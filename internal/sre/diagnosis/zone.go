package diagnosis

import (
	"fmt"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/graph"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/snapshot"
)

type ZoneSummary struct {
	Zone            model.Resource    `json:"zone"`
	State           model.AccessState `json:"state"`
	ManagementOnly  bool              `json:"managementOnly"`
	ManagementNodes int               `json:"managementNodes"`
	ComputeNodes    int               `json:"computeNodes"`
	StorageNodes    int               `json:"storageNodes"`
	MachinePools    int               `json:"machinePools"`
	VolumePools     int               `json:"volumePools"`
	HostedVMs       int               `json:"hostedVms"`
	DependentShoots int               `json:"dependentShoots"`
	Facts           []Fact            `json:"facts"`
	Gaps            []Gap             `json:"gaps"`
}

func DiagnoseZone(topology *graph.Graph, sources map[string]snapshot.SourceView, key model.ResourceKey) (ZoneSummary, error) {
	zone, ok := topology.Resource(key)
	if !ok || zone.Provenance.Kind != "Zone" {
		return ZoneSummary{}, fmt.Errorf("zone not found")
	}
	summary := ZoneSummary{Zone: zone, State: model.AccessOK}
	if role, _ := zone.Sanitized["capacityRole"].(string); role == "management-only" {
		summary.ManagementOnly, summary.State = true, model.AccessNotConfigured
	}
	zoneName := zone.Provenance.Name
	for _, resource := range topology.AllResources() {
		if resource.Provenance.Kind == "PhysicalNode" && resource.Provenance.SourceID == zoneName {
			role, _ := resource.Sanitized["physicalRole"].(string)
			switch role {
			case "management":
				summary.ManagementNodes++
			case "compute":
				summary.ComputeNodes++
			case "storage":
				summary.StorageNodes++
			}
		}
		if resource.Labels["topology.ironcore.dev/zone"] == zoneName {
			switch resource.Provenance.Kind {
			case "MachinePool":
				summary.MachinePools++
			case "VolumePool":
				summary.VolumePools++
			}
		}
	}
	for _, poolRelation := range topology.Incoming(zone.Key) {
		pool, _ := topology.Resource(poolRelation.From)
		if pool.Provenance.Kind != "MachinePool" {
			continue
		}
		for _, vmRelation := range topology.Incoming(pool.Key) {
			if vmRelation.Type == "scheduled-on" {
				summary.HostedVMs++
				vm, _ := topology.Resource(vmRelation.From)
				for _, shootRelation := range topology.Outgoing(vm.Key) {
					if shootRelation.Type == "owned-by-shoot" {
						summary.DependentShoots++
					}
				}
			}
		}
	}
	summary.Facts = []Fact{
		{Label: "Management nodes", Value: fmt.Sprintf("%d", summary.ManagementNodes)},
		{Label: "Compute nodes", Value: fmt.Sprintf("%d", summary.ComputeNodes)},
		{Label: "Storage nodes", Value: fmt.Sprintf("%d", summary.StorageNodes)},
		{Label: "Hosted VMs", Value: fmt.Sprintf("%d", summary.HostedVMs)},
	}
	for _, source := range sources {
		if source.SourceID == zoneName && (source.Status.Access != model.AccessOK || source.Status.Freshness != model.FreshnessFresh) {
			summary.Gaps = append(summary.Gaps, Gap{Area: source.SourceID, State: string(source.Status.Access), Message: "zone inventory is incomplete or not current"})
		}
	}
	return summary, nil
}
