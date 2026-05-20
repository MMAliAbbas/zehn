# LogicIgniter Operating Model Build Plan

Date: 2026-05-13
Author: Claude (concrete implementation plan)

> **Execution Status (2026-05-13, end of session)**
>
> Artifacts 1–11 shipped or staged. Artifact 12 (controlled drill) pending Zehn restart approval.
>
> - Artifact 1 (`verify-pr.sh` v1): merged to `logicigniter/scripts:main` via PR #10.
> - Artifact 2 (GitHub Actions + composite action): merged in same PR; scripts CI green; composite action ready for other repos once `LI_SCRIPTS_READ_TOKEN` is provisioned.
> - Artifacts 3, 4, 5, 6, 9, 10, 11: doctrine docs written under `.picoclaw/workspace/memory/`. Artifact 5 replaced the old cadence in place; v1 archived.
> - Artifact 7: new `cron/jobs.json` with 5 cadence jobs, all `enabled: false` pending approval. Previous 17-job file backed up to `jobs.json.pre-cadence-v2-20260513`.
> - Artifact 8: setup doc written (`LOGICIGNITER_EVENT_DISPATCH_SETUP.md`); rules NOT applied to `config.json` until Discord-GitHub webhook channels are wired.
> - Search hygiene: 3 stale memory docs and 6 stale operating-prompts archived.
> - LI repos touched: only `scripts/main` (one merge). Other LI repos untouched. Zehn not started; no cron enabled.
>
> Artifact 12 acceptance: one drill issue traverses claim → branch → verify-pr.sh → PR → review → merge → reconcile → Yaad write under autonomous execution. Awaiting Ali approval to start Zehn and enable the first cron jobs.

This document replaces the menu-of-options framing in
`supervision/CLAUDE_ZEHN_RECOVERY_AUDIT.md` §13. The audit stays as the
restart-safety reference. This is the build plan.

## Why The Company Isn't Operating

Twelve weeks of audits have already named the operational gaps. The
shortest honest version:

LogicIgniter has roles, schedules, and tools. It does not have an
**operating motion**. A real software company has events, decisions,
accountable owners, and durable progress. LogicIgniter has cron polling,
parallel monitors, ephemeral Discord output, and a `HEARTBEAT_OK` reply
loop. Same agents wake up every hour and rediscover the same state.

The fix is not more roles, more prompts, more cron jobs, or more
audits. The fix is to make work *terminal*, define the *motion*, and
*delete* what doesn't contribute. That is what this plan does.

## What A Real Software Company Has That LogicIgniter Doesn't

| Gap | Real company | LogicIgniter today |
| --- | --- | --- |
| Backlog with priority | One ranked queue, "Now / Next / Later" | 10 area-labeled queues, no ranking |
| Planning cadence | Weekly plan + daily standup + weekly review | Hourly cron polls everything |
| Release ladder | Per-product stage with explicit criteria | "51 apps launch together" — no per-app state |
| End-to-end owner | One engineer drives an issue to merge | Cron re-scans every cycle; no continuous ownership |
| Event-driven triggers | PR opened → reviewer pinged, issue filed → triaged | Polling every 30 min, no event routing |
| Scoreboard | WIP, cycle time, readiness %, blocker register | Discord messages saying `HEARTBEAT_OK` |
| Coordination meetings | Weekly all-hands, standups, design reviews | `start_agent_meeting` tool exists, unused |
| Durable memory | Wiki, decision log, ADRs | Yaad wired, unused — agents reason from scratch |
| Quality gates that fail | CI, QA approval, security review | No `verify-pr.sh`, no Actions in most repos |
| Failure with shape | Postmortems, retro, RCA | `HEARTBEAT_OK` masks failure |
| External signal | Customers, sales pipeline, support tickets | Closed-loop system; no customer-facing surface yet |
| Non-eng dept activity | Sales discovery, marketing drafts, CFO model | CRO/CMO/CFO/Legal/CHRO/CCO have personas, no outputs |

