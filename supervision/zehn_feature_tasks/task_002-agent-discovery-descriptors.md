# Task 002: Agent Discovery Descriptors

Slug: `002-agent-discovery-descriptors`

Docs-only allowed: no

## Goal

Add a small, upstream-clean agent discovery descriptor layer so agents can know
which configured peers exist without exposing Zehn-specific business logic.

## Allowed repos/files

- `pkg/agent/registry.go`
- `pkg/agent/registry_test.go`
- `pkg/agent/instance.go`
- `pkg/agent/definition.go`
- `pkg/agent/definition_test.go`
- `pkg/agent/context.go`
- `pkg/agent/context_test.go`
- `pkg/agent/prompt*.go`
- `pkg/agent/prompt_test.go`
- `docs/architecture/**`
- `docs/reference/**`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `pkg/agent/registry.go`
- `pkg/agent/instance.go`
- `pkg/agent/definition.go`
- `pkg/agent/context.go`
- `pkg/agent/prompt.go`
- upstream PicoClaw issue `#1934`
- upstream PicoClaw issue `#2148`
- upstream PicoClaw PR `#2158`

## Work

- Add an `AgentDescriptor` model with `id`, `name`, and `description`.
- Add registry methods to list descriptors and fetch one descriptor by ID.
- Prefer existing agent definition metadata where available; otherwise fall
  back to safe config/workspace-derived values.
- Keep the discovery layer generic and independent of private company,
  memory-system, channel, or GitHub workflow assumptions.
- If prompt injection is added in this task, keep it compact and omit it for
  single-agent setups.

## Acceptance criteria

- Multi-agent registries expose stable peer descriptors.
- Single-agent setups avoid unnecessary discovery prompt bloat.
- Descriptor output is deterministic for tests.
- No Zehn-private concepts appear in generic PicoClaw code.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./pkg/agent -run 'TestAgentRegistry|TestAgentDefinition|TestContext|TestPrompt' -count=1
go test ./pkg/agent
```
