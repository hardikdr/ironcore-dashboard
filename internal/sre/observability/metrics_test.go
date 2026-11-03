package observability

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMetricsExposeRequiredBoundedSignals(t *testing.T) {
	t.Parallel()
	metrics := NewMetrics()
	metrics.ObserveCollection("wdf4a", true, 250*time.Millisecond, map[string]int{"IronCoreVM": 12})
	metrics.ObserveSnapshotAge("wdf4a", 3*time.Second)
	metrics.ObserveAPI("/search", http.StatusOK, 20*time.Millisecond, true)
	metrics.ObserveSearchSize(12)
	metrics.ObserveGraphSize(40, 45)

	response := httptest.NewRecorder()
	metrics.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := response.Body.String()
	for _, name := range []string{
		"ironcore_sre_collection_success_total", "ironcore_sre_collection_duration_seconds",
		"ironcore_sre_collection_objects", "ironcore_sre_snapshot_age_seconds",
		"ironcore_sre_partial_responses_total", "ironcore_sre_api_requests_total",
		"ironcore_sre_api_latency_seconds", "ironcore_sre_graph_resources",
		"ironcore_sre_graph_relations", "ironcore_sre_search_results",
	} {
		if !strings.Contains(body, name) {
			t.Fatalf("metrics output missing %s: %s", name, body)
		}
	}
	if strings.Contains(body, "token") || strings.Contains(body, "password") {
		t.Fatalf("metrics leaked forbidden label: %s", body)
	}
}
