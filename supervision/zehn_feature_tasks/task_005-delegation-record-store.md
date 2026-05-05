# Task 005: Delegation Record Store

Slug: `005-delegation-record-store`

Docs-only allowed: no

## Goal

Add a durable local delegation record store so delegation requests and results
survive beyond the active model turn before Yaad or GitHub integration is
enabled.

## Allowed repos/files

- `pkg/agent/delegation*.go`
- `pkg/agent/delegation*_test.go`
- `pkg/config/config.go`
- `pkg/config/defaults.go`
- `pkg/config/config_test.go`
- `pkg/session/**`
- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `pkg/session/jsonl_backend.go`
- `pkg/session/manager.go`
- `pkg/agent/pipeline_finalize.go`
- `pkg/agent/delegation*.go`
- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`

## Work

- Define a delegation record schema with request, status, parent, target,
  timestamps, result, error, and artifact references.
- Store records under the configured workspace or PicoClaw home using stable
  filenames.
- Write records atomically.
- Update status on request, running, completion, failure, and cancellation.
- Keep the store generic and independent of Yaad.

## Acceptance criteria

- Record writes are atomic.
- Records are readable after process restart.
- Failed delegations leave useful error evidence.
- Store tests do not require network access.
- No secrets are written to records.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./pkg/agent -run 'Test.*Delegation.*Store|Test.*Delegation.*Record' -count=1 -v
go test ./pkg/agent ./pkg/session
```
