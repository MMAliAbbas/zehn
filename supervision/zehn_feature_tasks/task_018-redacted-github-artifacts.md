# Task 018: Redacted GitHub Artifacts

Slug: `018-redacted-github-artifacts`

Docs-only allowed: no

## Goal

Ensure meeting and delegation GitHub issues/comments are built from redacted
artifact data so external tracker publishing cannot leak secrets or private
operator input.

## Allowed repos/files

- `pkg/agent/github_artifacts.go`
- `pkg/agent/github_artifacts_test.go`
- `pkg/agent/delegation*.go`
- `pkg/agent/*delegation*_test.go`
- `pkg/agent/meeting*.go`
- `pkg/agent/meeting*_test.go`
- `docs/reference/**`
- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`
- `.picoclaw/workspace/memory/AGENT_MEETING_SYSTEM.md`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `pkg/agent/github_artifacts.go`
- `pkg/agent/delegation_store.go`
- `pkg/agent/meeting_store.go`
- `pkg/agent/delegation_memory.go`
- `pkg/agent/github_artifacts_test.go`
- `docs/reference/agent-delegation-meetings.md`

## Work

- Audit every GitHub issue body, issue title, and comment body generated for
  delegation and meeting artifacts.
- Build GitHub artifact content from redacted records where possible.
- Where raw outcome/request/result values must still be used, pass all strings
  through the same sensitive-data redaction path used by local records.
- Preserve useful tracker context: chair, sponsor, target, status, priority,
  recommendation, timeline, risks, approvals, follow-ups, and artifact links.
- Add deterministic tests with fake secret values in delegation task text,
  delegation result text, meeting recommendation, risks, timeline, and follow-up
  fields.
- Tests must assert that raw fake secrets are absent from GitHub issue and
  comment bodies while redacted placeholders remain readable.

## Acceptance criteria

- GitHub artifact publishing never bypasses the established redaction path.
- Existing local record redaction behavior is not weakened.
- Existing async artifact publishing behavior still records pending, created,
  failed, and skipped states as appropriate.
- Tests use fake writers only and do not call live GitHub.
- The implementation remains generic PicoClaw code.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./pkg/agent -run 'GitHub|Artifact|Redact|Delegation|Meeting' -count=1
go test ./pkg/agent -run 'GitHub|Artifact|Redact|Delegation|Meeting' -race
operations/audit-zehn-feature-task.sh 018-redacted-github-artifacts
```
