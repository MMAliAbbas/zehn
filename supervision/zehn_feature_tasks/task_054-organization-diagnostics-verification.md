# Task 054: Organization Diagnostics Verification

Slug: `054-organization-diagnostics-verification`

Docs-only allowed: no

## Goal

Add verification coverage and operator documentation proving the Organization
diagnostics workflow is useful, read-only, bounded, and safe for always-on
operation.

## Allowed repos/files

- `web/backend/api/**`
- `web/frontend/src/components/agent/organization/**`
- `web/frontend/src/api/**`
- `docs/reference/**`
- `supervision/**`

## Required reading

- `supervision/ZEHN_AGENT_ORGANIZATION_DIAGNOSTICS_PLAN.md`
- `docs/reference/agent-organization-live-verification.md`
- `web/backend/api/organization_test.go`
- `web/frontend/src/components/agent/organization/organization-state.test.ts`
- `web/frontend/src/components/agent/organization/detail-panels.tsx`

## Work

- Add or update a reference document describing how to verify organization
  diagnostics locally without mutating runtime state.
- Cover the full operator path: card status, failure reason, failures tab,
  record detail, live-log correlation, and stale/current failure distinction.
- Add tests or fixture coverage for at least one delegation failure and one
  meeting failure flowing through summary, list, and detail models.
- Confirm bounded display behavior for large task/result/error text.
- Confirm inaccessible or unrelated records do not leak through detail APIs.
- Confirm existing organization command-center behavior still works when no
  diagnostic fields are present.

## Acceptance criteria

- The diagnostics workflow has backend and frontend verification evidence.
- The documentation explains what operators can and cannot infer from the UI.
- The feature remains read-only and safe for always-on gateway operation.
- Old records without diagnostic fields still render safely.
- The final task audit passes.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run 'Agent|Organization|Activity|Failure|Meeting|Detail|Event' -count=1
cd /Users/aliai/zehn/web/frontend
node --test --experimental-strip-types src/components/agent/organization/organization-state.test.ts
pnpm lint
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 054-organization-diagnostics-verification
```

