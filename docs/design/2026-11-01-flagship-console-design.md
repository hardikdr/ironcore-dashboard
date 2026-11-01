# IronCore Dashboard — Flagship Console Design Spec

> Status: Active · 2026-11-01  
> Scope: Design system, shell, overview, resource workspaces, list UX, auth, console

---

## 1. Users and Jobs

### 1.1 Tenant Engineer (primary)
Provisions and operates IronCore resources in one or more namespaces. Daily jobs:
- Launch a machine of a given class, attach volumes and network
- Diagnose a stuck/pending/error resource: read conditions, events, relationships
- Check fleet state at a glance before/after a change
- Delete unused resources safely

### 1.2 Platform Operator
Manages pools, classes, quotas across the cluster. Needs:
- Cross-namespace fleet health at a glance
- Pool capacity and class availability
- Server (metal-operator) inventory and claim status

### 1.3 Read-Only Observer / Compliance Reviewer
Views resource state and events without mutations. Needs:
- Clear read-only indicators when lacking write permission
- Honest capability boundaries

---

## 2. Information Architecture

### 2.1 Navigation Groups
```
COMPUTE
  Machines           /machines
  Machine Pools      /machinepools
  Servers            /servers

STORAGE
  Volumes            /volumes
  Snapshots          /volumesnapshots
  Buckets            /buckets
  Classes & Pools    /volumeclasses + /volumepools + /bucketclasses + /bucketpools

NETWORKING
  Networks           /networks
  Interfaces         /networkinterfaces
  Load Balancers     /loadbalancers
  NAT Gateways       /natgateways
  Virtual IPs        /virtualips
  Policies           /networkpolicies

IPAM
  IP Prefixes        /prefixes
  Allocations        /prefixallocations

PLATFORM
  Quotas             /resourcequotas
  Operations         /operations  (new: events feed)
  Settings           /settings
```

### 2.2 URL Structure
All namespace-scoped resources: `/namespaces/:ns/<resource>` for deep linking.
Global resources (pools, classes, servers): `/<resource>` (cluster-scoped).
Detail pages: `/<resource>/:name` (namespace from store).

---

## 3. Visual Design System

### 3.1 Color Tokens
```
--ic-bg-base:        #070b14   (dark) / #f1f5f9   (light)
--ic-surface:        #0d1526   (dark) / #ffffff   (light)
--ic-surface-raised: #1a2540   (dark) / #f8fafc   (light)
--ic-border:         #1e2d4a   (dark) / #e2e8f0   (light)
--ic-text-primary:   #e2e8f0   (dark) / #0f172a   (light)
--ic-text-secondary: #94a3b8   (dark) / #64748b   (light)
--ic-text-muted:     #475569   (dark) / #94a3b8   (light)

Status (icon + text, never color alone):
--ic-status-running:    #22c55e  (green-500)
--ic-status-pending:    #f59e0b  (amber-500)
--ic-status-error:      #ef4444  (red-500)
--ic-status-unknown:    #64748b  (slate-500)
--ic-status-terminating:#a855f7  (purple-500)

Primary actions: #818cf8 (indigo-400 dark) / #6366f1 (indigo-500 light)
Accent: #2dd4bf (teal-400)
```

### 3.2 Typography Scale
```
display:  20px / 600  (page titles)
title:    16px / 600  (section headers, card titles)
body:     14px / 400  (default content)
caption:  12px / 400  (metadata, timestamps, labels)
mono:     13px / 400  (code, IDs, resource names)
```

### 3.3 Spacing
4px base unit. Cards: 16px internal padding. Section gaps: 16px. Page padding: 24px.

### 3.4 Surface Hierarchy
- Base (`--ic-bg-base`): page background
- Surface (`--ic-surface`): cards, panels
- Raised (`--ic-surface-raised`): input backgrounds, hover states, nested cards
- Border: 1px `--ic-border` on cards, tables, separators

### 3.5 Status Semantics
All status indicators must use icon + text label, not color alone.
- Running/Available/Bound: ● + "Running" (green)
- Pending/Provisioning: ◌ + "Pending" (amber, animated pulse)
- Error/Failed: ✕ + "Error" (red)
- Unknown/Terminating: — + "Unknown" / ⊘ + "Terminating"

StatusBadge component handles all cases. Never hardcode colors in views.

### 3.6 Elevation
No box-shadows on cards by default (flat). Use border only. Reserve shadows for overlays (dialogs, tooltips, dropdowns).

### 3.7 Motion
Reduced-motion: respect `prefers-reduced-motion`. Default transitions: 150ms ease-out. No decorative animations. Loading pulses should respect reduced-motion (static skeleton).

