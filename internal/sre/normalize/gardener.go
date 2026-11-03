package normalize

import (
	"fmt"
	"strings"
	"time"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/config"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Gardener struct {
	source      config.Source
	collectedAt time.Time
}

func NewGardener(source config.Source, collectedAt time.Time) *Gardener {
	return &Gardener{source: source, collectedAt: collectedAt.UTC()}
}

func (normalizer *Gardener) Installation() (model.Resource, error) {
	key, err := model.NewResourceKey(model.ResourceIdentity{SourceID: normalizer.source.ID, Group: "core", Kind: "GardenerInstallation", Name: normalizer.source.ID})
	if err != nil {
		return model.Resource{}, err
	}
	return model.Resource{
		Key: key, Type: "gardener-installation", Health: model.HealthUnknown,
		Provenance: normalizer.provenance("core", "v1", "GardenerInstallation", "", normalizer.source.ID, ""),
	}, nil
}

func (normalizer *Gardener) Normalize(gvr schema.GroupVersionResource, object *unstructured.Unstructured) ([]model.Resource, error) {
	kind, resourceType, ok := gardenerKind(gvr.Resource)
	if !ok {
		return nil, fmt.Errorf("unsupported Gardener resource %s", gvr.Resource)
	}
	key, err := model.NewResourceKey(model.ResourceIdentity{
		SourceID: normalizer.source.ID, Group: gvr.Group, Kind: kind,
		Namespace: object.GetNamespace(), Name: object.GetName(), UID: string(object.GetUID()),
	})
	if err != nil {
		return nil, err
	}
	state := firstNestedString(object.Object, []string{"status", "currentStatus", "phase"}, []string{"status", "lastOperation", "state"})
	providerID := firstNestedString(object.Object, []string{"spec", "providerID"}, []string{"status", "providerID"})
	resource := model.Resource{
		Key: key, Type: resourceType, CreatedAt: object.GetCreationTimestamp().Time.UTC(),
		Provenance: normalizer.provenance(gvr.Group, gvr.Version, kind, object.GetNamespace(), object.GetName(), string(object.GetUID())),
		Health:     healthFrom(state, normalizeConditions(object.Object)), State: state, Labels: object.GetLabels(), Annotations: safeAnnotations(object.GetAnnotations()),
		ProviderID: providerID, Conditions: normalizeConditions(object.Object), OwnerRefs: normalizeOwners(object),
	}
	resource.Selector, resource.SelectorExpressions = selectorAt(object.Object, "spec", "selector")
	if node := object.GetLabels()["node"]; node != "" {
		resource.Aliases = append(resource.Aliases, node)
	}
	resource.Sanitized = gardenerDetails(gvr.Resource, object.Object, providerID)

	result := []model.Resource{resource}
	if gvr.Resource == "shoots" {
		project := strings.TrimPrefix(object.GetNamespace(), "garden-")
		if project != "" && project != object.GetNamespace() {
			result[0].Aliases = append(result[0].Aliases, "shoot--"+project+"--"+object.GetName(), strings.ToUpper(project)+"/"+object.GetName())
		}
		workers, found, _ := unstructured.NestedSlice(object.Object, "spec", "provider", "workers")
		if found {
			for _, raw := range workers {
				worker, ok := raw.(map[string]any)
				if !ok {
					continue
				}
				name := stringValue(worker["name"])
				if name == "" {
					continue
				}
				workerKey, err := model.NewResourceKey(model.ResourceIdentity{SourceID: normalizer.source.ID, Group: gvr.Group, Kind: "WorkerGroup", Namespace: object.GetNamespace(), Name: object.GetName() + "--" + name})
				if err != nil {
					continue
				}
				result = append(result, model.Resource{
					Key: workerKey, Type: "worker-group", Health: model.HealthUnknown,
					Provenance: normalizer.provenance(gvr.Group, gvr.Version, "WorkerGroup", object.GetNamespace(), object.GetName()+"--"+name, ""),
					References: []model.ObjectReference{{Group: gvr.Group, Kind: "Shoot", Namespace: object.GetNamespace(), Name: object.GetName(), UID: string(object.GetUID()), FieldPath: ".spec.provider.workers"}},
					Sanitized:  map[string]any{"workerName": name, "minimum": worker["minimum"], "maximum": worker["maximum"]},
				})
			}
		}
	}
	return result, nil
}

func (normalizer *Gardener) provenance(group, version, kind, namespace, name, uid string) model.Provenance {
	return model.Provenance{
		SourceID: normalizer.source.ID, Environment: normalizer.source.Environment, ClusterRole: string(normalizer.source.Role), InstallationRole: normalizer.source.InstallationRole,
		Group: group, Version: version, Kind: kind, Namespace: namespace, Name: name, UID: uid,
		CollectedAt: normalizer.collectedAt, Freshness: model.FreshnessFresh, Access: model.AccessOK, Evidence: model.EvidenceObserved,
	}
}

func gardenerKind(resource string) (string, string, bool) {
	values := map[string][2]string{
		"shoots": {"Shoot", "shoot"}, "seeds": {"Seed", "seed"}, "cloudprofiles": {"CloudProfile", "cloud-profile"},
		"machinedeployments": {"McmMachineDeployment", "mcm-machine-deployment"}, "machinesets": {"McmMachineSet", "mcm-machine-set"}, "machines": {"McmMachine", "mcm-machine"},
	}
	value, ok := values[resource]
	return value[0], value[1], ok
}

func gardenerDetails(resource string, object map[string]any, providerID string) map[string]any {
	details := make(map[string]any)
	if providerID != "" {
		switch {
		case strings.HasPrefix(providerID, "ironcore-metal://"):
			details["providerType"] = "bare-metal"
		default:
			details["providerType"] = "virtual-machine"
		}
	}
	if resource == "shoots" {
		if providerType, found, _ := unstructured.NestedString(object, "spec", "provider", "type"); found {
			details["providerType"] = providerType
		}
		if _, found, _ := unstructured.NestedFieldNoCopy(object, "status", "access", "secretRef"); found {
			details["redactedReferences"] = []string{"redacted"}
		}
	}
	if len(details) == 0 {
		return nil
	}
	return details
}

func selectorAt(object map[string]any, path ...string) (map[string]string, []model.SelectorRequirement) {
	selector, found, _ := unstructured.NestedMap(object, path...)
	if !found {
		return nil, nil
	}
	wrapper := map[string]any{"spec": map[string]any{"networkInterfaceSelector": selector}}
	return normalizeSelector(wrapper)
}
