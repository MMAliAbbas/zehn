# Task 052: Organization Record Detail Drilldown

Slug: `052-organization-record-detail-drilldown`

Docs-only allowed: no

## Goal

Add a workbench drilldown for one selected activity record so operators can
inspect diagnostic detail without leaving the Organization page.

## Allowed repos/files

- `web/frontend/src/api/agents.ts`
- `web/frontend/src/components/agent/organization/**`
- `web/frontend/src/i18n/locales/**`
- `docs/reference/**`
- `supervision/**`

## Required reading

- `supervision/ZEHN_AGENT_ORGANIZATION_DIAGNOSTICS_PLAN.md`
- `web/frontend/src/api/agents.ts`
- `web/frontend/src/components/agent/organization/agent-detail-content.tsx`
- `web/frontend/src/components/agent/organization/detail-panels.tsx`
- `web/frontend/src/components/agent/organization/record-components.tsx`
- `web/frontend/src/components/agent/organization/organization-state.ts`

## Work

- Add selected-record state for the organization workbench.
- Add a Details action to failure, inbox, outbox, and meeting records where the
  backend says detail inspection is available.
- Fetch the safe detail endpoint for the selected record.
- Render detail sections for identity, reason, request/context summary, result
  summary, memory status, artifact status, participant status, and artifact
  references when available.
- Provide clear loading, not-found, permission-denied, and load-error states.
- Keep the implementation read-only and avoid any retry/replay/run controls.
- Preserve mobile sheet behavior and desktop persistent workbench behavior.

## Acceptance criteria

- Operators can open a selected failed record and see diagnostic details in the
  workbench.
- Non-failure inbox, outbox, and meeting records with details can also be
  inspected safely.
- Detail errors are visible and do not break the rest of the Organization page.
- No mutation controls are introduced.
- Existing record list tabs continue to work.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run 'Agent|Organization|Activity|Failure|Meeting|Detail' -count=1
cd /Users/aliai/zehn/web/frontend
node --test --experimental-strip-types src/components/agent/organization/organization-state.test.ts
pnpm lint
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 052-organization-record-detail-drilldown
```

