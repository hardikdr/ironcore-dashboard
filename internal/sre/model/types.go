package model

import "time"

type Freshness string

const (
	FreshnessFresh   Freshness = "fresh"
	FreshnessStale   Freshness = "stale"
	FreshnessUnknown Freshness = "unknown"
)

type AccessState string

const (
	AccessOK            AccessState = "ok"
	AccessDenied        AccessState = "denied"
	AccessUnreachable   AccessState = "unreachable"
	AccessUnsupported   AccessState = "unsupported"
	AccessNotConfigured AccessState = "not-configured"
	AccessUnknown       AccessState = "unknown"
)

type HealthState string

const (
	HealthHealthy  HealthState = "healthy"
	HealthDegraded HealthState = "degraded"
	HealthFailed   HealthState = "failed"
	HealthUnknown  HealthState = "unknown"
)

type EvidenceClass string

const (
	EvidenceObserved EvidenceClass = "observed-live"
	EvidenceDeclared EvidenceClass = "declared-gitops"
	EvidenceDerived  EvidenceClass = "derived"
	EvidenceUnknown  EvidenceClass = "unknown"
)

type EvidenceKind string

const (
	EvidenceUIDReference    EvidenceKind = "uid-reference"
	EvidenceOwnerReference  EvidenceKind = "owner-reference"
	EvidenceObjectReference EvidenceKind = "object-reference"
	EvidenceSelectorMatch   EvidenceKind = "selector-match"
	EvidenceProviderID      EvidenceKind = "provider-id"
	EvidenceExactLabel      EvidenceKind = "exact-label"
	EvidenceDerivedName     EvidenceKind = "derived-name"
	EvidenceUnproven        EvidenceKind = "unknown"
)

type Confidence string

const (
	ConfidenceProven       Confidence = "proven"
	ConfidenceCorroborated Confidence = "corroborated"
	ConfidenceDerived      Confidence = "derived"
	ConfidenceUnknown      Confidence = "unknown"
)

type Provenance struct {
	SourceID    string        `json:"sourceId"`
	Environment string        `json:"environment"`
	ClusterRole string        `json:"clusterRole"`
	Group       string        `json:"group"`
	Version     string        `json:"version"`
	Kind        string        `json:"kind"`
	Namespace   string        `json:"namespace,omitempty"`
	Name        string        `json:"name"`
	UID         string        `json:"uid,omitempty"`
	CollectedAt time.Time     `json:"collectedAt"`
	Freshness   Freshness     `json:"freshness"`
	Access      AccessState   `json:"access"`
	Evidence    EvidenceClass `json:"evidence"`
}

type ObjectReference struct {
	Group     string `json:"group,omitempty"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name"`
	UID       string `json:"uid,omitempty"`
	FieldPath string `json:"fieldPath,omitempty"`
}

type Condition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	Reason             string    `json:"reason,omitempty"`
	Message            string    `json:"message,omitempty"`
	LastTransitionTime time.Time `json:"lastTransitionTime,omitempty"`
}

type Resource struct {
	Key         ResourceKey       `json:"key"`
	Type        string            `json:"type"`
	Provenance  Provenance        `json:"provenance"`
	Health      HealthState       `json:"health"`
	State       string            `json:"state,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	ProviderID  string            `json:"providerId,omitempty"`
	Addresses   []string          `json:"addresses,omitempty"`
	Aliases     []string          `json:"aliases,omitempty"`
	Conditions  []Condition       `json:"conditions,omitempty"`
	References  []ObjectReference `json:"references,omitempty"`
	OwnerRefs   []ObjectReference `json:"ownerReferences,omitempty"`
	Selector    map[string]string `json:"selector,omitempty"`
	Sanitized   map[string]any    `json:"sanitized,omitempty"`
}

type Relation struct {
	From       ResourceKey  `json:"from"`
	To         ResourceKey  `json:"to"`
	Type       string       `json:"type"`
	Evidence   EvidenceKind `json:"evidence"`
	FieldPath  string       `json:"fieldPath"`
	Confidence Confidence   `json:"confidence"`
}

type SourceCluster struct{ Resource }
type Zone struct{ Resource }
type PhysicalNode struct{ Resource }
type CapiMachine struct{ Resource }
type MachinePool struct{ Resource }
type IronMachine struct{ Resource }
type MachineClass struct{ Resource }
type GardenerInstallation struct{ Resource }
type Shoot struct{ Resource }
type WorkerGroup struct{ Resource }
type McmMachineDeployment struct{ Resource }
type McmMachineSet struct{ Resource }
type McmMachine struct{ Resource }
type Network struct{ Resource }
type Nic struct{ Resource }
type LoadBalancer struct{ Resource }
type LoadBalancerRouting struct{ Resource }
type NATGateway struct{ Resource }
type VirtualIP struct{ Resource }
type VolumePool struct{ Resource }
type Volume struct{ Resource }
type Attachment struct{ Resource }
type ProviderComponent struct{ Resource }
type KubernetesEvent struct{ Resource }

type Observation struct {
	Resource
	Reason        string        `json:"reason"`
	Message       string        `json:"message"`
	Severity      string        `json:"severity"`
	Confidence    Confidence    `json:"confidence"`
	Related       []ResourceKey `json:"related,omitempty"`
	FirstObserved time.Time     `json:"firstObserved"`
	LastObserved  time.Time     `json:"lastObserved"`
}
