# Task 033: Agent Organization Text Log Events

Slug: `033-agent-organization-text-log-events`

Docs-only allowed: no

## Goal

Make recent-event enrichment useful with the launcher's current raw gateway log
format while keeping logs secondary to structured records.

## Allowed repos/files

- `web/backend/api/organization.go`
- `web/backend/api/organization_test.go`
- `web/backend/api/agent_activity_test.go`
- `web/backend/api/gateway.go`
- `docs/reference/**`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `web/backend/api/organization.go`
- `web/backend/api/gateway.go`
- `web/backend/api/log.go`
- `web/backend/api/organization_test.go`
- `docs/reference/agent-organization-live-verification.md`
- `supervision/ZEHN_AGENT_ORGANIZATION_UI_PLAN.md`

## Work

- Review the gateway log buffer path and the actual log line format appended by
  the launcher.
- Keep JSON structured log parsing when present.
- Add conservative parsing for raw text log lines that contain clear
  `agent_id=<id>` or similar key/value agent references.
- Do not infer agent matches from arbitrary message text or partial substrings.
- Keep redaction and message length limits for both JSON and text-derived
  events.
- Add tests for raw text matches, unrelated text, partial substring
  non-matches, sensitive value redaction, and malformed lines.
- Update docs to state logs are best-effort secondary evidence.

## Acceptance criteria

- Recent events can be populated from real launcher log lines when they include
  explicit agent key/value fields.
- Log-derived events never override structured current status.
- Raw text parsing is conservative and does not match arbitrary substrings.
- Sensitive values remain redacted in text-derived events.
- Backend tests cover JSON and text log paths.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run 'Agent|Organization|Event|Log|Redact|Text' -count=1
go test ./web/backend/api -run 'Agent|Organization|Event|Log|Redact|Text' -race
operations/audit-zehn-feature-task.sh 033-agent-organization-text-log-events
```

