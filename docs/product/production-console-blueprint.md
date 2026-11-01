# IronCore Dashboard — Production Console Blueprint

> Status: Draft v1.0 · 2026-11-01  
> Scope: Covers Phase 1 audit through Phase 3 delivery. Updated as work progresses.

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

### 2.1 Compute (`compute.ironcore.dev/v1alpha1`)

| CRD | API Backed | List | Detail | Create | Delete | Power |
|-----|-----------|------|--------|--------|--------|-------|
| `Machine` | ✅ Yes | ✅ | ✅ | ✅ | ✅ | ✅ (PATCH power) |
| `MachineClass` | ✅ Yes | ✅ (read-only) | ❌ | ❌ | ❌ | — |
| `MachinePool` | ❌ Not wired | ❌ | ❌ | ❌ | ❌ | — |

**Machine states:** `Pending`, `Running`, `Terminating`  
**Power states:** `On`, `Off`

### 2.2 Storage (`storage.ironcore.dev/v1alpha1`)

| CRD | API Backed | List | Detail | Create | Delete |
|-----|-----------|------|--------|--------|--------|
| `Volume` | ✅ Yes | ✅ | ✅ | ✅ | ✅ |
| `VolumeClass` | ✅ Yes | ✅ (read-only) | ❌ | ❌ | ❌ |
| `VolumePool` | ❌ Not wired | ❌ | ❌ | ❌ | ❌ |
| `VolumeSnapshot` | ❌ Not wired | ❌ | ❌ | ❌ | ❌ |
| `Bucket` | ❌ Not wired | ❌ | ❌ | ❌ | ❌ |
| `BucketClass` | ❌ Not wired | ❌ | ❌ | ❌ | ❌ |
| `BucketPool` | ❌ Not wired | ❌ | ❌ | ❌ | ❌ |

**Volume states:** `Pending`, `Available`, `Error`

### 2.3 Networking (`networking.ironcore.dev/v1alpha1`)

| CRD | API Backed | List | Detail | Create | Delete |
|-----|-----------|------|--------|--------|--------|
| `Network` | ✅ Yes | ✅ | ✅ | ✅ | ✅ |
| `NetworkInterface` | ✅ Yes | ✅ (via Network) | ❌ | ❌ | ❌ |
| `VirtualIP` | ✅ Yes | ✅ | ✅ | ✅ | ✅ |
| `LoadBalancer` | ✅ Yes | ✅ | ✅ | ✅ | ✅ |
| `NATGateway` | ❌ Not wired | ❌ | ❌ | ❌ | ❌ |
| `NetworkPolicy` | ❌ Not wired | ❌ | ❌ | ❌ | ❌ |
| `LoadBalancerRouting` | ❌ Not wired | ❌ | ❌ | ❌ | ❌ |

### 2.4 IPAM (`ipam.ironcore.dev/v1alpha1`)

| CRD | API Backed | List | Detail | Create | Delete |
|-----|-----------|------|--------|--------|--------|
| `Prefix` | ✅ Yes | ✅ | ✅ | ✅ | ✅ |
| `PrefixAllocation` | ❌ Not wired | ❌ | ❌ | ❌ | ❌ |

### 2.5 Core (`core.ironcore.dev/v1alpha1`)

| CRD | API Backed | List | Detail | Create | Delete |
|-----|-----------|------|--------|--------|--------|
| `ResourceQuota` | ❌ Not wired | ❌ | ❌ | ❌ | ❌ |

### 2.6 Metal (`metal.ironcore.dev/v1alpha1` via metal-operator)

| CRD | API Backed | List | Detail | Create | Delete |
|-----|-----------|------|--------|--------|--------|
| `Server` | ✅ Yes (dynamic) | ✅ | ✅ | ❌ | ❌ |

---

## 3. Navigation Architecture

The navigation is grouped into verticals that match operator mental models, not API groups. Each vertical maps to a section of sidebar navigation.

