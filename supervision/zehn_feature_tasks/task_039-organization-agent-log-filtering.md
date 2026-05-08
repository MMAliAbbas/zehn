# Task 039: Organization Agent Log Filtering

Slug: `039-organization-agent-log-filtering`

Docs-only allowed: no

## Goal

Make live logs useful for a selected agent by filtering or highlighting lines
that reference that agent.

## Allowed repos/files

- `web/frontend/src/components/agent/organization/**`
- `web/frontend/src/components/logs/**`
- `web/frontend/src/lib/**`
- `web/frontend/src/i18n/locales/**`
- `supervision/ZEHN_AGENT_ORGANIZATION_COMMAND_CENTER_PLAN.md`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `supervision/ZEHN_AGENT_ORGANIZATION_COMMAND_CENTER_PLAN.md`
- `web/frontend/src/components/agent/organization/detail-panels.tsx`
- `web/frontend/src/components/logs/ansi-log-line.tsx`
- `web/frontend/src/lib/ansi-log.ts`
- `web/backend/api/organization.go`

## Work

- Add selected-agent filtering and/or highlighting for live log lines.
- Include common agent reference fields such as `agent_id`, `target_agent_id`,
  `parent_agent_id`, `requester_id`, `sponsor_agent_id`, and `chair_agent_id`.
- Keep an option to view all logs for context.
- Preserve ANSI formatting as much as practical.
- Avoid parsing secrets or exposing sensitive values.

## Acceptance criteria

- Selected-agent log references are easier to find.
- The user can switch between all logs and selected-agent-related logs.
- Filtering does not remove logs permanently or affect the standalone Logs page.
- Behavior is deterministic for JSON logs and key-value text logs.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run "^  -count=1
cd /Users/aliai/zehn/web/frontend
pnpm lint
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 039-organization-agent-log-filtering
```
