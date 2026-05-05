# Task 009: GitHub Meeting Artifacts

Slug: `009-github-meeting-artifacts`

Docs-only allowed: no

## Goal

Add optional GitHub issue/comment/project artifact integration for delegation
and meetings while keeping GitHub as the work tracker, not the company brain.

## Allowed repos/files

- `pkg/agent/meeting*.go`
- `pkg/agent/delegation*.go`
- `pkg/tools/meeting*.go`
- `pkg/tools/delegate*.go`
- `pkg/tools/integration/**`
- `pkg/config/config.go`
- `pkg/config/defaults.go`
- `.picoclaw/workspace/memory/AGENT_MEETING_SYSTEM.md`
- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `.picoclaw/workspace/memory/AGENT_MEETING_SYSTEM.md`
- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`
- `pkg/tools/integration/**`
- current GitHub-related workspace skills under `.picoclaw/workspace/skills/github`

## Work

- Add a narrow GitHub artifact interface with fake-test implementation.
- Create issues only when executable or approval-tracked work exists.
- Add curated meeting summaries as issue body/comments.
- Add focused participant comments only for material positions, risks,
  commitments, dependencies, or acceptance criteria.
- Avoid posting raw internal transcript text.
- Do not require live GitHub access in unit tests.

## Acceptance criteria

- Meeting can produce no GitHub issue when no executable work exists.
- Meeting can produce one issue when executable work exists.
- Participant comments are curated and scoped.
- GitHub failures do not lose the meeting record.
- GitHub Project remains a tracker and is not treated as durable memory.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./pkg/agent -run 'Test.*Meeting.*GitHub|Test.*Delegation.*GitHub' -count=1 -v
go test ./pkg/tools ./pkg/agent
```
