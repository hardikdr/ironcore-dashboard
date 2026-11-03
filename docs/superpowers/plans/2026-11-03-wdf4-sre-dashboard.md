# WDF4 SRE Dashboard Implementation Plan

> **Purpose:** This document is the execution handoff for an AI agent implementing a read-only, multi-cluster SRE mode inside the existing IronCore Dashboard. Follow tasks in order. Do not contact or mutate live clusters while implementing unless the human explicitly authorizes a read-only validation session.

**Goal:** Add a first-class SRE Console to the existing IronCore Dashboard so an on-duty engineer can search any known VM, Shoot, physical node, pool, network, volume, or alert and traverse the proven dependency path across Gardener, central IronCore APIs, zonal CAPI clusters, provider components, and physical infrastructure.

**Product boundary:** Keep the current end-user console and all existing routes/CRUD behavior intact. Add an explicit `User Console / SRE Console` mode switch within the same application shell, theme, authentication boundary, and binary. The SRE Console is strictly read-only and uses a separate backend package, API prefix, data model, router, stores, and navigation.

**Canonical architecture contract:** `/Users/i331134/go/src/github.tools.sap/ironcore-dev/opsie/landscape-tooling/docs/ai-dashboard-handoff.md`

**Human architecture guide:** `/Users/i331134/go/src/github.tools.sap/ironcore-dev/opsie/landscape-tooling/docs/landscape-architecture.md`

**Safety contract:** `/Users/i331134/go/src/github.tools.sap/ironcore-dev/opsie/AGENTS.md` and `landscape-tooling/AGENTS.md`. The `opsie` repository and every reachable cluster are read-only. Never modify them as part of this plan.

---

## 1. Non-negotiable requirements

1. **Zero mutation capability in SRE mode.** No POST, PUT, PATCH, DELETE, terminal, exec, attach, copy, port-forward, restart, scale, cordon, drain, reconciliation trigger, or action menu. Enforce this in code and tests, not only in UI copy.
2. **No unrestricted `kubectl` from web requests.** Use typed/dynamic Kubernetes clients configured with read-only credentials. Local SSH/SOCKS tunnel lifecycle remains an operator or sidecar concern.
3. **Never collect or return Secret payloads, kubeconfigs, bearer tokens, client certificates, ignition data, or `.status.access.secretRef` contents.** Secret reference names may be retained only as redacted metadata when operationally necessary.
4. **Every record carries provenance:** source cluster, environment, API group, kind, namespace, name, UID, collected-at time, freshness, access result, and evidence class.
5. **Never merge similarly named resources without evidence.** Join priority: UID-bearing reference, ownerReference UID, explicit object reference, exact selector evaluation, provider ID, exact label, derived name, unknown.
6. **Qualify all Machine kinds in UI and code:** CAPI Machine, IronCore VM, or MCM Machine. Never display a bare “Machine” when ambiguity exists.
7. **Unknown is not healthy.** RBAC denial, unreachable sources, unsupported APIs, incomplete joins, and stale snapshots must remain explicit.
8. **Declared is not observed.** GitOps inventory may explain intended topology, but cannot prove current health.
9. **Snapshot counts are fixtures, not constants.** The 2026-11-03 counts may be used in tests only.
10. **Existing end-user routes remain compatible.** `/overview`, `/machines`, `/volumes`, `/networks`, and all current actions must continue to work against the existing single-cluster API.

---

## 2. Current product assessment

The current dashboard is a Go/Vue monolith:

- `cmd/server/main.go` creates one IronCore client, one Kubernetes clientset, and one dynamic client from a single kubeconfig or in-cluster config.
- `internal/server/server.go` exposes mutable end-user routes below `/api/v1` and serves the embedded Vue SPA.
- `internal/api` maps directly between HTTP calls and one IronCore API cluster.
- `frontend/src/layouts/DashboardLayout.vue` owns one namespace-oriented shell and global search over loaded end-user stores.
- `frontend/src/router/index.ts` contains one flat set of end-user routes.

This architecture is appropriate for the user console but cannot truthfully represent WDF4 because WDF4 data lives in multiple independent API servers. Do not stretch the existing stores or handlers with source flags. Introduce a separate SRE bounded context.

---

## 3. Target architecture

