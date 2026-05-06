# Task 019: Fail Closed Delegation Status

Slug: `019-fail-closed-delegation-status`

Docs-only allowed: no

## Goal

Make delegation status access fail closed when the caller agent identity is
missing, matching the stricter inbox behavior and preserving per-agent record
visibility.

## Allowed repos/files

- `pkg/tools/delegate*.go`
- `pkg/tools/delegate*_test.go`
- `pkg/agent/delegation*.go`
- `pkg/agent/*delegation*_test.go`
- `docs/reference/**`
- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `pkg/tools/delegate_status.go`
- `pkg/tools/delegate.go`
- `pkg/agent/delegation_store.go`
- `pkg/agent/delegation_records.go`
- `docs/reference/agent-delegation-meetings.md`

## Work

- Update the delegation status tool so list and direct-id lookup require a
  caller agent identity by default.
- Preserve legitimate per-agent visibility: parent, target, requesting agent,
  and explicit visible participants should still see their records.
- Keep the delegation inbox tool's existing fail-closed behavior intact.
- If an administrative all-records status view is needed, make it explicit and
  separately authorized rather than relying on missing caller identity.
- Add tests for missing caller identity, allowed caller identity, denied caller
  identity, direct-id lookup, and list filtering.

## Acceptance criteria

- Missing caller identity cannot list or read delegation records.
- Authorized agents can still read the records they should see.
- Unauthorized agents cannot infer private records through list or direct-id
  lookup.
- Existing delegation and async status tests still pass.
- The implementation remains generic PicoClaw code.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./pkg/tools ./pkg/agent -run 'DelegationStatus|DelegationInbox|Delegation|Delegate' -count=1
go test ./pkg/tools ./pkg/agent -run 'DelegationStatus|DelegationInbox|Delegation|Delegate' -race
operations/audit-zehn-feature-task.sh 019-fail-closed-delegation-status
```
