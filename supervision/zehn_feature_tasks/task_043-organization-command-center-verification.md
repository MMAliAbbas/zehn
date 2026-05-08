# Task 043: Organization Command Center Verification

Slug: `043-organization-command-center-verification`

Docs-only allowed: no

## Goal

Add final verification coverage and operator documentation for the organization
command center.

## Allowed repos/files

- `web/backend/api/**`
- `web/frontend/src/**`
- `docs/reference/**`
- `supervision/**`

## Required reading

- `supervision/ZEHN_AGENT_ORGANIZATION_COMMAND_CENTER_PLAN.md`
- `docs/reference/agent-organization-live-verification.md`
- `web/frontend/src/components/agent/organization/organization-page.tsx`
- `web/frontend/src/components/agent/organization/agent-card.tsx`
- `web/frontend/src/components/agent/organization/detail-panels.tsx`
- `web/frontend/src/hooks/use-gateway-logs.ts`

## Work

- Add or update tests for any new helper functions, status rules, or backend
  read-only fields introduced by tasks 035-042.
- Update operator documentation for the command center behavior.
- Document how to verify card selection, shortcuts, workbench tabs, live logs,
  org activity feed, and failure drilldown.
- Confirm the feature remains read-only.
- Confirm the organization page still works with no records, old failures,
  active delegation, active meeting, and gateway logs.

## Acceptance criteria

- There is a clear manual verification procedure for the new command center.
- Automated verification covers the new risk areas introduced by the upgrade.
- Frontend build passes.
- Relevant backend API tests pass.
- No private deployment-specific source code is introduced.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api ./pkg/agent ./pkg/config -run 'Agent|Organization|Activity|Inbox|Outbox|Meeting|Event|Failure' -count=1
cd /Users/aliai/zehn/web/frontend
pnpm lint
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 043-organization-command-center-verification
```
