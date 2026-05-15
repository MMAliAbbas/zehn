# Zehn Proactive Operating Cadence — Proposal v3

Date: 2026-05-15
Status: proposal, awaiting Ali approval. No new cron jobs are enabled by this change.
Supersedes (on activation): `.picoclaw/workspace/memory/LOGICIGNITER_OPERATING_CADENCE.md` v2.

## Why

Zehn is fully operational and functionally idle.

Live inspection (2026-05-15 ~22:00 local):

| Signal | Result |
| --- | --- |
| Gateway process | PID 85278, up 22h+ |
| Cron set | 4 jobs, all enabled, lastStatus=ok |
| Heartbeat | firing every 30 min |
| Gateway log activity | 42,683 lines / last 24h |
| Heartbeat-work-selection delegations | ~48 / last 24h, all completed |
| Yaad reachability | reachable; Yaad writes per cycle |
| COO daily scoreboard | generated for 2026-05-15 at 08:41 |
| Release Readiness Ladder | 0 services placed; 55 in `assessment-pending` |
| `zehn:ready` issues across 77 repos | **0** |
| CEO activations in last 24h | **0** (none scheduled until next Monday) |
| Outcomes shipped in last 24h | 0 PR merged by Zehn agents (the two LI PRs that merged today were authored interactively, not by Zehn) |

The system is doing exactly what v2 said to do — disciplined no-work-found reporting. The problem is that under v2, no role is creating fresh `zehn:ready` work daily. The CEO is structurally asleep 6 days out of 7. The non-engineering roles (CRO/CMO/CFO/Legal/CHRO/CCO/CDAO/CISO) are dormant; their Phase-1 deliverable specs in `LOGICIGNITER_NON_ENG_DELIVERABLES.md` were never operationalized.

The CEO's own `workspace-li-ceo/AGENT.md` already mandates the right posture:

> "Choose the next meaningful company outcome, not just the next inspection."
> "Run LogicIgniter like an active software company in development, not like a readiness monitor."
> "Activate CRO, CMO, CFO, Legal, CHRO, CCO, CISO, and CDAO for concrete internal deliverables, not passive status summaries."

The schedule denies it the chance to operate on that mandate.

## What v3 adds (and what it doesn't)

**Adds**: three activations.

1. **CEO Daily Sync** — `0 8 * * 1-5` (08:00 Mon–Fri). New cron `li-ceo-daily-sync`, target `li-ceo`. Bounded scan → one Yaad-recorded decision → up to 3 delegations with terminal-outcome expectations. Fires 30 min before the existing COO daily synthesis, so COO sees fresh CEO direction when generating the scoreboard.

2. **Non-Engineering Weekly Pulse** — `0 9 * * 2` (Tue 09:00). New cron `li-nonexec-weekly-pulse`, target `li-ceo`. CEO chairs a single-turn pulse covering eight non-eng roles. Each role reports last-week delivered, commits to one this-week deliverable. Each commitment becomes a `zehn:ready` issue against `LOGICIGNITER_NON_ENG_DELIVERABLES.md` Phase-1 specs.

3. **Release Ladder Assessment Sweep** — one-shot, no cron. Manual trigger: CEO runs the `logicigniter-release-ladder-assessment.md` operating prompt once (during the first CEO Daily Sync after v3 activates), delegating to CTO + architecture/backend specialists to classify the 55 `assessment-pending` services onto the seven Release Ladder stages. Terminal outcome: a Yaad-stored ladder snapshot the COO scoreboard reads on subsequent days.

**Does not change**:

- v2's heartbeat-driven work selection loop.
- v2's terminal-outcome rule, `Failure Reason:` discipline, Yaad write-back contract.
- The four existing v2 cron jobs (Monday plan, daily synthesis, ops monitor, Friday review).
- Approval-escalation policy: DNS/SSL/auth/billing/migrations/new-repo/etc. stay `approval:ali-required`.
- Open delegation (`allow_agents: ["*"]`) stays.
- `tools.exec.enable_deny_patterns: false` stays per Ali's explicit decision.

## Why this is not a return to the v1 polling storm

v1 had 17 cron jobs producing `HEARTBEAT_OK` returns. The problem wasn't the firings — it was **firings without terminal outcomes**. v2 fixed that.

v3 adds two activations that inherit every v2 discipline:

| v2 discipline | v3 inherits |
| --- | --- |
| `HEARTBEAT_OK` only when no actionable work found | Yes — CEO daily sync may return `HEARTBEAT_OK` if no proactive action surfaces |
| Bounded payload (≤ 250 chars in `jobs.json`) | Yes — both new payloads ≤ 250 chars referencing operating-prompt files |
| Terminal outcome rule | Yes — both prompts list the seven terminal-outcome categories from `workspace-li-ceo/AGENT.md` |
| Mandatory Yaad write per cycle | Yes — both prompts require a Yaad write before terminal close |
| `Failure Reason:` field on incomplete delegations | Yes — inherited from existing CEO doctrine |
| Discord output ≤ 500 chars with Yaad-entry pointer | Yes — same Discord-output contract as v2 |

Going from 4 jobs → 6 jobs is a 50% increase in cron volume. Going from 1 CEO firing/week → 5 CEO firings/week is the structural shift. Neither approaches the v1 17-job density.

## How v3 changes the daily flow

Today (v2):

```
07:00  …silence…
08:30  COO daily-synthesis — generates scoreboard, posts to Discord
09:00  …silence except heartbeat work-selection every 30 min returning no-work-found…
…
```

