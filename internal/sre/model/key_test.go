package model

import (
	"encoding/json"
	"testing"
)

func TestResourceKeyRoundTrip(t *testing.T) {
	t.Parallel()

	want := ResourceIdentity{
		SourceID:  "canary-ironcore",
		Group:     "compute.ironcore.dev",
		Kind:      "IronCoreVM",
		Namespace: "shoot--feos--fecode",
		Name:      "worker-a-0",
		UID:       "55bded79-0702-4f15-94e2-46f6e100af40",
	}

	key, err := NewResourceKey(want)
	if err != nil {
		t.Fatalf("NewResourceKey() error = %v", err)
	}
	if got := string(key); got != "canary-ironcore/compute.ironcore.dev/IronCoreVM/shoot--feos--fecode/worker-a-0/55bded79-0702-4f15-94e2-46f6e100af40" {
		t.Fatalf("key = %q", got)
	}

	got, err := ParseResourceKey(key)
	if err != nil {
		t.Fatalf("ParseResourceKey() error = %v", err)
	}
	if got != want {
		t.Fatalf("round trip = %#v, want %#v", got, want)
	}
}

func TestResourceKeyUsesPlaceholderForClusterScopedAndMissingUID(t *testing.T) {
	t.Parallel()

	key, err := NewResourceKey(ResourceIdentity{
		SourceID: "wdf4a",
		Group:    "core",
		Kind:     "PhysicalNode",
		Name:     "cn-wdf4a-7",
	})
	if err != nil {
		t.Fatalf("NewResourceKey() error = %v", err)
	}
	if got := string(key); got != "wdf4a/core/PhysicalNode/_/cn-wdf4a-7/_" {
		t.Fatalf("key = %q", got)
	}
	parsed, err := ParseResourceKey(key)
	if err != nil {
		t.Fatalf("ParseResourceKey() error = %v", err)
	}
	if parsed.Namespace != "" || parsed.UID != "" {
		t.Fatalf("placeholders leaked into identity: %#v", parsed)
	}
}

func TestResourceKeyRejectsAmbiguousOrMalformedSegments(t *testing.T) {
	t.Parallel()

	tests := []ResourceIdentity{
		{SourceID: "", Group: "core", Kind: "PhysicalNode", Name: "node-a"},
		{SourceID: "wdf4a", Group: "", Kind: "PhysicalNode", Name: "node-a"},
		{SourceID: "wdf4a", Group: "core", Kind: "Machine", Name: "ambiguous"},
		{SourceID: "wdf4a/other", Group: "core", Kind: "PhysicalNode", Name: "node-a"},
		{SourceID: "wdf4a", Group: "core", Kind: "PhysicalNode", Namespace: "_", Name: "node-a"},
	}
	for _, tt := range tests {
		if _, err := NewResourceKey(tt); err == nil {
			t.Fatalf("NewResourceKey(%#v) succeeded, want error", tt)
		}
	}

	for _, raw := range []ResourceKey{
		"wdf4a/core/PhysicalNode/_/node-a",
		"wdf4a/core/Machine/_/node-a/_",
		"wdf4a/core/PhysicalNode/_/node/a/_",
	} {
		if _, err := ParseResourceKey(raw); err == nil {
			t.Fatalf("ParseResourceKey(%q) succeeded, want error", raw)
		}
	}
}

func TestResourceKeyJSONIsStableString(t *testing.T) {
	t.Parallel()

	key := ResourceKey("wdf4a/core/PhysicalNode/_/cn-wdf4a-7/_")
	got, err := json.Marshal(key)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if string(got) != `"wdf4a/core/PhysicalNode/_/cn-wdf4a-7/_"` {
		t.Fatalf("JSON = %s", got)
	}
}