```text
                              one IronCore Dashboard binary
                                           |
                 +-------------------------+-------------------------+
                 |                                                   |
       Existing User Console                                New SRE Console
       /api/v1 + existing routes                            /api/sre/v1 + /sre/*
       single IronCore kubeconfig                           read-only multi-source
       CRUD preserved                                      no mutation routes
                                                                     |
                                  +----------------------------------+
                                  | Source registry + collectors      |
                                  | normalized snapshot store         |
                                  | relationship/index engine         |
                                  +----------------------------------+
                                    |          |          |          |
                           zonal CAPI A-G   Canary IC   Dev IC    Gardeners
                           read-only REST   read-only   read-only  read-only/RBAC
```

### Runtime components

1. **Source registry** loads declarative source metadata and prebuilt `rest.Config` objects. It never discovers credentials from Secrets.
2. **Collectors** poll list/get endpoints on a bounded interval with jitter, per-source timeout, concurrency limits, and cancellation. Watches are deferred until polling is proven safe.
3. **Snapshot store** atomically publishes immutable per-source snapshots. A failed collection does not erase the last successful snapshot; it changes access/freshness state.
4. **Normalizer** converts unstructured/typed objects into stable SRE records while preserving source references.
5. **Relationship engine** builds directional, evidence-bearing graph edges and reverse indexes.
6. **Observation engine** produces facts such as stale source, degraded provider, cross-zone storage placement, or unresolved selector. It does not guess policy violations.
7. **Search index** indexes exact identifiers and operator-friendly aliases across all normalized records.
8. **SRE HTTP API** exposes only GET endpoints under `/api/sre/v1`.
9. **SRE frontend** uses the existing shell/tokens but has separate routes, stores, views, and read-only vocabulary.

### Initial persistence decision

Use an in-memory atomic snapshot store for the first production slice. It minimizes operational dependencies and ensures collectors remain stateless/read-only. Define a storage interface so a later SQLite/PostgreSQL history implementation can be added without changing collectors or API contracts. Historical incident timelines are out of scope until current-state correctness is proven.

---

## 4. Source catalog and access strategy

### Zonal CAPI sources

| Source ID | Role | Expected current composition (fixture only) | Access |
|---|---|---:|---|
| `wdf4a` | Canary zone | 3 management, 16 compute, 4 storage | read-only zonal kubeconfig via SOCKS |
| `wdf4b` | Canary zone | 3 management, 16 compute, 4 storage | same |
| `wdf4c` | Canary zone | 3 management, 16 compute, 4 storage | same |
| `wdf4d` | Canary management-only | 3 management | same |
| `wdf4e` | Canary management-only | 3 management | same |
| `wdf4f` | Dev zone | 3 management, 2 compute, 4 storage | same |
| `wdf4g` | Dev zone | 3 management, 2 compute, 4 storage | same |

The operator tool `landscape-tooling/bin/zonal-capi` retrieves local kubeconfigs and creates SOCKS5 tunnels. The dashboard must **not** run this shell script from HTTP requests. For local/operator deployment, a supervisor starts tunnels before the dashboard and supplies the seven generated kubeconfig paths. For production, prefer network reachability plus least-privilege service-account kubeconfigs mounted read-only.

Each zonal kubeconfig already contains `proxy-url=socks5://localhost:<port>` when produced by `zonal-capi`. `client-go` must consume that config directly. Do not duplicate router IPs or SSH keys in application configuration.

Collect from each zonal cluster:

- core `nodes`, `pods`, `events`;
- `machines.cluster.x-k8s.io`;
- provider Deployments/DaemonSets/Pods in explicitly allowlisted namespaces;
- `machines.machine.sapcloud.io` only where the API exists and RBAC permits;
- `servers`/metal resources only where their fully qualified API identity is configured.

### Central IronCore sources

| Source ID | Environment | Kubernetes context | Zones |
|---|---|---|---|
| `canary-ironcore` | Canary | `garden-fe-canary--wdf4-external` | A, B, C |
| `dev-ironcore` | Dev | `garden-fe-dev--wdf4-external` | F, G |

Collect fully qualified resources:

- `machines.compute.ironcore.dev`, MachinePools, MachineClasses;
- Volumes, VolumePools, VolumeClasses, VolumeSnapshots;
- Networks, NetworkInterfaces, LoadBalancers, LoadBalancerRoutings, NATGateways, VirtualIPs, NetworkPolicies;
- IronCore control-plane Pods and Events from allowlisted namespaces.

