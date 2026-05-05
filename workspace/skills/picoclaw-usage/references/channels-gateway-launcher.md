# Channels, Gateway, And Launcher

## Launcher

`picoclaw-launcher` serves the browser dashboard on `127.0.0.1:18800` by default. First run creates a dashboard password. Sessions use HttpOnly cookies and reset when the launcher process restarts.

Remote dashboard access requires public binding and CIDR/auth review. Avoid public mode until provider credentials and channel/tool limits are configured.

## Gateway

`picoclaw gateway` runs channel integrations, Pico WebSocket, cron execution, heartbeat, media services, and the turn loop. The launcher starts it with allow-empty behavior so setup can continue before every credential is ready.

Webhook channels share the gateway HTTP host/port, default `127.0.0.1:18790`. Socket/stream channels such as Telegram long polling, Discord gateway, Slack Socket Mode, Feishu/DingTalk/WeCom stream-style connections may not need public HTTP ingress.

## Pico Web UI

Pico is the built-in local WebSocket channel used by the launcher chat UI. Use it first to confirm providers, memory, tools, and gateway health before external channels.

## External Channel Setup Order

1. Create bot/app credentials at the platform.
2. Put tokens in `.security.yml`.
3. Add channel config in `config.json`.
4. Set `allow_from` to known user IDs.
5. For groups/servers, set `group_trigger.mention_only` or prefixes.
6. Start/restart gateway.
7. Test direct message before group/server access.

## Useful Channel Starters

Telegram:

```json
{
  "channel_list": {
    "telegram": {
      "enabled": true,
      "type": "telegram",
      "allow_from": ["YOUR_USER_ID"],
      "use_markdown_v2": false
    }
  }
}
```

Discord:

```json
{
  "channel_list": {
    "discord": {
      "enabled": true,
      "type": "discord",
      "allow_from": ["YOUR_USER_ID"],
      "group_trigger": { "mention_only": true }
    }
  }
}
```

Slack uses Socket Mode with bot and app tokens. Telegram supports command registration and voice/audio/doc media. Discord supports attachments, reactions, typing, and voice hooks. Enable each capability deliberately.

