# Task 027: Agent Organization Frontend Page

Slug: `027-agent-organization-frontend-page`

Docs-only allowed: no

## Goal

Add a launcher frontend page that renders the configured agent organization
hierarchy with compact activity cards.

## Allowed repos/files

- `web/frontend/src/api/**`
- `web/frontend/src/components/agent/**`
- `web/frontend/src/components/app-sidebar.tsx`
- `web/frontend/src/routes/**`
- `web/frontend/src/i18n/locales/**`
- `web/frontend/src/routeTree.gen.ts`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `web/frontend/src/components/app-sidebar.tsx`
- `web/frontend/src/routes/agent/hub.tsx`
- `web/frontend/src/api/http.ts`
- `web/frontend/src/components/agent/tools/tools-page.tsx`
- `web/frontend/package.json`
- `supervision/ZEHN_AGENT_ORGANIZATION_UI_PLAN.md`

## Work

- Add a typed frontend API client for the organization snapshot endpoint.
- Add an Agent Organization page under the existing Agent navigation group.
- Render a hierarchy from the API payload, with stable ordering and graceful
  empty/error/loading states.
- Show compact cards with agent name, ID, status badge, current activity
  summary, and counts for inbox/outbox/meetings/errors when present.
- Keep the UI operational and dense; avoid marketing-style hero sections.
- Use existing UI primitives and icon library conventions.

## Acceptance criteria

- The page is reachable from the sidebar.
- The page renders explicit hierarchy and flat fallback payloads.
- Long agent names and IDs do not overflow compact cards.
- Loading, error, and empty states are professional and non-disruptive.
- The frontend build passes.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run 'Agent|Organization' -count=1
cd /Users/aliai/zehn/web/frontend
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 027-agent-organization-frontend-page
```

