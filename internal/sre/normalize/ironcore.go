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

type IronCore struct {
	source      config.Source
	collectedAt time.Time
}

func NewIronCore(source config.Source, collectedAt time.Time) *IronCore {
	return &IronCore{source: source, collectedAt: collectedAt.UTC()}
}

func (normalizer *IronCore) Normalize(gvr schema.GroupVersionResource, object *unstructured.Unstructured) (model.Resource, error) {
	kind, resourceType, ok := ironCoreKind(gvr.Resource)
	if !ok {
		return model.Resource{}, fmt.Errorf("unsupported IronCore resource %s", gvr.Resource)
	}
	key, err := model.NewResourceKey(model.ResourceIdentity{
		SourceID: normalizer.source.ID, Group: gvr.Group, Kind: kind,
		Namespace: object.GetNamespace(), Name: object.GetName(), UID: string(object.GetUID()),
	})
	if err != nil {
		return model.Resource{}, err
	}
	resource := model.Resource{
		Key: key, Type: resourceType, Labels: object.GetLabels(), Annotations: safeAnnotations(object.GetAnnotations()),
		CreatedAt: object.GetCreationTimestamp().Time.UTC(),
		Provenance: model.Provenance{
			SourceID: normalizer.source.ID, Environment: normalizer.source.Environment, ClusterRole: string(normalizer.source.Role),
			Group: gvr.Group, Version: gvr.Version, Kind: kind, Namespace: object.GetNamespace(), Name: object.GetName(), UID: string(object.GetUID()),
			CollectedAt: normalizer.collectedAt, Freshness: model.FreshnessFresh, Access: model.AccessOK, Evidence: model.EvidenceObserved,
		},
		Health: model.HealthUnknown,
	}
	resource.State, _, _ = unstructured.NestedString(object.Object, "status", "state")
	resource.ProviderID = firstNestedString(object.Object,
		[]string{"spec", "providerID"}, []string{"status", "machineID"}, []string{"status", "volumeID"})
	resource.Conditions = normalizeConditions(object.Object)
	resource.Health = healthFrom(resource.State, resource.Conditions)
	resource.OwnerRefs = normalizeOwners(object)
	resource.Addresses = normalizeAddresses(object.Object)
	resource.References = normalizeReferences(gvr.Resource, object.Object, object.GetNamespace())
	resource.Selector, resource.SelectorExpressions = normalizeSelector(object.Object)
	resource.Sanitized = normalizeSafeDetails(gvr.Resource, object.Object)
	return resource, nil
}

func ironCoreKind(resource string) (string, string, bool) {
	values := map[string][2]string{
		"machines": {"IronCoreVM", "ironcore-vm"}, "machinepools": {"MachinePool", "machine-pool"}, "machineclasses": {"MachineClass", "machine-class"},
		"volumes": {"Volume", "volume"}, "volumepools": {"VolumePool", "volume-pool"}, "volumeclasses": {"VolumeClass", "volume-class"}, "volumesnapshots": {"VolumeSnapshot", "volume-snapshot"},
		"networks": {"Network", "network"}, "networkinterfaces": {"Nic", "nic"}, "loadbalancers": {"LoadBalancer", "load-balancer"}, "loadbalancerroutings": {"LoadBalancerRouting", "load-balancer-routing"},
		"natgateways": {"NATGateway", "nat-gateway"}, "virtualips": {"VirtualIP", "virtual-ip"}, "networkpolicies": {"NetworkPolicy", "network-policy"},
		"prefixes": {"Prefix", "prefix"}, "prefixallocations": {"PrefixAllocation", "prefix-allocation"},
	}
	value, ok := values[resource]
	return value[0], value[1], ok
}

func normalizeReferences(resource string, object map[string]any, namespace string) []model.ObjectReference {
	var refs []model.ObjectReference
	add := func(group, kind, name, uid, fieldPath string) {
		if name != "" {
			refs = append(refs, model.ObjectReference{Group: group, Kind: kind, Namespace: namespace, Name: name, UID: uid, FieldPath: fieldPath})
		}
	}
	addNested := func(group, kind, fieldPath string, fields ...string) {
		name, _, _ := unstructured.NestedString(object, fields...)
		add(group, kind, name, "", fieldPath)
	}
	switch resource {
	case "machines":
		addNested("compute.ironcore.dev", "MachinePool", ".spec.machinePoolRef.name", "spec", "machinePoolRef", "name")
		addNested("compute.ironcore.dev", "MachineClass", ".spec.machineClassRef.name", "spec", "machineClassRef", "name")
		addArrayRefs(&refs, object, "spec", "networkInterfaces", "networkInterfaceRef", "networking.ironcore.dev", "Nic", namespace, ".spec.networkInterfaces")
		addArrayRefs(&refs, object, "status", "networkInterfaces", "networkInterfaceRef", "networking.ironcore.dev", "Nic", namespace, ".status.networkInterfaces")
		addArrayRefs(&refs, object, "spec", "volumes", "volumeRef", "storage.ironcore.dev", "Volume", namespace, ".spec.volumes")
		addArrayRefs(&refs, object, "status", "volumes", "volumeRef", "storage.ironcore.dev", "Volume", namespace, ".status.volumes")
	case "networkinterfaces":
		addNested("networking.ironcore.dev", "Network", ".spec.networkRef.name", "spec", "networkRef", "name")
		name, _, _ := unstructured.NestedString(object, "spec", "machineRef", "name")
		uid, _, _ := unstructured.NestedString(object, "spec", "machineRef", "uid")
		add("compute.ironcore.dev", "IronCoreVM", name, uid, ".spec.machineRef")
	case "loadbalancers", "natgateways":
		addNested("networking.ironcore.dev", "Network", ".spec.networkRef.name", "spec", "networkRef", "name")
	case "volumes":
		addNested("storage.ironcore.dev", "VolumePool", ".spec.volumePoolRef.name", "spec", "volumePoolRef", "name")
		addNested("storage.ironcore.dev", "VolumeClass", ".spec.volumeClassRef.name", "spec", "volumeClassRef", "name")
		addNested("storage.ironcore.dev", "VolumeSnapshot", ".spec.dataSource.volumeSnapshotRef.name", "spec", "dataSource", "volumeSnapshotRef", "name")
	}
	return refs
}

