package search

import (
	"testing"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
)

func TestSearchRanksExactIdentifiersBeforePrefixAndTokens(t *testing.T) {
	t.Parallel()

	resources := []model.Resource{
		searchResource(t, "central", "IronCoreVM", "ironcore-vm", "tenant", "vm-alpha", "uid-exact", []string{"10.20.0.17"}, "libvirt://host/vm-alpha", nil),
		searchResource(t, "central", "IronCoreVM", "ironcore-vm", "tenant", "uid-exact-prefix", "uid-other", nil, "", nil),
		searchResource(t, "gardener", "Shoot", "shoot", "garden-feos", "fecode", "shoot-uid", nil, "", []string{"shoot--feos--fecode", "FEOS/fecode"}),
	}
	index := New(resources)

	results := index.Search("uid-exact", Options{Limit: 10})
	if len(results) < 2 || results[0].Resource.Provenance.UID != "uid-exact" || results[0].Match.Kind != MatchExact || results[0].Match.Field != "uid" {
		t.Fatalf("results = %#v", results)
	}
	if results[1].Match.Kind == MatchExact {
		t.Fatalf("prefix result ranked as exact: %#v", results[1])
	}

	for query, field := range map[string]string{"10.20.0.17": "ip", "libvirt://host/vm-alpha": "providerId", "shoot--feos--fecode": "alias", "garden-feos/fecode": "namespace/name"} {
		got := index.Search(query, Options{Limit: 1})
		if len(got) != 1 || got[0].Match.Kind != MatchExact || got[0].Match.Field != field {
			t.Fatalf("search %q = %#v", query, got)
		}
	}
}

func TestSearchReturnsSafeHighlightRangesWithoutMarkup(t *testing.T) {
	t.Parallel()

	resource := searchResource(t, "central", "IronCoreVM", "ironcore-vm", "tenant", "safe-name", "uid", nil, "", []string{"vm-<script>alert(1)</script>"})
	result := New([]model.Resource{resource}).Search("script", Options{Limit: 1})
	if len(result) != 1 || result[0].Match.Fragment != "vm-<script>alert(1)</script>" || len(result[0].Match.Ranges) != 2 {
		t.Fatalf("result = %#v", result)
	}
	if result[0].Match.RenderedHTML != "" {
		t.Fatalf("search emitted HTML: %q", result[0].Match.RenderedHTML)
	}
}

func TestSearchFiltersTypesAndEnforcesLimit(t *testing.T) {
	t.Parallel()
	resources := []model.Resource{
		searchResource(t, "central", "IronCoreVM", "ironcore-vm", "tenant", "shared", "vm", nil, "", nil),
		searchResource(t, "central", "Volume", "volume", "tenant", "shared", "volume", nil, "", nil),
	}
	results := New(resources).Search("shared", Options{Types: []string{"volume"}, Limit: 1})
	if len(results) != 1 || results[0].Resource.Type != "volume" {
		t.Fatalf("results = %#v", results)
	}
}

func searchResource(t *testing.T, source, kind, resourceType, namespace, name, uid string, addresses []string, providerID string, aliases []string) model.Resource {
	t.Helper()
	key, err := model.NewResourceKey(model.ResourceIdentity{SourceID: source, Group: "test.example", Kind: kind, Namespace: namespace, Name: name, UID: uid})
	if err != nil {
		t.Fatal(err)
	}
	return model.Resource{Key: key, Type: resourceType, Addresses: addresses, ProviderID: providerID, Aliases: aliases, Provenance: model.Provenance{SourceID: source, Group: "test.example", Kind: kind, Namespace: namespace, Name: name, UID: uid}}
}
