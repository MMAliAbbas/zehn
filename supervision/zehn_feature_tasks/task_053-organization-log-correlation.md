# Task 053: Organization Log Correlation

Slug: `053-organization-log-correlation`

Docs-only allowed: no

## Goal

Make the Live Logs workbench section useful for investigating a selected agent
or activity record by filtering or highlighting relevant log lines without
changing gateway logging behavior.

## Allowed repos/files

- `web/frontend/src/api/**`
- `web/frontend/src/components/agent/organization/**`
- `web/frontend/src/i18n/locales/**`
- `docs/reference/**`
- `supervision/**`

## Required reading

- `supervision/ZEHN_AGENT_ORGANIZATION_DIAGNOSTICS_PLAN.md`
- `web/frontend/src/components/agent/organization/detail-panels.tsx`
- `web/frontend/src/components/agent/organization/status-components.tsx`
- `web/frontend/src/api/**`
- `docs/reference/agent-organization-live-verification.md`

## Work

- Add selected-agent and selected-record correlation controls to the existing
  organization live-log panel.
- Highlight or filter log lines containing the selected agent id, selected
  record id, or known peer id.
- Keep an "all logs" mode available so filtering does not hide global runtime
  issues.
- Show a clear empty state when no matching log lines are present.
- Do not add new backend log storage or mutate existing gateway log behavior.
- Preserve existing bounded log-buffer behavior.
- Add frontend helper tests for correlation matching if existing test structure
  supports it.

## Acceptance criteria

- Selecting an agent makes relevant live logs easier to spot.
- Selecting a detail record makes the record id easy to find in logs when
  present.
- Operators can switch back to all logs.
- Empty/error/stale log states remain clear.
- Log filtering does not increase backend load or unbounded frontend memory.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run 'Agent|Organization|Activity|Event' -count=1
cd /Users/aliai/zehn/web/frontend
node --test --experimental-strip-types src/components/agent/organization/organization-state.test.ts
pnpm lint
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 053-organization-log-correlation
```

