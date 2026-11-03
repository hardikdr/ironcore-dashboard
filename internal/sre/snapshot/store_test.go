package snapshot

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
)

func TestStorePublishesImmutableSnapshotAndComputesFreshnessAtReadTime(t *testing.T) {
	t.Parallel()

	store := NewStore()
	collectedAt := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	input := Snapshot{
		SourceID: "canary-ironcore",
		Resources: []model.Resource{{
			Key:    "canary-ironcore/compute.ironcore.dev/IronCoreVM/default/vm-a/vm-a-uid",
			Type:   "ironcore-vm",
			Labels: map[string]string{"zone": "wdf4a"},
			Sanitized: map[string]any{
				"conditions": []any{map[string]any{"type": "Ready", "status": "True"}},
			},
		}},
		ObjectCounts: map[string]int{"IronCoreVM": 1},
	}
	store.PublishSuccess("canary-ironcore", collectedAt, 120*time.Millisecond, input)

	input.Resources[0].Labels["zone"] = "corrupted"
	input.Resources[0].Sanitized["conditions"].([]any)[0].(map[string]any)["status"] = "False"
	input.ObjectCounts["IronCoreVM"] = 999

	view, ok := store.Read("canary-ironcore", collectedAt.Add(time.Minute), 2*time.Minute)
	if !ok {
		t.Fatal("Read() source not found")
	}
	if view.Status.Freshness != model.FreshnessFresh {
		t.Fatalf("freshness = %q", view.Status.Freshness)
	}
	if got := view.Snapshot.Resources[0].Labels["zone"]; got != "wdf4a" {
		t.Fatalf("stored label = %q", got)
	}
	if got := view.Snapshot.Resources[0].Sanitized["conditions"].([]any)[0].(map[string]any)["status"]; got != "True" {
		t.Fatalf("stored nested value = %v", got)
	}
	if got := view.Snapshot.ObjectCounts["IronCoreVM"]; got != 1 {
		t.Fatalf("stored object count = %d", got)
	}

	view.Snapshot.Resources[0].Labels["zone"] = "mutated-reader"
	again, _ := store.Read("canary-ironcore", collectedAt.Add(3*time.Minute), 2*time.Minute)
	if again.Status.Freshness != model.FreshnessStale || again.Snapshot.Resources[0].Labels["zone"] != "wdf4a" {
		t.Fatalf("second read = %#v", again)
	}
}

func TestStoreRetainsLastSuccessWhenPollFails(t *testing.T) {
	t.Parallel()

	store := NewStore()
	successAt := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	store.PublishSuccess("wdf4c", successAt, time.Second, Snapshot{
		SourceID:     "wdf4c",
		Resources:    []model.Resource{{Key: "wdf4c/core/PhysicalNode/_/node-c/node-c-uid", Type: "physical-node"}},
		ObjectCounts: map[string]int{"PhysicalNode": 1},
	})
	first, _ := store.Read("wdf4c", successAt, time.Minute)

	attemptAt := successAt.Add(90 * time.Second)
	store.PublishFailure("wdf4c", attemptAt, 15*time.Second, ErrorTimeout, model.AccessUnreachable)
	got, ok := store.Read("wdf4c", attemptAt, time.Minute)
	if !ok {
		t.Fatal("Read() source not found")
	}
	if len(got.Snapshot.Resources) != 1 || got.Status.LastSuccess != successAt {
		t.Fatalf("last success was not retained: %#v", got)
	}
	if got.Status.LastAttempt != attemptAt || got.Status.Duration != 15*time.Second || got.Status.ErrorClass != ErrorTimeout || got.Status.Access != model.AccessUnreachable {
		t.Fatalf("failure status = %#v", got.Status)
	}
	if got.Generation <= first.Generation {
		t.Fatalf("generation did not advance: first %d, failed %d", first.Generation, got.Generation)
	}
}

func TestAggregateIsSortedAndUsesOneGeneration(t *testing.T) {
	t.Parallel()

	store := NewStore()
	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	store.PublishFailure("wdf4b", now, time.Second, ErrorRBAC, model.AccessDenied)
	store.PublishSuccess("wdf4a", now, time.Second, Snapshot{SourceID: "wdf4a"})

	got := store.Aggregate(now, time.Minute)
	if len(got.Sources) != 2 || got.Sources[0].SourceID != "wdf4a" || got.Sources[1].SourceID != "wdf4b" {
		t.Fatalf("aggregate sources = %#v", got.Sources)
	}
	if got.Generation == 0 || got.Sources[1].Status.Freshness != model.FreshnessUnknown {
		t.Fatalf("aggregate = %#v", got)
	}
}

func TestStoreConcurrentPublishAndRead(t *testing.T) {
	store := NewStore()
	start := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	var group sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		worker := worker
		group.Add(1)
		go func() {
			defer group.Done()
			for iteration := 0; iteration < 200; iteration++ {
				sourceID := fmt.Sprintf("source-%d", worker%3)
				now := start.Add(time.Duration(iteration) * time.Second)
				if iteration%5 == 0 {
					store.PublishFailure(sourceID, now, time.Millisecond, ErrorTransport, model.AccessUnreachable)
				} else {
					store.PublishSuccess(sourceID, now, time.Millisecond, Snapshot{SourceID: sourceID, ObjectCounts: map[string]int{"objects": iteration}})
				}
				store.Read(sourceID, now, time.Minute)
				store.Aggregate(now, time.Minute)
			}
		}()
	}
	group.Wait()
}
