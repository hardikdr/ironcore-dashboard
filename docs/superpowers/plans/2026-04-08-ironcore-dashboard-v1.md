# IronCore Dashboard v1 — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a demo-ready v1 IronCore dashboard — a Go HTTP backend proxying the IronCore Kubernetes API, and a Vue 3 frontend — that lets engineers and operators list, inspect, create, and delete IronCore resources (Machines, Volumes, Networks, VIPs, Load Balancers, IP Prefixes) scoped to a namespace/project, with a Gardener-style namespace switcher.

**Architecture:** The Go backend wraps the IronCore `client-go` typed clients, exposes a thin REST API under `/api/v1/namespaces/:ns/...`, and serves the built Vue 3 SPA as static files from the same binary. The frontend uses Vue 3 + Vuetify 3 + Pinia and communicates only with the Go backend — never directly with the Kubernetes API. Auth in v1 is a single service account (kubeconfig on disk); OIDC is a v2 concern.

**Tech Stack:** Go 1.22+, `github.com/go-chi/chi/v5` (router), `github.com/ironcore-dev/ironcore` client-go (already vendored in the monorepo), Vue 3, Vite, Vuetify 3, Pinia, TypeScript.

---

## Workspace layout

All work happens inside:
```
dashboard-workspace/
├── ironcore/               ← IronCore APIs + client-go (READ ONLY — do not modify)
├── ironcore-in-a-box/      ← local Kind cluster (READ ONLY)
└── ironcore-dashboard/     ← THIS PROJECT (all tasks write here)
    ├── cmd/server/         ← Go binary entry point
    ├── internal/
    │   ├── client/         ← IronCore client-go wrapper
    │   ├── api/            ← HTTP handlers per resource
    │   └── server/         ← HTTP router + middleware
    ├── frontend/           ← Vue 3 SPA
    └── .agents/            ← agent prompt files (one folder per task)
```

---

## File map (created across all tasks)

```
ironcore-dashboard/
├── go.mod
├── go.sum
├── Makefile
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── client/
│   │   └── ironcore.go          # IronCore clientset factory
│   ├── api/
│   │   ├── types.go             # shared JSON response structs
│   │   ├── machines.go          # GET/POST/DELETE machines + PATCH power
│   │   ├── volumes.go           # GET/POST/DELETE volumes
│   │   ├── networks.go          # GET networks, networkinterfaces
│   │   ├── virtualips.go        # GET/POST virtualips
│   │   ├── loadbalancers.go     # GET loadbalancers
│   │   └── ipam.go              # GET prefixes
│   └── server/
│       └── server.go            # chi router, CORS, static file serving
├── frontend/
│   ├── package.json
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── index.html
│   └── src/
│       ├── main.ts
│       ├── App.vue
│       ├── plugins/
│       │   └── vuetify.ts       # theme: blue #1a5fa8
│       ├── router/
│       │   └── index.ts         # vue-router, namespace-scoped routes
│       ├── stores/
│       │   ├── namespace.ts     # active namespace + list
│       │   ├── machines.ts
│       │   ├── volumes.ts
│       │   └── networks.ts
│       ├── api/
│       │   └── client.ts        # fetch wrapper for /api/v1/...
│       ├── layouts/
│       │   └── DashboardLayout.vue  # topbar + sidebar + <router-view>
│       ├── components/
│       │   ├── TopBar.vue
│       │   ├── Sidebar.vue
│       │   ├── NamespaceSwitcher.vue
│       │   ├── StatCard.vue
│       │   ├── StatusBadge.vue
│       │   └── ResourceTable.vue
│       └── views/
│           ├── MachinesView.vue
│           ├── MachineDetailView.vue
│           ├── MachineCreateView.vue
│           ├── VolumesView.vue
│           ├── NetworksView.vue
│           ├── VirtualIPsView.vue
│           ├── LoadBalancersView.vue
│           └── IPPrefixesView.vue
└── Makefile
```

---

## Task 01 — Go backend scaffold

**Agent prompt file:** `.agents/task-01-backend-scaffold/PROMPT.md`

**Files:**
- Create: `go.mod`
- Create: `Makefile`
- Create: `cmd/server/main.go`
- Create: `internal/server/server.go`

- [ ] **Step 1: Initialise the Go module**

```bash
cd ironcore-dashboard
go mod init github.com/ironcore-dev/ironcore-dashboard
```

Expected: `go.mod` created with `module github.com/ironcore-dev/ironcore-dashboard`

- [ ] **Step 2: Add dependencies**

```bash
go get github.com/go-chi/chi/v5@latest
go get github.com/go-chi/cors@latest
go get k8s.io/client-go@latest
go get k8s.io/apimachinery@latest
# Replace directive so we use the local ironcore module
go mod edit -replace github.com/ironcore-dev/ironcore=../ironcore
go get github.com/ironcore-dev/ironcore@latest
go mod tidy
```

- [ ] **Step 3: Write `internal/server/server.go`**

```go
package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Server struct {
	router *chi.Mux
}

func New() *Server {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173", "http://localhost:8080"},
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	}))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	return &Server{router: r}
}

func (s *Server) Router() *chi.Mux { return s.router }

func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.router)
}
```

- [ ] **Step 4: Write `cmd/server/main.go`**

```go
package main

import (
	"flag"
	"log"

	"github.com/ironcore-dev/ironcore-dashboard/internal/server"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	flag.Parse()

	srv := server.New()
	log.Printf("IronCore Dashboard listening on %s", *addr)
	if err := srv.ListenAndServe(*addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
```

- [ ] **Step 5: Write `Makefile`**

```makefile
.PHONY: build run dev tidy

build:
	go build -o bin/ironcore-dashboard ./cmd/server

run:
	go run ./cmd/server --addr :8080

tidy:
	go mod tidy

test:
	go test ./...
```

- [ ] **Step 6: Run and verify**

```bash
make run
# In another terminal:
curl http://localhost:8080/healthz
```

Expected output: `ok`

- [ ] **Step 7: Commit**

```bash
git add go.mod go.sum Makefile cmd/ internal/server/
git commit -m "feat: scaffold Go backend with chi router and /healthz"
```

---

## Task 02 — IronCore client-go wrapper

**Agent prompt file:** `.agents/task-02-ironcore-client/PROMPT.md`

**Files:**
- Create: `internal/client/ironcore.go`

**Context:** The IronCore clientset is at `github.com/ironcore-dev/ironcore/client-go/ironcore/versioned`. It has five sub-clients: `ComputeV1alpha1()`, `StorageV1alpha1()`, `NetworkingV1alpha1()`, `IpamV1alpha1()`, `CoreV1alpha1()`. The kubeconfig path comes from a flag.

- [ ] **Step 1: Write `internal/client/ironcore.go`**

```go
package client

import (
	"fmt"

	versioned "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// New returns an IronCore versioned clientset.
// If kubeconfigPath is empty it falls back to in-cluster config.
func New(kubeconfigPath string) (versioned.Interface, error) {
	var cfg *rest.Config
	var err error

	if kubeconfigPath != "" {
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else {
		cfg, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, fmt.Errorf("build kubeconfig: %w", err)
	}

	cs, err := versioned.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("build ironcore clientset: %w", err)
	}
	return cs, nil
}
```

- [ ] **Step 2: Wire the client into `cmd/server/main.go`**

Replace `cmd/server/main.go` with:

```go
package main

import (
	"flag"
	"log"

	ironclient "github.com/ironcore-dev/ironcore-dashboard/internal/client"
	"github.com/ironcore-dev/ironcore-dashboard/internal/server"
)

func main() {
	addr       := flag.String("addr", ":8080", "HTTP listen address")
	kubeconfig := flag.String("kubeconfig", "", "path to kubeconfig (empty = in-cluster)")
	flag.Parse()

	cs, err := ironclient.New(*kubeconfig)
	if err != nil {
		log.Fatalf("ironcore client: %v", err)
	}

	srv := server.New(cs)
	log.Printf("IronCore Dashboard listening on %s", *addr)
	if err := srv.ListenAndServe(*addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
```

- [ ] **Step 3: Update `internal/server/server.go` to accept the clientset**

Change the `New` signature:

```go
import versioned "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"

type Server struct {
	router    *chi.Mux
	ironcore  versioned.Interface
}

func New(cs versioned.Interface) *Server {
	s := &Server{ironcore: cs}
	r := chi.NewRouter()
	// ... same middleware setup ...
	s.router = r
	return s
}
```

- [ ] **Step 4: Verify it compiles against ironcore-in-a-box**

```bash
make build
# Should produce bin/ironcore-dashboard with no errors
make run --kubeconfig ~/.kube/config
curl http://localhost:8080/healthz
```

Expected: `ok`

- [ ] **Step 5: Commit**

