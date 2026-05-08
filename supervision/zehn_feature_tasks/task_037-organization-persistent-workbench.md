# Task 037: Organization Persistent Workbench

Slug: `037-organization-persistent-workbench`

Docs-only allowed: no

## Goal

Add a persistent desktop workbench for the selected agent while preserving the
mobile sheet experience.

## Allowed repos/files

- `web/frontend/src/components/agent/organization/**`
- `web/frontend/src/i18n/locales/**`
- `supervision/ZEHN_AGENT_ORGANIZATION_COMMAND_CENTER_PLAN.md`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `supervision/ZEHN_AGENT_ORGANIZATION_COMMAND_CENTER_PLAN.md`
- `web/frontend/src/components/agent/organization/organization-page.tsx`
- `web/frontend/src/components/agent/organization/agent-detail-sheet.tsx`
- `web/frontend/src/components/agent/organization/detail-panels.tsx`
- `web/frontend/src/components/ui/sheet.tsx`

## Work

- Create a desktop workbench panel that reuses the existing overview, inbox,
  outbox, meetings, and recent panels.
- Keep the organization canvas visible while the workbench is open.
- On desktop, selecting an agent should update the workbench instead of hiding
  context behind a modal-only flow.
- On mobile/narrow layouts, keep the sheet as the primary detail view.
- Keep the page operational and dense; avoid large decorative cards or hero
  treatment.
- Avoid duplicating query logic more than necessary.

## Acceptance criteria

- Desktop users can inspect an agent without losing the org view.
- Mobile users retain a usable sheet-based detail view.
- Existing detail tabs still work.
- No backend or runtime mutation is introduced.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run 'Agent|Organization|Activity|Inbox|Outbox|Meeting|Event' -count=1
cd /Users/aliai/zehn/web/frontend
pnpm lint
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 037-organization-persistent-workbench
```
