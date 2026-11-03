package diagnosis

import (
	"testing"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/graph"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
)

func TestDiagnoseZoneDistinguishesManagementOnlyFromUnknown(t *testing.T) {
	t.Parallel()
	zone := diagnosisResource(t, "catalog", "Zone", "zone", "", "wdf4d", "zone-uid")
	zone.Sanitized = map[string]any{"capacityRole": "management-only"}
	node := diagnosisResource(t, "wdf4d", "PhysicalNode", "physical-node", "", "control-0", "node-uid")
	node.Sanitized = map[string]any{"physicalRole": "management"}

	got, err := DiagnoseZone(graph.Build([]model.Resource{zone, node}), nil, zone.Key)
	if err != nil {
		t.Fatal(err)
	}
	if !got.ManagementOnly || got.ManagementNodes != 1 || got.ComputeNodes != 0 || got.State != model.AccessNotConfigured {
		t.Fatalf("summary = %#v", got)
	}
}
