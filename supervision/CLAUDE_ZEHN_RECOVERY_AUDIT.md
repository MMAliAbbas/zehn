# Claude Zehn Recovery Audit

Date: 2026-05-13
Author: Claude (read-only audit pass, no runtime changes applied)
Working directory: `/Users/aliai/zehn`
Status: AUDIT ONLY — implementation pending Ali approval

This audit is grounded in file paths, log excerpts, command output, and source
references collected today. It builds on, and re-verifies, four prior Zehn
supervision documents:

- `supervision/ZEHN_PROMPT_MEMORY_RUNBOOK_AUDIT_20260512.md` (28 findings)
- `supervision/ZEHN_COMPANY_OPERATING_MODEL_ROLE_AUDIT_20260512.md` + remediation TODO
- `supervision/ZEHN_EXEC_SAFETY_GUARD_INVESTIGATION_20260512.md`
- `supervision/ZEHN_STOPPED_RUNTIME_INVESTIGATION_20260513.md`
- `supervision/ZEHN_LOGICIGNITER_AUTONOMY_FIX_PLAN_20260513.md`

Prior work substantially remediated agent persona files. The defects that
remain are operational and structural, not persona text.

---

## 1. Executive Summary

**Zehn is currently stopped, every cron job is disabled, and the runtime no
longer panics — but the design that was running before the stop cannot operate
reliably as a 48-hour autonomous company assistant.** Evidence shows the
gateway successfully executes scheduled work, but produces `HEARTBEAT_OK`
responses while open PRs and blockers remain unresolved. The dominant root
cause is a broken autonomous delivery loop: the contract requires a standard
verification wrapper that does not exist, several LogicIgniter subrepos have
no GitHub Actions to provide check signals, multi-repo branch posture is
ambiguous, and the cron schedule fans out before the underlying loop can
terminate.

The dominant problems are not Go bugs and not persona text. They are: the
missing `verify-pr.sh` standard, unspecified work-selection mechanics, cron
scheduling density without a defined replacement loop, and stale documents
under workspace memory that fragment search results and inflate on-demand
tool budget when agents follow `AGENT.md` doctrine pointers.

Recommended sequence (revised per review): fix the delivery loop first, then
reduce cron density only after the replacement work-selection mechanics are
defined, then run one controlled drill before restoring autonomy. Exec safety
remains `enable_deny_patterns: false` per Ali's explicit choice to observe
full potential; risk is logged, no reversal recommended. Open delegation
(`allow_agents: ["*"]`) stays per Ali's stated preference; the fix is in
routing prompts, failure reasons, and delegation visibility, not schema.

---

## 2. Verified Current Stopped State

Evidence collected 2026-05-13:

| Signal | Result | Source |
| --- | --- | --- |
| zehn / picoclaw / launcher / gateway processes | None running | `ps -Ao pid,etime,command \| grep -Ei 'zehn\|picoclaw\|launcher\|gateway'` (only postgres autovacuum returned) |
| Gateway listen port 18790 | No listener | `lsof -nP -iTCP:18790 -sTCP:LISTEN` (empty) |
| Web listen port 18800 | No listener | `lsof -nP -iTCP:18800 -sTCP:LISTEN` (empty) |
| User crontab | Empty | `crontab -l` returned no output |
| PicoClaw cron jobs | 17 jobs present, **all `enabled: false`** | `.picoclaw/workspace/cron/jobs.json`; grep for `"enabled": true` → 0 matches |
| `.picoclaw/.picoclaw.pid` | Stale (PID 85810, file dated 2026-05-13 01:53) | Pid file present, no process |
| LaunchAgent plist | Still installed | `/Users/aliai/Library/LaunchAgents/io.picoclaw.launcher.plist` (446 bytes, 2026-05-03) |
| LogicIgniter local services | `svc-identity` on :8090, `svc-bff` on :8091 still listening | `lsof` (pre-existing, not Zehn-managed) |

The launcher plist remains installed but the process is not running. A
`launchctl bootout` was used in the prior 2026-05-13 investigation. The plist
file should be reviewed before relaunching to confirm Ali's intended
auto-start behavior.

---

## 3. Evidence Table

Compact evidence for every finding cited below.

