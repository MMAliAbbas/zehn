# Phase 2 — Doctrine Reconciliation

## 2.1 — HEARTBEAT_OK canonical file + supersession markers

Canonical file: `workspace/memory/LOGICIGNITER_HEARTBEAT_ACCEPTANCE_CRITERIA.md` (6 scenarios: 0 public-site probe gate, 1 ready issue / no owner, 2 PR green but unmerged, 3 completed work / no successor, 4 GitHub or company state unavailable, 5 no work truly exists — only Scenario 5 permits the literal `HEARTBEAT_OK` token).

Files edited to remove rule restatements and add a pointer (in addition to the 7 named in the recovery plan, 3 more were found during verification):

| File | Edit |
|---|---|
| `workspace/operating-prompts/zehn-operations-monitor.md` | Removed L94–105 fail-closed visibility rule body and L99–103 YAAD_DEGRADED MEMORY.md mandate; replaced with pointer to canon + Yaad-retry guidance |
| `workspace/operating-prompts/logicigniter-ceo-daily-sync.md` | Replaced L78–82 HEARTBEAT_OK rules with pointer; removed L18–23 stale First-Run Condition |
| `workspace/operating-prompts/logicigniter-nonexec-weekly-pulse.md` | Replaced L113–114 unconditional invalidation with pointer; updated Yaad class list to canonical 8 |
| `workspace/memory/LOGICIGNITER_OPERATING_CADENCE.md` | Removed L72–75 public-site-probe gate restatement and L302–322 "What HEARTBEAT_OK Means" block; replaced with pointer |
| `workspace/memory/LOGICIGNITER_COMPANY_UTILIZATION_CONTRACT.md` | Removed L83–95 utilization-specific invalidation list; replaced with pointer |
| `workspace/memory/ZEHN_CURRENT_STATE.md` | Replaced L100–102 with pointer |
| `workspace/memory/ZEHN_OPERATING_CADENCE.md` | Added file-level "SUPERSEDED 2026-06-04" marker at top |
| `workspace/memory/LOGICIGNITER_COMPANY_OPERATING_CONTRACT.md` (extra, not in plan) | Replaced L56–57 rule restatement with pointer |
| `workspace/memory/LOGICIGNITER_WORK_SELECTION.md` (extra, not in plan) | Reframed L588–594 no-work-found rule as supplement to canon, with explicit pointer |
| `workspace/operating-prompts/logicigniter-ceo-operating-check.md` (extra, not in plan) | Removed L24–31 inline "do not return HEARTBEAT_OK" rules and L102–110 No-Action Report invalidation list; pointed both to canon |

Files NOT edited (passing references only, not rule restatements):
- `workspace/memory/LOGICIGNITER_GITHUB_CONTROL_PLANE.md` L358 (single passing ref)
- `workspace/memory/LOGICIGNITER_YAAD_SCHEMA_CONTRACT.md` L146–147 (single passing ref)
- `workspace/memory/LOGICIGNITER_TERMINAL_STATE_MACHINE.md` L74 (passing ref, leaves room for caller-allowed quiet)
- `workspace/operating-prompts/personal-operating-check.md` L36–37 (Personal agent domain, different scope)
- Scoreboard files (`20260528.md`, `20260603.md`, `20260604.md`) and archived `SCOREBOARD_LATEST_20260604_pre_symlink.md` — these are historical records with embedded annotations, not doctrine; archival semantics preserved.

## 2.2 — Yaad memory-class list (canon = 8 classes)

Canonical list: `fact`, `decision`, `summary`, `note`, `runbook`, `best_practice`, `anti_pattern`, `architecture_decision` (from `LOGICIGNITER_YAAD_SCHEMA_CONTRACT.md`). Verified against Yaad MCP `tools/list` during Phase 1.1 — all 8 are accepted; 7 others (`status`, `operational_event`, `blocker`, `handoff`, `constraint`, `profile`, `artifact`) are rejected.

`logicigniter-nonexec-weekly-pulse.md` L94–96 was the only prompt with a narrower 3-class list; updated to the canonical 8.

## 2.3 — Terminal state machine cleanup

`DISPATCHED_AND_SUMMARIZED` added to `LOGICIGNITER_TERMINAL_STATE_MACHINE.md` as a valid terminal token used by the bounded-coordinator-cron pattern (currently only `logicigniter-nonexec-weekly-pulse.md`). Now in exactly 2 places: the state machine doc and the consuming prompt.

## 2.4 — Brittle reference cleanup

- Hardcoded port `18790` in `zehn-operations-monitor.md` L43 — replaced with reference to `config.json:gateway.port`.
- `/Users/aliai/zehn/operations/logicigniter-work-queue-scan.py` (referenced in `logicigniter-ceo-daily-sync.md` L52) — verified to exist (21 KB, executable, 2026-06-04 19:43 mtime). Reference left in place; the path is a valid hard dependency of the queue-control rule. **Open follow-up**: this path is outside the audited workspace tree and would silently break the CEO Daily Sync if moved. Consider relocating the script into the runtime home or adding a doctrine pointer that this is a hard dependency.

## 2.5 — Stale prompt + reports hygiene

- `workspace/operating-prompts/logicigniter-coo-work-selection.md.bak-liveprobe-20260523T2350Z` → `workspace/operating-prompts/archive/`
- `workspace/reports/ZEHN_AUTONOMY_CONTROL_PLANE_REPAIR_PLAN_20260511.md` → `workspace/reports/archive/`
- `workspace/reports/ZEHN_IMPLEMENTATION_PERFORMANCE_REVIEW_20260511.md` → `workspace/reports/archive/`
- `workspace/reports/ZEHN_RUNTIME_AUDIT_20260511.md` → `workspace/reports/archive/`

## Phase 2 acceptance
- [x] HEARTBEAT_OK doctrine: canonical file chosen; 10 files updated to point to it; remaining occurrences are passing references or archived historical records.
- [x] Yaad memory-class list: 8 canonical classes consistent across prompts and the schema contract.
- [x] `DISPATCHED_AND_SUMMARIZED` registered in terminal state machine.
- [x] Hardcoded port `18790` removed; hardcoded script path left in place but verified to exist (and flagged for relocation).
- [x] 4 stale files moved to archive subdirs.

Phase 2 complete.
