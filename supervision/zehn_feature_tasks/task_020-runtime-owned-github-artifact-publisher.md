# Task 020: Runtime Owned GitHub Artifact Publisher

Slug: `020-runtime-owned-github-artifact-publisher`

Docs-only allowed: no

## Goal

Move GitHub artifact publishing from a package-global singleton to a runtime
owned component with explicit capacity, timeout, shutdown, and deterministic
test behavior.

## Allowed repos/files

- `pkg/agent/github_artifacts.go`
- `pkg/agent/github_artifacts_test.go`
- `pkg/agent/agent*.go`
- `pkg/agent/instance*.go`
- `pkg/agent/meeting*.go`
- `pkg/agent/meeting*_test.go`
- `pkg/agent/delegation*.go`
- `pkg/agent/*delegation*_test.go`
- `pkg/config/config.go`
- `pkg/config/defaults.go`
- `pkg/config/*_test.go`
- `docs/reference/**`
- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`
- `.picoclaw/workspace/memory/AGENT_MEETING_SYSTEM.md`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `pkg/agent/github_artifacts.go`
- `pkg/agent/agent_init.go`
- `pkg/agent/instance.go`
- `pkg/agent/delegation_executor.go`
- `pkg/config/config.go`
- `pkg/config/defaults.go`
- `docs/reference/agent-delegation-meetings.md`

## Work

- Replace the package-global GitHub artifact publisher with an instance owned
  by `AgentLoop` or the nearest existing runtime lifecycle owner.
- Add explicit shutdown/drain behavior so accepted publish jobs are not left in
  an unmanaged global goroutine at runtime close.
- Make publisher capacity and timeout configurable if there is an existing
  appropriate config section; otherwise keep conservative defaults close to the
  lifecycle owner without adding broad configuration churn.
- Preserve non-blocking user-visible completion for meetings and delegations.
- Preserve recorded artifact states: pending, created, failed, and skipped.
- Add tests for capacity rejection, timeout/failure recording, shutdown/drain,
  and isolation between separate runtime instances.

## Acceptance criteria

- No package-global mutable publisher is required for normal GitHub artifact
  publishing.
- Runtime shutdown has a deliberate story for accepted publisher jobs.
- Separate runtime instances do not share publisher capacity accidentally.
- Existing async artifact tests continue to pass under normal and race modes.
- The implementation remains generic PicoClaw code.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./pkg/agent ./pkg/config -run 'GitHub|Artifact|Publisher|Meeting|Delegation|Config' -count=1
go test ./pkg/agent ./pkg/config -run 'GitHub|Artifact|Publisher|Meeting|Delegation|Config' -race
operations/audit-zehn-feature-task.sh 020-runtime-owned-github-artifact-publisher
```
