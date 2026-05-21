# Zehn Autonomy Recovery Master Plan

Status: design plan only. Do not treat this file as an implemented change.
Date: 2026-05-21

## Purpose

Recover Zehn from fragile prompt-driven autonomy into a reliable always-on
operating system for Ali, LogicIgniter, and future organizations/projects.

This plan exists because the previous approach mixed prompts, cron, heartbeat,
delegation, GitHub work, runtime monitoring, Yaad memory, and Discord reporting
without a strong control-plane boundary. That made failures hard to isolate and
encouraged repeated restarts, prompt edits, and status summaries instead of
root-cause fixes.

The goal is not to make Zehn "busier." The goal is to make Zehn reliably:

- know what work exists;
- know who owns it;
- start only bounded work;
- observe whether work completed;
- preserve memory locally and in Yaad;
- surface stuck work before Ali has to ask;
- keep GitHub/repos clean;
- recover from provider, tool, Yaad, Discord, or process failures.

## Evidence Baseline

These are facts observed from the current runtime and repository. They are not
assumptions.

- Gateway process can remain alive while operational readiness is false:
  `/health` returned OK while `/ready` returned HTTP 503.
- `gateway.log` stopped advancing at `2026-05-21 12:08:07`.
- `heartbeat.log` shows heartbeat firing regularly until
  `2026-05-21 12:03:36`, then no terminal heartbeat result after that.
- At `2026-05-21 12:04-12:07`, heartbeat/company-control work was inside
  delegated `li-ceo -> li-coo -> specialist` activity.
- At `2026-05-21 12:08:07`, `li-qa` hit provider stream failure:
  `codex API call: stream error: stream ID 611; INTERNAL_ERROR`.
- At `2026-05-21 12:15`, cron started `zehn-operations-monitor`; its persisted
  state later had `lastRunAtMs` but no `nextRunAtMs`, consistent with an
  in-flight or non-finalized cron run.
- `pkg/heartbeat/service.go` runs heartbeat synchronously through the configured
  handler and has no explicit single-flight guard, timeout, terminal-state
  record, or stale-run reconciliation.
- Cron agent-turn jobs and heartbeat both route into agent/provider/tool
  execution paths and can overlap.
- Durable delegation records exist in running or failed states after parent
  work failed, which means lifecycle reconciliation is incomplete.
- The runtime had recent large source changes in the agent/provider/streaming
  path and local delegation/internal-channel path; this is a regression window,
  not proof of one exact cause.

## Non-Negotiable Operating Rules

These rules must constrain every implementation phase.

1. No fix without evidence.
   - Every code/config/prompt change must point to a specific observed failure,
     source path, or reproducible test.

2. No prompt dumping.
   - Existing prompt and memory files must be rewritten or surgically edited as
     coherent artifacts. Do not append broad instruction blocks on top of stale
     text.

3. Heartbeat must stay reliable.
   - Heartbeat is the liveness and operating-awareness pulse. It must not become
     an unbounded company-management job.

4. Company work must be bounded.
   - LogicIgniter operating checks may delegate, but each check must have a
     lease, timeout, owner, and terminal state.

5. Cron is not the primary autonomy mechanism.
   - Cron may run scheduled anchors, but continuous operating awareness should
     come from heartbeat plus a durable control plane.

6. Zehn-main is supervisor, not CEO.
   - `zehn-main` watches runtime health and routes organization checks. It must
     not perform LogicIgniter executive or implementation work directly.

7. CEO owns company priority.
   - `li-ceo` decides company priority and delegates to COO/CTO/CPO/other roles.

8. COO owns execution flow.
   - `li-coo` manages WIP, claims, PRs, stuck work, successor work, and dispatch.

9. Specialists execute role-matched work.
   - Backend/frontend/QA/security/devops/docs/etc. choose work by issue contract,
     not by scanning random repos from scratch.

10. Repos must never be left dirty.
    - Dirty repos are a control-plane failure, not a normal end state.

11. Yaad is canonical, local memory is required fallback.
    - Zehn must operate when Yaad is down and sync later when Yaad returns.

12. Discord is visibility, not internal truth.
    - Discord summaries are human-facing signals. Durable state belongs in local
      records, GitHub, and Yaad.

13. No silent failure.
    - If a role cannot perform its duty, it must record why, assign an owner or
      retry condition, and do something safe/useful if possible.

