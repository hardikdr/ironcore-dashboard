package acceptance

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/diagnosis"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/fixture"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/graph"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/search"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/snapshot"
)

func TestNineFixtureBackedAcceptanceScenarios(t *testing.T) {
	store, err := fixture.LoadDir(filepath.Join("..", "testdata"))
	if err != nil {
		t.Fatal(err)
	}
	aggregate := store.Aggregate(time.Date(2026, 11, 3, 12, 5, 0, 0, time.UTC), 2*time.Minute)
	var resources []model.Resource
	sources := map[string]snapshot.SourceView{}
	for _, source := range aggregate.Sources {
		resources = append(resources, source.Snapshot.Resources...)
		sources[source.SourceID] = source
	}
	topology := graph.Build(resources)
	vm := find(t, resources, "IronCoreVM", "worker-a-0")
	pool := find(t, resources, "MachinePool", "cn-wdf4a-7")
	node := find(t, resources, "PhysicalNode", "cn-wdf4a-7")
	shoot := find(t, resources, "Shoot", "fecode")
	lb := find(t, resources, "LoadBalancer", "api")
	nic := find(t, resources, "Nic", "worker-a-0-primary")
	volume := find(t, resources, "Volume", "worker-a-0-root")
	volumePool := find(t, resources, "VolumePool", "wdf4a")
	metal := find(t, resources, "McmMachine", "wdf4ai-worker-0")
	t.Run("1 VM-backed Shoot spans the physical path", func(t *testing.T) {
		require(t, topology, vm.Key, pool.Key, "scheduled-on")
		require(t, topology, pool.Key, node.Key, "registered-by")
		require(t, topology, vm.Key, shoot.Key, "owned-by-shoot")
	})
	t.Run("2 LB selector resolves NIC targets", func(t *testing.T) { require(t, topology, lb.Key, nic.Key, "selects-nic") })
	t.Run("3 persistent Volume placement", func(t *testing.T) { require(t, topology, volume.Key, volumePool.Key, "placed-on-volume-pool") })
	t.Run("4 bare-metal wdf4ai has no fabricated VM", func(t *testing.T) {
		for _, edge := range topology.Outgoing(metal.Key) {
			target, _ := topology.Resource(edge.To)
			if target.Provenance.Kind == "IronCoreVM" {
				t.Fatalf("bare-metal edge fabricated: %#v", edge)
			}
		}
	})
	t.Run("5 D and E are management-only not configured", func(t *testing.T) {
		for _, id := range []string{"wdf4d", "wdf4e"} {
			if sources[id].Status.Access != model.AccessNotConfigured {
				t.Fatalf("%s access = %s", id, sources[id].Status.Access)
			}
		}
	})
	t.Run("6 Gardener RBAC denial remains explicit", func(t *testing.T) {
		if sources["ironcore-consumer-gardener"].Status.Access != model.AccessDenied {
			t.Fatalf("consumer access = %s", sources["ironcore-consumer-gardener"].Status.Access)
		}
	})
	t.Run("7 stale Canary source remains stale", func(t *testing.T) {
		if sources["wdf4c"].Status.Freshness != model.FreshnessStale || sources["wdf4c"].Status.Access != model.AccessUnreachable {
			t.Fatalf("wdf4c = %#v", sources["wdf4c"].Status)
		}
	})
	t.Run("8 provider degradation does not fabricate user impact", func(t *testing.T) {
		summary, err := diagnosis.DiagnoseVM(topology, sources, vm.Key)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(strings.ToLower(summary.Localization), "provider failed") {
			t.Fatalf("unproven impact: %s", summary.Localization)
		}
	})
	t.Run("9 exact VM UID search returns full path", func(t *testing.T) {
		results := search.New(resources).Search("vm-canary-001", search.Options{Limit: 10})
		if len(results) != 1 || results[0].Resource.Key != vm.Key {
			t.Fatalf("search = %#v", results)
		}
		path := topology.Traverse(vm.Key, 5, 100)
		if !contains(path.Resources, node.Key) || !contains(path.Resources, shoot.Key) {
			t.Fatalf("full path incomplete")
		}
	})
}

func find(t *testing.T, resources []model.Resource, kind, name string) model.Resource {
	t.Helper()
	for _, resource := range resources {
		if resource.Provenance.Kind == kind && resource.Provenance.Name == name {
			return resource
		}
	}
	t.Fatalf("missing %s %s", kind, name)
	return model.Resource{}
}
func require(t *testing.T, topology *graph.Graph, from, to model.ResourceKey, kind string) {
	t.Helper()
	for _, edge := range topology.Outgoing(from) {
		if edge.To == to && edge.Type == kind {
			return
		}
	}
	t.Fatalf("missing %s edge from %s to %s", kind, from, to)
}
func contains(resources []model.Resource, key model.ResourceKey) bool {
	for _, resource := range resources {
		if resource.Key == key {
			return true
		}
	}
	return false
}