```
IronCore Dashboard
├── Overview                    /overview        (dashboard with stat cards)
├── Compute
│   ├── Machines                /machines        ✅ implemented
│   ├── Machine Classes         /machineclasses  planned Phase 3
│   └── Machine Pools           /machinepools    planned Phase 3
├── Storage
│   ├── Volumes                 /volumes         ✅ implemented
│   ├── Volume Classes          /volumeclasses   planned Phase 3
│   ├── Volume Pools            /volumepools     planned Phase 3
│   └── Snapshots               /snapshots       planned Phase 3
├── Networking
│   ├── Networks                /networks        ✅ implemented
│   ├── Network Interfaces      /networkinterfaces planned Phase 3
│   ├── Virtual IPs             /virtualips      ✅ implemented
│   ├── Load Balancers          /loadbalancers   ✅ implemented
│   ├── NAT Gateways            /natgateways     planned Phase 3
│   └── Network Policies        /networkpolicies planned Phase 3
├── IPAM
│   └── IP Prefixes             /prefixes        ✅ implemented
├── Operations
│   └── Servers                 /servers         ✅ implemented (metal-operator)
└── Administration
    └── Settings                /settings        ✅ implemented
```

**Not in navigation** (hidden from Phase 1):
- `PrefixAllocation` — view-only sub-resource of Prefix, shown inline on Prefix detail
- `LoadBalancerRouting` — inline on LoadBalancer detail
- `BucketClass`, `BucketPool`, `Bucket` — pending API wiring (Phase 3+)
- `ResourceQuota` — Phase 3 in Administration section

---

## 4. Current State vs. Production-Grade Gap

### 4.1 What works today (as of `worktree-production-console`)

| Area | State |
|------|-------|
| App shell (top bar, sidebar, nav drawer) | ✅ Implemented, mobile-responsive |
| Page transitions & loading states | ✅ Polished |
| Namespace switcher | ✅ Basic (in sidebar) |
| CRUD for Machines, Volumes, Networks, VIPs, LBs, Prefixes | ✅ |
| Machine power toggle | ✅ |
| Machine WebSocket console (simulated fallback) | ✅ |
| Server list (metal-operator) | ✅ |
| Demo mode (no kubeconfig required) | ✅ |
| Smoke test suite (6 tests) | ✅ All pass |
| WCAG 2.2 AA for navigation & core actions | ✅ (aria-labels, keyboard nav, color contrast) |

### 4.2 Production-grade gaps (Phases 2–3)

| Gap | Impact | Priority |
|-----|--------|----------|
| Global search (across all resource types) | High — console unusable at scale without it | P1 |
| Grouped nav with collapse/expand per vertical | Medium — current flat nav won't scale to 15+ items | P1 |
| Pagination on all list views | High — lists will time-out and OOM with real clusters | P1 |
| Column sort + filter on list views | High | P1 |
| Bulk actions (delete, power-toggle for machines) | Medium | P2 |
| Empty states (no machines in namespace, etc.) | Medium — UX confusion on fresh deployments | P2 |
| Error states with retry actions | Medium | P2 |
| Breadcrumbs on detail views | Medium | P2 |
| Resource event log on detail views | High for debugging | P2 |
| Recents / Favorites | Low | P3 |
| Notification panel | Low | P3 |
| MachinePool capacity view | Low | P3 |
| VolumeSnapshot create/restore | Low | P3 |
| NATGateway, NetworkPolicy views | Low | P3 |
| ResourceQuota view | Low | P3 |
| Bucket / BucketClass / BucketPool | Low (backend not wired) | P3+ |

---

## 5. Security and Tenancy Model

### 5.1 Backend security constraints (current)
- The backend uses the kubeconfig provided at server startup — no per-request auth.
- All API calls use the server's service account identity — no user impersonation.
- CORS allows only `localhost:5173` and `localhost:8080` in production mode.
- Demo mode allows `*` origins (acceptable for CI/preview only).

### 5.2 Production requirements (Phase 2+)
- **Authentication**: Bearer token forwarding from frontend to backend, or OIDC integration.
- **Authorization**: Backend must not expose resources the caller cannot access in Kubernetes — use SubjectAccessReview or run calls with the user's token.
- **Namespace scoping**: All list/get calls must be scoped to the namespace the user selected; the namespace switcher must validate namespace access before switching.
- **Audit log**: Write a structured access log to stdout in JSON (already partially done via chi `middleware.Logger`).

