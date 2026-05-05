# Gateway, Launcher, And Pico

## CLI Gateway

`picoclaw gateway` is implemented under `cmd/picoclaw/internal/gateway`. It supports debug, no-truncate, allow-empty, and host flags. The web launcher starts the gateway with `gateway -E` so the UI can come up before all credentials are configured.

## Imported Channels

`pkg/gateway/gateway.go` side-effect imports channel implementations: DingTalk, Discord, Feishu, IRC, LINE, MaixCam, OneBot, Pico, QQ, Slack, Teams webhook, Telegram, VK, WeCom, Weixin, WhatsApp, and WhatsApp native.

## Web Launcher

`web/backend/main.go` defaults to `127.0.0.1:18800`. Public binding, host, browser launch, and debug behavior are controlled by flags/config/env. Loopback can use one-shot local auto-login; dashboard auth uses an HttpOnly session cookie.

Middleware enforces CIDR allowlists, treats loopback specially, protects dashboard routes, and returns `401` for unauthorized `/pico/ws`.

`web/backend/api/gateway.go` starts or attaches to gateway processes, validates PID files through process inspection and health checks, caches the Pico token for proxying, and computes config signatures for restart-required status.

## Pico WebSocket

`pkg/channels/pico` exposes `/pico/ws`. Auth can use `Authorization: Bearer`, `Sec-WebSocket-Protocol: token.<token>`, or query tokens when allowed. It broadcasts message create/update/delete, media, typing, errors, placeholders, and tool-feedback events by session.

Pico supports server mode and client mode. `pico_client` dials a remote Pico WebSocket and reconnects. Treat WebSocket longevity as intentional runtime behavior, not a leak.

## Frontend

`web/frontend` uses React 19, Vite, TanStack Router, Jotai, and React Query. Pages cover chat, models, credentials, OAuth, config, raw config, channels, skills, tools, logs, login, and setup. Chat connects to `/pico/ws?session_id=...`, reconnects while the gateway runs, hydrates session history, and handles streamed updates.

