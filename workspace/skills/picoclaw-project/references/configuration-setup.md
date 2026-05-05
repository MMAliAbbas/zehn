# Configuration And Setup

## Paths

`PICOCLAW_HOME` controls the home directory; otherwise PicoClaw uses `~/.picoclaw`. `PICOCLAW_CONFIG` overrides the full config path. The default workspace is `<PICOCLAW_HOME>/workspace`.

Configuration samples live in `config/`, especially `config/config.example.json`. The launcher stores its own settings in `launcher-config.json`.

## Loading And Saving

`pkg/config.LoadConfig` loads JSON, migrates old versions to v3 with backups, merges `.security.yml`, applies environment expansion, normalizes gateway host values, expands multi-key model configs, initializes channel configs, and validates models.

`SaveConfig` writes `config.json` and `.security.yml` with mode `0600`. Secure values marshal to JSON as `[NOT_HERE]`; secret material is preserved or encrypted through YAML-backed secure fields.

## Secure Values

`SecureString` and `SecureStrings` support raw values, `enc://` values, and `file://` references. Credential sealing may require passphrases or SSH key material depending on configuration.

## Defaults To Review

Default tools are powerful: shell execution, cron, file read/write/edit/append/list, web, web fetch, skills, spawn/subturn, and send-file are enabled by default. Hardware and MCP are disabled by default. `exec.allow_remote` and `cron.allow_command` default true; tighten these for public or semi-public channels.

Isolation defaults to disabled while still restricting filesystem tools to the workspace when configured. Review `tools`, `isolation`, `gateway`, channel allowlists, provider credentials, and launcher public binding before any always-on deployment.

