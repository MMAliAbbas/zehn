# Zehn Upstream Publishing Checklist

Use this checklist before any upstream-facing push or pull request from the
Zehn workspace. The goal is to prepare a branch that is clean for PicoClaw
upstream review without deleting private files, rewriting history, changing
remotes, or force-pushing unless the operator explicitly approves that step.

## Advisory Audit

Run the local publishability audit from the repository root:

```bash
operations/audit-zehn-feature-task.sh 021-upstream-publishability-audit
```

The audit is advisory and safe to run unattended. It reports tracked
publishability risks in the current tree and in recent local history, including:

- `workspace/skills/**` and `.picoclaw/workspace/skills/**` paths.
- `supervision/zehn_feature_tasks/**` task files.
- `supervision/ZEHN_FEATURE_*` automation ledgers or prompt files.
- `supervision/ZEHN_UPSTREAM_PUBLISHING_CHECKLIST.md` itself.

Treat every warning as a blocker for an upstream-focused branch even though the
audit command exits successfully for feature-task automation.

## Current Tree Cleanup

1. Inspect the audit output and `git status --short`.
2. Decide which files are intended for upstream. PicoClaw-generic source,
   tests, and docs may belong upstream; Zehn supervision, local skills,
   private memory, Yaad-specific configuration, GitHub project automation, and
   Discord visibility operations do not.
3. Move upstream-worthy changes onto a clean branch based on upstream `main`.
4. Leave private paths such as `workspace/skills/**` and `supervision/**`
   unstaged for upstream commits.
5. Before committing, run:

```bash
git diff --name-only --cached
```

Confirm the staged set contains only upstream-intended paths.

## Branch Splitting

Split mixed Zehn work into separate branches before opening a pull request:

- One upstream branch per PicoClaw-generic bug fix, test hardening change, or
  documentation clarification.
- One private Zehn branch for supervision files, feature task records,
  `workspace/skills/**`, local memory, private adapters, or operator workflow
  changes.
- No upstream branch should depend on private Zehn automation state.

Safe branch-splitting sequence:

1. Create or update a clean upstream base locally.
2. Create a new `fix/...` or `feature/...` branch name that avoids `codex`,
   `agent`, and `ai`.
3. Select only upstream-safe changes with non-destructive staging commands such
   as `git add -p` or explicit path staging.
4. Commit the focused upstream change.
5. Run the publishability audit again and verify no warning path is staged or
   required by the branch.

## History Cleanup

If the audit reports private paths in history, do not push the branch upstream.
Recent history containing `workspace/skills/**`, `.picoclaw/**`, or
`supervision/**` private automation paths can expose local implementation
details even when the current tree looks clean.

History rebuild steps require explicit operator approval because they rewrite
commit identities and may require replacing local branches:

1. Confirm the upstream target and exact commit range to keep.
2. Create a backup branch or tag for the current local state.
3. Rebuild a clean branch from upstream `main` using cherry-pick, patch export,
   or manual reapplication of only upstream-safe changes.
4. Re-run tests and the publishability audit on the rebuilt branch.
5. Push only the rebuilt branch. Use force-push only when the operator has
   explicitly approved the remote branch replacement.

Never run destructive cleanup commands as part of the audit. The audit reports
risks; the operator chooses and approves any branch rebuild, history rewrite, or
remote update.
