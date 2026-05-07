# Task 034: Agent Organization Frontend Decomposition

Slug: `034-agent-organization-frontend-decomposition`

Docs-only allowed: no

## Goal

Split the organization frontend into smaller focused modules without changing
user-visible behavior.

## Allowed repos/files

- `web/frontend/src/api/**`
- `web/frontend/src/components/agent/organization/**`
- `web/frontend/src/i18n/locales/**`
- `docs/reference/**`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `web/frontend/src/components/agent/organization/organization-page.tsx`
- `web/frontend/src/api/agents.ts`
- `web/frontend/package.json`
- `docs/reference/agent-organization-live-verification.md`
- `supervision/ZEHN_AGENT_ORGANIZATION_UI_PLAN.md`

## Work

- Decompose the large organization page into focused files such as page shell,
  agent card, detail sheet, record panels, status/count components, and shared
  formatting helpers.
- Preserve the existing route, API contracts, i18n keys, visual layout, and
  behavior.
- Avoid introducing new dependencies or unrelated redesign.
- Keep component boundaries clear enough that later status and detail work can
  be reviewed in smaller files.
- Build after the split to catch export/import and type issues.

## Acceptance criteria

- `organization-page.tsx` becomes a small page composition module instead of a
  large all-in-one implementation.
- Extracted files have clear responsibilities and no circular imports.
- User-visible behavior remains the same.
- The frontend build passes.
- Documentation remains accurate after the split.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run 'Agent|Organization|Activity' -count=1
cd /Users/aliai/zehn/web/frontend
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 034-agent-organization-frontend-decomposition
```

