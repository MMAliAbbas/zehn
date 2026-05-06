# Task 028: Agent Activity Drilldown Frontend

Slug: `028-agent-activity-drilldown-frontend`

Docs-only allowed: no

## Goal

Add an agent detail drawer or panel for inbox, outbox, meetings, and recent
activity from the organization page.

## Allowed repos/files

- `web/frontend/src/api/**`
- `web/frontend/src/components/agent/**`
- `web/frontend/src/components/ui/**`
- `web/frontend/src/i18n/locales/**`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `web/frontend/src/components/agent/tools/tools-page.tsx`
- `web/frontend/src/components/ui/sheet.tsx`
- `web/frontend/src/components/ui/dialog.tsx`
- `web/frontend/src/api/http.ts`
- `supervision/ZEHN_AGENT_ORGANIZATION_UI_PLAN.md`

## Work

- Add frontend API functions for agent inbox, outbox, and meeting drill-down
  endpoints.
- Add a card action that opens an agent detail drawer or side panel.
- Provide tabs or segmented controls for Overview, Inbox, Outbox, Meetings, and
  Recent Events.
- Summarize records with status, title/task summary, peer agent, timestamps, and
  artifact/memory status where available.
- Avoid exposing long private content in the first view; use concise summaries.
- Add accessible loading, empty, and error states per tab.

## Acceptance criteria

- Agent cards can open a detail view without navigating away from the org page.
- Inbox/outbox/meeting data is visibly separated and labeled.
- The detail view remains usable on narrow screens.
- Failed or blocked records are visually distinguishable without dominating the
  page.
- The frontend build passes.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run 'Inbox|Outbox|Meeting|Agent' -count=1
cd /Users/aliai/zehn/web/frontend
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 028-agent-activity-drilldown-frontend
```

