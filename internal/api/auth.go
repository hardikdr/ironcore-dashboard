package api

import (
	"net/http"

	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// NamespaceAccessHandler performs SelfSubjectAccessReview checks to verify
// that the caller can list machines in the requested namespace.
//
// Route: GET /api/v1/auth/namespace-access?namespace=<ns>
// Returns: {"allowed": true/false, "namespace": "<ns>"}
type NamespaceAccessHandler struct {
	k8s kubernetes.Interface
}

func NewNamespaceAccessHandler(k8s kubernetes.Interface) *NamespaceAccessHandler {
	return &NamespaceAccessHandler{k8s: k8s}
}

func (h *NamespaceAccessHandler) Check(w http.ResponseWriter, r *http.Request) {
	ns := r.URL.Query().Get("namespace")
	if ns == "" {
		writeError(w, http.StatusBadRequest, "namespace query parameter is required")
		return
	}

	review := &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Namespace: ns,
				Verb:      "list",
				Group:     "compute.ironcore.dev",
				Resource:  "machines",
			},
		},
	}

	result, err := h.k8s.AuthorizationV1().SelfSubjectAccessReviews().Create(
		r.Context(), review, metav1.CreateOptions{},
	)
	if err != nil {
		// Fail open: if we cannot perform the check (e.g., RBAC not configured),
		// allow the switch so operators are not blocked in non-OIDC setups.
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"allowed":   true,
			"namespace": ns,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"allowed":   result.Status.Allowed,
		"namespace": ns,
	})
}
