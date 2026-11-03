package diagnosis

import (
	"fmt"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/graph"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/snapshot"
)

type ShootSummary struct {
	Shoot              model.Resource   `json:"shoot"`
	WorkerVMs          int              `json:"workerVms"`
	WorkerGroups       int              `json:"workerGroups"`
	RelationConfidence model.Confidence `json:"relationConfidence"`
	Facts              []Fact           `json:"facts"`
	Gaps               []Gap            `json:"gaps"`
}

func DiagnoseShoot(topology *graph.Graph, sources map[string]snapshot.SourceView, key model.ResourceKey) (ShootSummary, error) {
	shoot, ok := topology.Resource(key)
	if !ok || shoot.Provenance.Kind != "Shoot" {
		return ShootSummary{}, fmt.Errorf("Shoot not found")
	}
	summary := ShootSummary{Shoot: shoot, RelationConfidence: model.ConfidenceUnknown}
	for _, relation := range topology.Incoming(shoot.Key) {
		if relation.Type == "owned-by-shoot" {
			summary.WorkerVMs++
			if relation.Confidence == model.ConfidenceProven || relation.Confidence == model.ConfidenceCorroborated {
				summary.RelationConfidence = relation.Confidence
			}
		}
	}
	for _, resource := range topology.AllResources() {
		if resource.Provenance.Kind == "WorkerGroup" && resource.Provenance.SourceID == shoot.Provenance.SourceID && resource.Provenance.Namespace == shoot.Provenance.Namespace {
			summary.WorkerGroups++
		}
	}
	summary.Facts = []Fact{
		{Label: "Worker VMs", Value: fmt.Sprintf("%d", summary.WorkerVMs), Confidence: summary.RelationConfidence},
		{Label: "Worker groups", Value: fmt.Sprintf("%d", summary.WorkerGroups), Confidence: model.ConfidenceProven},
	}
	for _, source := range sources {
		if source.Status.Access != model.AccessOK || source.Status.Freshness != model.FreshnessFresh {
			summary.Gaps = append(summary.Gaps, Gap{Area: source.SourceID, State: string(source.Status.Access), Message: "Shoot topology may be partial"})
		}
	}
	return summary, nil
}
