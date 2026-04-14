# IronCore Dashboard

> **Proof of Concept** — A web-based dashboard for [IronCore](https://github.com/ironcore-dev/ironcore), providing a visual interface to manage bare-metal compute, networking, and storage resources.

## Architecture

```
┌─────────────────────────────────────────────┐
│                Single Binary                │
│                                             │
│  ┌──────────────┐   ┌─────────────────────┐ │
│  │  Vue 3 SPA   │   │    Go HTTP Server   │ │
│  │  (Vuetify)   │──▶│  (chi router)       │ │
│  │              │   │                     │ │
│  │  - Machines  │   │  REST API /api/v1/  │ │
│  │  - Volumes   │   │  Static file serve  │ │
│  │  - Networks  │   │                     │ │
│  │  - VIPs      │   └──────────┬──────────┘ │
│  │  - LBs       │              │            │
│  │  - IPAM      │   ┌──────────▼──────────┐ │
│  └──────────────┘   │  IronCore client-go │ │
│                     │  + k8s client-go    │ │
│                     └──────────┬──────────┘ │
└────────────────────────────────┼────────────┘
                                 │
                    ┌────────────▼────────────┐
                    │   Kubernetes / IronCore  │
                    │   API Server             │
                    └─────────────────────────┘
```

The frontend (Vue 3 + Vuetify) is compiled and **embedded into the Go binary** at build time via `go:embed`. The result is a single self-contained binary that serves both the REST API and the SPA with no external dependencies at runtime.

### Backend (`internal/`)

| Package | Responsibility |
|---------|----------------|
| `internal/client` | Builds the IronCore and standard k8s clientsets from a kubeconfig |
| `internal/api` | HTTP handlers for machines, volumes, networks, VIPs, load balancers, IPAM |
| `internal/server` | chi router wiring, CORS, static SPA fallback |

### Frontend (`frontend/`)

TypeScript, Vue 3, Vuetify 3, Pinia, Vue Router, Axios. Built with Vite, output embedded into the Go binary.

## Prerequisites

- Go 1.21+
- Node.js 18+
- Access to a Kubernetes cluster with IronCore installed

> **Note:** This module uses a `replace` directive in `go.mod` pointing to a local checkout of `../ironcore`. You need both repos checked out as siblings:
> ```
> workspace/
> ├── ironcore/
> └── ironcore-dashboard/
> ```
> Clone IronCore: `git clone https://github.com/ironcore-dev/ironcore ../ironcore`

## Testing with ironcore-in-a-box

[ironcore-in-a-box](https://github.com/ironcore-dev/ironcore-in-a-box) provides a local Kubernetes cluster with IronCore pre-installed — the fastest way to try the dashboard.

```bash
# 1. Start ironcore-in-a-box (follow its README)
#    This gives you a kubeconfig, typically at ~/.kube/config

# 2. Build the dashboard
make build

# 3. Run against your local cluster
./bin/ironcore-dashboard --kubeconfig ~/.kube/config --addr :8080

# 4. Open http://localhost:8080
```

### Development mode (hot-reload)

Run the backend and frontend separately for a faster feedback loop:

```bash
# Terminal 1 — Go backend (API on :8080)
make dev-backend

# Terminal 2 — Vite dev server (UI on :5173, proxies /api to :8080)
make dev-frontend
```

Then open http://localhost:5173.

## Build

```bash
# Build frontend + embed into Go binary
make build

# Binary output
./bin/ironcore-dashboard --help
```

## License

Apache 2.0
