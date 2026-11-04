# Local console development

## What happened

As inspected on 4 November 2026:

- Local `main` is `ac8a2d4`: the User Console plus an SRE implementation plan. It does not contain the SRE implementation. The last known remote main reference is `290c238`; no fetch was performed.
- `codex/wdf4-sre-dashboard` contains the initial SRE implementation through `a1cd4b0`. The executable `/tmp/wdf4-sre-dashboard-final` on port 8181 was built from that version. Go embeds the frontend at build time, so later source edits do not update that executable.
- Another agent changed seven frontend source files in the same worktree without committing them: the shell, palette, navigation, Landscape, Triage, Dependencies, and EvidenceGraph. It built and launched this version on port 8185 with `--sre-fixture-dir internal/sre/testdata` on 3 November. This is the preferred dark preview. It was not a separate application or separate backend. Port 8185 was no longer listening at the start of consolidation.
- The main checkout also had local generated index and dependency-lock changes. Those were preserved on a separate archive branch. `.pi/` remains untouched and untracked.

## Branches

| Local branch | Purpose |
| --- | --- |
| `main` | Unchanged original User Console and SRE plan |
| `archive/main-local-before-console-unification` | Existing local main build/index and dependency-lock state, saved before switching |
| `archive/sre-original-8181` | Original SRE implementation at `a1cd4b0`; retained, not selected as the visual baseline |
| `archive/sre-preferred-8185` | Preferred dark preview, captured in commit `511cf37` |
| `codex/unified-console` | Active combined version, developed in the main repository directory |

No branch was pushed or merged. Existing worktree folders are left intact for recovery, but are not used for this development workflow.

## Combined behavior

- `/overview` and all existing User resource routes use the original `DashboardLayout.vue` from the 8181 build. User views and visual structure are unchanged. A disposal guard prevents refresh timers from being installed after switching away during loading.
- The User header's SRE button navigates to `/sre/landscape` on the same server.
- `/sre/*` uses `SreDashboardLayout.vue` and the preferred dark preview pages.
- The dark console has a User/SRE switch for returning to `/overview`.
- The SRE palette is scoped to its own shell so it does not change the User Console's dark palette.
- Existing SRE collectors, API contracts, and detailed-resource routes remain: the preferred design depends on them. The archived version is not offered as a second SRE mode.

## Run the active branch locally

From the main repository directory:

```sh
cd /Users/i331134/go/src/github.com/ironcore-dev/dashboard-workspace/ironcore-dashboard
git branch --show-current
# Expected: codex/unified-console
./scripts/preview-local.sh
```

Open `http://127.0.0.1:8185/overview` for User or `http://127.0.0.1:8185/sre/landscape` for SRE. The switch stays on that same port. Stop with Ctrl-C. If that port is already occupied, choose another explicitly, for example `./scripts/preview-local.sh 8186`.

The script builds from the current checkout every time, then starts a loopback-only fixture server. It never reads a kubeconfig or starts collectors against a cluster. Installed frontend dependencies are required; it does not install packages automatically.

**Fixture limitation:** User pages and navigation are available, but the fixture server intentionally does not expose the User CRUD backend. User resource loading/creation is therefore unavailable in this preview; errors do not indicate a cluster problem. SRE shows sanitized local fixtures, not live infrastructure.

## Inspect actual main without creating another worktree

Stop the preview and check `git status --short` before switching. Commit or otherwise deliberately preserve your own work first. Do not discard changes to switch branches.

```sh
git switch main
# Inspect source in the same repository directory.
cd frontend
npm run build -- --emptyOutDir
npm run preview -- --host 127.0.0.1 --port 8186
```

This is a frontend-only preview of main; it has no connected backend and resource requests will fail. Stop it before returning with `git switch codex/unified-console`. Builds update tracked generated files in this repository; review/preserve generated changes before switching. The combined branch's script rebuilds the correct embedded frontend, avoiding stale binaries.

For a source-only comparison without switching or generating files:

```sh
git log --oneline main..codex/unified-console
git diff main..codex/unified-console -- frontend/src
git show main:frontend/src/router/index.ts
```

## What to move forward

Use the dark preview as the visual direction and preserve the User Console. Next work should validate the preferred SRE overview against API evidence: Landscape currently contains hard-coded zone capacities, and Triage has presentation controls without corresponding incident workflows. Some detail pages still inherit the original SRE presentation. Those are concrete follow-up tasks, not evidence that the preferred preview is a complete production console. Do not treat “Live”, capacity bars, or summary health labels in the preview as verified cluster health.

Agree on real source configuration and read-only access separately before any live validation. No real cluster validation or mutation was performed during this consolidation.

## Working agreement

Use the main repository directory and an ordinary local feature branch. Do not create new worktrees. Commit locally, report the branch/commit and checks, and wait for explicit instructions before any merge, push, deployment, or cluster change.

## Consolidation verification and efficiency follow-up

The combined preview is served on both 8181 and 8185 from the same newly built executable; these ports no longer represent separate designs. Either User/SRE switch stays on its current port. Both use local fixtures only.

Zone drill-down now uses two snapshot requests (zone summary and bounded graph), removing the extra zone-list lookup. Its related-inventory filter runs entirely in the browser. Graph requests are cancelled on navigation away, and direct relationship tables exclude unrelated traversal edges. Status refresh uses source metadata already present in the status response instead of requesting it again. No new API endpoints or metrics UI were added.

SRE HTTP handlers read immutable snapshots rather than querying Kubernetes. The independent collector scheduler limits concurrency, adds polling jitter, and waits for a collection cycle to finish before scheduling another. The configured default poll interval is 30 seconds; this is not a guarantee that every production-sized cluster can tolerate that interval. Tune it against actual collection duration and API request limits before live use. Existing `/metrics` instrumentation is available; no live load measurements were taken.

Validation: the frontend suite passed before the zone follow-up; all six focused zone, routing and lifecycle tests passed afterward, along with TypeScript checking and the final production frontend/backend builds. Backend tests passed; backend source was unchanged by this consolidation. The existing Vite large-bundle warning remains. Visual verification covers the User/SRE switch and zone drill-down; this is not a claim of a completed all-screen design audit.
