package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	versioned "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	metav1    "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			Name:      p.Name,
			Namespace: p.Namespace,
			Prefix:    prefix,
			Phase:     string(p.Status.Phase),
			CreatedAt: p.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}
