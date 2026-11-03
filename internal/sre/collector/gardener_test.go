package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/config"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestGardenerCollectorPublishesDeniedAndUnsupportedAccessGaps(t *testing.T) {
	t.Parallel()

	reader := &fakeReader{
		supported: map[schema.GroupVersionResource]bool{
			shootsGVR: true, seedsGVR: false, cloudProfilesGVR: false,
			mcmDeploymentsGVR: false, mcmSetsGVR: false, mcmMachinesGVR: false,
		},
		discoveryErrors: map[schema.GroupVersionResource]error{
			shootsGVR: apierrors.NewForbidden(schema.GroupResource{Group: shootsGVR.Group, Resource: shootsGVR.Resource}, "", errors.New("forbidden")),
		},
	}
	collector := NewGardenerCollector(config.Source{ID: "ironcore-consumer-gardener", Environment: "canary", Role: config.RoleGardener, InstallationRole: "consumer"}, reader, reader, GardenerOptions{PageSize: 50, RequestTimeout: time.Second})

	got, err := collector.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if !got.Partial || got.ObjectCounts["GardenerInstallation"] != 1 || got.ObjectCounts["Shoot"] != 0 {
		t.Fatalf("snapshot = %#v", got)
	}
	if len(got.Gaps) != 6 || got.Gaps[0].Access != model.AccessDenied {
		t.Fatalf("gaps = %#v", got.Gaps)
	}
	for _, call := range reader.calls {
		if call.gvr.Resource == "secrets" {
			t.Fatalf("collector attempted to list Secrets: %#v", call)
		}
	}
}

func TestGardenerCollectorCollectsOnlyServedAuthorizedInventory(t *testing.T) {
	t.Parallel()

	reader := &fakeReader{supported: map[schema.GroupVersionResource]bool{}}
	for _, gvr := range canonicalGardenerInventory() {
		reader.supported[gvr] = true
	}
	collector := NewGardenerCollector(config.Source{ID: "wdf4-bare-metal-gardener", Environment: "canary", Role: config.RoleGardener, InstallationRole: "bare-metal"}, reader, reader, GardenerOptions{PageSize: 67, RequestTimeout: time.Second})
	got, err := collector.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.Partial || len(got.Gaps) != 0 {
		t.Fatalf("snapshot = %#v", got)
	}
	for _, gvr := range canonicalGardenerInventory() {
		if !reader.sawPage(gvr, "", "", 67) {
			t.Fatalf("served resource %s was not listed", gvr)
		}
	}
}
