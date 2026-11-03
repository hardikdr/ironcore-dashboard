package graph

import (
	"testing"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
)

func TestResolvePoolNodeUsesTopologyZoneAndProviderID(t *testing.T) {
	t.Parallel()
	pool := testResource(t, "central", "MachinePool", "machine-pool", "", "pool-a", "pool-uid")
	pool.Labels = map[string]string{"topology.ironcore.dev/zone": "wdf4a"}
	pool.ProviderID = "libvirt-provider://node-a"
	zone := testResource(t, "catalog", "Zone", "zone", "", "wdf4a", "zone-a")
	node := testResource(t, "wdf4a", "PhysicalNode", "physical-node", "", "node-a", "node-a-uid")
	wrongZoneNode := testResource(t, "wdf4b", "PhysicalNode", "physical-node", "", "node-a", "node-b-uid")

	graph := Build([]model.Resource{pool, zone, node, wrongZoneNode})
	requireEdge(t, graph.Outgoing(pool.Key), pool.Key, zone.Key, "located-in-zone", model.EvidenceExactLabel)
	requireEdge(t, graph.Outgoing(pool.Key), pool.Key, node.Key, "registered-by", model.EvidenceProviderID)
	for _, edge := range graph.Outgoing(pool.Key) {
		if edge.To == wrongZoneNode.Key {
			t.Fatalf("pool joined same-named node in wrong zone: %#v", edge)
		}
	}
}
