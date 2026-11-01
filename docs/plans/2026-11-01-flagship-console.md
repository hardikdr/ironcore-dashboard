# IronCore Flagship Console — Execution Plan

> For agentic workers: execute this plan task-by-task. Steps use `- [ ]` for tracking.

**Goal:** Transform the IronCore dashboard from a functional prototype into a premium, operator-first cloud console.

**Architecture:** Vue 3 + Vuetify 3 + Pinia SPA, Go backend with chi router. Demo mode uses fixture data; kubeconfig mode uses real k8s clients. Work in worktree `production-console` only.

**Tech Stack:** Vue 3 (Composition API), TypeScript, Vuetify 3, Pinia, Vite, Go 1.21, chi router, IronCore client-go

---

## Milestone 1: Design System + Shell (Phase 1)

### Task 1.1: Design tokens in Vuetify theme
**Files:** `frontend/src/plugins/vuetify.ts`

- [ ] Update dark/light theme colors to spec tokens (surface-raised, border, text-secondary)
- [ ] Add `global` defaults: VCard border, VTable density compact, VBtn rounded
- [ ] Verify light mode has sufficient contrast (AA)
- [ ] Build: `cd frontend && npm run build`

### Task 1.2: StatusBadge component upgrade
**Files:** `frontend/src/components/common/StatusBadge.vue`

- [ ] Add icon to every state: ● Running (green), ◌ Pending (amber), ✕ Error (red), ⊘ Terminating (purple), — Unknown (slate)
- [ ] Use `prepend-icon` chip pattern: icon + text, never color alone
- [ ] Add `size` prop (small/default)
- [ ] Update all usages still passing color-only strings

### Task 1.3: Top bar improvements
**Files:** `frontend/src/layouts/DashboardLayout.vue`

- [ ] Replace amber "DEMO" chip with icon+text chip (warning icon)
- [ ] Add "LIVE" green chip in real mode (check `/api/v1/mode`)
- [ ] Shrink top bar height to 48px
- [ ] Ensure namespace picker shows current namespace name, not just "NS"
- [ ] Keyboard shortcut hint: show "Ctrl+K" near search area

### Task 1.4: Sidebar badge counts
**Files:** `frontend/src/components/nav/NavItem.vue`, `frontend/src/layouts/DashboardLayout.vue`

- [ ] Add optional `badge` prop to NavItem
- [ ] Wire machines pending+error count badge on "Machines" nav item
- [ ] Wire volumes error count badge on "Volumes" nav item
- [ ] Count from Pinia stores (only non-zero counts shown)

### Task 1.5: Global search upgrade
**Files:** `frontend/src/components/common/GlobalSearch.vue`

- [ ] Add example hint text under search input: 'Try "web-" or type a resource name'
- [ ] Group results by type with type icon prefix
- [ ] Show keyboard hint (↑↓ navigate, Enter to go)
- [ ] Store last 5 searches in localStorage, show as "Recent" section

---

## Milestone 2: Console Fix (Phase 5 — do early, it's a correctness blocker)

### Task 2.1: Fix fake console in real mode
**Files:** `internal/server/server.go`, `frontend/src/views/MachineDetailView.vue`, `frontend/src/components/terminal/MachineTerminal.vue`

- [ ] In `server.go`: in WebSocket console handler, if not demo mode, return HTTP 501 with JSON `{"error":"console_unavailable","detail":"Console requires IronCore API server exec subresource proxy. See docs/product/production-console-blueprint.md#console."}`
- [ ] In `MachineTerminal.vue`: handle 501 response — show `v-alert` with explanation instead of terminal emulator
- [ ] In `MachineDetailView.vue`: show "Console (demo only)" label on Terminal tab in demo mode; show "Console (unavailable)" in real mode
- [ ] Add backend test: real-mode console returns 501
- [ ] Build and verify: `go test ./internal/...`

---

## Milestone 3: Operator Overview (Phase 2)

