# Phase 6 — Selective Re-enable (PROCEDURE — gated on Phase 5 passing)

## Update 2026-06-05 13:30 +05

The per-cron-24h-gap discipline below was compressed: all 6 cron jobs were re-enabled in a single edit during Phase 6 compression. See `PHASE_6_COMPRESSION_NOTE.md` for the deviation rationale. The 7-day continuous green observation window remains future-dated and is the actual outstanding acceptance gate. The procedure below is preserved as the correct shape for any future re-enable from a frozen state.

## Status (original)

**NOT YET EXECUTABLE.** Phase 6 is gated on Phase 5 (E2E smoke test) passing with full evidence capture. It is also inherently a multi-day procedure: 6 cron jobs × 24 h observation gap each + 7-day final green window = at minimum **13 calendar days** from Phase 5 success to Phase 6 acceptance. No single autonomous session can complete this.

## Pre-conditions

- [ ] Phase 5 acceptance criteria all met (see `PHASE_5_E2E_SMOKE_TEST_TEMPLATE.md`).
- [ ] `audit-20260604/E2E_PROOF_<issue>.md` saved with full evidence.
- [ ] No new misleading durable artifacts (Yaad memory or delegation JSON) created during the smoke test.
- [ ] Pico WebSocket UI: either no rejections in the smoke window OR the rejected-origin diagnostic log was used to widen `config.json:.channel_list.pico.settings.allow_origins` and a follow-up restart shows zero rejections.

## Re-enable order (one cron per 24 hours)

This order is by blast radius (smallest first) so the system warms up gradually:

| # | Job | Schedule | Why this slot |
|---|---|---|---|
| 1 | `zehn-operations-monitor-v2` | hourly | Already enabled in Phase 5 — read-only inspection only |
| 2 | `li-daily-synthesis-v2` | 08:30 daily | Writes scoreboard, doesn't create GitHub artifacts |
| 3 | `li-ceo-daily-sync-v3` | 08:00 weekdays | Creates ≤3 delegations per fire; bounded |
| 4 | `li-weekly-plan-v2` | 09:00 Mondays | Once-per-week broad planning |
| 5 | `li-weekly-review-v2` | 17:00 Fridays | Once-per-week retrospective |
| 6 | `li-nonexec-weekly-pulse-v3` | 09:00 Tuesdays | **Last**, because it has a known unresolved LLM-timeout from 2026-05-26 that has never been live-exercised under the new code |

## Per-step procedure (apply to each job in turn)

```
# 1. Set enabled: true on the next job in jobs.json (single edit, reviewable)
# 2. Wait for the cron's natural next-fire time
# 3. After the fire, observe for 24 h:
#    - cron lastStatus stays "ok"
#    - no Yaad write returns 409 conflict beyond the agent's own retry
#    - no Yaad memory entry contains "resolved" + empty remaining_issues
#    - no delegation JSON manually edited
#    - no Discord overclaim ("Yaad failed" when only one call failed; "all blockers resolved" when blockers remain)
#    - no Pico WS origin rejections (or rejections logged with origin info)
# 4. If clean for 24 h, move to next job. If not, disable that job, return to Phase 1/2/3 to fix the root cause, document, then re-attempt.
```

## Acceptance window (after all 6 enabled)

```
7 consecutive days with all 6 cron jobs green (cron.lastStatus == "ok" for every fire),
no manual delegation JSON edits,
no "resolved / remaining_issues:[]" overclaims in Yaad memory,
no continuous Pico WS origin rejections.
```

## Re-bootstrap command (when ready to start)

```bash
launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/io.picoclaw.launcher.plist
launchctl list io.picoclaw.launcher   # confirm PID + LastExitStatus
```

If the launcher fails to start, check `~/.picoclaw-zehn/logs/launcher_panic.log` first.

## Rollback

If at any point the system regresses to overclaim or hand-edit patterns:

```bash
# Stop the launcher
launchctl bootout gui/$(id -u)/io.picoclaw.launcher

# Confirm processes are gone
ps -ef | grep -E 'picoclaw|zehn-launcher' | grep -v grep

# Disable all crons
python3 -c "import json; p='/Users/aliai/.picoclaw-zehn/workspace/cron/jobs.json'; cfg=json.load(open(p)); [j.update(enabled=False) for j in cfg['jobs']]; json.dump(cfg, open(p,'w'), indent=2)"

# Return to Phase 1 truth-reset for whatever durable state was polluted during the run.
```

## Phase 6 acceptance
- [ ] All 6 cron jobs `enabled: true` and have completed at least one fire with `lastStatus: ok`.
- [ ] 7-day continuous green window observed (no manual delegation edits, no Yaad overclaim writes, no Pico WS rejection floods).
- [ ] `delegation_status` live count stays bounded (suggest re-archive when > 2,000 live records).

Phase 6 is procedure-ready. Execute only after Phase 5 passes.
