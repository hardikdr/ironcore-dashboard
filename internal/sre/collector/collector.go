package collector

import (
	"context"
	"errors"
	"fmt"

	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/model"
	"github.com/ironcore-dev/ironcore-dashboard/internal/sre/snapshot"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

type Collector interface {
	SourceID() string
	Collect(context.Context) (snapshot.Snapshot, error)
}

type ResourceReader interface {
	List(context.Context, schema.GroupVersionResource, string, metav1.ListOptions) (*unstructured.UnstructuredList, error)
}

type APIDiscovery interface {
	Supports(context.Context, schema.GroupVersionResource) (bool, error)
}

type ClientReader struct {
	Dynamic   dynamic.Interface
	Discovery discovery.DiscoveryInterface
}

func (reader ClientReader) List(ctx context.Context, gvr schema.GroupVersionResource, namespace string, options metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	resource := reader.Dynamic.Resource(gvr)
	if namespace != metav1.NamespaceAll {
		return resource.Namespace(namespace).List(ctx, options)
	}
	return resource.List(ctx, options)
}

func (reader ClientReader) Supports(_ context.Context, gvr schema.GroupVersionResource) (bool, error) {
	resources, err := reader.Discovery.ServerResourcesForGroupVersion(gvr.GroupVersion().String())
	if err != nil {
		return false, err
	}
	for _, resource := range resources.APIResources {
		if resource.Name == gvr.Resource {
			return true, nil
		}
	}
	return false, nil
}

type CollectionError struct {
	Class  snapshot.ErrorClass
	Access model.AccessState
	Err    error
}

func (err *CollectionError) Error() string {
	return fmt.Sprintf("source collection failed (%s)", err.Class)
}

func (err *CollectionError) Unwrap() error { return err.Err }

func classifyError(err error) (snapshot.ErrorClass, model.AccessState) {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return snapshot.ErrorTimeout, model.AccessUnreachable
	case apierrors.IsForbidden(err):
		return snapshot.ErrorRBAC, model.AccessDenied
	case apierrors.IsUnauthorized(err):
		return snapshot.ErrorAuthentication, model.AccessDenied
	case apierrors.IsNotFound(err):
		return snapshot.ErrorUnsupported, model.AccessUnsupported
	default:
		return snapshot.ErrorTransport, model.AccessUnreachable
	}
}
