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
		ips := make([]string, 0, len(lb.Status.IPs))
		for _, ip := range lb.Status.IPs {
			ips = append(ips, ip.String())
		}
		resp = append(resp, LoadBalancerResponse{
			Name: lb.Name, Namespace: lb.Namespace,
			Type: string(lb.Spec.Type), IPs: ips,
			CreatedAt: lb.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *LoadBalancerHandler) Get(w http.ResponseWriter, r *http.Request) {
	ns, name := chi.URLParam(r, "ns"), chi.URLParam(r, "name")
	lb, err := h.cs.NetworkingV1alpha1().LoadBalancers(ns).Get(r.Context(), name, metav1.GetOptions{})
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	ips := make([]string, 0, len(lb.Status.IPs))
	for _, ip := range lb.Status.IPs {
		ips = append(ips, ip.String())
	}
	families := make([]string, 0, len(lb.Spec.IPFamilies))
	for _, f := range lb.Spec.IPFamilies {
		families = append(families, string(f))
	}
	ports := make([]LBPort, 0, len(lb.Spec.Ports))
	for _, p := range lb.Spec.Ports {
		proto := ""
		if p.Protocol != nil {
			proto = string(*p.Protocol)
		}
		ep := int32(0)
		if p.EndPort != nil {
			ep = *p.EndPort
		}
		ports = append(ports, LBPort{
			Protocol: proto,
			Port:     p.Port,
			EndPort:  ep,
		})
	}
	writeJSON(w, http.StatusOK, LoadBalancerDetailResponse{
		Name: lb.Name, Namespace: lb.Namespace,
		Type: string(lb.Spec.Type), IPFamilies: families,
		NetworkRef: lb.Spec.NetworkRef.Name,
		Ports: ports, IPs: ips,
		CreatedAt: lb.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

func (h *LoadBalancerHandler) Create(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	var req CreateLoadBalancerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	families := make([]corev1.IPFamily, 0, len(req.IPFamilies))
	for _, f := range req.IPFamilies {
		families = append(families, corev1.IPFamily(f))
	}
	ports := make([]networkingv1alpha1.LoadBalancerPort, 0, len(req.Ports))
	for _, p := range req.Ports {
		port := networkingv1alpha1.LoadBalancerPort{
			Port: p.Port,
		}
		if p.Protocol != "" {
			proto := corev1.Protocol(p.Protocol)
			port.Protocol = &proto
		}
		if p.EndPort != 0 {
			ep := p.EndPort
			port.EndPort = &ep
		}
		ports = append(ports, port)
	}
	lb := &networkingv1alpha1.LoadBalancer{
		ObjectMeta: metav1.ObjectMeta{Name: req.Name, Namespace: ns},
		Spec: networkingv1alpha1.LoadBalancerSpec{
			Type:       networkingv1alpha1.LoadBalancerType(req.Type),
			IPFamilies: families,
			NetworkRef: corev1.LocalObjectReference{Name: req.NetworkRef},
			Ports:      ports,
		},
	}
	created, err := h.cs.NetworkingV1alpha1().LoadBalancers(ns).Create(r.Context(), lb, metav1.CreateOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, LoadBalancerResponse{
		Name: created.Name, Namespace: created.Namespace,
		Type: string(created.Spec.Type),
		CreatedAt: created.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

func (h *LoadBalancerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ns, name := chi.URLParam(r, "ns"), chi.URLParam(r, "name")
	if err := h.cs.NetworkingV1alpha1().LoadBalancers(ns).Delete(r.Context(), name, metav1.DeleteOptions{}); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
