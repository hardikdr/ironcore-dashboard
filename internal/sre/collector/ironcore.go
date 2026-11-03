package collector

import (
	"context"
	"sort"
	"time"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/config"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/normalize"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/snapshot"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	ironMachinesGVR         = ironGVR("compute.ironcore.dev", "machines")
	machinePoolsGVR         = ironGVR("compute.ironcore.dev", "machinepools")
	machineClassesGVR       = ironGVR("compute.ironcore.dev", "machineclasses")
	volumesGVR              = ironGVR("storage.ironcore.dev", "volumes")
	volumePoolsGVR          = ironGVR("storage.ironcore.dev", "volumepools")
	volumeClassesGVR        = ironGVR("storage.ironcore.dev", "volumeclasses")
	volumeSnapshotsGVR      = ironGVR("storage.ironcore.dev", "volumesnapshots")
	networksGVR             = ironGVR("networking.ironcore.dev", "networks")
	nicsGVR                 = ironGVR("networking.ironcore.dev", "networkinterfaces")
	loadBalancersGVR        = ironGVR("networking.ironcore.dev", "loadbalancers")
	loadBalancerRoutingsGVR = ironGVR("networking.ironcore.dev", "loadbalancerroutings")
	natGatewaysGVR          = ironGVR("networking.ironcore.dev", "natgateways")
	virtualIPsGVR           = ironGVR("networking.ironcore.dev", "virtualips")
	networkPoliciesGVR      = ironGVR("networking.ironcore.dev", "networkpolicies")
	prefixesGVR             = ironGVR("ipam.ironcore.dev", "prefixes")
	prefixAllocationsGVR    = ironGVR("ipam.ironcore.dev", "prefixallocations")
)

type IronCoreOptions struct {
	PageSize       int64
	RequestTimeout time.Duration
	Now            func() time.Time
}

type IronCoreCollector struct {
	source  config.Source
	reader  ResourceReader
	options IronCoreOptions
}

func NewIronCoreCollector(source config.Source, reader ResourceReader, options IronCoreOptions) *IronCoreCollector {
	if options.PageSize <= 0 {
		options.PageSize = 250
	}
	if options.RequestTimeout <= 0 {
		options.RequestTimeout = 15 * time.Second
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &IronCoreCollector{source: source, reader: reader, options: options}
}

func (collector *IronCoreCollector) SourceID() string { return collector.source.ID }

func (collector *IronCoreCollector) Collect(ctx context.Context) (snapshot.Snapshot, error) {
	collectedAt := collector.options.Now().UTC()
	result := snapshot.Snapshot{SourceID: collector.source.ID, CollectedAt: collectedAt, ObjectCounts: make(map[string]int)}
	normalizer := normalize.NewIronCore(collector.source, collectedAt)
	successes := 0
	var lastErr error

	for _, group := range canonicalIronCoreGroups() {
		for _, gvr := range group {
			objects, err := listPaginated(ctx, collector.reader, gvr, metav1.NamespaceAll, collector.options.PageSize, collector.options.RequestTimeout)
			if err != nil {
				lastErr = err
				class, access := classifyError(err)
				result.Partial = true
				result.Gaps = append(result.Gaps, snapshot.AccessGap{Resource: gvr.Group, Access: access, Message: string(class)})
				break
			}
			successes++
			for index := range objects {
				resource, err := normalizer.Normalize(gvr, &objects[index])
				if err != nil {
					result.Partial = true
					result.Warnings = append(result.Warnings, "object with invalid identity was omitted")
					continue
				}
				result.Resources = append(result.Resources, resource)
				result.ObjectCounts[resource.Provenance.Kind]++
			}
		}
	}
	if successes == 0 {
		class, access := classifyError(lastErr)
		return snapshot.Snapshot{}, &CollectionError{Class: class, Access: access, Err: lastErr}
	}
	sort.Slice(result.Resources, func(i, j int) bool { return result.Resources[i].Key < result.Resources[j].Key })
	return result, nil
}

func canonicalIronCoreGroups() [][]schema.GroupVersionResource {
	return [][]schema.GroupVersionResource{
		{ironMachinesGVR, machinePoolsGVR, machineClassesGVR},
		{volumesGVR, volumePoolsGVR, volumeClassesGVR, volumeSnapshotsGVR},
		{networksGVR, nicsGVR, loadBalancersGVR, loadBalancerRoutingsGVR, natGatewaysGVR, virtualIPsGVR, networkPoliciesGVR},
		{prefixesGVR, prefixAllocationsGVR},
	}
}

func canonicalIronCoreInventory() []schema.GroupVersionResource {
	var result []schema.GroupVersionResource
	for _, group := range canonicalIronCoreGroups() {
		result = append(result, group...)
	}
	return result
}

func ironGVR(group, resource string) schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: group, Version: "v1alpha1", Resource: resource}
}

func listPaginated(ctx context.Context, reader ResourceReader, gvr schema.GroupVersionResource, namespace string, pageSize int64, timeout time.Duration) ([]unstructured.Unstructured, error) {
	var objects []unstructured.Unstructured
	continueValue := ""
	for {
		requestCtx, cancel := context.WithTimeout(ctx, timeout)
		list, err := reader.List(requestCtx, gvr, namespace, metav1.ListOptions{Limit: pageSize, Continue: continueValue})
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
