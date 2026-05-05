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

## Private Zehn Guidance

Keep Yaad and other Zehn-specific memory work private unless explicitly asked to design an upstream-safe extension point. Prefer private MCP/config integration first. Upstream contributions should fix general PicoClaw problems without pushing Zehn-specific agenda.