---

## 4. Shell Design

### 4.1 Top Bar (48px)
Left: hamburger | IronCore logo + "Console" wordmark | breadcrumb (route-level)
Center: Global command bar (Ctrl+K) — resource search + actions
Right: [Refresh all ↻] [DEMO|LIVE chip] [Namespace picker] [User/Settings]

Namespace picker: compact dropdown showing current namespace, count badge showing total namespaces available, "All namespaces" option for cluster-scoped views.

### 4.2 Sidebar
Width: 240px expanded, 56px rail. Groups collapsed/expanded state persisted to localStorage.
Active item: left border accent + background tint. Hover: background tint only.
Groups have group-level collapse toggle. Badge counts for attention items (errors/pending).

### 4.3 Demo/Live Indicator
Always visible chip in top bar:
- Demo: amber "DEMO" chip with warning icon — "Running on fixture data. Actions do not persist."
- Real: green "LIVE" chip — "Connected to <cluster>"
- Disconnected: red "OFFLINE" chip

### 4.4 Global Search (Ctrl+K)
- Instant filter across all cached resources (name, namespace, class)
- Recent searches persisted to localStorage (max 10)
- Group by resource type with icon
- Keyboard: arrows to navigate, Enter to go, Esc to close
- Example hint: "Try 'web-' or 'machine:pending'"

---

## 5. Operator Overview (Home)

### 5.1 Layout (1440px)
```
┌─────────────────────────────────────────────────────────────────┐
│ Needs Attention                                  [full-width row] │
│ ┌──────────────────────────────────────────────────────────────┐ │
│ │ Error/pending items with reason text and direct action links │ │
│ └──────────────────────────────────────────────────────────────┘ │
├──────────────┬────────────────────────┬────────────────────────── │
│ Fleet Stats  │ Capacity & Quotas      │ Recent Events             │
│ (2-col grid) │ (pool+quota bars)      │ (last 10 warning/errors)  │
├──────────────┴────────────────────────┴────────────────────────── │
│ Resource Inventory (compact table per type, cross-linked)         │
└─────────────────────────────────────────────────────────────────┘
```

### 5.2 Needs Attention
- Only shown when there are pending/error/warning items
- Each item: icon + resource name (link) + state badge + reason (from conditions/events)
- Max 8 items; "View all" link to filtered list
- Never shown for loading — wait for data before rendering

### 5.3 Fleet Stats
6 stat tiles: Machines total, Running, Volumes, Networks, VIPs, Load Balancers.
Each tile: large number, label, subtle state breakdown bar.
Loading: skeleton tile, not spinner.

### 5.4 Capacity & Quotas
- Machine pools: available/total capacity bars
- Resource quotas: used/hard bars, color by % used (green <70%, amber <90%, red ≥90%)
- Only shown if data available; skeleton if loading

### 5.5 Recent Events (Kubernetes Events)
- Last 10 Warning-type events across namespace
- Each: icon + resource link + reason + message truncated + age
- "View all operations" link to /operations

---

## 6. Resource Workspaces

### 6.1 Shared Identity Header
Every detail view starts with:
```
[← Back]  [Resource icon]  [Name]  [StatusBadge]  [Age]  [Actions...]
          [Namespace chip] [Class/Pool chip] [Key IDs as CopyChips]
```
Actions are capability-aware: shown if user likely has permission, disabled-with-tooltip if not, hidden only if completely inapplicable.

### 6.2 Tab Structure
```
Overview     — key facts, primary relationships (2-col card grid)
Configuration — spec fields (read-only or editable)
Relationships — graph/table of connected resources with links
Conditions    — StatusConditions component
Events        — ResourceEvents component
YAML          — ResourceYaml component (copy + download)
```
Not every resource needs all tabs — only render tabs that have content.

### 6.3 Two-Column Card Layout
Each column holds stacked cards. Cards:
- Consistent subtitle header (icon + label)
- Compact table of key→value rows
- Links where references exist (machineRef, networkRef, etc.)
- No empty cards; hide if no data

### 6.4 Machine Detail
Identity: name, state, power, class, IPs  
Overview: machine info + volume attachments + annotations  
Configuration: image, class, power policy  
Relationships: NICs→Networks, Volumes→VolumeClasses  
Conditions: VolumesReady, NetworkInterfacesReady, Ready  
Events: Kubernetes events for this machine  
Console: disabled in real mode with explanation; simulation in demo mode  
YAML: full resource serialization  

### 6.5 Volume Detail
Identity: name, state, class, size  
Overview: volume info + access info  
Relationships: attached machine (link), snapshot refs  
Conditions + Events + YAML

