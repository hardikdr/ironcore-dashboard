package model

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var (
	forbiddenFixtureKey   = regexp.MustCompile(`(?i)(^|[-_.])(token|password|kubeconfig|ignition|secretdata|privatekey|clientcertificate|clientkey|data|stringdata)($|[-_.])`)
	forbiddenFixtureValue = regexp.MustCompile(`(?i)(-----BEGIN [A-Z ]*(PRIVATE KEY|CERTIFICATE)-----|\bbearer\s+[a-z0-9._~+/=-]+|apiVersion:\s*v1\s+clusters:|ignition\s*:)`)
)

func TestRecordedFixturesContainNoCredentialMaterial(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "testdata")
	files := 0
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		files++
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var value any
		if err := json.Unmarshal(body, &value); err != nil {
			return fmt.Errorf("%s: invalid JSON: %w", path, err)
		}
		if err := rejectCredentialMaterial(value, "$"); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if files < 4 {
		t.Fatalf("scanned %d fixture files, want at least one per source family", files)
	}
}

func TestFixtureScannerRejectsCredentialShapes(t *testing.T) {
	t.Parallel()

	for _, fixture := range []map[string]any{
		{"bearer-token": "redacted"},
		{"status": map[string]any{"access": map[string]any{"secretData": "redacted"}}},
		{"safe": "Bearer abc.def.ghi"},
		{"safe": "-----BEGIN PRIVATE KEY-----"},
		{"ignition": map[string]any{"version": "3.4.0"}},
	} {
		if err := rejectCredentialMaterial(fixture, "$"); err == nil {
			t.Fatalf("scanner accepted forbidden fixture %#v", fixture)
		}
	}
}

func rejectCredentialMaterial(value any, path string) error {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if forbiddenFixtureKey.MatchString(strings.ReplaceAll(key, " ", "-")) {
				return fmt.Errorf("forbidden key at %s.%s", path, key)
			}
			if err := rejectCredentialMaterial(child, path+"."+key); err != nil {
				return err
			}
		}
	case []any:
		for index, child := range typed {
			if err := rejectCredentialMaterial(child, fmt.Sprintf("%s[%d]", path, index)); err != nil {
				return err
			}
		}
	case string:
		if forbiddenFixtureValue.MatchString(typed) {
			return fmt.Errorf("forbidden credential-like value at %s", path)
		}
	}
	return nil
}
