package search

import "strings"

type MatchKind string

const (
	MatchExact  MatchKind = "exact"
	MatchPrefix MatchKind = "prefix"
	MatchToken  MatchKind = "token"
)

var fieldWeights = map[string]int{
	"uid": 90, "ip": 85, "providerId": 80, "name": 75, "namespace/name": 70,
	"alias": 65, "label": 60,
}

func rank(query, value, field string) (int, MatchKind, bool) {
	lowerQuery, lowerValue := strings.ToLower(query), strings.ToLower(value)
	weight := fieldWeights[field]
	switch {
	case lowerValue == lowerQuery:
		return 3000 + weight, MatchExact, true
	case strings.HasPrefix(lowerValue, lowerQuery):
		return 2000 + weight, MatchPrefix, true
	default:
		for _, token := range strings.FieldsFunc(lowerValue, func(r rune) bool {
			return r == '/' || r == '-' || r == '_' || r == '.' || r == ':' || r == ' ' || r == '@'
		}) {
			if strings.HasPrefix(token, lowerQuery) {
				return 1000 + weight, MatchToken, true
			}
		}
	}
	return 0, "", false
}