### 5.3 Demo mode boundaries
- `--demo` flag uses hardcoded fixture data. It **must not** be used in production.
- No kubeconfig, no Kubernetes API calls, no real secrets.
- The smoke test suite (`make smoke-test`) always uses demo mode.

---

## 6. Hyperscaler Benchmark: IronCore vs. AWS EC2 Console

Based on reference screenshots (`../aws-example-screenshots/`):

| Feature | AWS EC2 Console | IronCore Current | Gap |
|---------|----------------|-----------------|-----|
| Resource list with inline search | ✅ | ✅ (basic) | Search is per-view, not global |
| Column sort + filter | ✅ | ❌ | Missing on all views |
| Pagination (server-side) | ✅ | ❌ | Lists load all items |
| Bulk select + actions | ✅ | ❌ | No bulk selection |
| Tag / label filtering | ✅ | ❌ | No label-based filter |
| Resource relationships (Machine → NIC → IP) | ✅ | Partial | Detail views show some links |
| Event history on detail | ✅ | ❌ | Not implemented |
| Instance connect (console) | ✅ | ✅ (simulated) | Real SSH passthrough pending infra |
| Monitoring / metrics | ✅ (CloudWatch) | ❌ | Out of scope (no metrics API) |

---

## 7. Phased Roadmap

### Phase 0 — Production baseline (✅ Complete)
- Demo mode server (`--demo`) with fixture data
- Fixed `mountFrontend` SPA fallback (fs.StatFS panic)
- 6 Playwright smoke tests, all passing
- `make smoke-test` CI target

### Phase 1 — Blueprint (✅ This document)
- IronCore CRD audit
- Capability matrix
- Navigation architecture
- Persona mapping
- Gap analysis

### Phase 2 — Console shell foundation
Acceptance criteria:
- [ ] Sidebar groups (Compute, Storage, Networking, IPAM, Operations, Administration)
- [ ] Sidebar collapse/expand per group, persisted to localStorage
- [ ] Global search bar (keyboard shortcut `/` or `Ctrl+K`), scoped to visible resources
- [ ] Pagination component (page size 20/50/100) on Machines, Volumes, Networks list views
- [ ] Column sort (by name, state, created) on list views
- [ ] Filter chips on list views (by state, by namespace if cross-ns)
- [ ] Breadcrumbs on all detail views
- [ ] Empty state component (no resources, no search results, no access)
- [ ] Error state component (API error, with retry button)
- [ ] Bulk select on Machine list with bulk delete + bulk power-toggle
- [ ] WCAG 2.2 AA: all new interactive elements keyboard-accessible and screen-reader labeled

### Phase 3 — Real vertical slices
Acceptance criteria:
- [ ] Compute: MachineClass list view (read-only), MachinePool list view (read-only, shows available capacity)
- [ ] Storage: VolumeClass list view; Volume list shows VolumePool column; VolumeSnapshot list (read-only)
- [ ] Networking: NetworkInterface list/detail; NATGateway list/detail
- [ ] IPAM: PrefixAllocation sub-resource shown inline on Prefix detail
- [ ] Administration: ResourceQuota view per namespace
- [ ] Backend API: wired endpoints for MachinePool, VolumePool, VolumeSnapshot, NATGateway, NetworkInterface (standalone), PrefixAllocation, ResourceQuota

---

## 8. Real vs. Demo Mode Feature Boundaries

Any feature that relies on live Kubernetes data is "Real Only". The smoke test suite always runs against demo mode. A feature that works only in real mode must be hidden in demo mode UI, not shown as broken.

| Feature | Demo Mode | Real Mode |
|---------|-----------|-----------|
| Fixture data (5 machines, 5 vols, etc.) | ✅ | ❌ |
| Namespace switcher values | 4 hardcoded | From k8s CoreV1().Namespaces() |
| Machine console (WebSocket) | Simulated shell | SSH passthrough (if NIC IP reachable) |
| Volume snapshots | Not shown | Shown when API wired |
| Events / audit log | Not shown | From k8s Events API |
| ResourceQuota | Not shown | From core.ironcore.dev API |

---

_Blueprint authored by Claude Code (Phase 1 automation). Update this file as capabilities land._
