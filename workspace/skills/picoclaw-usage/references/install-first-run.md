# Install And First Run

## Recommended Desktop Path

For macOS or desktop use, prefer the release launcher first:

```bash
picoclaw-launcher
```

Open `http://localhost:18800`, create the launcher password on first run, configure a provider, then start the gateway from the UI. On macOS, Gatekeeper may require System Settings -> Privacy & Security -> Open Anyway for the downloaded binary.

## CLI Path

Use CLI setup when the launcher is unavailable or for headless systems:

```bash
picoclaw onboard
picoclaw agent -m "What is 2+2?"
picoclaw agent
picoclaw gateway
```

Default config path is `~/.picoclaw/config.json`; default workspace is `~/.picoclaw/workspace`.

## Portable Or Service-Friendly Paths

Use environment variables for dedicated installs:

```bash
PICOCLAW_HOME=/srv/picoclaw picoclaw gateway
PICOCLAW_CONFIG=/srv/picoclaw/config.json picoclaw gateway
PICOCLAW_HOME=/srv/picoclaw PICOCLAW_CONFIG=/srv/picoclaw/main.json picoclaw gateway
```

`PICOCLAW_CONFIG` points to the exact config file. `PICOCLAW_HOME` controls default data directories such as workspace, sessions, memory, cron jobs, and global config location.

## Docker

Docker first run creates `docker/data/config.json` and workspace, then exits. Edit secrets/config, then start:

```bash
docker compose -f docker/docker-compose.yml --profile launcher up
docker compose -f docker/docker-compose.yml --profile launcher up -d
docker compose -f docker/docker-compose.yml logs -f
docker compose -f docker/docker-compose.yml --profile launcher down
```

For Docker or VM access from host networks, set `PICOCLAW_GATEWAY_HOST=0.0.0.0` or use launcher public mode only after configuring auth and allowlists.

## Source Build

Use source build for development or custom tags:

```bash
make deps
(cd web/frontend && pnpm install --frozen-lockfile)
make build
make build-launcher
make install
```

Prerequisites in docs are Go 1.25+ and Node.js 22+ with pnpm 10.33.0+ for the frontend.

