# Task 049: Organization Diagnostic Summary Model

Slug: `049-organization-diagnostic-summary-model`

Docs-only allowed: no

## Goal

Add a read-only backend diagnostic summary model so organization activity rows
can explain why a delegation or meeting is failed without exposing full raw
record bodies.

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
- `web/frontend/src/api/agents.ts`
- `docs/reference/agent-organization-live-verification.md`

## Work

- Extend the organization activity summary shape with compact diagnostic fields
  derived from existing delegation and meeting records.
- Include a short summary, reason, reason source, severity, current/stale flags,
  and whether detail inspection is available.
- Derive delegation reasons from record errors first, then failed durable-memory
  writes, failed artifact writes, and status fallback.
- Derive meeting reasons from meeting errors first, then failed participant
  turns, failed artifact writes, and status fallback.
- Keep compact summaries bounded and sanitized. Do not expose full task, result,
  notes, prompt, or recommendation bodies in list responses.
- Preserve existing JSON fields so current frontend callers remain compatible.
- Add backend tests for delegation error, memory error, artifact error, meeting
  error, participant failure, stale historical failure, and no-known-reason
  fallback.
- Update frontend TypeScript API types only for the new optional fields.

## Acceptance criteria

- Failed activity records include a useful compact reason when source records
  contain one.
- Non-failed records are not mislabeled as failures.
- Historical failures can be identified as stale when newer current activity is
  available.
- Existing organization, inbox, outbox, meeting, and failure endpoints remain
  read-only and backward compatible.
- Backend tests prove precedence and fallback behavior.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run 'Agent|Organization|Activity|Failure|Meeting' -count=1
cd /Users/aliai/zehn/web/frontend
pnpm lint
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 049-organization-diagnostic-summary-model
```

