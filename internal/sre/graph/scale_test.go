package graph

import (
	"testing"
	"time"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/search"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/testutil"
)

func TestGeneratedProductionCardinality(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping 65k-resource scale validation in short mode")
	}
	started := time.Now()
	resources := testutil.ScaleResources()
	if got, want := len(resources), testutil.ScaleVMs+testutil.ScaleNICs+testutil.ScaleLoadBalancers; got != want {
		t.Fatalf("generated resources = %d, want %d", got, want)
	}
	topology := Build(resources)
	if got := len(topology.AllResources()); got != len(resources) {
		t.Fatalf("graph resources = %d, want %d", got, len(resources))
	}
	index := search.New(resources)
	results := index.Search("vm-uid-09999", search.Options{Limit: 10})
	if len(results) != 1 || results[0].Resource.Provenance.Name != "vm-09999" {
		t.Fatalf("exact scale search = %#v", results)
	}
	if traversal := topology.Traverse(results[0].Resource.Key, 2, 500); len(traversal.Resources) <= 1 || len(traversal.Resources) > 500 || len(traversal.Relations) == 0 {
		t.Fatalf("bounded traversal resources = %d relations = %d", len(traversal.Resources), len(traversal.Relations))
	}
	t.Logf("validated %d resources (%d VMs, %d NICs, %d LBs) in %s", len(resources), testutil.ScaleVMs, testutil.ScaleNICs, testutil.ScaleLoadBalancers, time.Since(started))
}
