# Task 03 — Machines API Endpoints

## Prerequisite

Task 02 (IronCore client-go wrapper) must be complete. Verify:

```bash
make build  # must succeed
```

## Your job

Implement the full Machines CRUD API: list, get, create, delete, and power on/off. Also implement the machine classes endpoint needed by the Create wizard.

By the end:
- `GET  /api/v1/namespaces/{ns}/machines` → `[]MachineResponse`
- `GET  /api/v1/namespaces/{ns}/machines/{name}` → `MachineResponse`
- `POST /api/v1/namespaces/{ns}/machines` → `MachineResponse` (201)
- `DELETE /api/v1/namespaces/{ns}/machines/{name}` → 204
- `PATCH /api/v1/namespaces/{ns}/machines/{name}/power` → `MachineResponse`
- `GET  /api/v1/machineclasses` → `[]MachineClassResponse`

## Key IronCore types (from `github.com/ironcore-dev/ironcore/api/compute/v1alpha1`)

- `Machine.Spec.MachineClassRef.Name` → string (e.g. `"cx21"`)
- `Machine.Spec.Image` → string (e.g. `"ubuntu-22.04"`)
- `Machine.Spec.Power` → `computev1alpha1.Power` cast from `"On"` or `"Off"`
- `Machine.Status.State` → `"Pending"`, `"Running"`, `"Shutdown"`, `"Terminated"`, `"Terminating"`
- `Machine.Status.NetworkInterfaces[].IPs` → slice of IP addresses (call `.String()` on each)
- `Machine.Spec.Volumes[].Name` → string
- `MachineClass.Spec.Capabilities.Cpu()` → `*resource.Quantity` (call `.String()`)
- `MachineClass.Spec.Capabilities.Memory()` → `*resource.Quantity` (call `.String()`)

For volume attachments use ephemeral inline volumes:
```go
computev1alpha1.Volume{
    Name: v.Name,
    VolumeSource: computev1alpha1.VolumeSource{
        Ephemeral: &computev1alpha1.EphemeralVolumeSource{
            VolumeTemplate: &computev1alpha1.VolumeTemplateSpec{
                Spec: storagev1alpha1.VolumeSpec{
                    VolumeClassRef: &corev1.LocalObjectReference{Name: v.VolumeClass},
                    Resources: corev1alpha1.ResourceList{
                        corev1alpha1.ResourceStorage: *resource.NewQuantity(v.SizeBytes, resource.BinarySI),
                    },
                },
            },
        },
    },
}
```

For network interfaces use ephemeral inline NICs:
```go
computev1alpha1.NetworkInterface{
    Name: "primary",
    NetworkInterfaceSource: computev1alpha1.NetworkInterfaceSource{
        Ephemeral: &computev1alpha1.EphemeralNetworkInterfaceSource{
            NetworkInterfaceTemplate: &computev1alpha1.NetworkInterfaceTemplateSpec{
                Spec: networkingv1alpha1.NetworkInterfaceSpec{
                    NetworkRef: corev1.LocalObjectReference{Name: req.NetworkName},
                    IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol},
                },
            },
        },
    },
}
```

## Files to create / modify

| Action | File |
|--------|------|
| Create | `internal/api/types.go` |
| Create | `internal/api/machines.go` |
| Modify | `internal/server/server.go` (mount routes) |

## Step-by-step

### Step 1 — Write `internal/api/types.go`

