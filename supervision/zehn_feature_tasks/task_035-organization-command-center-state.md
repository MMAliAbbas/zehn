# Task 035: Organization Command Center State

Slug: `035-organization-command-center-state`

Docs-only allowed: no

## Goal

Introduce frontend state for selecting an agent and choosing a workbench section
without changing backend behavior.

## Allowed repos/files

- `web/frontend/src/components/agent/organization/**`
- `web/frontend/src/api/**`
- `supervision/ZEHN_AGENT_ORGANIZATION_COMMAND_CENTER_PLAN.md`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `supervision/ZEHN_AGENT_ORGANIZATION_COMMAND_CENTER_PLAN.md`
- `web/frontend/src/components/agent/organization/organization-page.tsx`
- `web/frontend/src/components/agent/organization/agent-card.tsx`
- `web/frontend/src/components/agent/organization/agent-detail-sheet.tsx`
- `web/frontend/src/components/agent/organization/types.ts`

## Work

- Add a small, explicit selected-agent state model to the organization page.
- Preserve the current organization hierarchy rendering.
- Allow agent cards to request selection without opening the details sheet by
  default on desktop.
- Add or update type definitions for the workbench sections.
- Keep the existing details sheet available so later tasks can choose desktop vs
  mobile behavior.
- Do not add backend endpoints or runtime mutations.

## Acceptance criteria

- Selecting an agent can be represented in page state.
- The selected agent can be passed down to cards/branches for active styling in
  later tasks.
- Existing details sheet behavior is not broken.
- The change is read-only and local to the organization frontend.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run "^  -count=1
cd /Users/aliai/zehn/web/frontend
pnpm lint
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 035-organization-command-center-state
```
