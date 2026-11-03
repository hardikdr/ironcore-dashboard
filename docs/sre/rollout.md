# SRE Console rollout and rollback

## Local fixture mode

Fixture mode is cluster-free: it loads sanitized JSON snapshots, exposes the GET-only SRE API, metrics, and the embedded SPA, and does not construct Kubernetes clients.

```bash
go run ./cmd/server --addr 127.0.0.1:8181 --sre-fixture-dir internal/sre/testdata
# open http://127.0.0.1:8181/sre/landscape
```

`--sre-config` and `--sre-fixture-dir` are mutually exclusive. With neither flag, the SRE backend is disabled by default and the existing user console behavior is unchanged.

## Controlled rollout

1. Run fixture mode and the nine-scenario acceptance suite.
2. Obtain explicit human approval for one live validation session. Approval is per session; it is not persistent authorization.
3. Start with one environment and read-only identities. Compare a small sample from every configured source against direct list/get results.
4. Run `SRE_LIVE_VALIDATION_APPROVED=yes-for-this-session SRE_BASE_URL=... scripts/validate-sre-live.sh`. It issues GET requests only.
5. Audit the HTTP log with `scripts/audit-sre-access-log.sh <log>` and require zero non-GET SRE requests.
6. Observe collection success, duration, snapshot age, partial responses, API latency, graph size, and search size at `/metrics`.
7. Add remaining sources only after access/freshness gaps and sample joins are understood.

## Rollback

Remove `--sre-config` and restart the dashboard. This stops SRE collection and removes `/api/sre/v1` without changing the established `/api/v1` user console. No cluster cleanup is required because SRE collection never creates, patches, or deletes cluster objects.

If one source is unhealthy but the console remains useful, remove that source from the operator-owned configuration or disable the entire SRE feature; do not broaden credentials as an emergency workaround.
