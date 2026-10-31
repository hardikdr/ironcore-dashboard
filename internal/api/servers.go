package api

import (
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var serverGVR = schema.GroupVersionResource{
	Group:    "metal.ironcore.dev",
	Version:  "v1alpha1",
	Resource: "servers",
}

type ServerHandler struct{ dyn dynamic.Interface }

func NewServerHandler(dyn dynamic.Interface) *ServerHandler {
	return &ServerHandler{dyn: dyn}
}

func (h *ServerHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.dyn.Resource(serverGVR).List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeJSON(w, http.StatusOK, []ServerResponse{})
		return
	}

	resp := make([]ServerResponse, 0, len(list.Items))
	for _, item := range list.Items {
		spec, _ := item.Object["spec"].(map[string]interface{})
		status, _ := item.Object["status"].(map[string]interface{})
		metadata, _ := item.Object["metadata"].(map[string]interface{})

		bmcAddr := ""
		if spec != nil {
			if bmc, ok := spec["bmc"].(map[string]interface{}); ok {
				bmcAddr, _ = bmc["address"].(string)
			}
		}

		powerState := "Unknown"
		claimStatus := "Unclaimed"
		if status != nil {
			if ps, ok := status["powerState"].(string); ok {
				powerState = ps
			}
			if cr, ok := status["claimRef"].(map[string]interface{}); ok && cr != nil {
				claimStatus = "Claimed"
			}
		}

		location := ""
		if metadata != nil {
			if lblMap, ok := metadata["labels"].(map[string]interface{}); ok {
				if loc, ok := lblMap["metal.ironcore.dev/location"].(string); ok {
					location = loc
				}
			}
		}

		resp = append(resp, ServerResponse{
			Name:        item.GetName(),
			BMCAddress:  bmcAddr,
			PowerState:  powerState,
			Location:    location,
			ClaimStatus: claimStatus,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}
