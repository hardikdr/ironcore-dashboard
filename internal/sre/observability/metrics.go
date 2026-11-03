package observability

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Metrics is a dependency-free Prometheus text collector. Labels are restricted
// to configured source IDs, normalized routes, status codes, and known kinds.
// It never accepts object payloads, names, namespaces, or credential material.
type Metrics struct {
	mu                 sync.RWMutex
	collectionSuccess  map[string]uint64
	collectionFailure  map[string]uint64
	collectionDuration map[string]float64
	collectionObjects  map[string]float64
	snapshotAge        map[string]float64
	apiRequests        map[string]uint64
	apiLatency         map[string]float64
	partialResponses   uint64
	searchResults      float64
	graphResources     float64
	graphRelations     float64
}

func NewMetrics() *Metrics {
	return &Metrics{
		collectionSuccess: map[string]uint64{}, collectionFailure: map[string]uint64{},
		collectionDuration: map[string]float64{}, collectionObjects: map[string]float64{},
		snapshotAge: map[string]float64{}, apiRequests: map[string]uint64{}, apiLatency: map[string]float64{},
	}
}

func (m *Metrics) ObserveCollection(source string, success bool, duration time.Duration, counts map[string]int) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	source = label(source)
	if success {
		m.collectionSuccess[source]++
	} else {
		m.collectionFailure[source]++
	}
	m.collectionDuration[source] = duration.Seconds()
	for kind, count := range counts {
		m.collectionObjects[source+"\x00"+label(kind)] = float64(count)
	}
}

func (m *Metrics) ObserveSnapshotAge(source string, age time.Duration) {
	if m == nil {
		return
	}
	if age < 0 {
		age = 0
	}
	m.mu.Lock()
	m.snapshotAge[label(source)] = age.Seconds()
	m.mu.Unlock()
}

func (m *Metrics) ObserveAPI(route string, status int, duration time.Duration, partial bool) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	key := label(route) + "\x00" + strconv.Itoa(status)
	m.apiRequests[key]++
	m.apiLatency[label(route)] = duration.Seconds()
	if partial {
		m.partialResponses++
	}
}

func (m *Metrics) ObserveSearchSize(size int) {
	if m != nil {
		m.mu.Lock()
		m.searchResults = float64(size)
		m.mu.Unlock()
	}
}
func (m *Metrics) ObserveGraphSize(resources, relations int) {
	if m != nil {
		m.mu.Lock()
		m.graphResources = float64(resources)
		m.graphRelations = float64(relations)
		m.mu.Unlock()
	}
}

func (m *Metrics) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		m.mu.RLock()
		defer m.mu.RUnlock()
		writeMapCounter(w, "ironcore_sre_collection_success_total", "source", m.collectionSuccess)
		writeMapCounter(w, "ironcore_sre_collection_failure_total", "source", m.collectionFailure)
		writeMapGauge(w, "ironcore_sre_collection_duration_seconds", "source", m.collectionDuration)
		writePairGauge(w, "ironcore_sre_collection_objects", "source", "kind", m.collectionObjects)
		writeMapGauge(w, "ironcore_sre_snapshot_age_seconds", "source", m.snapshotAge)
		writePairCounter(w, "ironcore_sre_api_requests_total", "route", "status", m.apiRequests)
		writeMapGauge(w, "ironcore_sre_api_latency_seconds", "route", m.apiLatency)
		fmt.Fprintf(w, "# TYPE ironcore_sre_partial_responses_total counter\nironcore_sre_partial_responses_total %d\n", m.partialResponses)
		fmt.Fprintf(w, "# TYPE ironcore_sre_search_results gauge\nironcore_sre_search_results %g\n", m.searchResults)
		fmt.Fprintf(w, "# TYPE ironcore_sre_graph_resources gauge\nironcore_sre_graph_resources %g\n", m.graphResources)
		fmt.Fprintf(w, "# TYPE ironcore_sre_graph_relations gauge\nironcore_sre_graph_relations %g\n", m.graphRelations)
	})
}

func label(value string) string {
	value = strings.ReplaceAll(value, "\\", "_")
	value = strings.ReplaceAll(value, "\"", "_")
	value = strings.ReplaceAll(value, "\n", "_")
	if len(value) > 128 {
		return value[:128]
	}
	return value
}
func keys[V any](values map[string]V) []string {
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}
func writeMapCounter(w http.ResponseWriter, name, labelName string, values map[string]uint64) {
	fmt.Fprintf(w, "# TYPE %s counter\n", name)
	for _, key := range keys(values) {
		fmt.Fprintf(w, "%s{%s=\"%s\"} %d\n", name, labelName, key, values[key])
	}
}
func writeMapGauge(w http.ResponseWriter, name, labelName string, values map[string]float64) {
	fmt.Fprintf(w, "# TYPE %s gauge\n", name)
	for _, key := range keys(values) {
		fmt.Fprintf(w, "%s{%s=\"%s\"} %g\n", name, labelName, key, values[key])
	}
}
func writePairGauge(w http.ResponseWriter, name, first, second string, values map[string]float64) {
	fmt.Fprintf(w, "# TYPE %s gauge\n", name)
	for _, key := range keys(values) {
		parts := strings.SplitN(key, "\x00", 2)
		fmt.Fprintf(w, "%s{%s=\"%s\",%s=\"%s\"} %g\n", name, first, parts[0], second, parts[1], values[key])
	}
}
func writePairCounter(w http.ResponseWriter, name, first, second string, values map[string]uint64) {
	fmt.Fprintf(w, "# TYPE %s counter\n", name)
	for _, key := range keys(values) {
		parts := strings.SplitN(key, "\x00", 2)
		fmt.Fprintf(w, "%s{%s=\"%s\",%s=\"%s\"} %d\n", name, first, parts[0], second, parts[1], values[key])
	}
}
