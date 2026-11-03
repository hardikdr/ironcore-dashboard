package graph

import (
	"testing"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
)

func TestResolveVMShootRetainsConflictingExactLabels(t *testing.T) {
	t.Parallel()
	vm := testResource(t, "central", "IronCoreVM", "ironcore-vm", "tenant", "vm-a", "vm-uid")
	vm.Labels = map[string]string{"kubernetes.io/cluster": "shoot--team--alpha"}
	shootA := testResource(t, "gardener-a", "Shoot", "shoot", "garden-team", "alpha", "shoot-a")
	shootB := testResource(t, "gardener-b", "Shoot", "shoot", "garden-team", "alpha", "shoot-b")
	shootA.Aliases, shootB.Aliases = []string{"shoot--team--alpha"}, []string{"shoot--team--alpha"}

	graph := Build([]model.Resource{vm, shootA, shootB})
	requireEdge(t, graph.Outgoing(vm.Key), vm.Key, shootA.Key, "owned-by-shoot", model.EvidenceExactLabel)
	requireEdge(t, graph.Outgoing(vm.Key), vm.Key, shootB.Key, "owned-by-shoot", model.EvidenceExactLabel)
	if len(graph.Observations()) == 0 {
		t.Fatal("ambiguous exact Shoot label produced no conflict observation")
	}
}