### Gardener sources

Model Gardener installation and role separately:

- `sap-canary-hosting-gardener`: hosts the Kubernetes clusters containing Canary and Dev IronCore control planes.
- `ironcore-consumer-gardener`: creates VM-backed Shoot workers via `gardener-extension-provider-ironcore`; complete installation identity is not yet proven. The `garden-onmetal` context is a candidate, not a fact.
- `wdf4-bare-metal-gardener`: runs in the Canary API Kubernetes cluster, Virtual Garden `garden.wdf4.canary.ironcore.dev`, Seed `wdf4a`, with `wdf4ai` using `ironcore-metal`.

If Shoot listing is RBAC denied, publish source access as `denied`, retain the partial topology from IronCore exact cluster labels, and display `unknown/RBAC denied` rather than zero Shoots.

---

## 5. Normalized model

Create these backend types under `internal/sre/model`:

```go
type ResourceKey string // <source>/<api-group>/<kind>/<namespace-or-_>/<name>/<uid-or-_>

type Provenance struct {
    SourceID, Environment, ClusterRole string
    Group, Version, Kind, Namespace, Name, UID string
    CollectedAt time.Time
    Freshness Freshness
    Access AccessState
}

type EvidenceKind string // uid-reference, owner-reference, object-reference,
                         // selector-match, provider-id, exact-label,
                         // derived-name, unknown

type Relation struct {
    From, To ResourceKey
    Type string
    Evidence EvidenceKind
    FieldPath string
    Confidence string // proven, corroborated, derived, unknown
}
```

Required records:

- `SourceCluster`, `Zone`, `PhysicalNode`, `CapiMachine`;
- `MachinePool`, `IronMachine`, `MachineClass`;
- `GardenerInstallation`, `Shoot`, `WorkerGroup`, `McmMachineDeployment`, `McmMachineSet`, `McmMachine`;
- `Network`, `Nic`, `LoadBalancer`, `LoadBalancerRouting`, `NATGateway`, `VirtualIP`;
- `VolumePool`, `Volume`, `Attachment`;
- `ProviderComponent`, `KubernetesEvent`;
- `Observation`.

Keep raw source objects outside normalized records. If raw inspection is required, retain a sanitized JSON representation produced by an allowlist/redaction pipeline. Never forward arbitrary Kubernetes JSON directly to browsers.

---

## 6. Authoritative relationship rules

Implement each join as a named, independently tested resolver. Every produced edge includes evidence and field path.

1. **IronCore VM → MachinePool:** `.spec.machinePoolRef.name` within the same central IronCore source.
2. **MachinePool → Zone:** `topology.ironcore.dev/zone` label. Name parsing is fallback-only and marked derived.
3. **MachinePool → physical node:** `spec.providerID == libvirt-provider://<node-name>` joined to the exact zonal Node name for the pool's authoritative zone.
4. **VM → Shoot:** exact labels `extension.ironcore.dev/cluster-name` or `kubernetes.io/cluster`. Preserve canonical `shoot--project--name`; display tuple parsing is derived.
5. **VM → NIC:** status reference or NIC `.spec.machineRef.{name,uid}`; prefer UID and emit a warning if UID/name disagree.
6. **NIC → Network:** `.spec.networkRef.name` in the same namespace/source.
7. **LoadBalancer → NIC:** evaluate `.spec.networkInterfaceSelector` exactly against NIC labels in the same namespace. Then traverse NIC → VM. Never draw an LB-to-VM edge before selector evaluation.
8. **LoadBalancer → LoadBalancerRouting:** ownerReference UID.
9. **VM → Volume:** persistent `.spec.volumes[].volumeRef` and ephemeral `.status.volumes[].volumeRef`; never assume generated names.
10. **Volume → VolumePool/zone:** `.spec.volumePoolRef`, topology label, or explicit selector evidence.
11. **Gardener MCM hierarchy:** MachineDeployment selector/owner → MachineSet; MCM Machine owner UID → MachineSet; node label → Shoot Node.
12. **Bare-metal split:** any MCM provider ID beginning `ironcore-metal://` follows the physical Server path and must not be represented as an IronCore VM.

