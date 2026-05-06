# Zehn Feature Live Verification

Updated: 2026-05-06T06:35:00+05:00

This is the staged local rollout plan for Zehn delegation and meeting behavior.
The default path is deterministic and local-only. It must not send live Discord
messages, must not write live GitHub issues or comments, and must not write live
Yaad memories.

## Current Standing

Implementation status:

- Feature automation tasks `001` through `023` are green.
- Delegation and meeting code audits passed focused normal and race tests.
- The remaining Zehn work is runtime rollout, config verification, and live
  staged validation.
- Upstream publishability is intentionally parked until Zehn has more live
  confidence; this local branch contains private skills/supervision history and
  must not be treated as an upstream-ready branch.

Runtime rollout checklist:

1. Rebuild the current source into the launcher and CLI binaries.
2. Restart through the launcher UI so gateway lifecycle stays standard.
3. Confirm active config path is the intended `PICOCLAW_HOME`.
4. Confirm `delegate_to_agent`, `delegation_status`, `delegation_inbox`, and
   `start_agent_meeting` are enabled only for intended agents.
5. Keep GitHub artifact writer disabled for the first local tests.
6. Run local CLI/Web delegation smoke.
7. Run local CLI/Web meeting smoke.
8. Enable Yaad MCP and verify one terminal delegation memory write.
9. Run one narrow Discord command in the allowlisted channel.
10. Enable GitHub artifacts last, using one low-risk approval/follow-up test.

## Baseline Local Smoke

Run this before every live stage:

```bash
cd /Users/aliai/zehn
operations/zehn-live-verification-smoke.sh
go test ./pkg/config -run '^$' -count=1
```

Expected result:

- The smoke helper reports zero failures.
- `go test ./pkg/config -run '^$' -count=1` passes without running package
  tests.
- No external service token is required.
- No Discord, GitHub, or Yaad side effect occurs.

## Stage 1: Local CLI Only

Purpose: verify that task automation and delegation/meeting documentation gates
are usable without starting the gateway or enabling integrations.

Operator gate: none. This stage is mandatory and local-only.

Commands:

```bash
cd /Users/aliai/zehn
bash -n operations/run-one-zehn-feature-task.sh
operations/audit-zehn-feature-task.sh 017-staged-local-live-verification
operations/zehn-live-verification-smoke.sh
```

Pass criteria:

- Bash syntax checks pass.
- The task audit passes.
- The smoke helper confirms that the rollout plan contains local, Yaad, GitHub,
  Discord, operator-gate, and rollback coverage.

Rollback/disable: no persistent service is started. Stop after this stage if any
check fails and fix the local files before proceeding.

## Stage 2: Local Gateway Only

Purpose: verify gateway startup with external integrations disabled. Discord is
still a human visibility layer only and must not be used as an internal
delegation bus.

Operator gate: confirm the active config has no enabled live Discord channel,
no live GitHub artifact writer, and no Yaad MCP server enabled.

Preparation:

```bash
cd /Users/aliai/zehn
unset ZEHN_LIVE_DISCORD_CONFIRM
unset ZEHN_LIVE_GITHUB_CONFIRM
unset ZEHN_LIVE_YAAD_CONFIRM
unset PICOCLAW_DELEGATION_MEMORY_STRICT
```

Run the gateway in the repo's normal local setup mode. Prefer loopback binding
and setup-tolerant startup:

```bash
go run ./cmd/picoclaw gateway -E
```

Pass criteria:

- The gateway starts locally.
- No Discord channel connects.
- Delegation and meeting records remain local workspace artifacts.
- GitHub artifact status is skipped when no writer is installed.
- Yaad memory status is unavailable or skipped when no writer is available.

Rollback/disable:

- Stop the gateway process.
- Keep Discord, GitHub, and Yaad config disabled.
- Leave `PICOCLAW_DELEGATION_MEMORY_STRICT` unset unless deliberately testing
  strict Yaad failure behavior.

## Stage 3: Yaad MCP Enabled

Purpose: verify that Yaad can be connected as the durable memory path without
enabling GitHub or Discord.

Live side effect: writes to Yaad memory only after the explicit gate below.

Operator gate:

```bash
export ZEHN_LIVE_YAAD_CONFIRM=write-yaad-memory
```

Required checks before the gate:

- The configured MCP server name is `yaad`, or it exposes a `memory_add` tool.
- The Yaad MCP server points at the intended local/private Yaad runtime.
- `PICOCLAW_DELEGATION_MEMORY_STRICT` is unset for the first test so local
  delegation records survive Yaad errors.
- GitHub and Discord remain disabled.

Dry run:

```bash
unset ZEHN_LIVE_YAAD_CONFIRM
operations/zehn-live-verification-smoke.sh
```

Live run:

```bash
test "${ZEHN_LIVE_YAAD_CONFIRM:-}" = "write-yaad-memory"
go run ./cmd/picoclaw mcp list
go run ./cmd/picoclaw mcp show yaad
```

Then run one narrow local delegation or meeting command through the local CLI or
gateway path that creates a completed delegation record. Confirm the record has
a Yaad durable memory status of `written` and a memory ID, or a recorded Yaad
failure if the service is intentionally unavailable.

Rollback/disable:

