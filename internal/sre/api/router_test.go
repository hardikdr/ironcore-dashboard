package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/observability"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/snapshot"
)

func TestRouterRegistersOnlyGETEndpoints(t *testing.T) {
	t.Parallel()

	router := New(snapshot.NewStore(), time.Minute).Router()
	var routes []string
	err := chi.Walk(router, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		if method != http.MethodGet {
			t.Fatalf("SRE route %s registered mutation method %s", route, method)
		}
		routes = append(routes, route)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(routes) != 17 {
		t.Fatalf("registered %d routes, want 17: %#v", len(routes), routes)
	}

	request := httptest.NewRequest(http.MethodPost, "/search?q=vm", strings.NewReader(`{}`))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST status = %d", response.Code)
	}
}

func TestSearchUsesETagOpaqueIDsAndPartialMetadata(t *testing.T) {
	t.Parallel()

	store, vm := apiTestStore(t)
	router := New(store, time.Minute).Router()
	request := httptest.NewRequest(http.MethodGet, "/search?q=vm-a&limit=10", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("search status = %d body = %s", response.Code, response.Body.String())
	}
	etag := response.Header().Get("ETag")
	if etag == "" {
		t.Fatal("search response has no ETag")
	}
	var envelope struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
		Meta Meta `json:"meta"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if len(envelope.Data) != 1 || envelope.Data[0].ID == "" || strings.Contains(envelope.Data[0].ID, "/") {
		t.Fatalf("search IDs = %#v", envelope.Data)
	}
	if !envelope.Meta.Partial || len(envelope.Meta.Sources) != 2 {
		t.Fatalf("meta = %#v", envelope.Meta)
	}
	decoded, err := DecodeID(envelope.Data[0].ID)
	if err != nil || decoded != vm.Key {
		t.Fatalf("DecodeID() = %q, %v", decoded, err)
	}

	request = httptest.NewRequest(http.MethodGet, "/search?q=vm-a", nil)
	request.Header.Set("If-None-Match", etag)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusNotModified {
		t.Fatalf("conditional search status = %d", response.Code)
	}
}

func TestDetailRejectsRawKeysAndSanitizedEndpointRedactsAgain(t *testing.T) {
	t.Parallel()

	store, vm := apiTestStore(t)
	router := New(store, time.Minute).Router()

	rawRequest := httptest.NewRequest(http.MethodGet, "/vms/"+string(vm.Key), nil)
	rawResponse := httptest.NewRecorder()
	router.ServeHTTP(rawResponse, rawRequest)
	if rawResponse.Code == http.StatusOK {
		t.Fatalf("raw key unexpectedly accepted: %s", rawResponse.Body.String())
	}

	request := httptest.NewRequest(http.MethodGet, "/resources/"+EncodeID(vm.Key)+"/sanitized", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("sanitized status = %d body = %s", response.Code, response.Body.String())
	}
	for _, forbidden := range []string{"super-secret", "Bearer abc.def.ghi", "privateKey", "token"} {
		if strings.Contains(response.Body.String(), forbidden) {
			t.Fatalf("sanitized response leaked %q: %s", forbidden, response.Body.String())
		}
	}
}

func TestGraphCapsDepthAndNodeLimit(t *testing.T) {
	t.Parallel()
	store, vm := apiTestStore(t)
	router := New(store, time.Minute).Router()
	request := httptest.NewRequest(http.MethodGet, "/graph?root="+EncodeID(vm.Key)+"&depth=999&limit=99999", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("graph status = %d body = %s", response.Code, response.Body.String())
	}
	if response.Body.Len() > MaxResponseBytes {
		t.Fatalf("graph response exceeded cap: %d", response.Body.Len())
	}
}

func TestAPIRoutesRecordBoundedMetrics(t *testing.T) {
	t.Parallel()
	store, _ := apiTestStore(t)
	metrics := observability.NewMetrics()
	router := NewWithMetrics(store, time.Minute, metrics).Router()
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/search?q=vm-a", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("search status = %d", response.Code)
	}
	metricResponse := httptest.NewRecorder()
	metrics.Handler().ServeHTTP(metricResponse, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := metricResponse.Body.String()
	for _, expected := range []string{`route="/search"`, "ironcore_sre_partial_responses_total 1", "ironcore_sre_search_results 1", "ironcore_sre_snapshot_age_seconds"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("metrics missing %q: %s", expected, body)
		}
	}
}

func apiTestStore(t *testing.T) (*snapshot.Store, model.Resource) {
	t.Helper()
	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	key, err := model.NewResourceKey(model.ResourceIdentity{SourceID: "central", Group: "compute.ironcore.dev", Kind: "IronCoreVM", Namespace: "tenant", Name: "vm-a", UID: "vm-uid"})
	if err != nil {
		t.Fatal(err)
	}
	vm := model.Resource{
		Key: key, Type: "ironcore-vm", State: "Running",
		Provenance: model.Provenance{SourceID: "central", Environment: "canary", Group: "compute.ironcore.dev", Kind: "IronCoreVM", Namespace: "tenant", Name: "vm-a", UID: "vm-uid", CollectedAt: now, Access: model.AccessOK, Freshness: model.FreshnessFresh},
		Sanitized:  map[string]any{"safe": "visible", "token": "super-secret", "nested": map[string]any{"privateKey": "Bearer abc.def.ghi"}},
	}
	store := snapshot.NewStore()
	store.PublishSuccess("central", now, time.Second, snapshot.Snapshot{SourceID: "central", Resources: []model.Resource{vm}, ObjectCounts: map[string]int{"IronCoreVM": 1}})
	store.PublishFailure("gardener", now, time.Second, snapshot.ErrorRBAC, model.AccessDenied)
	return store, vm
}
