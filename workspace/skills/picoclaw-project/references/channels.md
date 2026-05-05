# Channels

## Channel Manager

`pkg/channels.Manager` initializes enabled channel configs through registered factories, injects media store and owner references, registers webhook/health routes on a dynamic mux, starts per-channel text/media workers, and dispatches bus output.

Startup is partially tolerant: if some enabled channels fail, others can continue. If all enabled channels fail, startup fails. `gateway -E` permits empty or incomplete startup for setup flows.

## Delivery Semantics

Outbound workers enforce per-channel rate limits, split by semantic markers and max length, retry temporary/rate errors, and avoid retrying permanent failures. Placeholder, typing, reaction, and tool-feedback lifecycles are explicit. A TTL janitor cleans stale placeholder records.

## Access Control

In `pkg/channels/base.go`, an empty `allow_from` means open access and logs a warning. `"*"` explicitly allows all. Group behavior is permissive unless `group_trigger.mention_only` or prefixes are configured.

## Notable Channels

- Telegram: long polling, proxy/base URL, command registration, Markdown/HTML modes, group triggers, forum-topic session IDs, photos, voice, audio, docs, placeholders, typing, streaming, media send.
- Discord: gateway bot, mention and group filtering, attachments, tool feedback, reactions, typing, voice join/leave/listen, ASR/TTS hooks, reference expansion.
- Slack: Socket Mode, bot/app tokens, messages, app mentions, slash commands, thread-aware chat IDs, file upload/download, reactions, allowlist before media.
- Pico: local WebSocket UI and remote client bridge.

## Operational Rule

When configuring external channels, explicitly review `allow_from`, group trigger behavior, `exec.allow_remote`, `cron.allow_command`, file tools, and credential storage. Public chat surfaces plus default powerful tools are the highest-risk deployment combination.

