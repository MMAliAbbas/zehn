# Task 014: Idempotent Yaad Delegation Memory

Slug: `014-idempotent-yaad-delegation-memory`

Docs-only allowed: no

## Goal

Make delegation memory persistence idempotent so one delegation does not create
repeated long-term memories for requested, running, and completed transitions.

## Allowed repos/files

- `pkg/agent/delegation_memory.go`
- `pkg/agent/delegation_memory_test.go`
- `pkg/agent/delegation*.go`
- `pkg/agent/*delegation*_test.go`
- `pkg/tools/delegate*.go`
- `pkg/tools/delegate*_test.go`
- `docs/reference/**`
- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `pkg/agent/delegation.go`
- `pkg/agent/delegation_memory.go`
- `pkg/agent/delegation_records.go`
- `docs/reference/agent-delegation-meetings.md`
- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`
- Yaad MCP tool names available in the local Zehn config or tests, if present.

## Work

- Determine from local code/config/test doubles whether Yaad exposes an update
  or upsert-style MCP tool. Do not assume a tool exists without evidence.
- If an update/upsert tool exists, use a deterministic delegation memory key so
  repeated persistence updates the same memory.
- If no update/upsert tool exists, reduce writes to final and important
  blocker/failure states, and record skipped intermediate states locally.
- Keep strict mode behavior explicit and tested.
- Add tests proving a single delegation does not call `memory_add` repeatedly
  for normal requested/running/completed flow.

## Acceptance criteria

- Normal successful delegation creates or updates at most one durable Yaad
  memory entry.
- Failure/blocker transitions remain visible in local records and, where
  appropriate, durable memory.
- The implementation does not require live Yaad for tests.
- Local fallback behavior remains available when Yaad is unavailable.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./pkg/agent -run 'DelegationMemory|Delegation' -count=1
go test ./pkg/agent -run 'DelegationMemory|Delegation' -race
```
