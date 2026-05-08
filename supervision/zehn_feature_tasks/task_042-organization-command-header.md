# Task 042: Organization Command Header

Slug: `042-organization-command-header`

Docs-only allowed: no

## Goal

Upgrade the organization summary into a live command header that communicates
overall system activity and freshness.

## Allowed repos/files

- `web/frontend/src/components/agent/organization/**`
- `web/frontend/src/i18n/locales/**`
- `supervision/ZEHN_AGENT_ORGANIZATION_COMMAND_CENTER_PLAN.md`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `supervision/ZEHN_AGENT_ORGANIZATION_COMMAND_CENTER_PLAN.md`
- `web/frontend/src/components/agent/organization/status-components.tsx`
- `web/frontend/src/components/agent/organization/organization-page.tsx`
- `web/frontend/src/components/agent/organization/constants.ts`

## Work

- Replace or refine the current summary metric strip into a compact command
  header.
- Show active work, delegations, meetings, failures, hierarchy mode, and
  generated/last refresh time.
- Add a subtle live/refresh indicator tied to the organization query state.
- Keep the header dense and operational, not decorative.
- Do not add external health checks in this task unless already available from
  existing frontend state.

## Acceptance criteria

- The top of the organization page immediately communicates system activity.
- Refresh/live state is visible without distracting animation.
- The header remains responsive on narrow screens.
- No backend mutation is introduced.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run "^  -count=1
cd /Users/aliai/zehn/web/frontend
pnpm lint
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 042-organization-command-header
```
