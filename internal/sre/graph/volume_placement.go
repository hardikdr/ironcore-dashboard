package graph

import (
	"k8s.io/apimachinery/pkg/labels"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
)

func resolveVolumePlacement(graph *Graph) {
	for _, vm := range graph.byKind("IronCoreVM") {
		for _, ref := range vm.References {
			if ref.Kind != "Volume" {
				continue
			}
			matches, evidence := resolveSameSourceRef(graph, vm.Provenance.SourceID, "Volume", vm.Provenance.Namespace, ref)
			for _, volume := range matches {
				graph.addRelation(model.Relation{From: vm.Key, To: volume.Key, Type: "has-volume", Evidence: evidence, FieldPath: ref.FieldPath, Confidence: model.ConfidenceProven})
			}
		}
	}
	for _, volume := range graph.byKind("Volume") {
		placed := false
		for _, ref := range volume.References {
			if ref.Kind != "VolumePool" {
				continue
			}
			for _, pool := range graph.sameSourceName(volume.Provenance.SourceID, "VolumePool", "", ref.Name) {
				placed = true
				graph.addRelation(model.Relation{From: volume.Key, To: pool.Key, Type: "placed-on-volume-pool", Evidence: model.EvidenceObjectReference, FieldPath: ref.FieldPath, Confidence: model.ConfidenceProven})
			}
		}
		if placed {
			continue
		}
		selector, err := selectorFor(volume)
		if err != nil || selector == nil {
			continue
		}
		for _, pool := range graph.byKind("VolumePool") {
			if pool.Provenance.SourceID == volume.Provenance.SourceID && selector.Matches(labels.Set(pool.Labels)) {
				graph.addRelation(model.Relation{From: volume.Key, To: pool.Key, Type: "placed-on-volume-pool", Evidence: model.EvidenceSelectorMatch, FieldPath: ".spec.volumePoolSelector", Confidence: model.ConfidenceProven})
			}
		}
	}
}
