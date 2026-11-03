package search

import (
	"sort"
	"strings"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
)

type Range struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type Match struct {
	Field        string    `json:"field"`
	Kind         MatchKind `json:"kind"`
	Fragment     string    `json:"fragment"`
	Ranges       []Range   `json:"ranges"`
	RenderedHTML string    `json:"-"`
}

type Result struct {
	Resource model.Resource `json:"resource"`
	Match    Match          `json:"match"`
	Rank     int            `json:"rank"`
}

type Options struct {
	Types []string
	Limit int
}

type fieldValue struct {
	field string
	value string
}

type indexedResource struct {
	resource model.Resource
	fields   []fieldValue
}

type Index struct {
	resources []indexedResource
}

func New(resources []model.Resource) *Index {
	index := &Index{resources: make([]indexedResource, 0, len(resources))}
	for _, resource := range resources {
		fields := []fieldValue{{field: "name", value: resource.Provenance.Name}}
		if resource.Provenance.Namespace != "" {
			fields = append(fields, fieldValue{field: "namespace/name", value: resource.Provenance.Namespace + "/" + resource.Provenance.Name})
		}
		if resource.Provenance.UID != "" {
			fields = append(fields, fieldValue{field: "uid", value: resource.Provenance.UID})
		}
		if resource.ProviderID != "" {
			fields = append(fields, fieldValue{field: "providerId", value: resource.ProviderID})
		}
		for _, address := range resource.Addresses {
			fields = append(fields, fieldValue{field: "ip", value: address})
		}
		for _, alias := range resource.Aliases {
			fields = append(fields, fieldValue{field: "alias", value: alias})
		}
		for key, value := range resource.Labels {
			if key == "extension.ironcore.dev/cluster-name" || key == "kubernetes.io/cluster" || key == "topology.ironcore.dev/zone" {
				fields = append(fields, fieldValue{field: "label", value: value})
			}
		}
		index.resources = append(index.resources, indexedResource{resource: resource, fields: fields})
	}
	return index
}

func (index *Index) Search(query string, options Options) []Result {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}
	if options.Limit <= 0 || options.Limit > 200 {
		options.Limit = 50
	}
	types := make(map[string]struct{}, len(options.Types))
	for _, resourceType := range options.Types {
		types[resourceType] = struct{}{}
	}
	results := make([]Result, 0)
	for _, indexed := range index.resources {
		if len(types) != 0 {
			if _, ok := types[indexed.resource.Type]; !ok {
				continue
			}
		}
		best := Result{Resource: indexed.resource}
		for _, field := range indexed.fields {
			score, kind, matched := rank(query, field.value, field.field)
			if !matched || score <= best.Rank {
				continue
			}
			best.Rank = score
			best.Match = Match{Field: field.field, Kind: kind, Fragment: field.value, Ranges: highlightRanges(field.value, query)}
		}
		if best.Rank != 0 {
			results = append(results, best)
		}
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Rank == results[j].Rank {
			return results[i].Resource.Key < results[j].Resource.Key
		}
		return results[i].Rank > results[j].Rank
	})
	if len(results) > options.Limit {
		results = results[:options.Limit]
	}
	return results
}

func highlightRanges(value, query string) []Range {
	lowerValue, lowerQuery := strings.ToLower(value), strings.ToLower(query)
	if lowerQuery == "" {
		return nil
	}
	var result []Range
	for offset := 0; offset < len(lowerValue); {
		index := strings.Index(lowerValue[offset:], lowerQuery)
		if index < 0 {
			break
		}
		start := offset + index
		result = append(result, Range{Start: start, End: start + len(lowerQuery)})
		offset = start + len(lowerQuery)
	}
	return result
}
