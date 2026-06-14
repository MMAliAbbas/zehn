# Zehn Recovery Plan — 2026-06-04

Sibling doc: `ZEHN_FORENSIC_AUDIT_20260604.md` (read first).
Status: Zehn is **frozen** (launchctl bootout completed 22:54 +05; PIDs 79662/79664 gone; plist intact). No cron will fire until re-bootstrap.

---

## Operating principles for this recovery (read once, enforce always)

1. **No bulk scripts that rewrite artifacts.** Every prompt, memory, delegation, or Yaad write is a single deliberate human-reviewed action. The mess was caused by agents generating boilerplate at scale and other agents trusting it.
2. **Distinguish 7 levels of done**: configured / enabled / reachable / authorized / called / succeeded-once / live-proven-recurring. Never collapse them into "fixed" / "working".
3. **No "resolved" / "remaining_issues: []" claim without live evidence in the same session.** The 22:15:59 Yaad update is the cautionary case.
4. **No hand-editing of delegation JSON to make `delegation_status` look clean.** Either the run produced a terminal result or it didn't.
5. **Keep what works.** Don't tear down `77d13f90`, the Yaad token plumbing, the Discord channel wiring, the GitHub artifact creation flow, the cron scheduler, or the agent dispatch routing — all of these are reachable today.
6. **Strip what contradicts.** Doctrine fragmentation (8+ HEARTBEAT_OK rule sets) is the biggest single drag — pick one canon and supersede the rest.
7. **Logs and delegation records are evidence.** Don't rotate or delete them in the same session as the change you're trying to verify. Preserve.
8. **End-to-end is the only proof.** Code-tests and unit-tests do not count for "live-proven".

---

## Phase 0 — Freeze & Preserve  *(DONE 2026-06-04 22:54 +05)*

- [x] `launchctl bootout gui/501/io.picoclaw.launcher` — agent unloaded, no respawn
- [x] PIDs 79662 (launcher) and 79664 (gateway) confirmed gone
- [x] Plist `~/Library/LaunchAgents/io.picoclaw.launcher.plist` left intact for later re-bootstrap
- [x] Forensic audit saved to `audit-20260604/ZEHN_FORENSIC_AUDIT_20260604.md`
- [x] This plan saved to `audit-20260604/ZEHN_RECOVERY_PLAN_20260604.md`

---

## Phase 1 — Truth Reset (durable data layer)

Goal: undo the misleading durable state so future reads of memory/delegations don't carry forward today's overclaims. **Highest priority.** Do this before touching prompts or code — otherwise the lies become "consensus".

### 1.1 Yaad durable memory cleanup
- [ ] **Inventory** every Yaad memory entry written by Zehn agents on 2026-06-04 (UTC range covers both sessions). Use `mcp_yaad_memory_browse` filtered by `metadata.date: "2026-06-04"` and by `labels` containing `zehn-monitor`, `runtime-observability`, `runtime-health`, `resolved-blocker`. Do this from a **non-agent** session (CLI/MCP client) so we don't trigger a new agent cycle that writes more.
- [ ] For each entry, classify: **truthful** / **overclaim** / **out-of-date**.
- [ ] Specifically reconcile id `9de0d453-3b45-47d8-9272-16a2ba72d133`:
  - Current state (per 22:15:59 update): `failure_class: runtime-observability-degraded`, `actionable: false`, `remaining_issues: []`, `resolved_symptom: "...; Yaad reachable; delegation_status visibility working with stale lanes terminally resolved."`
  - Reality at that time: Pico WebSocket origin rejections ongoing; `memory_update` failed 3/4 times in the prior 4h; `mcp_yaad_memory_query` failed at 21:45:52.
  - **Decision required from Ali**: (a) update the entry to the truthful state (`actionable: true`, `remaining_issues: ["pico WebSocket origin rejected", "yaad memory_update intermittent (75% fail)", "mcp blank-error propagation"]`), or (b) delete it. Recommendation: (a), so the audit trail of the lie is preserved.
- [ ] Reconcile id `zehn-monitor:runtime-blockers-resolved 2026-06-04 21:15 +05` (memory_add at 21:16:12) — same overclaim shape. Same decision pattern.