This table is the diagnosis. The artifacts below are the treatment.

---

## The 12 Artifacts To Build

Each artifact has a **file path**, **purpose**, **acceptance criteria**,
**depends on**, and **size**. "Size" is rough: S = a few hours, M = a
day, L = several days. I am committing to build all 12 in the order
shown unless you redirect.

### Artifact 1 — `verify-pr.sh` v1 (size: M, P0)

**Path:** `/Users/aliai/logicigniter/scripts/verification/verify-pr.sh`

**Purpose:** Make "verified" a terminal state. Until this exists, no PR
ever truly merges with confidence, and every audit cited it as the
critical blocker.

**Behavior (v1, narrow):**
- Args: `--repo <name> --issue <num>` (optional `--pr <num>`).
- Detect repo from `git remote get-url origin`.
- Detect changed files vs `main`: `git diff --name-only main...HEAD`.
- Run repo-local `scripts/verify.sh` or `Makefile verify` target if
  present; capture exit + duration.
- Fallback by repo class:
  - Go svc-* and go-packages: `go test ./...` on changed package paths.
  - scripts and operations: `bash -n` on each changed `*.sh`.
  - business, supervision, proto, config: skipped with reason in
    evidence (v2 will add lint/schema).
- Write JSON evidence to
  `/Users/aliai/logicigniter/operations/.verify-pr/<repo>-<issue>-<ts>.json`
  with: repo, files, each step's command/duration/exit, final verdict
  (`pass` | `fail` | `skipped`), one-line `reason`.
- Exit 0 on `pass` or `skipped-with-reason`, non-zero on `fail`.

**Acceptance:** runs on one Go svc-* repo issue and one `scripts` repo
issue, produces deterministic pass/fail and a valid JSON file. v2 (adds
integration_tests, MCP runtime proof, doc lint) is deferred.

**Depends on:** nothing.

---

### Artifact 2 — Minimum GitHub Actions workflow per repo class (size: S, P0)

**Paths:**
- `/Users/aliai/logicigniter/scripts/.github/workflows/verify.yml`
- `/Users/aliai/logicigniter/integration_tests/.github/workflows/verify.yml`
- `/Users/aliai/logicigniter/business/.github/workflows/verify.yml`
- Template at `/Users/aliai/logicigniter/.github/workflow-templates/verify.yml`

**Purpose:** Make `gh pr checks` return non-empty. Today most repos have
no workflow, so `statusCheckRollup` is empty, and the merge gate can't
distinguish "no checks configured" from "checks failed."

**Behavior:** Each workflow runs `verify-pr.sh` (Artifact 1) on PR
open/sync.

**Acceptance:** Open one test PR per repo; `gh pr checks <pr>` shows a
green check from the new workflow.

**Depends on:** Artifact 1.

---

### Artifact 3 — Release Readiness Ladder (size: M, P0)

**Path:**
`/Users/aliai/.picoclaw-zehn/workspace/memory/LOGICIGNITER_RELEASE_READINESS_LADDER.md`

**Purpose:** Give CPO and CTO measurable progress per app. "All 51
launch together" remains the constraint, but per-app state becomes
observable.

**Content:** Seven stages per app:

1. **Skeleton** — repo exists, basic build succeeds.
2. **Backend** — service contract complete, gRPC/API implemented, unit
   tests pass.
3. **Frontend** — UI surface implemented, hooked to backend, basic
   navigation works.
4. **Integration** — service participates in cross-service flows,
   integration tests pass.
5. **Quality** — QA scenarios written and passing, security review
   done.
6. **Docs** — user/admin/operator docs exist and pass review.
7. **Launch-Ready** — verify-pr.sh green on main, post-merge reconcile
   tested, sign-off recorded in Yaad.

