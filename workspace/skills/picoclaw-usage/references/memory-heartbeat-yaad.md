# Memory, Sessions, Heartbeat, And Yaad

## Workspace Files

Default workspace:

```text
~/.picoclaw/workspace/
├── sessions/
├── memory/
├── state/
├── cron/
├── skills/
├── AGENT.md
├── HEARTBEAT.md
├── IDENTITY.md
├── SOUL.md
└── USER.md
```

`AGENT.md`, `SOUL.md`, `USER.md`, and `memory/MEMORY.md` are detected by modification time and reloaded on the next request; gateway restart is not required.

## Usage

- `IDENTITY.md`: who the assistant is.
- `SOUL.md`: tone, temperament, and operating principles.
- `USER.md`: user preferences, constraints, recurring context.
- `AGENT.md`: task behavior, tool rules, escalation style.
- `memory/MEMORY.md`: durable facts and standing knowledge.
- daily memory notes: `memory/YYYYMM/YYYYMMDD.md`.
- `sessions/`: conversation history and summaries.

## Session Sharing

`session.dimensions` controls memory sharing across channel, account, space, chat, topic, and sender. For personal use, default chat-scoped memory is usually safest. For shared groups, include `sender` when users should not share memory.

## Heartbeat

`HEARTBEAT.md` defines periodic tasks. Default interval is 30 minutes, minimum 5.

```json
{
  "heartbeat": {
    "enabled": true,
    "interval": 30
  }
}
```

Environment:

```bash
PICOCLAW_HEARTBEAT_ENABLED=false
PICOCLAW_HEARTBEAT_INTERVAL=60
```

Use heartbeat for light periodic checks and use spawned subturns for long-running work. Heartbeat, spawned work, and main turns inherit the same workspace restriction.

## Yaad

Keep Yaad private. Preferred first path:

1. Run Yaad separately.
2. Expose Yaad through a private MCP server or private local HTTP/stdin integration.
3. Add the MCP entry in local config or `.security.yml`-backed env files.
4. Test through Pico Web UI before any external channel.
5. Avoid upstream-visible Yaad names or required config fields.

Only consider a private Yaad-backed context manager after MCP/private integration proves reliable.

