package graph

import (
	"strings"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
)

func resolvePoolNode(graph *Graph) {
	for _, kind := range []string{"MachinePool", "VolumePool"} {
		for _, pool := range graph.byKind(kind) {
			zone := pool.Labels["topology.ironcore.dev/zone"]
			if zone != "" {
				for _, candidate := range graph.byKind("Zone") {
					if candidate.Provenance.Name == zone {
						graph.addRelation(model.Relation{From: pool.Key, To: candidate.Key, Type: "located-in-zone", Evidence: model.EvidenceExactLabel, FieldPath: ".metadata.labels[topology.ironcore.dev/zone]", Confidence: model.ConfidenceProven})
					}
				}
			}
			if kind != "MachinePool" || !strings.HasPrefix(pool.ProviderID, "libvirt-provider://") || zone == "" {
				continue
			}
			nodeName := strings.TrimPrefix(pool.ProviderID, "libvirt-provider://")
			var matches []model.Resource
			for _, node := range graph.byKind("PhysicalNode") {
				if node.Type != "physical-node" || node.Provenance.Name != nodeName {
					continue
				}
				nodeZone := node.Labels["topology.kubernetes.io/zone"]
				if node.Provenance.SourceID == zone || nodeZone == zone {
					matches = append(matches, node)
				}
			}
			for _, node := range matches {
				graph.addRelation(model.Relation{From: pool.Key, To: node.Key, Type: "registered-by", Evidence: model.EvidenceProviderID, FieldPath: ".spec.providerID", Confidence: model.ConfidenceProven})
			}
			if len(matches) > 1 {
				keys := []model.ResourceKey{pool.Key}
				for _, match := range matches {
					keys = append(keys, match.Key)
				}
				graph.addObservation("provider-id-conflict", "MachinePool provider ID matched multiple physical nodes in its authoritative zone", keys...)
			}
		}
	}
}
