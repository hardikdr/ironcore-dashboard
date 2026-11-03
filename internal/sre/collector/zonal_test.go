package collector

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/config"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestZonalCollectorUsesPaginationTimeoutsAndProviderAllowlists(t *testing.T) {
	t.Parallel()

	reader := &fakeReader{
		pages: map[string][]*unstructured.UnstructuredList{
			listKey(nodesGVR, ""): {
				pageWithContinue("next", nodeObject("node-a", map[string]string{"ironcore.dev/role": "compute", "custom": "kept"})),
				{Items: []unstructured.Unstructured{nodeObject("node-b", map[string]string{"node-role.kubernetes.io/control-plane": ""})}},
			},
		},
		supported: map[schema.GroupVersionResource]bool{mcmMachinesGVR: false, metalServersGVR: false},
	}
	collector := NewZonalCollector(config.Source{
		ID: "wdf4a", Environment: "canary", Role: config.RoleZonalCAPI, Zone: "wdf4a",
		ProviderNamespaces: []string{"ironcore-system", "metalnet-system"},
	}, reader, reader, ZonalOptions{PageSize: 100, RequestTimeout: 2 * time.Second})

	got, err := collector.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if got.SourceID != "wdf4a" || got.ObjectCounts["PhysicalNode"] != 2 {
		t.Fatalf("snapshot = %#v", got)
	}
	if len(got.Resources) < 2 || got.Resources[0].Type != "physical-node" || got.Resources[0].Sanitized["physicalRole"] != "compute" {
		t.Fatalf("normalized first node = %#v", got.Resources)
	}
	if got.Resources[0].Labels["custom"] != "kept" {
		t.Fatalf("raw labels were not retained: %#v", got.Resources[0].Labels)
	}
	if !reader.sawPage(nodesGVR, "", "", 100) || !reader.sawPage(nodesGVR, "", "next", 100) {
		t.Fatalf("node list calls = %#v", reader.calls)
	}
	for _, call := range reader.calls {
		if !call.hadDeadline {
			t.Fatalf("list call lacked timeout: %#v", call)
		}
	}
	for _, namespace := range []string{"ironcore-system", "metalnet-system"} {
		for _, gvr := range []schema.GroupVersionResource{podsGVR, eventsGVR, deploymentsGVR, daemonSetsGVR} {
			if !reader.sawResource(gvr, namespace) {
				t.Fatalf("missing %s list in namespace %q", gvr.Resource, namespace)
			}
		}
	}
	if !reader.sawResource(capiMachinesGVR, metav1.NamespaceAll) {
		t.Fatal("CAPI Machines were not listed cluster-wide")
	}
	if len(reader.discoveryCalls) != 2 {
		t.Fatalf("optional discovery calls = %d, want one per optional API", len(reader.discoveryCalls))
	}
	if len(got.Gaps) != 2 || got.Gaps[0].Access != model.AccessUnsupported || got.Gaps[1].Access != model.AccessUnsupported {
		t.Fatalf("optional gaps = %#v", got.Gaps)
	}
}

func TestZonalCollectorDistinguishesDeniedOptionalAPI(t *testing.T) {
	t.Parallel()

	reader := &fakeReader{
		supported: map[schema.GroupVersionResource]bool{mcmMachinesGVR: true, metalServersGVR: false},
		discoveryErrors: map[schema.GroupVersionResource]error{
			mcmMachinesGVR: apierrors.NewForbidden(schema.GroupResource{Group: mcmMachinesGVR.Group, Resource: mcmMachinesGVR.Resource}, "", errors.New("forbidden")),
		},
	}
	collector := NewZonalCollector(config.Source{ID: "wdf4a", Environment: "canary", Role: config.RoleZonalCAPI, Zone: "wdf4a"}, reader, reader, ZonalOptions{PageSize: 50, RequestTimeout: time.Second})

	got, err := collector.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if len(got.Gaps) != 2 || got.Gaps[0].Access != model.AccessDenied || got.Gaps[1].Access != model.AccessUnsupported {
		t.Fatalf("gaps = %#v", got.Gaps)
	}
	if !got.Partial {
		t.Fatal("snapshot with denied optional inventory is not partial")
	}
}

