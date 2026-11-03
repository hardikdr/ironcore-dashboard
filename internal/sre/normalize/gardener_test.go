package normalize

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/config"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestGardenerNormalizerCreatesShootAndWorkerGroupsWithoutAccessMaterial(t *testing.T) {
	t.Parallel()

	normalizer := NewGardener(config.Source{ID: "ironcore-consumer-gardener", Environment: "canary", Role: config.RoleGardener, InstallationRole: "consumer"}, time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC))
	shoot := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "core.gardener.cloud/v1beta1", "kind": "Shoot",
		"metadata": map[string]any{"namespace": "garden-feos", "name": "fecode", "uid": "shoot-uid"},
		"spec": map[string]any{"provider": map[string]any{"type": "ironcore", "workers": []any{
			map[string]any{"name": "worker-a", "minimum": int64(1), "maximum": int64(3)},
			map[string]any{"name": "worker-b", "minimum": int64(2), "maximum": int64(4)},
		}}},
		"status": map[string]any{"credentials": map[string]any{"rotation": map[string]any{"certificateAuthorities": map[string]any{"lastInitiationTime": "2026-11-03T10:00:00Z"}}}, "access": map[string]any{"secretRef": map[string]any{"name": "shoot-kubeconfig"}}},
	}}

	got, err := normalizer.Normalize(schema.GroupVersionResource{Group: "core.gardener.cloud", Version: "v1beta1", Resource: "shoots"}, shoot)
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if len(got) != 3 || got[0].Type != "shoot" || got[0].Provenance.InstallationRole != "consumer" {
		t.Fatalf("normalized resources = %#v", got)
	}
	if got[0].Aliases[0] != "shoot--feos--fecode" || got[1].Type != "worker-group" || got[2].Type != "worker-group" {
		t.Fatalf("shoot aliases/workers = %#v", got)
	}
	body, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"shoot-kubeconfig", "secretRef", "credentials", "certificateAuthorities"} {
		if strings.Contains(string(body), forbidden) {
			t.Fatalf("normalized JSON leaked %q: %s", forbidden, body)
		}
	}
}

func TestGardenerNormalizerKeepsIronCoreMetalAsBareMetal(t *testing.T) {
	t.Parallel()

	normalizer := NewGardener(config.Source{ID: "wdf4-bare-metal-gardener", Environment: "canary", Role: config.RoleGardener, InstallationRole: "bare-metal"}, time.Now())
	machine := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "machine.sapcloud.io/v1alpha1", "kind": "Machine",
		"metadata": map[string]any{"namespace": "garden-infra", "name": "wdf4ai-worker-0", "uid": "mcm-uid", "labels": map[string]any{"node": "wdf4ai-worker-0"}},
		"spec":     map[string]any{"providerID": "ironcore-metal://server-ai-01"},
		"status":   map[string]any{"currentStatus": map[string]any{"phase": "Running"}},
	}}

	got, err := normalizer.Normalize(schema.GroupVersionResource{Group: "machine.sapcloud.io", Version: "v1alpha1", Resource: "machines"}, machine)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Type != "mcm-machine" || got[0].ProviderID != "ironcore-metal://server-ai-01" || got[0].Sanitized["providerType"] != "bare-metal" {
		t.Fatalf("normalized MCM machine = %#v", got)
	}
	for _, ref := range got[0].References {
		if ref.Kind == "IronCoreVM" {
			t.Fatalf("bare-metal worker fabricated VM reference: %#v", ref)
		}
	}
	if got[0].Health != model.HealthHealthy {
		t.Fatalf("health = %q", got[0].Health)
	}
}
