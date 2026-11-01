package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	versioned "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NetworkPolicyHandler serves NetworkPolicy list per namespace.
type NetworkPolicyHandler struct{ cs versioned.Interface }

func NewNetworkPolicyHandler(cs versioned.Interface) *NetworkPolicyHandler {
	return &NetworkPolicyHandler{cs: cs}
}

func (h *NetworkPolicyHandler) List(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	list, err := h.cs.NetworkingV1alpha1().NetworkPolicies(ns).List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]NetworkPolicyResponse, 0, len(list.Items))
	for _, np := range list.Items {
		pt := make([]string, 0, len(np.Spec.PolicyTypes))
		for _, p := range np.Spec.PolicyTypes {
			pt = append(pt, string(p))
		}
		resp = append(resp, NetworkPolicyResponse{
			Name:         np.Name,
			Namespace:    np.Namespace,
			NetworkRef:   np.Spec.NetworkRef.Name,
			PolicyTypes:  pt,
			IngressRules: len(np.Spec.Ingress),
			EgressRules:  len(np.Spec.Egress),
			CreatedAt:    np.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

// BucketHandler serves Bucket list per namespace.
type BucketHandler struct{ cs versioned.Interface }

func NewBucketHandler(cs versioned.Interface) *BucketHandler {
	return &BucketHandler{cs: cs}
}

func (h *BucketHandler) List(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	list, err := h.cs.StorageV1alpha1().Buckets(ns).List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]BucketResponse, 0, len(list.Items))
	for _, b := range list.Items {
		classRef, poolRef, endpoint := "", "", ""
		if b.Spec.BucketClassRef != nil {
			classRef = b.Spec.BucketClassRef.Name
		}
		if b.Spec.BucketPoolRef != nil {
			poolRef = b.Spec.BucketPoolRef.Name
		}
		if b.Status.Access != nil {
			endpoint = b.Status.Access.Endpoint
		}
		resp = append(resp, BucketResponse{
			Name:           b.Name,
			Namespace:      b.Namespace,
			BucketClassRef: classRef,
			BucketPoolRef:  poolRef,
			State:          string(b.Status.State),
			Endpoint:       endpoint,
			CreatedAt:      b.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

// BucketClassHandler serves BucketClass list (cluster-scoped).
type BucketClassHandler struct{ cs versioned.Interface }

func NewBucketClassHandler(cs versioned.Interface) *BucketClassHandler {
	return &BucketClassHandler{cs: cs}
}

func (h *BucketClassHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.cs.StorageV1alpha1().BucketClasses().List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]BucketClassResponse, 0, len(list.Items))
	for _, bc := range list.Items {
		caps := map[string]string{}
		for k, v := range bc.Capabilities {
			caps[string(k)] = v.String()
		}
		resp = append(resp, BucketClassResponse{
			Name:         bc.Name,
			Capabilities: caps,
			CreatedAt:    bc.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

// BucketPoolHandler serves BucketPool list (cluster-scoped).
type BucketPoolHandler struct{ cs versioned.Interface }

func NewBucketPoolHandler(cs versioned.Interface) *BucketPoolHandler {
	return &BucketPoolHandler{cs: cs}
}

func (h *BucketPoolHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.cs.StorageV1alpha1().BucketPools().List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]BucketPoolResponse, 0, len(list.Items))
	for _, bp := range list.Items {
		classes := make([]string, 0, len(bp.Status.AvailableBucketClasses))
		for _, c := range bp.Status.AvailableBucketClasses {
			classes = append(classes, c.Name)
		}
		resp = append(resp, BucketPoolResponse{
			Name:                   bp.Name,
			State:                  string(bp.Status.State),
			AvailableBucketClasses: classes,
			CreatedAt:              bp.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}
