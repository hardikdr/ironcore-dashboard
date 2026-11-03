package graph

import (
	"fmt"
	"sort"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
)

const (
	MaxTraversalDepth = 5
	MaxTraversalNodes = 500
)

type Graph struct {
	resources    map[model.ResourceKey]model.Resource
	index        index
	outgoing     map[model.ResourceKey][]model.Relation
	incoming     map[model.ResourceKey][]model.Relation
	relationKeys map[string]struct{}
	observations []model.Observation
}

type Subgraph struct {
	Resources []model.Resource `json:"resources"`
	Relations []model.Relation `json:"relations"`
	Truncated bool             `json:"truncated"`
}

func New(resources []model.Resource) *Graph {
	graph := &Graph{
		resources: make(map[model.ResourceKey]model.Resource, len(resources)), outgoing: make(map[model.ResourceKey][]model.Relation),
		incoming: make(map[model.ResourceKey][]model.Relation), relationKeys: make(map[string]struct{}),
	}
	for _, resource := range resources {
		graph.resources[resource.Key] = resource
	}
	graph.index = newIndex(graph.resources)
	return graph
}

func Build(resources []model.Resource) *Graph {
	graph := New(resources)
	resolveVMPool(graph)
	resolvePoolNode(graph)
	resolveVMShoot(graph)
	resolveVMNIC(graph)
	resolveLBTargets(graph)
	resolveVolumePlacement(graph)
	resolveMCM(graph)
	return graph
}

func (graph *Graph) addRelation(relation model.Relation) {
	if _, fromOK := graph.resources[relation.From]; !fromOK {
		return
	}
	if _, toOK := graph.resources[relation.To]; !toOK {
		return
	}
	key := fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%s", relation.From, relation.To, relation.Type, relation.Evidence, relation.FieldPath)
	if _, exists := graph.relationKeys[key]; exists {
		return
	}
	graph.relationKeys[key] = struct{}{}
	graph.outgoing[relation.From] = append(graph.outgoing[relation.From], relation)
	graph.incoming[relation.To] = append(graph.incoming[relation.To], relation)
}

func (graph *Graph) addObservation(reason, message string, related ...model.ResourceKey) {
	name := fmt.Sprintf("graph-%03d", len(graph.observations)+1)
	key, _ := model.NewResourceKey(model.ResourceIdentity{SourceID: "graph", Group: "core", Kind: "Observation", Name: name})
	graph.observations = append(graph.observations, model.Observation{
		Resource: model.Resource{Key: key, Type: "observation", Health: model.HealthUnknown, Provenance: model.Provenance{SourceID: "graph", Group: "core", Kind: "Observation", Name: name, Access: model.AccessOK, Evidence: model.EvidenceDerived}},
		Reason:   reason, Message: message, Severity: "warning", Confidence: model.ConfidenceProven, Related: append([]model.ResourceKey(nil), related...),
	})
}

func (graph *Graph) Outgoing(key model.ResourceKey) []model.Relation {
	return append([]model.Relation(nil), graph.outgoing[key]...)
}

func (graph *Graph) Incoming(key model.ResourceKey) []model.Relation {
	return append([]model.Relation(nil), graph.incoming[key]...)
}

func (graph *Graph) Observations() []model.Observation {
	return append([]model.Observation(nil), graph.observations...)
}

func (graph *Graph) Resource(key model.ResourceKey) (model.Resource, bool) {
	resource, ok := graph.resources[key]
	return resource, ok
}

func (graph *Graph) AllResources() []model.Resource {
	keys := make([]string, 0, len(graph.resources))
	byString := make(map[string]model.ResourceKey, len(graph.resources))
	for key := range graph.resources {
		text := string(key)
		keys = append(keys, text)
		byString[text] = key
	}
	sort.Strings(keys)
	result := make([]model.Resource, 0, len(keys))
	for _, key := range keys {
		result = append(result, graph.resources[byString[key]])
	}
	return result
}

func (graph *Graph) Traverse(root model.ResourceKey, depth, maxNodes int) Subgraph {
	if _, ok := graph.resources[root]; !ok {
		return Subgraph{}
	}
	if depth <= 0 {
		depth = 1
	}
	if depth > MaxTraversalDepth {
		depth = MaxTraversalDepth
	}
	if maxNodes <= 0 || maxNodes > MaxTraversalNodes {
		maxNodes = MaxTraversalNodes
	}
	type queueItem struct {
		key   model.ResourceKey
		depth int
	}
	queue := []queueItem{{key: root}}
	seen := map[model.ResourceKey]struct{}{root: {}}
	result := Subgraph{}
	relationSeen := make(map[string]struct{})
	for len(queue) != 0 {
		item := queue[0]
		queue = queue[1:]
		result.Resources = append(result.Resources, graph.resources[item.key])
		if item.depth >= depth {
			continue
		}
		edges := append(graph.Outgoing(item.key), graph.Incoming(item.key)...)
		for _, edge := range edges {
			relationKey := fmt.Sprintf("%s\x00%s\x00%s", edge.From, edge.To, edge.Type)
			if _, exists := relationSeen[relationKey]; !exists {
				relationSeen[relationKey] = struct{}{}
				result.Relations = append(result.Relations, edge)
			}
			neighbor := edge.To
			if neighbor == item.key {
				neighbor = edge.From
			}
			if _, exists := seen[neighbor]; exists {
				continue
			}
			if len(seen) >= maxNodes {
				result.Truncated = true
				continue
			}
			seen[neighbor] = struct{}{}
			queue = append(queue, queueItem{key: neighbor, depth: item.depth + 1})
		}
	}
	return result
}
