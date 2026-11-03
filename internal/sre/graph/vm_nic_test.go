package graph

import (
	"testing"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
)

func TestResolveVMNICPrefersUIDAndConnectsNICNetwork(t *testing.T) {
	t.Parallel()
	vm := testResource(t, "central", "IronCoreVM", "ironcore-vm", "tenant", "vm-a", "vm-uid")
	nic := testResource(t, "central", "Nic", "nic", "tenant", "nic-a", "nic-uid")
	network := testResource(t, "central", "Network", "network", "tenant", "net-a", "net-uid")
	vm.References = []model.ObjectReference{{Kind: "Nic", Namespace: "tenant", Name: "nic-a", UID: "nic-uid", FieldPath: ".status.networkInterfaces"}}
	nic.References = []model.ObjectReference{
		{Kind: "IronCoreVM", Namespace: "tenant", Name: "wrong-name", UID: "vm-uid", FieldPath: ".spec.machineRef"},
		{Kind: "Network", Namespace: "tenant", Name: "net-a", FieldPath: ".spec.networkRef.name"},
	}

	graph := Build([]model.Resource{vm, nic, network})
	requireEdge(t, graph.Outgoing(vm.Key), vm.Key, nic.Key, "has-nic", model.EvidenceUIDReference)
	requireEdge(t, graph.Outgoing(nic.Key), nic.Key, network.Key, "connected-to-network", model.EvidenceObjectReference)
	if len(graph.Observations()) == 0 {
		t.Fatal("UID/name disagreement produced no observation")
	}
}
