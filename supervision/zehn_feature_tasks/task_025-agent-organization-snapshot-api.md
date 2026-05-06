# Task 025: Agent Organization Snapshot API

Slug: `025-agent-organization-snapshot-api`

Docs-only allowed: no

## Goal

Expose a read-only launcher API that returns configured agent hierarchy plus
structured activity summary.

## Allowed repos/files

- `web/backend/api/**`
- `pkg/agent/delegation_store.go`
- `pkg/agent/meeting_store.go`
- `pkg/config/config.go`
- `pkg/config/*organization*_test.go`
- `docs/reference/**`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `web/backend/api/router.go`
- `web/backend/api/config.go`
- `web/backend/api/gateway.go`
- `pkg/agent/delegation_store.go`
- `pkg/agent/meeting_store.go`
- `pkg/config/config.go`
- `supervision/ZEHN_AGENT_ORGANIZATION_UI_PLAN.md`

## Work

- Register a read-only endpoint such as `GET /api/agents/organization`.
- Load configured agents and optional organization metadata from the launcher
  config path.
- Read delegation and meeting records from the configured workspace-backed
  record directories.
- Return a normalized tree/snapshot payload suitable for a frontend org chart.
- Derive status from structured records first: active delegation target,
  active delegation requester, active meeting participant/chair, failed recent
  record, or idle.
- Do not parse logs for primary status in this task.
- Add handler tests for empty hierarchy, explicit hierarchy, active delegation,
  active meeting, failed record, and malformed config.

## Acceptance criteria

- The endpoint is read-only and requires normal launcher dashboard auth.
- Missing record directories return empty activity instead of errors.
- The API does not expose raw internal prompts or full response content by
  default.
- Status derivation is deterministic and covered by tests.
- The implementation remains generic launcher functionality.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api ./pkg/agent ./pkg/config -run 'Agent|Organization|Activity|Delegation|Meeting' -count=1
go test ./web/backend/api ./pkg/agent ./pkg/config -run 'Agent|Organization|Activity|Delegation|Meeting' -race
operations/audit-zehn-feature-task.sh 025-agent-organization-snapshot-api
```