For each of the 51 apps, the doc records: current stage, blockers, next
stage criteria not yet met, owner. This is the canonical readiness
state — copied into Yaad once stable.

**Acceptance:** All 51 apps have a row with current stage and at least
one named blocker or "ready for next stage" note.

**Depends on:** nothing (read-only of LI repo state).

---

### Artifact 4 — Work Selection Algorithm (size: S, P0)

**Path:**
`/Users/aliai/.picoclaw-zehn/workspace/memory/LOGICIGNITER_WORK_SELECTION.md`

**Purpose:** Replace "10 specialists each scan everything every hour"
with a concrete, ranked selection rule. This is the missing algorithm
under the cron jobs.

**Content (concrete rules, not aspirations):**

- **Source:** GitHub Issues across the LI repo set labeled `zehn:ready`.
- **Filter:** exclude `approval:ali-required` unless body has explicit
  Ali approval; exclude items with an active claim (in-progress within
  4 hours).
- **Rank:** primary by Release Ladder stage (Artifact 3): items that
  move an app forward in stage rank higher than internal cleanups;
  secondary by age (oldest first); tertiary by suite priority (Ali's
  approved order).
- **Claim:** specialist comments on the issue with agent ID + branch
  name + intended completion target; adds `zehn:in-progress` label.
- **Dedup:** one open claim per issue.
- **Retry:** verify failure twice → `zehn:blocked` + comment with
  failure summary; third failure → escalate to li-coo via
  `start_agent_meeting`.
- **Auto-release:** claims with no branch push within 4 hours
  auto-release (label flip back to `zehn:ready`, comment recording the
  release).

**Acceptance:** The COO can apply this algorithm by hand once and
produce a consistent ranked list.

**Depends on:** Artifact 3 (for rank input).

---

### Artifact 5 — Operating Cadence v2 (size: S, P0)

**Path:**
`/Users/aliai/.picoclaw-zehn/workspace/memory/LOGICIGNITER_OPERATING_CADENCE.md`
(replaces the current 248-line file, not appended to)

**Purpose:** Replace polling-only with a real cadence.

**Content:**

- **Weekly plan (Mon 9am):** CEO sets the week's outcome; COO confirms
  WIP capacity; CTO confirms tech direction. Posted to CEO Discord
  channel + written to Yaad.
- **Daily synthesis (08:30):** COO posts the scoreboard (Artifact 6),
  CEO posts the day's terminal target.
- **Continuous specialist work:** triggered by Artifact 4 selection,
  not by hourly cron. The sweep (Artifact 7) is reduced to 3×/day
  (08:00, 13:00, 18:00) — enough to detect new `zehn:ready` items
  without the polling-storm pattern.
- **Event triggers:** PR opened → reviewer routed; CI failure →
  on-call; `approval:ali-required` → Ali Discord ping.
- **Weekly review (Fri 17:00):** CEO posts what shipped, what slipped,
  what we learned (written to Yaad).
- **Monthly retro (last Friday of month):** decisions, anti-patterns,
  conventions written to Yaad.

**Acceptance:** Doc parses against `verify-zehn-role-personas.sh`; no
duplicate doctrine remains in the file; the prior 248-line version is
archived to `memory/archive/`.

**Depends on:** Artifacts 4, 6, 7 (referenced from this doc).

---

### Artifact 6 — COO Scoreboard (size: S, P1)

**Path:** generated daily at
`/Users/aliai/.picoclaw-zehn/workspace/memory/scoreboard/YYYYMMDD.md`
plus a live `/Users/aliai/.picoclaw-zehn/workspace/memory/scoreboard/LATEST.md`

**Purpose:** Make throughput visible. COO needs an instrument to
manage.

**Daily fields:**
- WIP count by ladder stage (Artifact 3).
- Open issue count by repo, with age histogram (0–24h, 24–72h, >72h).
- Open PR count by repo, with cycle-time-so-far.
- Blockers register: issue, owner, blocker, next check date.
- Stuck items: items >72h without state change.
- Yaad writes since last scoreboard.
- Verify-pr.sh pass/fail count since last scoreboard.

