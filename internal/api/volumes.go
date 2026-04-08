package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	corev1alpha1    "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	versioned       "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	corev1          "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1          "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type VolumeHandler struct{ cs versioned.Interface }

func NewVolumeHandler(cs versioned.Interface) *VolumeHandler { return &VolumeHandler{cs: cs} }

func (h *VolumeHandler) List(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	list, err := h.cs.StorageV1alpha1().Volumes(ns).List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]VolumeResponse, 0, len(list.Items))
	for _, v := range list.Items {
		resp = append(resp, volumeToResponse(&v))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *VolumeHandler) Create(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	var req struct {
		Name        string `json:"name"`
		VolumeClass string `json:"volumeClass"`
		SizeBytes   int64  `json:"sizeBytes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	vol := &storagev1alpha1.Volume{
		ObjectMeta: metav1.ObjectMeta{Name: req.Name, Namespace: ns},
		Spec: storagev1alpha1.VolumeSpec{
			VolumeClassRef: &corev1.LocalObjectReference{Name: req.VolumeClass},
			Resources: corev1alpha1.ResourceList{
				corev1alpha1.ResourceStorage: *resource.NewQuantity(req.SizeBytes, resource.BinarySI),
			},
		},
	}
	created, err := h.cs.StorageV1alpha1().Volumes(ns).Create(r.Context(), vol, metav1.CreateOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, volumeToResponse(created))
}

func (h *VolumeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ns, name := chi.URLParam(r, "ns"), chi.URLParam(r, "name")
	if err := h.cs.StorageV1alpha1().Volumes(ns).Delete(r.Context(), name, metav1.DeleteOptions{}); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func volumeToResponse(v *storagev1alpha1.Volume) VolumeResponse {
	var sizeBytes int64
	if q, ok := v.Status.Resources[corev1alpha1.ResourceStorage]; ok {
		sizeBytes = q.Value()
	}
	vc := ""
	if v.Spec.VolumeClassRef != nil {
		vc = v.Spec.VolumeClassRef.Name
	}
	return VolumeResponse{
		Name: v.Name, Namespace: v.Namespace,
		State:       string(v.Status.State),
		SizeBytes:   sizeBytes,
		VolumeClass: vc,
		CreatedAt:   v.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
	}
}
