package collector

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/config"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/snapshot"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	nodesGVR        = schema.GroupVersionResource{Version: "v1", Resource: "nodes"}
	podsGVR         = schema.GroupVersionResource{Version: "v1", Resource: "pods"}
	eventsGVR       = schema.GroupVersionResource{Version: "v1", Resource: "events"}
	deploymentsGVR  = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	daemonSetsGVR   = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonsets"}
	capiMachinesGVR = schema.GroupVersionResource{Group: "cluster.x-k8s.io", Version: "v1beta1", Resource: "machines"}
	mcmMachinesGVR  = schema.GroupVersionResource{Group: "machine.sapcloud.io", Version: "v1alpha1", Resource: "machines"}
	metalServersGVR = schema.GroupVersionResource{Group: "metal.ironcore.dev", Version: "v1alpha1", Resource: "servers"}
)

type ZonalOptions struct {
	PageSize       int64
	RequestTimeout time.Duration
	Now            func() time.Time
}

type ZonalCollector struct {
	source    config.Source
	reader    ResourceReader
	discovery APIDiscovery
	options   ZonalOptions
}

func NewZonalCollector(source config.Source, reader ResourceReader, discovery APIDiscovery, options ZonalOptions) *ZonalCollector {
	if options.PageSize <= 0 {
		options.PageSize = 250
	}
	if options.RequestTimeout <= 0 {
		options.RequestTimeout = 15 * time.Second
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &ZonalCollector{source: source, reader: reader, discovery: discovery, options: options}
}

func (collector *ZonalCollector) SourceID() string { return collector.source.ID }

func (collector *ZonalCollector) Collect(ctx context.Context) (snapshot.Snapshot, error) {
	collectedAt := collector.options.Now().UTC()
	result := snapshot.Snapshot{
		SourceID:     collector.source.ID,
		CollectedAt:  collectedAt,
		ObjectCounts: make(map[string]int),
	}

	type inventory struct {
		gvr       schema.GroupVersionResource
		namespace string
	}
	required := []inventory{{gvr: nodesGVR}, {gvr: capiMachinesGVR}}
	for _, namespace := range collector.source.ProviderNamespaces {
		for _, gvr := range []schema.GroupVersionResource{podsGVR, eventsGVR, deploymentsGVR, daemonSetsGVR} {
			required = append(required, inventory{gvr: gvr, namespace: namespace})
		}
	}

	successes := 0
	var lastRequiredErr error
	for _, item := range required {
		objects, err := collector.listAll(ctx, item.gvr, item.namespace)
		if err != nil {
			lastRequiredErr = err
			class, access := classifyError(err)
			result.Partial = true
			result.Gaps = append(result.Gaps, snapshot.AccessGap{Resource: resourceName(item.gvr, item.namespace), Access: access, Message: string(class)})
			continue
		}
		successes++
		collector.appendNormalized(&result, item.gvr, objects, collectedAt)
	}
	if successes == 0 {
		class, access := classifyError(lastRequiredErr)
		return snapshot.Snapshot{}, &CollectionError{Class: class, Access: access, Err: lastRequiredErr}
	}

	for _, gvr := range []schema.GroupVersionResource{mcmMachinesGVR, metalServersGVR} {
		supported, err := collector.discovery.Supports(ctx, gvr)
		if err != nil {
			class, access := classifyError(err)
			result.Partial = true
			result.Gaps = append(result.Gaps, snapshot.AccessGap{Resource: gvr.String(), Access: access, Message: string(class)})
			continue
		}
		if !supported {
			result.Partial = true
			result.Gaps = append(result.Gaps, snapshot.AccessGap{Resource: gvr.String(), Access: model.AccessUnsupported, Message: "API is not served"})
			continue
		}
		objects, err := collector.listAll(ctx, gvr, metav1.NamespaceAll)
		if err != nil {
			class, access := classifyError(err)
			result.Partial = true
			result.Gaps = append(result.Gaps, snapshot.AccessGap{Resource: gvr.String(), Access: access, Message: string(class)})
			continue
		}
		collector.appendNormalized(&result, gvr, objects, collectedAt)
	}

	sort.Slice(result.Resources, func(i, j int) bool { return result.Resources[i].Key < result.Resources[j].Key })
	return result, nil
}

func (collector *ZonalCollector) listAll(ctx context.Context, gvr schema.GroupVersionResource, namespace string) ([]unstructured.Unstructured, error) {
	var objects []unstructured.Unstructured
	continueValue := ""
	for {
		requestCtx, cancel := context.WithTimeout(ctx, collector.options.RequestTimeout)
		list, err := collector.reader.List(requestCtx, gvr, namespace, metav1.ListOptions{Limit: collector.options.PageSize, Continue: continueValue})
		cancel()
		if err != nil {
			return nil, err
		}
		objects = append(objects, list.Items...)
		if list.GetContinue() == "" {
			return objects, nil
		}
		continueValue = list.GetContinue()
	}
}

func (collector *ZonalCollector) appendNormalized(result *snapshot.Snapshot, gvr schema.GroupVersionResource, objects []unstructured.Unstructured, collectedAt time.Time) {
	for index := range objects {
		resource, ok := collector.normalize(gvr, &objects[index], collectedAt)
		if !ok {
			result.Partial = true
			result.Warnings = append(result.Warnings, "object with invalid identity was omitted")
			continue
		}
		result.Resources = append(result.Resources, resource)
		result.ObjectCounts[resource.Provenance.Kind]++
	}
}

func (collector *ZonalCollector) normalize(gvr schema.GroupVersionResource, object *unstructured.Unstructured, collectedAt time.Time) (model.Resource, bool) {
	group := gvr.Group
	if group == "" {
		group = "core"
	}
	kind, resourceType := normalizedZonalKind(gvr)
	identity := model.ResourceIdentity{
		SourceID:  collector.source.ID,
		Group:     group,
		Kind:      kind,
		Namespace: object.GetNamespace(),
		Name:      object.GetName(),
		UID:       string(object.GetUID()),
	}
	key, err := model.NewResourceKey(identity)
	if err != nil {
		return model.Resource{}, false
	}
	labels := object.GetLabels()
	resource := model.Resource{
		Key: key, Type: resourceType, Labels: labels,
		Provenance: model.Provenance{
			SourceID: collector.source.ID, Environment: collector.source.Environment, ClusterRole: string(collector.source.Role),
			Group: group, Version: gvr.Version, Kind: kind, Namespace: object.GetNamespace(), Name: object.GetName(), UID: string(object.GetUID()),
			CollectedAt: collectedAt, Freshness: model.FreshnessFresh, Access: model.AccessOK, Evidence: model.EvidenceObserved,
		},
		Health: model.HealthUnknown,
	}
	if gvr == nodesGVR {
		role := physicalRole(labels)
		resource.Sanitized = map[string]any{"physicalRole": role}
		if readyCondition(object) == "True" {
			resource.Health = model.HealthHealthy
		}
	}
	if phase, found, _ := unstructured.NestedString(object.Object, "status", "phase"); found {
		resource.State = phase
	}
	return resource, true
}

func normalizedZonalKind(gvr schema.GroupVersionResource) (string, string) {
	switch gvr {
	case nodesGVR:
		return "PhysicalNode", "physical-node"
	case capiMachinesGVR:
		return "CapiMachine", "capi-machine"
	case mcmMachinesGVR:
		return "McmMachine", "mcm-machine"
	case metalServersGVR:
		return "PhysicalNode", "physical-server"
	case eventsGVR:
		return "KubernetesEvent", "kubernetes-event"
	default:
		return "ProviderComponent", "provider-component"
	}
}

func physicalRole(labels map[string]string) string {
	if role := labels["ironcore.dev/role"]; role != "" {
		return role
	}
	for _, role := range []string{"management", "control-plane", "compute", "storage"} {
		if _, ok := labels["node-role.ironcore.dev/"+role]; ok {
			return role
		}
		if _, ok := labels["node-role.kubernetes.io/"+role]; ok {
			if role == "control-plane" {
				return "management"
			}
			return role
		}
	}
	return "unknown"
}

func readyCondition(object *unstructured.Unstructured) string {
	conditions, found, _ := unstructured.NestedSlice(object.Object, "status", "conditions")
	if !found {
		return ""
	}
	for _, raw := range conditions {
		condition, ok := raw.(map[string]any)
		if ok && condition["type"] == "Ready" {
			status, _ := condition["status"].(string)
			return status
		}
	}
	return ""
}

func resourceName(gvr schema.GroupVersionResource, namespace string) string {
	if namespace == "" {
		return gvr.String()
	}
	return fmt.Sprintf("%s in namespace %s", gvr.String(), namespace)
}
