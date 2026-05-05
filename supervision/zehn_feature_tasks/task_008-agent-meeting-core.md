# Task 008: Agent Meeting Core

Slug: `008-agent-meeting-core`

Docs-only allowed: no

## Goal

Implement the core chaired meeting workflow on top of delegation: sponsor,
chair, participants, meeting record, participant turns, and consolidated chair
recommendation.

## Allowed repos/files

- `pkg/agent/meeting*.go`
- `pkg/agent/meeting*_test.go`
- `pkg/agent/delegation*.go`
- `pkg/tools/meeting*.go`
- `pkg/tools/meeting*_test.go`
- `pkg/agent/agent_init.go`
- `pkg/config/config.go`
- `pkg/config/defaults.go`
- `.picoclaw/workspace/memory/AGENT_MEETING_SYSTEM.md`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `.picoclaw/workspace/memory/AGENT_MEETING_SYSTEM.md`
- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`
- `pkg/agent/delegation*.go`
- `pkg/tools/delegate*.go`
- `pkg/agent/registry.go`

## Work

- Add a meeting record schema with ID, title, sponsor, chair, participants,
  goal, constraints, notes, recommendation, timeline, risks, approvals, and
  artifact refs.
- Add `start_agent_meeting` capability or equivalent generic tool.
- Use delegation for each participant turn.
- Let the chair synthesize one consolidated recommendation.
- Store meeting records locally first.
- Keep meeting internals private and do not post raw debate to Discord or
  GitHub.

## Acceptance criteria

- CEO can sponsor a meeting chaired by a department head.
- Department head can chair a domain meeting.
- Participant responses are preserved in the meeting record.
- User-facing output defaults to one consolidated recommendation.
- Meeting record includes participants, timeline, risks, and follow-ups.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./pkg/agent -run 'Test.*Meeting|Test.*Delegation' -count=1 -v
go test ./pkg/tools -run 'Test.*Meeting|Test.*Delegate' -count=1 -v
go test ./pkg/agent ./pkg/tools
```
