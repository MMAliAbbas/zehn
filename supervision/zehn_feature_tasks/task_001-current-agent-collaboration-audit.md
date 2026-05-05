# Task 001: Current Agent Collaboration Audit

Slug: `001-current-agent-collaboration-audit`

Docs-only allowed: yes

## Goal

Create an evidence-backed audit of the current Zehn/PicoClaw multi-agent,
subturn, routing, session, and message-tool behavior before implementing
delegation or meetings.

## Allowed repos/files

- `supervision/zehn_feature_tasks/**`
- `supervision/ZEHN_FEATURE_AUTOMATION_STATUS.md`
- `supervision/ZEHN_FEATURE_AUTOMATION_FAILURES.md`
- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`
- `.picoclaw/workspace/memory/AGENT_MEETING_SYSTEM.md`
- `.picoclaw/workspace/memory/ZEHN_SETUP_PLANNING.md`
- `docs/architecture/**`
- `docs/reference/**`

## Required reading

- `pkg/agent/agent_message.go`
- `pkg/agent/agent_init.go`
- `pkg/agent/registry.go`
- `pkg/agent/subturn.go`
- `pkg/tools/spawn.go`
- `pkg/tools/subagent.go`
- `pkg/tools/integration/message.go`
- `pkg/channels/discord/discord.go`
- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`
- `.picoclaw/workspace/memory/AGENT_MEETING_SYSTEM.md`

## Work

- Document current behavior of inbound routing, subturn spawning, `spawn`,
  `subagent`, `message`, Discord self-message handling, session allocation, and
  target-agent permission checks.
- Prove whether `agent_id` in `spawn` currently selects the target agent or only
  gates permission.
- Identify exact source files that must change for target-agent delegation.
- Identify exact source files that should remain unchanged to preserve upstream
  behavior.
- Record upstream issue references for PicoClaw multi-agent discovery and
  delegation.

## Acceptance criteria

- The audit states what exists today, what is missing, and what must not be
  overloaded.
- The audit cites concrete source files and functions.
- The audit confirms Discord is visibility only and not the internal delegation
  bus.
- No runtime code is changed.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./pkg/agent ./pkg/tools
rg -n "agent-to-agent|delegate_to_agent|start_agent_meeting|spawn" .picoclaw/workspace/memory supervision/zehn_feature_tasks
```
