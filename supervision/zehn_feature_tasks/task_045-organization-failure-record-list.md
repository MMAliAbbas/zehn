# Task 045: Organization Failure Record List

Slug: `045-organization-failure-record-list`

Docs-only allowed: no

## Goal

Make the organization command center's failure drilldown show the actual recent
failure records behind an agent's failure count instead of only the current or
last failure summary.

## Allowed repos/files

- `web/backend/api/organization.go`
- `web/backend/api/organization_test.go`
- `web/frontend/src/api/agents.ts`
- `web/frontend/src/components/agent/organization/**`
- `docs/reference/**`
- `supervision/**`

## Required reading

- `web/backend/api/organization.go`
- `web/backend/api/organization_test.go`
- `web/frontend/src/api/agents.ts`
- `web/frontend/src/components/agent/organization/agent-detail-content.tsx`
- `web/frontend/src/components/agent/organization/detail-panels.tsx`
- `docs/reference/agent-organization-live-verification.md`

## Work

- Add a read-only failures endpoint for a selected agent, or extend the existing
  activity API shape in a way that returns recent failed delegation and meeting
  records for that selected agent.
- Keep the response scoped to records visible to the selected agent.
- Update the failures tab to fetch and render the recent failed records.
- Preserve the current and last-failure summary as context, but do not let it
  masquerade as the complete drilldown when more failures exist.
- Add backend and frontend helper tests for multi-failure behavior.

## Acceptance criteria

- When an agent has more than one failed visible record, the failures tab shows
  a recent list instead of only one historical item.
- The failures count and the drilldown content no longer conflict.
- Delegation and meeting failures are both represented.
- Empty and load-error states remain clear.
- The feature remains read-only.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run 'Agent|Organization|Failure|Inbox|Meeting' -count=1
cd /Users/aliai/zehn/web/frontend
pnpm lint
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 045-organization-failure-record-list
```
