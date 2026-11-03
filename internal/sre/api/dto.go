package api

import (
	"time"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
)

type Envelope struct {
	Data any  `json:"data"`
	Meta Meta `json:"meta"`
}

type Meta struct {
	GeneratedAt time.Time    `json:"generatedAt"`
	Sources     []SourceMeta `json:"sources"`
	Partial     bool         `json:"partial"`
	Warnings    []string     `json:"warnings"`
	Continue    string       `json:"continue,omitempty"`
}

type SourceMeta struct {
	ID          string            `json:"id"`
	Environment string            `json:"environment,omitempty"`
	CollectedAt time.Time         `json:"collectedAt,omitempty"`
	Freshness   model.Freshness   `json:"freshness"`
	Access      model.AccessState `json:"access"`
	Duration    time.Duration     `json:"duration"`
	ErrorClass  string            `json:"errorClass,omitempty"`
	ObjectCount map[string]int    `json:"objectCount,omitempty"`
}

type ResourceDTO struct {
	ID       string         `json:"id"`
	Resource model.Resource `json:"resource"`
}

type SearchDTO struct {
	ID           string         `json:"id"`
	Resource     model.Resource `json:"resource"`
	MatchedField string         `json:"matchedField"`
	MatchKind    string         `json:"matchKind"`
	Fragment     string         `json:"fragment"`
	Ranges       any            `json:"ranges"`
}
