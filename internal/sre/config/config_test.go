package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadAppliesSafeDefaults(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	kubeconfig := filepath.Join(dir, "source.yaml")
	writeFile(t, kubeconfig, "apiVersion: v1\nkind: Config\n", 0o600)
	configPath := filepath.Join(dir, "sre.yaml")
	writeFile(t, configPath, "enabled: true\nsources:\n  - id: wdf4a\n    environment: canary\n    role: zonal-capi\n    zone: wdf4a\n    kubeconfig: "+kubeconfig+"\n", 0o600)

	got, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.PollInterval != 30*time.Second || got.StaleAfter != 2*time.Minute || got.RequestTimeout != 15*time.Second {
		t.Fatalf("defaults = poll %s, stale %s, timeout %s", got.PollInterval, got.StaleAfter, got.RequestTimeout)
	}
	if got.QPS != 5 || got.Burst != 10 || got.MaxConcurrent != 3 {
		t.Fatalf("client defaults = qps %v burst %d concurrency %d", got.QPS, got.Burst, got.MaxConcurrent)
	}
}

func TestLoadRejectsDuplicateIDsInlineCredentialsAndUnsafeFiles(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"duplicate source": "enabled: true\nsources:\n  - {id: wdf4a, environment: canary, role: zonal-capi, zone: wdf4a, kubeconfig: SOURCE}\n  - {id: wdf4a, environment: canary, role: zonal-capi, zone: wdf4a, kubeconfig: SOURCE}\n",
		"inline token":     "enabled: true\nsources:\n  - {id: wdf4a, environment: canary, role: zonal-capi, zone: wdf4a, kubeconfig: SOURCE, token: abc}\n",
		"inline server":    "enabled: true\nsources:\n  - {id: wdf4a, environment: canary, role: zonal-capi, zone: wdf4a, kubeconfig: SOURCE, server: https://example.invalid}\n",
		"missing zone":     "enabled: true\nsources:\n  - {id: wdf4a, environment: canary, role: zonal-capi, kubeconfig: SOURCE}\n",
		"unknown role":     "enabled: true\nsources:\n  - {id: wdf4a, environment: canary, role: arbitrary, kubeconfig: SOURCE}\n",
	}
	for name, body := range tests {
		name, body := name, body
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			kubeconfig := filepath.Join(dir, "source.yaml")
			writeFile(t, kubeconfig, "apiVersion: v1\nkind: Config\n", 0o600)
			path := filepath.Join(dir, "sre.yaml")
			writeFile(t, path, strings.ReplaceAll(body, "SOURCE", kubeconfig), 0o600)
			if _, err := Load(path); err == nil {
				t.Fatal("Load() succeeded, want error")
			}
		})
	}

	t.Run("unsafe source file permissions", func(t *testing.T) {
		dir := t.TempDir()
		kubeconfig := filepath.Join(dir, "source.yaml")
		writeFile(t, kubeconfig, "apiVersion: v1\nkind: Config\n", 0o644)
		path := filepath.Join(dir, "sre.yaml")
		writeFile(t, path, "enabled: true\nsources:\n  - id: wdf4a\n    environment: canary\n    role: zonal-capi\n    zone: wdf4a\n    kubeconfig: "+kubeconfig+"\n", 0o600)
		if _, err := Load(path); err == nil || strings.Contains(err.Error(), kubeconfig) {
			t.Fatalf("Load() error = %v, want redacted unsafe-permission error", err)
		}
	})
}

func TestLoadDisabledConfigNeedsNoSources(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "sre.yaml")
	writeFile(t, path, "enabled: false\n", 0o600)
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Enabled || len(got.Sources) != 0 {
		t.Fatalf("disabled config = %#v", got)
	}
}

func writeFile(t *testing.T, path, body string, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), mode); err != nil {
		t.Fatal(err)
	}
}