### Task 3.1: Needs Attention component
**Files:** `frontend/src/components/common/NeedsAttention.vue` (new), `frontend/src/views/OverviewView.vue`

- [ ] Create `NeedsAttention.vue`: takes `items: AttentionItem[]` prop, renders alert-style list
- [ ] AttentionItem: `{ label, icon, color, link, reason }`
- [ ] Only renders if items.length > 0 (no empty panel)
- [ ] Wire in OverviewView: pending+error machines, error volumes, error networks
- [ ] Add reason text from conditions where available
- [ ] Skeleton: don't render until stores loaded

### Task 3.2: Stat tiles with breakdown bars
**Files:** `frontend/src/components/common/StatCard.vue`

- [ ] Add optional `breakdown` prop: `{ label, value, color }[]`
- [ ] Render micro-bar below count showing state distribution
- [ ] Loading: show skeleton number, not zero
- [ ] Update OverviewView to pass breakdown for machines (running/pending/error)

### Task 3.3: Pool capacity section
**Files:** `frontend/src/views/OverviewView.vue`

- [ ] Add "Capacity" section: machine pool available/total capacity
- [ ] Resource quota bars: used/hard per resource type
- [ ] Use `v-progress-linear` with color tiers
- [ ] Only show if pools/quotas data available; hide section if empty

### Task 3.4: Recent events on overview
**Files:** `frontend/src/views/OverviewView.vue`

- [ ] Add "Recent Events" panel: last 10 Warning events from `/api/v1/namespaces/:ns/events`
- [ ] Each row: type icon + resource link + reason + message (truncated 80 chars) + age
- [ ] "View all →" link to `/operations`
- [ ] Empty state: "No recent warnings" (green check)
- [ ] Only fetch when namespace is active

---

## Milestone 4: Operations View (Phase 5)

### Task 4.1: Operations/Events view
**Files:** `frontend/src/views/OperationsView.vue` (new), `frontend/src/router/index.ts`

- [ ] Create `OperationsView.vue`: full-page events browser
- [ ] Table: Type icon | Reason | Resource (link) | Message | Count | First | Last
- [ ] Filter chips: Warning / Normal / by resource kind (Machine, Volume, Network)
- [ ] Text search on reason + message
- [ ] Sort by lastTime (default desc)
- [ ] Refresh age in header with manual refresh button
- [ ] Register route `/operations` in router
- [ ] Add "Operations" nav item under PLATFORM group

### Task 4.2: Backend events endpoint improvements
**Files:** `internal/api/events.go`

- [ ] Verify events endpoint supports `kind` filter query param
- [ ] Add `type` filter param (Warning/Normal)
- [ ] Return events sorted by lastTime desc
- [ ] Demo fixtures: add 3-5 Warning events to demo handler

---

## Milestone 5: Resource Detail Workspaces (Phase 3)

### Task 5.1: Shared ResourceHeader component
**Files:** `frontend/src/components/common/ResourceHeader.vue` (new)

- [ ] Props: `title`, `icon`, `state`, `namespace`, `createdAt`, `class?`, `pool?`, `ids?: CopyChip[]`, `actions: Action[]`
- [ ] Renders: breadcrumb | name + StatusBadge | age | namespace chip | class chip | IDs
- [ ] Actions: primary button + overflow v-menu for secondary actions
- [ ] Replace PageHeader usage in Machine/Volume/Network detail views

### Task 5.2: Machine detail workspace upgrade
**Files:** `frontend/src/views/MachineDetailView.vue`

- [ ] Use ResourceHeader
- [ ] Overview tab: two-column card grid (Machine Info left, Volume Attachments + Annotations right)
- [ ] Relationships tab (new): NICs table with network links + volumes table with class info
- [ ] Configuration tab (new): image, class, power policy fields
- [ ] Consolidate existing tabs: Overview, Configuration, Relationships, Conditions, Events, Console, YAML
- [ ] Remove dead zone: no empty card bodies

### Task 5.3: Volume detail workspace upgrade
**Files:** `frontend/src/views/VolumeDetailView.vue`

