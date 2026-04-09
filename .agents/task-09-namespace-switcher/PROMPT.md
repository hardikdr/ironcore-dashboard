# Task 09 — Go Backend Serves Built Frontend (Integration)

## Prerequisites

All of Tasks 01–08 must be complete:
- `make build` produces `bin/ironcore-dashboard`
- `cd frontend && npm run build` succeeds (Vue SPA compiles)
- All API endpoints return valid JSON

Verify:
```bash
make build
cd frontend && npm run build && cd ..
ls dist/frontend/index.html  # must exist
```

## Your job

1. Add the standard k8s clientset to the server so `GET /api/v1/namespaces` works
2. Embed the built Vue SPA into the Go binary using `embed`
3. Update the Makefile so `make build` also builds the frontend
4. Verify the single binary serves both API and frontend at http://localhost:8080

## Files to modify

| Action | File |
|--------|------|
| Modify | `internal/server/server.go` (add k8s client, embed frontend, namespace route) |
| Modify | `cmd/server/main.go` (create k8s client) |
| Modify | `Makefile` (add `build-frontend` target, wire into `build`) |
| Modify | `frontend/vite.config.ts` (ensure `outDir` is `../dist/frontend`) |

## Step-by-step

### Step 1 — Verify `vite.config.ts` output dir

Check `frontend/vite.config.ts` has:
```typescript
build: {
    outDir: '../dist/frontend'
}
```

If it says something else, fix it.

### Step 2 — Update `internal/server/server.go`

Replace the full file with this version that embeds the frontend and adds the k8s client:

```go
package server

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/ironcore-dev/ironcore-dashboard/internal/api"
	versioned "github.com/ironcore-dev/ironcore/client-go/ironcore/versioned"
	metav1    "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s       "k8s.io/client-go/kubernetes"
)

//go:embed ../../dist/frontend
var frontendFS embed.FS

type Server struct {
	router   *chi.Mux
	ironcore versioned.Interface
	k8s      *k8s.Clientset
}

func New(cs versioned.Interface, k8sClient *k8s.Clientset) *Server {
	s := &Server{ironcore: cs, k8s: k8sClient}
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173", "http://localhost:8080"},
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	}))

	// Health check
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Namespace list (for project/namespace switcher)
	r.Get("/api/v1/namespaces", func(w http.ResponseWriter, r *http.Request) {
		list, err := s.k8s.CoreV1().Namespaces().List(r.Context(), metav1.ListOptions{})
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		names := make([]string, 0, len(list.Items))
		for _, ns := range list.Items {
			names = append(names, ns.Name)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(names)
	})

	// Machine routes
	mh := api.NewMachineHandler(cs)
	r.Get("/api/v1/machineclasses", mh.ListMachineClasses)
	r.Route("/api/v1/namespaces/{ns}/machines", func(r chi.Router) {
		r.Get("/", mh.List)
		r.Post("/", mh.Create)
		r.Get("/{name}", mh.Get)
		r.Delete("/{name}", mh.Delete)
		r.Patch("/{name}/power", mh.PatchPower)
	})

	// Volume routes
	vh := api.NewVolumeHandler(cs)
	r.Route("/api/v1/namespaces/{ns}/volumes", func(r chi.Router) {
		r.Get("/", vh.List)
		r.Post("/", vh.Create)
		r.Delete("/{name}", vh.Delete)
	})

	// Networking routes
	nh  := api.NewNetworkHandler(cs)
	vip := api.NewVirtualIPHandler(cs)
	lb  := api.NewLoadBalancerHandler(cs)
	r.Route("/api/v1/namespaces/{ns}", func(r chi.Router) {
		r.Get("/networks",          nh.ListNetworks)
		r.Get("/networkinterfaces", nh.ListNetworkInterfaces)
		r.Get("/virtualips",        vip.List)
		r.Get("/loadbalancers",     lb.List)
	})

	// IPAM routes
	iph := api.NewIPAMHandler(cs)
	r.Get("/api/v1/namespaces/{ns}/prefixes", iph.ListPrefixes)

	// Serve built Vue SPA for all other routes
	sub, err := fs.Sub(frontendFS, "dist/frontend")
	if err != nil {
		panic("failed to sub frontendFS: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		// SPA fallback: serve index.html for unknown routes so vue-router handles them
		_, err := sub.(fs.StatFS).Stat(r.URL.Path[1:])
		if err != nil {
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})

	s.router = r
	return s
}

func (s *Server) Router() *chi.Mux { return s.router }

func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.router)
}
```

> **Important:** The `//go:embed ../../dist/frontend` directive requires the `dist/frontend/` directory to exist at compile time. Always run `make build-frontend` before `go build`.

### Step 3 — Update `cmd/server/main.go`

```go
package main

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
	kubeconfig := flag.String("kubeconfig", "", "path to kubeconfig (empty = in-cluster)")
	flag.Parse()

	// IronCore clientset
	cs, err := ironclient.New(*kubeconfig)
	if err != nil {
		log.Fatalf("ironcore client: %v", err)
	}

	// Standard k8s clientset (for namespace listing)
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

### Step 4 — Update `Makefile`

```makefile
.PHONY: build build-frontend run dev-backend dev-frontend tidy test

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

### Step 5 — Full integration test

```bash
make build
./bin/ironcore-dashboard --addr :8080 --kubeconfig ~/.kube/config
```

Open http://localhost:8080 — you should see the full dashboard served from the single Go binary.

Verify:
- http://localhost:8080 loads the Vue SPA (blue topbar, sidebar)
- http://localhost:8080/api/v1/namespaces returns `["default", ...]`
- http://localhost:8080/api/v1/namespaces/default/machines returns `[]`
- Navigating to `/machines`, `/volumes`, `/networks` etc. all work (vue-router SPA navigation)
- Refreshing the browser at `/machines` still works (SPA fallback to index.html)

### Step 6 — Commit

```bash
git add internal/server/server.go cmd/server/main.go Makefile
git commit -m "feat: embed Vue SPA in Go binary, k8s namespace listing, single binary serve"
```

## Done criteria

- `make build` compiles both frontend and backend with no errors
- `./bin/ironcore-dashboard --kubeconfig ~/.kube/config` serves the full dashboard at port 8080
- Browser refresh works on any SPA route (not just `/`)
- `GET /api/v1/namespaces` returns real namespace names from the cluster
- No `dist/` directory is needed at runtime — it's baked into the binary

## Troubleshooting

**`embed: pattern ../../dist/frontend: directory does not exist`**
Run `make build-frontend` first, then `go build`.

**SPA routes 404 on refresh**
The SPA fallback in the `/*` handler must serve `index.html` for unknown paths. Check the fallback logic in server.go.

**Namespace switcher shows empty list**
Check that `GET /api/v1/namespaces` returns data by running `curl http://localhost:8080/api/v1/namespaces`. If it errors, check that the k8s clientset is properly created in main.go.