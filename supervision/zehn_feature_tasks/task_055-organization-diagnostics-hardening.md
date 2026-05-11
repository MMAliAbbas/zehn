# Task 055: Organization Diagnostics Hardening

Slug: `055-organization-diagnostics-hardening`

Docs-only allowed: no

## Goal

Harden the Organization diagnostics implementation after review so failure
detail remains safe, selected-record log filtering is semantically clear, and
per-agent activity endpoints avoid unnecessary full-snapshot rebuild work.

## Allowed repos/files

- `web/backend/api/organization.go`
- `web/backend/api/organization_test.go`
- `web/frontend/src/api/agents.ts`
- `web/frontend/src/components/agent/organization/**`
- `web/frontend/src/i18n/locales/**`
- `docs/reference/**`
- `supervision/**`

## Required reading

- `supervision/ZEHN_AGENT_ORGANIZATION_DIAGNOSTICS_PLAN.md`
- `web/backend/api/organization.go`
- `web/backend/api/organization_test.go`
- `web/frontend/src/api/agents.ts`
- `web/frontend/src/components/agent/organization/detail-panels.tsx`
- `web/frontend/src/components/agent/organization/organization-log-correlation.ts`
- `web/frontend/src/components/agent/organization/organization-state.test.ts`
- `docs/reference/agent-organization-live-verification.md`

## Work

- Replace or relabel detail fields that currently look like diagnostic
  summaries but are actually bounded raw excerpts from task, result, meeting,
  chair, participant, note, or recommendation text.
- Prefer structured diagnostic summaries from status, mode, priority,
  approval-required flag, participant status, memory status, artifact status,
  and known error sources before using any raw excerpt.
- If any bounded raw excerpt remains, expose it with an explicit field name or
  UI label that says it is an excerpt, not a derived diagnosis.
- Preserve redaction and strict display bounds for every detail string.
- Tighten selected-record live-log correlation so the "Selected Record" mode
  primarily matches the selected record id and known peer ids. If selected-agent
  context is still included, rename the UI copy to make that scope explicit.
- Add frontend tests covering strict selected-record correlation semantics and
  empty-state copy.
- Refactor per-agent inbox, outbox, meetings, failures, and detail handlers so
  current/stale annotation does not require rebuilding the full organization
  snapshot when a cheaper direct derivation from the same record stores is
  practical.
- Keep all endpoints read-only. Do not mutate records, config, memory, channels,
  GitHub artifacts, or external systems.

## Acceptance criteria

- Detail responses no longer present raw task/result/meeting text as if it were
  a derived diagnostic summary.
- Any remaining raw excerpt is clearly labeled as an excerpt and remains
  bounded/redacted.
- Selected-record log filtering behavior matches its UI label and tests.
- Per-agent activity endpoints avoid avoidable full organization snapshot
  rebuilds while preserving current/stale behavior.
- Existing diagnostics, visibility, stale/current, and hidden-record tests still
  pass.
- The feature remains read-only and safe for always-on gateway operation.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run 'Agent|Organization|Activity|Failure|Meeting|Detail|Event' -count=1
cd /Users/aliai/zehn/web/frontend
node --test --experimental-strip-types src/components/agent/organization/organization-state.test.ts
node --test --experimental-strip-types src/components/agent/organization/failure-records.test.ts
pnpm lint
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 055-organization-diagnostics-hardening
```

