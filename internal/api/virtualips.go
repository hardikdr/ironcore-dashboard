package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	versioned "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	metav1    "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type VirtualIPHandler struct{ cs versioned.Interface }

func NewVirtualIPHandler(cs versioned.Interface) *VirtualIPHandler {
	return &VirtualIPHandler{cs: cs}
}

func (h *VirtualIPHandler) List(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	list, err := h.cs.NetworkingV1alpha1().VirtualIPs(ns).List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]VirtualIPResponse, 0, len(list.Items))
	for _, v := range list.Items {
		ip := ""
		if v.Status.IP != nil {
			ip = v.Status.IP.String()
		}
		resp = append(resp, VirtualIPResponse{
			Name:      v.Name,
			Namespace: v.Namespace,
			IP:        ip,
			Type:      string(v.Spec.Type),
			IPFamily:  string(v.Spec.IPFamily),
			CreatedAt: v.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}
