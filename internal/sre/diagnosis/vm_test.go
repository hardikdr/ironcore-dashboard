package diagnosis

import (
	"strings"
	"testing"
	"time"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/graph"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/snapshot"
)

func TestDiagnoseVMReportsFactsAndExplicitUnknowns(t *testing.T) {
	t.Parallel()
	vm := diagnosisResource(t, "central", "IronCoreVM", "ironcore-vm", "tenant", "vm-a", "vm-uid")
	vm.State, vm.Health = "Running", model.HealthHealthy
	sources := map[string]snapshot.SourceView{
		"central":  {SourceID: "central", Status: snapshot.SourceStatus{Access: model.AccessOK, Freshness: model.FreshnessFresh}},
		"gardener": {SourceID: "gardener", Status: snapshot.SourceStatus{Access: model.AccessDenied, Freshness: model.FreshnessUnknown}},
	}

	got, err := DiagnoseVM(graph.Build([]model.Resource{vm}), sources, vm.Key)
	if err != nil {
		t.Fatal(err)
	}
	if !hasFact(got.Facts, "VM state", "Running") || !hasFact(got.Facts, "MachinePool", "unknown") || !hasFact(got.Facts, "Shoot ownership", "unknown") {
		t.Fatalf("facts = %#v", got.Facts)
	}
	if !strings.Contains(got.Localization, "VM Running") || !strings.Contains(got.Localization, "Shoot ownership unknown") {
		t.Fatalf("localization = %q", got.Localization)
	}
	if len(got.Gaps) == 0 || got.Gaps[0].State != string(model.AccessDenied) {
		t.Fatalf("gaps = %#v", got.Gaps)
	}
}

func TestDiagnoseVMBuildsEvidenceBackedPath(t *testing.T) {
	t.Parallel()
	vm := diagnosisResource(t, "central", "IronCoreVM", "ironcore-vm", "tenant", "vm-a", "vm-uid")
	vm.State = "Running"
	pool := diagnosisResource(t, "central", "MachinePool", "machine-pool", "", "pool-a", "pool-uid")
	node := diagnosisResource(t, "zone-a", "PhysicalNode", "physical-node", "", "node-a", "node-uid")
	shoot := diagnosisResource(t, "gardener", "Shoot", "shoot", "garden-team", "alpha", "shoot-uid")
	nic := diagnosisResource(t, "central", "Nic", "nic", "tenant", "nic-a", "nic-uid")
	volume := diagnosisResource(t, "central", "Volume", "volume", "tenant", "vol-a", "vol-uid")
	vm.References = []model.ObjectReference{{Kind: "MachinePool", Name: "pool-a"}, {Kind: "Nic", Namespace: "tenant", Name: "nic-a", UID: "nic-uid"}, {Kind: "Volume", Namespace: "tenant", Name: "vol-a", UID: "vol-uid"}}
	vm.Labels = map[string]string{"kubernetes.io/cluster": "shoot--team--alpha"}
	pool.ProviderID, pool.Labels = "libvirt-provider://node-a", map[string]string{"topology.ironcore.dev/zone": "zone-a"}
	shoot.Aliases = []string{"shoot--team--alpha"}

	got, err := DiagnoseVM(graph.Build([]model.Resource{vm, pool, node, shoot, nic, volume}), nil, vm.Key)
	if err != nil {
		t.Fatal(err)
	}
	for _, label := range []string{"MachinePool", "Physical host", "Shoot ownership", "NICs", "Volumes"} {
		if !hasFactLabel(got.Facts, label) {
			t.Fatalf("facts %#v missing %q", got.Facts, label)
		}
	}
}

func diagnosisResource(t *testing.T, source, kind, resourceType, namespace, name, uid string) model.Resource {
	t.Helper()
	key, err := model.NewResourceKey(model.ResourceIdentity{SourceID: source, Group: "test.example", Kind: kind, Namespace: namespace, Name: name, UID: uid})
	if err != nil {
		t.Fatal(err)
	}
	return model.Resource{Key: key, Type: resourceType, Provenance: model.Provenance{SourceID: source, Group: "test.example", Kind: kind, Namespace: namespace, Name: name, UID: uid, CollectedAt: time.Now()}}
}

func hasFact(facts []Fact, label, value string) bool {
	for _, fact := range facts {
		if fact.Label == label && fact.Value == value {
			return true
		}
	}
	return false
}

func hasFactLabel(facts []Fact, label string) bool {
	for _, fact := range facts {
		if fact.Label == label && fact.Value != "unknown" {
			return true
		}
	}
	return false
}