## Current Architectural Problem

The current system has useful parts, but the boundaries are not strong enough.

Heartbeat currently says "light control loop," but it still delegates one sync
company operating check to CEO every cycle. CEO may delegate to COO, and COO may
delegate to specialists. In practice, that turns heartbeat into an unbounded
tree of agent turns unless every layer voluntarily stays small.

Cron jobs add more agent turns on top of that. Delegations can fail, provider
streams can fail, and tool runs can take minutes. When any of these paths does
not terminally reconcile, the system can look alive but stop responding.

The missing architecture is a real control plane:

- explicit work leases;
- active run registry;
- single-flight guards;
- timeouts;
- terminal state reconciliation;
- local/Yaad memory queue;
- role-owned work selection;
- runtime-vs-business separation;
- observable start/end/failure records for heartbeat, cron, delegation, and
  provider/tool runs.

## Target Architecture

### 1. Runtime Supervisor Layer

Owner: `zehn-main`

Responsibilities:

- verify gateway liveness and readiness;
- detect stuck heartbeat, cron, delegation, provider, and tool runs;
- check Yaad reachability and local sync backlog;
- check Discord delivery health;
- check dirty runtime/repo guardrails through read-only inspection;
- route bounded organization operating checks;
- never implement LogicIgniter work directly.

Key artifact:

- local runtime state ledger under Zehn home, with current active runs and last
  terminal states.

### 2. Heartbeat Layer

Owner: heartbeat service plus `zehn-main`

Responsibilities:

- fire at configured interval;
- create a heartbeat run record before work starts;
- refuse overlap if previous heartbeat is still active;
- enforce max duration;
- record terminal state: OK, ACTION, BLOCKED, FAILED, STALE;
- trigger at most one bounded supervisor action per cycle;
- never hide failed required checks behind `HEARTBEAT_OK`.

Important design decision:

Heartbeat may ask CEO for a company check, but it must not wait forever for a
deep delegation tree. It should create or renew a bounded company-check lease
and then observe its lifecycle.

### 3. Organization Control Plane

Owner: `li-ceo` and `li-coo`

Responsibilities:

- maintain active initiatives;
- map initiatives to GitHub scopes and repos;
- decide priority;
- keep WIP floor and ceiling;
- identify idle/stuck/completed work;
- create successor work only when a prior unit reaches terminal state;
- escalate to Ali only for approval boundaries.

Key artifact:

- `LOGICIGNITER_ACTIVE_INITIATIVES.md`, or a structured successor to it, must
  become a concise control-board artifact, not a long narrative dump.

### 4. GitHub Work Queue Layer

Owner: `li-coo`, specialists, reviewer roles

Responsibilities:

- GitHub issues are executable work units.
- Issues must carry labels, scope, acceptance criteria, verification command,
  risk, approval status, and owner/review requirements.
- Specialists only claim matching ready issues.
- Claims must be leases, not permanent labels with no timeout.
- PRs must move through review, merge, and post-merge reconcile.
- Completed work must either close the initiative lane or create successor work.

### 5. Execution Layer

Owner: specialist agents

Responsibilities:

- inspect target repo first;
- use correct repo context before mutation;
- branch from current main or issue branch as specified;
- run the verification wrapper or a documented fallback;
- commit, push, PR, review, and leave repo clean;
- report blockers with exact command/path/evidence.

### 6. Memory Layer

Owner: all agents, supervised by `zehn-main`

Responsibilities:

- write durable facts and changed state to Yaad when reachable;
- write local pending memory when Yaad is unavailable;
- sync local pending memory to Yaad later;
- deduplicate or update durable memories instead of adding repeated summaries;
- reject unsupported Yaad schema keys/classes before runtime use.

### 7. Visibility Layer

Owner: Discord + org UI

Responsibilities:

- show current state, not just historical logs;
- show agent active/idle/failed with reason;
- show inbox/outbox/delegations/meetings;
- show run leases and stale runs;
- make failures clickable to evidence.

This layer must not be the source of truth.

## Phased Implementation Strategy

Each phase must be independently testable and reversible. Do not start a later
phase until the prior phase passes its verification gate.

### Phase 0: Freeze And Baseline

Goal: preserve evidence and stop making blind changes.

Scope:

- no runtime behavior change;
- collect current config, heartbeat, cron, delegation, logs, sessions, and git
  status;
