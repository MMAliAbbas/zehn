# Task 057: GitHub Artifact Repo Routing

Slug: `057-github-artifact-repo-routing`

Docs-only allowed: no

## Goal

Make the Zehn `GitHubArtifactWriter` from task 056 route each issue to the
correct LogicIgniter repo instead of posting everything to a single default
repo. A frontend-task issue should land in `logicigniter/svc-logicigniter-web`,
a backend-handler task for `svc-contentaudit-grpc` should land in
`logicigniter/svc-contentaudit-grpc`, doctrine work should land in
`logicigniter/supervision`, etc. — based on signals already present in the
delegation/meeting record (title, body, labels) plus the live LI repo
inventory on disk.

This is the follow-up the user explicitly demanded after reviewing the
task-056 PR: "Writer currently posts to one default repo — this is not
correct, please find a way that it decides the repo itself, if it need to
create a frontend issue it should be in proper repo not a default one."

## Allowed repos/files

- `pkg/tools/integration/github_artifacts.go` (Zehn-side; add `Repo` field)
- `pkg/agent/zehn_github_artifact_writer.go` (Zehn-only)
- `pkg/agent/zehn_github_repo_resolver.go` (new, Zehn-only)
- `pkg/agent/zehn_github_repo_resolver_test.go` (new, Zehn-only)
- `pkg/agent/zehn_github_artifact_writer_test.go` (Zehn-only)
- `supervision/zehn_feature_tasks/**`

## Required reading

- `pkg/tools/integration/github_artifacts.go`
- `pkg/agent/zehn_github_artifact_writer.go` (from task 056)
- `pkg/agent/github_artifacts.go` (the publishers that build the request)
- `supervision/zehn_feature_tasks/task_056-github-artifact-writer-implementation.md`
- `.picoclaw/workspace/memory/LADDER_SNAPSHOT_LATEST.md` (the kind of work
  that needs per-service routing today)

## Work

- Add an optional `Repo string` field to `integrationtools.GitHubIssueRequest`
  so callers can explicitly steer routing. Existing callers compile
  unchanged because the field is optional.
- Add `pkg/agent/zehn_github_repo_resolver.go` containing
  `zehnGitHubRepoResolver` with:
  - `liRoot` field (defaults to `/Users/aliai/logicigniter`, injectable
    for tests).
  - `defaultRepo` field for the fall-through case.
  - Lazy, once-only discovery of the known LI repo set by listing
    `liRoot` entries and keeping those whose `.git` is an actual
    directory (filters out worktree clones such as
    `svc-logicigniter-web-issue83` whose `.git` is a file).
- Resolution order, applied in `resolveRepo`:
  1. Explicit `req.Repo` (caller wins; bare names normalized to
     `<owner>/<name>`).
  2. `Target repo: <name>` line at the start of a body line.
  3. Longest exact-name match of a known repo in the title.
  4. Longest exact-name match of a known repo in the body.
  5. `area:*` label mapping (frontend, portal, supervision, integration,
     devops, ops, operations, docs, proto, infra, business, shared-ui).
     Only applied if the mapped repo exists in the discovered set.
  6. Fallback to `defaultRepo`.
- Word-boundary matching for repo names so `svc-bill` does **not** match
  inside `svc-billing-grpc`.
- Wire the resolver into `zehnGitHubArtifactWriter`'s constructor so
  `CreateIssue` calls `resolver.resolveRepo(req)` before invoking `gh`.

## Acceptance criteria

- `GitHubIssueRequest.Repo` exists and is optional.
- New resolver file exists with a public-to-package `resolveRepo` method.
- Discovery filters worktree clones (`.git` as file) and dotfile dirs.
- Resolver returns the right repo for each tested precedence rule.
- `CreateIssue` routes to the resolved repo, not the constructor-only
  default. Two integration tests in the writer test file demonstrate
  this end-to-end via `--repo <resolved>` in the captured args.
- Word-boundary matching prevents `svc-bill` from triggering on
  `svc-billing`.
- Resolver is nil-safe and tolerates a missing `liRoot`.
- All previous task-056 tests still pass; existing writer happy-path test
  is updated to use a hermetic `t.TempDir()`-based resolver so it does not
  depend on the contents of `/Users/aliai/logicigniter/` on the test host.

## Verification commands

```bash
cd /Users/aliai/zehn
go build ./pkg/agent/... ./pkg/tools/integration/...
go test ./pkg/agent -run 'GitHub|Artifact|Publisher|Resolver|Repo' -count=1
go test ./pkg/agent -run 'GitHub|Artifact|Publisher|Resolver|Repo' -count=1 -race
go vet ./pkg/agent ./pkg/tools/integration
```

## Out of scope

- Populating `req.Repo` from the delegation/meeting record itself. Today,
  the publishers in `pkg/agent/github_artifacts.go` build `GitHubIssueRequest`
  with `Repo` left empty, relying on the resolver to figure it out from
  title/body/labels. A future task can plumb an explicit `TargetRepo` field
  through the delegation/meeting records and the `delegate_to_agent` /
  `start_agent_meeting` tools so agents can steer routing directly without
  embedding repo hints in title/body. Not blocking — resolver-from-signals
  works for the immediate need (CTO sweep candidates).
- Repo creation. If the resolver picks a repo that doesn't exist on
  GitHub, `gh issue create` will error and the delegation's
  `github_artifact.status` will record `failed` with the gh stderr. That
  is the correct failure mode; explicit repo creation is a separate
  approval-gated action (`approval:new-repo`).
- Cross-org routing. Resolver always uses the same owner as `defaultRepo`.
  Multi-owner routing is not needed today.