**Acceptance:** A scoreboard file exists each weekday; the LATEST
symlink points to it; COO Discord channel receives a one-screen summary
referencing the file.

**Depends on:** Artifacts 1, 3, 4.

---

### Artifact 7 — New Cron Job Set (size: S, P0)

**Path:** `/Users/aliai/.picoclaw-zehn/workspace/cron/jobs.json` (replace
the 17-job model in place; the current file is already in `.picoclaw/`
backups via existing snapshot pattern).

**Purpose:** Stop the polling storm. Hand work to the operating
cadence.

**Final set (4 active jobs after restart, all with payload ≤ 250 chars
referencing operating prompts, not duplicating doctrine):**

| Job | Schedule | Target | Purpose |
| --- | --- | --- | --- |
| `li-weekly-plan` | `0 9 * * 1` | li-ceo | Mon 9am weekly plan |
| `li-daily-synthesis` | `30 8 * * *` | li-coo | 08:30 scoreboard post |
| `li-specialist-sweep` | `0 8,13,18 * * *` | li-coo | 3×/day selection + delegation |
| `zehn-operations-monitor` | `15 * * * *` | zehn-main | Hourly Zehn health (read-only) |

Plus heartbeat unchanged (built-in, 30 min).

**Deleted:** the 10 specialist work-queue jobs, the daily training job,
the CEO hourly check (folded into li-weekly-plan + event triggers), the
COO hourly check (folded into li-daily-synthesis), the engineering
30-min check (replaced by event triggers), the GitHub control-plane
reconciler (folded into specialist-sweep), the personal check (moved to
heartbeat-routed delegation), the second zehn-operations-monitor slot.

**Acceptance:** `jobs.json` parses; 17-job set archived; new file
contains exactly 4 enabled-FALSE jobs awaiting restart approval.

**Depends on:** Artifact 5 (cadence) and Artifact 4 (selection rules).

---

### Artifact 8 — Event-Driven Dispatch Rules (size: S, P1)

**Path:** updates to
`/Users/aliai/.picoclaw-zehn/config.json` → `agents.dispatch.rules`

**Purpose:** Add reactive routing alongside the cron cadence.

**Add rules for:**
- Discord webhook of GitHub PR-opened in any LI repo → route to
  matching specialist for review.
- Discord webhook of issue labeled `zehn:ready` → route to
  li-specialist-sweep (next run).
- Discord webhook of CI failure → route to li-devops.
- Discord webhook of `approval:ali-required` → ping Ali channel.

(If Discord-GitHub webhook integration isn't already wired, that's
a smaller pre-work artifact; the dispatch rules themselves are just
config additions and the audit confirmed all dispatch keys are
supported.)

**Acceptance:** New rules added to `agents.dispatch.rules` with valid
config keys (per source audit); pre-restart syntax check passes.

**Depends on:** Discord-GitHub webhook setup (separate, ~30 min).

---

### Artifact 9 — Yaad Read-First / Write-Back Enforcement (size: M, P0)

**Paths:**
- `/Users/aliai/.picoclaw-zehn/workspace/memory/LOGICIGNITER_YAAD_SCHEMA_CONTRACT.md`
  — add the read/write posture rules
- `/Users/aliai/.picoclaw-zehn/workspace-li-ceo/AGENT.md` — add 2-line
  Yaad rule
- `/Users/aliai/.picoclaw-zehn/workspace-li-coo/AGENT.md` — same
- Each `workspace-li-*` specialist `AGENT.md` — same

**Purpose:** Stop agents rediscovering company structure on every run.
Make Yaad the brain.

**Content:**
- **Read-first rule:** Before scanning the filesystem for company
  structure, app inventory, or prior decisions, query
  `organization:logicigniter` Yaad memory.
