package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/diagnosis"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/graph"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/search"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/snapshot"
)

type currentState struct {
	aggregate snapshot.Aggregate
	resources []model.Resource
	graph     *graph.Graph
	search    *search.Index
	meta      Meta
}

func (api *API) current() currentState {
	aggregate := api.store.Aggregate(time.Now().UTC(), api.staleAfter)
	state := currentState{aggregate: aggregate}
	state.meta.GeneratedAt = aggregate.GeneratedAt
	for _, source := range aggregate.Sources {
		environment := ""
		if len(source.Snapshot.Resources) != 0 {
			environment = source.Snapshot.Resources[0].Provenance.Environment
		}
		if !source.Status.LastSuccess.IsZero() {
			api.metrics.ObserveSnapshotAge(source.SourceID, time.Since(source.Status.LastSuccess))
		}
		state.resources = append(state.resources, source.Snapshot.Resources...)
		state.meta.Sources = append(state.meta.Sources, SourceMeta{
			ID: source.SourceID, Environment: environment, CollectedAt: source.Status.LastSuccess, Freshness: source.Status.Freshness,
			Access: source.Status.Access, Duration: source.Status.Duration, ErrorClass: string(source.Status.ErrorClass), ObjectCount: source.Snapshot.ObjectCounts,
		})
		if source.Snapshot.Partial || source.Status.Access != model.AccessOK || source.Status.Freshness != model.FreshnessFresh {
			state.meta.Partial = true
		}
		state.meta.Warnings = append(state.meta.Warnings, source.Snapshot.Warnings...)
		for _, gap := range source.Snapshot.Gaps {
			state.meta.Warnings = append(state.meta.Warnings, fmt.Sprintf("%s: %s", gap.Resource, gap.Access))
		}
	}
	state.graph = graph.Build(state.resources)
	state.search = search.New(state.resources)
	return state
}

func (api *API) status(w http.ResponseWriter, request *http.Request) {
	state := api.current()
	api.write(w, request, state, http.StatusOK, map[string]any{"status": "ok", "generation": state.aggregate.Generation})
}

func (api *API) sources(w http.ResponseWriter, request *http.Request) {
	state := api.current()
	api.write(w, request, state, http.StatusOK, state.meta.Sources)
}

func (api *API) summary(w http.ResponseWriter, request *http.Request) {
	state := api.current()
	counts := make(map[string]int)
	for _, resource := range filterResources(state.resources, request) {
		counts[resource.Type]++
	}
	api.write(w, request, state, http.StatusOK, counts)
}

func (api *API) search(w http.ResponseWriter, request *http.Request) {
	state := api.current()
	query := strings.TrimSpace(request.URL.Query().Get("q"))
	if len(query) < 2 {
		writeError(w, http.StatusBadRequest, "query must contain at least two characters")
		return
	}
	limit := boundedInt(request.URL.Query().Get("limit"), 50, 1, 200)
	types := splitCSV(request.URL.Query().Get("types"))
	results := state.search.Search(query, search.Options{Types: types, Limit: limit})
	api.metrics.ObserveSearchSize(len(results))
	data := make([]SearchDTO, 0, len(results))
	for _, result := range results {
		data = append(data, SearchDTO{ID: EncodeID(result.Resource.Key), Resource: publicResource(result.Resource), MatchedField: result.Match.Field, MatchKind: string(result.Match.Kind), Fragment: result.Match.Fragment, Ranges: result.Match.Ranges})
	}
	api.write(w, request, state, http.StatusOK, data)
}

func (api *API) zones(w http.ResponseWriter, request *http.Request) { api.listKind(w, request, "Zone") }
func (api *API) vms(w http.ResponseWriter, request *http.Request) {
	api.listKind(w, request, "IronCoreVM")
}
func (api *API) shoots(w http.ResponseWriter, request *http.Request) {
	api.listKind(w, request, "Shoot")
}

