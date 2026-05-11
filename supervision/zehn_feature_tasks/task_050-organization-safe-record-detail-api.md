# Task 050: Organization Safe Record Detail API

Slug: `050-organization-safe-record-detail-api`

Docs-only allowed: no

## Goal

Add a safe read-only detail endpoint for selected organization activity records
so an operator can inspect the reason, context summary, and supporting evidence
behind a failure without opening local JSON files manually.

## Allowed repos/files

- `web/backend/api/organization.go`
- `web/backend/api/organization_test.go`
- `web/frontend/src/api/agents.ts`
- `docs/reference/**`
- `supervision/**`

## Required reading

- `supervision/ZEHN_AGENT_ORGANIZATION_DIAGNOSTICS_PLAN.md`
- `web/backend/api/organization.go`
- `web/backend/api/organization_test.go`
- `pkg/agent/delegation_store.go`
- `pkg/agent/meeting_store.go`
- `web/frontend/src/api/agents.ts`

## Work

- Add a read-only endpoint for a configured agent to inspect one visible
  delegation or meeting record by id.
- Require the selected agent to be related to the record through the same
  visibility rules used by inbox, outbox, meetings, and failures.
- Return a safe detail model containing identity, status, role, peer, timestamps,
  compact reason, request/context summary, result summary, memory status,
  artifact status, participant status when applicable, and artifact references.
- Bound all string fields intended for display. Do not dump full task, result,
  note, recommendation, or transcript content.
- Return clear `404` or `403` style errors for unknown or non-visible records.
- Add backend tests for delegation detail visibility, meeting detail visibility,
  hidden record rejection, bounded summaries, and artifact/memory diagnostics.
- Add TypeScript API types and a fetch helper for the new detail endpoint.

## Acceptance criteria

- A selected agent can fetch details for visible delegation and meeting records.
- A selected agent cannot fetch unrelated private records.
- Details include the diagnostic reason and supporting status fields needed for
  operator inspection.
- Large raw bodies are summarized or truncated, not exposed wholesale.
- Existing list endpoints and organization snapshot behavior do not regress.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run 'Agent|Organization|Activity|Failure|Meeting|Detail' -count=1
cd /Users/aliai/zehn/web/frontend
pnpm lint
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 050-organization-safe-record-detail-api
```