Build reverse indexes for every relation so the UI can traverse Shoot → VM → host and host → VM → Shoot with identical evidence.

---

## 7. Operator workflows that define usefulness

### “My VM X is not working”

The global SRE search must accept exact name, UID, IP, provider ID, canonical Shoot cluster label, physical node, pool, namespace/name, and user-facing Shoot name. An exact VM result opens a troubleshooting workspace that answers, in order:

1. Is the source fresh and accessible?
2. What does the IronCore VM's authoritative state/conditions say?
3. Which MachinePool, zone, and physical node host it?
4. Are the zonal provider Pod and MachinePool heartbeat healthy?
5. Which Shoot/worker group owns it, and is that join proven or derived?
6. Which NICs, IPs, Networks, LBs, NAT/Public IPs are attached or selecting it?
7. Which Volumes are requested/attached and where are their VolumePools?
8. Are there recent warning Events at any layer?
9. Which data is missing or RBAC denied?

The page should produce a concise “failure localization” summary, not an invented root cause. Example: `VM Running; NIC Available; provider healthy; Shoot relation proven; consumer Gardener health unknown due RBAC`.

### “A Shoot is degraded”

Search Shoot or canonical cluster label, then show worker groups, MCM hierarchy, expected/ready counts, each VM's zone/host, network frontends, volumes, and gaps. Group workers by worker group and zone.

### “Zone B has a problem”

Open Zone B and see management nodes, compute/storage nodes, provider components, pool capacity, hosted VMs, dependent Shoots, current warning Events, and dependencies entering/leaving the zone.

### “Which workloads depend on this physical server?”

Open the physical node and traverse MachinePool → hosted VMs → Shoot labels, NICs, Volumes, and LBs. Counts must be clickable and resolve to the exact set.

### “Is this an outage or a visibility gap?”

Every screen shows source freshness and access. Stale, unreachable, denied, unsupported, and not configured use distinct states.

---

## 8. SRE HTTP API

All endpoints are GET-only under `/api/sre/v1`. Register the router with only `r.Get`; add a route-audit test that fails if any other method is introduced.

```text
GET /api/sre/v1/status
GET /api/sre/v1/sources
GET /api/sre/v1/summary?environment=&zone=
GET /api/sre/v1/search?q=&types=&environment=&zone=&state=&limit=
GET /api/sre/v1/zones
GET /api/sre/v1/zones/{zone}
GET /api/sre/v1/nodes/{resourceKey}
GET /api/sre/v1/vms
GET /api/sre/v1/vms/{resourceKey}
GET /api/sre/v1/shoots
GET /api/sre/v1/shoots/{resourceKey}
GET /api/sre/v1/networks/{resourceKey}
GET /api/sre/v1/loadbalancers/{resourceKey}
GET /api/sre/v1/volumes/{resourceKey}
GET /api/sre/v1/observations
GET /api/sre/v1/graph?root={resourceKey}&depth=3
GET /api/sre/v1/resources/{resourceKey}/sanitized
```

API envelopes include:

```json
{
  "data": {},
  "meta": {
    "generatedAt": "...",
    "sources": [{"id": "...", "collectedAt": "...", "freshness": "fresh", "access": "ok"}],
    "partial": false,
    "warnings": []
  }
}
```

Use opaque URL-safe resource IDs derived from `ResourceKey` rather than accepting raw cluster/namespace/name path fragments. Search results include matched field and exact-vs-fuzzy rank.

---

## 9. Frontend information architecture

### Shared shell integration

Add a persistent console mode switch near product identity:

- `User Console` retains namespace selector and current navigation.
- `SRE Console` replaces namespace context with environment/source freshness and shows read-only status.
- Mode is URL-derived (`/sre/*`) and persisted only as a convenience; deep links remain authoritative.

Use the same semantic tokens, toolbar dimensions, focus behavior, responsive navigation, and component language. Do not create a second visual framework.

### SRE routes

```text
/sre/landscape
/sre/triage
/sre/zones
/sre/zones/:zone
/sre/nodes/:id
/sre/shoots
/sre/shoots/:id
/sre/vms
/sre/vms/:id
/sre/dependencies/:id
/sre/networking
/sre/loadbalancers/:id
/sre/storage
/sre/volumes/:id
/sre/sources
```

### Required views

