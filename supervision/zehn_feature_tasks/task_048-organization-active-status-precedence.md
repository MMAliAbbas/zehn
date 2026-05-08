# Task 048: Organization Active Status Precedence

Slug: `048-organization-active-status-precedence`

Docs-only allowed: no

## Goal

Make the organization command center show "what is happening now" by keeping
active work and active meetings visible even when newer completed records exist
for the same agent.

## Allowed repos/files

- `web/backend/api/organization.go`
- `web/backend/api/organization_test.go`
- `web/frontend/src/components/agent/organization/**`
- `docs/reference/**`
- `supervision/**`

## Required reading

- `web/backend/api/organization.go`
- `web/backend/api/organization_test.go`
- `web/frontend/src/components/agent/organization/formatting.ts`
- `web/frontend/src/components/agent/organization/status-components.tsx`
- `docs/reference/agent-organization-live-verification.md`

## Work

- Review current activity selection rules for active, failed, and completed
  records.
- Adjust status precedence so a selected agent with any active visible work or
  meeting is not shown as idle because of a newer completed record.
- Preserve separate last-failure metadata so historical failures stay visible
  without incorrectly overriding active work.
- Add tests covering older active work plus newer completed records, active
  meeting plus newer completed records, and newer failures.
- Document the final status precedence rules.

## Acceptance criteria

- Active delegation target records keep the agent in a working state until the
  active record ends.
- Active meetings keep participants in a meeting state until the meeting ends.
- Completed records do not hide active current work.
- Newer failures still surface clearly according to the documented failure
  policy.
- Existing active and failure tests continue to pass.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run 'Agent|Organization|Activity|Failure|Meeting' -count=1
cd /Users/aliai/zehn/web/frontend
pnpm lint
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 048-organization-active-status-precedence
```
