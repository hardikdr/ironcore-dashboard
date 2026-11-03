package graph

import (
	"testing"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
)

func TestResolveMCMHierarchyAndBareMetalSplit(t *testing.T) {
	t.Parallel()
	deployment := testResource(t, "gardener", "McmMachineDeployment", "mcm-machine-deployment", "shoot-ns", "workers", "deployment-uid")
	deployment.Selector = map[string]string{"deployment": "workers"}
	set := testResource(t, "gardener", "McmMachineSet", "mcm-machine-set", "shoot-ns", "workers-abc", "set-uid")
	set.Labels = map[string]string{"deployment": "workers"}
	machine := testResource(t, "gardener", "McmMachine", "mcm-machine", "shoot-ns", "worker-0", "machine-uid")
	machine.OwnerRefs = []model.ObjectReference{{Kind: "McmMachineSet", Namespace: "shoot-ns", Name: "workers-abc", UID: "set-uid", FieldPath: ".metadata.ownerReferences"}}
	machine.ProviderID = "ironcore-metal://server-ai-01"
	server := testResource(t, "wdf4a", "PhysicalNode", "physical-server", "", "server-ai-01", "server-uid")
	vm := testResource(t, "central", "IronCoreVM", "ironcore-vm", "shoot-ns", "worker-0", "vm-uid")

	graph := Build([]model.Resource{deployment, set, machine, server, vm})
	requireEdge(t, graph.Outgoing(deployment.Key), deployment.Key, set.Key, "owns-machine-set", model.EvidenceSelectorMatch)
	requireEdge(t, graph.Outgoing(set.Key), set.Key, machine.Key, "owns-mcm-machine", model.EvidenceOwnerReference)
	requireEdge(t, graph.Outgoing(machine.Key), machine.Key, server.Key, "provisioned-on-physical-server", model.EvidenceProviderID)
	for _, edge := range graph.Outgoing(machine.Key) {
		if edge.To == vm.Key {
			t.Fatalf("bare-metal MCM machine linked to IronCore VM: %#v", edge)
		}
	}
}
