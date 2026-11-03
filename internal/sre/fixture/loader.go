// Package fixture loads recorded, sanitized JSON snapshots for local-only demos.
package fixture

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/snapshot"
)

const maxFixtureBytes = 16 << 20

type sourceRecord struct {
	ID        string            `json:"id"`
	Access    model.AccessState `json:"access"`
	Freshness model.Freshness   `json:"freshness"`
}

type envelope struct {
	RecordedAt time.Time        `json:"recordedAt"`
	Source     *sourceRecord    `json:"source"`
	Sources    []sourceRecord   `json:"sources"`
	Resources  []model.Resource `json:"resources"`
}

// LoadDir performs no network I/O. It accepts only regular .json files and
// publishes their sanitized resources into the same immutable store used live.
func LoadDir(directory string) (*snapshot.Store, error) {
	info, err := os.Stat(directory)
	if err != nil {
		return nil, fmt.Errorf("fixture directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("fixture path is not a directory")
	}
	declared := map[string]sourceRecord{}
	resources := map[string][]model.Resource{}
	collected := map[string]time.Time{}
	err = filepath.WalkDir(directory, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("fixture %q is not a regular file", path)
		}
		if info.Size() > maxFixtureBytes {
			return fmt.Errorf("fixture %q exceeds size cap", path)
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var current envelope
		if err := json.Unmarshal(body, &current); err != nil {
			return fmt.Errorf("decode fixture %q: %w", path, err)
		}
		if current.RecordedAt.IsZero() {
			return fmt.Errorf("fixture %q has no recordedAt", path)
		}
		if current.Source != nil {
			declared[current.Source.ID] = *current.Source
		}
		for _, source := range current.Sources {
			declared[source.ID] = source
		}
		for _, resource := range current.Resources {
			sourceID := resource.Provenance.SourceID
			if sourceID == "" {
				return fmt.Errorf("fixture %q contains resource without source ID", path)
			}
			resources[sourceID] = append(resources[sourceID], resource)
			if _, ok := declared[sourceID]; !ok {
				declared[sourceID] = sourceRecord{ID: sourceID, Access: model.AccessOK, Freshness: model.FreshnessFresh}
			}
			collected[sourceID] = current.RecordedAt
		}
		for id := range declared {
			if collected[id].IsZero() {
				collected[id] = current.RecordedAt
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(declared))
	for id := range declared {
		if id == "" {
			return nil, fmt.Errorf("fixture contains empty source ID")
		}
		ids = append(ids, id)
	}
	sort.Strings(ids)
	store := snapshot.NewStore()
	for _, id := range ids {
		record := declared[id]
		at := collected[id]
		counts := map[string]int{}
		for _, resource := range resources[id] {
			counts[resource.Provenance.Kind]++
		}
		store.PublishSuccess(id, at, 0, snapshot.Snapshot{SourceID: id, CollectedAt: at, Resources: resources[id], ObjectCounts: counts})
		if record.Access != model.AccessOK {
			store.PublishFailure(id, at, 0, errorClass(record.Access), record.Access)
		}
	}
	return store, nil
}

func errorClass(access model.AccessState) snapshot.ErrorClass {
	switch access {
	case model.AccessDenied:
		return snapshot.ErrorRBAC
	case model.AccessUnreachable:
		return snapshot.ErrorTransport
	case model.AccessUnsupported, model.AccessNotConfigured:
		return snapshot.ErrorUnsupported
	default:
		return snapshot.ErrorUnknown
	}
}
