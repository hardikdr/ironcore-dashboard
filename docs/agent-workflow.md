# Agent workflow

This project uses short, evidence-driven agent iterations. The goal is continuous
improvement without allowing an unattended model to accumulate unreviewed changes.

## Start an interactive lead-agent session

From this repository:

```bash
pi --name ironcore-dashboard-lead
```

Inside Pi, run `/improve` or `/improve accessibility`, review the proposed scope, and
let the agent complete one increment. Inspect the Git diff before committing.

## Safe autonomy ladder

1. **Interactive:** one task, human approves important decisions.
2. **Bounded loop:** two or three iterations, tests required after every iteration.
3. **Isolated worker:** run in a dedicated Git worktree or disposable VM.
4. **Scheduled worker:** a queue dispatches issues to isolated workers and opens draft PRs.
5. **Cluster workers:** Kubernetes Jobs run workers with resource limits, restricted
   credentials, persistent logs, and a human approval channel.

Do not jump directly to stage 5. A local Kubernetes cluster is useful for integration
testing, but it is not itself an agent orchestrator or a safety boundary.

## Definition of done for an iteration

- One bounded user-visible or engineering improvement.
- Acceptance criteria recorded in the session.
- Relevant tests/builds pass, or failures are clearly documented.
- Diff reviewed for scope, security, and maintainability.
- No push, deployment, or cluster mutation without explicit approval.

## Notifications

Use local desktop notifications first. For remote notifications, send events from the
orchestrator—not arbitrary agent shell commands—to a narrow webhook service. WhatsApp
requires an approved provider/API and should be added only after the local workflow is
stable; never give a coding worker unrestricted messaging credentials.

