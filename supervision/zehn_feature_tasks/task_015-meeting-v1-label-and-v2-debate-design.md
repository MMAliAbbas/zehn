# Task 015: Meeting V1 Label And V2 Debate Design

Slug: `015-meeting-v1-label-and-v2-debate-design`

Docs-only allowed: no

## Goal

Make the current chaired meeting implementation honest and explicit as
`meeting v1`, then design the next multi-round debate loop without pretending
the current sequential participant flow is real-time discussion.

## Allowed repos/files

- `pkg/agent/meeting*.go`
- `pkg/agent/meeting*_test.go`
- `pkg/tools/meeting*.go`
- `pkg/tools/meeting*_test.go`
- `docs/reference/**`
- `.picoclaw/workspace/memory/AGENT_MEETING_SYSTEM.md`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `pkg/agent/meeting.go`
- `pkg/agent/meeting_records.go`
- `pkg/tools/meeting.go`
- `docs/reference/agent-delegation-meetings.md`
- `.picoclaw/workspace/memory/AGENT_MEETING_SYSTEM.md`

## Work

- Update docs and user-visible tool descriptions so the current implementation
  is named as a chaired sequential meeting flow.
- Preserve the current chair synthesis behavior for v1.
- Add a design section for v2 multi-round debate: turn order, participant
  visibility into prior turns, chair interventions, stopping criteria, token
  limits, failure handling, and audit trail.
- Add tests where code-facing labels or tool schemas change.
- Do not implement v2 debate mechanics in this task unless it is a tiny
  compatibility-neutral scaffold required by the docs/tests.

## Acceptance criteria

- No docs or tool text imply v1 is live real-time debate.
- The v2 debate design is specific enough to become implementation tasks.
- Existing meeting tests still pass.
- Meeting output still includes recommendation, participants, timeline, risks,
  approvals, and follow-ups.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./pkg/agent ./pkg/tools -run 'Meeting|meeting' -count=1
go test ./pkg/agent ./pkg/tools -run 'Meeting|meeting' -race
```
