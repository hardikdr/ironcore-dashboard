package collector

import (
	"context"
	"errors"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/snapshot"
)

type SchedulerOptions struct {
	MaxConcurrent  int
	Interval       time.Duration
	JitterFraction float64
	Random         func() float64
}

type Scheduler struct {
	collectors []Collector
	store      *snapshot.Store
	options    SchedulerOptions
}

func NewScheduler(collectors []Collector, store *snapshot.Store, options SchedulerOptions) *Scheduler {
	if options.MaxConcurrent <= 0 {
		options.MaxConcurrent = 1
	}
	if options.Interval <= 0 {
		options.Interval = 30 * time.Second
	}
	if options.JitterFraction <= 0 {
		options.JitterFraction = 0.1
	}
	if options.JitterFraction > 0.5 {
		options.JitterFraction = 0.5
	}
	if options.Random == nil {
		options.Random = rand.Float64
	}
	return &Scheduler{collectors: append([]Collector(nil), collectors...), store: store, options: options}
}

func (scheduler *Scheduler) Poll(ctx context.Context) error {
	semaphore := make(chan struct{}, scheduler.options.MaxConcurrent)
	errorsByCollector := make(chan error, len(scheduler.collectors))
	var group sync.WaitGroup
	for _, current := range scheduler.collectors {
		current := current
		group.Add(1)
		go func() {
			defer group.Done()
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				errorsByCollector <- ctx.Err()
				return
			}

			started := time.Now()
			result, err := current.Collect(ctx)
			completed := time.Now()
			duration := completed.Sub(started)
			if err != nil {
				class, access := snapshot.ErrorUnknown, model.AccessUnreachable
				var collectionErr *CollectionError
				if errors.As(err, &collectionErr) {
					class, access = collectionErr.Class, collectionErr.Access
				}
				scheduler.store.PublishFailure(current.SourceID(), completed.UTC(), duration, class, access)
				errorsByCollector <- err
				return
			}
			collectedAt := result.CollectedAt
			if collectedAt.IsZero() {
				collectedAt = completed.UTC()
			}
			scheduler.store.PublishSuccess(current.SourceID(), collectedAt, duration, result)
		}()
	}
	group.Wait()
	close(errorsByCollector)
	var collected []error
	for err := range errorsByCollector {
		collected = append(collected, err)
	}
	return errors.Join(collected...)
}

func (scheduler *Scheduler) Run(ctx context.Context) error {
	for {
		_ = scheduler.Poll(ctx)
		timer := time.NewTimer(scheduler.NextDelay())
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func (scheduler *Scheduler) NextDelay() time.Duration {
	span := float64(scheduler.options.Interval) * scheduler.options.JitterFraction
	offset := (scheduler.options.Random()*2 - 1) * span
	return scheduler.options.Interval + time.Duration(offset)
}
