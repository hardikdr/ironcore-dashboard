package graph

import (
	"fmt"
	"testing"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
)

func testResource(t *testing.T, source, kind, resourceType, namespace, name, uid string) model.Resource {
	t.Helper()
	group := groupForKind(kind)
	key, err := model.NewResourceKey(model.ResourceIdentity{SourceID: source, Group: group, Kind: kind, Namespace: namespace, Name: name, UID: uid})
	if err != nil {
		t.Fatal(err)
	}
	return model.Resource{Key: key, Type: resourceType, Provenance: model.Provenance{SourceID: source, Group: group, Kind: kind, Namespace: namespace, Name: name, UID: uid}}
}

func groupForKind(kind string) string {
	switch kind {
	case "IronCoreVM", "MachinePool", "MachineClass":
		return "compute.ironcore.dev"
	case "Nic", "Network", "LoadBalancer", "LoadBalancerRouting":
		return "networking.ironcore.dev"
	case "Volume", "VolumePool":
		return "storage.ironcore.dev"
	case "Shoot", "WorkerGroup":
		return "core.gardener.cloud"
	case "McmMachine", "McmMachineSet", "McmMachineDeployment":
		return "machine.sapcloud.io"
	default:
		return "core"
	}
}

func requireEdge(t *testing.T, relations []model.Relation, from, to model.ResourceKey, relationType string, evidence model.EvidenceKind) {
	t.Helper()
	for _, relation := range relations {
		if relation.From == from && relation.To == to && relation.Type == relationType && relation.Evidence == evidence {
			return
		}
	}
	t.Fatalf("missing %s edge %s -> %s (%s) in %#v", relationType, from, to, evidence, relations)
}

func TestGraphStoresSymmetricIndexesAndCapsTraversal(t *testing.T) {
	t.Parallel()

	resources := make([]model.Resource, 0, 20)
	for index := 0; index < 20; index++ {
		resource := testResource(t, "source", "Network", "network", "tenant", fmt.Sprintf("n-%02d", index), fmt.Sprintf("uid-%02d", index))
		resources = append(resources, resource)
	}
	graph := New(resources)
	for index := 1; index < len(resources); index++ {
		graph.addRelation(model.Relation{From: resources[index-1].Key, To: resources[index].Key, Type: "test", Evidence: model.EvidenceObjectReference, Confidence: model.ConfidenceProven})
	}
	requireEdge(t, graph.Outgoing(resources[0].Key), resources[0].Key, resources[1].Key, "test", model.EvidenceObjectReference)
	requireEdge(t, graph.Incoming(resources[1].Key), resources[0].Key, resources[1].Key, "test", model.EvidenceObjectReference)

	got := graph.Traverse(resources[0].Key, 99, 3)
	if len(got.Resources) != 3 || !got.Truncated {
		t.Fatalf("traversal = resources %d truncated %v", len(got.Resources), got.Truncated)
	}
}
