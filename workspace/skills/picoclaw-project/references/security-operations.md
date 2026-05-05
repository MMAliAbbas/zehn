# Security And Operations

## Highest-Risk Surfaces

Review these areas first during audits:

- Launcher public binding, CIDR allowlists, setup/login flows, cookies, and local auto-login.
- Gateway PID ownership, restart, health probes, and stale process handling.
- External channel allowlists, group triggers, attachment handling, and media downloads.
- Shell execution, background sessions, PTY, command guards, and isolation.
- File tools, symlink handling, workspace restriction, archive extraction, and skill installation.
- Provider credentials, OAuth flows, secure strings, `.security.yml`, logs, and JSON redaction.
- MCP stdio/http server configuration, env files, custom headers, and tool exposure.
- Cron command scheduling and remote-channel command execution.

## Operational Setup

For a dedicated 16 GB Intel MacBook Pro, prefer a conservative always-on profile:

- Run launcher bound to loopback unless remote dashboard access is explicitly required.
- Enable only necessary external channels first.
- Disable remote shell and command cron unless a trusted internal-only channel needs them.
- Keep workspace and config on local disk with backups.
- Use one or two reliable providers plus fallback rather than many untested model entries.
- Treat Discord/Telegram/Slack as production ingress surfaces once enabled.

## Audit Discipline

Do not infer safety from naming. Trace the actual call path, config default, and test coverage. Separate sandbox test failures from product failures. Record uncertainty explicitly.

