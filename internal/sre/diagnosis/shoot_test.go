package diagnosis

import (
	"testing"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/graph"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
)

func TestDiagnoseShootCountsOnlyEvidenceBackedWorkers(t *testing.T) {
	t.Parallel()
	shoot := diagnosisResource(t, "gardener", "Shoot", "shoot", "garden-team", "alpha", "shoot-uid")
	shoot.Aliases = []string{"shoot--team--alpha"}
	vm := diagnosisResource(t, "central", "IronCoreVM", "ironcore-vm", "tenant", "vm-a", "vm-uid")
	vm.Labels = map[string]string{"kubernetes.io/cluster": "shoot--team--alpha"}

	got, err := DiagnoseShoot(graph.Build([]model.Resource{shoot, vm}), nil, shoot.Key)
	if err != nil {
		t.Fatal(err)
	}
	if got.WorkerVMs != 1 || got.RelationConfidence != model.ConfidenceCorroborated {
		t.Fatalf("summary = %#v", got)
	}
}