- **Write-back rule:** Every terminal outcome — merged, approved,
  blocked-with-owner, escalated, deferred, replaced — writes a Yaad
  entry with decision, evidence pointer, owner, date.
- **Failure fallback:** If Yaad write fails, log `Failure Reason:` and
  write the same content locally to `memory/MEMORY.md`; next run
  retries Yaad.

**Acceptance:** Drill (Artifact 12) produces at least two successful
Yaad MCP calls (one read, one write) with entry IDs in the gateway log.

**Depends on:** nothing (Yaad is already wired per audit).

---

### Artifact 10 — Non-Engineering Role Deliverable Specs (size: M, P1)

**Path:**
`/Users/aliai/.picoclaw-zehn/workspace/memory/LOGICIGNITER_NON_ENG_DELIVERABLES.md`

**Purpose:** Activate CRO, CMO, CFO, Legal, CHRO, CCO. Right now they
have personas and no concrete output. Pre-launch, they should still
produce things.

**Content — concrete deliverables per role for the next 30 days:**

- **CFO (li-cfo):** monthly financial model (revenue scenarios per
  suite, cost model, runway). Output: `business/finance/model.md` PR.
- **CRO (li-cro):** ICP doc per suite (10 docs). Output:
  `business/sales/icp-<suite>.md`. Plus pricing-strategy v0.
- **CMO (li-cmo):** launch narrative per suite (10 docs). Output:
  `business/marketing/launch-narrative-<suite>.md`.
- **CCO (li-cco):** customer support playbook per suite. Output:
  `business/cco/support-playbook-<suite>.md`.
- **CHRO (li-chro):** hiring plan v0 + interview rubric for the first 3
  roles. Output: `business/people/hiring-plan.md`.
- **Legal (li-legal):** terms-of-service draft, privacy policy draft,
  DPA template. Output: `business/legal/{terms,privacy,dpa}.md`.

Each deliverable enters the same work-selection algorithm (Artifact 4),
so it shows up in COO's scoreboard.

**Acceptance:** Every non-eng role has at least one concrete issue in
`zehn:ready` state pointing at one of these deliverables.

**Depends on:** Artifact 4.

---

### Artifact 11 — Synthetic Customer Scenarios (size: M, P1)

**Path:** `/Users/aliai/logicigniter/business/customer-scenarios/`

**Purpose:** Pre-launch, there are no real customers. Without external
signal, prioritization is hypothetical. Synthetic personas exercising
the apps give CRO/CMO/CCO a target.

**Content:** 10 personas (one per suite), each with: name, company
size, pain, expected workflow across the relevant apps, success
criteria. These become test cases QA can run against, demo scripts CMO
can write narratives for, and ICP candidates CRO can validate against.

**Acceptance:** Each suite has at least one persona file; CRO ICP docs
(Artifact 10) reference them.

**Depends on:** Artifact 10.

---

### Artifact 12 — One Controlled Drill (size: S, P0 gating)

**Purpose:** Prove the whole loop works end-to-end before re-enabling
broader autonomy.

**Steps:**
1. Pick one `zehn:ready` issue in `scripts` (low-risk, shell change) or
   one Go svc-* repo.
2. Enable only Artifact 7's 4 cron jobs.
3. Let li-specialist-sweep select and delegate.
4. Watch one specialist claim, branch, verify (Artifact 1), PR.
5. Watch reviewer routing (Artifact 8) trigger and complete.
6. Watch verify pass, merge gate, post-merge reconcile.
7. Watch Yaad write of terminal outcome (Artifact 9).
8. Watch COO scoreboard (Artifact 6) update.

**Acceptance:** One issue traverses `zehn:ready → in-progress →
verified → reviewed → merged → reconciled → recorded`. Gateway log
shows ≥2 Yaad MCP successes. Scoreboard shows the increment.

**Depends on:** all prior artifacts.

---

