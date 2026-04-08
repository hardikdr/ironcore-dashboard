package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	versioned "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	metav1    "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NetworkHandler struct{ cs versioned.Interface }

func NewNetworkHandler(cs versioned.Interface) *NetworkHandler { return &NetworkHandler{cs: cs} }

func (h *NetworkHandler) ListNetworks(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	list, err := h.cs.NetworkingV1alpha1().Networks(ns).List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]NetworkResponse, 0, len(list.Items))
	for _, n := range list.Items {
		resp = append(resp, NetworkResponse{
			Name:      n.Name,
			Namespace: n.Namespace,
			CreatedAt: n.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *NetworkHandler) ListNetworkInterfaces(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	list, err := h.cs.NetworkingV1alpha1().NetworkInterfaces(ns).List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]NetworkInterfaceResponse, 0, len(list.Items))
	for _, ni := range list.Items {
		ips := []string{}
		for _, ip := range ni.Status.IPs {
			ips = append(ips, ip.String())
		}
		machine := ""
		if ni.Spec.MachineRef != nil {
			machine = ni.Spec.MachineRef.Name
		}
		resp = append(resp, NetworkInterfaceResponse{
			Name:      ni.Name,
			Namespace: ni.Namespace,
			State:     string(ni.Status.State),
			IPs:       ips,
			Network:   ni.Spec.NetworkRef.Name,
			Machine:   machine,
			CreatedAt: ni.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}
