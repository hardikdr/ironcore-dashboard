package graph

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
)

func TestRecordedFixturesBuildGoldenPathAndReverseTraversal(t *testing.T) {
	t.Parallel()

	var resources []model.Resource
	err := filepath.WalkDir(filepath.Join("..", "testdata"), func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() || filepath.Ext(path) != ".json" {
			return walkErr
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var envelope struct {
			Resources []model.Resource `json:"resources"`
		}
		if err := json.Unmarshal(body, &envelope); err != nil {
			return err
		}
		resources = append(resources, envelope.Resources...)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	graph := Build(resources)
	vm := findFixture(t, resources, "IronCoreVM", "worker-a-0")
	pool := findFixture(t, resources, "MachinePool", "cn-wdf4a-7")
	node := findFixture(t, resources, "PhysicalNode", "cn-wdf4a-7")
	shoot := findFixture(t, resources, "Shoot", "fecode")
	nic := findFixture(t, resources, "Nic", "worker-a-0-primary")
	lb := findFixture(t, resources, "LoadBalancer", "api")
	volume := findFixture(t, resources, "Volume", "worker-a-0-root")
	volumePool := findFixture(t, resources, "VolumePool", "wdf4a")

	requireEdge(t, graph.Outgoing(vm.Key), vm.Key, pool.Key, "scheduled-on", model.EvidenceObjectReference)
	requireEdge(t, graph.Outgoing(pool.Key), pool.Key, node.Key, "registered-by", model.EvidenceProviderID)
	requireEdge(t, graph.Outgoing(vm.Key), vm.Key, shoot.Key, "owned-by-shoot", model.EvidenceExactLabel)
	requireEdge(t, graph.Outgoing(vm.Key), vm.Key, nic.Key, "has-nic", model.EvidenceUIDReference)
	requireEdge(t, graph.Outgoing(lb.Key), lb.Key, nic.Key, "selects-nic", model.EvidenceSelectorMatch)
	requireEdge(t, graph.Outgoing(vm.Key), vm.Key, volume.Key, "has-volume", model.EvidenceUIDReference)
	requireEdge(t, graph.Outgoing(volume.Key), volume.Key, volumePool.Key, "placed-on-volume-pool", model.EvidenceObjectReference)

	reverse := graph.Traverse(node.Key, 5, 100)
	if !containsResource(reverse.Resources, vm.Key) || !containsResource(reverse.Resources, shoot.Key) {
		t.Fatalf("reverse traversal from node omitted VM/Shoot: %#v", reverse.Resources)
	}
}

func findFixture(t *testing.T, resources []model.Resource, kind, name string) model.Resource {
	t.Helper()
	for _, resource := range resources {
		if resource.Provenance.Kind == kind && resource.Provenance.Name == name {
			return resource
		}
	}
	t.Fatalf("fixture missing %s %s", kind, name)
	return model.Resource{}
}

func containsResource(resources []model.Resource, key model.ResourceKey) bool {
	for _, resource := range resources {
		if resource.Key == key {
			return true
		}
	}
	return false
}
