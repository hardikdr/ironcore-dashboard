package diagnosis

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/graph"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/snapshot"
)

type Fact struct {
	Label      string             `json:"label"`
	Value      string             `json:"value"`
	SourceID   string             `json:"sourceId,omitempty"`
	Resource   model.ResourceKey  `json:"resource,omitempty"`
	Evidence   model.EvidenceKind `json:"evidence,omitempty"`
	Confidence model.Confidence   `json:"confidence,omitempty"`
}

type Gap struct {
	Area    string `json:"area"`
	State   string `json:"state"`
	Message string `json:"message"`
}

type VMSummary struct {
	VM           model.Resource `json:"vm"`
	Localization string         `json:"localization"`
	Facts        []Fact         `json:"facts"`
	Gaps         []Gap          `json:"gaps"`
}

func DiagnoseVM(topology *graph.Graph, sources map[string]snapshot.SourceView, key model.ResourceKey) (VMSummary, error) {
	vm, ok := topology.Resource(key)
	if !ok || vm.Provenance.Kind != "IronCoreVM" {
		return VMSummary{}, fmt.Errorf("IronCore VM not found")
	}
	summary := VMSummary{VM: vm}
	state := vm.State
	if state == "" {
		state = "unknown"
	}
	summary.Facts = append(summary.Facts, Fact{Label: "VM state", Value: state, SourceID: vm.Provenance.SourceID, Resource: vm.Key})
	if source, found := sources[vm.Provenance.SourceID]; found {
		summary.Facts = append(summary.Facts,
			Fact{Label: "Source access", Value: string(source.Status.Access), SourceID: source.SourceID},
			Fact{Label: "Source freshness", Value: string(source.Status.Freshness), SourceID: source.SourceID})
	}

	pool, poolEdge := firstOutgoing(topology, vm.Key, "scheduled-on")
	appendRelatedFact(&summary.Facts, "MachinePool", pool, poolEdge)
	var node model.Resource
	var nodeEdge model.Relation
	if pool.Key != "" {
		node, nodeEdge = firstOutgoing(topology, pool.Key, "registered-by")
	}
	appendRelatedFact(&summary.Facts, "Physical host", node, nodeEdge)
	shoot, shootEdge := firstOutgoing(topology, vm.Key, "owned-by-shoot")
	appendRelatedFact(&summary.Facts, "Shoot ownership", shoot, shootEdge)
	summary.Facts = append(summary.Facts,
		countFact(topology, vm.Key, "has-nic", "NICs"),
		countFact(topology, vm.Key, "has-volume", "Volumes"),
	)
	providerCount := 0
	if node.Key != "" {
		for _, resource := range topology.AllResources() {
			if resource.Type == "provider-component" && resource.Provenance.SourceID == node.Provenance.SourceID {
				providerCount++
			}
		}
	}
	if providerCount == 0 {
		summary.Facts = append(summary.Facts, Fact{Label: "Provider health", Value: "unknown"})
	} else {
		summary.Facts = append(summary.Facts, Fact{Label: "Provider components", Value: fmt.Sprintf("%d observed", providerCount)})
	}
	for _, source := range sources {
		if source.Status.Access != model.AccessOK || source.Status.Freshness != model.FreshnessFresh {
			state := string(source.Status.Access)
			if source.Status.Access == model.AccessOK {
				state = string(source.Status.Freshness)
			}
			summary.Gaps = append(summary.Gaps, Gap{Area: source.SourceID, State: state, Message: "source inventory is incomplete or not current"})
		}
	}
	sort.Slice(summary.Gaps, func(i, j int) bool { return summary.Gaps[i].Area < summary.Gaps[j].Area })
	parts := []string{"VM " + state}
	if shoot.Key == "" {
		parts = append(parts, "Shoot ownership unknown")
	} else {
		parts = append(parts, "Shoot relation "+string(shootEdge.Confidence))
	}
	if node.Key == "" {
		parts = append(parts, "physical host unknown")
	} else {
		parts = append(parts, "physical host "+node.Provenance.Name)
	}
	if len(summary.Gaps) != 0 {
		parts = append(parts, fmt.Sprintf("%d visibility gap(s)", len(summary.Gaps)))
	}
	summary.Localization = strings.Join(parts, "; ")
	return summary, nil
}

func firstOutgoing(topology *graph.Graph, key model.ResourceKey, relationType string) (model.Resource, model.Relation) {
	for _, relation := range topology.Outgoing(key) {
		if relation.Type == relationType {
			resource, _ := topology.Resource(relation.To)
			return resource, relation
		}
	}
	return model.Resource{}, model.Relation{}
}

func appendRelatedFact(facts *[]Fact, label string, resource model.Resource, relation model.Relation) {
	if resource.Key == "" {
		*facts = append(*facts, Fact{Label: label, Value: "unknown", Confidence: model.ConfidenceUnknown})
		return
	}
	*facts = append(*facts, Fact{Label: label, Value: resource.Provenance.Name, SourceID: resource.Provenance.SourceID, Resource: resource.Key, Evidence: relation.Evidence, Confidence: relation.Confidence})
}

func countFact(topology *graph.Graph, key model.ResourceKey, relationType, label string) Fact {
	count := 0
	for _, relation := range topology.Outgoing(key) {
		if relation.Type == relationType {
			count++
		}
	}
	if count == 0 {
		return Fact{Label: label, Value: "unknown", Confidence: model.ConfidenceUnknown}
	}
	return Fact{Label: label, Value: fmt.Sprintf("%d", count), Confidence: model.ConfidenceProven}
}
