# Task 004: Synchronous Delegate Tool

Slug: `004-delegate-tool-sync`

Docs-only allowed: no

## Goal

Expose the target-agent delegation primitive through a synchronous tool that a
configured agent can use to request bounded advice, review, or work from an
allowed peer agent.

## Allowed repos/files

- `pkg/tools/delegate*.go`
- `pkg/tools/delegate*_test.go`
- `pkg/tools/registry.go`
- `pkg/tools/registry_test.go`
- `pkg/agent/agent_init.go`
- `pkg/agent/agent_test.go`
- `pkg/config/config.go`
- `pkg/config/defaults.go`
- `pkg/config/config_test.go`
- `docs/reference/**`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `pkg/tools/subagent.go`
- `pkg/tools/spawn.go`
- `pkg/tools/registry.go`
- `pkg/agent/agent_init.go`
- `pkg/config/defaults.go`
- `pkg/config/config.go`
- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`

## Work

- Add a tool named `delegate_to_agent` for Zehn-private use, or a generic
  `delegate` tool with a Zehn-facing alias if that fits local conventions
  better.
- Required parameters: `agent_id`, `task`.
- Optional parameters: `thread_key`, `priority`, `due`, `artifact_refs`.
- Return the target response to the caller in sync mode.
- Return structured tool errors for missing task, missing agent ID, denied
  target, missing target, and delegated execution failure.
- Register the tool only when its config flag is enabled.

## Acceptance criteria

- Tool schema is explicit and minimal.
- Tool result is useful to the caller and concise to the user.
- Permission behavior matches `subagents.allow_agents`.
- The tool does not depend on Discord or GitHub.
- Existing `spawn` and `subagent` tool tests still pass.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./pkg/tools -run 'Test.*Delegate|TestSpawnTool|TestSubagentTool|TestToolRegistry' -count=1 -v
go test ./pkg/agent -run 'Test.*Delegate|Test.*SharedTools|Test.*Tool' -count=1 -v
go test ./pkg/tools ./pkg/agent ./pkg/config
```
