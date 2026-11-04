# IronCore Dashboard Agent Guide

## Mission

Build a production-quality IronCore management dashboard inspired by the Gardener
Dashboard. Reuse interaction ideas and architecture lessons, but do not copy code
or branding blindly. Prefer small, testable improvements over broad rewrites.

## Workspace map

- `../dashboard`: Gardener Dashboard reference. Read-only.
- `../ironcore`: IronCore APIs and client libraries. Read-only unless explicitly asked.
- `../ironcore-in-a-box`: local cluster environment. Read-only unless explicitly asked.
- `.`: the IronCore Dashboard product. Make changes only here by default.

## Product priorities

1. Correct, safe handling of IronCore resources.
2. Clear resource status, errors, loading states, and destructive-action confirmation.
3. Fast operator workflows inspired by Gardener Dashboard.
4. Accessibility, responsive layout, and consistent visual language.
5. Automated tests and maintainable code.

## Required workflow

1. Inspect the current implementation and relevant reference code.
2. Select one bounded improvement with an observable acceptance criterion.
3. State the intended change before editing.
4. Implement the smallest coherent change.
5. Run focused tests, then broader verification when appropriate.
6. Review the diff for regressions, security issues, and accidental scope expansion.
7. Summarize what changed, evidence from verification, and the best next improvement.

Never silently weaken or delete tests. Never edit generated files unless the generating
workflow requires it. Do not modify `.env`, credentials, kubeconfigs, or Git metadata.
Do not push, deploy, merge, install system software, or contact external services unless
the user explicitly authorizes that action.

## Commands

```bash
# Full product build
make build

# Backend tests
go test ./...

# Frontend type-check and production build
cd frontend && npm run build

# Development servers
make dev-backend
make dev-frontend
```

Before choosing a command, inspect `Makefile` and the relevant package configuration.
Use the existing dependency manager and lockfiles.

## Iteration policy

For autonomous work, perform at most one product change per iteration. Keep a human in
the loop for architecture changes, new dependencies, API changes, authentication,
authorization, destructive operations, cluster changes, and external integrations.
Stop if tests repeatedly fail, requirements conflict, credentials are needed, or the
working tree contains unrelated changes that overlap the task.


## Local branch workflow (user preference)

Work in the main repository directory on an ordinary local feature branch. Do not
create new worktrees. Keep all changes and commits local. Do not merge into main
or push without explicit user instruction. Preserve existing worktrees as archives
unless the user separately authorizes their removal. See docs/local-console-development.md.
