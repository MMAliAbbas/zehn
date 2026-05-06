# Task 021: Upstream Publishability Audit

Slug: `021-upstream-publishability-audit`

Docs-only allowed: no

## Goal

Add a non-destructive publishability audit that identifies private/local files,
branch-history risks, and contribution-splitting requirements before any
upstream-facing push or pull request.

## Allowed repos/files

- `operations/**`
- `docs/reference/**`
- `supervision/ZEHN_UPSTREAM_PUBLISHING_CHECKLIST.md`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `CONTRIBUTING.md`
- `AGENTS.md`
- `operations/run-one-zehn-feature-task.sh`
- `operations/audit-zehn-feature-task.sh`
- `docs/reference/agent-delegation-meetings.md`
- `workspace/skills/picoclaw-project/references/upstream-contribution.md`

## Work

- Create a non-destructive audit script or extend an existing audit script to
  report publishability risks without rewriting history.
- The audit must detect tracked private/local skill paths in the current tree
  and in recent local history.
- The audit must detect task/supervision files that are private automation
  artifacts and should not be included in upstream-focused commits.
- Document the exact safe cleanup strategy for preparing upstream-clean
  contribution branches, including branch splitting and history rebuild steps
  that require explicit operator approval.
- Do not run destructive git commands and do not rewrite history in this task.

## Acceptance criteria

- There is an operator-facing checklist for upstream publish preparation.
- There is a deterministic local audit command that reports current
  publishability blockers.
- The audit is advisory by default and safe to run unattended.
- The task does not remove private files, rewrite commits, force-push, or
  change remotes.

## Verification commands

```bash
cd /Users/aliai/zehn
bash -n operations/run-one-zehn-feature-task.sh
bash -n operations/audit-zehn-feature-task.sh
operations/audit-zehn-feature-task.sh --publishability-self-test
go test ./pkg/agent -run '^$' -count=1
test -f supervision/ZEHN_UPSTREAM_PUBLISHING_CHECKLIST.md
grep -i 'history' supervision/ZEHN_UPSTREAM_PUBLISHING_CHECKLIST.md
grep -i 'workspace/skills' supervision/ZEHN_UPSTREAM_PUBLISHING_CHECKLIST.md
operations/audit-zehn-feature-task.sh 021-upstream-publishability-audit
```
