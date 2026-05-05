# Tools, Automation, MCP, And Skills

## Tool Posture

PicoClaw has powerful tools. Before connecting remote chat channels, review:

```json
{
  "tools": {
    "exec": {
      "enabled": true,
      "allow_remote": false,
      "enable_deny_patterns": true
    },
    "cron": {
      "enabled": true,
      "allow_command": false,
      "exec_timeout_minutes": 5
    },
    "web": {
      "enabled": true
    }
  }
}
```

The exec guard blocks many dangerous direct commands, but it is not a complete sandbox for untrusted build pipelines. Build tools can spawn child processes after the top-level command is allowed.

## Web

DuckDuckGo is enabled by default. Brave, Tavily, Perplexity, Baidu, GLM Search, and SearXNG can be configured. `web_fetch` has a default 10 MB fetch limit. Use `private_host_whitelist` before allowing fetches to private/internal hosts.

## Cron

Cron jobs live in `<workspace>/cron/jobs.json`.

CLI examples:

```bash
picoclaw cron add --name "Daily summary" --message "Summarize today's logs" --cron "0 18 * * *"
picoclaw cron add --name "Ping" --message "heartbeat" --every 300 --deliver
picoclaw cron list
picoclaw cron disable <job-id>
picoclaw cron remove <job-id>
```

The agent-facing cron tool supports one-time, interval, and cron-expression jobs. Command jobs run through the exec tool and should stay internal-only.

## MCP

The MCP CLI edits config; the gateway starts configured servers.

```bash
picoclaw mcp add filesystem -- npx -y @modelcontextprotocol/server-filesystem /tmp
picoclaw mcp add context7 --transport http https://mcp.context7.com/mcp
picoclaw mcp add github --env-file .env.github -- npx -y @modelcontextprotocol/server-github
picoclaw mcp list --status
picoclaw mcp show filesystem
picoclaw mcp test filesystem
```

Use `--deferred` for large MCP servers so tools are discoverable on demand instead of always loaded into context.

## Skills

Skills load from:

1. `<workspace>/skills`
2. `~/.picoclaw/skills`
3. builtin embedded skills

CLI:

```bash
picoclaw skills search "web scraping"
picoclaw skills install <skill-name>
picoclaw skills list
```

Chat commands:

```text
/list skills
/list mcp
/show mcp github
/use <skill> <message>
/use <skill>
/use clear
/btw <question>
```

Treat remote skill installation as trusted-code installation.

