package snapshot

import (
	"time"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
)

type ErrorClass string

const (
	ErrorNone           ErrorClass = ""
	ErrorTransport      ErrorClass = "transport"
	ErrorTimeout        ErrorClass = "timeout"
	ErrorAuthentication ErrorClass = "authentication"
	ErrorRBAC           ErrorClass = "rbac"
	ErrorUnsupported    ErrorClass = "unsupported"
	ErrorUnknown        ErrorClass = "unknown"
)

type Snapshot struct {
	SourceID     string           `json:"sourceId"`
	CollectedAt  time.Time        `json:"collectedAt"`
	Resources    []model.Resource `json:"resources"`
	ObjectCounts map[string]int   `json:"objectCounts"`
	Partial      bool             `json:"partial"`
	Warnings     []string         `json:"warnings,omitempty"`
	Gaps         []AccessGap      `json:"gaps,omitempty"`
}

type AccessGap struct {
	Resource string            `json:"resource"`
	Access   model.AccessState `json:"access"`
	Message  string            `json:"message"`
}

type SourceStatus struct {
	LastAttempt time.Time         `json:"lastAttempt"`
	LastSuccess time.Time         `json:"lastSuccess,omitempty"`
	Duration    time.Duration     `json:"duration"`
	ErrorClass  ErrorClass        `json:"errorClass,omitempty"`
	Access      model.AccessState `json:"access"`
	Freshness   model.Freshness   `json:"freshness"`
}

type SourceView struct {
	SourceID   string       `json:"sourceId"`
	Generation uint64       `json:"generation"`
	Snapshot   Snapshot     `json:"snapshot"`
	Status     SourceStatus `json:"status"`
}

type Aggregate struct {
	Generation  uint64       `json:"generation"`
	GeneratedAt time.Time    `json:"generatedAt"`
	Sources     []SourceView `json:"sources"`
}
