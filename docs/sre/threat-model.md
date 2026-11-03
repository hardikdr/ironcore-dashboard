# SRE Console threat model

The SRE Console is a read-only aggregation boundary. It reads configured Kubernetes APIs, normalizes a narrow allow-list of operational fields into immutable snapshots, and exposes only GET endpoints. It must never become a credential browser, a cluster mutation proxy, or an unbounded topology query service.

| Threat | Control | Verification |
|---|---|---|
| Credentials, Secrets, ignition, or kubeconfigs leak through fixtures, normalization, logs, or API responses | Collectors never list Secret resources; normalizers allow-list safe fields; nested response scrubbing removes credential-like keys and values; structured logs contain metadata and counts only | Fixture scanner, normalizer leak tests, API redaction tests, structured-log test |
| SSRF through user-controlled cluster endpoints | Endpoints come only from operator-owned kubeconfig files loaded at startup; the API accepts no URL, kubeconfig, or endpoint parameter | Config/registry tests; no request DTO contains endpoint fields |
| Kubeconfig path injection or source impersonation | Kubeconfig references must be absolute, clean, regular files; impersonation fields are rejected; source IDs are validated and unique | Config and registry tests |
| Oversized objects or responses exhaust memory | Collectors use pagination and deadlines; snapshots retain sanitized projections; responses are capped at 2 MiB; list and graph limits are bounded | Collector and API cap tests |
| Label cardinality explodes metrics | Metric labels are limited to configured source IDs, normalized routes, status codes, and known kinds; names, namespaces, UIDs, labels, and payloads are excluded | Observability unit test |
| Graph explosion or adversarial selectors | Traversal depth is capped at 5 and nodes at 500; selectors are evaluated server-side; frontend renders only returned evidence and warns when truncated | Graph/selector tests and `EvidenceGraph.spec.ts` |
| Stale or partial data is mistaken for health | Last good snapshot is retained with explicit access/freshness metadata; partial warnings remain visible; unknown is never converted to healthy | Snapshot, source-state, Landscape, and Triage tests |
| One slow source blocks all reads | Collectors run concurrently with request timeouts; immutable snapshots remain readable; failure does not erase last success | Scheduler isolation/failure tests and race suite |
| Source identity is spoofed or joins cross tenants | Canonical keys contain source/group/kind/namespace/name/UID; resolvers require evidence; duplicate IDs and impersonation are rejected | Key, registry, and resolver tests |
| Mutation is introduced accidentally | SRE routes register GET only; upstream transport rejects non-GET/HEAD; SRE UI has no mutation action | Route/transport/UI audits |

`/metrics` is intended for the dashboard's trusted network boundary and contains no per-resource labels. Deployment ingress owns authentication and authorization. Live validation requires separate explicit approval and read-only identities.
