package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/config"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestIronCoreCollectorKeepsSuccessfulGroupsWhenOneGroupFails(t *testing.T) {
	t.Parallel()

	vm := unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "compute.ironcore.dev/v1alpha1", "kind": "Machine",
		"metadata": map[string]any{"namespace": "tenant", "name": "vm-a", "uid": "vm-a-uid"},
		"status":   map[string]any{"state": "Running"},
	}}
	reader := &fakeReader{
		pages: map[string][]*unstructured.UnstructuredList{
			listKey(ironMachinesGVR, ""): {{Items: []unstructured.Unstructured{vm}}},
		},
		listErrors: map[string]error{listKey(networksGVR, ""): errors.New("network API unavailable")},
	}
	collector := NewIronCoreCollector(config.Source{ID: "canary-ironcore", Environment: "canary", Role: config.RoleCentralIronCore}, reader, IronCoreOptions{PageSize: 100, RequestTimeout: time.Second})

	got, err := collector.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if !got.Partial || got.ObjectCounts["IronCoreVM"] != 1 || len(got.Resources) != 1 {
		t.Fatalf("snapshot = %#v", got)
	}
	if len(got.Gaps) != 1 || got.Gaps[0].Access != model.AccessUnreachable {
		t.Fatalf("gaps = %#v", got.Gaps)
	}
}

func TestIronCoreCollectorUsesCanonicalFixedInventory(t *testing.T) {
	t.Parallel()

	reader := &fakeReader{}
	collector := NewIronCoreCollector(config.Source{ID: "canary-ironcore", Environment: "canary", Role: config.RoleCentralIronCore}, reader, IronCoreOptions{PageSize: 73, RequestTimeout: time.Second})
	if _, err := collector.Collect(context.Background()); err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	for _, gvr := range canonicalIronCoreInventory() {
		if !reader.sawPage(gvr, "", "", 73) {
			t.Fatalf("canonical inventory did not list %s; calls %#v", gvr, reader.calls)
		}
	}
}
