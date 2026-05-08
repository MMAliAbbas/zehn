# Task 041: Organization Failure Drilldown

Slug: `041-organization-failure-drilldown`

Docs-only allowed: no

## Goal

Make failures understandable and distinguish old/stale failures from current
work.

## Allowed repos/files

- `web/backend/api/**`
- `web/frontend/src/api/**`
- `web/frontend/src/components/agent/organization/**`
- `web/frontend/src/i18n/locales/**`
- `supervision/ZEHN_AGENT_ORGANIZATION_COMMAND_CENTER_PLAN.md`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `supervision/ZEHN_AGENT_ORGANIZATION_COMMAND_CENTER_PLAN.md`
- `web/backend/api/organization.go`
- `web/backend/api/organization_test.go`
- `web/frontend/src/components/agent/organization/detail-panels.tsx`
- `web/frontend/src/components/agent/organization/formatting.ts`

## Work

- Add a failure-focused workbench section or improve the Recent section so
  failure records are easy to inspect.
- Show record id, peer agent, role, status, created/updated/completed time, and
  artifact references when available.
- Indicate when a failure is the last failure but not the current activity.
- Avoid making an old failure visually look like a current running blocker when
  newer successful activity exists.
- Add backend summary fields only if needed, and keep them read-only.

## Acceptance criteria

- Old failures are clearly understandable as old failures.
- Current failures remain visible and high-signal.
- The user can jump from an Errors shortcut to failure details.
- Tests cover stale failure semantics if backend status logic changes.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run 'Agent|Organization|Activity|Failure|Event' -count=1
cd /Users/aliai/zehn/web/frontend
pnpm lint
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 041-organization-failure-drilldown
```
