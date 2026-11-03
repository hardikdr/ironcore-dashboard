package graph

import "github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"

func resolveVMPool(graph *Graph) {
	for _, vm := range graph.byKind("IronCoreVM") {
		for _, ref := range vm.References {
			if ref.Kind != "MachinePool" {
				continue
			}
			for _, pool := range graph.sameSourceName(vm.Provenance.SourceID, "MachinePool", "", ref.Name) {
				graph.addRelation(model.Relation{From: vm.Key, To: pool.Key, Type: "scheduled-on", Evidence: model.EvidenceObjectReference, FieldPath: ref.FieldPath, Confidence: model.ConfidenceProven})
			}
		}
	}
}