| ID | Claim | Evidence path / command | Notes |
| --- | --- | --- | --- |
| E-01 | All 17 cron jobs disabled | `.picoclaw/workspace/cron/jobs.json` — every `"enabled": false` | Confirmed by `grep -c '"enabled": true'` = 0 |
| E-02 | Cron uses Discord exclusively as channel | jobs.json `"channel": "discord"` x17 | No alternative reporting surface for autonomy |
| E-03 | Cron payload sizes are large | jobs.json payloads 600–1300 chars each (e.g. ceo-operating-check ~1620 chars) | Context bloat per scheduled turn |
| E-04 | Engineering check exceeds 12-min budget | launcher.log: `'logicigniter-engineering-check' completed in 738222ms` (12m18s) at 2026-05-13 09:42:18 | `tools.cron.exec_timeout_minutes = 12` |
| E-05 | zehn-operations-monitor exceeds budget | launcher.log: `completed in 977030ms` (16m17s) at 07:38:14, `787951ms` (13m8s) at 10:58:20 | Repeats |
| E-06 | Most cron runs end with `HEARTBEAT_OK` | gateway.log: `agent.go:555 message:"Response: HEARTBEAT_OK"` for li-cto turn-328 at 2026-05-13 11:41:54 | No forward movement |
| E-07 | Internal delegation outbound channel unrecognized | gateway.log: `pkg/channels/manager.go:1128 message:"Unknown channel for outbound message"` 2026-05-13T11:41:34 | Recurring warning |
| E-08 | Standard verify-pr.sh missing | `find /Users/aliai/logicigniter -maxdepth 4 -name 'verify-pr.sh'` → no results; `ls /Users/aliai/logicigniter/scripts/verification/` shows 25 scripts, no `verify-pr.sh` | Referenced by CEO/engineering prompts, jobs.json text, PR template |
| E-09 | LogicIgniter is a multi-repo root | 60+ subdirs under `/Users/aliai/logicigniter`, each with its own `.git` (spot-checked svc-mainttriage-grpc, scripts, operations, business, integration_tests) | Many on `chore/*` branches, none currently dirty |
| E-10 | Several LI repos are on feature branches, not main | `business@chore/51-current-root-mcp-final-readiness-reconciliation`, `operations@chore/3-final-readiness-ledger-validation`, `scripts@chore/3-standard-pr-verification` | Agents must reconcile vs. main |
| E-11 | Yaad MCP wired and authenticated | `.picoclaw/config.json` `tools.mcp.servers.yaad` block; `YAAD_AGENT_TOKEN` set in environment; `.picoclaw/secrets/yaad-zehn-mbp-i7.env` present | Functional |
| E-12 | Workspace memory directory contains large historical files | `wc -l .picoclaw/workspace/memory/*.md`: `ZEHN_SETUP_PLANNING.md` 1219 lines, `ZEHN_READINESS_AUDIT.md` 772 lines, `AGENT_MEETING_SYSTEM.md` 495 lines | **Correction:** these are NOT auto-loaded into the system prompt. `pkg/agent/memory.go:32-44` shows only `memory/MEMORY.md` plus `GetRecentDailyNotes(3)` from `memory/YYYYMM/YYYYMMDD.md` are auto-loaded. `pkg/agent/context.go:367` tracks only `memory/MEMORY.md` for prompt cache. Cost manifests as search-hit noise + on-demand `read_file` tool budget when an agent follows `AGENT.md` doctrine pointers like "follow X.md". |
| E-13 | Delegation log bloat | `du -sh .picoclaw/workspace/delegations` = 9.1M (1186 files) | Workspace search hits historical examples |
| E-14 | Gateway log unbounded | `du -sh .picoclaw/logs/gateway.log` = 133M (249340 lines) | Auto-rotation not configured |
| E-15 | Heartbeat fires but produces no useful work | `.picoclaw/workspace/heartbeat.log` — pattern `Heartbeat OK - silent` every 30 min from 2026-05-10 through 2026-05-13 11:23 | By design (HEARTBEAT.md forbids exec) |
| E-16 | Exec safety guard now disabled at config level | `.picoclaw/config.json`: `tools.exec.enable_deny_patterns = false` | Contradicts prior audit which recorded `true`; bypasses pkg/tools/shell.go defaults |
| E-17 | All `agents.list[*].subagents.allow_agents = ["*"]` | config.json | Any agent can delegate to any other — no hierarchical containment |
| E-18 | `agents.defaults.context_window = 32768`, `max_tokens = 4096` | config.json | Tight budget when role files + memory + cron payload all loaded |
| E-19 | OpenAI model name | config.json: `model_list[0].model = "gpt-5.5"` | Needs explicit confirmation that this string exists and is intended |
| E-20 | All config keys are valid PicoClaw struct fields | Source survey of `pkg/config/config.go`, `pkg/config/migration.go`, `pkg/config/diagnostics.go` | `decodeJSONWithDiagnostics` rejects unknown fields; load would fail otherwise |
| E-21 | Built-in deny patterns are at `pkg/tools/shell.go:50–98` | source | 36 patterns, only active when `enable_deny_patterns=true` |
| E-22 | Structured agent loader tracks AGENT.md / SOUL.md / USER.md / memory/MEMORY.md | `pkg/agent/definition.go`, `pkg/agent/context.go` | Changes to those files invalidate prompt cache |
| E-23 | `gateway_panic.log` has no actual panic | `grep -c -E 'panic\|fatal\|FATAL' gateway_panic.log` = 0; only `Error: bind: address already in use` / `bind: operation not permitted` lines from prior launcher races | Misnamed file; functions as stderr stream |
| E-24 | LogicIgniter has no GitHub Actions in target repos | LI survey: no `.github/workflows/*.yml` in `scripts`, `integration_tests` (matches `ZEHN_LOGICIGNITER_AUTONOMY_FIX_PLAN_20260513.md`) | Empty `statusCheckRollup` is expected, not a failure |
| E-25 | LogicIgniter has CodeGraph at root | `/Users/aliai/logicigniter/.codegraph/codegraph.db` 273 MB; per-subrepo indices 144 KB–308 KB | Tools `codegraph_*` available |
| E-26 | `/Users/aliai/zehn/AGENTS.md` does not exist | `ls /Users/aliai/zehn/AGENTS.md` → No such file | Top-level Zehn repo has no AGENTS.md (audit prompt referenced it) |
| E-27 | LogicIgniter top-level guidance is at `/Users/aliai/logicigniter/.agents/AGENTS.md` | 16 KB; 33 domain subdirs | Multi-domain agent reference book, not executable |

---

## 4. Broken Assumptions

Assumptions that the prior setup made which the current evidence does not
support.

1. **"Cron payloads can be the long-form agent contract."** Each cron message
   carries 600–1300 chars of policy text duplicating workspace memory and
   operating-prompt files. The agent reads the same constraints from
   `AGENT.md`, `memory/MEMORY.md`, the operating-prompt file, and the cron
   message. This wastes context, slows turns, and increases drift.
2. **"17 scheduled jobs make the company more autonomous."** Evidence (E-04,
   E-05, E-06) shows the dense schedule produces busy logs but
   `HEARTBEAT_OK`-dominated outcomes. Throughput is throttled by the
   serialized cron queue and 12-minute per-job budget.
3. **"Heartbeat is the supervisor signal."** Per `HEARTBEAT.md`, heartbeat
   forbids shell exec. So it cannot inspect logs, GitHub, or repos. It is a
   liveness probe, not a monitor. Yet `ZEHN_OPERATING_CADENCE.md` treats it
   as the supervisor.
4. **"Specialists can complete work because they have scheduled queues."** A
   specialist queue check produces `HEARTBEAT_OK` because (a) it cannot find
   matching `zehn:ready` issues, or (b) the open PRs have empty
   `statusCheckRollup` (no GitHub Actions configured in those LI subrepos),
   so the queue check has no terminal verdict to give.
5. **"`verify-pr.sh` is the standard verification."** It does not exist in
   any LogicIgniter subrepo at any depth (E-08). Multiple prompts treat it as
   a hard gate.
6. **"Docker-based local stack is the readiness contract."** The 2026-05-13
   runtime investigation already confirmed Docker is not present on this
   host. `scripts/local-preview/README.md` and `start-real-stack.sh` still
   point agents at Docker.
7. **"Workspace memory is a single source of truth."** The workspace memory
   directory mixes historical readiness audits, setup planning, and current
   operating contracts. **Correction (post-review):** these extra files do
   NOT auto-load into the system prompt — `pkg/agent/memory.go` and
   `pkg/agent/context.go` only auto-load `memory/MEMORY.md` plus the last 3
   days of `memory/YYYYMM/YYYYMMDD.md`. The real cost is (a) workspace
   search returns stale content, and (b) when `AGENT.md` says "follow X.md",
   the agent spends tool budget on `read_file` to load it. Not active-prompt
   bloat, but fragmentation and on-demand cost.
8. **"Subagent containment is enforced by config."** Every entry in
   `agents.list[*]` has `subagents.allow_agents: ["*"]`. There is no
   hierarchical containment in schema. Per Ali's stated preference, open
   delegation is intentional; relevance and responsibility are enforced
   through routing prompts, not schema.
