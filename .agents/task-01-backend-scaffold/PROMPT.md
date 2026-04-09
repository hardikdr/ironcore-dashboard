# Task 01 — Go Backend Scaffold

## Your job

Scaffold the Go HTTP backend for the IronCore Dashboard. By the end of this task:
- `ironcore-dashboard/` is a valid Go module
- `go run ./cmd/server` starts a server that responds `ok` to `GET /healthz`
- All dependencies compile cleanly

## Working directory

Always `cd` into `ironcore-dashboard/` before running commands. The workspace layout is:

```
dashboard-workspace/
├── ironcore/               ← IronCore APIs (READ ONLY — do not modify)
├── ironcore-in-a-box/      ← local Kind cluster (READ ONLY)
└── ironcore-dashboard/     ← ALL your work goes here
```

## Files to create

| File | Purpose |
|------|---------|
| `go.mod` | Go module definition |
| `go.sum` | Dependency checksums (generated) |
| `Makefile` | Build + run targets |
| `cmd/server/main.go` | Entry point |
| `internal/server/server.go` | chi router + CORS + /healthz |

## Step-by-step

### Step 1 — Initialise the Go module

```bash
cd ironcore-dashboard
go mod init github.com/ironcore-dev/ironcore-dashboard
```

Expected: `go.mod` created with `module github.com/ironcore-dev/ironcore-dashboard`

### Step 2 — Add dependencies

```bash
go get github.com/go-chi/chi/v5@latest
go get github.com/go-chi/cors@latest
go get k8s.io/client-go@latest
go get k8s.io/apimachinery@latest
go mod edit -replace github.com/ironcore-dev/ironcore=../ironcore
go get github.com/ironcore-dev/ironcore@latest
go mod tidy
```

### Step 3 — Write `internal/server/server.go`

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

### Step 4 — Write `cmd/server/main.go`

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

### Step 5 — Write `Makefile`

```makefile
.PHONY: build run dev tidy test

build:
	go build -o bin/ironcore-dashboard ./cmd/server

run:
	go run ./cmd/server --addr :8080

tidy:
	go mod tidy

test:
	go test ./...
```

### Step 6 — Verify

```bash
make run &
sleep 2
curl http://localhost:8080/healthz
```

Expected output: `ok`

If you see `ok`, kill the background process and continue.

### Step 7 — Commit

```bash
git add go.mod go.sum Makefile cmd/ internal/server/
git commit -m "feat: scaffold Go backend with chi router and /healthz"
```

## Done criteria

- `curl http://localhost:8080/healthz` returns `ok`
- `make build` produces `bin/ironcore-dashboard` with no errors
- All files listed above exist

## Next task

After this task completes, Task 02 (IronCore client-go wrapper) can start. Hand off by ensuring `go.mod` has the `replace` directive for `../ironcore`.