## What Gets Deleted Or Archived

These are removed deliberately, not by accident, because they
contribute to the loop-of-monitoring pattern:

**Cron jobs deleted from `jobs.json`** (per Artifact 7):
- 10 area work-queue jobs (architect, backend, frontend, ux,
  integration, data-ai, devops, qa, security, docs)
- daily training
- CEO hourly check, COO hourly check, engineering 30-min check
- GitHub control-plane reconciler hourly
- personal hourly check
- second zehn-operations-monitor slot

**Files archived to `.picoclaw/workspace/memory/archive/`:**
- `ZEHN_SETUP_PLANNING.md` (1219 lines, historical)
- `ZEHN_READINESS_AUDIT.md` (772 lines, stale facts)
- `LOGICIGNITER_SOLUTION_PORTFOLIO_PLAN.md` (contains stale "app owners
  remain responsible")
- `LOGICIGNITER_OPERATING_CADENCE.md` v1 (replaced by Artifact 5)

**Workspace draft files archived to per-agent `archive/`:**
- All `*-issue-*-body.md`, `tmp-pr-*-body.md` per the prior F-027
  finding.

**Delegations directory** (`.picoclaw/workspace/delegations/`, 1186
files, 9.1 MB): keep on disk for forensic value; move files older than
14 days to `.picoclaw/workspace/delegations/archive/YYYYMM/`.

**Gateway log rotation:** add `logrotate`-equivalent script to roll
`gateway.log` at 50 MB to `gateway.log.1`, .2, .3, max 3 rolls.

---

## Order Of Implementation

I will build in this order. Each step is verified before the next.

**Week 1 — Unblock the loop**
1. Artifact 1: `verify-pr.sh` v1
2. Artifact 2: GitHub Actions workflows
3. Artifact 9: Yaad enforcement in AGENT.md files
4. Artifact 5: Operating Cadence v2 (replace existing doc)

**Week 2 — Define the motion**
5. Artifact 3: Release Readiness Ladder
6. Artifact 4: Work Selection Algorithm
7. Artifact 6: COO Scoreboard
8. Artifact 7: New cron job set (file change only, kept disabled)

**Week 3 — Activate the rest of the company**
9. Artifact 10: Non-eng role deliverable specs
10. Artifact 11: Synthetic customer scenarios
11. Artifact 8: Event-driven dispatch rules

**Week 4 — Drill**
12. Artifact 12: One controlled drill, then expand autonomy gradually

Total: roughly 4 weeks of focused work. No more cron jobs added. No
more roles added. Only what ships.

---

## Hard Rules I Will Follow

- No Zehn restart, no cron enable, no Discord, no external channel
  until you explicitly approve.
- No edits to existing files until you approve each artifact.
- No invented config keys (audit confirmed every current key against
  source; same rule applies to additions).
- Delete and simplify before adding. Every new doc must displace an
  old one, not stack on top.
- No Go code changes unless an artifact's acceptance proves config /
  prompt / script cannot satisfy it.
- Every LI repo I touch ends clean.
- Secrets stay in `.security.yml` / `.env`, never in `config.json` or
  committed docs.
- Every artifact has an acceptance criterion that can be checked.

---

## What I'm Asking For Right Now

One decision: **approve me to build Artifact 1 (`verify-pr.sh` v1).**

That's the unblock for everything else. It's a single shell script in
one LI subrepo (`/Users/aliai/logicigniter/scripts/`). It costs me a
few hours. Acceptance is testable in 10 minutes after it's written.

I will not start Zehn. I will not enable cron. I will not touch
anything else in this approval round. Once Artifact 1 lands and you
verify it, I'll ask for Artifact 2 and proceed in the order above.

If you'd rather I batch the approvals (e.g., approve Week 1 as a
unit), say so and I'll proceed that way instead. But after weeks of
audit cycles I think the most useful thing I can do is ship one real
artifact today and let you see it work.
