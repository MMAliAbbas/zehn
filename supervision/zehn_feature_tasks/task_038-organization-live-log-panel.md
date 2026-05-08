# Task 038: Organization Live Log Panel

Slug: `038-organization-live-log-panel`

Docs-only allowed: no

## Goal

Embed a live gateway log panel into the organization workbench using the
existing incremental gateway logs API.

## Allowed repos/files

- `web/frontend/src/components/agent/organization/**`
- `web/frontend/src/components/logs/**`
- `web/frontend/src/hooks/**`
- `web/frontend/src/api/**`
- `web/frontend/src/i18n/locales/**`
- `supervision/ZEHN_AGENT_ORGANIZATION_COMMAND_CENTER_PLAN.md`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `supervision/ZEHN_AGENT_ORGANIZATION_COMMAND_CENTER_PLAN.md`
- `web/frontend/src/hooks/use-gateway-logs.ts`
- `web/frontend/src/components/logs/logs-panel.tsx`
- `web/frontend/src/api/gateway.ts`
- `web/frontend/src/components/agent/organization/detail-panels.tsx`

## Work

- Add a Live Logs section to the organization workbench.
- Reuse or generalize the existing gateway log polling hook.
- Show incremental gateway logs without requiring navigation to the Logs page.
- Preserve auto-scroll behavior only when the user is near the bottom.
- Handle gateway stopped/stale/error states without breaking the page.
- Do not clear logs from the organization workbench in this task.

## Acceptance criteria

- Live logs are visible from the organization workbench.
- The implementation reuses the existing `/api/gateway/logs` polling behavior.
- Logs do not force-scroll while the user is reading older lines.
- The standalone Logs page still works.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run "^  -count=1
cd /Users/aliai/zehn/web/frontend
pnpm lint
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 038-organization-live-log-panel
```