```go
package api

import (
	"encoding/json"
	"net/http"
)

type MachineResponse struct {
	Name         string   `json:"name"`
	Namespace    string   `json:"namespace"`
	State        string   `json:"state"`
	Power        string   `json:"power"`
	MachineClass string   `json:"machineClass"`
	Image        string   `json:"image"`
	IPs          []string `json:"ips"`
	Volumes      []string `json:"volumes"`
	CreatedAt    string   `json:"createdAt"`
}

type VolumeResponse struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	State       string `json:"state"`
	SizeBytes   int64  `json:"sizeBytes"`
	VolumeClass string `json:"volumeClass"`
	CreatedAt   string `json:"createdAt"`
}

type NetworkResponse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	CreatedAt string `json:"createdAt"`
}

type NetworkInterfaceResponse struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	State     string   `json:"state"`
	IPs       []string `json:"ips"`
	Network   string   `json:"network"`
	Machine   string   `json:"machine"`
	CreatedAt string   `json:"createdAt"`
}

type VirtualIPResponse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	IP        string `json:"ip"`
	Type      string `json:"type"`
	IPFamily  string `json:"ipFamily"`
	CreatedAt string `json:"createdAt"`
}

type LoadBalancerResponse struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Type      string   `json:"type"`
	IPs       []string `json:"ips"`
	CreatedAt string   `json:"createdAt"`
}

type PrefixResponse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Prefix    string `json:"prefix"`
	Phase     string `json:"phase"`
	CreatedAt string `json:"createdAt"`
}

type CreateMachineRequest struct {
	Name         string             `json:"name"`
	MachineClass string             `json:"machineClass"`
	Image        string             `json:"image"`
	NetworkName  string             `json:"networkName"`
	Volumes      []VolumeAttachment `json:"volumes"`
	Power        string             `json:"power"`
}

type VolumeAttachment struct {
	Name        string `json:"name"`
	SizeBytes   int64  `json:"sizeBytes"`
	VolumeClass string `json:"volumeClass"`
}

type PatchPowerRequest struct {
	Power string `json:"power"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
```

### Step 2 — Write `internal/api/machines.go`

```go
package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	computev1alpha1  "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	corev1alpha1     "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	storagev1alpha1  "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	versioned        "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	corev1           "k8s.io/api/core/v1"
	metav1           "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type MachineHandler struct{ cs versioned.Interface }

func NewMachineHandler(cs versioned.Interface) *MachineHandler {
	return &MachineHandler{cs: cs}
}

func (h *MachineHandler) List(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	list, err := h.cs.ComputeV1alpha1().Machines(ns).List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]MachineResponse, 0, len(list.Items))
	for _, m := range list.Items {
		resp = append(resp, machineToResponse(&m))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *MachineHandler) Get(w http.ResponseWriter, r *http.Request) {
	ns, name := chi.URLParam(r, "ns"), chi.URLParam(r, "name")
	m, err := h.cs.ComputeV1alpha1().Machines(ns).Get(r.Context(), name, metav1.GetOptions{})
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, machineToResponse(m))
}

func (h *MachineHandler) Create(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	var req CreateMachineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	power := computev1alpha1.Power("On")
	if req.Power == "Off" {
		power = computev1alpha1.Power("Off")
	}

	volumes := make([]computev1alpha1.Volume, 0, len(req.Volumes))
	for _, v := range req.Volumes {
		volumes = append(volumes, computev1alpha1.Volume{
			Name: v.Name,
			VolumeSource: computev1alpha1.VolumeSource{
				Ephemeral: &computev1alpha1.EphemeralVolumeSource{
					VolumeTemplate: &computev1alpha1.VolumeTemplateSpec{
						Spec: storagev1alpha1.VolumeSpec{
							VolumeClassRef: &corev1.LocalObjectReference{Name: v.VolumeClass},
							Resources: corev1alpha1.ResourceList{
								corev1alpha1.ResourceStorage: *resource.NewQuantity(v.SizeBytes, resource.BinarySI),
							},
						},
					},
				},
			},
		})
	}

	machine := &computev1alpha1.Machine{
		ObjectMeta: metav1.ObjectMeta{Name: req.Name, Namespace: ns},
		Spec: computev1alpha1.MachineSpec{
			MachineClassRef: corev1.LocalObjectReference{Name: req.MachineClass},
			Image:           req.Image,
			Power:           power,
			Volumes:         volumes,
			NetworkInterfaces: []computev1alpha1.NetworkInterface{
				{
					Name: "primary",
					NetworkInterfaceSource: computev1alpha1.NetworkInterfaceSource{
						Ephemeral: &computev1alpha1.EphemeralNetworkInterfaceSource{
							NetworkInterfaceTemplate: &computev1alpha1.NetworkInterfaceTemplateSpec{
								Spec: networkingv1alpha1.NetworkInterfaceSpec{
									NetworkRef: corev1.LocalObjectReference{Name: req.NetworkName},
									IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol},
								},
							},
						},
					},
				},
			},
		},
	}

	created, err := h.cs.ComputeV1alpha1().Machines(ns).Create(r.Context(), machine, metav1.CreateOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, machineToResponse(created))
}

