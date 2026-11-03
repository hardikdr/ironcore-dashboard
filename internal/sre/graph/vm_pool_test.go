package graph

import (
	"testing"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
)

func TestResolveVMPoolUsesSameSourceObjectReference(t *testing.T) {
	t.Parallel()
	vm := testResource(t, "central", "IronCoreVM", "ironcore-vm", "tenant", "vm-a", "vm-uid")
	pool := testResource(t, "central", "MachinePool", "machine-pool", "", "pool-a", "pool-uid")
	other := testResource(t, "other", "MachinePool", "machine-pool", "", "pool-a", "other-pool")
	vm.References = []model.ObjectReference{{Group: "compute.ironcore.dev", Kind: "MachinePool", Name: "pool-a", FieldPath: ".spec.machinePoolRef.name"}}

	graph := Build([]model.Resource{vm, pool, other})
	requireEdge(t, graph.Outgoing(vm.Key), vm.Key, pool.Key, "scheduled-on", model.EvidenceObjectReference)
	if len(graph.Outgoing(vm.Key)) != 1 {
		t.Fatalf("VM edges crossed source boundary: %#v", graph.Outgoing(vm.Key))
	}
}