```bash
git add internal/client/ internal/server/ cmd/
git commit -m "feat: wire IronCore client-go into server"
```

---

## Task 03 — Machines API endpoints

**Agent prompt file:** `.agents/task-03-machines-api/PROMPT.md`

**Files:**
- Create: `internal/api/types.go`
- Create: `internal/api/machines.go`
- Modify: `internal/server/server.go` (mount machine routes)

**Context:** Key types from `github.com/ironcore-dev/ironcore/api/compute/v1alpha1`:
- `Machine.Spec.MachineClassRef.Name` → size name (e.g. `cx21`)
- `Machine.Spec.Image` → OS image string
- `Machine.Spec.Power` → `"On"` or `"Off"`
- `Machine.Status.State` → `"Pending"`, `"Running"`, `"Shutdown"`, `"Terminated"`, `"Terminating"`
- `Machine.Status.NetworkInterfaces[].IPs` → slice of IP strings
- `Machine.Spec.Volumes[].Name` → volume attachment name

- [ ] **Step 1: Write `internal/api/types.go`**

```go
package api

// MachineResponse is the JSON shape returned to the frontend.
type MachineResponse struct {
	Name       string   `json:"name"`
	Namespace  string   `json:"namespace"`
	State      string   `json:"state"`
	Power      string   `json:"power"`
	MachineClass string `json:"machineClass"`
	Image      string   `json:"image"`
	IPs        []string `json:"ips"`
	Volumes    []string `json:"volumes"`
	CreatedAt  string   `json:"createdAt"`
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

// CreateMachineRequest is the JSON body for POST /namespaces/:ns/machines
type CreateMachineRequest struct {
	Name         string              `json:"name"`
	MachineClass string              `json:"machineClass"`
	Image        string              `json:"image"`
	NetworkName  string              `json:"networkName"`
	Volumes      []VolumeAttachment  `json:"volumes"`
	Power        string              `json:"power"` // "On" or "Off"
}

type VolumeAttachment struct {
	Name      string `json:"name"`
	SizeBytes int64  `json:"sizeBytes"`
	VolumeClass string `json:"volumeClass"`
}

// PatchPowerRequest is the JSON body for PATCH .../power
type PatchPowerRequest struct {
	Power string `json:"power"` // "On" or "Off"
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

> Note: add `import ("encoding/json" "net/http")` at top of types.go.

- [ ] **Step 2: Write `internal/api/machines.go`**

```go
package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	corev1alpha1    "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	versioned       "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	corev1          "k8s.io/api/core/v1"
	metav1          "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MachineHandler struct{ cs versioned.Interface }

func NewMachineHandler(cs versioned.Interface) *MachineHandler {
	return &MachineHandler{cs: cs}
}

// GET /api/v1/namespaces/{ns}/machines
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

// GET /api/v1/namespaces/{ns}/machines/{name}
func (h *MachineHandler) Get(w http.ResponseWriter, r *http.Request) {
	ns, name := chi.URLParam(r, "ns"), chi.URLParam(r, "name")
	m, err := h.cs.ComputeV1alpha1().Machines(ns).Get(r.Context(), name, metav1.GetOptions{})
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, machineToResponse(m))
}

