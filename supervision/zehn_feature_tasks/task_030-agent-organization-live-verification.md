# Task 030: Agent Organization Live Verification

Slug: `030-agent-organization-live-verification`

Docs-only allowed: no

## Goal

Add verification coverage and operator documentation for the agent organization
page, activity summaries, and drill-down APIs.

## Allowed repos/files

- `web/backend/api/**`
- `web/frontend/src/**`
- `docs/reference/**`
- `supervision/**`

## Required reading

- `supervision/ZEHN_AGENT_ORGANIZATION_UI_PLAN.md`
- `web/backend/api/router.go`
- `web/frontend/src/components/app-sidebar.tsx`
- `web/frontend/package.json`
- `docs/reference/agent-delegation-meetings.md`

## Work

- Add or update tests to cover the full read-only org-page flow.
- Document how an operator verifies the page against a running launcher and
  gateway.
- Include staged checks for config-only state, active delegation, active
  meeting, failed record, missing record directory, and recent-event
  enrichment.
- Confirm no endpoint mutates config, records, memory, channels, or external
  artifacts.
- Confirm frontend build and backend tests pass together.

## Acceptance criteria

- There is clear evidence that the page works with and without activity records.
- Operator docs explain what the status badges mean.
- Tests cover the high-risk cases for visibility and stale/missing data.
- The final implementation remains read-only from the launcher page.
- The task does not introduce private deployment-specific source code.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api ./pkg/agent ./pkg/config -run 'Agent|Organization|Activity|Inbox|Outbox|Meeting|Event' -count=1
cd /Users/aliai/zehn/web/frontend
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 030-agent-organization-live-verification
```

