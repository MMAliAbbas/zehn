# Operations And Troubleshooting

## Daily Commands

```bash
picoclaw status
picoclaw version
picoclaw model
picoclaw agent -m "health check"
picoclaw gateway
picoclaw mcp list --status
picoclaw cron list
```

Set log level:

```json
{
  "gateway": {
    "log_level": "debug"
  }
}
```

Or:

```bash
PICOCLAW_LOG_LEVEL=debug picoclaw gateway
```

## Troubleshooting Order

1. Confirm binary runs: `picoclaw version`.
2. Confirm config path: check `PICOCLAW_HOME` and `PICOCLAW_CONFIG`.
3. Confirm provider works in Pico Web UI or `picoclaw agent -m`.
4. Check secrets mapping in `.security.yml`.
5. Confirm gateway is running and bound to expected host/port.
6. Test one channel in direct message before groups.
7. Check `allow_from` and group trigger settings.
8. Check tool restrictions if web/file/exec/cron actions fail.
9. For MCP, run `picoclaw mcp test <name>` and then restart gateway.
10. Inspect workspace sessions, memory, and cron job files only after runtime checks.

## Common Problems

- Launcher opens but chat fails: gateway may not be running, provider config may be missing, or Pico token/proxy attachment failed.
- External bot sees messages but does not answer: `allow_from`, mention-only trigger, missing message-content intent, or wrong chat/user ID.
- Webhook channel unreachable: gateway host defaults to `127.0.0.1`; public webhook channels need reachable host/proxy.
- MCP server appears configured but no tools: gateway restart may be needed, server may be deferred, or `tools.mcp.enabled` is false.
- Cron job exists but does not execute: cron tool disabled, gateway not running, command job blocked by exec settings, or job is disabled.
- Secret still missing: `.security.yml` key must match model name/channel key exactly and use the right plural/singular field.

## Backup Targets

Back up at least:

- `config.json`
- `.security.yml`
- `launcher-config.json` and launcher auth store
- `workspace/sessions`
- `workspace/memory`
- `workspace/cron`
- `workspace/skills`
- custom MCP env files

