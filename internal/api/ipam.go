package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	ipamv1alpha1 "github.com/ironcore-dev/ironcore/api/ipam/v1alpha1"
	versioned "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IPAMHandler struct{ cs versioned.Interface }

func NewIPAMHandler(cs versioned.Interface) *IPAMHandler { return &IPAMHandler{cs: cs} }

func (h *IPAMHandler) ListPrefixes(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	list, err := h.cs.IpamV1alpha1().Prefixes(ns).List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]PrefixResponse, 0, len(list.Items))
	for _, p := range list.Items {
		prefix := ""
		if p.Spec.Prefix != nil {
			prefix = p.Spec.Prefix.String()
		}
		resp = append(resp, PrefixResponse{
			Name: p.Name, Namespace: p.Namespace,
			Prefix: prefix, Phase: string(p.Status.Phase),
			CreatedAt: p.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *IPAMHandler) GetPrefix(w http.ResponseWriter, r *http.Request) {
	ns, name := chi.URLParam(r, "ns"), chi.URLParam(r, "name")
	p, err := h.cs.IpamV1alpha1().Prefixes(ns).Get(r.Context(), name, metav1.GetOptions{})
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	prefix := ""
	if p.Spec.Prefix != nil {
		prefix = p.Spec.Prefix.String()
	}
	parentRef := ""
	if p.Spec.ParentRef != nil {
		parentRef = p.Spec.ParentRef.Name
	}
	used := make([]string, 0, len(p.Status.Used))
	for _, u := range p.Status.Used {
		used = append(used, u.String())
	}
	writeJSON(w, http.StatusOK, PrefixDetailResponse{
		Name: p.Name, Namespace: p.Namespace,
		IPFamily: string(p.Spec.IPFamily),
		Prefix:   prefix, Phase: string(p.Status.Phase),
		ParentRef: parentRef, Used: used,
		CreatedAt: p.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

func (h *IPAMHandler) CreatePrefix(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	var req CreatePrefixRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	spec := ipamv1alpha1.PrefixSpec{
		IPFamily: corev1.IPFamily(req.IPFamily),
	}
	if req.Prefix != "" {
		parsed := commonv1alpha1.MustParseIPPrefix(req.Prefix)
		spec.Prefix = &parsed
	} else if req.PrefixLength > 0 {
		spec.PrefixLength = req.PrefixLength
	}
	if req.ParentRef != "" {
		spec.ParentRef = &corev1.LocalObjectReference{Name: req.ParentRef}
	}
	pfx := &ipamv1alpha1.Prefix{
		ObjectMeta: metav1.ObjectMeta{Name: req.Name, Namespace: ns},
		Spec:       spec,
	}
	created, err := h.cs.IpamV1alpha1().Prefixes(ns).Create(r.Context(), pfx, metav1.CreateOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	createdPrefix := ""
	if created.Spec.Prefix != nil {
		createdPrefix = created.Spec.Prefix.String()
	}
	writeJSON(w, http.StatusCreated, PrefixResponse{
		Name: created.Name, Namespace: created.Namespace,
		Prefix: createdPrefix, Phase: string(created.Status.Phase),
		CreatedAt: created.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

func (h *IPAMHandler) DeletePrefix(w http.ResponseWriter, r *http.Request) {
	ns, name := chi.URLParam(r, "ns"), chi.URLParam(r, "name")
	if err := h.cs.IpamV1alpha1().Prefixes(ns).Delete(r.Context(), name, metav1.DeleteOptions{}); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
