# Task 031: Agent Organization Current Status Semantics

Slug: `031-agent-organization-current-status-semantics`

Docs-only allowed: no

## Goal

Make agent organization status reflect current operational state without letting
old failed records permanently override newer active work.

## Allowed repos/files

- `web/backend/api/organization.go`
- `web/backend/api/organization_test.go`
- `web/backend/api/agent_activity_test.go`
- `docs/reference/**`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `web/backend/api/organization.go`
- `web/backend/api/organization_test.go`
- `web/backend/api/agent_activity_test.go`
- `docs/reference/agent-organization-live-verification.md`
- `supervision/ZEHN_AGENT_ORGANIZATION_UI_PLAN.md`

## Work

- Review how organization status is derived from delegation and meeting
  records.
- Keep `last_failure` and failure counters visible, but ensure a stale failed
  record does not permanently mask a newer active delegation or meeting.
- Define deterministic precedence between active work, active meetings, recent
  failed records, completed records, and idle state.
- Add tests where an older failed record and a newer running delegation exist
  for the same agent.
- Add tests where a newer failed record should still be visible as the current
  status when no newer active record supersedes it.
- Update operator docs so status badges explain current status versus retained
  failure evidence.

## Acceptance criteria

- Current status matches the most relevant current or newest operational
  record, not simply the highest severity historical record.
- Failed records remain visible through `last_failure` and failure counts.
- Status precedence is deterministic and covered by tests.
- Existing no-leak and read-only behavior remains intact.
- The implementation remains generic launcher functionality.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run 'Agent|Organization|Activity|Status|Failure|Current' -count=1
go test ./web/backend/api -run 'Agent|Organization|Activity|Status|Failure|Current' -race
operations/audit-zehn-feature-task.sh 031-agent-organization-current-status-semantics
```

