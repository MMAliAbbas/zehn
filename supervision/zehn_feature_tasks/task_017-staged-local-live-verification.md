# Task 017: Staged Local Live Verification

Slug: `017-staged-local-live-verification`

Docs-only allowed: no

## Goal

Add a staged local verification plan and deterministic smoke-test helpers for
Zehn delegation/meeting rollout with Discord, Yaad, and GitHub enabled one at a
time.

## Allowed repos/files

- `operations/**`
- `docs/reference/**`
- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`
- `.picoclaw/workspace/memory/AGENT_MEETING_SYSTEM.md`
- `supervision/ZEHN_FEATURE_AUTOMATION_STATUS.md`
- `supervision/ZEHN_FEATURE_AUTOMATION_FAILURES.md`
- `supervision/ZEHN_FEATURE_LIVE_VERIFICATION.md`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `docs/reference/agent-delegation-meetings.md`
- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`
- `.picoclaw/workspace/memory/AGENT_MEETING_SYSTEM.md`
- `workspace/skills/picoclaw-project/references/channels.md`
- `workspace/skills/picoclaw-project/references/memory-context.md`
- `workspace/skills/picoclaw-project/references/security-operations.md`

## Work

- Document a staged local verification sequence:
  local CLI only, local gateway only, Yaad MCP enabled, GitHub artifact writer
  enabled with dry-run/fake mode if available, Discord summaries enabled, then
  one narrow live Discord command.
- Add smoke-test helper scripts only if they can run without secrets by default.
- The default verification must not send live Discord messages, write live
  GitHub issues, or write live Yaad memories.
- Include explicit operator gates for each live integration.
- Update the final automation status only after deterministic local checks pass.

## Acceptance criteria

- The live rollout plan is executable step by step without guessing.
- Live side effects are opt-in and clearly named.
- The plan covers rollback/disable steps for Discord, Yaad, and GitHub.
- Verification does not require live external services.

## Verification commands

```bash
cd /Users/aliai/zehn
bash -n operations/run-one-zehn-feature-task.sh
go test ./pkg/config -run '^$' -count=1
test -f supervision/ZEHN_FEATURE_LIVE_VERIFICATION.md
grep -i 'Discord' supervision/ZEHN_FEATURE_LIVE_VERIFICATION.md
grep -i 'Yaad' supervision/ZEHN_FEATURE_LIVE_VERIFICATION.md
grep -i 'GitHub' supervision/ZEHN_FEATURE_LIVE_VERIFICATION.md
operations/audit-zehn-feature-task.sh 017-staged-local-live-verification
```
