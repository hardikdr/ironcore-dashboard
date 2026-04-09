# Task 04 — Volumes, Networking, and IPAM API Endpoints

## Prerequisite

Task 02 (IronCore client-go wrapper) must be complete. Verify:

```bash
make build  # must succeed
```

This task runs **in parallel with Task 03** — you will each add different handler files. Task 03 creates `internal/api/types.go` (shared types). If Task 03 is not done yet, create `types.go` yourself using the content below — it is idempotent, both tasks produce the same file.

## Your job

Implement list/create/delete endpoints for Volumes, Networks, NetworkInterfaces, VirtualIPs, LoadBalancers, and IP Prefixes. Also add the `/api/v1/namespaces` endpoint (namespace list for the project switcher).

By the end:
- `GET  /api/v1/namespaces` → `[]string` (namespace names)
- `GET  /api/v1/namespaces/{ns}/volumes` → `[]VolumeResponse`
- `POST /api/v1/namespaces/{ns}/volumes` → `VolumeResponse` (201)
- `DELETE /api/v1/namespaces/{ns}/volumes/{name}` → 204
- `GET  /api/v1/namespaces/{ns}/networks` → `[]NetworkResponse`
- `GET  /api/v1/namespaces/{ns}/networkinterfaces` → `[]NetworkInterfaceResponse`
- `GET  /api/v1/namespaces/{ns}/virtualips` → `[]VirtualIPResponse`
- `GET  /api/v1/namespaces/{ns}/loadbalancers` → `[]LoadBalancerResponse`
- `GET  /api/v1/namespaces/{ns}/prefixes` → `[]PrefixResponse`

## Key IronCore types

**Volumes** (`github.com/ironcore-dev/ironcore/api/storage/v1alpha1`):
- `cs.StorageV1alpha1().Volumes(ns)`
- `Volume.Status.State` → `"Pending"`, `"Available"`, `"Error"`
- `Volume.Status.Resources[corev1alpha1.ResourceStorage]` → `resource.Quantity` (call `.Value()` for bytes)
- `Volume.Spec.VolumeClassRef.Name` → string

**Networks** (`github.com/ironcore-dev/ironcore/api/networking/v1alpha1`):
- `cs.NetworkingV1alpha1().Networks(ns)` — no state field, just list
- `cs.NetworkingV1alpha1().NetworkInterfaces(ns)`
- `NetworkInterface.Status.IPs` → slice (call `.String()` on each)
- `NetworkInterface.Status.State` → string
- `NetworkInterface.Spec.NetworkRef.Name` → string
- `NetworkInterface.Spec.MachineRef` → `*corev1.LocalObjectReference` (may be nil)

**VirtualIPs** (`networking/v1alpha1`):
- `cs.NetworkingV1alpha1().VirtualIPs(ns)`
- `VirtualIP.Status.IP` → `*commonv1alpha1.IP` (may be nil, call `.String()`)
- `VirtualIP.Spec.Type` → string
- `VirtualIP.Spec.IPFamily` → `corev1.IPFamily`

**LoadBalancers** (`networking/v1alpha1`):
- `cs.NetworkingV1alpha1().LoadBalancers(ns)`
- `LoadBalancer.Status.IPs` → slice of `commonv1alpha1.IP`
- `LoadBalancer.Spec.Type` → string

**Prefixes** (`github.com/ironcore-dev/ironcore/api/ipam/v1alpha1`):
- `cs.IpamV1alpha1().Prefixes(ns)`
- `Prefix.Spec.Prefix` → `*commonv1alpha1.IPPrefix` (may be nil, call `.String()`)
- `Prefix.Status.Phase` → `"Pending"`, `"Allocated"`

**Namespace listing** uses the standard k8s client-go:
- Import `k8s.io/client-go/kubernetes`
- `k8sClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})`

## Files to create / modify

| Action | File |
|--------|------|
| Create (if not exists) | `internal/api/types.go` |
| Create | `internal/api/volumes.go` |
| Create | `internal/api/networks.go` |
| Create | `internal/api/virtualips.go` |
| Create | `internal/api/loadbalancers.go` |
| Create | `internal/api/ipam.go` |
| Modify | `internal/server/server.go` (add k8s client, mount all routes) |
| Modify | `cmd/server/main.go` (create k8s client alongside ironcore client) |

## Step-by-step

### Step 1 — Ensure `internal/api/types.go` exists

If it doesn't exist yet, create it (same file as Task 03 — safe to create):

```go
package api

import (
	"encoding/json"
	"net/http"
)

// ... (all response structs: MachineResponse, VolumeResponse, NetworkResponse,
// NetworkInterfaceResponse, VirtualIPResponse, LoadBalancerResponse, PrefixResponse,
// CreateMachineRequest, VolumeAttachment, PatchPowerRequest)
// ... writeJSON() and writeError() helpers

// Full content is in the implementation plan at:
// docs/superpowers/plans/2026-04-08-ironcore-dashboard-v1.md — Task 03, Step 1
```

### Step 2 — Write `internal/api/volumes.go`

```go
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
		State:     string(v.Status.State),
		SizeBytes: sizeBytes, VolumeClass: vc,
		CreatedAt: v.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
	}
}
```

