# SRE data contract

The SRE console is a read-only, multi-source bounded context. Its normalized records never contain credentials, Secret payloads, kubeconfigs, ignition, bearer tokens, certificates, or private keys. Raw Kubernetes objects are not part of this contract; a separately tested allowlist/redaction pipeline may expose a sanitized representation.

## Identity and provenance

Every record uses the key `<source>/<api-group>/<kind>/<namespace-or-_>/<name>/<uid-or-_>`. Cluster-scoped objects and unavailable UIDs use `_`. A bare `Machine` kind is invalid: records must say `CapiMachine`, `IronCoreVM`, or `McmMachine`.

Each record carries its source, environment, cluster role, API group/version, qualified kind, namespace, name, UID, collection time, freshness, access state, and evidence class.

## States

Freshness is independent of health:

- `fresh`: the last successful collection is within the source threshold.
- `stale`: a prior successful snapshot exists but exceeds the threshold.
- `unknown`: no successful collection time is available.

Access is explicit:

- `ok`: requested inventory was collected.
- `denied`: the API exists but RBAC rejected the request.
- `unreachable`: transport or authentication failed and no current data is available.
- `unsupported`: the API is not served by that source.
- `not-configured`: the topology intentionally lacks the role or capability.
- `unknown`: there is not enough evidence to classify access.

Health is never inferred from access:

- `healthy`: authoritative required conditions are satisfied.
- `degraded`: an authoritative required component is failing but service remains partially available.
- `failed`: an authoritative failure exists or all required replicas are unavailable.
- `unknown`: no authoritative condition or sufficient evidence exists.

## Evidence

Record evidence classes are `observed-live`, `declared-gitops`, `derived`, and `unknown`. Declared topology explains intent but does not prove health.

Relationship evidence is ordered: `uid-reference`, `owner-reference`, `object-reference`, `selector-match`, `provider-id`, `exact-label`, `derived-name`, then `unknown`. Every edge records the source field path and confidence (`proven`, `corroborated`, `derived`, or `unknown`). Similar names alone never produce a proven edge.

## Offline fixtures

Fixtures under `internal/sre/testdata` are sanitized, dated observations for deterministic tests. Counts are not production constants. They cover zones A–G, D/E management-only topology, stale and unreachable sources, RBAC denial, a VM-backed Shoot path, network selector data, storage placement, and an `ironcore-metal://` worker that must never be represented as an IronCore VM.