9. **"The safety guard is the protection layer."** `enable_deny_patterns` is
   `false` (E-16) — built-in deny patterns (`pkg/tools/shell.go:50–98`) are
   not enforced. **Per Ali's explicit direction, this is intentional** to
   observe full system potential. The user-facing prompt rules ("never push
   to main, never touch secrets") and the path allowlist in `tools` are the
   active guardrails. Prior exec-safety investigation assumed `true`; the
   current state supersedes that.
10. **"`zehn-main` and `personal` and `li-ceo` are distinct."** The role
    text is now correctly distinct (verified in workspace/AGENT.md,
    workspace-personal/AGENT.md, workspace-li-ceo/AGENT.md). The remaining
    risk is operational, not text-level.

---

## 5. Stale / Contradictory Instruction Inventory

Files containing stale or contradictory instructions. **Correction
(post-review):** these files do NOT auto-load into the system prompt (per
`pkg/agent/memory.go` and `pkg/agent/context.go`, only `memory/MEMORY.md` and
recent daily notes auto-load). Risk is: (a) workspace search surfaces them
as authoritative-looking matches, and (b) `AGENT.md` files include "follow
X.md" doctrine pointers that cause agents to read these via `read_file`,
consuming tool budget. Source: prior `ZEHN_PROMPT_MEMORY_RUNBOOK_AUDIT_20260512.md`
findings, re-validated by spot-checking today.

- `.picoclaw/workspace/memory/ZEHN_SETUP_PLANNING.md` (1219 lines) — mixes
  current setup, old MCQ planning, old "open draft PRs" guidance, old
  87-agent model, and old Discord "mention-only" claims. Status: present and
  active.
- `.picoclaw/workspace/memory/ZEHN_READINESS_AUDIT.md` (772 lines) — stale
  exec capability table; the live config diverges (E-16).
- `.picoclaw/workspace/memory/LOGICIGNITER_SOLUTION_PORTFOLIO_PLAN.md` —
  contains "App owners remain responsible…" language even though no
  `workspace-li-app-*` agents exist (the org tree only has 41 agents now).
- `.picoclaw/workspace/memory/LOGICIGNITER_SOFTWARE_DELIVERY_SYSTEM.md` —
  approval-gate text broader than the live `APPROVAL_ESCALATION_MATRIX.md`,
  which grants standing setup/development authority.
- Cron prompts in `.picoclaw/workspace/cron/jobs.json` — every CEO and
  engineering payload still mentions `verify-pr.sh` as the merge gate; the
  wrapper does not exist (E-08).
- `.picoclaw/workspace/operating-prompts/logicigniter-ceo-operating-check.md`
  and `logicigniter-engineering-check.md` — both still treat `verify-pr.sh`
  as the canonical verification command.
- `.picoclaw/workspace/operating-prompts/zehn-operations-monitor.md` —
  defines `zehn-main` monitor scope correctly, but its scheduled job
  consistently runs 13–16 minutes (E-05), which is the slowest scheduled
  cadence, and frequently produces no actionable findings because heartbeat
  silence is by design.
- Many historical delegation JSONs under
  `.picoclaw/workspace/delegations/*.json` (1186 files, 9.1M) include older
  prompt bodies that contradict current policy (draft PRs, app-owner
  delegation targets, no-GitHub-artifacts during setup).

---

## 6. Config Validity Review

`/Users/aliai/zehn/.picoclaw/config.json` was validated against the PicoClaw
Go source (`pkg/config/config.go`, `pkg/config/migration.go`,
`pkg/config/diagnostics.go`).

**Result: every key resolves to a defined struct field.** PicoClaw's loader
calls `decodeJSONWithDiagnostics` → `collectUnknownJSONFields`, which
**rejects unknown JSON keys**. So if any key were invented, the config would
fail to load at launch. The fact that the gateway started repeatedly through
2026-05-13 01:53 confirms current keys are accepted.

Concrete settings worth flagging operationally (not invalid, but worth
intentional confirmation):

| Key | Value | Concern |
| --- | --- | --- |
| `tools.exec.enable_deny_patterns` | `false` | Built-in destructive-command/process-control patterns at `pkg/tools/shell.go:50–98` are NOT enforced. Prompts are now the only guardrail. The prior exec-safety investigation incorrectly assumed `true`. |
| `tools.exec.allow_remote` | `true` | Required for LogicIgniter work; intended. |
| `tools.exec.timeout_seconds` | `720` | 12 min per exec — matches cron job budget. |
| `tools.cron.exec_timeout_minutes` | `12` | Multiple jobs exceed this (E-04, E-05). |
| `tools.cron.allow_command` | `false` | Cron payloads must use `agent_turn`, not raw command — correct. |
| `agents.defaults.max_tool_iterations` | `50` | CEO/CTO checks have hit max-iter previously (prior audit F-004). |
| `agents.defaults.context_window` | `32768` | Per-turn prompt actually carries: AGENT.md + SOUL.md + USER.md + `memory/MEMORY.md` + recent daily notes (last 3 days) + peer-agent catalog + tool/skills descriptors + cron payload + any docs the agent explicitly `read_file`s. Other `memory/*.md` files are NOT auto-loaded. Tight budget mainly when cron payload is verbose, peer-agent catalog is large, and the agent's first action is to read several pointer docs. |
| `agents.defaults.max_tokens` | `4096` | Output budget per turn. |
| `agents.defaults.async_delegation.max_concurrent` | `9` | Generous; combined with `allow_agents: ["*"]` enables wide fanout. |
| `agents.defaults.restrict_to_workspace` | `false` | Required; LI repo work is outside workspace. |
| `agents.defaults.allow_read_outside_workspace` | `true` | Intended. |
| `tools.allow_read_paths` | `^/Users/aliai/logicigniter`, `^/Users/aliai/projects`, `^/Users/aliai/zehn/\.picoclaw/workspace/memory` | Read-only path allowlist; intended. |
| `tools.allow_write_paths` | `^/Users/aliai/logicigniter`, `^/Users/aliai/projects` | Write allowlist excludes Zehn `.picoclaw` workspace — agents cannot write to their own workspace via the file tools. |
| `tools.mcp.servers.yaad` | `enabled: true, deferred: true, type: http, url, Bearer ${YAAD_AGENT_TOKEN}` | Token resolved via env (`YAAD_AGENT_TOKEN` is set). |
| `heartbeat.{enabled, interval}` | `true, 30` | Every 30 min; HEARTBEAT.md forbids shell exec. |
| `agents.list[*].subagents.allow_agents` | `["*"]` for all 41 agents | No hierarchical containment. |
| `agents.organization.roots` | `["zehn-main", "personal", "li-ceo"]` | Three-root model is intentional and consistent with current role files. |
| `channel_list.discord.allow_from` | One ID, mention_only=false | Correct; matches user preference. |
| `channel_list.pico.settings.allow_origins` | `127.0.0.1:18800` + `localhost:18800` | Local-only, fine. |
| `model_list[0].model` | `gpt-5.5` | Needs explicit confirmation. The auth method is `oauth`; no API key visible in config. |
| `build_info.version` | `nightly-65-g6e1fab80` from 2026-05-02 | The running binary may not match — gateway logs show `v0.2.8-78-g08a4abac` from 2026-05-12 (so a newer rebuild has occurred since this config was last serialized). |

`.picoclaw/.security.yml` contains the Discord and pico tokens. **Secrets are
in the correct location**, not in `config.json`. No leaked secrets observed
in committed supervision docs (spot-checked).

---

## 7. Agent Role / Persona Quality Review

Verified files (today):

- `.picoclaw/workspace/AGENT.md` — defines `zehn-main` as system monitor,
  not the LogicIgniter CEO. Routing boundary is explicit. ✓
- `.picoclaw/workspace/memory/MEMORY.md` — concise (58 lines), aligned. ✓
- `.picoclaw/workspace-personal/AGENT.md` — personal-only scope, routes LI
  to `li-ceo`. ✓
- `.picoclaw/workspace-li-ceo/AGENT.md` — CEO mandate, terminal-outcome
  rule, execution authority, repo-access doctrine. ✓
- `.picoclaw/workspace-li-ceo/memory/MEMORY.md` — concise; canonical 10
  suites listed. ✓
- `.picoclaw/workspace-li-coo/AGENT.md` — throughput owner, WIP/blocker
  doctrine. ✓
- `.picoclaw/workspace-li-backend-developer/AGENT.md` — scoped to
  area:backend, repo-clean rule, engineering quality doctrine. ✓

**The persona-text quality problem identified by the 2026-05-12 audit has
largely been remediated.** The roles read like operators, not generic titles
with disclaimers.

Remaining gaps:

- The active **company operating prompts** (operating-prompts/*.md) and the
  cron payloads still embed the missing `verify-pr.sh` as a hard step.
- Workspace shared memory (`workspace/memory/*.md`) includes large historical
  files. **Correction:** these do not override role text by auto-loading
  (only `memory/MEMORY.md` + recent daily notes auto-load per
  `pkg/agent/memory.go`). They can still mislead agents by (a) polluting
  workspace search results and (b) being explicitly loaded via `read_file`
  when `AGENT.md` or another doctrine file references them.

---

## 8. Cron / Heartbeat / Autonomy Review

`.picoclaw/workspace/cron/jobs.json` lists 17 jobs, all disabled:

| Job | Schedule | Last status |
| --- | --- | --- |
| logicigniter-ceo-operating-check | `5 * * * *` | ok |
| personal-operating-check | `9 * * * *` | ok |
| logicigniter-coo-control-board-check | `1 * * * *` | ok |
| logicigniter-engineering-check | `*/30 * * * *` | ok |
| logicigniter-github-control-plane-reconciler | `3 * * * *` | ok |
| Daily LogicIgniter agent training | `30 8 * * *` | ok |
| logicigniter-architect-work-queue | `12 * * * *` | ok |
| logicigniter-backend-work-queue | `17 * * * *` | ok |
| logicigniter-frontend-work-queue | `22 * * * *` | ok |
| logicigniter-ux-work-queue | `27 * * * *` | ok |
| logicigniter-integration-work-queue | `32 * * * *` | ok |
| logicigniter-data-ai-work-queue | `37 * * * *` | ok |
| logicigniter-devops-work-queue | `42 * * * *` | ok |
| logicigniter-qa-work-queue | `47 * * * *` | ok |
| logicigniter-security-work-queue | `52 * * * *` | ok |
| logicigniter-docs-work-queue | `57 * * * *` | ok |
| zehn-operations-monitor | `15,45 * * * *` | ok |

Observations:

- **Density**: at any given hour, 17 jobs fire across CTO, COO, CEO,
  operations, ops-monitor, and 10 specialist queues. With per-job runs in
  the 1–16 minute range, jobs serialize and starve each other.
- **Budget overruns**: engineering-check and zehn-operations-monitor have
  exceeded 12 min (`tools.cron.exec_timeout_minutes`) at least 3 times in
  the last 8 hours of evidence (E-04, E-05). Once timed out, the cron logs
  `ok` because the command returned — the business outcome is opaque.
- **Output**: routine cron turns end with `HEARTBEAT_OK` (E-06). The cron
  framework treats that as success even though no work moved.
- **Channel only Discord**: every job posts to a Discord channel. If
  Discord delivery fails (E-07-style outbound channel warnings), there is
  no fallback artifact (e.g. local report file).
- **Heartbeat**: 30-min interval, `Heartbeat OK - silent` for the full 3
  days inspected (E-15). HEARTBEAT.md forbids `exec` so it cannot detect
  many of the failure modes the operations monitor is asked to catch.

---

## 9. GitHub / Project / Issue Execution Flow Review

The intended loop:
`issue → claim → branch → implementation → verification → PR → review → merge → post-merge reconcile`.

What blocks this loop today:

1. **`verify-pr.sh` does not exist** anywhere in `/Users/aliai/logicigniter`
   (E-08). CEO operating prompt, engineering prompt, and 5+ cron payloads
   reference it.
2. **No GitHub Actions workflows** in `scripts`, `integration_tests`, and
   most subrepos (E-24). `gh pr checks` returns "no checks reported", which
   the prior 2026-05-13 investigation flagged as a stuck signal. Agents
   currently treat empty `statusCheckRollup` as ambiguous.
3. **Multi-repo branch posture**: `business`, `operations`, `scripts` are
   currently on `chore/*` feature branches rather than main (E-10). Agents
   that assume "checkout main" can fail.
4. **Codex review confusion**: prompts already note `👀` ≠ approval, but
   delegation records still cite "Codex commented" as a merge signal.
5. **PR template at `.github/.github/pull_request_template.md`** also
   references `verify-pr.sh` — a separate repo from where the script would
   live.

---

## 10. LogicIgniter Workspace / Repo Access Review

Live multi-repo root: `/Users/aliai/logicigniter`. Confirmed:

- 60+ subrepos (50+ `svc-*` services, plus `business`, `operations`,
  `supervision`, `scripts`, `integration_tests`, `proto`, `go-packages`,
  `infra`, `infra_research`, `config`, `keycloak`, `logicigniter-runner`,
  `.agents`, `.github`).
- Each subdir is an independent git repo. Spot-checked working trees are
  clean.
- `.agents/AGENTS.md` (16 KB) + 33 domain subdirs = reference book for AI
  agents working in LogicIgniter.
- `.codegraph/codegraph.db` 273 MB at root; per-subrepo indices available.
- Control-plane scripts: `scripts/local-preview/start-mcp-runtime-proof.sh`
  (host-native, per autonomy fix plan), `operations/final-readiness-verify.sh`,
  `operations/engineering-pipeline/{pr.sh, worker.sh}`,
  `operations/codex-coder/run.sh`.
- Missing: `verify-pr.sh`; missing dedicated post-merge reconciler in
  LogicIgniter (`operations/logicigniter-post-merge-reconcile.sh` exists
  here in the Zehn `operations/` dir — untracked).

Effect: agents that need "where is the code, where is the script, what
branch is it on" can answer with `.agents/AGENTS.md` + CodeGraph, but the
absence of `verify-pr.sh` and a standard restart wrapper forces ad hoc
fallback, which the exec safety guard can block when patterns trigger.

---

## 11. Yaad Memory Review

Configured:
- `tools.mcp.servers.yaad` block: `enabled: true`, `deferred: true`, `type:
  http`, `url: https://yaad.mmaliabbas.com/mcp`, `Authorization: Bearer
  ${YAAD_AGENT_TOKEN}`.
- `YAAD_AGENT_TOKEN` and `YAAD_API_TOKEN` are present in the shell
  environment.
- `.picoclaw/secrets/yaad-zehn-mbp-i7.env` (274 bytes) provides token at
  launcher start.

Doctrine files (verified):
- `.picoclaw/workspace/memory/LOGICIGNITER_YAAD_SCHEMA_CONTRACT.md` —
  defines `scope_type=organization`, `external_key=logicigniter` for LI-wide
  facts, `user_persona:ali` for personal preferences.
- `MEMORY.md` files for `zehn-main`, `li-ceo`, etc. all reference Yaad
  policy.

Operational concerns:
- **No direct evidence** in logs that Yaad reads/writes succeed at scale.
  The last 8 hours of gateway events show no `yaad` MCP tool failures, but
  also no clear "wrote to Yaad" traces. The audit cannot confirm Yaad is
  actually being used as durable memory without a controlled test.
- **No mechanism enforces Yaad-first** behavior. If agents rediscover LI
  structure from filesystem on every turn, no system signal will flag that
  as a Yaad bypass.
- Yaad MCP tool is `deferred` in config — the tool catalog loads lazily.
  This is fine but means it's invisible until first requested.

**Required Yaad read/write verification contract** (added per review,
implementation detail in §13 Step 7a):
- **Read-first contract**: CEO, COO, and specialists must read canonical
  `organization:logicigniter` Yaad memory before scanning the filesystem
  for company structure, app/service inventory, current portfolio state,
  or blocker history. Filesystem scan is the fallback when Yaad has no
  matching entry, not the default action.
- **Write-back contract**: Every terminal outcome from CEO, COO, and
  specialists must produce one durable Yaad write under
  `organization:logicigniter` (or `user_persona:ali` for personal scope)
  capturing the decision, evidence pointer, owner, and date. "Terminal"
  here means merged, approved, blocked-with-owner, escalated, deferred,
  or replaced/closed (per the COMPANY_OPERATING_CONTRACT).
- **Evidence requirement**: a Yaad write succeeds → log line includes
  the resulting Yaad entry ID; a Yaad write fails → agent records the
  failure under `Failure Reason:` and falls back to writing the same
  content to local `memory/MEMORY.md` so the next run can retry.

---

## 12. Risk Ranking

### P0 — must address before restart

- **P0-A: Missing `verify-pr.sh` blocks the autonomous delivery loop.**
  Every CEO/engineering scheduled run cites it as the merge gate. A
  standard wrapper is the contract that makes "verified" a terminal state.
  (E-08)
- **P0-B: Work-selection mechanics for the delivery loop are unspecified.**
  Without explicit selection, claim, dedup, retry, and escalation rules,
  the loop cannot terminate even when verification works. The current
  17-job cron schedule papers over the absence of a real selection
  algorithm.
- **P0-C: Cron over-scheduling without a defined replacement loop.** 17
  jobs, many over the 12-min budget, mostly returning `HEARTBEAT_OK`.
  Reducing density without specifying the replacement work-selection
  algorithm just relocates the loop. (E-03, E-04, E-05, E-06)
- **P0-D: Cron payload text duplicates doctrine and embeds stale
  `verify-pr.sh` references.** Long payloads (600–1300 chars) duplicate
  what AGENT.md and operating-prompts already say. (E-03)

### P1 — must address before claiming 48-hour autonomy

- **P1-A: Stale documents fragment workspace search.** Large historical
  files in `.picoclaw/workspace/memory/` (`ZEHN_SETUP_PLANNING.md` 1219
  lines, `ZEHN_READINESS_AUDIT.md` 772 lines) surface in agent searches
  and trigger on-demand `read_file` budget when `AGENT.md` doctrine
  pointers cite them. **Not** active-prompt bloat (per source review), but
  still a discoverability and on-demand-budget problem. (E-12)
- **P1-B: `zehn-operations-monitor` runs 13–16 min and produces little.**
  Either bound it tighter, give it a smaller scope, or move it to
  event-driven monitoring. (E-05)
- **P1-C: `Unknown channel for outbound message` warning** — internal
  delegation responses have no home. Route to delegation inbox or silence
  at source. (E-07)
- **P1-D: No GitHub Actions in `scripts` / `integration_tests`** — empty
  `statusCheckRollup` looks broken to the verification path. Add a minimal
  workflow or document the expectation explicitly. (E-24)
- **P1-E: Multi-repo branch posture ambiguity.** `business`, `operations`,
  `scripts` on `chore/*` feature branches; agents must reconcile against
  main before assuming a clean state. (E-10)
- **P1-F: `model_list[0].model = "gpt-5.5"`** — confirm the string is
  correct for the provider currently in use. (E-19)

### Recorded user decisions (not reversal candidates)

- **`tools.exec.enable_deny_patterns: false`** — explicit user choice to
  observe full system potential. Risk: built-in destructive-command and
  process-control patterns (`pkg/tools/shell.go:50–98`) are not enforced;
  the only active guardrails are the path allowlists in `tools.allow_*_paths`
  and prompt-level rules. Logged as a known risk, no reversal recommended.
  (E-16)
- **`agents.list[*].subagents.allow_agents: ["*"]`** — explicit user
  preference for open delegation. Relevance and responsibility should be
  enforced via routing prompts, failure-reason discipline, and delegation
  visibility, not schema containment. (E-17)

### P2 — should address before scaling

- **P2-A: `gateway.log` unbounded (133 MB), `delegations/` 9.1 MB, 1186
  files.** Set log rotation and add periodic delegation archiving.
- **P2-B: `gateway_panic.log` is misnamed.** It's actually stderr.
- **P2-C: Multi-repo branch posture.** `business`, `operations`, `scripts`
  on `chore/*` branches — confirm whether Zehn should reconcile to `main`.
- **P2-D: Stale LaunchAgent plist** at
  `~/Library/LaunchAgents/io.picoclaw.launcher.plist`. Confirm intended
  auto-start behavior before next reboot.
- **P2-E: Historical drafts under workspace-li-operations/** (e.g.
  `issue-7-body.md`) still present — flagged by 2026-05-12 audit F-027.
- **P2-F: No top-level `AGENTS.md` at `/Users/aliai/zehn`** — if the audit
  prompt references one, it does not exist in this repo. Decide whether
  the canonical one lives in `.picoclaw/workspace/AGENT.md` (current) or a
  new top-level file.

---

## 13. Minimal Recovery Plan (for review before any implementation)

Goal: restore a reliable autonomous delivery loop, then re-expand autonomy
in stages. Ordering revised per review: **fix the loop before reducing
cron**; do not reverse exec-safety or subagent decisions Ali has made.

### Step 1 — Build the `verify-pr.sh` standard wrapper, v1 (P0)

This is the single highest-leverage move. The wrapper is the contract that
makes "verified" a terminal state. Without it, agents either skip
verification or stall.

**v1 scope is deliberately narrow** to ship a working contract fast, then
expand per repo type after the first drill (revised per review):

Create `/Users/aliai/logicigniter/scripts/verification/verify-pr.sh`. v1 must:

1. Accept `--repo <repo> --issue <issue>` and optionally `--pr <pr>`.
2. **Detect the repo** by inspecting the working directory's git config /
   remote and matching against a known list.
3. **Detect changed files** vs. the repo's main branch
   (`git diff --name-only main...HEAD`).
4. **Run repo-local verification if present**: if the repo has its own
   `scripts/verify.sh`, `Makefile` target `verify`, or
   `scripts/verification/verify-service.sh`, run that and capture exit
   status.
5. **Fallback per repo class**:
   - **Go repos** (`svc-*`, `go-packages`, `integration_tests` excluded
     until v2): `go test ./...` on the affected packages.
   - **Shell/script repos** (`scripts`, `operations`): `bash -n` on every
     changed `*.sh` file.
   - **Docs/config repos** (`supervision`, `proto`, `config`, `business`):
     run **only** tools that already exist (`yamllint`, `protoc --lint_out`,
     etc.); if none exist, exit 0 with `skipped: no tools` in the evidence.
6. **Write structured JSON evidence** to
   `operations/.verify-pr/<repo>-<issue>-<timestamp>.json` with: detected
   repo, detected changed files, each step's command/duration/exit code,
   final verdict (`pass` / `fail` / `skipped`), and a one-line `reason`.
7. Exit 0 on full pass or skipped-with-reason; exit non-zero on any
   failure, with the failing stage named in stderr.
8. Be idempotent and safe to re-run.

**Out of v1 scope (deferred to v2 after first drill):**
- `integration_tests` repo (needs host-native env wiring per
  `ZEHN_STOPPED_RUNTIME_INVESTIGATION_20260513.md`).
- MCP-runtime-proof orchestration.
- Repo-specific deep verification (`audit-*.mjs`, etc.) — these stay
  callable separately but are not part of the v1 contract.

Acceptance for v1: `verify-pr.sh` returns deterministic `pass` / `fail` /
`skipped` for at least one issue in a Go svc-* repo and one issue in a
shell-script repo, with a valid JSON evidence file produced for each.

### Step 2 — Specify work-selection mechanics (P0)

Before any cron consolidation, define the replacement loop:

1. **Selection**: where the queue lives (GitHub Issues with `zehn:ready`
   + matching `area:*`), and the priority order (oldest unclaimed first,
   skip `approval:ali-required` unless explicit body approval present).
2. **Claim**: how a specialist marks an issue as theirs (`zehn:in-progress`
   label + issue comment with agent ID + timestamp + intended branch
   name). Stale claims (>4 hours with no branch push) auto-release.
3. **Dedup**: one open claim per issue; cron passes that see an existing
   in-progress claim and a corresponding open PR skip without warning.
4. **Branch + verify + PR**: from claim, the same agent owns through
   branch push and PR open, using `verify-pr.sh` (Step 1).
5. **Review**: PR review queue is a first-class queue alongside issue
   queue; reviewers route by repo/area regardless of who implemented.
6. **Merge**: gated on `verify-pr.sh` pass + internal review + Codex
   thumbs-up (not 👀) + label-based risk policy.
7. **Post-merge reconcile**: existing
   `operations/logicigniter-post-merge-reconcile.sh` runs, with a
   reporting line per affected service.
8. **Retry + escalate**: a claim that fails verification twice is
   commented with the failure summary, label changes to
   `zehn:blocked`, and a precise next-action is required; three failures
   escalates to li-coo with a meeting request.

Record this in a new doc (e.g.
`.picoclaw/workspace/memory/LOGICIGNITER_AUTONOMOUS_DELIVERY_LOOP.md`)
referenced from `AGENT.md` for execution roles. Agents already have the
COMPANY_OPERATING_CONTRACT; this adds the missing concrete mechanics.

### Step 3 — Improve routing, failure visibility, and delegation status (P0)

Open delegation (`allow_agents: ["*"]`) stays. Make it work better:

1. **Routing discipline**: each role's `AGENT.md` already names a default
   parent and peer set. Reinforce in operating prompts: "delegate to the
   most relevant role; record why; if unsure, ask li-coo before fanning
   out".
2. **Failure reasons**: every delegation that returns without a terminal
   outcome must include a one-line `Failure Reason:` field (e.g.,
   "tool budget exhausted before verification", "external service
   unreachable", "issue body missing acceptance criteria"). Cron payloads
   should ask for this field explicitly.
3. **Delegation status visibility**: `delegation_status` and
   `delegation_inbox` tools are already enabled. Add a routine
   coordinator check (li-coo) that lists open delegations older than 30
   minutes with their failure reasons, posted to the COO channel.
4. **Capacity**: `async_delegation.max_concurrent: 9` is generous; leave
   it. Document explicitly that simultaneous wide fanout from `zehn-main`
   or `li-ceo` is expected behavior, not a misconfiguration.

### Step 4 — Replace cron payload duplication with references (P1)

Only after Steps 1–3 are in place:

1. Rewrite each cron job message to be ≤ 250 chars:
   - one line naming the operating prompt file;
   - the role's matching label/scope;
   - the terminal-outcome contract reference;
   - **no inline policy text** (it's in the prompt file already).
2. Keep the `verify-pr.sh` references intact (Step 1 made them work).
3. Keep the standing-authority language out of the cron message; it lives
   in `LOGICIGNITER_APPROVAL_ESCALATION_MATRIX.md` and the cron message
   can just reference that.

### Step 5 — Reduce cron density with a defined sweep mechanism (P1)

Only after Step 2's mechanics are written:

1. Replace the 10 specialist work-queue jobs with a single
   `logicigniter-specialist-sweep` (every 30 minutes) routed to li-coo.
   The sweep:
   - reads `zehn:ready` issues across all configured repos via `gh`;
   - groups them by `area:*` label;
   - for each non-empty group with no existing in-progress claim, opens a
     **delegation** to the matching specialist with the specific issue
     number(s) and the Step 2 mechanics reference;
   - also lists PR review queue items by area; delegates review work the
     same way;
   - returns a terminal report listing what it delegated, what it
     skipped (and why), and what it escalated;
   - never returns `HEARTBEAT_OK` if there is open work in any tracked
     area.
   - **Strict no-work-found contract** (per review): if zero claimable
     work is found, the sweep must report — never silently — at least:
     * **Repos searched**: explicit list of repos queried;
     * **Labels queried**: `zehn:ready`, each `area:*` queried, any
       blocker filter applied;
     * **Issues found**: total count and per-repo count, including
       in-progress and review-queue items that were intentionally
       skipped (with reason: "claimed by X", "blocked by Y",
       "approval:ali-required without explicit body approval", etc.);
     * **Why nothing was claimable**: one-line reason summarizing the
       above (e.g., "all 7 ready issues are already in-progress; 3 PRs
       awaiting Codex review").
     Only after that report may the sweep return `HEARTBEAT_OK`. This
     prevents silent fake-success.
2. Disable `Daily LogicIgniter agent training` until the delivery loop is
   proven (it can be re-enabled later as a focused study cadence).
3. Reduce `zehn-operations-monitor` from `15,45 * * * *` to `15 * * * *`
   until the per-run budget overruns are gone, and require it to write a
   structured findings file in addition to its Discord message.
4. Keep CEO (`5 * * * *`), COO (`1 * * * *`), engineering check
   (`*/30 * * * *`), GitHub control-plane (`3 * * * *`), and personal
   (`9 * * * *`) cadences. These are the routing layer.

### Step 6 — Search hygiene for workspace memory (P1)

This is search/discovery cleanup, not prompt-context optimization
(corrected per review).

1. Move historical / superseded docs into
   `.picoclaw/workspace/memory/archive/` (still on disk, still readable
   via explicit `read_file`, but out of default workspace search results):
   - `ZEHN_SETUP_PLANNING.md`
   - `ZEHN_READINESS_AUDIT.md`
   - `LOGICIGNITER_SOLUTION_PORTFOLIO_PLAN.md` (contains the stale "app
     owners remain responsible" line)
2. Remove "follow X.md" doctrine pointers in any `AGENT.md` that targets
   a now-archived doc; replace with a pointer to the current source of
   truth.
3. Audit `AGENT.md` doctrine pointers across all 41 agent workspaces and
   confirm each cited file still exists and remains canonical.
4. Add a one-page index at `.picoclaw/workspace/memory/INDEX.md` listing
   active doctrine files vs. archived files.

### Step 7 — Record the explicit user decisions

1. Add a short note to `.picoclaw/workspace/memory/ZEHN_CURRENT_STATE.md`:
   - `tools.exec.enable_deny_patterns: false` is an explicit Ali decision
     (date), pending observation of full potential. Built-in destructive
     and process-control deny patterns are inactive; path allowlist + prompt
     discipline are the active guardrails.
   - `agents.list[*].subagents.allow_agents: ["*"]` is an explicit Ali
     decision; open delegation is intended.
2. These notes are for future agents that read CURRENT_STATE so they don't
   re-flag the same items.

### Step 7a — Yaad read-first / write-back verification (P1)

Tie the §11 Yaad contract into the operating doctrine so the first drill
exercises it end-to-end:

1. Update `LOGICIGNITER_YAAD_SCHEMA_CONTRACT.md` to state the read-first
   and write-back rules explicitly (currently it only documents the
   schema, not the read/write posture).
2. Update `li-ceo` / `li-coo` / specialist `AGENT.md` files with a
   one-line "Read canonical `organization:logicigniter` Yaad memory before
   filesystem scan; write terminal decisions back to Yaad" requirement,
   pointing at the schema contract.
3. Add a Yaad self-check in the §13 Step 8 drill: before the chosen
   specialist starts work, it must produce a Yaad read of any prior
   entries for that repo/area, and after the loop closes, a Yaad write
   recording the outcome. The drill is not "passed" until both Yaad
   operations succeed and their entry IDs appear in logs.
4. Acceptance: at the end of Step 8, the gateway log contains at least
   two successful `yaad` MCP tool calls (one read, one write) and the
   Yaad entry IDs appear in the drill's evidence summary.

### Step 8 — Restart with one controlled drill

Only after Steps 1–7 and 7a:

1. Pick one low-risk `zehn:ready` issue in a repo that already has clear
   verification (e.g., `scripts` with a shell change, or one of the
   `svc-*-grpc` services with a `go test` change).
2. Enable a minimal cron set:
   - `logicigniter-coo-control-board-check` (1×/hour),
   - `logicigniter-specialist-sweep` (every 30 min),
   - `zehn-operations-monitor` (1×/hour),
   - heartbeat (unchanged).
3. Let the loop run for that one issue: claim → branch → verify → PR →
   review → merge → reconcile.
4. Inspect logs and the Step 2 evidence file. Iterate before enabling more
   cron jobs.
5. Only after one full loop is proven, re-enable CEO check, engineering
   check, GitHub control-plane reconciler, and personal check.

---

## 14. Files Proposed for Deletion / Replacement / Consolidation

**Proposed deletions** (none destructive; move to dated archive):
- `.picoclaw/config.json.pre-delegation-rollout` (2026-05-03 snapshot) —
  move to `.picoclaw/backups/`.
- `.picoclaw/config.json.pre-org-hierarchy-20260507074641` — move to
  `.picoclaw/backups/`.
- `.picoclaw/.DS_Store` — delete.

**Proposed archive moves** (active → archive subdir):
- `.picoclaw/workspace/memory/ZEHN_SETUP_PLANNING.md`
- `.picoclaw/workspace/memory/ZEHN_READINESS_AUDIT.md`
- `.picoclaw/workspace/memory/LOGICIGNITER_SOLUTION_PORTFOLIO_PLAN.md`
- Anything under `.picoclaw/workspace-*/` matching `*-issue-*-body.md`,
  `tmp-pr-*-body.md` (per prior F-027).

**Proposed consolidations**:
- Merge `LOGICIGNITER_OPERATING_CADENCE.md` and `ZEHN_OPERATING_CADENCE.md`
  into one document with a clear "Zehn system" section and a "LogicIgniter
  company" section.
- Trim `AGENT_DELEGATION_SYSTEM.md` and `AGENT_MEETING_SYSTEM.md` from
  combined 765 lines to one-pagers.
- Cron payload bodies → single shared `cron-payload-policy.md` and
  per-job 1-line references.

**Proposed retention** (preserve as-is):
- All workspace `AGENT.md`, `SOUL.md`, `USER.md`, `MEMORY.md` files (good
  quality after prior remediation).
- `LOGICIGNITER_COMPANY_OPERATING_CONTRACT.md`,
  `LOGICIGNITER_GITHUB_CONTROL_PLANE.md`,
  `LOGICIGNITER_APPROVAL_ESCALATION_MATRIX.md`,
  `LOGICIGNITER_YAAD_SCHEMA_CONTRACT.md`,
  `LOGICIGNITER_REPO_ACCESS_DOCTRINE.md`,
  `LOGICIGNITER_ENGINEERING_QUALITY_DOCTRINE.md`.

---

## 15. Files Proposed for Code / Config / Script Changes

**Config (`.picoclaw/config.json`)** — minimal, no reversal of Ali decisions:
- Confirm `model_list[0].model = "gpt-5.5"` string is correct for the
  provider in use. No change recommended without confirmation.
- **No change** to `tools.exec.enable_deny_patterns` (stays `false` per
  explicit Ali decision; risk is logged in §12 and recorded in
  CURRENT_STATE per §13 Step 7).
- **No change** to `agents.list[*].subagents.allow_agents` (stays `["*"]`
  per explicit Ali decision; relevance enforced through routing, not
  schema).

**Cron jobs (`.picoclaw/workspace/cron/jobs.json`)** — only after §13 Steps
1–3 are done:
- Rewrite payloads to ≤ 250 chars referencing operating prompts.
- Replace 10 specialist queues with one `logicigniter-specialist-sweep`
  whose mechanics are defined in §13 Step 5.
- Disable daily training job until the delivery loop is proven.

**Operating prompts**:
- After §13 Step 1, the `verify-pr.sh` references work as intended; keep
  them.
- Add a `Failure Reason:` field requirement to delegation-returning
  prompts (li-coo, li-cto, li-engineering).
- Cap CEO and engineering prompts to a bounded inspection scope (e.g.
  "at most 6 active items per run") to stop 12-min budget overruns.

**Scripts to add**:
- `/Users/aliai/logicigniter/scripts/verification/verify-pr.sh` — required
  (§13 Step 1).
- A small `operations/zehn-restart.sh` in the Zehn repo that performs
  controlled stop/start of launcher + gateway with safe defaults.
- Optional: `.picoclaw/workspace/memory/INDEX.md` listing active vs.
  archived doctrine files (§13 Step 6).

**New memory doc**:
- `.picoclaw/workspace/memory/LOGICIGNITER_AUTONOMOUS_DELIVERY_LOOP.md`
  capturing §13 Step 2 mechanics in canonical form.

**Updates to existing memory**:
- `.picoclaw/workspace/memory/ZEHN_CURRENT_STATE.md` — append the
  explicit-decision notes from §13 Step 7.
- `.picoclaw/workspace/memory/LOGICIGNITER_YAAD_SCHEMA_CONTRACT.md` —
  add explicit read-first / write-back operating rules per §13 Step 7a.
- `li-ceo`, `li-coo`, and specialist `AGENT.md` files — add the one-line
  Yaad read-first / write-back requirement per §13 Step 7a.

**Go code**:
- No code change is required for the recovery plan above. If the
  `Unknown channel for outbound message` warning is operationally noisy,
  `pkg/channels/manager.go:1128` can be quieted by routing internal
  delegation responses to the delegation inbox channel; that is a P2
  consideration, not a recovery prerequisite.

---

## 16. Explicit "Do Not Touch Yet" List

Until Ali approves the plan, do not:

1. **Start** Zehn, gateway, launcher, cron, Discord, or any external
   channel.
2. **Edit** any file under `.picoclaw/`, `operations/`, `supervision/`, or
   any `LogicIgniter` repo.
3. **Enable** any cron job in `jobs.json`.
4. **Bootstrap** the `io.picoclaw.launcher.plist` via `launchctl`.
5. **Delete** any historical delegation, session, or evidence file.
6. **Reformat** `gateway.log` or rotate it.
7. **Push** any branch or open any PR in any LogicIgniter subrepo from
   automation.
8. **Touch** secrets in `.picoclaw/.security.yml` or
   `.picoclaw/secrets/`.
9. **Run** any verification, restart, or post-merge script except the
   read-only inspection commands cited in this audit.
10. **Modify** Go source under `pkg/`, `web/`, or `cmd/` of the Zehn repo.

The four untracked supervision files added during the prior audit
(`supervision/ZEHN_COMPANY_OPERATING_MODEL_ROLE_AUDIT_20260512.md`,
`supervision/ZEHN_COMPANY_OPERATING_MODEL_ROLE_REMEDIATION_TODO_20260512.md`,
`supervision/ZEHN_EXEC_SAFETY_GUARD_INVESTIGATION_20260512.md`,
`supervision/ZEHN_STOPPED_RUNTIME_INVESTIGATION_20260513.md`) and the two
untracked operations verifiers (`operations/verify-logicigniter-cron-routing.sh`,
`operations/verify-zehn-role-personas.sh`) remain in place. This audit
adds one more untracked file (this document). All other files are
unchanged.

---

## 17. Decision Requested

Before any implementation, Ali to approve:

1. The revised Minimal Recovery Plan in §13 (steps 1–8 plus 7a), or a
   modified version. Notable changes from prior drafts: loop fix first;
   verify-pr.sh v1 is deliberately narrow (Go test + bash -n only, JSON
   evidence) with v2 expansion after the first drill; specialist sweep
   has a strict no-work-found reporting rule; Yaad read-first / write-back
   contract added; exec safety and open delegation recorded as explicit
   user decisions (not reversed).
2. **verify-pr.sh v1 scope (§13 Step 1)**: confirm v1 covers the right
   minimum — repo detection, changed-files diff, repo-local verify if
   present, Go `go test ./...`, shell `bash -n`, docs only-if-tools-exist,
   structured JSON evidence. v2 (integration_tests, MCP runtime proof,
   deep audits) deferred.
3. **§13 Step 2 mechanics**: confirm stale-claim auto-release window
   (proposed: 4 hours) and failure-retry threshold (proposed: 2 fails →
   `zehn:blocked`, 3 fails → escalate to li-coo).
4. **§13 Step 5 sweep owner**: `li-coo` (throughput accountability) vs.
   `li-operations` (GitHub project hygiene). Current plan: li-coo.
5. **First-drill repo (§13 Step 8)**: `scripts` (shell change), one
   `svc-*-grpc` (go test change), or `business` (doc/config change).
6. **§13 Step 7a Yaad contract**: confirm read-first / write-back is the
   intended posture for CEO, COO, and specialists, and that the drill
   should fail without two confirmed Yaad operations.
7. Whether `zehn-main` should acquire one bounded read-only Zehn-log
   diagnostic capability (currently forbidden by `HEARTBEAT.md`). This
   would let the operations monitor catch failure modes heartbeat misses.

Once approved, implementation will proceed in the order specified, with
each step verified before the next is started. No restart will occur
until §13 Steps 1–7 are complete and one drill issue is staged.
