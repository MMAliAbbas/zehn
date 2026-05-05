# Task 007: Yaad Delegation Persistence

Slug: `007-yaad-delegation-persistence`

Docs-only allowed: no

## Goal

Connect delegation records to Yaad as the preferred durable memory layer while
keeping the generic delegation core usable without Yaad.

## Allowed repos/files

- `pkg/agent/delegation*.go`
- `pkg/agent/delegation*_test.go`
- `pkg/mcp/**`
- `pkg/tools/integration/mcp_tool.go`
- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`
- `.picoclaw/workspace/memory/MEMORY.md`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`
- `.picoclaw/workspace/memory/MEMORY.md`
- `/Users/aliai/.codex/skills/yaad-memory/SKILL.md`
- `pkg/mcp/manager.go`
- `pkg/agent/agent_mcp.go`
- `pkg/tools/integration/mcp_tool.go`

## Work

- Add a narrow interface for durable delegation memory writes.
- Implement a Yaad-backed adapter through existing MCP facilities where safe.
- Fall back to local record store when Yaad is unavailable.
- Never block core delegation on Yaad outages unless configured as strict.
- Store request, result, status, decisions, and follow-ups in Yaad.

## Acceptance criteria

- Delegation still works when Yaad is unavailable.
- Yaad write success is recorded in the local delegation record.
- Yaad write failure is recorded without losing the delegation result.
- Tests use fake Yaad/MCP adapters and do not call the real Yaad service.
- No secrets are written to Yaad records.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./pkg/agent -run 'Test.*Delegation.*Yaad|Test.*Delegation.*Memory' -count=1 -v
go test ./pkg/mcp ./pkg/tools/integration ./pkg/agent
```
