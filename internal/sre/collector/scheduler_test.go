package collector

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/snapshot"
)

func TestSchedulerLimitsConcurrentSourcePolls(t *testing.T) {
	t.Parallel()

	started := make(chan string, 5)
	release := make(chan struct{})
	var active atomic.Int32
	var maximum atomic.Int32
	collectors := make([]Collector, 0, 5)
	for index := 0; index < 5; index++ {
		collectors = append(collectors, &blockingCollector{id: string(rune('a' + index)), started: started, release: release, active: &active, maximum: &maximum})
	}
	scheduler := NewScheduler(collectors, snapshot.NewStore(), SchedulerOptions{MaxConcurrent: 2, Interval: time.Minute, JitterFraction: 0.1})
	done := make(chan error, 1)
	go func() { done <- scheduler.Poll(context.Background()) }()

	<-started
	<-started
	select {
	case third := <-started:
		t.Fatalf("third collector %q started before a slot was released", third)
	case <-time.After(25 * time.Millisecond):
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatalf("Poll() error = %v", err)
	}
	if got := maximum.Load(); got != 2 {
		t.Fatalf("maximum concurrency = %d", got)
	}
}

func TestSchedulerPublishesFailureWithoutErasingSuccess(t *testing.T) {
	t.Parallel()

	store := snapshot.NewStore()
	collector := &sequenceCollector{id: "wdf4a", outcomes: []collectorOutcome{
		{snapshot: snapshot.Snapshot{SourceID: "wdf4a", ObjectCounts: map[string]int{"PhysicalNode": 1}}},
		{err: &CollectionError{Class: "timeout", Access: model.AccessUnreachable, Err: errors.New("deadline")}},
	}}
	scheduler := NewScheduler([]Collector{collector}, store, SchedulerOptions{MaxConcurrent: 1, Interval: time.Minute, JitterFraction: 0.1})
	if err := scheduler.Poll(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := scheduler.Poll(context.Background()); err == nil {
		t.Fatal("second Poll() succeeded, want collection error")
	}
	view, ok := store.Read("wdf4a", time.Now(), time.Hour)
	if !ok || view.Snapshot.ObjectCounts["PhysicalNode"] != 1 || view.Status.ErrorClass != snapshot.ErrorTimeout {
		t.Fatalf("stored view = %#v", view)
	}
}

func TestSchedulerDelayIncludesBoundedJitter(t *testing.T) {
	t.Parallel()

	values := []float64{0, 0.5, 1}
	index := 0
	scheduler := NewScheduler(nil, snapshot.NewStore(), SchedulerOptions{
		Interval: time.Minute, JitterFraction: 0.1,
		Random: func() float64 { value := values[index]; index++; return value },
	})
	for want, got := range map[time.Duration]time.Duration{
		54 * time.Second: scheduler.NextDelay(),
		60 * time.Second: scheduler.NextDelay(),
		66 * time.Second: scheduler.NextDelay(),
	} {
		if got != want {
			t.Fatalf("delay = %s, want %s", got, want)
		}
	}
}

func TestSlowSourceDoesNotBlockOtherCollectors(t *testing.T) {
	t.Parallel()
	store := snapshot.NewStore()
	started, release := make(chan string, 1), make(chan struct{})
	slow := &blockingCollector{id: "slow", started: started, release: release, active: &atomic.Int32{}, maximum: &atomic.Int32{}}
	fast := &sequenceCollector{id: "fast", outcomes: []collectorOutcome{{snapshot: snapshot.Snapshot{SourceID: "fast", ObjectCounts: map[string]int{"IronCoreVM": 1}}}}}
	scheduler := NewScheduler([]Collector{slow, fast}, store, SchedulerOptions{MaxConcurrent: 2, Interval: time.Minute})
	done := make(chan error, 1)
	go func() { done <- scheduler.Poll(context.Background()) }()
	<-started
	deadline := time.Now().Add(time.Second)
	for {
		if view, ok := store.Read("fast", time.Now(), time.Minute); ok && view.Snapshot.ObjectCounts["IronCoreVM"] == 1 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("fast collector was blocked by slow source")
		}
		time.Sleep(time.Millisecond)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestStructuredCollectionLogsExcludePayloads(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	current := &sequenceCollector{id: "safe-source", outcomes: []collectorOutcome{{snapshot: snapshot.Snapshot{SourceID: "safe-source", ObjectCounts: map[string]int{"IronCoreVM": 1}, Warnings: []string{"credential-do-not-log"}}}}}
	scheduler := NewScheduler([]Collector{current}, snapshot.NewStore(), SchedulerOptions{Logger: logger})
	if err := scheduler.Poll(context.Background()); err != nil {
		t.Fatal(err)
	}
	logged := output.String()
	if !strings.Contains(logged, `"source_id":"safe-source"`) || !strings.Contains(logged, `"phase":"collect"`) {
		t.Fatalf("structured fields missing: %s", logged)
	}
	if strings.Contains(logged, "credential-do-not-log") {
		t.Fatalf("collector payload leaked to logs: %s", logged)
	}
}

type blockingCollector struct {
	id      string
	started chan<- string
	release <-chan struct{}
	active  *atomic.Int32
	maximum *atomic.Int32
}

func (collector *blockingCollector) SourceID() string { return collector.id }

func (collector *blockingCollector) Collect(ctx context.Context) (snapshot.Snapshot, error) {
	current := collector.active.Add(1)
	defer collector.active.Add(-1)
	for {
		maximum := collector.maximum.Load()
		if current <= maximum || collector.maximum.CompareAndSwap(maximum, current) {
			break
		}
	}
	collector.started <- collector.id
	select {
	case <-collector.release:
		return snapshot.Snapshot{SourceID: collector.id}, nil
	case <-ctx.Done():
		return snapshot.Snapshot{}, ctx.Err()
	}
}

type collectorOutcome struct {
	snapshot snapshot.Snapshot
	err      error
}

type sequenceCollector struct {
	mu       sync.Mutex
	id       string
	outcomes []collectorOutcome
}

func (collector *sequenceCollector) SourceID() string { return collector.id }

func (collector *sequenceCollector) Collect(context.Context) (snapshot.Snapshot, error) {
	collector.mu.Lock()
	defer collector.mu.Unlock()
	outcome := collector.outcomes[0]
	collector.outcomes = collector.outcomes[1:]
	return outcome.snapshot, outcome.err
}
