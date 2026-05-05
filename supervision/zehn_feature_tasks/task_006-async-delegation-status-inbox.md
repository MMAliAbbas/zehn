# Task 006: Async Delegation Status And Inbox

Slug: `006-async-delegation-status-inbox`

Docs-only allowed: no

## Goal

Add async delegation support with status and inbox tools so long-running agent
assignments remain durable and inspectable after the original user turn ends.

## Allowed repos/files

- `pkg/agent/delegation*.go`
- `pkg/agent/delegation*_test.go`
- `pkg/tools/delegate*.go`
- `pkg/tools/delegate*_test.go`
- `pkg/agent/agent_init.go`
- `pkg/agent/agent_test.go`
- `pkg/config/config.go`
- `pkg/config/defaults.go`
- `pkg/config/config_test.go`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `pkg/agent/subturn.go`
- `pkg/tools/spawn_status.go`
- `pkg/tools/spawn_status_test.go`
- `pkg/tools/delegate*.go`
- `pkg/agent/delegation*.go`
- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`

## Work

- Add async mode to delegation without blocking the caller until completion.
- Add `delegation_status` for listing and inspecting delegation records.
- Add `delegation_inbox` for target agents to inspect assigned work.
- Ensure async results are persisted even if the parent turn is gone.
- Ensure status visibility is scoped so one agent cannot inspect unrelated
  private records unless configured.

## Acceptance criteria

- Async request returns a delegation ID immediately.
- Status moves through requested/running/completed or failed.
- Target inbox lists only appropriate target-agent work.
- Completed async result is recoverable from the record store.
- Existing sync delegation still works.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./pkg/tools -run 'Test.*Delegation.*Status|Test.*Delegation.*Inbox|Test.*Delegate' -count=1 -v
go test ./pkg/agent -run 'Test.*Delegation.*Async|Test.*Delegation.*Record' -count=1 -v
go test ./pkg/agent ./pkg/tools
```