func (api *API) listKind(w http.ResponseWriter, request *http.Request, kind string) {
	state := api.current()
	var resources []model.Resource
	for _, resource := range filterResources(state.resources, request) {
		if resource.Provenance.Kind == kind {
			resources = append(resources, resource)
		}
	}
	sort.Slice(resources, func(i, j int) bool { return resources[i].Key < resources[j].Key })
	page, next := paginate(resources, request)
	state.meta.Continue = next
	data := make([]ResourceDTO, 0, len(page))
	for _, resource := range page {
		data = append(data, ResourceDTO{ID: EncodeID(resource.Key), Resource: publicResource(resource)})
	}
	api.write(w, request, state, http.StatusOK, data)
}

func (api *API) zone(w http.ResponseWriter, request *http.Request) {
	state := api.current()
	name := chi.URLParam(request, "zone")
	for _, resource := range state.resources {
		if resource.Provenance.Kind == "Zone" && resource.Provenance.Name == name {
			summary, err := diagnosis.DiagnoseZone(state.graph, sourceMap(state.aggregate), resource.Key)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "diagnosis failed")
				return
			}
			api.write(w, request, state, http.StatusOK, summary)
			return
		}
	}
	writeError(w, http.StatusNotFound, "zone not found")
}

func (api *API) node(w http.ResponseWriter, request *http.Request) {
	api.detail(w, request, "PhysicalNode")
}
func (api *API) network(w http.ResponseWriter, request *http.Request) {
	api.detail(w, request, "Network")
}
func (api *API) loadBalancer(w http.ResponseWriter, request *http.Request) {
	api.detail(w, request, "LoadBalancer")
}
func (api *API) volume(w http.ResponseWriter, request *http.Request) {
	api.detail(w, request, "Volume")
}

func (api *API) detail(w http.ResponseWriter, request *http.Request, kind string) {
	state := api.current()
	resource, ok := resourceFromID(state, chi.URLParam(request, "id"))
	if !ok || resource.Provenance.Kind != kind {
		writeError(w, http.StatusNotFound, "resource not found")
		return
	}
	api.write(w, request, state, http.StatusOK, ResourceDTO{ID: EncodeID(resource.Key), Resource: publicResource(resource)})
}

func (api *API) vm(w http.ResponseWriter, request *http.Request) {
	state := api.current()
	resource, ok := resourceFromID(state, chi.URLParam(request, "id"))
	if !ok {
		writeError(w, http.StatusNotFound, "VM not found")
		return
	}
	summary, err := diagnosis.DiagnoseVM(state.graph, sourceMap(state.aggregate), resource.Key)
	if err != nil {
		writeError(w, http.StatusNotFound, "VM not found")
		return
	}
	api.write(w, request, state, http.StatusOK, summary)
}

func (api *API) shoot(w http.ResponseWriter, request *http.Request) {
	state := api.current()
	resource, ok := resourceFromID(state, chi.URLParam(request, "id"))
	if !ok {
		writeError(w, http.StatusNotFound, "Shoot not found")
		return
	}
	summary, err := diagnosis.DiagnoseShoot(state.graph, sourceMap(state.aggregate), resource.Key)
	if err != nil {
		writeError(w, http.StatusNotFound, "Shoot not found")
		return
	}
	api.write(w, request, state, http.StatusOK, summary)
}

func (api *API) observations(w http.ResponseWriter, request *http.Request) {
	state := api.current()
	api.write(w, request, state, http.StatusOK, state.graph.Observations())
}