// POST /api/v1/namespaces/{ns}/machines
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
							Resources:      corev1alpha1.ResourceList{
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

// DELETE /api/v1/namespaces/{ns}/machines/{name}
func (h *MachineHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ns, name := chi.URLParam(r, "ns"), chi.URLParam(r, "name")
	if err := h.cs.ComputeV1alpha1().Machines(ns).Delete(r.Context(), name, metav1.DeleteOptions{}); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PATCH /api/v1/namespaces/{ns}/machines/{name}/power
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

- [ ] **Step 3: Mount machine routes in `internal/server/server.go`**

Add inside `New()` after middleware setup:

```go
mh := api.NewMachineHandler(cs)
r.Route("/api/v1/namespaces/{ns}/machines", func(r chi.Router) {
    r.Get("/", mh.List)
    r.Post("/", mh.Create)
    r.Get("/{name}", mh.Get)
    r.Delete("/{name}", mh.Delete)
    r.Patch("/{name}/power", mh.PatchPower)
})
```

- [ ] **Step 4: Test against ironcore-in-a-box**

```bash
# Start the cluster if not running:
cd ../ironcore-in-a-box && make deploy && cd ../ironcore-dashboard

make run -- --kubeconfig ~/.kube/config
curl http://localhost:8080/api/v1/namespaces/default/machines
```

Expected: `[]` (empty JSON array — no machines yet)

- [ ] **Step 5: Commit**

```bash
git add internal/api/ internal/server/
git commit -m "feat: machines CRUD + power patch API endpoints"
```

---

## Task 04 — Remaining API endpoints (Volumes, Networking, IPAM)

**Agent prompt file:** `.agents/task-04-remaining-apis/PROMPT.md`

**Files:**
- Create: `internal/api/volumes.go`
- Create: `internal/api/networks.go`
- Create: `internal/api/virtualips.go`
- Create: `internal/api/loadbalancers.go`
- Create: `internal/api/ipam.go`
- Modify: `internal/server/server.go` (mount all new routes)

**Context:**
- Volumes: `cs.StorageV1alpha1().Volumes(ns)` — `Volume.Status.State` is `"Pending"`, `"Available"`, `"Error"`. Size in `Volume.Status.Resources[corev1alpha1.ResourceStorage]`.
- Networks: `cs.NetworkingV1alpha1().Networks(ns)` — no Status.State, just list.
- NetworkInterfaces: `cs.NetworkingV1alpha1().NetworkInterfaces(ns)` — `Status.IPs`, `Status.State`, `Spec.NetworkRef.Name`, `Spec.MachineRef.Name`.
- VirtualIPs: `cs.NetworkingV1alpha1().VirtualIPs(ns)` — `Status.IP`, `Spec.Type`, `Spec.IPFamily`.
- LoadBalancers: `cs.NetworkingV1alpha1().LoadBalancers(ns)` — `Status.IPs`, `Spec.Type`.
- Prefixes: `cs.IpamV1alpha1().Prefixes(ns)` — `Spec.Prefix`, `Status.Phase` (`"Pending"`, `"Allocated"`).

- [ ] **Step 1: Write `internal/api/volumes.go`**

```go
package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	corev1alpha1    "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	versioned       "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	corev1          "k8s.io/api/core/v1"
	metav1          "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type VolumeHandler struct{ cs versioned.Interface }

func NewVolumeHandler(cs versioned.Interface) *VolumeHandler { return &VolumeHandler{cs: cs} }

func (h *VolumeHandler) List(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	list, err := h.cs.StorageV1alpha1().Volumes(ns).List(r.Context(), metav1.ListOptions{})
	if err != nil { writeError(w, http.StatusInternalServerError, err.Error()); return }
	resp := make([]VolumeResponse, 0, len(list.Items))
	for _, v := range list.Items { resp = append(resp, volumeToResponse(&v)) }
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
		writeError(w, http.StatusBadRequest, err.Error()); return
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
	if err != nil { writeError(w, http.StatusInternalServerError, err.Error()); return }
	writeJSON(w, http.StatusCreated, volumeToResponse(created))
}

func (h *VolumeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ns, name := chi.URLParam(r, "ns"), chi.URLParam(r, "name")
	if err := h.cs.StorageV1alpha1().Volumes(ns).Delete(r.Context(), name, metav1.DeleteOptions{}); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error()); return
	}
	w.WriteHeader(http.StatusNoContent)
}

func volumeToResponse(v *storagev1alpha1.Volume) VolumeResponse {
	var sizeBytes int64
	if q, ok := v.Status.Resources[corev1alpha1.ResourceStorage]; ok {
		sizeBytes = q.Value()
	}
	vc := ""
	if v.Spec.VolumeClassRef != nil { vc = v.Spec.VolumeClassRef.Name }
	return VolumeResponse{
		Name: v.Name, Namespace: v.Namespace,
		State: string(v.Status.State),
		SizeBytes: sizeBytes, VolumeClass: vc,
		CreatedAt: v.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
	}
}
```

- [ ] **Step 2: Write `internal/api/networks.go`**

```go
package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	versioned "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	metav1    "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NetworkHandler struct{ cs versioned.Interface }

func NewNetworkHandler(cs versioned.Interface) *NetworkHandler { return &NetworkHandler{cs: cs} }

func (h *NetworkHandler) ListNetworks(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	list, err := h.cs.NetworkingV1alpha1().Networks(ns).List(r.Context(), metav1.ListOptions{})
	if err != nil { writeError(w, http.StatusInternalServerError, err.Error()); return }
	resp := make([]NetworkResponse, 0, len(list.Items))
	for _, n := range list.Items {
		resp = append(resp, NetworkResponse{
			Name: n.Name, Namespace: n.Namespace,
			CreatedAt: n.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *NetworkHandler) ListNetworkInterfaces(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	list, err := h.cs.NetworkingV1alpha1().NetworkInterfaces(ns).List(r.Context(), metav1.ListOptions{})
	if err != nil { writeError(w, http.StatusInternalServerError, err.Error()); return }
	resp := make([]NetworkInterfaceResponse, 0, len(list.Items))
	for _, ni := range list.Items {
		ips := []string{}
		for _, ip := range ni.Status.IPs { ips = append(ips, ip.String()) }
		machine := ""
		if ni.Spec.MachineRef != nil { machine = ni.Spec.MachineRef.Name }
		resp = append(resp, NetworkInterfaceResponse{
			Name: ni.Name, Namespace: ni.Namespace,
			State: string(ni.Status.State), IPs: ips,
			Network: ni.Spec.NetworkRef.Name, Machine: machine,
			CreatedAt: ni.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}
```

- [ ] **Step 3: Write `internal/api/virtualips.go`**

```go
package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	versioned "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	metav1    "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type VirtualIPHandler struct{ cs versioned.Interface }

func NewVirtualIPHandler(cs versioned.Interface) *VirtualIPHandler { return &VirtualIPHandler{cs: cs} }

func (h *VirtualIPHandler) List(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	list, err := h.cs.NetworkingV1alpha1().VirtualIPs(ns).List(r.Context(), metav1.ListOptions{})
	if err != nil { writeError(w, http.StatusInternalServerError, err.Error()); return }
	resp := make([]VirtualIPResponse, 0, len(list.Items))
	for _, v := range list.Items {
		ip := ""
		if v.Status.IP != nil { ip = v.Status.IP.String() }
		resp = append(resp, VirtualIPResponse{
			Name: v.Name, Namespace: v.Namespace,
			IP: ip, Type: string(v.Spec.Type),
			IPFamily: string(v.Spec.IPFamily),
			CreatedAt: v.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}
```

- [ ] **Step 4: Write `internal/api/loadbalancers.go`**

```go
package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	versioned "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	metav1    "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LoadBalancerHandler struct{ cs versioned.Interface }

func NewLoadBalancerHandler(cs versioned.Interface) *LoadBalancerHandler {
	return &LoadBalancerHandler{cs: cs}
}

func (h *LoadBalancerHandler) List(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	list, err := h.cs.NetworkingV1alpha1().LoadBalancers(ns).List(r.Context(), metav1.ListOptions{})
	if err != nil { writeError(w, http.StatusInternalServerError, err.Error()); return }
	resp := make([]LoadBalancerResponse, 0, len(list.Items))
	for _, lb := range list.Items {
		ips := []string{}
		for _, ip := range lb.Status.IPs { ips = append(ips, ip.String()) }
		resp = append(resp, LoadBalancerResponse{
			Name: lb.Name, Namespace: lb.Namespace,
			Type: string(lb.Spec.Type), IPs: ips,
			CreatedAt: lb.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}
```

- [ ] **Step 5: Write `internal/api/ipam.go`**

```go
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
	if err != nil { writeError(w, http.StatusInternalServerError, err.Error()); return }
	resp := make([]PrefixResponse, 0, len(list.Items))
	for _, p := range list.Items {
		prefix := ""
		if p.Spec.Prefix != nil { prefix = p.Spec.Prefix.String() }
		resp = append(resp, PrefixResponse{
			Name: p.Name, Namespace: p.Namespace,
			Prefix: prefix, Phase: string(p.Status.Phase),
			CreatedAt: p.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}
```

- [ ] **Step 6: Add namespace listing endpoint and mount all routes in `server.go`**

In `internal/server/server.go`, add inside `New()`:

```go
// Namespace list (for project/namespace switcher)
r.Get("/api/v1/namespaces", func(w http.ResponseWriter, r *http.Request) {
    nsList, err := cs.CoreV1alpha1() /* actually k8s core */ ...
    // Use standard k8s client for namespace listing:
    // (handled by adding a k8s.io/client-go *kubernetes.Clientset alongside ironcore cs)
})

vh  := api.NewVolumeHandler(cs)
nh  := api.NewNetworkHandler(cs)
vip := api.NewVirtualIPHandler(cs)
lb  := api.NewLoadBalancerHandler(cs)
ip  := api.NewIPAMHandler(cs)

r.Route("/api/v1/namespaces/{ns}", func(r chi.Router) {
    r.Get("/volumes",           vh.List)
    r.Post("/volumes",          vh.Create)
    r.Delete("/volumes/{name}", vh.Delete)

    r.Get("/networks",              nh.ListNetworks)
    r.Get("/networkinterfaces",     nh.ListNetworkInterfaces)

    r.Get("/virtualips",            vip.List)
    r.Get("/loadbalancers",         lb.List)
    r.Get("/prefixes",              ip.ListPrefixes)
})
```

> Note: For namespace listing, add a standard `k8s.io/client-go` `*kubernetes.Clientset` to the Server alongside the IronCore clientset. `GET /api/v1/namespaces` calls `k8sClient.CoreV1().Namespaces().List(...)`.

- [ ] **Step 7: Verify all endpoints compile and return empty arrays**

```bash
make build
make run -- --kubeconfig ~/.kube/config
curl http://localhost:8080/api/v1/namespaces/default/volumes
curl http://localhost:8080/api/v1/namespaces/default/networks
curl http://localhost:8080/api/v1/namespaces/default/virtualips
curl http://localhost:8080/api/v1/namespaces/default/prefixes
```

Expected: `[]` for each

- [ ] **Step 8: Commit**

```bash
git add internal/api/ internal/server/
git commit -m "feat: volumes, networks, VIPs, LBs, IPAM list endpoints"
```

---

## Task 05 — Vue 3 frontend scaffold

**Agent prompt file:** `.agents/task-05-frontend-scaffold/PROMPT.md`

**Files:**
- Create: `frontend/package.json`
- Create: `frontend/vite.config.ts`
- Create: `frontend/tsconfig.json`
- Create: `frontend/index.html`
- Create: `frontend/src/main.ts`
- Create: `frontend/src/App.vue`
- Create: `frontend/src/plugins/vuetify.ts`
- Create: `frontend/src/router/index.ts`
- Create: `frontend/src/api/client.ts`
- Create: `frontend/src/stores/namespace.ts`
- Create: `frontend/src/layouts/DashboardLayout.vue`

- [ ] **Step 1: Create `frontend/package.json`**

```json
{
  "name": "ironcore-dashboard-frontend",
  "version": "0.1.0",
  "private": true,
  "scripts": {
    "dev": "vite",
    "build": "vue-tsc && vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "vue": "^3.4.0",
    "vue-router": "^4.3.0",
    "pinia": "^2.1.0",
    "vuetify": "^3.5.0",
    "@mdi/font": "^7.4.0",
    "axios": "^1.6.0"
  },
  "devDependencies": {
    "@vitejs/plugin-vue": "^5.0.0",
    "typescript": "^5.3.0",
    "vite": "^5.1.0",
    "vue-tsc": "^2.0.0"
  }
}
```

- [ ] **Step 2: Create `frontend/vite.config.ts`**

```typescript
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:8080'
    }
  },
  build: {
    outDir: '../dist/frontend'
  }
})
```

- [ ] **Step 3: Create `frontend/src/plugins/vuetify.ts`**

```typescript
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import 'vuetify/styles'
import '@mdi/font/css/materialdesignicons.css'

export default createVuetify({
  components,
  directives,
  theme: {
    defaultTheme: 'light',
    themes: {
      light: {
        colors: {
          primary:    '#1a5fa8',
          secondary:  '#475569',
          success:    '#16a34a',
          warning:    '#d97706',
          error:      '#b91c1c',
          info:       '#0369a1',
          background: '#f4f6fa',
          surface:    '#ffffff',
        }
      }
    }
  }
})
```

- [ ] **Step 4: Create `frontend/src/api/client.ts`**

```typescript
const BASE = '/api/v1'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error || res.statusText)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

export const api = {
  namespaces: {
    list: () => request<string[]>('/namespaces')
  },
  machines: {
    list:   (ns: string) => request<Machine[]>(`/namespaces/${ns}/machines`),
    get:    (ns: string, name: string) => request<Machine>(`/namespaces/${ns}/machines/${name}`),
    create: (ns: string, body: CreateMachineRequest) =>
      request<Machine>(`/namespaces/${ns}/machines`, { method: 'POST', body: JSON.stringify(body) }),
    delete: (ns: string, name: string) =>
      request<void>(`/namespaces/${ns}/machines/${name}`, { method: 'DELETE' }),
    power:  (ns: string, name: string, power: 'On'|'Off') =>
      request<Machine>(`/namespaces/${ns}/machines/${name}/power`, {
        method: 'PATCH', body: JSON.stringify({ power })
      })
  },
  volumes: {
    list:   (ns: string) => request<Volume[]>(`/namespaces/${ns}/volumes`),
    create: (ns: string, body: CreateVolumeRequest) =>
      request<Volume>(`/namespaces/${ns}/volumes`, { method: 'POST', body: JSON.stringify(body) }),
    delete: (ns: string, name: string) =>
      request<void>(`/namespaces/${ns}/volumes/${name}`, { method: 'DELETE' })
  },
  networks: {
    list:               (ns: string) => request<Network[]>(`/namespaces/${ns}/networks`),
    listInterfaces:     (ns: string) => request<NetworkInterface[]>(`/namespaces/${ns}/networkinterfaces`)
  },
  virtualIPs: {
    list: (ns: string) => request<VirtualIP[]>(`/namespaces/${ns}/virtualips`)
  },
  loadBalancers: {
    list: (ns: string) => request<LoadBalancer[]>(`/namespaces/${ns}/loadbalancers`)
  },
  prefixes: {
    list: (ns: string) => request<Prefix[]>(`/namespaces/${ns}/prefixes`)
  }
}

// ── Type definitions (mirror backend JSON) ──────────────────────────────
export interface Machine {
  name: string; namespace: string; state: string; power: string
  machineClass: string; image: string; ips: string[]; volumes: string[]; createdAt: string
}
export interface CreateMachineRequest {
  name: string; machineClass: string; image: string
  networkName: string; volumes: VolumeAttachment[]; power: string
}
export interface VolumeAttachment { name: string; sizeBytes: number; volumeClass: string }
export interface Volume {
  name: string; namespace: string; state: string; sizeBytes: number; volumeClass: string; createdAt: string
}
export interface CreateVolumeRequest { name: string; volumeClass: string; sizeBytes: number }
export interface Network { name: string; namespace: string; createdAt: string }
export interface NetworkInterface {
  name: string; namespace: string; state: string; ips: string[]; network: string; machine: string; createdAt: string
}
export interface VirtualIP { name: string; namespace: string; ip: string; type: string; ipFamily: string; createdAt: string }
export interface LoadBalancer { name: string; namespace: string; type: string; ips: string[]; createdAt: string }
export interface Prefix { name: string; namespace: string; prefix: string; phase: string; createdAt: string }
```

- [ ] **Step 5: Create `frontend/src/stores/namespace.ts`**

```typescript
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '@/api/client'

export const useNamespaceStore = defineStore('namespace', () => {
  const namespaces = ref<string[]>([])
  const active     = ref<string>('default')

  async function load() {
    namespaces.value = await api.namespaces.list()
    if (namespaces.value.length && !namespaces.value.includes(active.value)) {
      active.value = namespaces.value[0]
    }
  }

  function setActive(ns: string) { active.value = ns }

  return { namespaces, active, load, setActive }
})
```

- [ ] **Step 6: Create `frontend/src/router/index.ts`**

```typescript
import { createRouter, createWebHistory } from 'vue-router'
import DashboardLayout from '@/layouts/DashboardLayout.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: DashboardLayout,
      redirect: '/machines',
      children: [
        { path: 'machines',           component: () => import('@/views/MachinesView.vue') },
        { path: 'machines/new',       component: () => import('@/views/MachineCreateView.vue') },
        { path: 'machines/:name',     component: () => import('@/views/MachineDetailView.vue') },
        { path: 'volumes',            component: () => import('@/views/VolumesView.vue') },
        { path: 'networks',           component: () => import('@/views/NetworksView.vue') },
        { path: 'virtualips',         component: () => import('@/views/VirtualIPsView.vue') },
        { path: 'loadbalancers',      component: () => import('@/views/LoadBalancersView.vue') },
        { path: 'prefixes',           component: () => import('@/views/IPPrefixesView.vue') },
      ]
    }
  ]
})

export default router
```

- [ ] **Step 7: Create `frontend/src/layouts/DashboardLayout.vue`**

```vue
<template>
  <v-app :theme="'light'">
    <TopBar />
    <v-navigation-drawer permanent width="220">
      <Sidebar />
    </v-navigation-drawer>
    <v-main>
      <router-view />
    </v-main>
  </v-app>
</template>

<script setup lang="ts">
import TopBar from '@/components/TopBar.vue'
import Sidebar from '@/components/Sidebar.vue'
import { useNamespaceStore } from '@/stores/namespace'
import { onMounted } from 'vue'

const nsStore = useNamespaceStore()
onMounted(() => nsStore.load())
</script>
```

- [ ] **Step 8: Create `frontend/src/main.ts`**

```typescript
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import vuetify from './plugins/vuetify'

createApp(App).use(createPinia()).use(router).use(vuetify).mount('#app')
```

- [ ] **Step 9: Create `frontend/src/App.vue`**

```vue
<template><router-view /></template>
```

- [ ] **Step 10: Create `frontend/index.html`**

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>IronCore Dashboard</title>
  </head>
  <body>
    <div id="app"></div>
    <script type="module" src="/src/main.ts"></script>
  </body>
</html>
```

- [ ] **Step 11: Install deps and verify dev server starts**

```bash
cd frontend && npm install
npm run dev
```

Expected: Vite dev server at http://localhost:5173 — blank page (no views yet), no console errors.

- [ ] **Step 12: Commit**

```bash
git add frontend/
git commit -m "feat: Vue 3 + Vuetify 3 + Pinia + router scaffold"
```

---

## Task 06 — TopBar, Sidebar, NamespaceSwitcher components

**Agent prompt file:** `.agents/task-06-machines-view/PROMPT.md`

**Files:**
- Create: `frontend/src/components/TopBar.vue`
- Create: `frontend/src/components/Sidebar.vue`
- Create: `frontend/src/components/NamespaceSwitcher.vue`
- Create: `frontend/src/components/StatusBadge.vue`
- Create: `frontend/src/components/StatCard.vue`
- Create: `frontend/src/views/MachinesView.vue`
- Create: `frontend/src/stores/machines.ts`

**Context:** Match the mockup in `docs/architecture.md` — blue topbar (`#1a5fa8`), IronCore Fe logo in top-left, sidebar grouped by Compute/Storage/Networking/IPAM, namespace switcher dropdown in the top-right like Gardener.

- [ ] **Step 1: Create `frontend/src/components/StatusBadge.vue`**

```vue
<template>
  <v-chip :color="color" size="small" label>
    <v-icon start size="x-small">mdi-circle</v-icon>
    {{ label }}
  </v-chip>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ state: string }>()

const color = computed(() => ({
  Running: 'success', Available: 'success', Allocated: 'success',
  Pending: 'warning',
  Shutdown: 'default', Stopped: 'default',
  Error: 'error', Terminated: 'error', Terminating: 'error',
}[props.state] ?? 'default'))

const label = computed(() => props.state || '—')
</script>
```

- [ ] **Step 2: Create `frontend/src/components/StatCard.vue`**

```vue
<template>
  <v-card variant="outlined" rounded="lg">
    <v-card-text class="pa-4">
      <div class="text-h4 font-weight-bold" :style="{ color: valueColor }">{{ value }}</div>
      <div class="text-caption text-uppercase text-medium-emphasis mt-1">{{ label }}</div>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
defineProps<{ value: string | number; label: string; valueColor?: string }>()
</script>
```

- [ ] **Step 3: Create `frontend/src/components/NamespaceSwitcher.vue`**

```vue
<template>
  <v-menu>
    <template #activator="{ props }">
      <v-btn v-bind="props" variant="outlined" color="white" size="small" class="text-white border-white">
        <v-icon start>mdi-package-variant</v-icon>
        {{ nsStore.active }}
        <v-icon end>mdi-chevron-down</v-icon>
      </v-btn>
    </template>
    <v-list density="compact" min-width="200">
      <v-list-subheader>Switch Project / Namespace</v-list-subheader>
      <v-list-item
        v-for="ns in nsStore.namespaces"
        :key="ns"
        :value="ns"
        :active="ns === nsStore.active"
        active-color="primary"
        @click="nsStore.setActive(ns)"
      >
        <v-list-item-title>{{ ns }}</v-list-item-title>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<script setup lang="ts">
import { useNamespaceStore } from '@/stores/namespace'
const nsStore = useNamespaceStore()
</script>
```

- [ ] **Step 4: Create `frontend/src/components/TopBar.vue`**

```vue
<template>
  <v-app-bar color="primary" elevation="2" height="52">
    <template #prepend>
      <div class="d-flex align-center ga-2 px-4 border-e border-white border-opacity-20" style="min-width:220px">
        <v-avatar size="32" rounded="sm">
          <v-img src="/ironcore-logo.png" />
        </v-avatar>
        <div>
          <div class="text-white font-weight-bold" style="font-size:15px;line-height:1.1">IronCore</div>
          <div class="text-white text-opacity-60" style="font-size:10px">Cloud Dashboard</div>
        </div>
      </div>
    </template>

    <v-tabs color="white" selected-class="border-b-white" class="ml-2">
      <v-tab :to="{ path: '/machines' }" prepend-icon="mdi-monitor">Machines</v-tab>
      <v-tab :to="{ path: '/volumes' }"  prepend-icon="mdi-database">Volumes</v-tab>
      <v-tab :to="{ path: '/networks' }" prepend-icon="mdi-lan">Networking</v-tab>
      <v-tab :to="{ path: '/prefixes' }" prepend-icon="mdi-ip-network">IPAM</v-tab>
    </v-tabs>

    <template #append>
      <div class="d-flex align-center ga-3 pr-4">
        <NamespaceSwitcher />
        <div class="d-flex align-center ga-1 text-white text-opacity-70 text-caption">
          <v-icon size="10" color="success">mdi-circle</v-icon>
          ironcore-in-a-box
        </div>
        <v-btn icon="mdi-help-circle-outline" color="white" variant="text" size="small" />
        <v-avatar color="secondary" size="30" class="cursor-pointer">
          <span class="text-caption font-weight-bold">DV</span>
        </v-avatar>
      </div>
    </template>
  </v-app-bar>
</template>

<script setup lang="ts">
import NamespaceSwitcher from './NamespaceSwitcher.vue'
</script>
```

> Note: copy `/dashboard-workspace/ironcore-logo.png` to `frontend/public/ironcore-logo.png` so `<v-img src="/ironcore-logo.png" />` resolves.

- [ ] **Step 5: Create `frontend/src/components/Sidebar.vue`**

```vue
<template>
  <v-list density="compact" nav class="pt-2">
    <div class="text-caption text-uppercase text-medium-emphasis px-3 pt-2 pb-1 font-weight-bold">
      Compute
    </div>
    <v-list-item to="/machines"      prepend-icon="mdi-monitor"     title="Machines" rounded="lg" />

    <v-divider class="my-2" />
    <div class="text-caption text-uppercase text-medium-emphasis px-3 pt-2 pb-1 font-weight-bold">
      Storage
    </div>
    <v-list-item to="/volumes"       prepend-icon="mdi-database"    title="Volumes"   rounded="lg" />
    <v-list-item to="/snapshots"     prepend-icon="mdi-camera"      title="Snapshots" rounded="lg" />

    <v-divider class="my-2" />
    <div class="text-caption text-uppercase text-medium-emphasis px-3 pt-2 pb-1 font-weight-bold">
      Networking
    </div>
    <v-list-item to="/networks"      prepend-icon="mdi-lan"         title="Networks"       rounded="lg" />
    <v-list-item to="/virtualips"    prepend-icon="mdi-earth"       title="Virtual IPs"    rounded="lg" />
    <v-list-item to="/loadbalancers" prepend-icon="mdi-scale-balance" title="Load Balancers" rounded="lg" />

    <v-divider class="my-2" />
    <div class="text-caption text-uppercase text-medium-emphasis px-3 pt-2 pb-1 font-weight-bold">
      IP Management
    </div>
    <v-list-item to="/prefixes"      prepend-icon="mdi-ip-network"  title="IP Prefixes" rounded="lg" />

    <template #append>
      <v-divider class="my-2" />
      <v-list-item prepend-icon="mdi-cog-outline" title="Settings" rounded="lg" />
    </template>
  </v-list>
</template>
```

- [ ] **Step 6: Create `frontend/src/stores/machines.ts`**

```typescript
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api, type Machine } from '@/api/client'
import { useNamespaceStore } from './namespace'

export const useMachinesStore = defineStore('machines', () => {
  const items   = ref<Machine[]>([])
  const loading = ref(false)
  const error   = ref<string | null>(null)

  async function load() {
    const ns = useNamespaceStore().active
    loading.value = true; error.value = null
    try { items.value = await api.machines.list(ns) }
    catch (e: any) { error.value = e.message }
    finally { loading.value = false }
  }

  async function deleteMachine(name: string) {
    const ns = useNamespaceStore().active
    await api.machines.delete(ns, name)
    await load()
  }

  async function setPower(name: string, power: 'On' | 'Off') {
    const ns = useNamespaceStore().active
    await api.machines.power(ns, name, power)
    await load()
  }

  return { items, loading, error, load, deleteMachine, setPower }
})
```

- [ ] **Step 7: Create `frontend/src/views/MachinesView.vue`**

```vue
<template>
  <v-container fluid class="pa-6">
    <!-- Page header -->
    <div class="d-flex align-center justify-space-between mb-6">
      <div>
        <h1 class="text-h5 font-weight-bold">Machines</h1>
        <div class="text-caption text-medium-emphasis">{{ nsStore.active }}</div>
      </div>
      <div class="d-flex ga-2">
        <v-btn variant="outlined" prepend-icon="mdi-refresh" @click="store.load()">Refresh</v-btn>
        <v-btn color="primary" prepend-icon="mdi-plus" :to="{ path: '/machines/new' }">Create Machine</v-btn>
      </div>
    </div>

    <!-- Stat cards -->
    <v-row class="mb-6" dense>
      <v-col v-for="s in stats" :key="s.label" cols="auto">
        <StatCard :value="s.value" :label="s.label" :value-color="s.color" />
      </v-col>
    </v-row>

    <!-- Search -->
    <v-text-field
      v-model="search"
      prepend-inner-icon="mdi-magnify"
      placeholder="Search machines…"
      variant="outlined"
      density="compact"
      hide-details
      class="mb-4"
      style="max-width: 320px"
    />

    <!-- Table -->
    <v-card variant="outlined" rounded="lg">
      <v-card-title class="d-flex align-center ga-2 pa-4 bg-grey-lighten-5">
        <v-icon color="primary">mdi-monitor</v-icon>
        All Machines
        <v-chip size="x-small" color="primary">{{ store.items.length }}</v-chip>
        <v-spacer />
      </v-card-title>

      <v-data-table
        :headers="headers"
        :items="filtered"
        :loading="store.loading"
        density="compact"
        hover
        item-key="name"
        :no-data-text="store.error ?? 'No machines found'"
      >
        <template #item.state="{ item }">
          <StatusBadge :state="item.state" />
        </template>
        <template #item.machineClass="{ item }">
          <span class="font-weight-bold">{{ item.machineClass }}</span>
        </template>
        <template #item.ips="{ item }">
          <v-chip v-for="ip in item.ips" :key="ip" size="x-small" color="info" class="mr-1">{{ ip }}</v-chip>
        </template>
        <template #item.volumes="{ item }">
          <v-chip v-for="v in item.volumes" :key="v" size="x-small" variant="outlined" class="mr-1">{{ v }}</v-chip>
        </template>
        <template #item.actions="{ item }">
          <div class="d-flex ga-1">
            <v-btn
              size="x-small" variant="outlined"
              :icon="item.power === 'On' ? 'mdi-power-off' : 'mdi-power'"
              :color="item.power === 'On' ? 'warning' : 'success'"
              :title="item.power === 'On' ? 'Power Off' : 'Power On'"
              @click.stop="store.setPower(item.name, item.power === 'On' ? 'Off' : 'On')"
            />
            <v-btn
              size="x-small" variant="outlined" icon="mdi-delete" color="error"
              title="Delete"
              @click.stop="confirmDelete(item.name)"
            />
          </div>
        </template>
      </v-data-table>
    </v-card>

    <!-- Delete confirm dialog -->
    <v-dialog v-model="deleteDialog" max-width="420">
      <v-card>
        <v-card-title>Delete machine?</v-card-title>
        <v-card-text>Machine <strong>{{ pendingDelete }}</strong> will be permanently deleted.</v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="deleteDialog = false">Cancel</v-btn>
          <v-btn color="error" @click="doDelete">Delete</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useMachinesStore } from '@/stores/machines'
import { useNamespaceStore } from '@/stores/namespace'
import StatusBadge from '@/components/StatusBadge.vue'
import StatCard from '@/components/StatCard.vue'

const store   = useMachinesStore()
const nsStore = useNamespaceStore()
const search  = ref('')
const deleteDialog  = ref(false)
const pendingDelete = ref('')

const headers = [
  { title: 'Name',         key: 'name'         },
  { title: 'Status',       key: 'state'        },
  { title: 'Size',         key: 'machineClass' },
  { title: 'Image',        key: 'image'        },
  { title: 'Volumes',      key: 'volumes'      },
  { title: 'IP Addresses', key: 'ips'          },
  { title: '',             key: 'actions', sortable: false },
]

const filtered = computed(() =>
  store.items.filter(m => m.name.toLowerCase().includes(search.value.toLowerCase()))
)

const stats = computed(() => [
  { label: 'Running',  value: store.items.filter(m => m.state === 'Running').length,  color: '#16a34a' },
  { label: 'Pending',  value: store.items.filter(m => m.state === 'Pending').length,  color: '#d97706' },
  { label: 'Shutdown', value: store.items.filter(m => m.state === 'Shutdown').length, color: '#94a3b8' },
  { label: 'Total',    value: store.items.length,                                     color: '#1a2332' },
])

function confirmDelete(name: string) { pendingDelete.value = name; deleteDialog.value = true }
async function doDelete() { await store.deleteMachine(pendingDelete.value); deleteDialog.value = false }

onMounted(() => store.load())
watch(() => nsStore.active, () => store.load())
</script>
```

- [ ] **Step 8: Verify in browser**

```bash
# Backend running: make run -- --kubeconfig ~/.kube/config
cd frontend && npm run dev
```

Open http://localhost:5173 — should show the full layout: blue topbar with logo, sidebar, namespace switcher dropdown, Machines view with stats and table.

- [ ] **Step 9: Commit**

```bash
git add frontend/src/
git commit -m "feat: TopBar, Sidebar, NamespaceSwitcher, MachinesView with stats table"
```

---

## Task 07 — Create Machine wizard

**Agent prompt file:** `.agents/task-07-create-machine/PROMPT.md`

**Files:**
- Create: `frontend/src/views/MachineCreateView.vue`

The wizard has accordion sections (Name → Machine Type → OS Image → Network → Storage) with a sticky Summary panel on the right. On submit it calls `api.machines.create(ns, payload)` then navigates back to `/machines`.

- [ ] **Step 1: Fetch available machine classes and networks on mount**

In `frontend/src/api/client.ts`, add:

```typescript
machineClasses: {
  list: () => request<MachineClass[]>('/machineclasses')
},
```

Add to the backend `GET /api/v1/machineclasses` in `internal/api/machines.go`:

```go
// GET /api/v1/machineclasses (cluster-scoped, no namespace)
func (h *MachineHandler) ListMachineClasses(w http.ResponseWriter, r *http.Request) {
	list, err := h.cs.ComputeV1alpha1().MachineClasses().List(r.Context(), metav1.ListOptions{})
	if err != nil { writeError(w, http.StatusInternalServerError, err.Error()); return }
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
```

Mount in `server.go`: `r.Get("/api/v1/machineclasses", mh.ListMachineClasses)`

Add `MachineClass` type to `client.ts`:

```typescript
export interface MachineClass { name: string; cpu: string; ram: string }
```

- [ ] **Step 2: Create `frontend/src/views/MachineCreateView.vue`**

```vue
<template>
  <v-container fluid class="pa-6">
    <div class="d-flex align-center ga-2 mb-2 text-caption text-medium-emphasis">
      <router-link to="/machines" class="text-primary text-decoration-none">Machines</router-link>
      <span>›</span> Launch a Machine
    </div>
    <h1 class="text-h5 font-weight-bold mb-1">Launch a Machine</h1>
    <p class="text-medium-emphasis mb-6">Configure and launch a new IronCore virtual machine.</p>

    <div class="d-flex ga-6 align-start">
      <!-- ── Left: accordion form ── -->
      <div style="flex:1;min-width:0">

        <!-- Name -->
        <v-card variant="outlined" rounded="lg" class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold pa-4 pb-2">Name</v-card-title>
          <v-card-text>
            <v-text-field v-model="form.name" label="Machine name" variant="outlined" density="compact"
              placeholder="e.g. web-server-01" :rules="[v => !!v || 'Required']" />
          </v-card-text>
        </v-card>

        <!-- Machine type -->
        <v-card variant="outlined" rounded="lg" class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold pa-4 pb-2">Machine Type</v-card-title>
          <v-card-text>
            <v-radio-group v-model="form.machineClass" hide-details>
              <v-table density="compact">
                <thead>
                  <tr>
                    <th></th><th>Type</th><th>vCPU</th><th>Memory</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="mc in machineClasses" :key="mc.name"
                    class="cursor-pointer"
                    :class="{ 'bg-primary-lighten-5': form.machineClass === mc.name }"
                    @click="form.machineClass = mc.name">
                    <td><v-radio :value="mc.name" hide-details /></td>
                    <td><strong>{{ mc.name }}</strong></td>
                    <td>{{ mc.cpu }}</td>
                    <td>{{ mc.ram }}</td>
                  </tr>
                </tbody>
              </v-table>
            </v-radio-group>
          </v-card-text>
        </v-card>

        <!-- OS Image -->
        <v-card variant="outlined" rounded="lg" class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold pa-4 pb-2">OS Image</v-card-title>
          <v-card-text>
            <v-select v-model="form.image" :items="imageOptions" label="OS Image"
              variant="outlined" density="compact" />
          </v-card-text>
        </v-card>

        <!-- Network -->
        <v-card variant="outlined" rounded="lg" class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold pa-4 pb-2">Network</v-card-title>
          <v-card-text>
            <v-select v-model="form.networkName" :items="networkNames" label="Network"
              variant="outlined" density="compact" />
          </v-card-text>
        </v-card>

        <!-- Volumes -->
        <v-card variant="outlined" rounded="lg" class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold pa-4 pb-2">Storage</v-card-title>
          <v-card-text>
            <div v-for="(vol, i) in form.volumes" :key="i" class="d-flex ga-3 align-center mb-3">
              <v-text-field v-model="vol.name" label="Volume name" variant="outlined" density="compact" hide-details />
              <v-text-field v-model.number="vol.sizeGiB" label="Size (GiB)" type="number" variant="outlined"
                density="compact" hide-details style="max-width:110px" />
              <v-select v-model="vol.volumeClass" :items="['standard','fast-ssd']" label="Class"
                variant="outlined" density="compact" hide-details style="max-width:130px" />
              <v-btn icon="mdi-delete" size="small" variant="text" color="error"
                :disabled="i === 0" @click="form.volumes.splice(i, 1)" />
            </div>
            <v-btn variant="dashed" color="primary" prepend-icon="mdi-plus" size="small" @click="addVolume">
              Add volume
            </v-btn>
          </v-card-text>
        </v-card>

        <v-alert v-if="submitError" type="error" class="mb-4">{{ submitError }}</v-alert>

        <div class="d-flex ga-3">
          <v-btn color="primary" :loading="submitting" @click="submit">Launch Machine</v-btn>
          <v-btn variant="outlined" :to="{ path: '/machines' }">Cancel</v-btn>
        </div>
      </div>

      <!-- ── Right: Summary panel ── -->
      <v-card variant="outlined" rounded="lg" style="width:260px;position:sticky;top:80px">
        <v-card-title class="text-subtitle-2 font-weight-bold pa-4 pb-2">Summary</v-card-title>
        <v-divider />
        <v-list density="compact" class="pa-2">
          <v-list-item title="Name"         :subtitle="form.name || '—'" />
          <v-list-item title="Machine type" :subtitle="form.machineClass || '—'" />
          <v-list-item title="OS Image"     :subtitle="form.image || '—'" />
          <v-list-item title="Network"      :subtitle="form.networkName || '—'" />
          <v-list-item title="Volumes"
            :subtitle="form.volumes.map(v => `${v.name} ${v.sizeGiB}GiB`).join(', ')" />
        </v-list>
        <v-card-actions class="pa-4 pt-0">
          <v-btn color="primary" block :loading="submitting" @click="submit">Launch Machine</v-btn>
        </v-card-actions>
      </v-card>
    </div>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api, type MachineClass, type Network } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'

const router  = useRouter()
const nsStore = useNamespaceStore()

const machineClasses = ref<MachineClass[]>([])
const networks       = ref<Network[]>([])
const submitting     = ref(false)
const submitError    = ref('')

const imageOptions = [
  'ubuntu-22.04', 'ubuntu-24.04', 'debian-12', 'almalinux-9'
]

const networkNames = computed(() => networks.value.map(n => n.name))

const form = ref({
  name: '',
  machineClass: '',
  image: 'ubuntu-22.04',
  networkName: '',
  volumes: [{ name: 'root', sizeGiB: 50, volumeClass: 'standard' }]
})

function addVolume() {
  form.value.volumes.push({ name: `data-${form.value.volumes.length}`, sizeGiB: 100, volumeClass: 'standard' })
}

async function submit() {
  if (!form.value.name || !form.value.machineClass || !form.value.networkName) {
    submitError.value = 'Name, machine type, and network are required.'; return
  }
  submitting.value = true; submitError.value = ''
  try {
    await api.machines.create(nsStore.active, {
      name:         form.value.name,
      machineClass: form.value.machineClass,
      image:        form.value.image,
      networkName:  form.value.networkName,
      power:        'On',
      volumes:      form.value.volumes.map(v => ({
        name:        v.name,
        sizeBytes:   v.sizeGiB * 1024 * 1024 * 1024,
        volumeClass: v.volumeClass
      }))
    })
    router.push('/machines')
  } catch (e: any) {
    submitError.value = e.message
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  machineClasses.value = await api.machineClasses.list()
  networks.value       = await api.networks.list(nsStore.active)
  if (machineClasses.value.length) form.value.machineClass = machineClasses.value[0].name
  if (networks.value.length)       form.value.networkName  = networks.value[0].name
})
</script>
```

- [ ] **Step 3: Verify end-to-end**

```bash
# With backend running against ironcore-in-a-box:
# 1. Open http://localhost:5173/machines/new
# 2. Fill in name, pick a machine class, pick network, click Launch
# 3. Should redirect to /machines and show new machine with Pending state
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/views/MachineCreateView.vue internal/api/machines.go
git commit -m "feat: Create Machine wizard with summary panel"
```

---

## Task 08 — Volumes, Networking, IPAM views

**Agent prompt file:** `.agents/task-08-volumes-networking-views/PROMPT.md`

**Files:**
- Create: `frontend/src/stores/volumes.ts`
- Create: `frontend/src/views/VolumesView.vue`
- Create: `frontend/src/views/NetworksView.vue`
- Create: `frontend/src/views/VirtualIPsView.vue`
- Create: `frontend/src/views/LoadBalancersView.vue`
- Create: `frontend/src/views/IPPrefixesView.vue`

Each view follows the same pattern as `MachinesView.vue`: page header with title + namespace, optional stat cards, search field, `v-data-table` with appropriate columns, `StatusBadge` for state fields.

- [ ] **Step 1: Create `frontend/src/stores/volumes.ts`**

```typescript
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api, type Volume } from '@/api/client'
import { useNamespaceStore } from './namespace'

export const useVolumesStore = defineStore('volumes', () => {
  const items   = ref<Volume[]>([])
  const loading = ref(false)
  const error   = ref<string | null>(null)

  async function load() {
    const ns = useNamespaceStore().active
    loading.value = true; error.value = null
    try { items.value = await api.volumes.list(ns) }
    catch (e: any) { error.value = e.message }
    finally { loading.value = false }
  }

  async function deleteVolume(name: string) {
    await api.volumes.delete(useNamespaceStore().active, name)
    await load()
  }

  return { items, loading, error, load, deleteVolume }
})
```

- [ ] **Step 2: Create `frontend/src/views/VolumesView.vue`**

```vue
<template>
  <v-container fluid class="pa-6">
    <div class="d-flex align-center justify-space-between mb-6">
      <div>
        <h1 class="text-h5 font-weight-bold">Volumes</h1>
        <div class="text-caption text-medium-emphasis">{{ nsStore.active }}</div>
      </div>
      <v-btn color="primary" prepend-icon="mdi-plus">Create Volume</v-btn>
    </div>

    <v-card variant="outlined" rounded="lg">
      <v-card-title class="d-flex align-center ga-2 pa-4 bg-grey-lighten-5">
        <v-icon color="primary">mdi-database</v-icon>
        Volumes
        <v-chip size="x-small" color="primary">{{ store.items.length }}</v-chip>
      </v-card-title>
      <v-data-table
        :headers="headers" :items="store.items"
        :loading="store.loading" density="compact" hover
        :no-data-text="store.error ?? 'No volumes'"
      >
        <template #item.state="{ item }"><StatusBadge :state="item.state" /></template>
        <template #item.sizeBytes="{ item }">{{ formatBytes(item.sizeBytes) }}</template>
        <template #item.actions="{ item }">
          <v-btn size="x-small" variant="outlined" icon="mdi-delete" color="error"
            @click="store.deleteVolume(item.name)" />
        </template>
      </v-data-table>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { useVolumesStore } from '@/stores/volumes'
import { useNamespaceStore } from '@/stores/namespace'
import StatusBadge from '@/components/StatusBadge.vue'

const store   = useVolumesStore()
const nsStore = useNamespaceStore()

const headers = [
  { title: 'Name',         key: 'name'        },
  { title: 'State',        key: 'state'       },
  { title: 'Size',         key: 'sizeBytes'   },
  { title: 'Volume Class', key: 'volumeClass' },
  { title: 'Created',      key: 'createdAt'   },
  { title: '',             key: 'actions', sortable: false },
]

function formatBytes(b: number) {
  if (!b) return '—'
  const gb = b / 1024 / 1024 / 1024
  return gb >= 1000 ? `${(gb/1024).toFixed(1)} TB` : `${Math.round(gb)} GB`
}

onMounted(() => store.load())
watch(() => nsStore.active, () => store.load())
</script>
```

- [ ] **Step 3: Create `frontend/src/views/NetworksView.vue`**

```vue
<template>
  <v-container fluid class="pa-6">
    <h1 class="text-h5 font-weight-bold mb-6">Networking</h1>

    <v-card variant="outlined" rounded="lg" class="mb-6">
      <v-card-title class="d-flex align-center ga-2 pa-4 bg-grey-lighten-5">
        <v-icon color="primary">mdi-lan</v-icon> Networks
        <v-chip size="x-small" color="primary">{{ networks.length }}</v-chip>
      </v-card-title>
      <v-data-table :headers="netHeaders" :items="networks" density="compact" hover :loading="loading">
        <template #item.createdAt="{ item }">{{ item.createdAt }}</template>
      </v-data-table>
    </v-card>

    <v-card variant="outlined" rounded="lg">
      <v-card-title class="d-flex align-center ga-2 pa-4 bg-grey-lighten-5">
        <v-icon color="primary">mdi-network-outline</v-icon> Network Interfaces
        <v-chip size="x-small" color="primary">{{ interfaces.length }}</v-chip>
      </v-card-title>
      <v-data-table :headers="nicHeaders" :items="interfaces" density="compact" hover :loading="loading">
        <template #item.state="{ item }"><StatusBadge :state="item.state" /></template>
        <template #item.ips="{ item }">
          <v-chip v-for="ip in item.ips" :key="ip" size="x-small" color="info" class="mr-1">{{ ip }}</v-chip>
        </template>
      </v-data-table>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type Network, type NetworkInterface } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'
import StatusBadge from '@/components/StatusBadge.vue'

const nsStore    = useNamespaceStore()
const networks   = ref<Network[]>([])
const interfaces = ref<NetworkInterface[]>([])
const loading    = ref(false)

const netHeaders = [
  { title: 'Name', key: 'name' }, { title: 'Created', key: 'createdAt' }
]
const nicHeaders = [
  { title: 'Name', key: 'name' }, { title: 'State', key: 'state' },
  { title: 'IPs', key: 'ips' },   { title: 'Network', key: 'network' },
  { title: 'Machine', key: 'machine' },
]

async function load() {
  loading.value = true
  const ns = nsStore.active
  ;[networks.value, interfaces.value] = await Promise.all([
    api.networks.list(ns), api.networks.listInterfaces(ns)
  ])
  loading.value = false
}

onMounted(load)
watch(() => nsStore.active, load)
</script>
```

- [ ] **Step 4: Create remaining views (VirtualIPsView, LoadBalancersView, IPPrefixesView)**

Each follows the exact same pattern. Create `frontend/src/views/VirtualIPsView.vue`:

```vue
<template>
  <v-container fluid class="pa-6">
    <h1 class="text-h5 font-weight-bold mb-6">Virtual IPs</h1>
    <v-card variant="outlined" rounded="lg">
      <v-card-title class="d-flex align-center ga-2 pa-4 bg-grey-lighten-5">
        <v-icon color="primary">mdi-earth</v-icon> Virtual IPs
        <v-chip size="x-small" color="primary">{{ items.length }}</v-chip>
      </v-card-title>
      <v-data-table :headers="headers" :items="items" density="compact" hover :loading="loading">
        <template #item.ip="{ item }">
          <v-chip v-if="item.ip" size="x-small" color="deep-purple" class="font-weight-bold">{{ item.ip }}</v-chip>
          <span v-else class="text-medium-emphasis">—</span>
        </template>
      </v-data-table>
    </v-card>
  </v-container>
</template>
<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type VirtualIP } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'
const nsStore = useNamespaceStore()
const items   = ref<VirtualIP[]>([])
const loading = ref(false)
const headers = [
  { title: 'Name', key: 'name' }, { title: 'IP', key: 'ip' },
  { title: 'Type', key: 'type' }, { title: 'IP Family', key: 'ipFamily' },
  { title: 'Created', key: 'createdAt' },
]
async function load() { loading.value=true; items.value = await api.virtualIPs.list(nsStore.active); loading.value=false }
onMounted(load); watch(() => nsStore.active, load)
</script>
```

Create `frontend/src/views/LoadBalancersView.vue` with columns: Name / Type / IPs / Created.
Create `frontend/src/views/IPPrefixesView.vue` with columns: Name / Prefix / Phase / Created. Use `StatusBadge` for Phase.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/stores/volumes.ts frontend/src/views/
git commit -m "feat: Volumes, Networks, VIPs, Load Balancers, IPAM views"
```

---

## Task 09 — Go backend serves built frontend + integration

**Agent prompt file:** `.agents/task-09-namespace-switcher/PROMPT.md`

**Files:**
- Modify: `internal/server/server.go` (serve static frontend)
- Modify: `Makefile` (build frontend then embed)
- Create: `frontend/tsconfig.json` (path aliases)

- [ ] **Step 1: Add `k8s.io/client-go` for namespace listing**

In `internal/server/server.go`, add a standard k8s clientset alongside the IronCore one:

```go
import k8s "k8s.io/client-go/kubernetes"

type Server struct {
    router   *chi.Mux
    ironcore versioned.Interface
    k8s      *k8s.Clientset
}
```

In `cmd/server/main.go`:

```go
k8sCfg, _ := clientcmd.BuildConfigFromFlags("", *kubeconfig)
k8sClient, _ := k8s.NewForConfig(k8sCfg)
srv := server.New(cs, k8sClient)
```

Add the namespace list handler:

```go
r.Get("/api/v1/namespaces", func(w http.ResponseWriter, r *http.Request) {
    list, err := s.k8s.CoreV1().Namespaces().List(r.Context(), metav1.ListOptions{})
    if err != nil { writeError(w, http.StatusInternalServerError, err.Error()); return }
    names := make([]string, 0, len(list.Items))
    for _, ns := range list.Items { names = append(names, ns.Name) }
    writeJSON(w, http.StatusOK, names)
})
```

- [ ] **Step 2: Serve the Vue build from Go using `embed`**

In `internal/server/server.go`, add static file serving at the bottom of the route setup:

```go
import "embed"
import "io/fs"

//go:embed ../../dist/frontend
var frontendFS embed.FS

// After all API routes:
sub, _ := fs.Sub(frontendFS, "dist/frontend")
r.Handle("/*", http.FileServer(http.FS(sub)))
```

- [ ] **Step 3: Update `Makefile`**

```makefile
.PHONY: build build-frontend run dev tidy test

build-frontend:
	cd frontend && npm install && npm run build

build: build-frontend
	go build -o bin/ironcore-dashboard ./cmd/server

run:
	go run ./cmd/server --addr :8080 --kubeconfig $(HOME)/.kube/config

dev-backend:
	go run ./cmd/server --addr :8080 --kubeconfig $(HOME)/.kube/config

dev-frontend:
	cd frontend && npm run dev

test:
	go test ./...

tidy:
	go mod tidy
```

- [ ] **Step 4: Full integration test**

```bash
make build
./bin/ironcore-dashboard --addr :8080 --kubeconfig ~/.kube/config
```

Open http://localhost:8080 — full dashboard served from single binary, no separate Vite server needed.

- [ ] **Step 5: Commit**

```bash
git add .
git commit -m "feat: serve built Vue SPA from Go binary, k8s namespace listing"
```

---

## Task 10 — Machine detail view + final polish

**Agent prompt file:** `.agents/task-10-integration/PROMPT.md`

**Files:**
- Create: `frontend/src/views/MachineDetailView.vue`

- [ ] **Step 1: Create `frontend/src/views/MachineDetailView.vue`**

```vue
<template>
  <v-container fluid class="pa-6">
    <div class="d-flex align-center ga-2 mb-4 text-caption text-medium-emphasis">
      <router-link to="/machines" class="text-primary text-decoration-none">Machines</router-link>
      <span>›</span> {{ route.params.name }}
    </div>

    <div v-if="machine" class="d-flex align-center justify-space-between mb-6">
      <div>
        <h1 class="text-h5 font-weight-bold">{{ machine.name }}</h1>
        <StatusBadge :state="machine.state" class="mt-1" />
      </div>
      <div class="d-flex ga-2">
        <v-btn variant="outlined"
          :prepend-icon="machine.power === 'On' ? 'mdi-power-off' : 'mdi-power'"
          :color="machine.power === 'On' ? 'warning' : 'success'"
          @click="togglePower">
          {{ machine.power === 'On' ? 'Power Off' : 'Power On' }}
        </v-btn>
        <v-btn variant="outlined" color="error" prepend-icon="mdi-delete" @click="doDelete">Delete</v-btn>
      </div>
    </div>

    <v-row v-if="machine">
      <v-col cols="12" md="6">
        <v-card variant="outlined" rounded="lg">
          <v-card-title class="pa-4 text-subtitle-1 font-weight-bold">Details</v-card-title>
          <v-divider />
          <v-list density="compact">
            <v-list-item title="Machine type"  :subtitle="machine.machineClass" />
            <v-list-item title="OS Image"      :subtitle="machine.image" />
            <v-list-item title="Power state"   :subtitle="machine.power" />
            <v-list-item title="Namespace"     :subtitle="machine.namespace" />
            <v-list-item title="Created"       :subtitle="machine.createdAt" />
          </v-list>
        </v-card>
      </v-col>
      <v-col cols="12" md="6">
        <v-card variant="outlined" rounded="lg">
          <v-card-title class="pa-4 text-subtitle-1 font-weight-bold">Network</v-card-title>
          <v-divider />
          <v-card-text>
            <div v-if="machine.ips.length">
              <v-chip v-for="ip in machine.ips" :key="ip" color="info" size="small" class="mr-1 mb-1">{{ ip }}</v-chip>
            </div>
            <span v-else class="text-medium-emphasis">No IPs assigned</span>
          </v-card-text>
        </v-card>

        <v-card variant="outlined" rounded="lg" class="mt-4">
          <v-card-title class="pa-4 text-subtitle-1 font-weight-bold">Volumes</v-card-title>
          <v-divider />
          <v-card-text>
            <v-chip v-for="v in machine.volumes" :key="v" variant="outlined" size="small" class="mr-1 mb-1">
              {{ v }}
            </v-chip>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-progress-circular v-else indeterminate color="primary" class="mt-10 d-block mx-auto" />
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, type Machine } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'
import StatusBadge from '@/components/StatusBadge.vue'

const route   = useRoute()
const router  = useRouter()
const nsStore = useNamespaceStore()
const machine = ref<Machine | null>(null)

onMounted(async () => {
  machine.value = await api.machines.get(nsStore.active, route.params.name as string)
})

async function togglePower() {
  if (!machine.value) return
  const next = machine.value.power === 'On' ? 'Off' : 'On'
  machine.value = await api.machines.power(nsStore.active, machine.value.name, next)
}

async function doDelete() {
  if (!machine.value) return
  await api.machines.delete(nsStore.active, machine.value.name)
  router.push('/machines')
}
</script>
```

- [ ] **Step 2: Final build and smoke test**

```bash
make build
./bin/ironcore-dashboard --addr :8080 --kubeconfig ~/.kube/config
```

Checklist:
- [ ] http://localhost:8080 loads the dashboard
- [ ] Namespace switcher dropdown shows available namespaces
- [ ] Machines list loads (empty `[]` if no machines, no errors)
- [ ] Click "Create Machine" → wizard loads with machine classes and networks from the API
- [ ] Create a machine → redirected to list, machine appears with Pending state
- [ ] Click machine name → detail view shows
- [ ] Power on/off works
- [ ] Delete works
- [ ] Volumes tab shows volumes
- [ ] Networks tab shows networks + interfaces
- [ ] Virtual IPs, Load Balancers, IPAM tabs load without errors

- [ ] **Step 3: Final commit**

```bash
git add .
git commit -m "feat: machine detail view, full integration smoke test passing"
```

---

## Parallelism map

These tasks can run in parallel once their dependencies are met:

```
Task 01 (backend scaffold)
    └── Task 02 (ironcore client)
            └── Task 03 (machines API)  ──┐
            └── Task 04 (other APIs)    ──┤
                                          ├── Task 09 (integration + static serve)
Task 05 (frontend scaffold)               │       └── Task 10 (detail + polish)
    └── Task 06 (TopBar + MachinesView) ──┤
    └── Task 07 (create wizard)         ──┤
    └── Task 08 (other views)           ──┘
```

**Safe to run in parallel:**
- Task 03 + Task 04 (both extend the API, different files)
- Task 06 + Task 07 + Task 08 (all frontend views, no shared file edits)

**Must be sequential:**
- Task 01 → 02 → 03/04
- Task 05 → 06/07/08
- All of the above → Task 09 → Task 10