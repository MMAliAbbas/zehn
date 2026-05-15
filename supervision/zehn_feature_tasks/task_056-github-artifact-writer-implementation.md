# Task 056: GitHub Artifact Writer Implementation

Slug: `056-github-artifact-writer-implementation`

Docs-only allowed: no

## Goal

Land the production `GitHubArtifactWriter` implementation that tasks 009 and
020 scaffolded but never wired in. Every delegation and meeting record
currently records `github_artifact: skipped — github artifact writer disabled`
because `al.githubArtifacts` is nil at runtime. The interface, storage
records, async publisher, and test fakes all exist; only the production
writer and the init-time wire-up are missing.

This task adds both, in Zehn-only files, with minimal touch to upstream-clean
code so future upstream sync from `sipeed/picoclaw` stays low-conflict.

## Allowed repos/files

- `pkg/agent/zehn_github_artifact_writer.go` (new, Zehn-only)
- `pkg/agent/zehn_github_artifact_writer_test.go` (new, Zehn-only)
- `pkg/agent/zehn_init_hook.go` (new, Zehn-only)
- `pkg/agent/agent_init.go` (additive only — one helper call)
- `supervision/zehn_feature_tasks/**`

## Required reading

- `pkg/agent/github_artifacts.go`
- `pkg/agent/agent_init.go`
- `pkg/tools/integration/github_artifacts.go`
- `pkg/agent/github_artifacts_test.go`
- `supervision/zehn_feature_tasks/task_009-github-meeting-artifacts.md`
- `supervision/zehn_feature_tasks/task_020-runtime-owned-github-artifact-publisher.md`

## Work

- Implement a fork-private `GitHubArtifactWriter` that shells out to the `gh`
  CLI. The implementation lives in `pkg/agent/zehn_github_artifact_writer.go`
  and is prefixed `zehn_` so the file name is obvious during upstream sync
  and cannot collide with future upstream additions.
- The writer's `CreateIssue` invokes `gh issue create --repo <repo> --title …
  --body … --label …` and parses the printed issue URL on success.
- The writer's `CreateComment` invokes `gh issue comment <number-or-url>
  --body … [--repo …]` and treats a clean exit as success.
- `execCmd` is an injectable field so tests can substitute a fake without
  touching `os/exec`.
- The writer enforces a per-call timeout (default 30s) honored via
  `context.WithTimeout`.
- Add `pkg/agent/zehn_init_hook.go` with one helper
  `wireZehnGitHubArtifactWriter(al, cfg)` that:
  - Returns immediately if `al` is nil.
  - Returns immediately if `gh` is not on `PATH` (so test/CI runs without
    `gh` continue to see a nil writer and the legacy "disabled" record).
  - Otherwise constructs the writer with a sensible default repo
    (`logicigniter/supervision`) and calls
    `al.SetGitHubArtifactWriter(...)`.
- Add **one call** to `wireZehnGitHubArtifactWriter(al, cfg)` in
  `pkg/agent/agent_init.go` immediately after `registerSharedTools(...)`
  and before `return al`. This is the only edit to a file shared with
  upstream-clean code paths; the call is additive and behaves as a no-op
  when `gh` is unavailable.

## Acceptance criteria

- New file `pkg/agent/zehn_github_artifact_writer.go` exists with
  `CreateIssue` and `CreateComment` methods that satisfy
  `integrationtools.GitHubArtifactWriter`.
- New file `pkg/agent/zehn_init_hook.go` exists with
  `wireZehnGitHubArtifactWriter`.
- `pkg/agent/agent_init.go` carries exactly one helper call
  (`wireZehnGitHubArtifactWriter(al, cfg)`) plus an explanatory comment
  line; no other edits.
- New unit tests in `pkg/agent/zehn_github_artifact_writer_test.go` cover:
  - `CreateIssue` happy path returns the parsed number + URL.
  - `CreateIssue` rejects empty title.
  - `CreateIssue` propagates `gh` errors.
  - `CreateComment` happy path.
  - `parseIssueNumberFromURL` handles canonical URLs, trailing slash,
    malformed input.
- `go test ./pkg/agent -run 'GitHub|Artifact' -count=1 -race` passes.
- Existing async-publisher tests in `github_artifacts_test.go` continue to
  pass (they construct `AgentLoop` directly without going through
  `newAgentLoop`, so the wire-up does not affect them).
- The change is additive only; no behavior change for environments where
  `gh` is not on PATH or fails.

## Verification commands

```bash
cd /Users/aliai/zehn
go build ./...
go test ./pkg/agent -run 'GitHub|Artifact' -count=1
go test ./pkg/agent -run 'GitHub|Artifact' -count=1 -race
go vet ./pkg/agent
operations/audit-zehn-feature-task.sh 056-github-artifact-writer-implementation
```

## Out of scope

- Per-service repo routing (writer currently posts to a single default repo).
  Cross-repo routing requires extending `GitHubIssueRequest` with a `Repo`
  field and updating the delegation/meeting publishers to populate it.
  Tracked as follow-up.
- Sync with `upstream/main` (189 commits behind as of 2026-05-15). Tracked
  separately so a sync conflict does not muddy this fix's verification.
- Discord output formatting for created issues; existing publishers already
  carry the URL into the delegation record, which Discord summaries reference.
