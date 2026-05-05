---
name: picoclaw-usage
description: Use when setting up, migrating, configuring, operating, troubleshooting, or hardening PicoClaw/Zehn as a personal always-on assistant on another computer, including providers, channels, tools, cron, MCP, skills, memory, and Yaad-private integration.
---

# PicoClaw Usage

## First Response Pattern

Start by identifying the target machine, install method, exposure model, provider choice, channels, and risk level. For Ali's dedicated Intel i7 MacBook Pro with 16 GB RAM, assume a local-first always-on setup unless told otherwise: launcher bound to loopback, gateway local, external channels allowlisted, powerful tools tightened before Discord/Telegram/Slack are enabled.

Do not treat PicoClaw as only a development repo. It is also a runtime to operate: install, configure secrets, choose models, start launcher/gateway, connect channels, tune tools, add MCP servers, create skills, write workspace memory, schedule tasks, and monitor logs.

## Load References By Task

- Install and first run: `references/install-first-run.md`
- Personal operating profile: `references/personal-operator-profile.md`
- Providers, models, and secrets: `references/providers-secrets.md`
- Channels, gateway, and launcher: `references/channels-gateway-launcher.md`
- Tools, cron, MCP, and skills: `references/tools-automation-mcp-skills.md`
- Memory, sessions, heartbeat, and Yaad: `references/memory-heartbeat-yaad.md`
- Operations and troubleshooting: `references/operations-troubleshooting.md`

## Setup Workflow

1. Pick install path: release binary/launcher for normal use, source build only when needed.
2. Run launcher or `picoclaw onboard`.
3. Set `PICOCLAW_HOME` if the install needs a portable or service-friendly data directory.
4. Configure provider models and keep API keys in `.security.yml`.
5. Start with Pico Web UI or CLI chat before external channels.
6. Enable one external channel at a time with `allow_from`.
7. Tighten `tools.exec`, `tools.cron`, file access, and MCP exposure.
8. Add workspace identity/preferences/memory files.
9. Add cron/heartbeat automation only after normal chat and tools work.

## Safety Defaults

Before exposing PicoClaw to external messaging platforms, review `allow_from`, group triggers, `tools.exec.enabled`, `tools.exec.allow_remote`, `tools.cron.allow_command`, MCP servers, file access, and skill installation. Keep Yaad private: use private MCP/config integration first, not upstream-visible config.

