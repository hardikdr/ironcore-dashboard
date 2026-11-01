package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// EventHandler serves Kubernetes Events for a namespace, optionally filtered by
// involving resource name and kind.
type EventHandler struct{ k8s *kubernetes.Clientset }

func NewEventHandler(k8s *kubernetes.Clientset) *EventHandler {
	return &EventHandler{k8s: k8s}
}

// List handles GET /api/v1/namespaces/{ns}/events[?regarding=<name>&kind=<Kind>]
func (h *EventHandler) List(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	regarding := r.URL.Query().Get("regarding")
	kind := r.URL.Query().Get("kind")

	fieldSelector := ""
	if regarding != "" && kind != "" {
		fieldSelector = "involvedObject.name=" + regarding + ",involvedObject.kind=" + kind
	} else if regarding != "" {
		fieldSelector = "involvedObject.name=" + regarding
	}

	list, err := h.k8s.CoreV1().Events(ns).List(r.Context(), metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]EventResponse, 0, len(list.Items))
	for _, ev := range list.Items {
		firstTime := ""
		if !ev.FirstTimestamp.IsZero() {
			firstTime = ev.FirstTimestamp.UTC().Format("2006-01-02T15:04:05Z")
		}
		lastTime := ""
		if !ev.LastTimestamp.IsZero() {
			lastTime = ev.LastTimestamp.UTC().Format("2006-01-02T15:04:05Z")
		}
		resp = append(resp, EventResponse{
			Namespace:   ev.Namespace,
			Name:        ev.Name,
			Reason:      ev.Reason,
			Message:     ev.Message,
			Type:        ev.Type,
			Count:       ev.Count,
			FirstTime:   firstTime,
			LastTime:    lastTime,
			Regarding:   ev.InvolvedObject.Name,
			RegardsKind: ev.InvolvedObject.Kind,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}
