# Task 016: Async GitHub Artifact Publisher

Slug: `016-async-github-artifact-publisher`

Docs-only allowed: no

## Goal

Move GitHub issue/comment publishing for delegation and meetings behind a
controlled asynchronous artifact publisher so slow GitHub calls do not block
the user-visible result path.

## Allowed repos/files

- `pkg/agent/github_artifacts.go`
- `pkg/agent/github_artifacts_test.go`
- `pkg/agent/meeting*.go`
- `pkg/agent/meeting*_test.go`
- `pkg/agent/delegation*.go`
- `pkg/agent/*delegation*_test.go`
- `pkg/tools/integration/github*.go`
- `pkg/tools/integration/github*_test.go`
- `docs/reference/**`
- `.picoclaw/workspace/memory/AGENT_MEETING_SYSTEM.md`
- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `pkg/agent/github_artifacts.go`
- `pkg/tools/integration/github.go`
- `pkg/agent/meeting.go`
- `pkg/agent/delegation.go`
- `docs/reference/agent-delegation-meetings.md`

## Work

- Introduce a bounded or otherwise controlled GitHub artifact publisher path.
- Meeting/delegation completion should return after local record completion,
  not after all GitHub network work completes.
- Preserve artifact status recording: pending, created, failed, and skipped
  should be distinguishable if the existing record model can support it cleanly.
- Add timeouts or context boundaries for GitHub publishing.
- Add tests for successful async publish, publish failure recorded without
  failing the completed meeting/delegation, and disabled writer behavior.

## Acceptance criteria

- GitHub publishing cannot indefinitely delay a completed meeting or
  delegation response.
- Publishing errors are visible in local records.
- Tests use fake writers only and do not call live GitHub.
- Existing issue/comment body formatting remains covered.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./pkg/agent ./pkg/tools/integration -run 'GitHub|Artifact|Meeting|Delegation' -count=1
go test ./pkg/agent ./pkg/tools/integration -run 'GitHub|Artifact|Meeting|Delegation' -race
```
