package graph

import (
	"testing"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
)

func TestResolveLBTargetsEvaluatesSelectorWithinSourceAndNamespace(t *testing.T) {
	t.Parallel()
	lb := testResource(t, "central", "LoadBalancer", "load-balancer", "tenant", "api", "lb-uid")
	lb.Selector = map[string]string{"cluster": "alpha"}
	lb.SelectorExpressions = []model.SelectorRequirement{{Key: "role", Operator: "In", Values: []string{"worker"}}}
	matching := testResource(t, "central", "Nic", "nic", "tenant", "nic-a", "nic-a-uid")
	matching.Labels = map[string]string{"cluster": "alpha", "role": "worker"}
	otherNamespace := testResource(t, "central", "Nic", "nic", "other", "nic-b", "nic-b-uid")
	otherNamespace.Labels = matching.Labels
	otherSource := testResource(t, "dev", "Nic", "nic", "tenant", "nic-c", "nic-c-uid")
	otherSource.Labels = matching.Labels

	graph := Build([]model.Resource{lb, matching, otherNamespace, otherSource})
	requireEdge(t, graph.Outgoing(lb.Key), lb.Key, matching.Key, "selects-nic", model.EvidenceSelectorMatch)
	if len(graph.Outgoing(lb.Key)) != 1 {
		t.Fatalf("selector crossed namespace/source boundary: %#v", graph.Outgoing(lb.Key))
	}
}