1. **Landscape:** Canary/Dev control planes, A-G, role counts, pools, VM state, affected Shoots, alerts, freshness. D/E render management-only.
2. **Triage:** active observations, affected layer, proven impact, uncertainty, source freshness, fleet health table.
3. **Zone detail:** CAPI cluster, physical nodes by role, pools, hosted VMs, dependent Shoots, providers, Events, inbound/outbound dependencies.
4. **Physical node:** Node + CAPI Machine, conditions, provider Pods, pool capacity, hosted VMs, dependent Shoots, NICs, Volumes, observations.
5. **Shoot detail:** installation, project/name, health, worker groups, MCM hierarchy, VMs by zone, physical placement, networking/storage, access gaps.
6. **VM troubleshooting workspace:** Overview, Compute path, Networking, Storage, Ownership, Events/source.
7. **Dependency explorer:** directional graph whose edges expose evidence; never turn it into a decorative topology canvas.
8. **Networking:** public IP → LB → selector → NIC → VM → pool/zone and VM/NIC → Network → NAT → egress IP.
9. **Storage:** VolumePools by zone, Volumes, attachments, compute-vs-storage placement observations.
10. **Sources:** access, last success, duration, object counts, errors, freshness thresholds—without endpoints or credentials.

### Search behavior

Replace the SRE-mode command dialog with server-backed search after two characters, debounced and cancelable. Group exact results first: VMs, Shoots, physical nodes, pools, IPs, then other resources. Keyboard selection must work. Search never requires all collections to be preloaded in the browser.

---

## 10. Health, freshness, and observations

Define independently:

- `fresh`: successful collection within source threshold;
- `stale`: last success exists but exceeds threshold;
- `unreachable`: transport/authentication failure with no usable current snapshot;
- `denied`: RBAC denied for requested inventory;
- `unsupported`: API not served;
- `not-configured`: topology intentionally lacks that role;
- `unknown`: no authoritative condition or insufficient evidence;
- `degraded`: an authoritative required component is failing but service remains partially available;
- `failed`: authoritative failure/all required replicas unavailable.

Do not collapse control-plane health, provider health, VM state, and Shoot health. A Running VM does not prove a healthy Shoot.

Initial observation rules:

- stale/unreachable/denied source;
- provider Deployment/DaemonSet unavailable or Pod warning state;
- MachinePool not ready or heartbeat stale when exposed;
- VM failed/pending beyond configurable duration;
- UID/name disagreement in a reference;
- LB selector matches zero NICs or unresolved label inventory;
- attachment requested but status reference absent beyond threshold;
- compute and persistent storage zones differ (informational until policy defines severity);
- GitOps-declared source missing from live catalog (discrepancy, not automatic outage).

Every observation includes reason, message, source references, evidence, first/last observed timestamps within the current process, severity, and confidence.

---

## 11. Configuration and deployment

Add `--sre-config` to the server. If absent, SRE mode is disabled and existing behavior is unchanged. Do not add a demo fallback to production startup.

Example configuration shape (no credentials):

```yaml
enabled: true
pollInterval: 30s
staleAfter: 2m
sources:
  - id: wdf4a
    environment: canary
    role: zonal-capi
    zone: wdf4a
    kubeconfig: /etc/ironcore-dashboard/sources/wdf4a.yaml
  - id: canary-ironcore
    environment: canary
    role: central-ironcore
    context: garden-fe-canary--wdf4-external
    kubeconfig: /etc/ironcore-dashboard/sources/central.yaml
```

Validate config at startup:

- unique source IDs;
- role-specific required fields;
- kubeconfig files have safe permissions where observable;
- no inline token/cert/key fields;
- allowed API inventory is fixed by source role, not arbitrary config;
- all SRE client configs use a dashboard-specific user agent, QPS/burst limits, timeout, and no impersonation.

Production deployment options:

1. **Preferred:** least-privilege service-account kubeconfigs mounted read-only, with network reachability to each API.
2. **Operator workstation:** `zonal-capi` starts local tunnels outside the dashboard; generated kubeconfigs are mounted/read directly.
3. **Future:** one read-only collector agent per network boundary publishing sanitized snapshots to an aggregator. Use only if direct central connectivity becomes operationally fragile.

Do not put SSH private keys, tunnel management, or Secret retrieval into the dashboard process.

