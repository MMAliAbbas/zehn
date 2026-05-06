---
name: picoclaw-project
description: Use when working in the PicoClaw/Zehn repository on setup, configuration, gateway and channel operation, launcher behavior, tools, MCP, cron, skills, memory, providers, audits, tests, or upstream contribution planning.
---

# PicoClaw Project

## Operating Rules

Start from source evidence. Read `AGENTS.md`, `CONTRIBUTING.md`, and the relevant reference file before changing code or configuration. For Go work, also use the `go125` skill and match existing table-driven test style. Do not make broad runtime changes until you trace the exact long-running flow they affect.

For this workspace, branch names must be `feature/...` or `fix/...` and must not contain `codex`, `agent`, or `ai`.

## Load References By Task

- Repository map and commands: `references/repository-map.md`
- Runtime architecture and turn lifecycle: `references/runtime-architecture.md`
- Configuration, secrets, and setup: `references/configuration-setup.md`
- Gateway, launcher, and Pico WebSocket: `references/gateway-launcher.md`
- External messaging channels: `references/channels.md`
- Tools, MCP, cron, skills, and sandboxing: `references/tools-mcp-cron-skills.md`
- Memory, sessions, context, and Yaad positioning: `references/memory-context.md`
- Security and operations audit focus: `references/security-operations.md`
- Testing and verification expectations: `references/testing-verification.md`
- Upstream contribution strategy: `references/upstream-contribution.md`

## Project Posture

PicoClaw is an always-on, multi-channel Go 1.25 personal assistant runtime with a web launcher, local Pico WebSocket UI, external channel gateways, scheduled automation, tool execution, MCP integration, skills, providers, and session memory. Treat long-running sessions, streaming, steering, gateway lifecycle, and channel delivery as core behavior.

Avoid narrow fixes that add blanket HTTP timeouts, kill idle connections, constrain active sessions, or change process lifecycle semantics without checking the relevant runtime path and tests. PicoClaw expects durable sessions and always-on connections; safety work must preserve that model.

## Delegation And Meetings

The Zehn fork includes opt-in durable target-agent delegation and chaired meeting v1 tools. `delegate_to_agent` is distinct from legacy `spawn`/`subagent`: it runs another configured agent by ID, records a local delegation artifact, can run sync or async, respects `subagents.allow_agents`, and persists terminal summaries to configured durable memory. `start_agent_meeting` is meeting v1: a sponsor asks a chair agent to consult required participants sequentially and produce one consolidated recommendation; it is not live multi-agent debate.

When changing this area, audit the end-to-end path across `pkg/tools/delegate*.go`, `pkg/tools/meeting.go`, `pkg/agent/delegation*.go`, `pkg/agent/meeting*.go`, `pkg/agent/github_artifacts.go`, config defaults, and record stores. Keep GitHub artifacts redacted, status/inbox visibility fail-closed, async executors bounded, publisher lifecycle owned by `AgentLoop`, and Yaad/private-memory metadata generic by default.

## Private Zehn Guidance

Keep Yaad and other Zehn-specific memory work private unless explicitly asked to design an upstream-safe extension point. Prefer private MCP/config integration first. Upstream contributions should fix general PicoClaw problems without pushing Zehn-specific agenda.