func addArrayRefs(refs *[]model.ObjectReference, object map[string]any, section, array, reference, group, kind, namespace, fieldPath string) {
	items, found, _ := unstructured.NestedSlice(object, section, array)
	if !found {
		return
	}
	for _, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		ref, ok := item[reference].(map[string]any)
		if !ok {
			continue
		}
		name, _ := ref["name"].(string)
		uid, _ := ref["uid"].(string)
		if name != "" {
			*refs = append(*refs, model.ObjectReference{Group: group, Kind: kind, Namespace: namespace, Name: name, UID: uid, FieldPath: fieldPath})
		}
	}
}

func normalizeSelector(object map[string]any) (map[string]string, []model.SelectorRequirement) {
	selector, found, _ := unstructured.NestedMap(object, "spec", "networkInterfaceSelector")
	if !found {
		return nil, nil
	}
	labels := make(map[string]string)
	if rawLabels, ok := selector["matchLabels"].(map[string]any); ok {
		for key, value := range rawLabels {
			if text, ok := value.(string); ok {
				labels[key] = text
			}
		}
	}
	var expressions []model.SelectorRequirement
	if rawExpressions, ok := selector["matchExpressions"].([]any); ok {
		for _, raw := range rawExpressions {
			expression, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			requirement := model.SelectorRequirement{Key: stringValue(expression["key"]), Operator: stringValue(expression["operator"])}
			if values, ok := expression["values"].([]any); ok {
				for _, value := range values {
					if text, ok := value.(string); ok {
						requirement.Values = append(requirement.Values, text)
					}
				}
			}
			expressions = append(expressions, requirement)
		}
	}
	return labels, expressions
}

func normalizeConditions(object map[string]any) []model.Condition {
	rawConditions, found, _ := unstructured.NestedSlice(object, "status", "conditions")
	if !found {
		return nil
	}
	conditions := make([]model.Condition, 0, len(rawConditions))
	for _, raw := range rawConditions {
		condition, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		result := model.Condition{Type: stringValue(condition["type"]), Status: stringValue(condition["status"]), Reason: stringValue(condition["reason"]), Message: stringValue(condition["message"])}
		if timestamp := stringValue(condition["lastTransitionTime"]); timestamp != "" {
			result.LastTransitionTime, _ = time.Parse(time.RFC3339, timestamp)
		}
		conditions = append(conditions, result)
	}
	return conditions
}

func normalizeOwners(object *unstructured.Unstructured) []model.ObjectReference {
	owners := object.GetOwnerReferences()
	result := make([]model.ObjectReference, 0, len(owners))
	for _, owner := range owners {
		group := strings.SplitN(owner.APIVersion, "/", 2)[0]
		result = append(result, model.ObjectReference{Group: group, Kind: owner.Kind, Namespace: object.GetNamespace(), Name: owner.Name, UID: string(owner.UID), FieldPath: ".metadata.ownerReferences"})
	}
	return result
}

func normalizeAddresses(object map[string]any) []string {
	values, found, _ := unstructured.NestedSlice(object, "status", "ips")
	if !found {
		return nil
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if text, ok := value.(string); ok {
			result = append(result, text)
		}
	}
	return result
}

func normalizeSafeDetails(resource string, object map[string]any) map[string]any {
	details := make(map[string]any)
	var redacted []string
	for _, item := range []struct {
		path   []string
		marker string
	}{
		{[]string{"spec", "ignitionRef"}, "redacted"},
		{[]string{"spec", "imagePullSecretRef"}, "redacted"},
		{[]string{"spec", "encryption", "secretRef"}, "redacted"},
		{[]string{"status", "access", "secretRef"}, "redacted"},
	} {
		if _, found, _ := unstructured.NestedFieldNoCopy(object, item.path...); found {
			redacted = append(redacted, item.marker)
		}
	}
	if len(redacted) != 0 {
		details["redactedReferences"] = redacted
	}
	if resource == "machines" {
		if power, found, _ := unstructured.NestedString(object, "spec", "power"); found {
			details["power"] = power
		}
	}
	if len(details) == 0 {
		return nil
	}
	return details
}

func firstNestedString(object map[string]any, paths ...[]string) string {
	for _, path := range paths {
		if value, found, _ := unstructured.NestedString(object, path...); found {
			return value
		}
	}
	return ""
}

func healthFrom(state string, conditions []model.Condition) model.HealthState {
	for _, condition := range conditions {
		if condition.Type == "Ready" {
			switch condition.Status {
			case "True":
				return model.HealthHealthy
			case "False":
				return model.HealthDegraded
			}
		}
	}
	switch strings.ToLower(state) {
	case "running", "available", "ready":
		return model.HealthHealthy
	case "error", "failed", "terminated":
		return model.HealthFailed
	default:
		return model.HealthUnknown
	}
}

func safeAnnotations(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	result := make(map[string]string)
	for key, value := range input {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "credential") || strings.Contains(lower, "ignition") {
			continue
		}
		result[key] = value
	}
	return result
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}
