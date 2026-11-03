package graph

import (
	"testing"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
)

func TestResolveVolumePlacementFollowsExplicitReferencesAndZone(t *testing.T) {
	t.Parallel()
	vm := testResource(t, "central", "IronCoreVM", "ironcore-vm", "tenant", "vm-a", "vm-uid")
	volume := testResource(t, "central", "Volume", "volume", "tenant", "vol-a", "vol-uid")
	pool := testResource(t, "central", "VolumePool", "volume-pool", "", "wdf4b", "pool-uid")
	zone := testResource(t, "catalog", "Zone", "zone", "", "wdf4b", "zone-uid")
	vm.References = []model.ObjectReference{{Kind: "Volume", Namespace: "tenant", Name: "vol-a", UID: "vol-uid", FieldPath: ".status.volumes"}}
	volume.References = []model.ObjectReference{{Kind: "VolumePool", Name: "wdf4b", FieldPath: ".spec.volumePoolRef.name"}}
	pool.Labels = map[string]string{"topology.ironcore.dev/zone": "wdf4b"}

	graph := Build([]model.Resource{vm, volume, pool, zone})
	requireEdge(t, graph.Outgoing(vm.Key), vm.Key, volume.Key, "has-volume", model.EvidenceUIDReference)
	requireEdge(t, graph.Outgoing(volume.Key), volume.Key, pool.Key, "placed-on-volume-pool", model.EvidenceObjectReference)
	requireEdge(t, graph.Outgoing(pool.Key), pool.Key, zone.Key, "located-in-zone", model.EvidenceExactLabel)
}