---

## 12. Detailed implementation tasks

### Task 1 — Freeze contracts with recorded fixtures

**Create:**

- `internal/sre/testdata/{zonal,canary,dev,gardener}/...json`
- `internal/sre/model/types.go`
- `internal/sre/model/key.go`
- `internal/sre/model/key_test.go`
- `docs/sre/data-contract.md`

- [ ] Build sanitized fixtures from the dated handoff facts; manually review for Secret/token/endpoint leakage.
- [ ] Include RBAC-denied, unreachable, stale, D/E management-only, VM-backed Shoot, and `ironcore-metal` cases.
- [ ] Implement stable global key parsing/formatting and round-trip tests.
- [ ] Document evidence, freshness, access, and health enums.
- [ ] Add a fixture scanner test rejecting keys matching token, password, kubeconfig, Secret data, ignition, PEM, or bearer patterns.

**Acceptance:** fixtures are sufficient to develop offline; no cluster access is required for normal tests.

### Task 2 — Add SRE configuration and source registry

**Create/modify:**

- `internal/sre/config/config.go`, `config_test.go`
- `internal/sre/source/registry.go`, `registry_test.go`
- `cmd/server/main.go`

- [ ] Add `--sre-config`; absent means disabled.
- [ ] Parse source metadata and create `rest.Config` using client-go only.
- [ ] Apply timeouts, user agent, QPS/burst, and transport wrapping.
- [ ] Reject inline credentials and duplicate IDs.
- [ ] Return source-access diagnostics without leaking paths or endpoints.
- [ ] Keep existing single-cluster clients untouched.

**Acceptance:** starting without SRE config is byte-for-byte behavior-compatible at the HTTP contract level; invalid SRE config fails closed.

### Task 3 — Implement immutable snapshot store

**Create:**

- `internal/sre/snapshot/store.go`, `store_test.go`
- `internal/sre/snapshot/types.go`

- [ ] Support atomic publish/read per source and aggregate generation IDs.
- [ ] Preserve last success when a later poll fails.
- [ ] Track last attempt, last success, duration, error class, access, and object counts.
- [ ] Compute freshness at read time.
- [ ] Make concurrent reads race-free; run `go test -race ./internal/sre/...`.

### Task 4 — Implement zonal CAPI collectors

**Create:**

- `internal/sre/collector/collector.go`
- `internal/sre/collector/zonal.go`, `zonal_test.go`
- `internal/sre/collector/scheduler.go`, `scheduler_test.go`

- [ ] Poll core Nodes/Pods/Events, CAPI Machines, and configured provider namespaces.
- [ ] Discover optional APIs once per successful discovery cycle; unsupported is distinct from denied.
- [ ] Bound each list with timeout and pagination.
- [ ] Limit concurrent source polls and add jitter.
- [ ] Never invoke shell commands or tunnel lifecycle.
- [ ] Normalize physical roles from authoritative labels/config, retaining raw labels as evidence.

### Task 5 — Implement central IronCore collectors

**Create:**

- `internal/sre/collector/ironcore.go`, `ironcore_test.go`
- `internal/sre/normalize/ironcore.go`, `ironcore_test.go`

- [ ] List the compute, storage, network, and IPAM inventory from the canonical contract.
- [ ] Preserve namespace, UID, conditions, refs, selectors, provider IDs, topology labels, and timestamps.
- [ ] Sanitize Secret references and exclude data fields.
- [ ] Handle one API group failing without discarding successful groups; mark snapshot partial.

### Task 6 — Implement Gardener/MCM collectors

**Create:**

- `internal/sre/collector/gardener.go`, `gardener_test.go`
- `internal/sre/normalize/gardener.go`, `gardener_test.go`

- [ ] Model installation role separately from source cluster.
- [ ] Collect Shoots, Seeds, CloudProfiles, MCM deployments/sets/machines only when served and authorized.
- [ ] Publish access gaps instead of empty lists.
- [ ] Preserve `ironcore-metal://` as an explicit bare-metal provider type.
- [ ] Never dereference Shoot access Secrets.

### Task 7 — Build the relationship graph

**Create:**

- `internal/sre/graph/graph.go`, `index.go`
- one resolver file/test per rule: `vm_pool.go`, `pool_node.go`, `vm_shoot.go`, `vm_nic.go`, `lb_targets.go`, `volume_placement.go`, `mcm.go`