- [ ] Use ResourceHeader
- [ ] Convert to tabbed layout: Overview | Relationships | Conditions | Events | YAML
- [ ] Relationships tab: attached machine link, snapshot refs
- [ ] Overview: two cards (Details, Status)

### Task 5.4: Network detail workspace upgrade
**Files:** `frontend/src/views/NetworkDetailView.vue`

- [ ] Use ResourceHeader
- [ ] Tabs: Overview | Relationships | Conditions | Events | YAML
- [ ] Relationships: NICs table (name, machine link, IPs, state)
- [ ] Overview: network info + peerings table

---

## Milestone 6: List View Upgrades (Phase 4)

### Task 6.1: List header pattern
**Files:** `frontend/src/composables/useListView.ts` (new)

- [ ] Composable providing: `lastFetch`, `refresh()`, `loading`, `error`, `searchQuery`, `activeFilters`
- [ ] Persists sort column + direction to localStorage per route
- [ ] Apply to MachinesView, VolumesView, NetworksView

### Task 6.2: Machines list upgrade
**Files:** `frontend/src/views/MachinesView.vue`

- [ ] Add "Last updated: Xs ago ↻" header with refresh button
- [ ] Selection: checkbox column, floating bulk action bar when items selected
- [ ] Bulk actions: Power On/Off (confirmation), Delete (typed confirmation)
- [ ] Empty state when filtered: "No machines match filters" + Clear filters button
- [ ] Error state with retry button

### Task 6.3: Column chooser
**Files:** `frontend/src/components/common/ColumnChooser.vue` (new)

- [ ] Dropdown panel with toggle for optional columns
- [ ] Persists to localStorage per resource type
- [ ] Wire into MachinesView (optional columns: IPs, annotations)

---

## Milestone 7: Auth Honesty (Phase 5)

### Task 7.1: Document auth boundaries
**Files:** `docs/product/production-console-blueprint.md`

- [ ] Update security section: clarify bearer token is received but not forwarded to k8s clients yet
- [ ] Document kubeconfig mode vs future token-forwarded mode
- [ ] Remove any claim that per-user SSAR is enforced (it runs as server SA)

### Task 7.2: Frontend auth state communication
**Files:** `frontend/src/views/SettingsView.vue`, `frontend/src/layouts/DashboardLayout.vue`

- [ ] Settings: show clear note "Token stored for future use — current mode uses server kubeconfig"
- [ ] Remove any UI text implying the token gates real-mode access

---

## Milestone 8: Quality and Documentation (Phase 6)

### Task 8.1: Fix documentation
**Files:** `docs/product/production-console-blueprint.md`, `docs/demo/presentation-walkthrough.md`

- [ ] Update test count to actual number
- [ ] Remove reference to nonexistent `hack/demo-resources.yaml`
- [ ] Correct console capability section
- [ ] Update Phase 5 completion status

### Task 8.2: Smoke test updates
**Files:** `frontend/e2e/smoke.spec.ts`

- [ ] Add test: Operations view loads
- [ ] Add test: Console tab in demo mode shows simulation label
- [ ] Add test: Needs Attention panel shows/hides correctly
- [ ] Update screenshot baselines after visual changes

### Task 8.3: Final verification
- [ ] `go test ./internal/...` — all pass
- [ ] `go build ./...` — clean
- [ ] `cd frontend && npm run build` — no errors, chunks <500KB warning addressed
- [ ] `cd frontend && npx vue-tsc --noEmit` — no errors
- [ ] `cd frontend && npm audit` — note any unfixed vulns
- [ ] Full Playwright suite passes
- [ ] Git status clean (no uncommitted source changes)
- [ ] Push branch

---

## Review Gates

After each milestone:
1. `npm run build` passes
2. `go test ./...` passes
3. No TypeScript errors
4. Commit with meaningful message

After Milestone 8:
- Inspect screenshots manually
- Verify demo mode shows correct labels everywhere
- Push to `origin/worktree-production-console`