### 1.2 Delegation record reconciliation
- [ ] Re-read both hand-edited records and decide formal disposition:
  - `delegation-20260602T030622.167596000Z-bae902771a5a.json` — current `status: failed`/`type: stale_superseded`. Acceptable IF we explicitly tag `manually_closed_by: "recovery-session-20260604"` so future readers know. Edit the field in place (small, reviewable). Do NOT close it as if it had a real terminal result.
  - `delegation-20260604T104857.649096000Z-6f5980bb85bb.json` — `status: completed`, GitHub artifacts real. Same treatment: add `manually_closed_by` field. Result content already discloses "Recovered stale delegation after audit." — fine.
- [ ] **Sample 20 random delegation records** across the 4,325 to check whether the manual-edit pattern is broader than 2 cases. Look for `provider: local-cleanup` / `provider: local-recovery` / "Recovered ... after audit." strings.
  - If pattern is isolated: noted, move on.
  - If pattern is widespread: escalate; we may need a separate audit pass before continuing.

### 1.3 Local ledger truthing
- [ ] `workspace/memory/MEMORY.md` — L27–37 are ledger-style entries that violate L8's own "fallback only, not historical ledger" posture. Decide: enforce L8 (delete L27–37) OR change L8's posture rule. Pick one. Recommendation: enforce L8, move L27–37 content to `archive/MEMORY_FALLBACK_HISTORY_20260604.md` for forensic preservation.
- [ ] `workspace/memory/LOGICIGNITER_OPERATING_CYCLE_LEDGER.md` (151 KB / 977 lines) — keep only the **Current Cycle** section (top ~30 lines). Move L71–977 verbatim to `archive/LOGICIGNITER_OPERATING_CYCLE_LEDGER_20260604_snapshot.md`. The ledger is a "live control" doc; treating it as append-only is what created the clutter.
- [ ] `workspace/memory/LADDER_SNAPSHOT_LATEST.md` — reconcile L1–11 (says "recovered/superseded") with L15 (still says "Yaad read failed before filesystem scan…"). Pick one. If snapshot is no longer authoritative, move it to `archive/` and replace with a 5-line pointer.

### 1.4 Scoreboard truthing
- [ ] Confirm `scoreboard/20260604.md` is the actual generated scoreboard from the 08:30 cron.
- [ ] Replace `scoreboard/LATEST.md` (regular file with disclaimer) with a **symlink** → `20260604.md` (or whatever the last truthful dated file is). One single `ln -sf` op.
- [ ] Note the May 30 – Jun 2 gap in a one-line entry in `scoreboard/README.md` (create if absent). Don't fabricate the missing days.

**Phase 1 acceptance:** A read of `MEMORY.md`, the ledger header, `LATEST.md`, and Yaad memory id `9de0d453-...` all tell the same story about runtime state. If they don't, Phase 1 isn't done.

---

## Phase 2 — Doctrine Reconciliation (prompts + memory contracts)

Goal: one canonical answer to "what is HEARTBEAT_OK?", "what Yaad classes are valid?", "what's a terminal state?". This is human edit work, not agent.

