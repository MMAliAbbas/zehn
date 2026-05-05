# Task 012: Branch Hygiene And Runner Scoped Staging

Slug: `012-branch-hygiene-and-runner-scoped-staging`

Docs-only allowed: no

## Goal

Prevent future Zehn feature automation runs from committing local skills,
private operational notes, or other unscoped files, and document the required
pre-push cleanup for the already-unpushed task history.

## Allowed repos/files

- `operations/run-one-zehn-feature-task.sh`
- `operations/audit-zehn-feature-task.sh`
- `supervision/ZEHN_FEATURE_AUTOMATION_STATUS.md`
- `supervision/ZEHN_FEATURE_AUTOMATION_FAILURES.md`
- `supervision/ZEHN_FEATURE_AUTOMATION_BRANCH_HYGIENE.md`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `operations/run-one-zehn-feature-task.sh`
- `operations/audit-zehn-feature-task.sh`
- `supervision/ZEHN_FEATURE_AUTOMATION_STATUS.md`
- `supervision/ZEHN_FEATURE_AUTOMATION_FAILURES.md`
- `CONTRIBUTING.md`
- `workspace/skills/picoclaw-project/references/upstream-contribution.md`

## Work

- Ensure auto-commit stages only files matched by the selected task's
  `Allowed repos/files`.
- Add or update automation checks so local skills under `workspace/skills/**`
  cannot be included by future feature task commits unless a task explicitly
  allows that path and the reviewer accepts it.
- Document the current unpushed-history blocker: commit `a25b1f52` added local
  `workspace/skills/picoclaw-project/**` and `workspace/skills/picoclaw-usage/**`
  files, so the branch must be rewritten or rebuilt before pushing.
- Do not rewrite history inside this task. Produce exact operator instructions
  for the cleanup step that should happen after all hardening tasks pass.

## Acceptance criteria

- The runner cannot stage arbitrary dirty files during `--commit`.
- The task scope guard still rejects changes outside the selected task scope.
- The hygiene document names the local-skill leak, why a delete commit is not
  enough, and the preferred cleanup path before push.
- Verification does not require network access or live Zehn services.

## Verification commands

```bash
cd /Users/aliai/zehn
bash -n operations/run-one-zehn-feature-task.sh
bash -n operations/audit-zehn-feature-task.sh
go test ./pkg/config -run '^$' -count=1
operations/audit-zehn-feature-task.sh --runner-scope-self-test
operations/audit-zehn-feature-task.sh 012-branch-hygiene-and-runner-scoped-staging
git show --name-only --oneline a25b1f52 | grep 'workspace/skills/picoclaw-project'
git show --name-only --oneline a25b1f52 | grep 'workspace/skills/picoclaw-usage'
```
