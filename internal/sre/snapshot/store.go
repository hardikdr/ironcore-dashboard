package snapshot

import (
	"sort"
	"sync"
	"time"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
)

type sourceState struct {
	generation uint64
	snapshot   Snapshot
	status     SourceStatus
}

type Store struct {
	mu         sync.RWMutex
	generation uint64
	sources    map[string]sourceState
}

func NewStore() *Store {
	return &Store{sources: make(map[string]sourceState)}
}

func (store *Store) PublishSuccess(sourceID string, collectedAt time.Time, duration time.Duration, snapshot Snapshot) uint64 {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.generation++
	snapshot.SourceID = sourceID
	snapshot.CollectedAt = collectedAt
	store.sources[sourceID] = sourceState{
		generation: store.generation,
		snapshot:   cloneSnapshot(snapshot),
		status: SourceStatus{
			LastAttempt: collectedAt,
			LastSuccess: collectedAt,
			Duration:    duration,
			Access:      model.AccessOK,
		},
	}
	return store.generation
}

func (store *Store) PublishFailure(sourceID string, attemptedAt time.Time, duration time.Duration, errorClass ErrorClass, access model.AccessState) uint64 {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.generation++
	state, found := store.sources[sourceID]
	if !found {
		state.snapshot.SourceID = sourceID
	}
	state.generation = store.generation
	state.status.LastAttempt = attemptedAt
	state.status.Duration = duration
	state.status.ErrorClass = errorClass
	state.status.Access = access
	store.sources[sourceID] = state
	return store.generation
}

func (store *Store) Read(sourceID string, now time.Time, staleAfter time.Duration) (SourceView, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	state, ok := store.sources[sourceID]
	if !ok {
		return SourceView{}, false
	}
	return makeView(sourceID, state, now, staleAfter), true
}

func (store *Store) Aggregate(now time.Time, staleAfter time.Duration) Aggregate {
	store.mu.RLock()
	defer store.mu.RUnlock()

	ids := make([]string, 0, len(store.sources))
	for id := range store.sources {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result := Aggregate{
		Generation:  store.generation,
		GeneratedAt: now,
		Sources:     make([]SourceView, 0, len(ids)),
	}
	for _, id := range ids {
		result.Sources = append(result.Sources, makeView(id, store.sources[id], now, staleAfter))
	}
	return result
}

func makeView(sourceID string, state sourceState, now time.Time, staleAfter time.Duration) SourceView {
	status := state.status
	status.Freshness = freshnessAt(status.LastSuccess, now, staleAfter)
	return SourceView{
		SourceID:   sourceID,
		Generation: state.generation,
		Snapshot:   cloneSnapshot(state.snapshot),
		Status:     status,
	}
}

func freshnessAt(lastSuccess, now time.Time, staleAfter time.Duration) model.Freshness {
	if lastSuccess.IsZero() {
		return model.FreshnessUnknown
	}
	if staleAfter > 0 && now.Sub(lastSuccess) > staleAfter {
		return model.FreshnessStale
	}
	return model.FreshnessFresh
}

func cloneSnapshot(input Snapshot) Snapshot {
	output := input
	output.ObjectCounts = cloneStringIntMap(input.ObjectCounts)
	output.Warnings = append([]string(nil), input.Warnings...)
	output.Gaps = append([]AccessGap(nil), input.Gaps...)
	output.Resources = make([]model.Resource, len(input.Resources))
	for index := range input.Resources {
		output.Resources[index] = cloneResource(input.Resources[index])
	}
	return output
}

func cloneResource(input model.Resource) model.Resource {
	output := input
	output.Labels = cloneStringMap(input.Labels)
	output.Annotations = cloneStringMap(input.Annotations)
	output.Addresses = append([]string(nil), input.Addresses...)
	output.Aliases = append([]string(nil), input.Aliases...)
	output.Conditions = append([]model.Condition(nil), input.Conditions...)
	output.References = append([]model.ObjectReference(nil), input.References...)
	output.OwnerRefs = append([]model.ObjectReference(nil), input.OwnerRefs...)
	output.Selector = cloneStringMap(input.Selector)
	if input.Sanitized != nil {
		output.Sanitized = make(map[string]any, len(input.Sanitized))
		for key, value := range input.Sanitized {
			output.Sanitized[key] = cloneValue(value)
		}
	}
	return output
}

func cloneStringMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func cloneStringIntMap(input map[string]int) map[string]int {
	if input == nil {
		return nil
	}
	output := make(map[string]int, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func cloneValue(input any) any {
	switch typed := input.(type) {
	case map[string]any:
		output := make(map[string]any, len(typed))
		for key, value := range typed {
			output[key] = cloneValue(value)
		}
		return output
	case []any:
		output := make([]any, len(typed))
		for index, value := range typed {
			output[index] = cloneValue(value)
		}
		return output
	case map[string]string:
		return cloneStringMap(typed)
	case []string:
		return append([]string(nil), typed...)
	default:
		return typed
	}
}