- create a single timestamped evidence bundle;
- identify the exact last-good and first-bad windows for heartbeat behavior.

Verification:

- evidence bundle exists;
- includes config hash, binary hash, git commit, heartbeat log tail, gateway log
  tail, cron jobs state, active process list, ready/health output, recent
  delegation states, and recent session files;
- no source/config/prompt files changed.

Exit criteria:

- we can state the regression window without relying on memory.

### Phase 1: Runtime Lifecycle Guardrails

Goal: prevent heartbeat/cron/delegation from silently wedging the system.

Scope:

- add heartbeat single-flight protection;
- add heartbeat run start/end/fail records;
- add heartbeat max-duration handling;
- add stale heartbeat detection;
- ensure cron job state finalizes or records failure if agent turn does not
  return within its allowed runtime;
- ensure provider/tool failures terminally mark parent delegation records.

Verification:

- unit tests prove overlapping heartbeat attempts are skipped or recorded;
- unit tests prove handler timeout produces failed terminal state;
- unit tests prove failed delegated turn updates local delegation state;
- manual test with a fake slow handler does not wedge next heartbeat forever.

Exit criteria:

- heartbeat cannot silently overlap itself or remain active forever.

### Phase 2: Runtime And Company Work Separation

Goal: make heartbeat reliable while still triggering company awareness.

Scope:

- keep heartbeat as supervisor pulse;
- introduce a bounded company-check lease rather than a deep synchronous wait;
- `zehn-main` may request one company check if none is active or overdue;
- CEO/COO work proceeds under its own lease and terminal record;
- heartbeat reports stale company check as a failure, not by recursively doing
  the company work itself.

Verification:

- a heartbeat cycle with no active company work creates exactly one company
  check lease;
- a second heartbeat while company check is active does not start another one;
- a stale company check is reported with owner/evidence;
- heartbeat returns OK only when runtime is healthy and company check state is
  terminal/non-actionable.

Exit criteria:

- heartbeat drives awareness without becoming the work executor.

### Phase 3: Delegation Lifecycle Reconciliation

Goal: make every delegation terminal, inspectable, and recoverable.

Scope:

- each delegation has parent, target, task, lease, due time, status, error,
  terminal evidence, and optional Yaad memory ID;
- async executor capacity must be visible and configurable;
- failed provider/tool calls mark the relevant delegation failed or blocked;
- stale running records are reconciled on startup and by supervisor check;
- UI and status tools expose reason, not only status.

Verification:

- tests cover sync success, sync provider failure, async accepted, async
  capacity full, target panic/error, stale reconciliation;
- status tool shows failure reason;
- UI can display failure reason without parsing gateway logs.

Exit criteria:

- no long-lived `running` delegation can exist without lease metadata and
  escalation behavior.

### Phase 4: Local Memory Queue And Yaad Sync

Goal: make memory useful even when Yaad is down.

Scope:

- local memory write-ahead queue for durable facts/state changes;
- schema validation before enqueue;
- sync worker with retries and dedupe/update behavior;
- clear distinction between local runtime context and durable Yaad memory;
- visible backlog and last sync state.

Verification:

- Yaad unavailable: memory is queued locally and agent receives explicit status;
- Yaad restored: queued memory syncs and records Yaad IDs;
- duplicate summary attempts update or skip rather than create junk;
- unsupported schema is rejected locally before MCP call.

Exit criteria:

- Yaad outage does not make Zehn forget or spam invalid memories.

### Phase 5: LogicIgniter Company Control Board

Goal: make LogicIgniter operate like a company, not as repo scanning loops.

Scope:

- active initiatives are the top-level units;
- CEO owns initiative priority;
- COO owns execution flow;
- CTO/CPO own technical/product direction;
- specialists execute role-matched issue work;
- control board records state, WIP floor, owner, next decision, blockers,
  current PRs/issues, and successor rule.

Verification:

- CEO can list active initiatives without scanning every repo;
- COO can find ready/stuck/completed work against initiatives;
- completion of one work unit either updates initiative terminal state or
  creates/requests successor work;
- no role returns `HEARTBEAT_OK` when a known initiative has unowned ready work.

Exit criteria:

- Ali should not need to ask "what next?" for an active initiative.

### Phase 6: GitHub Execution Contract Hardening