### 2.1 Pick canonical files; explicitly supersede the rest
- [ ] **HEARTBEAT_OK canon**: pick `workspace/memory/LOGICIGNITER_HEARTBEAT_ACCEPTANCE_CRITERIA.md` as the single source (it's scenario-keyed and structured). In each of the following files, add a **3-line "supersedes" marker** at the top pointing to the canon. Do not duplicate or restate the rules in these files:
  - `workspace/operating-prompts/zehn-operations-monitor.md` (remove L94–105 fail-closed rule body; replace with pointer)
  - `workspace/operating-prompts/logicigniter-ceo-daily-sync.md` (remove L78–82)
  - `workspace/operating-prompts/logicigniter-nonexec-weekly-pulse.md` (remove L113–114)
  - `workspace/memory/LOGICIGNITER_OPERATING_CADENCE.md` (remove L72–75 public-site-probe gate + L302–322 "What HEARTBEAT_OK means" block — leave a one-line "See LOGICIGNITER_HEARTBEAT_ACCEPTANCE_CRITERIA.md")
  - `workspace/memory/LOGICIGNITER_COMPANY_UTILIZATION_CONTRACT.md` (remove L83)
  - `workspace/memory/LOGICIGNITER_TERMINAL_STATE_MACHINE.md` (remove L74 pointer)
  - `workspace/memory/ZEHN_CURRENT_STATE.md` (remove L100–102)
  - `workspace/memory/ZEHN_OPERATING_CADENCE.md` (May 21 file) — mark **entire file** as superseded by `LOGICIGNITER_OPERATING_CADENCE.md` at the top; move to `archive/` after one week if no agent reads it.
- [ ] Decide if the canonical `LOGICIGNITER_HEARTBEAT_ACCEPTANCE_CRITERIA.md` itself needs trimming. Read it first; only edit if it's inconsistent internally.

### 2.2 Yaad memory-class list — pick one
- [ ] Canonical Yaad class set should be the **broad 8** (matches what zehn-operations-monitor.md L80–82 lists): `summary`, `decision`, `fact`, `note`, `runbook`, `best_practice`, `anti_pattern`, `architecture_decision`. Confirm this matches what Yaad backend actually accepts.
- [ ] Update `logicigniter-nonexec-weekly-pulse.md` L94–96 to use the same 8 (not just `summary`/`decision`/`fact`).
- [ ] If any of the 8 classes are NOT supported by current Yaad backend, document which and remove. Verify via `mcp_yaad_scope_type_list` or vendor docs — don't assume.

### 2.3 Terminal state machine cleanup
- [ ] Either add `DISPATCHED_AND_SUMMARIZED` to `LOGICIGNITER_TERMINAL_STATE_MACHINE.md` as a valid terminal token, OR remove it from `logicigniter-nonexec-weekly-pulse.md`. Pick one; don't have it in only one place.

### 2.4 Strip brittle references
- [ ] Remove hardcoded port `18790` from `zehn-operations-monitor.md` L43. Replace with reference to `config.json` `.gateway.port`.
- [ ] Resolve `/Users/aliai/zehn/operations/logicigniter-work-queue-scan.py` (referenced in `logicigniter-ceo-daily-sync.md` L51–58 but path is unverified). Either (a) confirm it exists and add a 1-line pointer in a memory doc as the source-of-truth path, or (b) remove the reference. Don't leave the hardcoded path as load-bearing.
- [ ] Reconcile stale "First-Run Condition" in CEO Daily Sync (L18–23) — the sweep already ran 2026-06-04 08:00; either remove the block or gate it on a flag check.

### 2.5 Stale prompt hygiene
- [ ] Move `operating-prompts/logicigniter-coo-work-selection.md.bak-liveprobe-20260523T2350Z` → `operating-prompts/archive/`.
- [ ] Move `workspace/reports/ZEHN_AUTONOMY_CONTROL_PLANE_REPAIR_PLAN_20260511.md`, `ZEHN_IMPLEMENTATION_PERFORMANCE_REVIEW_20260511.md`, `ZEHN_RUNTIME_AUDIT_20260511.md` → `workspace/reports/archive/` (or delete if confident).

**Phase 2 acceptance:** A grep for `HEARTBEAT_OK` across `workspace/operating-prompts/` and `workspace/memory/` shows the canonical file plus pointers to it — no inline rule restatements. A grep for `DISPATCHED_AND_SUMMARIZED` returns hits in only one doctrinal location.

---

## Phase 3 — Runtime Code Fixes (in `/Users/aliai/zehn`)

Goal: fix the three concrete code-level failures the audit surfaced. Each must pass a live test, not just a unit test.

### 3.1 Pico WebSocket origin rejection
- [ ] Inspect `pkg/channels/pico/pico.go:990` — the `Upgrader.CheckOrigin` rejection that fired continuously 21:47–21:49.
- [ ] Likely fixes: (a) widen the allowed-origin list to include `http://127.0.0.1:18800` and `http://[::1]:18800`, or (b) read allowed origins from `config.json` and confirm the running browser/UI's origin matches.
- [ ] Live-prove: open Pico UI in a browser, confirm WebSocket connects, gateway.log shows no `CheckOrigin` rejections for ≥5 min.

### 3.2 MCP blank-error propagation
- [ ] Inspect `pkg/tools/registry.go:340` (and likely `pkg/mcp/manager.go:576` `sending "tools/call":`). The truncated error at 21:45:52 had no underlying cause attached.
- [ ] Fix should surface (at minimum) the transport-level error (HTTP status, network error, JSON-RPC error code) so the agent prompt can react with specificity instead of broad "Yaad failed".
- [ ] Live-prove: induce a Yaad failure (e.g., temporarily DNS-block `yaad.mmaliabbas.com`), confirm the error string now includes the underlying transport detail.

### 3.3 Yaad `memory_update` conflict handling
- [ ] Today: 3/4 updates failed (1 Bad Gateway, 2 `conflict`). Optimistic-concurrency on `expected_version` is being violated because the agent reads version N, then by the time it writes, another writer (or the same memory's auto-increment) is at N+1.
- [ ] Fix shape — pick one, don't do both:
  - **Read-modify-write retry** at the MCP client layer (max 3 retries, exponential backoff, refetch version each try).
  - OR **conflict surfaces as a structured error** (`{kind: "version-conflict", current_version: N+1}`) so the agent prompt can decide whether to retry or supersede.
- [ ] Live-prove: a cron monitor run completes a `memory_update` on `9de0d453-...` without manual intervention, with success in gateway.log.

### 3.4 Codex empty-output reconstruction (already changed warn→debug in `77d13f90`)
- [ ] No action — log noise reduction only. But verify the underlying empty-output condition isn't masking a real provider bug. Read `pkg/providers/oauth/codex_provider.go` `hydrateCodexResponseOutput`. If 8+ empty outputs in 24h, escalate to provider-side issue, don't keep silently reconstructing.

**Phase 3 acceptance:** Each fix has a one-page evidence note (`audit-20260604/CODE_FIX_<topic>.md`) with: code change diff, unit test result, **live-runtime log slice** showing the fix in action under realistic load.

---

## Phase 4 — Clutter Containment

Goal: bounded retention. Do NOT delete the forensic record; archive it where future runs won't trip over it.

### 4.1 Delegation records
- [ ] Move delegation records with `completed_at` older than 14 days into `workspace/delegations/archive/YYYY-MM/`. Single deliberate move (preserve mtimes). Confirm `delegation_status` tool reads only the live dir.
- [ ] Verify `delegation_status` after archive shows reduced count and no errors.

### 4.2 Gateway log rotation
- [ ] Stop gateway is already done. Rename `logs/gateway.log` → `logs/gateway.log.20260604-frozen` and gzip. Same for `logs/gateway_panic.log` → `logs/gateway_panic.log.20260604-frozen.gz`.
- [ ] Add a `logs/README.md` documenting that `gateway_panic.log` is misnamed — it's cron stdout, not Go panics.

### 4.3 Scoreboard
- [ ] Replace `scoreboard/LATEST.md` (regular file with disclaimer) with `ln -sf 20260604.md LATEST.md`. Confirm the dated file is the truthful content.

### 4.4 Archive stale memory docs
- [ ] Move `workspace/memory/RELEASE_LADDER_ASSESSMENT_STATUS_20260518T0603.md` → `archive/`.
- [ ] Move `workspace/memory/ZEHN_OPERATING_CADENCE.md` (May 21) → `archive/` after Phase 2.1 supersession marker is in place for at least 7 days with no agent reads.
- [ ] Move `workspace/memory/ZEHN_CURRENT_STATE.md` → either refresh fully with a Jun-4 mtime + current state, OR archive it. If keep, it MUST be updated when state changes.

**Phase 4 acceptance:** `find workspace/ -type f -mtime -7 | wc -l` shows under 100 (was 245). `du -sh logs/gateway.log*` shows live log under 5 MB. Delegation `ls workspace/delegations/ | wc -l` shows under 200 live records (was 4,325).

---

## Phase 5 — Supervised End-to-End Smoke Test

**Goal: one full canonical loop, proven, with full log evidence captured.** Acceptance for "Zehn is live-proven autonomous LogicIgniter operator" is this single test, not a checklist of capabilities.

### Loop definition
`Ali picks one small ready task` → `scanner finds it` → `zehn-main delegates to specialist` → `specialist returns terminal result with a real GitHub artifact` → `Yaad memory write succeeds` → `Discord summary visible` → `scanner state advances`.

### Procedure
- [ ] **Ali picks ONE task.** Should be a small, scoped, low-blast-radius GitHub issue with clear acceptance criteria. Recommendation: an existing labeled `area:docs` or `area:frontend` issue under 2 hours of agent work, NOT a code-deploy or money-touching task.
- [ ] **Re-bootstrap launchd in test mode**: `launchctl bootstrap gui/501 ~/Library/LaunchAgents/io.picoclaw.launcher.plist`.
- [ ] **Before re-bootstrap, disable all cron jobs except `zehn-operations-monitor-v2`.** Edit `workspace/cron/jobs.json` and set `enabled: false` on the other 5. (This is a single deliberate edit, reviewable.)
- [ ] Watch `tail -f logs/gateway.log | grep -E 'mcp_yaad|delegation|discord|tool execution|agent.turn'` live, with the smoke-test task running.
- [ ] At each loop stage, capture log slices to `audit-20260604/E2E_PROOF_<task-id>.md`:
  - scanner discovery (timestamp, evidence query)
  - delegation creation (delegation_id + initial JSON)
  - specialist turn start/end (turn_id, iterations, tool calls)
  - GitHub artifact creation (PR/issue/comment URL with verification via `gh`)
  - Yaad memory write (memory_id, version, content)
  - Discord summary published (message_id, length)
  - delegation closure (terminal status, completed_at)
  - scanner re-runs and no longer surfaces this task

### Acceptance for Phase 5
- [ ] All 7 stages above have evidence captured in the proof doc.
- [ ] No hand-edits to delegation JSON.
- [ ] No "Yaad failed" or "all resolved" prose in any LLM output.
- [ ] No `Upgrader.CheckOrigin` rejections in `logs/gateway.log` during the smoke test.

If Phase 5 fails on any stage, do not "fix forward" — return to Phase 1/2/3 for the specific stage that broke.

---

## Phase 6 — Selective Re-enable

Only after Phase 5 passes.

- [ ] Re-bootstrap launchd if not already: `launchctl bootstrap gui/501 ~/Library/LaunchAgents/io.picoclaw.launcher.plist`. Confirm with `launchctl list io.picoclaw.launcher`.
- [ ] Enable cron jobs one at a time, **24 hours between each**, watching `lastStatus` and `gateway.log`:
  1. `zehn-operations-monitor-v2` (already enabled in test mode; observe 24h)
  2. `li-daily-synthesis-v2`
  3. `li-ceo-daily-sync-v3`
  4. `li-weekly-plan-v2`
  5. `li-weekly-review-v2`
  6. `li-nonexec-weekly-pulse-v3` — last, because of unresolved LLM-timeout from 2026-05-26.
- [ ] After each enable, watch for 24h that:
  - cron `lastStatus: ok`
  - Yaad writes succeed
  - Discord messages visible
  - no new ledger / memory contradictions written

### Phase 6 acceptance
- [ ] 7 consecutive days with all 6 cron jobs green, no manual delegation JSON edits, no "resolved / remaining_issues: []" overclaims in Yaad memory.

---

## Always-on rules (the permanent DO-NOT list)

These belong in `~/.picoclaw-zehn/workspace/memory/MEMORY.md` after Phase 1.3 cleanup, pinned at the top.

1. **Never mark a task "fixed" without a live log slice proving it.** Code-test passing ≠ live-proven.
2. **Never write a Yaad memory entry claiming "resolved" or "remaining_issues: []" without verifying the runtime matrix in the same session.**
3. **Never hand-edit delegation JSON to clean up `delegation_status` output.** Either re-run, supersede, or formally tag `manually_closed_by`.
4. **Never delete or rotate logs in the same session as a behavior change.** Logs are the only way to verify the change worked.
5. **Never restate doctrine inline in a prompt.** Point to the canonical file.
6. **Never bulk-script artifact writes.** One deliberate edit at a time.
7. **Never collapse "configured/enabled/reachable/authorized/called/succeeded/live-proven" into "working".**

---

## What this plan deliberately does NOT do

- It does not refactor `pkg/agent/*` beyond the three identified bugs.
- It does not migrate Yaad, Discord, GitHub, or any external dependency.
- It does not rebuild the workspace from scratch — the prompt/memory/skill assets are mostly fine; the doctrine layer is the problem.
- It does not add new operating doctrine, new prompts, or new memory categories. Reduction-only.
- It does not promise an end date. Estimated phase order is 1 → 2 → 3 → 4 → 5 → 6. Each phase has its own acceptance gate.

---

## Phase-by-phase TODO summary (for quick scanning)

- [x] **0.** Freeze & preserve (DONE)
- [ ] **1.** Truth reset — Yaad memory, delegation records, local ledger, scoreboard
- [ ] **2.** Doctrine reconciliation — pick canonical HEARTBEAT_OK, Yaad classes, terminal states; strip brittle refs
- [ ] **3.** Code fixes — Pico WS origin, MCP blank-error, Yaad update conflicts
- [ ] **4.** Clutter containment — archive delegations, rotate logs, symlink scoreboard, archive stale memory
- [ ] **5.** Supervised E2E smoke test with full evidence capture
- [ ] **6.** Selective cron re-enable, 24h gap each, 7-day green window for acceptance
