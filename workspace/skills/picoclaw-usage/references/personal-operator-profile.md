# Personal Operating Profile

## Dedicated Intel MacBook Pro, 16 GB RAM

Use the Mac as a local always-on control plane, not a public server by default.

Recommended baseline:

- Launch with `picoclaw-launcher` on loopback.
- Keep gateway on `127.0.0.1` unless a webhook channel requires public ingress.
- Use Telegram, Discord, Slack, or Pico first because they do not require custom public webhooks for basic use.
- Add only necessary channels; allowlist the user's IDs before enabling group/server access.
- Keep workspace on local disk and include it in backups.
- Configure one primary paid provider and one fallback before adding many experimental models.
- Use Ollama/LM Studio only for lightweight local tasks unless model quality is acceptable.
- Disable or restrict remote shell and command cron until trust boundaries are clear.

## Good Initial Capability Stack

Start small:

1. Pico Web UI for local chat and debugging.
2. One high-quality provider model.
3. Web search/fetch.
4. File tools restricted to workspace.
5. One messaging channel with `allow_from`.
6. MCP only for specific trusted systems.
7. Cron reminders before command-running cron.
8. Yaad as a private MCP or private context source after basic operation is stable.

## SaaS Company Use

For running company work, split responsibilities through sessions, routing, and workspace instructions rather than many risky channels at once. Use `AGENT.md`, `USER.md`, `IDENTITY.md`, `SOUL.md`, `memory/MEMORY.md`, and project-specific skills to define operating behavior.

Keep credentials in `.security.yml`, environment files, or external secret storage. Never put production tokens in shared config examples or upstream PRs.

