package testutil

import (
	"fmt"
	"time"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
)

const (
	ScaleVMs           = 10_000
	ScaleNICs          = 50_000
	ScaleLoadBalancers = 5_000
)

// ScaleResources generates sanitized resources at the minimum validation cardinalities.
func ScaleResources() []model.Resource {
	resources := make([]model.Resource, 0, ScaleVMs+ScaleNICs+ScaleLoadBalancers)
	for index := 0; index < ScaleVMs; index++ {
		name, uid := fmt.Sprintf("vm-%05d", index), fmt.Sprintf("vm-uid-%05d", index)
		resources = append(resources, resource("IronCoreVM", "ironcore-vm", name, uid, nil))
	}
	for index := 0; index < ScaleNICs; index++ {
		vmIndex := index % ScaleVMs
		refs := []model.ObjectReference{{Group: "scale.ironcore.dev", Kind: "IronCoreVM", Namespace: "scale", Name: fmt.Sprintf("vm-%05d", vmIndex), UID: fmt.Sprintf("vm-uid-%05d", vmIndex), FieldPath: ".spec.machineRef"}}
		resources = append(resources, resource("Nic", "nic", fmt.Sprintf("nic-%05d", index), fmt.Sprintf("nic-uid-%05d", index), refs))
	}
	for index := 0; index < ScaleLoadBalancers; index++ {
		resources = append(resources, resource("LoadBalancer", "load-balancer", fmt.Sprintf("lb-%05d", index), fmt.Sprintf("lb-uid-%05d", index), nil))
	}
	return resources
}

func resource(kind, resourceType, name, uid string, references []model.ObjectReference) model.Resource {
	key, err := model.NewResourceKey(model.ResourceIdentity{SourceID: "scale", Group: "scale.ironcore.dev", Kind: kind, Namespace: "scale", Name: name, UID: uid})
	if err != nil {
		panic(err)
	}
	return model.Resource{Key: key, Type: resourceType, Health: model.HealthHealthy, State: "Ready", References: references, Provenance: model.Provenance{SourceID: "scale", Environment: "generated", Group: "scale.ironcore.dev", Version: "v1", Kind: kind, Namespace: "scale", Name: name, UID: uid, CollectedAt: time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC), Freshness: model.FreshnessFresh, Access: model.AccessOK, Evidence: model.EvidenceObserved}}
}
