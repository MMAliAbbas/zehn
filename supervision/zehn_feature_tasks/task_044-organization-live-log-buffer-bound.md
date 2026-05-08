# Task 044: Organization Live Log Buffer Bound

Slug: `044-organization-live-log-buffer-bound`

Docs-only allowed: no

## Goal

Prevent the organization command center from retaining an unbounded browser-side
gateway log buffer during long-running sessions.

## Allowed repos/files

- `web/frontend/src/hooks/use-gateway-logs.ts`
- `web/frontend/src/hooks/gateway-logs-state.ts`
- `web/frontend/src/hooks/use-gateway-logs.test.ts`
- `web/frontend/src/components/agent/organization/detail-panels.tsx`
- `docs/reference/**`
- `supervision/**`

## Required reading

- `web/frontend/src/hooks/use-gateway-logs.ts`
- `web/frontend/src/hooks/gateway-logs-state.ts`
- `web/frontend/src/hooks/use-gateway-logs.test.ts`
- `web/frontend/src/components/agent/organization/detail-panels.tsx`
- `docs/reference/agent-organization-live-verification.md`

## Work

- Add an explicit maximum retained log-line count for the live gateway log
  buffer used by the organization command center.
- Keep polling offsets correct even when older visible lines are discarded.
- Preserve run-id replacement behavior and clear-log behavior.
- Add focused tests for truncation, offset handling, and run-id replacement.
- Document the retention limit and the operator implication for live log
  review.

## Acceptance criteria

- A long-running organization page cannot grow the in-browser gateway log array
  without bound.
- The latest log lines remain visible after truncation.
- Incremental polling continues from the gateway's total offset, not from the
  truncated visible buffer length.
- Existing live-log filtering still works for selected agents.
- Frontend build and relevant backend tests pass.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run 'Agent|Organization' -count=1
cd /Users/aliai/zehn/web/frontend
node --test --experimental-strip-types src/hooks/use-gateway-logs.test.ts
pnpm lint
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 044-organization-live-log-buffer-bound
```
