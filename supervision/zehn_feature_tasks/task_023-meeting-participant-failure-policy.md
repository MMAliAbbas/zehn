# Task 023: Meeting Participant Failure Policy

Slug: `023-meeting-participant-failure-policy`

Docs-only allowed: no

## Goal

Make chaired meeting v1 participant failure handling explicit and configurable
enough for reliable operational workflows without implementing full v2 debate.

## Allowed repos/files

- `pkg/agent/meeting*.go`
- `pkg/agent/meeting*_test.go`
- `pkg/tools/meeting*.go`
- `pkg/tools/meeting*_test.go`
- `pkg/config/config.go`
- `pkg/config/defaults.go`
- `pkg/config/*_test.go`
- `docs/reference/**`
- `.picoclaw/workspace/memory/AGENT_MEETING_SYSTEM.md`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `pkg/agent/meeting.go`
- `pkg/agent/meeting_records.go`
- `pkg/tools/meeting.go`
- `pkg/config/config.go`
- `docs/reference/agent-delegation-meetings.md`
- `.picoclaw/workspace/memory/AGENT_MEETING_SYSTEM.md`

## Work

- Define the meeting v1 failure policy for participant errors, chair errors,
  context cancellation, and record-store failures.
- Preserve the default conservative behavior unless a safer existing config
  pattern supports optional participants or partial completion.
- If optional participant support is added, make it explicit in the tool schema
  and record output, and ensure chair synthesis receives a clear failed/absent
  participant list.
- Add tests for participant failure, chair failure, cancellation, and any
  optional/partial behavior introduced.
- Update docs so operators know when a meeting fails versus completes with
  missing participant input.

## Acceptance criteria

- Meeting failure behavior is explicit in code, docs, and tests.
- Partial or optional participant behavior, if added, is never implicit.
- Chair synthesis cannot silently ignore failed required participants.
- Existing meeting v1 behavior remains understandable and deterministic.
- The implementation remains generic PicoClaw code.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./pkg/agent ./pkg/tools ./pkg/config -run 'Meeting|meeting|Participant|Failure|Cancel' -count=1
go test ./pkg/agent ./pkg/tools ./pkg/config -run 'Meeting|meeting|Participant|Failure|Cancel' -race
operations/audit-zehn-feature-task.sh 023-meeting-participant-failure-policy
```
