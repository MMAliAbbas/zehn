# Task 051: Organization Failure Reason Frontend

Slug: `051-organization-failure-reason-frontend`

Docs-only allowed: no

## Goal

Render diagnostic reasons in the Organization page so failed cards and failure
lists explain what went wrong instead of showing only a failed status.

## Allowed repos/files

- `web/frontend/src/api/agents.ts`
- `web/frontend/src/components/agent/organization/**`
- `web/frontend/src/i18n/locales/**`
- `docs/reference/**`
- `supervision/**`

## Required reading

- `supervision/ZEHN_AGENT_ORGANIZATION_DIAGNOSTICS_PLAN.md`
- `web/frontend/src/api/agents.ts`
- `web/frontend/src/components/agent/organization/agent-card.tsx`
- `web/frontend/src/components/agent/organization/detail-panels.tsx`
- `web/frontend/src/components/agent/organization/formatting.ts`
- `web/frontend/src/components/agent/organization/record-components.tsx`

## Work

- Show a concise failure reason on agent cards when the current activity is a
  failure and a diagnostic reason is available.
- In the Failures tab, show reason, reason source, severity, and current/stale
  status for each failed record.
- Preserve timestamps, peer, role, status, and artifact references already shown.
- Add empty/fallback copy for failures that have no known reason.
- Keep historical failures visually muted when newer current activity exists.
- Avoid oversized cards or wrapping that breaks the command-center layout.
- Add or update frontend helper tests for formatting current/stale diagnostics
  if existing test structure supports it.

## Acceptance criteria

- A failed agent card shows a useful reason when one exists.
- The Failures tab explains why each visible failure failed.
- Current failures and historical failures are visually distinguishable.
- Records without a known reason remain understandable and do not look broken.
- Layout remains stable across the existing command-center card grid and
  workbench.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run 'Agent|Organization|Activity|Failure|Meeting' -count=1
cd /Users/aliai/zehn/web/frontend
node --test --experimental-strip-types src/components/agent/organization/organization-state.test.ts
pnpm lint
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 051-organization-failure-reason-frontend
```

