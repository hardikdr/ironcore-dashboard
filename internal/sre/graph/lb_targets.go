package graph

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
)

func resolveLBTargets(graph *Graph) {
	for _, lb := range graph.byKind("LoadBalancer") {
		selector, err := selectorFor(lb)
		if err != nil {
			graph.addObservation("invalid-selector", "LoadBalancer selector could not be evaluated", lb.Key)
			continue
		}
		if selector != nil {
			for _, nic := range graph.byKind("Nic") {
				if nic.Provenance.SourceID != lb.Provenance.SourceID || nic.Provenance.Namespace != lb.Provenance.Namespace {
					continue
				}
				if selector.Matches(labels.Set(nic.Labels)) {
					graph.addRelation(model.Relation{From: lb.Key, To: nic.Key, Type: "selects-nic", Evidence: model.EvidenceSelectorMatch, FieldPath: ".spec.networkInterfaceSelector", Confidence: model.ConfidenceProven})
				}
			}
		}
		for _, ref := range lb.References {
			if ref.Kind == "Network" {
				for _, network := range graph.sameSourceName(lb.Provenance.SourceID, "Network", lb.Provenance.Namespace, ref.Name) {
					graph.addRelation(model.Relation{From: lb.Key, To: network.Key, Type: "connected-to-network", Evidence: model.EvidenceObjectReference, FieldPath: ref.FieldPath, Confidence: model.ConfidenceProven})
				}
			}
		}
	}
	for _, routing := range graph.byKind("LoadBalancerRouting") {
		for _, owner := range routing.OwnerRefs {
			if owner.Kind != "LoadBalancer" {
				continue
			}
			matches, _ := resolveSameSourceRef(graph, routing.Provenance.SourceID, "LoadBalancer", routing.Provenance.Namespace, owner)
			for _, lb := range matches {
				graph.addRelation(model.Relation{From: lb.Key, To: routing.Key, Type: "has-routing", Evidence: model.EvidenceOwnerReference, FieldPath: ".metadata.ownerReferences", Confidence: model.ConfidenceProven})
			}
		}
	}
}

func selectorFor(resource model.Resource) (labels.Selector, error) {
	if len(resource.Selector) == 0 && len(resource.SelectorExpressions) == 0 {
		return nil, nil
	}
	selector := &metav1.LabelSelector{MatchLabels: resource.Selector}
	for _, expression := range resource.SelectorExpressions {
		selector.MatchExpressions = append(selector.MatchExpressions, metav1.LabelSelectorRequirement{Key: expression.Key, Operator: metav1.LabelSelectorOperator(expression.Operator), Values: expression.Values})
	}
	return metav1.LabelSelectorAsSelector(selector)
}
