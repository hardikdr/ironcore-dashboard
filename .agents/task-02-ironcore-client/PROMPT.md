# Task 02 — IronCore client-go Wrapper

## Prerequisite

Task 01 (backend scaffold) must be complete. Verify:

```bash
curl http://localhost:8080/healthz  # should return: ok
ls go.mod                           # should exist
```

## Your job

Wire the IronCore `client-go` typed clientset into the Go server. By the end:
- `internal/client/ironcore.go` builds an IronCore clientset from a kubeconfig path (or in-cluster)
- `cmd/server/main.go` accepts a `--kubeconfig` flag and passes the clientset to the server
- `internal/server/server.go` accepts `versioned.Interface` in its constructor

## Working directory

All work in `ironcore-dashboard/`. The IronCore module is at `../ironcore` (already linked via `replace` in `go.mod`).

The key import path for the clientset:
```
github.com/ironcore-dev/ironcore/client-go/ironcore/versioned
```

The clientset interface (`versioned.Interface`) exposes these sub-clients:
- `ComputeV1alpha1()` — machines, machine classes, machine pools
- `StorageV1alpha1()` — volumes, volume classes
- `NetworkingV1alpha1()` — networks, network interfaces, virtual IPs, load balancers
- `IpamV1alpha1()` — prefixes
- `CoreV1alpha1()` — resource quotas

## Files to create / modify

| Action | File |
|--------|------|
| Create | `internal/client/ironcore.go` |
| Modify | `cmd/server/main.go` |
| Modify | `internal/server/server.go` |

## Step-by-step

### Step 1 — Write `internal/client/ironcore.go`

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

### Step 2 — Update `cmd/server/main.go`

Replace the file content entirely:

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

### Step 3 — Update `internal/server/server.go`

Change `New()` to accept the clientset. Replace the file:

```go
package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	versioned "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
)

type Server struct {
	router   *chi.Mux
	ironcore versioned.Interface
}

func New(cs versioned.Interface) *Server {
	s := &Server{ironcore: cs}
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

	s.router = r
	return s
}

func (s *Server) Router() *chi.Mux { return s.router }

func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.router)
}
```

### Step 4 — Verify it compiles

```bash
make build
# Should produce bin/ironcore-dashboard with no errors
```

If ironcore-in-a-box is running (`cd ../ironcore-in-a-box && make deploy`), also test:

```bash
go run ./cmd/server --kubeconfig ~/.kube/config &
sleep 2
curl http://localhost:8080/healthz
```

Expected: `ok`

### Step 5 — Commit

```bash
git add internal/client/ internal/server/server.go cmd/server/main.go
git commit -m "feat: wire IronCore client-go into server"
```

## Done criteria

- `make build` succeeds with no errors
- `internal/client/ironcore.go` exists and compiles
- `server.New()` accepts `versioned.Interface`

## Next tasks (can run in parallel after this)

- Task 03: Machines API endpoints
- Task 04: Volumes, Networking, IPAM endpoints