package collector

import (
	"context"
	"sort"
	"time"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/config"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/normalize"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/snapshot"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	shootsGVR         = schema.GroupVersionResource{Group: "core.gardener.cloud", Version: "v1beta1", Resource: "shoots"}
	seedsGVR          = schema.GroupVersionResource{Group: "core.gardener.cloud", Version: "v1beta1", Resource: "seeds"}
	cloudProfilesGVR  = schema.GroupVersionResource{Group: "core.gardener.cloud", Version: "v1beta1", Resource: "cloudprofiles"}
	mcmDeploymentsGVR = schema.GroupVersionResource{Group: "machine.sapcloud.io", Version: "v1alpha1", Resource: "machinedeployments"}
	mcmSetsGVR        = schema.GroupVersionResource{Group: "machine.sapcloud.io", Version: "v1alpha1", Resource: "machinesets"}
)

type GardenerOptions struct {
	PageSize       int64
	RequestTimeout time.Duration
	Now            func() time.Time
}

type GardenerCollector struct {
	source    config.Source
	reader    ResourceReader
	discovery APIDiscovery
	options   GardenerOptions
}

func NewGardenerCollector(source config.Source, reader ResourceReader, discovery APIDiscovery, options GardenerOptions) *GardenerCollector {
	if options.PageSize <= 0 {
		options.PageSize = 250
	}
	if options.RequestTimeout <= 0 {
		options.RequestTimeout = 15 * time.Second
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &GardenerCollector{source: source, reader: reader, discovery: discovery, options: options}
}

func (collector *GardenerCollector) SourceID() string { return collector.source.ID }

func (collector *GardenerCollector) Collect(ctx context.Context) (snapshot.Snapshot, error) {
	collectedAt := collector.options.Now().UTC()
	result := snapshot.Snapshot{SourceID: collector.source.ID, CollectedAt: collectedAt, ObjectCounts: make(map[string]int)}
	normalizer := normalize.NewGardener(collector.source, collectedAt)
	installation, err := normalizer.Installation()
	if err != nil {
		return snapshot.Snapshot{}, err
	}
	result.Resources = append(result.Resources, installation)
	result.ObjectCounts[installation.Provenance.Kind]++

	reachable := false
	var lastTransportErr error
	for _, gvr := range canonicalGardenerInventory() {
		supported, err := collector.discovery.Supports(ctx, gvr)
		if err != nil {
			class, access := classifyError(err)
			if access == model.AccessDenied || access == model.AccessUnsupported {
				reachable = true
			} else {
				lastTransportErr = err
			}
			result.Partial = true
			result.Gaps = append(result.Gaps, snapshot.AccessGap{Resource: gvr.String(), Access: access, Message: string(class)})
			continue
		}
		reachable = true
		if !supported {
			result.Partial = true
			result.Gaps = append(result.Gaps, snapshot.AccessGap{Resource: gvr.String(), Access: model.AccessUnsupported, Message: "API is not served"})
			continue
		}
		objects, err := listPaginated(ctx, collector.reader, gvr, metav1.NamespaceAll, collector.options.PageSize, collector.options.RequestTimeout)
		if err != nil {
			class, access := classifyError(err)
			result.Partial = true
			result.Gaps = append(result.Gaps, snapshot.AccessGap{Resource: gvr.String(), Access: access, Message: string(class)})
			continue
		}
		for index := range objects {
			resources, err := normalizer.Normalize(gvr, &objects[index])
			if err != nil {
				result.Partial = true
				result.Warnings = append(result.Warnings, "object with invalid identity was omitted")
				continue
			}
			for _, resource := range resources {
				result.Resources = append(result.Resources, resource)
				result.ObjectCounts[resource.Provenance.Kind]++
			}
		}
	}
	if !reachable && lastTransportErr != nil {
		class, access := classifyError(lastTransportErr)
		return snapshot.Snapshot{}, &CollectionError{Class: class, Access: access, Err: lastTransportErr}
	}
	sort.Slice(result.Resources, func(i, j int) bool { return result.Resources[i].Key < result.Resources[j].Key })
	return result, nil
}

func canonicalGardenerInventory() []schema.GroupVersionResource {
	return []schema.GroupVersionResource{shootsGVR, seedsGVR, cloudProfilesGVR, mcmDeploymentsGVR, mcmSetsGVR, mcmMachinesGVR}
}
