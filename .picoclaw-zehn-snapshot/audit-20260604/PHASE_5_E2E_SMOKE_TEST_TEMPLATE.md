# Phase 5 — Supervised E2E Smoke Test (TEMPLATE)

## Update 2026-06-05 13:30 +05

Phase 5 was subsequently fired — but not via the template procedure below. When the launcher was re-bootstrapped at 10:54 +05, the existing in-flight CEO delegation `delegation-20260604T172915.971462000Z-824f9283d096` (queued just before the freeze) resumed naturally and produced one full canonical-loop completion: PR `logicigniter/svc-logicigniter-web#139` + comment `4628906133` on issue #127 + Yaad `memory_add`. See `E2E_PROOF_svc-logicigniter-web-127.md` for the live evidence. The template below is preserved as the *procedure* for future smoke tests; today's run did not use it.

## Status (original)

**NOT YET FIRED.** Phase 5 requires Ali's explicit action: (a) pick one low-blast-radius GitHub issue, (b) re-bootstrap the launcher, (c) supervise the run with logs open. None of these were done autonomously because the safety value of "Ali picks" is structural — the recovery plan calls it out specifically.

## Pre-fire state (completed autonomously)

- [x] All non-monitor cron jobs **disabled** in `workspace/cron/jobs.json`. Only `zehn-operations-monitor-v2` (hourly read-only inspection) remains enabled. Pre-edit backup at `workspace/cron/jobs.json.20260605-pre-phase5`.
- [x] Code fixes from Phase 3 are in `/Users/aliai/zehn` working tree (uncommitted; HEAD still `77d13f90`):
  - `pkg/channels/pico/pico.go` — rejected-origin diagnostic logging
  - `pkg/tools/integration/mcp_tool.go` — MCP blank-error annotation
- [x] Binaries rebuilt: `build/picoclaw-darwin-amd64` (37 MB, Phase 3 fixes baked in) and `build/picoclaw-launcher-darwin-amd64` (23 MB). Symlinks `build/picoclaw` and `build/picoclaw-launcher` updated. Launcher reports `v0.2.9-209-g77d13f90-dirty` because the two edits are not yet committed.
- [x] Logs rotated, live files empty, ready for fresh capture.
- [x] Doctrine reconciled (Phase 2): a fired agent run will read one canonical HEARTBEAT_OK rule set.
- [x] Yaad durable memory truthed (Phase 1.1): no `resolved/remaining_issues:[]` overclaim sitting where a CEO cycle would re-read it.
- [x] Delegation records archived to 1,400 live (Phase 4): `delegation_status` will be fast and not surface the May-month stuck-running zombies.

## What Ali needs to do

### Step 1 — Decide whether to commit the Phase 3 code edits before the smoke test

The 2 edits in `/Users/aliai/zehn` are uncommitted. Either:
- (a) commit them as a single PR or local commit before re-bootstrapping (cleaner version string, easier to roll back), or
- (b) leave them uncommitted, accept the `-dirty` version string for the smoke test, and commit afterwards.

Recommendation: (a). Suggested commit message:
> `fix: surface MCP blank-cause errors and Pico rejected-origin diagnostics`
>
> - pkg/channels/pico/pico.go: log the rejected origin/allowed list/remote addr when CheckOrigin fails so the gorilla "request origin not allowed" error becomes diagnosable
> - pkg/tools/integration/mcp_tool.go: annotate MCP CallTool errors with server name + tool name, and flag blank-cause errors so agents stop narrating "Yaad failed" when the SDK has swallowed the underlying transport error
>
> Driven by 2026-06-04 forensic audit. See /Users/aliai/.picoclaw-zehn/audit-20260604/PHASE_3_CODE_FIXES.md.

### Step 2 — Pick ONE low-blast-radius task

Constraints from the recovery plan:
- existing labeled `area:docs` or `area:frontend` issue (low blast)
- < 2 hours of agent work
- NOT a code-deploy, money-touching, secrets, infra, or new-repo task
- has clear acceptance criteria readable from the issue body

Candidate selection process:
```
gh search issues "user:logicigniter is:open is:issue label:zehn:ready label:area:docs" --json url,title,labels --limit 10
gh search issues "user:logicigniter is:open is:issue label:zehn:ready label:area:frontend" --json url,title,labels --limit 10
```

Pick one, record the issue URL + acceptance criteria here under "## Chosen task" before re-bootstrapping.

### Step 3 — Re-bootstrap launchd

```
launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/io.picoclaw.launcher.plist
launchctl list io.picoclaw.launcher   # confirm PID
```

The launcher will start the gateway. With only `zehn-operations-monitor-v2` enabled, the only cron firing will be at `15 * * * *` (every hour at HH:15) for read-only monitoring.

### Step 4 — Trigger the loop

Manually message the CEO Discord channel or `zehn-main` with the chosen task. The expected loop:

1. **scanner discovery** — message routes to `zehn-main` → `li-ceo` → `li-coo`; `li-coo` runs the work-queue scanner (`/Users/aliai/zehn/operations/logicigniter-work-queue-scan.py`) and surfaces the issue
2. **delegation creation** — `li-coo` dispatches to the matching specialist (`li-frontend-developer` or `li-docs` per `area:*` label)
3. **specialist turn start/end** — turn_id, iterations, tool calls visible in `gateway.log`
4. **GitHub artifact creation** — PR or issue comment posted (verifiable via `gh`)
5. **Yaad memory write** — `memory_add` for the decision, `memory_update` for the cycle ledger entry
6. **Discord summary** — published to the routed channel
7. **delegation closure** — terminal `status: completed` with `completed_at` timestamp, no manual editing

### Step 5 — Capture evidence

Create `audit-20260604/E2E_PROOF_<issue-number>.md` while the run is happening. For each stage above:
- log slice from `gateway.log` (timestamp range, key event lines)
- IDs (turn_id, delegation_id, memory_id, message_id)
- artifact URLs (`gh pr view`, `gh issue view`, `gh api ... comments`)
- duration metric

Suggested live log filter while watching:
```
tail -F /Users/aliai/.picoclaw-zehn/logs/gateway.log \
  | jq -c 'select(.tool // .event_kind // .message | tostring | test("mcp_yaad|delegation|discord|tool execution|agent.turn|CheckOrigin"))'
```

## Phase 5 acceptance criteria (from recovery plan)

- [ ] All 7 loop stages have evidence captured in the proof doc.
- [ ] No hand-edits to delegation JSON during the run.
- [ ] No "Yaad failed" or "all resolved" prose in any LLM output for the run.
- [ ] No `Upgrader.CheckOrigin` rejections in `logs/gateway.log` during the run (if any, the new diagnostic logging at `pkg/channels/pico/pico.go:138` will reveal the rejected origin so the config can be widened).

If any stage fails, do NOT "fix forward" — stop, return to Phase 1/2/3 for the broken stage, fix the root cause, then re-run from Step 4.

## Why I stopped here

Re-bootstrapping the launcher and picking the task are decisions that affect real GitHub state and re-start an autonomous system that just had a confirmed pattern of overclaim. The recovery plan explicitly names "Ali picks" as the gate. Honoring that boundary is more important than chasing a green checkmark on the goal hook.
