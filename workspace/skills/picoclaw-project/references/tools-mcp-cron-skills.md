# Tools, MCP, Cron, And Skills

## Tool Registry

`pkg/tools.Registry` separates core and hidden tools, supports hidden TTL promotion, sorts definitions for deterministic prompts and provider cache behavior, validates arguments, injects channel/session context, supports async results, recovers from panics, and normalizes media.

Built-in tools are registered through `pkg/agent/agent_init.go`. They include web, web fetch, file tools, shell execution, cron, media tools, reactions, TTS, image loading, skills search/install, MCP tools, spawn/subturn tools, and hardware tools when enabled.

## Shell Tool

`pkg/tools/shell.go` supports sync and background execution, polling, reading, writing, killing, sending keys, and optional PTY. It has deny-pattern guards for destructive shell commands, command substitution, global installs, package managers, docker run/exec, git push, ssh remote execution, eval/source, and similar hazards.

Filesystem restrictions and subprocess isolation route through `pkg/isolation`. Do not treat these guards as complete security boundaries; configure exposure conservatively.

## Cron

`pkg/tools/cron.go` schedules `at`, `every`, and cron-expression jobs. Jobs persist under `<workspace>/cron/jobs.json` and store channel/chat context plus optional commands. Command jobs are restricted to internal channels unless explicitly confirmed. `pkg/cron.Service` updates status, errors, next run times, and disables or removes one-shot jobs.

## MCP

`pkg/mcp.Manager` supports stdio, SSE, and HTTP MCP servers. Env files are resolved relative to the workspace. HTTP supports custom headers. Partial startup is tolerated if some servers connect; startup fails when every enabled MCP server fails. Stdio subprocesses use isolated command transport and graceful close/terminate/kill behavior.

MCP tools can be registered eagerly or hidden behind discovery using regex and BM25 search helpers.

## Skills

Skills are installed and discovered through `pkg/skills` and tool surfaces. Skill install/import is a security-sensitive path because it can write executable instructions into the workspace. Review origin metadata, archive extraction, overwrite behavior, and registry trust before enabling remote skill installation.

## Delegation And Meeting Tools

The delegation/meeting tools are disabled unless enabled in config:

- `delegate_to_agent`: ask a configured target agent to perform one task. Supports sync and async modes through the runtime delegation runner.
- `delegation_status`: list or inspect delegation records visible to the calling agent. It must have caller identity; missing identity is an error.
- `delegation_inbox`: list delegation work assigned to the calling target agent.
- `start_agent_meeting`: start meeting v1, a private chaired sequential meeting with required participants and one chair recommendation.

Do not overload these tools onto Discord channel routing. Discord can provide human-visible summaries, but internal delegation should use the runtime agent registry, internal sessions, local records, durable memory, and configured artifact writers.

When reviewing tool registration, confirm `pkg/agent/agent_init.go` registers these tools only when explicitly enabled and that `delegate_to_agent` enforces both target existence and `subagents.allow_agents`.
