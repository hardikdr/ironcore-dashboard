package graph

import (
	"strings"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
)

type index struct {
	byKind           map[string][]model.ResourceKey
	bySourceKindName map[string][]model.ResourceKey
	bySourceKindUID  map[string][]model.ResourceKey
	byAlias          map[string][]model.ResourceKey
}

func newIndex(resources map[model.ResourceKey]model.Resource) index {
	result := index{
		byKind: make(map[string][]model.ResourceKey), bySourceKindName: make(map[string][]model.ResourceKey),
		bySourceKindUID: make(map[string][]model.ResourceKey), byAlias: make(map[string][]model.ResourceKey),
	}
	for key, resource := range resources {
		kind := resource.Provenance.Kind
		result.byKind[kind] = append(result.byKind[kind], key)
		nameKey := sourceKindName(resource.Provenance.SourceID, kind, resource.Provenance.Namespace, resource.Provenance.Name)
		result.bySourceKindName[nameKey] = append(result.bySourceKindName[nameKey], key)
		if resource.Provenance.UID != "" {
			uidKey := sourceKindUID(resource.Provenance.SourceID, kind, resource.Provenance.UID)
			result.bySourceKindUID[uidKey] = append(result.bySourceKindUID[uidKey], key)
		}
		for _, alias := range resource.Aliases {
			result.byAlias[alias] = append(result.byAlias[alias], key)
		}
	}
	return result
}

func sourceKindName(source, kind, namespace, name string) string {
	return strings.Join([]string{source, kind, namespace, name}, "\x00")
}

func sourceKindUID(source, kind, uid string) string {
	return strings.Join([]string{source, kind, uid}, "\x00")
}

func (graph *Graph) byKind(kind string) []model.Resource {
	return graph.resourcesForKeys(graph.index.byKind[kind])
}

func (graph *Graph) sameSourceName(source, kind, namespace, name string) []model.Resource {
	return graph.resourcesForKeys(graph.index.bySourceKindName[sourceKindName(source, kind, namespace, name)])
}

func (graph *Graph) sameSourceUID(source, kind, uid string) []model.Resource {
	return graph.resourcesForKeys(graph.index.bySourceKindUID[sourceKindUID(source, kind, uid)])
}

func (graph *Graph) resourcesForKeys(keys []model.ResourceKey) []model.Resource {
	result := make([]model.Resource, 0, len(keys))
	for _, key := range keys {
		result = append(result, graph.resources[key])
	}
	return result
}
