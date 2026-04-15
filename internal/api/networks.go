package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	versioned           "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	metav1              "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (h *NetworkHandler) GetNetwork(w http.ResponseWriter, r *http.Request) {
	ns, name := chi.URLParam(r, "ns"), chi.URLParam(r, "name")
	n, err := h.cs.NetworkingV1alpha1().Networks(ns).Get(r.Context(), name, metav1.GetOptions{})
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	peerings := make([]NetworkPeeringStatus, 0, len(n.Status.Peerings))
	for _, p := range n.Status.Peerings {
		peerings = append(peerings, NetworkPeeringStatus{
			Name:  p.Name,
			State: string(p.State),
		})
	}
	writeJSON(w, http.StatusOK, NetworkDetailResponse{
		Name:       n.Name,
		Namespace:  n.Namespace,
		State:      string(n.Status.State),
		ProviderID: n.Spec.ProviderID,
		Peerings:   peerings,
		CreatedAt:  n.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

func (h *NetworkHandler) CreateNetwork(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	var req CreateNetworkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	net := &networkingv1alpha1.Network{
		ObjectMeta: metav1.ObjectMeta{Name: req.Name, Namespace: ns},
	}
	created, err := h.cs.NetworkingV1alpha1().Networks(ns).Create(r.Context(), net, metav1.CreateOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, NetworkResponse{
		Name:      created.Name,
		Namespace: created.Namespace,
		CreatedAt: created.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

func (h *NetworkHandler) DeleteNetwork(w http.ResponseWriter, r *http.Request) {
	ns, name := chi.URLParam(r, "ns"), chi.URLParam(r, "name")
	if err := h.cs.NetworkingV1alpha1().Networks(ns).Delete(r.Context(), name, metav1.DeleteOptions{}); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
