# Task 003: Target Agent Delegation Primitive

Slug: `003-target-agent-delegation-primitive`

Docs-only allowed: no

## Goal

Add a source-level primitive that can run a delegated turn against a real target
`AgentInstance` while preserving target workspace, prompt, model, tools, memory,
and sessions.

## Allowed repos/files

- `pkg/agent/delegation*.go`
- `pkg/agent/delegation*_test.go`
- `pkg/agent/subturn.go`
- `pkg/agent/subturn_test.go`
- `pkg/agent/agent_init.go`
- `pkg/agent/agent_test.go`
- `pkg/session/**`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `pkg/agent/subturn.go`
- `pkg/agent/turn_state.go`
- `pkg/agent/pipeline.go`
- `pkg/agent/pipeline_setup.go`
- `pkg/agent/pipeline_finalize.go`
- `pkg/agent/registry.go`
- `pkg/session/allocator.go`
- `pkg/session/key.go`
- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`

## Work

- Add a generic delegation request/result type.
- Add an internal `RunAgentDelegation` or equivalent method on `AgentLoop`.
- Enforce parent-to-target permission through `AgentRegistry.CanSpawnSubagent`.
- Resolve the target agent by ID and fail clearly when missing.
- Run the target as the target agent, not as a copy of the caller.
- Use a private internal delegation session scope.
- Preserve existing `spawn` and `subagent` behavior unless a test proves a
  narrow compatibility fix is required.

## Acceptance criteria

- Delegated execution uses the target agent ID in events, session scope, and
  context.
- Delegated execution uses the target workspace and target prompt files.
- Permission denial is explicit and tested.
- Missing target agent is explicit and tested.
- Existing subturn tests still pass.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./pkg/agent -run 'Test.*Delegat|TestSpawnSubTurn|TestAgentRegistry_CanSpawnSubagent' -count=1 -v
go test ./pkg/agent
```
