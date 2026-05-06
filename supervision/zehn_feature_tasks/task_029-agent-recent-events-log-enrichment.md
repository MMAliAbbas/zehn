# Task 029: Agent Recent Events Log Enrichment

Slug: `029-agent-recent-events-log-enrichment`

Docs-only allowed: no

## Goal

Add optional recent-event enrichment from gateway logs while keeping structured
records as the source of truth for current agent status.

## Allowed repos/files

- `web/backend/api/**`
- `web/frontend/src/api/**`
- `web/frontend/src/components/agent/**`
- `web/frontend/src/hooks/**`
- `docs/reference/**`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `web/backend/api/log.go`
- `web/backend/api/gateway.go`
- `web/frontend/src/hooks/use-gateway-logs.ts`
- `web/frontend/src/components/logs/logs-page.tsx`
- `supervision/ZEHN_AGENT_ORGANIZATION_UI_PLAN.md`

## Work

- Add a bounded recent-events field or endpoint for org-page drill-down.
- Derive events from existing launcher/gateway log buffer only as enrichment.
- Filter events conservatively by configured agent ID and known structured log
  fields where possible.
- Never use log-derived data to override structured status from records.
- Redact or omit sensitive content and long raw lines in API responses.
- Add tests for no-log, matching-log, unrelated-log, and redaction behavior.

## Acceptance criteria

- Recent events improve observability without becoming the primary status
  source.
- Log parsing failures cannot break the organization page.
- Responses are bounded and safe for UI display.
- Tests cover event matching and non-matching behavior.
- Documentation states that logs are secondary evidence.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run 'Agent|Organization|Event|Log|Redact' -count=1
go test ./web/backend/api -run 'Agent|Organization|Event|Log|Redact' -race
cd /Users/aliai/zehn/web/frontend
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 029-agent-recent-events-log-enrichment
```

