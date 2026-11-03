package graph

import (
	"strings"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
)

func resolveMCM(graph *Graph) {
	for _, deployment := range graph.byKind("McmMachineDeployment") {
		selector, err := selectorFor(deployment)
		if err == nil && selector != nil {
			for _, set := range graph.byKind("McmMachineSet") {
				if set.Provenance.SourceID == deployment.Provenance.SourceID && set.Provenance.Namespace == deployment.Provenance.Namespace && selector.Matches(labels.Set(set.Labels)) {
					graph.addRelation(model.Relation{From: deployment.Key, To: set.Key, Type: "owns-machine-set", Evidence: model.EvidenceSelectorMatch, FieldPath: ".spec.selector", Confidence: model.ConfidenceProven})
				}
			}
		}
	}
	for _, set := range graph.byKind("McmMachineSet") {
		for _, owner := range set.OwnerRefs {
			if owner.Kind != "McmMachineDeployment" && owner.Kind != "MachineDeployment" {
				continue
			}
			matches, _ := resolveSameSourceRef(graph, set.Provenance.SourceID, "McmMachineDeployment", set.Provenance.Namespace, owner)
			for _, deployment := range matches {
				graph.addRelation(model.Relation{From: deployment.Key, To: set.Key, Type: "owns-machine-set", Evidence: model.EvidenceOwnerReference, FieldPath: owner.FieldPath, Confidence: model.ConfidenceProven})
			}
		}
	}
	for _, machine := range graph.byKind("McmMachine") {
		for _, owner := range machine.OwnerRefs {
			if owner.Kind != "McmMachineSet" && owner.Kind != "MachineSet" {
				continue
			}
			matches, _ := resolveSameSourceRef(graph, machine.Provenance.SourceID, "McmMachineSet", machine.Provenance.Namespace, owner)
			for _, set := range matches {
				graph.addRelation(model.Relation{From: set.Key, To: machine.Key, Type: "owns-mcm-machine", Evidence: model.EvidenceOwnerReference, FieldPath: owner.FieldPath, Confidence: model.ConfidenceProven})
			}
		}
		if nodeName := machine.Labels["node"]; nodeName != "" {
			for _, node := range graph.sameSourceName(machine.Provenance.SourceID, "ShootNode", "", nodeName) {
				graph.addRelation(model.Relation{From: machine.Key, To: node.Key, Type: "joined-as-shoot-node", Evidence: model.EvidenceExactLabel, FieldPath: ".metadata.labels[node]", Confidence: model.ConfidenceProven})
			}
		}
		if strings.HasPrefix(machine.ProviderID, "ironcore-metal://") {
			serverName := strings.TrimPrefix(machine.ProviderID, "ironcore-metal://")
			var matches []model.Resource
			for _, server := range graph.byKind("PhysicalNode") {
				if server.Type == "physical-server" && server.Provenance.Name == serverName {
					matches = append(matches, server)
					graph.addRelation(model.Relation{From: machine.Key, To: server.Key, Type: "provisioned-on-physical-server", Evidence: model.EvidenceProviderID, FieldPath: ".spec.providerID", Confidence: model.ConfidenceProven})
				}
			}
			if len(matches) > 1 {
				graph.addObservation("physical-server-conflict", "ironcore-metal provider ID matched multiple physical servers", machine.Key)
			}
		}
	}
}
