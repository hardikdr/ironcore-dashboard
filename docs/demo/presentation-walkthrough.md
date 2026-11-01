# IronCore Dashboard — Team Demo Walkthrough

> Estimated time: 7–10 minutes  
> Mode: Demo server (`make demo` or `./bin/ironcore-dashboard --demo --addr :8181`)  
> URL: http://localhost:8181  
> Last verified: 2026-11-01 (Phase 4, commit on `worktree-production-console`)

---

## Before You Start

```bash
# Build and start the demo server
cd ironcore-dashboard
make demo     # builds Go binary + starts on :8181

# Alternatively:
go build -o bin/ironcore-dashboard ./cmd/server
./bin/ironcore-dashboard --demo --addr :8181
```

Open http://localhost:8181 in Chrome or Firefox at **1440×900** resolution.

> **Honest disclosure**: This demo runs against fixture data — no kubeconfig or cluster needed.
> All resource counts, states, and timestamps are hardcoded. Actions (create, delete, power toggle) 
> appear to succeed but do not persist. The real-mode backend speaks to actual IronCore CRDs via 
> client-go; the API and response shapes are identical between demo and real modes.

---

## Demo Script (7 minutes)

### 0:00 — Overview Dashboard (1 min)

Open **http://localhost:8181**. You land on the Overview page.

**Point out:**
- Six stat tiles across the top: Machines (5), Running (3), Volumes (5), Networks (2), Virtual IPs, Load Balancers — all from real IronCore API response shapes.
- Two charts (Machine States, Storage Overview) using live store data.
- "Recent Resources" table with cross-resource timeline — not just one type.
- "Quick Actions" panel — direct shortcuts to common operations.
- Cluster health indicator (top right) — polls `/healthz`.
- Namespace switcher — all views react immediately on change.

**IronCore differentiator:** The overview aggregates across `compute`, `storage`, and `networking` API groups simultaneously. AWS EC2 console is one service; IronCore is a unified API for bare-metal + cloud compute + S3-compatible storage.

---

### 1:00 — Navigation and Search (1 min)

**Show collapsible nav groups:**
- Click "COMPUTE" label — collapses the section, persisted to localStorage.
- Re-expand. Show Machines, Machine Pools, Topology.

**Press `Ctrl+K`** (or click the Search bar):
- Type "web" — instantly shows machine `web-frontend-01`.
- Type "prod" — shows networks, volumes, prefixes matching "prod".
- Arrow keys to navigate, Enter to jump.

**Show rail mode:**
- Click the hamburger icon (☰) — nav collapses to icon-only rail mode.
- Hover icons to see tooltips. Click hamburger again to restore.

**IronCore differentiator:** Search spans all resource types across the IronCore API group hierarchy in a single keystroke — no switching between Machines, Volumes, and Networks separately.

---

### 2:00 — Machines (1.5 min)

Navigate to **Machines**.

**Point out:**
- Running/Pending/Terminating state filter chips — click "Pending" to filter.
- Table: Name, State badge, Power, Class, IPs, Created.
- Click on a machine name (e.g., `web-frontend-01`) → detail view.

**On detail:**
- Overview tab: Class, Power, IPs, attached volumes.
- Console tab: WebSocket terminal — in demo mode this opens a simulated shell.
  - Type `ls`, `uptime` — pre-programmed responses.

**Show bulk actions:**
- Go back to list, select 2 machines via checkboxes.
- Click "Power Off" — confirmation dialog, cancel it.

**IronCore differentiator:** Machine console via WebSocket proxied through the dashboard backend to the machine's NIC IP — no SSH bastion required.

---

### 3:30 — Storage (1 min)

Navigate to **Volumes**.

- Available/Pending/Error filter chips.
- Click a volume — detail with `sizeGiB`, `volumeClass`, `accessDriver`, `volumeID`.

Navigate to **Snapshots**.

- Point-in-time snapshots backed by `VolumeSnapshot` CRD.
- Link from snapshot → source volume.
- Delete action with confirmation dialog.

---

### 4:30 — Networking (1 min)

Navigate to **Networks** → click `prod-network`.

- Detail: peerings, provider ID.
- "Network Interfaces" tab — all NICs attached to this network, each with machine ref.

Navigate to **Load Balancers** → detail:
- Type (Public/Internal), IP families, ports, network ref.

Navigate to **NAT Gateways**:
- Shows IPs allocated, network ref.

**IronCore differentiator:** Networking is a first-class API (`networking.ironcore.dev`), not a wrapper around cloud provider primitives. Load balancers, NICs, and NAT gateways are all Kubernetes CRDs with controller-managed lifecycle.

---

### 5:30 — Object Storage (30 sec)

Navigate to **Object Storage → Buckets** (if Phase 4 nav is present):
- Bucket class, pool assignment, state, endpoint.

Navigate to **Bucket Classes**:
- Capabilities (IOPS, throughput) from the `BucketClass` CRD.

**IronCore differentiator:** S3-compatible object storage provisioned via the same Kubernetes control plane as compute and network — no separate AWS/GCS credentials.

---

### 6:00 — Operations (1 min)

Navigate to **Servers** (metal-operator):
- Physical bare-metal servers: BMC address, power state, claim status.
- Detail view shows all fields from the `Server` CRD.

Navigate to **Resource Quotas**:
- Hard limits and current usage per resource type.
- Progress bars turn yellow >70%, red >90%.

**IronCore differentiator:** The metal-operator integration means the dashboard surfaces the physical layer — you can see which server hosts are available before they become MachinePool capacity.

---

### 7:00 — Responsive & Accessibility (optional 1 min)

In Chrome DevTools, switch to **iPhone 14** device emulation (390×844):
- Drawer is closed — no content obscured.
- Tap hamburger — drawer slides in as overlay, content dims.
- Tap any nav item — drawer closes automatically.
- All controls accessible: search collapses to magnifier icon, namespace to compact dropdown.

---

## Key IronCore Differentiators to Emphasize

| IronCore | Traditional Cloud Console |
|---------|--------------------------|
| All resource types via unified Kubernetes CRD API | Each service has its own console/API |
| Metal-operator bridges bare-metal ↔ cloud abstraction | No bare-metal visibility |
| Network policies as Kubernetes CRDs, not VPC rules | Provider-specific constructs |
| IPAM (`Prefix`, `PrefixAllocation`) as first-class resources | Manual or separate IPAM tool |
| S3-compatible storage via same control plane | Separate storage service |
| Open-source, self-hosted | Vendor-controlled |

---

## Honest Capability Disclosure

| Capability | Status |
|-----------|--------|
| All resource lists | ✅ Real API handlers in production mode |
| Machine CRUD + power toggle | ✅ Real |
| Machine console | ✅ Real (WebSocket via NIC IP) / Simulated in demo |
| Volume CRUD | ✅ Real |
| Network CRUD, VirtualIP, LoadBalancer | ✅ Real |
| NATGateway, VolumeSnapshot (list + delete) | ✅ Real |
| NetworkInterface, NetworkPolicy (read-only) | ✅ Real |
| Bucket, BucketClass, BucketPool | ✅ Real |
| ResourceQuota, PrefixAllocation | ✅ Real |
| Server (metal-operator) | ✅ Real (dynamic k8s client) |
| Kubernetes Events on detail pages | ❌ Not yet implemented (P1 gap) |
| Status conditions on detail pages | ❌ Not yet implemented |
| Authentication (OIDC/bearer token) | ❌ Not implemented — uses server SA |
| Tenant isolation (SAR) | ❌ Not implemented |

---

## Recommended Demo Data (real mode)

If showing against a real IronCore cluster:

```bash
# Create demo namespace
kubectl create namespace demo

# Create a MachineClass (if not present)
kubectl apply -f hack/demo-machineclass.yaml

# Pre-provision 2 machines, 3 volumes, 1 network
kubectl apply -f hack/demo-resources.yaml
```

Fixtures are in `hack/demo-resources.yaml` (TODO: create this file for Phase 5).

---

## Run Commands

```bash
# Demo mode (no cluster needed)
make demo
# → http://localhost:8181

# Smoke tests (requires demo to be stopped first)  
make smoke-test
# → runs 25 Playwright tests, captures screenshots to frontend/e2e/screenshots/

# Real mode (requires KUBECONFIG)
export KUBECONFIG=~/.kube/config
make run
# → http://localhost:8181

# Frontend type check only
cd frontend && npm run build
```
