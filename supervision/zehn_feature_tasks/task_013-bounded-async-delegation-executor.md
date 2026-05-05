# Task 013: Bounded Async Delegation Executor

Slug: `013-bounded-async-delegation-executor`

Docs-only allowed: no

## Goal

Replace unbounded async delegation goroutine launch with a controlled executor
that has clear capacity, cancellation, shutdown, and testable failure behavior.

## Allowed repos/files

- `pkg/agent/delegation*.go`
- `pkg/agent/*delegation*_test.go`
- `pkg/agent/agent*.go`
- `pkg/agent/instance*.go`
- `pkg/config/config.go`
- `pkg/config/defaults.go`
- `pkg/config/*_test.go`
- `pkg/tools/delegate*.go`
- `pkg/tools/delegate*_test.go`
- `docs/reference/**`
- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `pkg/agent/delegation.go`
- `pkg/agent/delegation_records.go`
- `pkg/tools/delegate.go`
- `pkg/agent/turn_state.go`
- `pkg/agent/instance.go`
- `docs/reference/agent-delegation-meetings.md`
- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`

## Work

- Remove `RunAgentDelegationAsync`'s unbounded `go func()` path using
  `context.Background()`.
- Add a bounded async executor or queue owned by `AgentLoop` or a nearby
  lifecycle owner.
- Define behavior when the queue is full, the parent context is canceled before
  enqueue, the loop is shutting down, and an async task fails after enqueue.
- Preserve existing sync delegation semantics.
- Record rejected/failed async delegations in the delegation record store with
  user-visible status.
- Add deterministic tests for capacity, cancellation before enqueue, successful
  async completion, and shutdown/no goroutine leak behavior where practical.

## Acceptance criteria

- Async delegation cannot create unlimited goroutines under load.
- Async work does not ignore cancellation before it is accepted.
- Accepted async work has a deliberate runtime context and shutdown story.
- Existing `delegate` sync behavior and existing `spawn`/`subagent` tests still
  pass.
- The implementation is generic PicoClaw code, not Zehn-specific.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./pkg/agent ./pkg/tools ./pkg/config -run 'Delegation|Delegate|Async' -count=1
go test ./pkg/agent ./pkg/tools ./pkg/config -race
```