Goal: make autonomous coding reliable without restricting useful authority.

Scope:

- issue template/contract for executable work;
- claim lease with timestamp and owner;
- branch naming by issue number;
- standard verification wrapper;
- PR review/merge/reconcile flow;
- dirty repo detection and cleanup;
- successor issue rule.

Verification:

- simulated ready issue is claimed by correct specialist;
- wrong specialist skips with reason;
- blocked issue is not claimed;
- PR green/reviewed moves to merge/reconcile owner;
- dirty repo creates a blocker and cleanup owner.

Exit criteria:

- specialists can safely move one issue through issue -> branch -> PR -> review
  -> merge/reconcile without Ali driving the sequence.

### Phase 7: UI And Observability

Goal: make the system understandable while it runs.

Scope:

- org screen shows current active run state;
- cards show idle/working/failed/stale with reason;
- inbox/outbox/delegations/meetings are clickable;
- runtime health panel distinguishes process health, readiness, heartbeat,
  cron, provider, Yaad, Discord, GitHub, dirty repos;
- logs are supporting evidence, not primary state.

Verification:

- UI displays a synthetic failed delegation with reason;
- UI displays an active company-check lease;
- UI displays stale heartbeat/company check;
- no browser-only state is needed for recovery.

Exit criteria:

- Ali can see why a role is stuck without asking Discord.

### Phase 8: Multi-Organization Generalization

Goal: support more organizations/projects without LogicIgniter-specific leaks.

Scope:

- organization registry;
- per-organization active initiatives;
- per-organization CEO/COO/etc. mapping;
- shared supervisor heartbeat;
- organization-scoped Yaad memory;
- organization-scoped GitHub/workspace/project roots.

Verification:

- LogicIgniter still works unchanged;
- a second organization can be configured without copying LogicIgniter prompts;
- heartbeat routes each org through its own control lease.

Exit criteria:

- adding another org is configuration plus role artifacts, not code surgery.

## Implementation Order

Do not implement all phases at once.

Recommended order:

1. Phase 0: freeze and baseline.
2. Phase 1: lifecycle guardrails.
3. Phase 2: heartbeat/company separation.
4. Stop and run a 24-hour observation window.
5. Phase 3: delegation reconciliation.
6. Phase 4: local memory queue and Yaad sync.
7. Phase 5 and 6: LogicIgniter company execution.
8. Phase 7: UI.
9. Phase 8: multi-org.

The 24-hour observation after Phase 2 is mandatory. If heartbeat still wedges,
do not proceed to company-control changes.

## Testing Strategy

Required test classes:

- unit tests for heartbeat single-flight and timeout;
- unit tests for cron finalization on handler timeout/error;
- unit tests for delegation terminal states;
- integration tests with fake provider failures;
- integration tests with fake Yaad unavailable/restored;
- local live verification with Discord disabled/enabled in stages;
- 24-hour soak with heartbeat active and cron minimized.

No phase is complete because "logs looked okay." Each phase needs explicit
commands and expected output.

## Rollback Strategy

Each phase must be reversible.

- Runtime behavior changes behind config flags until proven.
- Prompt/control-board changes committed separately from Go code changes.
- No migration deletes historical records.
- Stale records can be archived but not silently discarded.
- If heartbeat reliability worsens, revert Phase 1/2 before touching higher
  layers.

## What Not To Do

- Do not add more cron jobs to compensate for broken heartbeat.
- Do not add more text to every agent persona.
- Do not restart repeatedly and call that recovery.
- Do not make `zehn-main` act as CEO.
- Do not make Discord the internal delegation bus.
- Do not hide provider failures behind `HEARTBEAT_OK`.
- Do not let every heartbeat scan every repo.
- Do not make Yaad availability a hard dependency for basic operation.
- Do not implement UI polling of logs as the source of truth.

## First Concrete Work Package

The first implementation package should be Phase 0 and Phase 1 only.

It should produce:

- one evidence bundle command/script;
- heartbeat run state records;
- heartbeat single-flight guard;
- heartbeat timeout/failure terminal state;
- tests proving overlap and timeout behavior;
- no LogicIgniter prompt rewrites;
- no GitHub work-flow changes;
- no UI changes.

Reason: if the heartbeat lifecycle is not reliable, every higher-level company
operating feature will remain fragile regardless of how good the prompts are.

