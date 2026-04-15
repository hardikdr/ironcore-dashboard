package api

import (
	"net/http"

	versioned "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type VolumeClassHandler struct{ cs versioned.Interface }

func NewVolumeClassHandler(cs versioned.Interface) *VolumeClassHandler {
	return &VolumeClassHandler{cs: cs}
}

func (h *VolumeClassHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.cs.StorageV1alpha1().VolumeClasses().List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]VolumeClassResponse, 0, len(list.Items))
	for _, vc := range list.Items {
		storage := vc.Capabilities.Storage().String()
		resp = append(resp, VolumeClassResponse{Name: vc.Name, Storage: storage})
	}
	writeJSON(w, http.StatusOK, resp)
}