### Step 3 — Write `internal/api/networks.go`

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
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
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
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]NetworkInterfaceResponse, 0, len(list.Items))
	for _, ni := range list.Items {
		ips := []string{}
		for _, ip := range ni.Status.IPs {
			ips = append(ips, ip.String())
		}
		machine := ""
		if ni.Spec.MachineRef != nil {
			machine = ni.Spec.MachineRef.Name
		}
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

### Step 4 — Write `internal/api/virtualips.go`

```go
package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	versioned "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	metav1    "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type VirtualIPHandler struct{ cs versioned.Interface }

func NewVirtualIPHandler(cs versioned.Interface) *VirtualIPHandler {
	return &VirtualIPHandler{cs: cs}
}

func (h *VirtualIPHandler) List(w http.ResponseWriter, r *http.Request) {
	ns := chi.URLParam(r, "ns")
	list, err := h.cs.NetworkingV1alpha1().VirtualIPs(ns).List(r.Context(), metav1.ListOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]VirtualIPResponse, 0, len(list.Items))
	for _, v := range list.Items {
		ip := ""
		if v.Status.IP != nil {
			ip = v.Status.IP.String()
		}
		resp = append(resp, VirtualIPResponse{
			Name: v.Name, Namespace: v.Namespace,
			IP: ip, Type: string(v.Spec.Type),
			IPFamily:  string(v.Spec.IPFamily),
			CreatedAt: v.CreationTimestamp.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}
```

### Step 5 — Write `internal/api/loadbalancers.go`

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
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]LoadBalancerResponse, 0, len(list.Items))
	for _, lb := range list.Items {
		ips := []string{}
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
```

### Step 6 — Write `internal/api/ipam.go`

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
```

### Step 7 — Add k8s Clientset to server and mount all routes

Update `internal/server/server.go` — add a standard k8s clientset and namespace listing:

```go
import (
    // ... existing imports ...
    k8s  "k8s.io/client-go/kubernetes"
    "github.com/ironcore-dev/ironcore-dashboard/internal/api"
)

type Server struct {
    router   *chi.Mux
    ironcore versioned.Interface
    k8s      *k8s.Clientset
}

func New(cs versioned.Interface, k8sClient *k8s.Clientset) *Server {
    s := &Server{ironcore: cs, k8s: k8sClient}
    r := chi.NewRouter()
    // ... middleware ...

    r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("ok"))
    })

    // Namespace list
    r.Get("/api/v1/namespaces", func(w http.ResponseWriter, r *http.Request) {
        list, err := s.k8s.CoreV1().Namespaces().List(r.Context(), metav1.ListOptions{})
        if err != nil {
            api.WriteErrorPublic(w, http.StatusInternalServerError, err.Error())
            return
        }
        names := make([]string, 0, len(list.Items))
        for _, ns := range list.Items {
            names = append(names, ns.Name)
        }
        api.WriteJSONPublic(w, http.StatusOK, names)
    })

    vh  := api.NewVolumeHandler(cs)
    nh  := api.NewNetworkHandler(cs)
    vip := api.NewVirtualIPHandler(cs)
    lb  := api.NewLoadBalancerHandler(cs)
    iph := api.NewIPAMHandler(cs)

    r.Route("/api/v1/namespaces/{ns}", func(r chi.Router) {
        r.Get("/volumes",            vh.List)
        r.Post("/volumes",           vh.Create)
        r.Delete("/volumes/{name}",  vh.Delete)
        r.Get("/networks",           nh.ListNetworks)
        r.Get("/networkinterfaces",  nh.ListNetworkInterfaces)
        r.Get("/virtualips",         vip.List)
        r.Get("/loadbalancers",      lb.List)
        r.Get("/prefixes",           iph.ListPrefixes)
    })

    s.router = r
    return s
}
```

> Note: `writeJSON` and `writeError` are package-private in `api`. If you need them in `server.go`, either make them exported (`WriteJSON`, `WriteError`) in `types.go` or inline small responses directly. The namespace list handler can use a local helper.

Update `cmd/server/main.go` to create both clients:

```go
import (
    "flag"
    "log"
    ironclient "github.com/ironcore-dev/ironcore-dashboard/internal/client"
    "github.com/ironcore-dev/ironcore-dashboard/internal/server"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
)

func main() {
    addr       := flag.String("addr", ":8080", "HTTP listen address")
    kubeconfig := flag.String("kubeconfig", "", "path to kubeconfig")
    flag.Parse()

    cs, err := ironclient.New(*kubeconfig)
    if err != nil {
        log.Fatalf("ironcore client: %v", err)
    }

    k8sCfg, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
    if err != nil {
        log.Fatalf("k8s config: %v", err)
    }
    k8sClient, err := kubernetes.NewForConfig(k8sCfg)
    if err != nil {
        log.Fatalf("k8s client: %v", err)
    }

    srv := server.New(cs, k8sClient)
    log.Printf("IronCore Dashboard listening on %s", *addr)
    if err := srv.ListenAndServe(*addr); err != nil {
        log.Fatalf("server error: %v", err)
    }
}
```

### Step 8 — Verify all endpoints compile and return empty arrays

```bash
make build
go run ./cmd/server --kubeconfig ~/.kube/config &
sleep 2
curl http://localhost:8080/api/v1/namespaces
curl http://localhost:8080/api/v1/namespaces/default/volumes
curl http://localhost:8080/api/v1/namespaces/default/networks
curl http://localhost:8080/api/v1/namespaces/default/virtualips
curl http://localhost:8080/api/v1/namespaces/default/prefixes
```

Expected: each returns `[]` (empty JSON array)

### Step 9 — Commit

```bash
git add internal/api/ internal/server/server.go cmd/server/main.go
git commit -m "feat: volumes, networks, VIPs, LBs, IPAM list endpoints + namespace API"
```

## Done criteria

- All 9 new endpoints exist and return valid JSON
- `make build` succeeds with no errors
- `GET /api/v1/namespaces` returns a list of namespace names