package source

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	sreconfig "github.com/ironcore-dev/ironcore-dashboard/internal/sre/config"
)

func TestRegistryBuildsBoundedReadOnlyConfigs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	kubeconfig := filepath.Join(dir, "source.yaml")
	writeKubeconfig(t, kubeconfig, "")
	cfg := sreconfig.Config{
		Enabled:        true,
		RequestTimeout: 9 * time.Second,
		QPS:            7,
		Burst:          11,
		Sources: []sreconfig.Source{{
			ID: "wdf4a", Environment: "canary", Role: sreconfig.RoleZonalCAPI,
			Zone: "wdf4a", Kubeconfig: kubeconfig,
		}},
	}

	registry, err := NewRegistry(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}
	entry, ok := registry.Get("wdf4a")
	if !ok {
		t.Fatal("source wdf4a not registered")
	}
	if entry.Config.Timeout != 9*time.Second || entry.Config.QPS != 7 || entry.Config.Burst != 11 {
		t.Fatalf("rest config bounds = timeout %s qps %v burst %d", entry.Config.Timeout, entry.Config.QPS, entry.Config.Burst)
	}
	if entry.Config.UserAgent != "ironcore-dashboard-sre/wdf4a" {
		t.Fatalf("user agent = %q", entry.Config.UserAgent)
	}
	if entry.Config.WrapTransport == nil {
		t.Fatal("read-only transport wrapper is nil")
	}

	called := 0
	base := roundTripperFunc(func(*http.Request) (*http.Response, error) {
		called++
		return &http.Response{StatusCode: http.StatusOK}, nil
	})
	transport := entry.Config.WrapTransport(base)
	if _, err := transport.RoundTrip(mustRequest(t, http.MethodGet)); err != nil || called != 1 {
		t.Fatalf("GET error = %v, base calls = %d", err, called)
	}
	if _, err := transport.RoundTrip(mustRequest(t, http.MethodPost)); !errors.Is(err, ErrMutationBlocked) || called != 1 {
		t.Fatalf("POST error = %v, base calls = %d", err, called)
	}
}

func TestRegistryRejectsImpersonationWithoutLeakingConnectionDetails(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	kubeconfig := filepath.Join(dir, "source.yaml")
	writeKubeconfig(t, kubeconfig, "    as: cluster-admin\n")
	cfg := sreconfig.Config{
		Enabled: true, RequestTimeout: time.Second, QPS: 1, Burst: 1,
		Sources: []sreconfig.Source{{ID: "wdf4a", Environment: "canary", Role: sreconfig.RoleZonalCAPI, Zone: "wdf4a", Kubeconfig: kubeconfig}},
	}

	registry, err := NewRegistry(context.Background(), cfg)
	if err == nil {
		t.Fatal("NewRegistry() succeeded, want error")
	}
	if strings.Contains(err.Error(), kubeconfig) || strings.Contains(err.Error(), "127.0.0.1") {
		t.Fatalf("error leaked connection details: %v", err)
	}
	if diagnostics := registry.Diagnostics(); len(diagnostics) != 1 || diagnostics[0].SourceID != "wdf4a" || diagnostics[0].Access != "denied" {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
}

func mustRequest(t *testing.T, method string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, "https://source.invalid/api", nil)
	if err != nil {
		t.Fatal(err)
	}
	return req
}

func writeKubeconfig(t *testing.T, path, authExtra string) {
	t.Helper()
	body := "apiVersion: v1\nkind: Config\nclusters:\n- name: source\n  cluster:\n    server: https://127.0.0.1:6443\ncontexts:\n- name: source\n  context:\n    cluster: source\n    user: dashboard\ncurrent-context: source\nusers:\n- name: dashboard\n  user:\n" + authExtra
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }
