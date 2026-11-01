# IronCore Dashboard — Production Console Blueprint

> Status: Draft v2.0 · 2026-11-01  
> Scope: Covers Phase 0–4 delivery. Updated to reflect Phase 3 shipped, Phase 4 in progress.

---

## 1. Personas

### 1.1 Tenant Engineer
A developer or SRE who provisions and operates cloud resources (VMs, volumes, networks) inside one or more namespaces they own. Primary day-to-day user. Needs to: launch and resize machines, attach/detach volumes and NICs, inspect states, read logs, access consoles, and manage IP prefixes.

**Jobs to be done:**
- "Launch a machine of class `x1-medium` in namespace `production` with an attached volume."
- "Why is my machine stuck in `Pending`? What events are attached to it?"
- "What IPs are allocated from prefix `10.0.0.0/16`?"

### 1.2 Platform Operator
An infrastructure admin who manages machine pools, volume pools, and class definitions across the cluster. Needs namespace-wide or cluster-wide views and bulk operations.

**Jobs to be done:**
- "Show me all machines across all namespaces and their pool assignment."
- "Which MachinePool has capacity headroom?"
- "Delete 20 stale volumes in the `dev` namespace."

### 1.3 Security / Compliance Reviewer
Reads audit information, reviews resource quotas, and validates that tenants are within limits.

**Jobs to be done:**
- "What is the ResourceQuota for namespace `staging`?"
- "Which machines are running without an associated network interface?"

---

## 2. IronCore CRD Capability Matrix

Legend: ✅ Fully implemented (real API + frontend) · ⚠️ Demo-only (no real API handler) · 🔧 Partial (real API but limited UI) · ❌ Not implemented

### 2.1 Compute (`compute.ironcore.dev/v1alpha1`)

| CRD | Real API | List | Detail | Create | Delete | Notes |
|-----|---------|------|--------|--------|--------|-------|
| `Machine` | ✅ | ✅ | ✅ | ✅ | ✅ | Power toggle (On/Off), WebSocket console, bulk delete/power |
| `MachineClass` | ✅ | ✅ | ❌ | ❌ | ❌ | Capabilities from API; list used in Machine create wizard |
| `MachinePool` | ✅ | ✅ | ❌ | ❌ | ❌ | Phase 3: real handler; shows available classes |

**Machine states:** `Pending`, `Running`, `Terminating`  
**Power states:** `On`, `Off`

### 2.2 Storage (`storage.ironcore.dev/v1alpha1`)

| CRD | Real API | List | Detail | Create | Delete | Notes |
|-----|---------|------|--------|--------|--------|-------|
| `Volume` | ✅ | ✅ | ✅ | ✅ | ✅ | |
| `VolumeClass` | ✅ | ✅ | ❌ | ❌ | ❌ | Used in Volume create wizard |
| `VolumePool` | ✅ | ✅ | ❌ | ❌ | ❌ | Phase 3: real handler; shows available classes |
| `VolumeSnapshot` | ✅ | ✅ | ❌ | ❌ | ✅ | Phase 3: real handler; delete wired |
| `Bucket` | ✅ | ✅ | ❌ | ❌ | ✅ | Phase 4: real handler; S3-compatible object storage |
| `BucketClass` | ✅ | ✅ | ❌ | ❌ | ❌ | Phase 4: cluster-scoped, read-only |
| `BucketPool` | ✅ | ✅ | ❌ | ❌ | ❌ | Phase 4: cluster-scoped, read-only |

**Volume states:** `Pending`, `Available`, `Error`  
**Bucket states:** `Pending`, `Available`, `Error`

### 2.3 Networking (`networking.ironcore.dev/v1alpha1`)

| CRD | Real API | List | Detail | Create | Delete | Notes |
|-----|---------|------|--------|--------|--------|-------|
| `Network` | ✅ | ✅ | ✅ | ✅ | ✅ | |
| `NetworkInterface` | ✅ | ✅ | ❌ | ❌ | ❌ | Listed via Networks detail tab + standalone /networkinterfaces; owned by machines |
| `VirtualIP` | ✅ | ✅ | ✅ | ✅ | ✅ | |
| `LoadBalancer` | ✅ | ✅ | ✅ | ✅ | ✅ | |
| `NATGateway` | ✅ | ✅ | ❌ | ❌ | ✅ | Phase 3: real handler |
| `NetworkPolicy` | ✅ | ✅ | ❌ | ❌ | ❌ | Phase 4: real handler; read-only (infrastructure-managed) |
| `LoadBalancerRouting` | ❌ | ❌ | ❌ | ❌ | ❌ | Shown inline on LoadBalancer detail only |

### 2.4 IPAM (`ipam.ironcore.dev/v1alpha1`)

| CRD | Real API | List | Detail | Create | Delete | Notes |
|-----|---------|------|--------|--------|--------|-------|
| `Prefix` | ✅ | ✅ | ✅ | ✅ | ✅ | |
| `PrefixAllocation` | ✅ | ✅ | ❌ | ❌ | ❌ | Phase 3: real handler; namespace-scoped list |

### 2.5 Core (`core.ironcore.dev/v1alpha1`)

| CRD | Real API | List | Detail | Create | Delete | Notes |
|-----|---------|------|--------|--------|--------|-------|
| `ResourceQuota` | ✅ | ✅ | ❌ | ❌ | ❌ | Phase 3: real handler; shows hard/used with progress bars |

### 2.6 Metal (`metal.ironcore.dev/v1alpha1` via metal-operator)

| CRD | Real API | List | Detail | Create | Delete | Notes |
|-----|---------|------|--------|--------|--------|-------|
| `Server` | ✅ (dynamic) | ✅ | ✅ | ❌ | ❌ | BMC address, power state, claim status |

---

## 3. Navigation Architecture

```
IronCore Dashboard
├── Overview                    /overview
├── Compute
│   ├── Machines                /machines        ✅ Full CRUD + power + console
│   ├── Machine Pools           /machinepools    ✅ Read-only
│   └── Topology                /topology        ✅ Visual graph
├── Storage
│   ├── Volumes                 /volumes         ✅ Full CRUD
│   └── Snapshots               /volumesnapshots ✅ List + delete
├── Networking
│   ├── Networks                /networks        ✅ Full CRUD
│   ├── Virtual IPs             /virtualips      ✅ Full CRUD
│   ├── Load Balancers          /loadbalancers   ✅ Full CRUD
│   ├── NAT Gateways            /natgateways     ✅ List + delete
│   ├── Network Interfaces      /networkinterfaces ✅ Read-only list (Phase 4)
│   └── Network Policies        /networkpolicies ✅ Read-only list (Phase 4)
├── Object Storage
│   ├── Buckets                 /buckets         ✅ List + delete (Phase 4)
│   ├── Bucket Classes          /bucketclasses   ✅ Read-only (Phase 4)
│   └── Bucket Pools            /bucketpools     ✅ Read-only (Phase 4)
├── IPAM
│   └── IP Prefixes             /prefixes        ✅ Full CRUD
├── Operations
│   ├── Servers                 /servers         ✅ Read-only (metal-operator)
│   └── Resource Quotas         /resourcequotas  ✅ Read-only with usage meters
└── Settings                    /settings        ✅
```

**Sub-resources shown inline (not in nav):**
- `PrefixAllocation` — on Prefix detail page
- `LoadBalancerRouting` — on LoadBalancer detail
- `NetworkInterface` — tab on Network detail (MachineRef shown as link)

---

## 4. Responsive Breakpoints

| Breakpoint | Viewport | Behavior |
|-----------|---------|---------|
| Mobile (xs) | < 600px | Drawer closed by default, hamburger opens overlay; search collapses to icon; namespace to compact menu; refresh hidden |
| Tablet (sm/md) | 600–1280px | Drawer open, cluster status collapses to dot icon |
| Desktop (lg+) | > 1280px | Full bar, drawer with rail toggle |

Tested at 390px, 768px, and 1440px in both light and dark themes. All 20 Playwright tests pass.

---

## 5. Current State vs. Production-Grade Gap

### 5.1 What works today (Phase 4, branch `worktree-production-console`)

| Area | State |
|------|-------|
| App shell — responsive 390/768/1440, light + dark | ✅ |
| Drawer closes on mobile nav, overlay mode | ✅ |
| Global search (Ctrl+K, /) across all resource types | ✅ |
| Grouped nav with localStorage collapse/expand | ✅ |
| Pagination (20/50/100) on all list views | ✅ |
| State filter chips on Machines, Volumes | ✅ |
| Bulk delete + bulk power toggle on Machines | ✅ |
| Empty states and error states with retry | ✅ |
| CRUD: Machine, Volume, Network, VirtualIP, LoadBalancer, Prefix | ✅ |
| Machine power toggle (On/Off via PATCH) | ✅ |
| Machine WebSocket console (real + simulated fallback) | ✅ |
| MachinePool, VolumePool list (real API) | ✅ |
| VolumeSnapshot list + delete (real API) | ✅ |
| NATGateway list + delete (real API) | ✅ |
| NetworkInterface list (real API — both inline + standalone) | ✅ |
| NetworkPolicy list (real API, read-only) | ✅ |
| Bucket list + delete (real API) | ✅ |
| BucketClass + BucketPool list (real API, read-only) | ✅ |
| ResourceQuota with hard/used progress bars (real API) | ✅ |
| PrefixAllocation list (real API) | ✅ |
| Server list + detail (metal-operator, dynamic client) | ✅ |
| Demo mode — clearly labeled, all fixtures present | ✅ |
| Smoke tests: 20 tests, all passing | ✅ |
| Go build: clean, no warnings | ✅ |
| vue-tsc: clean, zero TS errors | ✅ |

### 5.2 Remaining production gaps

| Gap | Impact | Priority |
|-----|--------|----------|
| Bearer token forwarding / OIDC auth | Critical for production deployment | P1 |
| Per-request Kubernetes impersonation or SAR | Critical for multi-tenant security | P1 |
| Kubernetes Events endpoint on detail views | High — needed to debug Pending machines | P1 |
| Status conditions on detail pages | High — debugging aid | P1 |
| Resource relationships panel (Machine → NIC → IP) | High — navigable graph | P2 |
| Column sort on list views | Medium | P2 |
| VolumeSnapshot create from Volume | Medium | P2 |
| Raw YAML view on all detail pages | Medium | P2 |
| Breadcrumbs on detail views | Low | P3 |
| Notification / alert panel | Low | P3 |
| MachinePool capacity detail (per machine type) | Low | P3 |
| Recents / Favorites | Low | P3 |
| Server-side pagination (list with label selectors) | Medium — needed at 500+ resources | P2 |

---

## 6. Security and Tenancy Model

### 6.1 Backend security constraints (current)
- Backend uses the kubeconfig provided at startup — no per-request auth.
- All API calls use the server's service account identity — no user impersonation.
- Demo mode (`--demo`) explicitly short-circuits all Kubernetes calls.

### 6.2 Demo mode boundaries
- `--demo` flag uses hardcoded fixture data. **Must not be used in production.**
- No kubeconfig, no Kubernetes API calls, no real secrets.
- The smoke test suite (`make smoke-test`) always uses demo mode.
- Demo server banner displayed prominently in future iteration.

### 6.3 Production deployment requirements (not yet implemented)
1. **Authentication**: OIDC/Bearer token forwarding from frontend `Authorization: Bearer <token>` to backend, which then passes it to the Kubernetes API.
2. **Authorization**: Backend runs all Kubernetes calls with the user's token — never the service account token. This prevents cross-tenant data leakage.
3. **Namespace scoping**: Validate namespace access with `SubjectAccessReview` before switching.

---

## 7. Real vs. Demo Mode Feature Boundaries

| Feature | Demo Mode | Real Mode |
|---------|-----------|-----------|
| All resource lists | ✅ Fixture data | ✅ From real Kubernetes API |
| Namespace switcher values | 4 hardcoded | From k8s CoreV1().Namespaces() |
| Machine console (WebSocket) | Simulated shell (`$PS1` prompt) | SSH passthrough via NIC IP |
| Events / audit log | Not shown | From k8s Events API (P1 gap) |
| Status conditions | Not shown | From resource .Status.Conditions |
| Buckets (Phase 4) | Demo fixtures | From StorageV1alpha1().Buckets() |
| NetworkPolicies (Phase 4) | Demo fixtures | From NetworkingV1alpha1().NetworkPolicies() |

---

## 8. Phased Roadmap

### Phase 0 — Production baseline ✅
- Demo mode, SPA fallback fix, smoke tests, `make smoke-test`

### Phase 1 — Blueprint ✅
- CRD audit, capability matrix, navigation architecture, persona mapping, gap analysis

### Phase 2 — Console shell foundation ✅
- Grouped collapsible nav, global search (Ctrl+K), pagination, filter chips, bulk actions, EmptyState/ErrorState components

### Phase 3 — Real vertical slices ✅
- MachinePool, VolumePool, VolumeSnapshot, NATGateway, PrefixAllocation, ResourceQuota
- All six backed by real IronCore client-go handlers
- Demo fixtures for all six

### Phase 4 — Responsive + extended resources (in progress)
- ✅ Responsive layout: mobile drawer, compact top bar, no horizontal overflow
- ✅ Expanded smoke tests: 20 tests (vs 6 in Phase 3), responsive at 390/768/1440
- ✅ Phase 4 resources: NetworkInterface list, NetworkPolicy list, Bucket+BucketClass+BucketPool
- 🔄 Events endpoint: real Kubernetes Events per resource (next milestone)
- 🔄 Status conditions: display conditions from resource status on detail pages

### Phase 5 — Production hardening (planned)
- OIDC authentication, per-request bearer token forwarding
- SubjectAccessReview namespace validation
- Status conditions on all detail views
- Kubernetes Events panel on all detail views
- Server-side pagination with label selectors
- Raw YAML view via `kubectl.io/v1` or direct API

---

_Blueprint authored by Claude Code (Phase 1) and updated through Phase 4 automation. Update as capabilities land._
