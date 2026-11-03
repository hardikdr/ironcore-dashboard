package source

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/config"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var ErrMutationBlocked = errors.New("SRE source transport blocks mutation methods")

type Entry struct {
	Metadata config.Source
	Config   *rest.Config
}

type Diagnostic struct {
	SourceID string            `json:"sourceId"`
	Access   model.AccessState `json:"access"`
	Message  string            `json:"message,omitempty"`
}

type Registry struct {
	entries     map[string]Entry
	diagnostics []Diagnostic
}

func NewRegistry(ctx context.Context, cfg config.Config) (*Registry, error) {
	registry := &Registry{entries: make(map[string]Entry, len(cfg.Sources))}
	if !cfg.Enabled {
		return registry, nil
	}

	var firstErr error
	for _, metadata := range cfg.Sources {
		if err := ctx.Err(); err != nil {
			return registry, err
		}
		restConfig, err := loadRESTConfig(metadata)
		if err != nil {
			registry.diagnostics = append(registry.diagnostics, Diagnostic{
				SourceID: metadata.ID,
				Access:   model.AccessUnreachable,
				Message:  "source configuration could not be loaded",
			})
			if firstErr == nil {
				firstErr = fmt.Errorf("source %q configuration is invalid", metadata.ID)
			}
			continue
		}
		if restConfig.Impersonate.UserName != "" || len(restConfig.Impersonate.Groups) != 0 || len(restConfig.Impersonate.Extra) != 0 {
			registry.diagnostics = append(registry.diagnostics, Diagnostic{
				SourceID: metadata.ID,
				Access:   model.AccessDenied,
				Message:  "source impersonation is forbidden",
			})
			if firstErr == nil {
				firstErr = fmt.Errorf("source %q requests forbidden impersonation", metadata.ID)
			}
			continue
		}

		applyClientBounds(restConfig, metadata.ID, cfg)
		registry.entries[metadata.ID] = Entry{Metadata: metadata, Config: restConfig}
		registry.diagnostics = append(registry.diagnostics, Diagnostic{SourceID: metadata.ID, Access: model.AccessUnknown, Message: "not collected"})
	}
	sort.Slice(registry.diagnostics, func(i, j int) bool {
		return registry.diagnostics[i].SourceID < registry.diagnostics[j].SourceID
	})
	return registry, firstErr
}

func loadRESTConfig(source config.Source) (*rest.Config, error) {
	rules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: source.Kubeconfig}
	overrides := &clientcmd.ConfigOverrides{CurrentContext: source.Context}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides).ClientConfig()
}

func applyClientBounds(restConfig *rest.Config, sourceID string, cfg config.Config) {
	restConfig.Timeout = cfg.RequestTimeout
	restConfig.QPS = cfg.QPS
	restConfig.Burst = cfg.Burst
	restConfig.UserAgent = "ironcore-dashboard-sre/" + sourceID
	previous := restConfig.WrapTransport
	restConfig.WrapTransport = func(next http.RoundTripper) http.RoundTripper {
		if previous != nil {
			next = previous(next)
		}
		return &readOnlyRoundTripper{next: next}
	}
}

type readOnlyRoundTripper struct {
	next http.RoundTripper
}

func (transport *readOnlyRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		return nil, ErrMutationBlocked
	}
	return transport.next.RoundTrip(request)
}

func (registry *Registry) Get(id string) (Entry, bool) {
	entry, ok := registry.entries[id]
	return entry, ok
}

func (registry *Registry) Count() int { return len(registry.entries) }

func (registry *Registry) Entries() []Entry {
	ids := make([]string, 0, len(registry.entries))
	for id := range registry.entries {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result := make([]Entry, 0, len(ids))
	for _, id := range ids {
		result = append(result, registry.entries[id])
	}
	return result
}

func (registry *Registry) Diagnostics() []Diagnostic {
	return append([]Diagnostic(nil), registry.diagnostics...)
}