func TestZonalCollectorClassifiesTotalRequiredFailure(t *testing.T) {
	t.Parallel()

	reader := &fakeReader{
		listError: errors.New("transport down"),
		supported: map[schema.GroupVersionResource]bool{mcmMachinesGVR: false, metalServersGVR: false},
	}
	collector := NewZonalCollector(config.Source{ID: "wdf4a", Environment: "canary", Role: config.RoleZonalCAPI, Zone: "wdf4a"}, reader, reader, ZonalOptions{PageSize: 50, RequestTimeout: time.Second})

	_, err := collector.Collect(context.Background())
	var collectionErr *CollectionError
	if !errors.As(err, &collectionErr) || collectionErr.Class != "transport" || collectionErr.Access != model.AccessUnreachable {
		t.Fatalf("Collect() error = %#v", err)
	}
}

type listCall struct {
	gvr         schema.GroupVersionResource
	namespace   string
	continueVal string
	limit       int64
	hadDeadline bool
}

type fakeReader struct {
	pages           map[string][]*unstructured.UnstructuredList
	pageIndexes     map[string]int
	listError       error
	supported       map[schema.GroupVersionResource]bool
	discoveryErrors map[schema.GroupVersionResource]error
	calls           []listCall
	discoveryCalls  []schema.GroupVersionResource
}

func (reader *fakeReader) List(ctx context.Context, gvr schema.GroupVersionResource, namespace string, options metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	_, hadDeadline := ctx.Deadline()
	reader.calls = append(reader.calls, listCall{gvr: gvr, namespace: namespace, continueVal: options.Continue, limit: options.Limit, hadDeadline: hadDeadline})
	if reader.listError != nil {
		return nil, reader.listError
	}
	key := listKey(gvr, namespace)
	if reader.pageIndexes == nil {
		reader.pageIndexes = make(map[string]int)
	}
	index := reader.pageIndexes[key]
	reader.pageIndexes[key]++
	if index < len(reader.pages[key]) {
		return reader.pages[key][index].DeepCopy(), nil
	}
	return &unstructured.UnstructuredList{}, nil
}

func (reader *fakeReader) Supports(_ context.Context, gvr schema.GroupVersionResource) (bool, error) {
	reader.discoveryCalls = append(reader.discoveryCalls, gvr)
	if err := reader.discoveryErrors[gvr]; err != nil {
		return false, err
	}
	return reader.supported[gvr], nil
}

func (reader *fakeReader) sawPage(gvr schema.GroupVersionResource, namespace, continueVal string, limit int64) bool {
	for _, call := range reader.calls {
		if call.gvr == gvr && call.namespace == namespace && call.continueVal == continueVal && call.limit == limit {
			return true
		}
	}
	return false
}

func (reader *fakeReader) sawResource(gvr schema.GroupVersionResource, namespace string) bool {
	for _, call := range reader.calls {
		if call.gvr == gvr && call.namespace == namespace {
			return true
		}
	}
	return false
}

func listKey(gvr schema.GroupVersionResource, namespace string) string {
	return fmt.Sprintf("%s/%s/%s|%s", gvr.Group, gvr.Version, gvr.Resource, namespace)
}

func nodeObject(name string, labels map[string]string) unstructured.Unstructured {
	return unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "v1", "kind": "Node",
		"metadata": map[string]any{"name": name, "uid": name + "-uid", "labels": stringMapToAny(labels)},
		"status":   map[string]any{"conditions": []any{map[string]any{"type": "Ready", "status": "True"}}},
	}}
}

func pageWithContinue(continueValue string, objects ...unstructured.Unstructured) *unstructured.UnstructuredList {
	page := &unstructured.UnstructuredList{Items: objects}
	page.SetContinue(continueValue)
	return page
}

func stringMapToAny(input map[string]string) map[string]any {
	output := make(map[string]any, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}