### 6.6 Network Detail  
Identity: name, state, providerID  
Overview: network info + peerings  
Relationships: NICs using this network (table with machine links)  
Conditions + Events + YAML

---

## 7. List Views

### 7.1 List Header
```
[Page Title] [count badge]        [Last updated: Xs ago ↻]  [Create]
[Search...] [Filter chips...] [Sort ▼]               [⚙ Columns]
```

### 7.2 Table
- Checkbox selection column (hidden when not selectable)
- Name column always first, always a link
- State column always uses StatusBadge
- Age column always last before actions
- Actions column: v-menu three-dot, never inline icon buttons
- Empty state: illustration + message + primary action
- Error state: error icon + message + retry button
- Loading: skeleton rows (3-5 rows), not full spinner

### 7.3 Selection & Bulk Actions
- Select-all checkbox in header
- Floating selection bar when items selected: count + bulk actions + deselect
- Bulk actions: power toggle (machines), delete (with confirmation)

### 7.4 Filter Chips
- Predefined state filters (Running, Pending, Error) with aria-pressed
- Text search (instant, client-side)
- Active filters shown as dismissible chips
- "Clear all" when any filter active

---

## 8. Operations View (New)

A dedicated `/operations` view showing Kubernetes Events across the namespace:
- Sortable by lastTime (default: newest first)
- Filter: Warning/Normal type, by resource kind, by reason
- Each row: type icon | reason | resource link | message | count | age
- Refresh age shown in header
- No fabricated audit data — only real Kubernetes Events

---

## 9. Auth and Security Model

### 9.1 Modes
**Demo mode** (`--demo` flag): fixture data, always-open SSAR, console simulation. Clearly labeled.

**Kubeconfig mode** (current default): server uses static kubeconfig SA identity. SSAR runs as server SA (may fail closed per namespace). Bearer token from frontend is accepted but not yet forwarded to k8s clients. This is the deployable mode for now.

**Token-forwarded mode** (future): request-scoped clients using user's bearer token. Fail closed on SSAR. Implement when OIDC integration is needed.

### 9.2 Current Honest Boundary
- Frontend stores bearer token in localStorage and sends in Authorization header
- Backend middleware reads token but current k8s/IronCore clients still use server SA
- SSAR uses server SA identity — may allow or deny based on SA permissions, not user
- This must be clearly documented; never claim per-user auth is enforced

### 9.3 Console Security
Real mode: WebSocket handler currently dials raw TCP port 22 — this is wrong. 
Fix: show "Console unavailable — requires IronCore API server exec subresource proxy" in real mode.
Demo mode: keep simulation, label it clearly as "DEMO SIMULATION".

---

## 10. Responsive Rules

| Breakpoint | Behavior |
|-----------|---------|
| 1440px+ | Full 3-column layouts, full sidebar |
| 1024px  | 2-column layouts, sidebar rail default |
| 768px   | 2→1 column cards, sidebar overlay |
| 390px   | Single column, mobile nav overlay, compact top bar |

---

## 11. Accessibility

- WCAG AA contrast on all text/icon combinations
- All interactive elements reachable by keyboard, logical tab order
- Focus ring always visible (no `outline: none`)
- Status communicated by icon + text, not color alone
- `aria-pressed` on filter toggles, `aria-label` on icon-only buttons
- `role="group"` on filter chip groups
- Skip-to-content link

---

## 12. Real / Demo / Read-Only Boundaries

| Feature | Demo | Kubeconfig mode | Token-forwarded (future) |
|---------|------|-----------------|--------------------------|
| Resource lists | Fixture data | Real k8s | Real k8s |
| Create/delete | Simulated | Real (SA identity) | Real (user identity) |
| Console | Simulated (labeled) | Disabled with message | Exec subresource proxy |
| SSAR namespace check | Always allowed | SA identity | User identity (fail closed) |
| Bearer token forwarding | N/A | Header received, not forwarded | Header forwarded to k8s |
| Events | Fixture | Real Kubernetes Events | Real Kubernetes Events |

---

## 13. Acceptance Criteria

- [ ] No dead controls (disabled-with-reason or removed)
- [ ] No fake auth, fake terminal, fabricated real-mode data
- [ ] Consistent detail/list/create patterns across all supported resources
- [ ] Clean worktree, logical commits, no uncommitted source/generated churn
- [ ] Desktop/tablet/mobile × light/dark all presentation-ready
- [ ] `go test ./...`, `go build`, `vue-tsc --noEmit`, `npm run build` all pass
- [ ] Full Playwright smoke suite passes
- [ ] Blueprint and walkthrough match code
