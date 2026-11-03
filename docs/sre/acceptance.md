# Fixture-backed acceptance record

The deterministic acceptance suite is `go test -v ./internal/sre/acceptance`. It loads only sanitized JSON under `internal/sre/testdata`; it performs no network or cluster operation.

| # | Scenario | Automated assertion |
|---|---|---|
| 1 | VM-backed Shoot spanning the physical path | VM → MachinePool → physical node and VM → Shoot evidence exists |
| 2 | LB selector resolving NICs/VM context | server graph contains `selects-nic` with selector evidence |
| 3 | Persistent Volume placement | Volume → VolumePool object-reference exists |
| 4 | Bare-metal `wdf4ai` | no edge from the `ironcore-metal` MCM machine to an IronCore VM |
| 5 | D/E management-only | `wdf4d` and `wdf4e` remain `not-configured`, not failed or healthy compute |
| 6 | Gardener RBAC denied | consumer Gardener source remains `denied` |
| 7 | Stale Canary source | `wdf4c` retains stale last-good status and unreachable access |
| 8 | Provider degradation without proven user impact | diagnosis does not fabricate a provider-caused VM failure |
| 9 | Exact VM search and full path | exact UID wins and traversal reaches node and Shoot |

These checks validate join semantics, not a live environment. A live comparison requires a separately approved session and a small manually verified sample from every configured source.
