# Task 036: Organization Clickable Activity Shortcuts

Slug: `036-organization-clickable-activity-shortcuts`

Docs-only allowed: no

## Goal

Make inbox, outbox, meetings, and errors on agent cards usable shortcuts into
the selected agent workbench.

## Allowed repos/files

- `web/frontend/src/components/agent/organization/**`
- `web/frontend/src/i18n/locales/**`
- `supervision/ZEHN_AGENT_ORGANIZATION_COMMAND_CENTER_PLAN.md`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `supervision/ZEHN_AGENT_ORGANIZATION_COMMAND_CENTER_PLAN.md`
- `web/frontend/src/components/agent/organization/agent-card.tsx`
- `web/frontend/src/components/agent/organization/status-components.tsx`
- `web/frontend/src/components/agent/organization/agent-detail-sheet.tsx`
- `web/frontend/src/components/agent/organization/types.ts`

## Work

- Convert nonzero activity count pills into accessible buttons or button-like
  controls.
- Clicking Inbox opens/selects the inbox section for that agent.
- Clicking Outbox opens/selects the outbox section for that agent.
- Clicking Meetings opens/selects the meetings section for that agent.
- Clicking Errors opens/selects a failure-focused section or recent/failure tab
  if the dedicated failure tab is not implemented yet.
- Prevent nested click handling from causing accidental card selection changes.
- Keep visual density close to the current card design.

## Acceptance criteria

- Activity counts are actionable.
- Keyboard users can focus and activate the shortcuts.
- The shortcut target is deterministic and visible.
- Existing cards still render correctly when counts are zero.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run "^  -count=1
cd /Users/aliai/zehn/web/frontend
pnpm lint
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 036-organization-clickable-activity-shortcuts
```
