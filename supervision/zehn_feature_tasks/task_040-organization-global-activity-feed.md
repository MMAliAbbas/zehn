# Task 040: Organization Global Activity Feed

Slug: `040-organization-global-activity-feed`

Docs-only allowed: no

## Goal

Add an org-wide activity feed so the page shows what changed recently across
agents without requiring the operator to inspect cards one by one.

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
- `web/frontend/src/api/agents.ts`
- `web/frontend/src/components/agent/organization/organization-page.tsx`

## Work

- Add a compact org-wide recent activity feed to the organization page.
- Prefer deriving feed items from existing organization snapshot data first.
- If backend support is needed, add read-only fields or a read-only endpoint.
- Include delegation, meeting, failure, and recent log event summaries where
  available.
- Keep feed entries concise and clickable/selectable when tied to a known
  agent.

## Acceptance criteria

- The operator can see recent org-level activity without opening each agent.
- Feed entries indicate agent, type, status, and timestamp when available.
- Any backend additions are read-only and covered by tests.
- The page remains usable when there are no records.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run 'Agent|Organization|Activity|Inbox|Outbox|Meeting|Event' -count=1
cd /Users/aliai/zehn/web/frontend
pnpm lint
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 040-organization-global-activity-feed
```
