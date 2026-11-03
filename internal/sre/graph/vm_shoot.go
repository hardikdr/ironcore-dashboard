package graph

import "github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"

func resolveVMShoot(graph *Graph) {
	for _, vm := range graph.byKind("IronCoreVM") {
		cluster := vm.Labels["extension.ironcore.dev/cluster-name"]
		if cluster == "" {
			cluster = vm.Labels["kubernetes.io/cluster"]
		}
		if cluster == "" {
			continue
		}
		var matches []model.Resource
		for _, key := range graph.index.byAlias[cluster] {
			candidate := graph.resources[key]
			if candidate.Provenance.Kind == "Shoot" {
				matches = append(matches, candidate)
			}
		}
		for _, shoot := range matches {
			graph.addRelation(model.Relation{From: vm.Key, To: shoot.Key, Type: "owned-by-shoot", Evidence: model.EvidenceExactLabel, FieldPath: ".metadata.labels[kubernetes.io/cluster]", Confidence: model.ConfidenceCorroborated})
		}
		if len(matches) > 1 {
			keys := []model.ResourceKey{vm.Key}
			for _, match := range matches {
				keys = append(keys, match.Key)
			}
			graph.addObservation("shoot-label-conflict", "exact Shoot cluster label matched multiple Gardener installations", keys...)
		}
	}
}
