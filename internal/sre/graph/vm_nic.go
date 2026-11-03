package graph

import "github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"

func resolveVMNIC(graph *Graph) {
	for _, vm := range graph.byKind("IronCoreVM") {
		for _, ref := range vm.References {
			if ref.Kind != "Nic" {
				continue
			}
			matches, evidence := resolveSameSourceRef(graph, vm.Provenance.SourceID, "Nic", vm.Provenance.Namespace, ref)
			for _, nic := range matches {
				graph.addRelation(model.Relation{From: vm.Key, To: nic.Key, Type: "has-nic", Evidence: evidence, FieldPath: ref.FieldPath, Confidence: model.ConfidenceProven})
			}
		}
	}
	for _, nic := range graph.byKind("Nic") {
		for _, ref := range nic.References {
			switch ref.Kind {
			case "IronCoreVM":
				matches, evidence := resolveSameSourceRef(graph, nic.Provenance.SourceID, "IronCoreVM", nic.Provenance.Namespace, ref)
				for _, vm := range matches {
					graph.addRelation(model.Relation{From: vm.Key, To: nic.Key, Type: "has-nic", Evidence: evidence, FieldPath: ref.FieldPath, Confidence: model.ConfidenceProven})
				}
			case "Network":
				for _, network := range graph.sameSourceName(nic.Provenance.SourceID, "Network", nic.Provenance.Namespace, ref.Name) {
					graph.addRelation(model.Relation{From: nic.Key, To: network.Key, Type: "connected-to-network", Evidence: model.EvidenceObjectReference, FieldPath: ref.FieldPath, Confidence: model.ConfidenceProven})
				}
			}
		}
	}
}

func resolveSameSourceRef(graph *Graph, source, kind, namespace string, ref model.ObjectReference) ([]model.Resource, model.EvidenceKind) {
	if ref.UID != "" {
		matches := graph.sameSourceUID(source, kind, ref.UID)
		for _, match := range matches {
			if ref.Name != "" && match.Provenance.Name != ref.Name {
				graph.addObservation("uid-name-disagreement", "reference UID and name resolve to different identities", match.Key)
			}
		}
		return matches, model.EvidenceUIDReference
	}
	return graph.sameSourceName(source, kind, namespace, ref.Name), model.EvidenceObjectReference
}