func (h *MachineHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ns, name := chi.URLParam(r, "ns"), chi.URLParam(r, "name")
	if err := h.cs.ComputeV1alpha1().Machines(ns).Delete(r.Context(), name, metav1.DeleteOptions{}); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MachineHandler) PatchPower(w http.ResponseWriter, r *http.Request) {
	ns, name := chi.URLParam(r, "ns"), chi.URLParam(r, "name")
	var req PatchPowerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	m, err := h.cs.ComputeV1alpha1().Machines(ns).Get(r.Context(), name, metav1.GetOptions{})
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	m.Spec.Power = computev1alpha1.Power(req.Power)
	updated, err := h.cs.ComputeV1alpha1().Machines(ns).Update(r.Context(), m, metav1.UpdateOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, machineToResponse(updated))
}

func (h *MachineHandler) ListMachineClasses(w http.ResponseWriter, r *http.Request) {
	list, err := h.cs.ComputeV1alpha1().MachineClasses().List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	type MCResponse struct {
		Name string `json:"name"`
		CPU  string `json:"cpu"`
		RAM  string `json:"ram"`
	}
	resp := make([]MCResponse, 0, len(list.Items))
	for _, mc := range list.Items {
		cpu := mc.Spec.Capabilities.Cpu().String()
		ram := mc.Spec.Capabilities.Memory().String()
		resp = append(resp, MCResponse{Name: mc.Name, CPU: cpu, RAM: ram})
	}
	writeJSON(w, http.StatusOK, resp)
}

func machineToResponse(m *computev1alpha1.Machine) MachineResponse {
	ips := []string{}
	for _, ni := range m.Status.NetworkInterfaces {
		for _, ip := range ni.IPs {
			ips = append(ips, ip.String())
		}
	}
	vols := []string{}
	for _, v := range m.Spec.Volumes {
		vols = append(vols, v.Name)
	}
	return MachineResponse{
		Name:         m.Name,
		Namespace:    m.Namespace,
		State:        string(m.Status.State),
		Power:        string(m.Spec.Power),
		MachineClass: m.Spec.MachineClassRef.Name,
		Image:        m.Spec.Image,
		IPs:          ips,
		Volumes:      vols,
		CreatedAt:    m.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
	}
}
```

### Step 3 — Mount routes in `internal/server/server.go`

Add these lines inside `New()`, after the middleware setup:

```go
import "github.com/ironcore-dev/ironcore-dashboard/internal/api"

// Inside New(), after middleware:
mh := api.NewMachineHandler(cs)
r.Get("/api/v1/machineclasses", mh.ListMachineClasses)
r.Route("/api/v1/namespaces/{ns}/machines", func(r chi.Router) {
    r.Get("/", mh.List)
    r.Post("/", mh.Create)
    r.Get("/{name}", mh.Get)
    r.Delete("/{name}", mh.Delete)
    r.Patch("/{name}/power", mh.PatchPower)
})
```

### Step 4 — Test

```bash
make build
go run ./cmd/server --kubeconfig ~/.kube/config &
sleep 2
curl http://localhost:8080/api/v1/namespaces/default/machines
curl http://localhost:8080/api/v1/machineclasses
```

Expected: both return `[]` (empty JSON array — no machines/classes yet in cluster)

### Step 5 — Commit

```bash
git add internal/api/ internal/server/server.go
git commit -m "feat: machines CRUD + power patch + machine classes API"
```

## Done criteria

- All 6 endpoints exist and return valid JSON
- `make build` succeeds
- Empty cluster returns `[]` for list endpoints (not 500 errors)