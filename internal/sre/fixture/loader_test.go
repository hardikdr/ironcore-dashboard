package fixture

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
)

func TestLoadDirPublishesSanitizedOfflineLandscape(t *testing.T) {
	t.Parallel()
	store, err := LoadDir(filepath.Join("..", "testdata"))
	if err != nil {
		t.Fatal(err)
	}
	aggregate := store.Aggregate(time.Date(2026, 11, 3, 12, 5, 0, 0, time.UTC), 2*time.Minute)
	if len(aggregate.Sources) < 10 {
		t.Fatalf("fixture sources = %d, want at least 10", len(aggregate.Sources))
	}
	var resources int
	states := map[model.AccessState]bool{}
	for _, source := range aggregate.Sources {
		resources += len(source.Snapshot.Resources)
		states[source.Status.Access] = true
	}
	if resources < 10 {
		t.Fatalf("fixture resources = %d", resources)
	}
	for _, state := range []model.AccessState{model.AccessOK, model.AccessDenied, model.AccessUnreachable, model.AccessNotConfigured} {
		if !states[state] {
			t.Fatalf("fixtures omit access state %q", state)
		}
	}
}