func (api *API) graph(w http.ResponseWriter, request *http.Request) {
	state := api.current()
	root, err := DecodeID(request.URL.Query().Get("root"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid root")
		return
	}
	depth := boundedInt(request.URL.Query().Get("depth"), 3, 1, graph.MaxTraversalDepth)
	limit := boundedInt(request.URL.Query().Get("limit"), 200, 1, graph.MaxTraversalNodes)
	result := state.graph.Traverse(root, depth, limit)
	api.metrics.ObserveGraphSize(len(result.Resources), len(result.Relations))
	if len(result.Resources) == 0 {
		writeError(w, http.StatusNotFound, "root not found")
		return
	}
	for index := range result.Resources {
		result.Resources[index] = publicResource(result.Resources[index])
	}
	api.write(w, request, state, http.StatusOK, result)
}

func (api *API) sanitized(w http.ResponseWriter, request *http.Request) {
	state := api.current()
	resource, ok := resourceFromID(state, chi.URLParam(request, "id"))
	if !ok {
		writeError(w, http.StatusNotFound, "resource not found")
		return
	}
	api.write(w, request, state, http.StatusOK, map[string]any{"id": EncodeID(resource.Key), "details": sanitizeValue(resource.Sanitized)})
}

func (api *API) write(w http.ResponseWriter, request *http.Request, state currentState, status int, data any) {
	if state.meta.Partial {
		if marker, ok := w.(interface{ markPartial() }); ok {
			marker.markPartial()
		}
	}
	etag := fmt.Sprintf(`W/"sre-%s-%d"`, api.etagSeed, state.aggregate.Generation)
	w.Header().Set("ETag", etag)
	if request.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	payload, err := json.Marshal(Envelope{Data: sanitizeValue(data), Meta: state.meta})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "response encoding failed")
		return
	}
	if len(payload) > MaxResponseBytes {
		writeError(w, http.StatusRequestEntityTooLarge, "response exceeds size cap")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(payload)
}

func sourceMap(aggregate snapshot.Aggregate) map[string]snapshot.SourceView {
	result := make(map[string]snapshot.SourceView, len(aggregate.Sources))
	for _, source := range aggregate.Sources {
		result[source.SourceID] = source
	}
	return result
}

func resourceFromID(state currentState, id string) (model.Resource, bool) {
	key, err := DecodeID(id)
	if err != nil {
		return model.Resource{}, false
	}
	return state.graph.Resource(key)
}

func publicResource(resource model.Resource) model.Resource {
	resource.Sanitized = nil
	return resource
}

func filterResources(resources []model.Resource, request *http.Request) []model.Resource {
	environment, zone, stateFilter := request.URL.Query().Get("environment"), request.URL.Query().Get("zone"), request.URL.Query().Get("state")
	result := make([]model.Resource, 0, len(resources))
	for _, resource := range resources {
		if environment != "" && resource.Provenance.Environment != environment {
			continue
		}
		if zone != "" && resource.Provenance.Name != zone && resource.Labels["topology.ironcore.dev/zone"] != zone {
			continue
		}
		if stateFilter != "" && !strings.EqualFold(resource.State, stateFilter) && !strings.EqualFold(string(resource.Health), stateFilter) {
			continue
		}
		result = append(result, resource)
	}
	return result
}

func paginate(resources []model.Resource, request *http.Request) ([]model.Resource, string) {
	limit := boundedInt(request.URL.Query().Get("limit"), 100, 1, 500)
	offset := boundedInt(request.URL.Query().Get("continue"), 0, 0, len(resources))
	end := offset + limit
	if end > len(resources) {
		end = len(resources)
	}
	next := ""
	if end < len(resources) {
		next = strconv.Itoa(end)
	}
	return resources[offset:end], next
}

func boundedInt(raw string, fallback, minimum, maximum int) int {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

var forbiddenResponseKey = regexp.MustCompile(`(?i)(token|password|secret|privatekey|clientkey|clientcertificate|ignition|credential|stringdata)`)
var forbiddenResponseValue = regexp.MustCompile(`(?i)(-----BEGIN [A-Z ]*(PRIVATE KEY|CERTIFICATE)-----|\bbearer\s+[a-z0-9._~+/=-]+)`)

func sanitizeValue(value any) any {
	body, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	var generic any
	if err := json.Unmarshal(body, &generic); err != nil {
		return nil
	}
	return scrub(generic)
}

func scrub(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, child := range typed {
			if forbiddenResponseKey.MatchString(key) {
				continue
			}
			result[key] = scrub(child)
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for index, child := range typed {
			result[index] = scrub(child)
		}
		return result
	case string:
		if forbiddenResponseValue.MatchString(typed) {
			return "[redacted]"
		}
	}
	return value
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