```bash
unset ZEHN_LIVE_YAAD_CONFIRM
unset PICOCLAW_DELEGATION_MEMORY_STRICT
```

Disable or remove the Yaad MCP server from the active private config, then
restart the gateway. Confirm later delegation records show Yaad unavailable or
skipped rather than blocking local completion.

## Stage 4: GitHub Artifact Writer Enabled

Purpose: verify executable delegation and meeting artifacts with fake or dry-run
GitHub first. GitHub Project is the tracker, not the company brain; durable
memory remains in local records and Yaad.

Live side effect: creates GitHub issues/comments only after the explicit gate
below.

Operator gate:

```bash
export ZEHN_LIVE_GITHUB_CONFIRM=create-github-artifact
```

Required checks before the gate:

- Prefer fake or dry-run writer mode when available.
- If no fake/dry-run writer exists in the active runtime, leave the writer
  disabled and confirm records show `github artifact writer disabled`.
- Yaad may remain enabled from Stage 3, but Discord remains disabled.
- The test request must be executable or approval-tracked so artifact creation
  is expected; advisory delegations should not create issues.

Dry run:

```bash
unset ZEHN_LIVE_GITHUB_CONFIRM
operations/zehn-live-verification-smoke.sh
```

Live run:

```bash
test "${ZEHN_LIVE_GITHUB_CONFIRM:-}" = "create-github-artifact"
```

Run one narrow executable delegation or approval-tracked meeting. Confirm:

- Exactly one intended GitHub issue is created or one fake/dry-run artifact is
  recorded.
- Participant comments are curated and material.
- Raw internal prompts and private participant turns are not written to GitHub.
- The local delegation or meeting record stores the GitHub artifact status.

Rollback/disable:

```bash
unset ZEHN_LIVE_GITHUB_CONFIRM
```

Disable the artifact writer or GitHub MCP/config path, restart the gateway, and
confirm new executable records show `github artifact writer disabled`. Close or
label any live test issue according to the project cleanup convention.

## Stage 5: Discord Summaries Enabled

Purpose: verify concise Discord visibility summaries without using Discord as
the internal delegation or meeting bus.

Live side effect: sends Discord summary messages only after the explicit gate
below.

Operator gate:

```bash
export ZEHN_LIVE_DISCORD_CONFIRM=send-discord-summary
```

Required checks before the gate:

- Use a private test server/channel and a bot token intended for verification.
- `allow_from` is restricted to the operator account or test channel.
- Group trigger settings require an explicit mention or prefix.
- Remote shell, command cron, and broad file tools are disabled for this live
  channel unless separately approved.
- Yaad and GitHub remain at the previously verified stage, or are disabled if
  isolating Discord.

Dry run:

```bash
unset ZEHN_LIVE_DISCORD_CONFIRM
operations/zehn-live-verification-smoke.sh
```

Live run:

```bash
test "${ZEHN_LIVE_DISCORD_CONFIRM:-}" = "send-discord-summary"
go run ./cmd/picoclaw gateway -E
```

Run one local delegation or meeting action that emits visibility summaries.
Confirm Discord receives only concise status/recommendation text. It must not
receive raw internal prompts, raw participant turns, secrets, or unredacted
private payloads.

Rollback/disable:

```bash
unset ZEHN_LIVE_DISCORD_CONFIRM
```

Disable the Discord channel in the active config, remove or rotate the test bot
token if exposed, and restart the gateway. Confirm no new Discord messages are
sent when repeating the local action.

## Stage 6: One Narrow Live Discord Command

Purpose: verify the complete user-visible path with the smallest possible live
Discord input after local CLI, gateway, Yaad, GitHub, and Discord summaries have
each passed independently.

Live side effect: one operator-authored Discord command plus the bot's concise
reply/summary.

Operator gate:

```bash
export ZEHN_LIVE_DISCORD_COMMAND_CONFIRM=send-one-command
```

Command shape:

```text
@<bot> run Zehn verification: ask li-ceo for a one-sentence readiness summary;
do not create external tasks; do not contact customers; do not spend money.
```

Pass criteria:

- The bot responds in the intended Discord channel.
- The route uses the configured agent identity.
- No GitHub issue is created unless the request explicitly asks for executable
  or approval-tracked work.
- Yaad writes only curated durable summary material.
- Discord output remains concise and human-visible, not an internal transcript.

Rollback/disable:

```bash
unset ZEHN_LIVE_DISCORD_COMMAND_CONFIRM
unset ZEHN_LIVE_DISCORD_CONFIRM
unset ZEHN_LIVE_GITHUB_CONFIRM
unset ZEHN_LIVE_YAAD_CONFIRM
```

Stop the gateway, disable Discord, disable GitHub artifact writer, and disable
Yaad MCP in the active private config. Restart only the local gateway stage if a
post-rollback check is needed.

## Live Integration Rules

- Live side effects are opt-in and must be named with the `ZEHN_LIVE_*_CONFIRM`
  gates above.
- Default verification must not send live Discord messages.
- Default verification must not write live GitHub issues or comments.
- Default verification must not write live Yaad memories.
- Run only one new live integration at a time.
- Record the exact command, config profile, timestamp, and observed artifact IDs
  in the operator notes before proceeding to the next stage.