- [ ] Apply the twelve authoritative rules in section 6.
- [ ] Store outgoing and incoming edges.
- [ ] Detect conflicts and retain both evidence plus an observation.
- [ ] Cap graph traversal depth/node count in API handlers.
- [ ] Ensure no selector edge crosses namespace/source unintentionally.

**Acceptance:** all nine canonical scenarios can be constructed from fixtures, including reverse traversal.

### Task 8 — Implement search and troubleshooting summaries

**Create:**

- `internal/sre/search/index.go`, `rank.go`, tests
- `internal/sre/diagnosis/vm.go`, `shoot.go`, `zone.go`, tests

- [ ] Index names, UID, IPs, provider IDs, canonical Shoot label, display Shoot tuple, pool, node, namespace/name.
- [ ] Prefer exact matches, then prefix, then token matches; no opaque fuzzy result above an exact identifier.
- [ ] Build fact-only summaries with explicit unknowns.
- [ ] Return matched field and highlighted fragments safely.

### Task 9 — Expose GET-only SRE API

**Create/modify:**

- `internal/sre/api/router.go`, handlers, response DTOs, tests
- `internal/server/server.go`

- [ ] Implement section 8 endpoints.
- [ ] Add GET-only route audit and CORS/auth integration.
- [ ] Add pagination, filters, response size caps, graph caps, and request timeouts.
- [ ] Add ETags based on aggregate generation.
- [ ] Return 206/partial metadata only through explicit response metadata; do not silently omit failed sources.
- [ ] Test redaction and absence of mutation verbs.

### Task 10 — Integrate SRE shell and stores

**Create/modify:**

- `frontend/src/router/index.ts`
- `frontend/src/layouts/DashboardLayout.vue`
- `frontend/src/components/app/ConsoleModeSwitch.vue`
- `frontend/src/components/sre/SreNavigation.vue`
- `frontend/src/stores/sre/{status,search,landscape}.ts`
- `frontend/src/api/sre.ts`

- [ ] Add `/sre/*` lazy routes and mode-aware navigation.
- [ ] Replace namespace picker with environment/freshness context in SRE mode.
- [ ] Show a persistent read-only indicator.
- [ ] Keep theme, responsive behavior, and keyboard search.
- [ ] Ensure SRE stores never import existing CRUD stores or actions.
- [ ] Add router/component tests proving existing user routes remain unchanged.

### Task 11 — Implement Landscape, Sources, and Triage

**Create:**

- `frontend/src/views/sre/LandscapeView.vue`
- `TriageView.vue`, `SourcesView.vue`
- components under `frontend/src/components/sre/`
- focused Vitest suites

- [ ] Implement the approved Landscape Map and On-call Triage concepts.
- [ ] Make aggregate counts link to filtered inventories.
- [ ] Distinguish D/E not configured, RBAC denied, unreachable, stale, and healthy.
- [ ] Mark fixture/snapshot times visibly.
- [ ] Render useful unknown/partial states at mobile sizes.

### Task 12 — Implement VM search and troubleshooting workspace

**Create:**

- `frontend/src/views/sre/VmsView.vue`, `VmDetailView.vue`
- `frontend/src/components/sre/GlobalSreSearch.vue`
- `FailureLocalization.vue`, `EvidenceBadge.vue`, `SourceDrawer.vue`
- tests

- [ ] Implement the “VM X is not working” workflow in section 7.
- [ ] Provide Overview, Compute path, Networking, Storage, Ownership, Events/source tabs.
- [ ] Every relation opens an evidence explanation.
- [ ] Raw source drawer is sanitized and read-only.
- [ ] Search works without browser-preloading inventories.

### Task 13 — Implement Zone, node, and Shoot workflows

**Create:**

- `ZoneDetailView.vue`, `PhysicalNodeDetailView.vue`, `ShootsView.vue`, `ShootDetailView.vue`
- associated components/tests

- [ ] Support bidirectional Zone/node/Shoot/VM traversal.
- [ ] Preserve Gardener installation role and provider type.
- [ ] Render `wdf4ai` as bare metal with no fabricated VM path.
- [ ] Show partial topology when Gardener access is denied.

### Task 14 — Implement dependency, networking, and storage views

