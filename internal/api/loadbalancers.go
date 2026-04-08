package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	versioned "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	metav1    "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LoadBalancerHandler struct{ cs versioned.Interface }

func NewLoadBalancerHandler(cs versioned.Interface) *LoadBalancerHandler {
	return &LoadBalancerHandler{cs: cs}
}

func (h *LoadBalancerHandler) List(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	list, err := h.cs.NetworkingV1alpha1().LoadBalancers(ns).List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]LoadBalancerResponse, 0, len(list.Items))
	for _, lb := range list.Items {
		ips := []string{}
		for _, ip := range lb.Status.IPs {
			ips = append(ips, ip.String())
		}
		resp = append(resp, LoadBalancerResponse{
			Name:      lb.Name,
			Namespace: lb.Namespace,
			Type:      string(lb.Spec.Type),
			IPs:       ips,
			CreatedAt: lb.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}
