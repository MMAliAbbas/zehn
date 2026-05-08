# Task 047: Organization Configured Scope Totals

Slug: `047-organization-configured-scope-totals`

Docs-only allowed: no

## Goal

Keep organization command-center totals and global activity feed scoped to the
configured organization agents so stale or unrelated records do not distort the
visible dashboard.

## Allowed repos/files

- `web/backend/api/organization.go`
- `web/backend/api/organization_test.go`
- `web/frontend/src/components/agent/organization/**`
- `docs/reference/**`
- `supervision/**`

## Required reading

- `web/backend/api/organization.go`
- `web/backend/api/organization_test.go`
- `web/frontend/src/components/agent/organization/status-components.tsx`
- `docs/reference/agent-organization-live-verification.md`

## Work

- Ensure organization summary totals count only delegations and meetings that
  reference at least one configured organization agent.
- Ensure global recent activity feed entries are omitted or safely classified
  when they do not reference a configured organization agent.
- Add tests with stale records that reference only unknown agents.
- Keep selected-agent inbox, outbox, and meeting endpoints unchanged unless a
  bug is found during implementation.
- Document how totals should behave after agents are removed or renamed.

## Acceptance criteria

- Dashboard totals can be reconciled with the visible configured organization.
- Unknown-agent historical records do not inflate active, delegation, meeting,
  or failure counts.
- Recent activity entries remain clickable to a visible configured agent.
- Existing tests for configured-agent records continue to pass.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run 'Agent|Organization|Activity|Event|Failure' -count=1
cd /Users/aliai/zehn/web/frontend
pnpm lint
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 047-organization-configured-scope-totals
```
