# Zehn Feature Automation Branch Hygiene

Updated: 2026-05-06T03:05:00+05:00

## Current Blocker

Do not push the current Zehn feature automation history until the local-skill
leak has been removed from history.

Commit `a25b1f52` (`chore(zehn-features): complete 004-delegate-tool-sync`)
added private local skill files under:

- `workspace/skills/picoclaw-project/**`
- `workspace/skills/picoclaw-usage/**`

Those files are operator-local PicoClaw/Zehn working material, not upstream
PicoClaw source. They must not appear in a pushed branch or pull request.

## Why A Delete Commit Is Not Enough

A follow-up commit that deletes `workspace/skills/picoclaw-project/**` and
`workspace/skills/picoclaw-usage/**` would only remove the files from the final
tree. The leaked content would still remain in commit `a25b1f52`, so anyone who
can fetch the branch could still inspect it with `git show`, `git checkout`, or
history search. Because the branch is still unpushed, the correct cleanup is to
rewrite or rebuild the branch before the first push.

## Required Cleanup After Hardening Passes

Run this cleanup only after all branch-hardening tasks are green. Do not do it
inside task 012.

Preferred path: rebuild a clean branch from the last safe published base and
cherry-pick only reviewed commits, skipping the leaked commit content.

```bash
cd /Users/aliai/zehn
git status --short
git branch backup/zehn-feature-automation-leaky-history
git switch -c feature/zehn-delegation-meeting-clean 7ef2d355
git cherry-pick 9cf1c579
git cherry-pick 8d143567
git cherry-pick 5339f546
git cherry-pick a25b1f52 --no-commit
git restore --staged --worktree -- workspace/skills/picoclaw-project workspace/skills/picoclaw-usage
git commit -m "chore(zehn-features): complete 004-delegate-tool-sync"
git cherry-pick 063ee2d6
git cherry-pick 9f4988ad
git cherry-pick 30aaaab4
git cherry-pick ec5de1b4
git cherry-pick 64a5be8e
git cherry-pick af697dbb
git cherry-pick 5f742a6b
git cherry-pick dcd0530e
```

Then verify that the rebuilt branch has no tracked local skills and no leaked
skill paths in history:

```bash
git ls-files 'workspace/skills/**'
git log --name-only --format='%H' -- workspace/skills/picoclaw-project workspace/skills/picoclaw-usage
```

Both commands must print no file paths before the branch is pushed.

If the cherry-pick sequence conflicts, stop and resolve only the scoped source
changes from the selected commits. Do not copy `workspace/skills/**` into the
rebuilt branch. Keep `backup/zehn-feature-automation-leaky-history` local until
the clean branch has been reviewed and pushed successfully.