**Create:**

- `DependencyExplorerView.vue`, `NetworkingSreView.vue`, `StorageSreView.vue`
- graph layout/evidence components and tests

- [ ] Render directed edges only from graph API evidence.
- [ ] Resolve LB selectors server-side and show matched NIC/VM/zone counts.
- [ ] Show VolumePool placement and cross-zone observations.
- [ ] Provide an accessible table fallback for every graph.

### Task 15 — Security, scale, and failure hardening

- [ ] Threat-model credential/Secret leakage, SSRF, kubeconfig path injection, oversized Kubernetes objects, label cardinality, graph explosion, stale data, and source impersonation.
- [ ] Add structured logs with source ID and collection phase, never object payloads.
- [ ] Add Prometheus metrics for collection success/duration/object counts, snapshot age, partial responses, API latency, and graph/search size.
- [ ] Load-test with at least 10,000 VMs, 50,000 NICs, and 5,000 LBs in generated fixtures.
- [ ] Confirm one slow/unreachable source does not block API reads or other collectors.
- [ ] Run race detector, frontend tests/build, Go tests/build, and secret scanners.

### Task 16 — Read-only live validation and rollout

- [ ] Require explicit human approval for each live validation session.
- [ ] Use only list/get through configured read-only identities.
- [ ] Compare dashboard joins with a small, manually verified sample from each source.
- [ ] Validate all nine acceptance scenarios from the canonical handoff.
- [ ] Verify HTTP access logs contain no non-GET `/api/sre/v1` calls.
- [ ] Ship disabled by default, then fixture mode, then one environment, then all sources.
- [ ] Document rollback as disabling `--sre-config`; user console remains available.

---

## 13. Testing strategy

### Unit tests

- key generation/parsing;
- status/freshness/access classification;
- every relationship resolver and conflict case;
- selector semantics;
- redaction;
- search ranking;
- observation rules.

### Contract tests

- source-role discovery and RBAC denial;
- partial snapshot publication;
- GET-only router;
- DTO JSON schemas and pagination;
- existing `/api/v1` behavior unchanged.

### Frontend tests

- mode switching/deep links;
- VM exact search and keyboard navigation;
- freshness/access/unknown rendering;
- evidence drawer;
- D/E management-only state;
- bare-metal Shoot;
- responsive name/status/context/action retention;
- no mutation controls in SRE routes.

### E2E fixtures

Use recorded sanitized fixtures for deterministic browser tests. Include:

1. VM-backed Shoot spanning A/B/C;
2. LB selector resolving to NICs and VMs;
3. persistent Volume placement;
4. bare-metal `wdf4ai`;
5. D/E management-only;
6. Gardener RBAC denied;
7. stale Canary source;
8. provider degradation with no proven user impact;
9. exact VM search returning the full path.

---

## 14. Definition of done

The SRE Console is complete only when:

- an operator can search an exact VM/IP/UID/provider ID and reach a one-page fact-based troubleshooting workspace;
- Shoot → MCM → IronCore VM → MachinePool → physical node works when evidence exists;
- physical node → hosted VMs → dependent Shoots works in reverse;
- networking and storage paths preserve selector/reference semantics;
- all values disclose source and freshness;
- denied/stale/unreachable/not-configured/unknown remain distinct;
- `ironcore-metal` is never represented as an IronCore VM;
- no SRE endpoint or UI exposes a mutation capability;
- no Secret or access payload reaches logs, fixtures, APIs, or browsers;
- existing end-user console routes and actions continue to pass their tests;
- the feature can be disabled without affecting the user console.

---

## 15. Recommended delivery order

Deliver in reviewable milestones:

1. **Offline truth model:** Tasks 1–3.
2. **Read-only collection:** Tasks 4–6.
3. **Relationships and query API:** Tasks 7–9.
4. **One-product shell + Landscape/Triage:** Tasks 10–11.
5. **VM troubleshooting golden path:** Task 12.
6. **Zone/Shoot reverse paths:** Task 13.
7. **Network/storage depth:** Task 14.
8. **Hardening and controlled rollout:** Tasks 15–16.

Do not begin broad UI implementation before fixture-backed joins and partial-data semantics are passing. The dashboard's credibility depends more on truthful relationships and uncertainty than on visual completeness.
