# Task 032: Agent Organization Live Refresh

Slug: `032-agent-organization-live-refresh`

Docs-only allowed: no

## Goal

Refresh the organization page and open activity drill-downs often enough to act
as an operational dashboard without overloading the launcher.

## Allowed repos/files

- `web/frontend/src/api/**`
- `web/frontend/src/components/agent/**`
- `web/frontend/src/i18n/locales/**`
- `docs/reference/**`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `web/frontend/src/components/agent/organization/organization-page.tsx`
- `web/frontend/src/api/agents.ts`
- `web/frontend/package.json`
- `docs/reference/agent-organization-live-verification.md`
- `supervision/ZEHN_AGENT_ORGANIZATION_UI_PLAN.md`

## Work

- Add modest polling for the organization snapshot while the page is mounted.
- Add modest polling for open inbox, outbox, and meeting detail queries while
  their tab is visible.
- Avoid aggressive refresh intervals; prefer a conservative interval that keeps
  the page useful without hammering local APIs.
- Keep loading/error states stable during background refreshes.
- Add code comments only if needed to explain refresh policy.
- Update operator docs with the refresh interval and expectations.

## Acceptance criteria

- The organization page updates without manual browser refresh.
- Drill-down tabs update while open and do not fetch while closed.
- Background refresh does not replace visible content with a full loading state
  on every poll.
- The frontend build passes.
- Documentation states the page is near-live polling, not a guaranteed realtime
  event stream.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run 'Agent|Organization|Activity' -count=1
cd /Users/aliai/zehn/web/frontend
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 032-agent-organization-live-refresh
```

