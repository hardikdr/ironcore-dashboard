package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	versioned "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			Name: v.Name, Namespace: v.Namespace,
			IP: ip, Type: string(v.Spec.Type), IPFamily: string(v.Spec.IPFamily),
			CreatedAt: v.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *VirtualIPHandler) Get(w http.ResponseWriter, r *http.Request) {
	ns, name := chi.URLParam(r, "ns"), chi.URLParam(r, "name")
	v, err := h.cs.NetworkingV1alpha1().VirtualIPs(ns).Get(r.Context(), name, metav1.GetOptions{})
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	ip := ""
	if v.Status.IP != nil {
		ip = v.Status.IP.String()
	}
	targetRef := ""
	if v.Spec.TargetRef != nil {
		targetRef = v.Spec.TargetRef.Name
	}
	writeJSON(w, http.StatusOK, VirtualIPDetailResponse{
		Name: v.Name, Namespace: v.Namespace,
		Type: string(v.Spec.Type), IPFamily: string(v.Spec.IPFamily),
		IP: ip, TargetRef: targetRef,
		CreatedAt: v.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

func (h *VirtualIPHandler) Create(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	var req CreateVirtualIPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	vip := &networkingv1alpha1.VirtualIP{
		ObjectMeta: metav1.ObjectMeta{Name: req.Name, Namespace: ns},
		Spec: networkingv1alpha1.VirtualIPSpec{
			Type:     networkingv1alpha1.VirtualIPType(req.Type),
			IPFamily: corev1.IPFamily(req.IPFamily),
		},
	}
	created, err := h.cs.NetworkingV1alpha1().VirtualIPs(ns).Create(r.Context(), vip, metav1.CreateOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, VirtualIPResponse{
		Name: created.Name, Namespace: created.Namespace,
		Type: string(created.Spec.Type), IPFamily: string(created.Spec.IPFamily),
		CreatedAt: created.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

func (h *VirtualIPHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ns, name := chi.URLParam(r, "ns"), chi.URLParam(r, "name")
	if err := h.cs.NetworkingV1alpha1().VirtualIPs(ns).Delete(r.Context(), name, metav1.DeleteOptions{}); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
