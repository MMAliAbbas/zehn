# Task 026: Agent Inbox Outbox API

Slug: `026-agent-inbox-outbox-api`

Docs-only allowed: no

## Goal

Expose read-only launcher APIs for agent inbox, outbox, and meeting drill-down
without leaking unrelated private records.

## Allowed repos/files

- `web/backend/api/**`
- `pkg/agent/delegation_store.go`
- `pkg/agent/meeting_store.go`
- `pkg/tools/delegate_status.go`
- `pkg/tools/delegate_status_test.go`
- `docs/reference/**`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `web/backend/api/router.go`
- `web/backend/api/config.go`
- `pkg/agent/delegation_store.go`
- `pkg/agent/meeting_store.go`
- `pkg/tools/delegate_status.go`
- `supervision/ZEHN_AGENT_ORGANIZATION_UI_PLAN.md`

## Work

- Add read-only endpoints for an agent's inbox, outbox, and meetings.
- Treat inbox as records targeted to the selected agent.
- Treat outbox as records requested by the selected agent.
- Treat meetings as records where the selected agent is sponsor, chair, or
  participant.
- Reuse the same visibility/redaction principles as delegation status tools.
- Support pagination or a conservative limit with stable newest-first sorting.
- Add tests proving unrelated records are not returned for a selected agent.

## Acceptance criteria

- Inbox/outbox endpoints return only records related to the requested agent.
- Meeting endpoint returns only meetings related to the requested agent.
- Responses are summarized enough for UI display and do not expose full private
  internal prompts by default.
- Unknown agent IDs return a clear client error.
- Tests cover visible records, unrelated records, missing stores, and limit
  behavior.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api ./pkg/agent ./pkg/tools -run 'Inbox|Outbox|Delegation|Meeting|Visibility' -count=1
go test ./web/backend/api ./pkg/agent ./pkg/tools -run 'Inbox|Outbox|Delegation|Meeting|Visibility' -race
operations/audit-zehn-feature-task.sh 026-agent-inbox-outbox-api
```

