# Task 022: Generic Memory Artifact Metadata

Slug: `022-generic-memory-artifact-metadata`

Docs-only allowed: no

## Goal

Keep durable memory artifact metadata generic by default, with private runtime
identity supplied through narrow configuration or adapter fields instead of
hard-coded product labels in generic package code.

## Allowed repos/files

- `pkg/agent/delegation_memory.go`
- `pkg/agent/delegation_memory_test.go`
- `pkg/agent/delegation*.go`
- `pkg/agent/*delegation*_test.go`
- `pkg/config/config.go`
- `pkg/config/defaults.go`
- `pkg/config/*_test.go`
- `docs/reference/**`
- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `pkg/agent/delegation_memory.go`
- `pkg/agent/delegation_memory_test.go`
- `pkg/config/config.go`
- `pkg/config/defaults.go`
- `docs/reference/agent-delegation-meetings.md`

## Work

- Audit durable delegation memory payload metadata for hard-coded private or
  product-specific labels.
- Replace hard-coded metadata with generic defaults or narrow configurable
  metadata values.
- Preserve the existing memory payload shape and useful searchable labels.
- Preserve idempotent terminal-state memory write behavior.
- Add tests proving default metadata is generic and configured metadata can be
  applied without changing core memory semantics.

## Acceptance criteria

- Generic package code does not hard-code private runtime branding in durable
  memory payload metadata.
- Private deployments can still configure recognizable memory source labels.
- Existing strict/unavailable/idempotent memory tests still pass.
- The implementation remains generic PicoClaw code.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./pkg/agent ./pkg/config -run 'DelegationMemory|Memory|Delegation|Config' -count=1
go test ./pkg/agent ./pkg/config -run 'DelegationMemory|Memory|Delegation|Config' -race
operations/audit-zehn-feature-task.sh 022-generic-memory-artifact-metadata
```
