# Zehn Runtime Artifact Audit - 2026-05-27

Scope: active runtime artifacts only. This audit intentionally excludes Go/source changes as a proposed fix path. It inspects config, cron, heartbeat, operating prompts, agent files, memory artifacts, and operational scripts.

## Current Runtime State

- At audit start, source repo `/Users/aliai/zehn` was clean and synced with origin.
- Active Zehn home is `/Users/aliai/.picoclaw-zehn`.
- Runtime lifecycle was not changed by this audit or remediation.
- No Go runtime/source files were modified for this audit or remediation.

## Evidence Summary

### 1. Active scanner creates a single-lane bottleneck

Evidence:

- `/Users/aliai/zehn/operations/logicigniter-work-queue-scan.py:521-558`
- Live read-only scanner run on 2026-05-27 returned:
  - `open_prs: 3`
  - `ready: 8`
  - `unblock_candidates: 24`
  - `next_action.type: REVIEW_PR`
  - `next_action.target: logicigniter/svc-webhookrouter-grpc#1`

Cause:

- `choose_next_action()` ranks `open_prs` before `ready`, `unblock_candidates`, and `malformed`.
- The first open PR selected is unlabeled `svc-webhookrouter-grpc#1`.
- This means the COO loop can repeatedly route one PR lane even while ready work and unblock candidates exist elsewhere.

Impact:

- This is a concrete reason Zehn appears busy but does not behave like a full company.
- It is not a Go bug.
- It is an artifact/script workflow bug: the scanner encodes a global priority rule that can starve other initiatives and roles.

### 2. Cron contains duplicate internal retry jobs for the same blocker

Evidence:

- `/Users/aliai/.picoclaw-zehn/workspace/cron/jobs.json:154-490`
- Enabled cron inventory shows 17 internal retry jobs targeting variants of:
  - `svc-webhookrouter-grpc#1`
  - `proto#11`
  - private `go-packages` CI read-path blocker

Cause:

- Prior agents repeatedly scheduled one-shot/internal retry jobs for the same blocker rather than maintaining one canonical blocker record.

Impact:

- Creates repeated work, duplicated messages, and false activity.
- Reinforces the scanner bottleneck above.
- Explains why Zehn keeps returning to the same PR/blocker lane.

### 3. Active prompts instruct a CEO/COO flow, but the scanner undermines it

Evidence:

- `HEARTBEAT.md:20-28` says heartbeat is dispatcher/watchdog and allows one control-plane action.
- `HEARTBEAT.md:52-77` delegates one async CEO operating cycle.
- `logicigniter-ceo-operating-check.md:31-73` tells CEO to pick the highest-priority active initiative/blocker and delegate execution to COO.
- `logicigniter-coo-work-selection.md:19-25` requires COO to run the deterministic scanner and perform exactly one action from `next_action`.

Cause:

- The high-level prompt model is reasonable, but COO is forced to obey a deterministic scanner whose priority ordering is too narrow.

Impact:

- The failure is not simply "bad prompts" or "too much text"; it is an interaction failure:
  - prompts delegate correctly to COO;
  - COO follows scanner;
  - scanner picks one global PR lane repeatedly.

### 4. Verification scripts check the wrong runtime home unless overridden

Evidence:

- `operations/verify-logicigniter-cron-routing.sh:4-6`
  - uses `$root/.picoclaw/config.json`
  - uses `$root/.picoclaw/workspace/cron/jobs.json`
- `operations/verify-zehn-role-personas.sh:4-5`
  - sets `HOME_DIR="$ROOT/.picoclaw"`
- Active runtime is `/Users/aliai/.picoclaw-zehn`, not `/Users/aliai/zehn/.picoclaw`.

Impact:

- Some prior verification results may have validated stale or nonexistent runtime paths unless `PICOCLAW_HOME` or equivalent context was manually supplied.
- This can create false confidence after edits.

### 5. Agent/memory context is bloated and uneven

Evidence:

- Active agent/persona/memory line counts:
  - `workspace-li-coo/memory/MEMORY.md`: 2379 lines
  - `workspace/memory/MEMORY.md`: 907 lines
  - `workspace-li-ciso/memory/MEMORY.md`: 904 lines
  - `workspace-li-frontend-developer/memory/MEMORY.md`: 487 lines
  - `workspace-li-devops/memory/MEMORY.md`: 460 lines
  - `workspace-li-ceo/AGENT.md`: 219 lines

Impact:

- This increases context load and makes agents more likely to follow stale historical state.
- The worst offender is COO memory, which is the key execution-control role.

### 6. Active memory still contains stale failure facts that may influence operation

Evidence:

- `workspace/memory/scoreboard/20260517.md` records missing release ladder path under `/Users/aliai/logicigniter/LOGICIGNITER_RELEASE_READINESS_LADDER.md`.
- `workspace/memory/scoreboard/20260522.md` records `invalid memory_class "profile"` and Docker-path failures.
- `workspace/memory/MEMORY.md` and several role memories contain old Docker/local-stack decisions.
- `workspace/memory/LOGICIGNITER_YAAD_SCHEMA_CONTRACT.md:49` correctly says to retry without invalid memory class, but old scoreboards still preserve failed calls.

Impact:

- Historical scoreboards are not necessarily wrong as history, but if active prompts read them as current truth they become stale operational input.
- This supports an artifact cleanup need, but only after the scanner/cron priority bug is fixed.

### 7. Internal channel usage is intentional but exposed by duplicate jobs

Evidence:

- `workspace/memory/AGENT_DELEGATION_SYSTEM.md:93` documents:
  - `internal:delegation:<parent_agent_id>:<target_agent_id>:<thread_key>`
- `cron/jobs.json:154-490` uses `payload.channel: internal` for many retry jobs.

Interpretation:

- `internal` is not inherently accidental in artifacts; it is part of the delegation addressing scheme.
- The problem is not the existence of `internal`; the problem is duplicate retry scheduling and whether the runtime/channel manager handles internal cron payloads correctly.

### 8. Post-merge reconcile script is generally sound but service-map-dependent

Evidence:

- `operations/logicigniter-post-merge-reconcile.sh:63-69` uses a trusted repo allowlist.
- `operations/logicigniter-post-merge-reconcile.sh:97-109` refuses dirty repos and fast-forwards main.
- `operations/logicigniter-post-merge-reconcile.sh:137-212` maps repos to restart/health behavior.

Risk:

- It depends on `/Users/aliai/logicigniter/scripts/local-preview/app-reconcile-map.sh`.
- It uses localhost health checks for some services while public domains are the user-facing requirement.

This is not currently the primary autonomous-loop failure, but it needs a later domain-health audit.

## Hard Issues Found

1. **Scanner priority starvation**: any open PR outranks ready work and unblock candidates, causing repeated focus on the same PR lane.
2. **Duplicate internal retry storm**: 17 enabled retry jobs target the same private-module CI blocker family.
3. **Verification path mismatch**: multiple verifier scripts default to `/Users/aliai/zehn/.picoclaw` while the actual home is `/Users/aliai/.picoclaw-zehn`.
4. **COO memory bloat**: the role responsible for execution-control has a 2379-line active memory file.
5. **Stale scoreboards remain in active memory paths**: old Docker/path/Yaad failures are preserved where active prompts may read them.

## What This Means

The main failure is not proven to be a PicoClaw Go/runtime bug.

The evidence points to a bad operating-control layer:

- heartbeat delegates to CEO;
- CEO delegates to COO;
- COO obeys scanner;
- scanner repeatedly selects one open PR lane;
- cron duplicates retry the same blocker;
- active memories preserve old failures and make agents treat old blockers as still central.

## Fix Direction, Based On Evidence

Do not start by editing Go.

Fix order should be:

1. Repair scanner prioritization so it chooses company-wide work by initiative, age, owner, blocker status, and starvation prevention, not "first open PR globally".
2. Collapse duplicate internal retry cron jobs into one canonical blocker checkpoint or move them into the GitHub issue itself.
3. Update verification scripts to target `PICOCLAW_HOME=/Users/aliai/.picoclaw-zehn` by default or require it explicitly.
4. Reduce COO active memory to current doctrine plus pointers; archive historical terminal records elsewhere.
5. Archive stale scoreboards/reports out of active memory lookup paths or mark them explicitly historical.

Only after these are corrected should runtime behavior be re-tested.

## Remediation Applied

Source-side remediation:

- `operations/logicigniter-work-queue-scan.py` now treats only workflow-labeled PRs as globally actionable. Generic unlabeled open PRs remain visible in the snapshot but no longer starve ready issues and unblock candidates.
- `operations/verify-logicigniter-work-queue-scan.sh` verifies that unlabeled PRs alone do not produce a `REVIEW_PR` next action.
- Runtime verification scripts now default to `/Users/aliai/.picoclaw-zehn` and fall back there if an inherited `PICOCLAW_HOME` points at a non-runtime path.
- `operations/apply-zehn-runtime-artifact-fixes.py` applies backed-up runtime artifact cleanup under the active Zehn home without editing Go files or restarting Zehn.

Runtime remediation applied to `/Users/aliai/.picoclaw-zehn`:

- Removed duplicate enabled internal retry cron jobs for the stale `svc-webhookrouter-grpc` blocker family.
- Trimmed oversized role `memory/MEMORY.md` files into concise active doctrine files and preserved previous content under timestamped runtime backups.
- Archived old active scoreboard files older than `20260526.md`.

Backups created:

- `/Users/aliai/.picoclaw-zehn/recovery-backups/20260526T225130Z-runtime-artifact-fixes`
- `/Users/aliai/.picoclaw-zehn/recovery-backups/20260526T225323Z-runtime-artifact-fixes`

Verification passed after remediation:

- `python3 -m py_compile operations/logicigniter-work-queue-scan.py operations/apply-zehn-runtime-artifact-fixes.py`
- `bash -n operations/verify-logicigniter-work-queue-scan.sh operations/verify-logicigniter-cron-routing.sh operations/verify-zehn-prompt-memory-remediation.sh operations/verify-zehn-role-personas.sh`
- `operations/verify-logicigniter-work-queue-scan.sh`
- `operations/verify-logicigniter-cron-routing.sh`
- `operations/verify-zehn-prompt-memory-remediation.sh`
- `operations/verify-zehn-role-personas.sh`