After v3 activation:

```
08:00  CEO daily-sync — reads scoreboard from yesterday, reads Yaad
       terminal-outcome log, scans blockers, picks ONE forward move,
       writes Yaad decision, delegates with terminal-outcome expectation
08:30  COO daily-synthesis — generates scoreboard incorporating CEO's
       fresh direction; heartbeat-work-selection picks up the new
       `zehn:ready` items CEO created
09:00+ Specialists begin single-chain implementing delegations
…
Tue 09:00 (additional): CEO chairs non-eng weekly pulse — eight roles
       commit to one deliverable each, each commitment becomes a
       `zehn:ready` issue
```

## Activation steps (Ali's call, in order)

1. **Review and merge this PR.** Doctrine + prompts + jobs.json delta land on `main`. New jobs ship `enabled: false` — nothing fires yet.
2. **Optional: pre-seed Yaad** with the v3 cadence doc as a `runbook`-class entry under `organization:logicigniter`. CEO will read it on first run.
3. **Flip `li-ceo-daily-sync.enabled` to `true`** in `jobs.json` and reload cron. First firing: next 08:00 weekday.
4. **Watch one CEO daily-sync run** (~5 min). Inspect: Yaad decision ID, Discord message, delegations created.
5. **If healthy, flip `li-nonexec-weekly-pulse.enabled` to `true`.** First firing: next Tuesday 09:00.
6. **First CEO Daily Sync after #3 runs the ladder-assessment sweep** automatically per the prompt's first-run condition. Sweep terminal-outcome lands within 1–2 days.

## Rollback

Per added job: flip `enabled: false` in `jobs.json` and reload. No state cleanup required; any in-flight delegations complete normally.

If the model proves wrong (CEO daily sync produces noise rather than direction, or non-eng pulse fails to drive deliverables), the cadence doc has a documented v2 fallback — flip both new jobs off and the system returns to v2.

## Files Changed (Mix Of Tracked And Live-Untracked)

`.picoclaw/` is gitignored — it is the agent runtime, not source-controlled.
The runtime files below are **already in place on disk** with the new
jobs DISABLED until you activate them. Only the supervision proposal doc
ships in this PR; the runtime files are reviewable on disk at the paths
listed.

| File | Location | Tracked? | Status after this change |
| --- | --- | --- | --- |
| `supervision/ZEHN_PROACTIVE_CADENCE_PROPOSAL_20260515.md` | git | yes (this PR) | the doc you are reading |
| `.picoclaw/workspace/memory/LOGICIGNITER_OPERATING_CADENCE.md` | runtime | no | v2 → v3 in-place; live on disk |
| `.picoclaw/workspace/operating-prompts/logicigniter-ceo-daily-sync.md` | runtime | no | new file, live on disk |
| `.picoclaw/workspace/operating-prompts/logicigniter-nonexec-weekly-pulse.md` | runtime | no | new file, live on disk |
| `.picoclaw/workspace/operating-prompts/logicigniter-release-ladder-assessment.md` | runtime | no | new file, live on disk |
| `.picoclaw/workspace/cron/jobs.json` | runtime | no | +2 entries (`li-ceo-daily-sync`, `li-nonexec-weekly-pulse`), both `enabled: false` |

Inspection commands (read-only):

```bash
# Cadence doc (v3) — diff against pre-v3 state in git history of /Users/aliai/zehn
diff <(git show HEAD:.picoclaw/workspace/memory/LOGICIGNITER_OPERATING_CADENCE.md 2>/dev/null || echo "not in HEAD — gitignored") \
     /Users/aliai/zehn/.picoclaw/workspace/memory/LOGICIGNITER_OPERATING_CADENCE.md

# New operating prompts
cat /Users/aliai/zehn/.picoclaw/workspace/operating-prompts/logicigniter-ceo-daily-sync.md
cat /Users/aliai/zehn/.picoclaw/workspace/operating-prompts/logicigniter-nonexec-weekly-pulse.md
cat /Users/aliai/zehn/.picoclaw/workspace/operating-prompts/logicigniter-release-ladder-assessment.md

# jobs.json delta — confirms 4 enabled (v2 set) + 2 disabled (new v3)
python3 -c "import json; d=json.load(open('/Users/aliai/zehn/.picoclaw/workspace/cron/jobs.json')); [print(f\"{'ON ' if j['enabled'] else 'off'}  {j['name']:30s} {j['schedule']['expr']}\") for j in d['jobs']]"
```

Activation is a two-line edit of `jobs.json` per the steps under
"Activation steps" above — flip `enabled: false` to `enabled: true` on
the two new entries, then reload cron.

## Open questions Ali may want to settle in review

1. **Daily-sync time**: 08:00 Mon–Fri proposed. Could be 07:30 or 08:15. Constraint: must be before 08:30 COO synthesis so the scoreboard incorporates CEO direction.
2. **Non-eng pulse day**: Tuesday 09:00 proposed (gives Monday weekly-plan time to set the week's outcome before non-eng roles commit). Could be Monday 10:00 instead, immediately after weekly-plan.
3. **Channel routing**: CEO daily sync proposed for the CEO Discord channel (`1487902195734024353`). Could alternatively post to a new `ali-direct` channel or be silent (Yaad-only) for low-event days.
4. **Ladder assessment owner**: CEO → CTO → architecture/backend specialists proposed. CPO could alternatively be the assessment chair since the ladder is portfolio readiness, not pure architecture.
