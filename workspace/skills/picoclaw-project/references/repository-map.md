# Repository Map

## Layout

- `cmd/picoclaw`: main Cobra CLI. Subcommands include `onboard`, `agent`, `auth`, `gateway`, `status`, `cron`, `mcp`, `migrate`, `skills`, `model`, `update`, and `version`.
- `cmd/picoclaw/internal`: command helpers and subcommand implementations.
- `cmd/picoclaw-launcher-tui`: terminal launcher entrypoint.
- `pkg/agent`: core turn pipeline, context assembly, provider calls, tool execution, steering, finalization, hooks, and subturns.
- `pkg/bus`: internal message bus and streaming delegate.
- `pkg/channels`: channel manager, common channel contracts, delivery workers, placeholders, media, and channel implementations.
- `pkg/gateway`: gateway process startup, reload, services, PID ownership, cron, health, heartbeat, media, and channel wiring.
- `pkg/config`: config load/save, migrations, env expansion, secure strings, defaults, channel decoding, validation.
- `pkg/tools`: built-in tool registry and implementations.
- `pkg/mcp`: MCP server manager and transports.
- `pkg/cron`: persisted scheduled job service.
- `pkg/memory`, `pkg/session`, `pkg/state`: workspace prompt memory, session stores, and state support.
- `pkg/providers`: OpenAI-compatible, Gemini, Anthropic, Azure, Bedrock, OAuth, CLI, local, and fallback behavior.
- `web/backend`: Go launcher HTTP API, dashboard auth, gateway process manager, middleware, launcher config.
- `web/frontend`: React 19, Vite, TanStack Router, Jotai, React Query launcher UI.
- `docs`, `examples`, `config`, `docker`, `scripts`, `assets`: docs, samples, deployment files, helper scripts, and static assets.

## Commands

Use Makefile targets because they set project build tags, cache paths, and toolchain behavior.

- `make generate`: run code generation.
- `make build`: generate and build `build/picoclaw`.
- `make run`: build and run the local CLI.
- `make test`: run Go tests, including launcher backend tests.
- `make check`: dependency checks, formatting, vet, tests, and docs lint.
- `make lint` / `make fix`: run or fix Go lint checks.
- `make build-launcher`: build the web launcher binary through `web/Makefile`.
- `cd web/frontend && pnpm dev`: run Vite frontend development server.
- `cd web/frontend && pnpm build`: type-check and build frontend.
- `cd web/frontend && pnpm lint` / `pnpm format`: frontend lint and formatting.

## Build Details

The root Makefile uses `GO_BUILD_TAGS=goolm,stdjson`, repo-local Go caches, and `GOTOOLCHAIN=local`. Launcher backend builds use CGO on Darwin. Direct `go test ./...` can differ from `make test`; prefer Make targets for repository-level claims.

