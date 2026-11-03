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

func TestIronCoreNormalizerPreservesOperationalFieldsAndReferences(t *testing.T) {
	t.Parallel()

	collectedAt := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	normalizer := NewIronCore(config.Source{ID: "canary-ironcore", Environment: "canary", Role: config.RoleCentralIronCore}, collectedAt)
	object := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "compute.ironcore.dev/v1alpha1", "kind": "Machine",
		"metadata": map[string]any{
			"namespace": "shoot--feos--fecode", "name": "worker-a-0", "uid": "vm-001",
			"creationTimestamp": "2026-11-03T10:00:00Z",
			"labels":            map[string]any{"extension.ironcore.dev/cluster-name": "shoot--feos--fecode"},
		},
		"spec": map[string]any{
			"machinePoolRef":    map[string]any{"name": "cn-wdf4a-7"},
			"machineClassRef":   map[string]any{"name": "medium"},
			"ignitionRef":       map[string]any{"name": "boot-material", "key": "ignition.yaml"},
			"networkInterfaces": []any{map[string]any{"name": "primary", "networkInterfaceRef": map[string]any{"name": "nic-001"}}},
			"volumes":           []any{map[string]any{"name": "root", "volumeRef": map[string]any{"name": "volume-001"}}},
		},
		"status": map[string]any{
			"machineID": "libvirt://vm-001", "state": "Running",
			"conditions":        []any{map[string]any{"type": "Ready", "status": "True", "reason": "Provisioned", "lastTransitionTime": "2026-11-03T10:05:00Z"}},
			"networkInterfaces": []any{map[string]any{"name": "primary", "networkInterfaceRef": map[string]any{"name": "nic-001"}}},
			"volumes":           []any{map[string]any{"name": "root", "state": "Attached", "volumeRef": map[string]any{"name": "volume-001"}}},
		},
	}}

	got, err := normalizer.Normalize(schema.GroupVersionResource{Group: "compute.ironcore.dev", Version: "v1alpha1", Resource: "machines"}, object)
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if got.Type != "ironcore-vm" || got.Provenance.Kind != "IronCoreVM" || got.ProviderID != "libvirt://vm-001" || got.State != "Running" || got.Health != model.HealthHealthy {
		t.Fatalf("normalized VM = %#v", got)
	}
	if got.CreatedAt != time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC) || len(got.Conditions) != 1 {
		t.Fatalf("timestamps/conditions = created %s conditions %#v", got.CreatedAt, got.Conditions)
	}
	for _, want := range []string{"MachinePool:cn-wdf4a-7", "MachineClass:medium", "Nic:nic-001", "Volume:volume-001"} {
		if !hasReference(got.References, want) {
			t.Fatalf("references %#v missing %s", got.References, want)
		}
	}
	body, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"boot-material", "ignition.yaml", "ignitionRef"} {
		if strings.Contains(string(body), forbidden) {
			t.Fatalf("normalized JSON leaked %q: %s", forbidden, body)
		}
	}
}

func TestIronCoreNormalizerRedactsSecretReferencesAndPayloadFields(t *testing.T) {
	t.Parallel()

	normalizer := NewIronCore(config.Source{ID: "canary-ironcore", Environment: "canary", Role: config.RoleCentralIronCore}, time.Now())
	object := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "storage.ironcore.dev/v1alpha1", "kind": "Volume",
		"metadata": map[string]any{"namespace": "tenant", "name": "volume-a", "uid": "volume-a-uid"},
		"spec": map[string]any{
			"volumePoolRef": map[string]any{"name": "wdf4a"},
			"encryption":    map[string]any{"secretRef": map[string]any{"name": "encryption-material"}},
			"data":          map[string]any{"encryptionKey": "super-secret"},
		},
		"status": map[string]any{
			"state":  "Available",
			"access": map[string]any{"secretRef": map[string]any{"name": "volume-access"}, "handle": "provider-handle", "token": "do-not-copy"},
		},
	}}

	got, err := normalizer.Normalize(schema.GroupVersionResource{Group: "storage.ironcore.dev", Version: "v1alpha1", Resource: "volumes"}, object)
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	body, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"encryption-material", "volume-access", "super-secret", "do-not-copy", "secretRef", `"data"`} {
		if strings.Contains(string(body), forbidden) {
			t.Fatalf("normalized JSON leaked %q: %s", forbidden, body)
		}
	}
	redactions, ok := got.Sanitized["redactedReferences"].([]string)
	if !ok || len(redactions) != 2 {
		t.Fatalf("redaction metadata = %#v", got.Sanitized)
	}
}

func TestIronCoreNormalizerPreservesSelectorExpressions(t *testing.T) {
	t.Parallel()

	normalizer := NewIronCore(config.Source{ID: "canary-ironcore", Environment: "canary", Role: config.RoleCentralIronCore}, time.Now())
	object := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "networking.ironcore.dev/v1alpha1", "kind": "LoadBalancer",
		"metadata": map[string]any{"namespace": "tenant", "name": "api", "uid": "lb-uid"},
		"spec": map[string]any{
			"networkRef": map[string]any{"name": "tenant-net"},
			"networkInterfaceSelector": map[string]any{
				"matchLabels":      map[string]any{"kubernetes.io/cluster": "shoot--feos--fecode"},
				"matchExpressions": []any{map[string]any{"key": "role", "operator": "In", "values": []any{"worker", "ingress"}}},
			},
		},
	}}

	got, err := normalizer.Normalize(schema.GroupVersionResource{Group: "networking.ironcore.dev", Version: "v1alpha1", Resource: "loadbalancers"}, object)
	if err != nil {
		t.Fatal(err)
	}
	if got.Selector["kubernetes.io/cluster"] != "shoot--feos--fecode" || len(got.SelectorExpressions) != 1 || got.SelectorExpressions[0].Operator != "In" {
		t.Fatalf("selector = labels %#v expressions %#v", got.Selector, got.SelectorExpressions)
	}
}

func hasReference(references []model.ObjectReference, want string) bool {
	for _, ref := range references {
		if ref.Kind+":"+ref.Name == want {
			return true
		}
	}
	return false
}
