# Zehn Yaad Clutter Prevention Audit - 2026-05-18

Status: investigation only; no config, prompt, code, GitHub, or Yaad mutations.

## Purpose

Identify why Zehn can still create Yaad clutter, duplicate memory, stale local
memory, or invalid Yaad references after the 2026-05-18 Yaad cleanup.

Canonical LogicIgniter Yaad scope:

```json
{"scope_type":"organization","external_key":"logicigniter"}
```

## Sources Inspected

- `.picoclaw/config.json`
- `.picoclaw/workspace/cron/jobs.json`
- `.picoclaw/workspace/HEARTBEAT.md`
- `.picoclaw/workspace/memory/LOGICIGNITER_YAAD_SCHEMA_CONTRACT.md`
- `.picoclaw/workspace/memory/ZEHN_CURRENT_STATE.md`
- `.picoclaw/workspace/memory/LOGICIGNITER_WORK_SELECTION.md`
- `.picoclaw/workspace/operating-prompts/logicigniter-ceo-daily-sync.md`
- `.picoclaw/workspace/operating-prompts/zehn-operations-monitor.md`
- `.picoclaw/logs/gateway.log`
- `pkg/agent/delegation_memory.go`
- `pkg/config/config.go`
- agent local memory files under `.picoclaw/workspace-*/memory/MEMORY.md`
- Direct read-only Yaad MCP calls via `https://yaad.mmaliabbas.com/mcp`
  using the existing agent token environment; no Yaad writes were performed.

## Current Safeguards That Are Working

- The current schema contract correctly names `organization:logicigniter`.
- The current schema contract explicitly rejects invalid scope type `company`.
- The current schema contract explicitly rejects unsupported memory classes such
  as `event`, `episodic`, `operating_state`, `operating_status`,
  `operational_finding`, `project_note`, `risk`, and `semantic`.
- Recent logs show valid successful `mcp_yaad_memory_add`,
  `mcp_yaad_profile_upsert`, and `mcp_yaad_scope_list` calls after Yaad became
  reachable again.
- Current state doc says old `li-app-*` agents are historical and current 51
  app records are portfolio context, not persistent app-owner agents.

## Re-Audit Corrections

The first pass of this report overstated one important point: several
near-duplicate timestamps in the gateway log are **not proven duplicate
successful Yaad writes**. Deeper correlation shows many of them are invalid
schema attempts followed by a corrected retry. The durable-memory issue is still
real, but the evidence supports a narrower conclusion:

- Proven: agents still make invalid Yaad calls (`binding_mode`, `memory_class`,
  and `scope_type`) and then often retry successfully.
- Proven: repeated no-work heartbeat/monitor summaries are being written to
  local memory and Yaad.
- Proven: built-in delegation memory is idempotent at the local delegation
  record level, but its primary configured scope is `project:zehn`, not
  `organization:logicigniter`.
- Not proven from log pairs alone: every close timestamp pair represents a
  duplicate successful Yaad memory. Confirming semantic duplicates requires
  Yaad memory IDs or result-level queries, not just gateway timestamps.

The report below has been revised to keep these distinctions explicit.

## Strict Re-Audit Evidence Table

This section separates runtime facts from prompt intent and from Yaad-side
effects.

| Area | Evidence | Finding | Confidence |
|---|---|---|---|
| Active agents | `.picoclaw/config.json` has 42 agents in `agents.list`; `jq -r '.agents.list[].id'` returns no `li-app-*`; `find .picoclaw -maxdepth 1 -name 'workspace-li-app-*'` returns 0. | Local runtime no longer has persistent 51 app-owner agents. | High |
| Active org hierarchy | `.picoclaw/config.json` `agents.organization.nodes` lists executives, departments, 10 bundle owners, and specialists. | Current organization model is represented in config. | High |
| Heartbeat runtime path | `pkg/gateway/gateway.go` creates a heartbeat handler that calls `agentLoop.ProcessHeartbeat`; `pkg/heartbeat/service.go` builds the prompt and invokes the handler; `pkg/agent/agent_message.go` uses the default agent with `SessionKey: "heartbeat"`, `NoHistory: true`, and no tool denylist. | Heartbeat no-`exec` is not runtime-enforced in the inspected path. | High |
| Heartbeat prompt | `.picoclaw/workspace/HEARTBEAT.md` line-level content says `Do not use shell execution (exec) during heartbeat`. | The intended heartbeat policy is no shell execution. | High |
| Heartbeat actual behavior | Gateway log around `2026-05-18T17:27` shows `li-coo` in `internal:delegation:zehn-main:li-coo:logicigniter-heartbeat-work-check` using `exec`, including long filesystem/GitHub sweep commands. | Heartbeat-triggered delegated work violated the prompt-level no-shell rule. | High |
| Cron runtime path | `pkg/tools/cron.go` runs `payload.command` through exec when present; otherwise it calls `ProcessDirectWithChannel` with sender `cron`. | Agent cron turns have normal tool availability unless prompt/config restricts them. | High |
| Operations monitor prompt | `.picoclaw/workspace/operating-prompts/zehn-operations-monitor.md` explicitly says `Use simple exec commands when useful`. | Operations-monitor exec usage is expected, not itself a contradiction. | High |
| Current crons | `.picoclaw/workspace/cron/jobs.json` has six recurring enabled jobs and two enabled one-shot `at` jobs. One-shot times resolve to `2026-05-18 20:01:12 +0500` and `2026-05-20 06:03:11 +0500`. | There are still Release Ladder one-shot follow-ups in the active cron file. | High |
| Built-in delegation memory | `pkg/agent/delegation_memory.go` writes Yaad `memory_add` only for terminal statuses and skips if local durable memory status is already `written`. | Built-in delegation memory is locally idempotent for completed local records. | High |
| Built-in delegation scope | `pkg/agent/delegation_memory.go` builds scopes `project:<metadata.ProjectKey>`, `agent:<parent>`, and `agent:<target>`; `.picoclaw/config.json` sets `project_key: zehn`. | Built-in delegation summaries are on the Zehn project track, not the LogicIgniter organization track. | High |
| Yaad project:zehn records | Read-only Yaad `memory_browse` on `project:zehn` with label `delegation` returned 20 active delegation summary records, e.g. `zehn-main to li-coo` and specialist-to-specialist delegations. | The Zehn delegation memory track exists and is populated separately from org memory. | High |
| Yaad org no-work clutter | Read-only Yaad `memory_browse` on `organization:logicigniter` with labels `coo-heartbeat`, `work-selection`, `no-claimable-work`, `no-work-found` returned `count=50` active records. | Repeated no-work/work-selection memories are proven in Yaad, not only inferred from logs. | High |
| Yaad stale app-agent check | Read-only Yaad query for `li-app app owner persistent agents old 51` returned the current authoritative hybrid-agent memory `1a1aaa3c-...` among top results, stating 51 apps are not persistent `li-app-*` agents. | The top active org memory now reinforces the corrected specialist/bundle model. This does not prove no stale `li-app-*` record exists anywhere. | Medium |
| Invalid Yaad schema | Gateway log contains post-contract invalid calls: `binding_mode` values `explicit`, `strong`, `primary`; `memory_class` values `episodic`, `event`; `scope_type` `company`. | Agents still sometimes invent invalid Yaad schema values, then often retry successfully. | High |
| Prompt schema contract | `.picoclaw/workspace/memory/LOGICIGNITER_YAAD_SCHEMA_CONTRACT.md` correctly lists allowed classes and forbids invented scope/binding values. | The core schema contract is mostly correct. | High |
| Prompt write pressure | The same schema contract says "Every terminal outcome must produce one durable Yaad write"; many `AGENT.md` files repeat "Write durable Yaad memory after every terminal outcome"; `LOGICIGNITER_WORK_SELECTION.md` says each terminal state writes Yaad before returning. | Current prompts push broad write frequency and do not define a stable idempotency key/update rule. | High |
| No-work reporting pressure | `LOGICIGNITER_WORK_SELECTION.md` requires strict no-work-found reports before `HEARTBEAT_OK`; Yaad evidence shows at least 50 active no-work/work-selection memories. | The no-work reporting doctrine is contributing to durable clutter. | High |

## Strict Re-Audit Conclusions

1. **The strongest proven problem is not stale app-agent state; it is durable
   no-work/work-selection clutter.** Yaad has at least 50 active
   `organization:logicigniter` records matching COO heartbeat/work-selection
   no-work labels.
2. **The second proven problem is prompt-only enforcement.** Heartbeat forbids
   `exec`, but the heartbeat runtime path exposes normal tools, and a
   heartbeat-triggered COO delegation used `exec` anyway.
3. **The third proven problem is schema misuse despite a correct contract.**
   Invalid calls still happen, but many are corrected on retry. The fix target
   is the first attempted call and the prompt/tool contract exposure, not just
   fallback behavior.
4. **The built-in delegation writer is not the source of duplicate terminal
   writes in the way the first pass implied.** It writes terminal statuses and
   skips already-written local records. Its real concern is scope split:
   `project:zehn` for built-in delegation summaries versus
   `organization:logicigniter` for company memories.
5. **The current active config has no persistent `li-app-*` agents locally.**
   Active Yaad top results also include the corrected hybrid-agent doctrine, but
   a complete proof that no stale `li-app-*` memory exists anywhere would need a
   full paginated Yaad search/export.

## Remaining Unknowns

- Exact total count of stale or duplicate Yaad memories across all pages and
  all labels. The direct queries above prove at least 50 relevant no-work
  records, but not the total corpus size.
- Whether Yaad supports a practical update/upsert key pattern that Zehn agents
  can use today without code changes. `memory_update` exists, but prompts do not
  currently require agents to query/update an existing rolling record before
  adding a new one.
- Whether heartbeat tool restriction should be solved by prompt, by routing
  heartbeat through a narrower tool registry, or by changing the heartbeat
  delegation prompt to make `li-coo` use read-only/file/Yaad/GitHub tools only.
- Whether the two enabled one-shot Release Ladder cron jobs are still useful or
  stale. They are active in `jobs.json`; this audit did not evaluate their
  business relevance.

## Findings

### F-001: Heartbeat says no shell, but heartbeat-driven flows still use shell

Evidence:

- `.picoclaw/workspace/HEARTBEAT.md` says:
  "Do not use shell execution (`exec`) during heartbeat."
- `pkg/heartbeat/service.go` builds the heartbeat prompt and calls the handler;
  it does not apply a heartbeat-specific tool denylist.
- `pkg/agent/agent_message.go` runs heartbeat through the default agent with
  `SessionKey: "heartbeat"`, `NoHistory: true`, `SuppressToolFeedback: true`,
  and `SendResponse: false`; it does not restrict the tool registry.
- Gateway log shows heartbeat/session-driven `zehn-main` and delegated `li-coo`
  runs using `exec`.
- Concrete recent example:
  - `2026-05-18T17:27:19+05:00`
  - agent `li-coo`
  - chat `internal:delegation:zehn-main:li-coo:logicigniter-heartbeat-work-check`
  - tool `exec`
  - command reads `LOGICIGNITER_WORK_SELECTION.md` and lists
    `/Users/aliai/logicigniter`.

Impact:

- This contradicts the heartbeat operating boundary.
- The boundary is prompt-only in the inspected runtime path, so the model can
  violate it unless the prompt is tightened or runtime/tool routing changes.
- It turns each heartbeat into a broad filesystem/GitHub scan, which then often
  produces local artifacts and Yaad summaries.
- It increases repeated "no claimable work" memory writes.

### F-002: COO heartbeat creates too many durable no-work records

Evidence:

- `.picoclaw/workspace-li-coo/memory/MEMORY.md` is 518 lines after the latest
  observed heartbeat append.
- It contains repeated timestamped "COO heartbeat work-check" entries.
- `.picoclaw/workspace-li-coo/.artifacts` contains 374 files.
- Recent Yaad calls include many `li-coo` memory add/profile attempts for
  repeated work-selection scans. Some timestamps below are failed attempts that
  were followed by corrected retries; the point here is repeated write pressure,
  not that every listed timestamp produced a successful memory:
  - `2026-05-17T23:29:16+05:00`
  - `2026-05-18T00:30:08+05:00`
  - `2026-05-18T01:29:18+05:00`
  - `2026-05-18T02:01:39+05:00`
  - `2026-05-18T02:01:51+05:00`
  - `2026-05-18T03:00:53+05:00`
  - `2026-05-18T03:01:11+05:00`
  - `2026-05-18T16:50:51+05:00`
  - `2026-05-18T17:07:07+05:00`

Impact:

- Yaad becomes an archive of operational scans instead of a compact durable
  memory layer.
- Retrieval can surface repeated "0 claimable" or old blocked-state records.
- Local boot memory also grows with repeated facts.

### F-003: Retry pairs and semantic duplication risk are real; duplicate writes need result-level proof

Evidence from gateway log:

- `li-qa` attempted a Yaad write at `2026-05-17T17:43:17/18+05:00`; the first
  attempt failed with `invalid binding_mode "primary"`, then the agent retried.
- `zehn-main` attempted a Yaad write at `2026-05-17T22:16:41+05:00`; that
  attempt failed with `invalid binding_mode "strong"`, then the agent retried
  around `22:16:56`.
- `li-qa` attempted a Yaad write at `2026-05-18T00:54:22+05:00`; that attempt
  failed with `invalid binding_mode "primary"`, then the agent retried around
  `00:54:33`.
- `li-coo` attempted a Yaad write at `2026-05-18T02:01:39/40+05:00`; the first
  attempt failed with `invalid binding_mode "explicit"`, then the agent retried.
- `li-coo` attempted a Yaad write at `2026-05-18T06:14:18/19+05:00`; the first
  attempt failed with `invalid memory_class "episodic"`, then the agent searched
  for the schema contract and succeeded at `06:14:40/41`.

Impact:

- The system wastes tool iterations on invalid calls, then often recovers.
- The retries can still create semantic clutter when the same no-work,
  monitor, or review state is written repeatedly across cycles.
- However, the close timestamp pairs above should not be treated as proven
  duplicate successful memories without querying Yaad memory IDs/results.

### F-004: Invalid Yaad schema calls still happened after the schema contract

Evidence from gateway log:

- `2026-05-18T02:01:40+05:00`: `invalid binding_mode "explicit"`.
- `2026-05-18T02:44:10+05:00`: `invalid memory_class "episodic"`.
- `2026-05-18T02:44:21+05:00`: `invalid memory_class "event"`.
- `2026-05-18T03:00:53+05:00`: `invalid memory_class "episodic"`.
- `2026-05-18T06:05:20+05:00`: `invalid binding_mode "strong"`.
- `2026-05-18T06:06:50+05:00`: `invalid scope_type "company"`.
- `2026-05-18T06:09:49+05:00`: `invalid memory_class "event"`.
- `2026-05-18T06:14:19+05:00`: `invalid memory_class "episodic"`.

Impact:

- The schema contract exists, but not every prompt/run path reliably follows it.
- Invalid calls waste tool iterations and may trigger fallback local-memory
  writes.

### F-005: Delegation durable memory uses `project:zehn`, not `organization:logicigniter`

Evidence:

- `.picoclaw/config.json` has:
  - `agents.defaults.delegation_memory.metadata.project_key = "zehn"`
  - labels include `zehn`, `logicigniter`, `delegation`
  - source is `zehn-delegation`
- `pkg/agent/delegation_memory.go` always creates a project scope for
  delegation memory:
  - `scope_type: project`
  - `external_key: metadata.ProjectKey`
- `pkg/agent/delegation_memory.go` only writes terminal statuses
  (`completed`, `failed`, `cancelled`) and skips writing if the local delegation
  record already has durable memory status `written`.
- `pkg/config/config.go` exposes only `project_key`, `labels`, and `source` for
  delegation memory metadata. It does not expose a config-only way to make
  delegation memory use `organization:logicigniter`.

Impact:

- LogicIgniter delegation summaries written by the built-in delegation writer
  are not naturally under the canonical LogicIgniter organization scope.
- Agents also manually write LogicIgniter summaries under
  `organization:logicigniter`, creating two parallel memory tracks.
- This is a scope split problem, not evidence that the built-in delegation
  writer itself is repeatedly adding duplicate terminal records.
- Fixing this cleanly may require a code/config design change, not just prompt
  edits.

### F-006: Zehn operations monitor is instructed to write a Yaad summary every actionable hour

Evidence:

- `.picoclaw/workspace/operating-prompts/zehn-operations-monitor.md` says full
  findings should be written as a Yaad `summary` under
  `organization:logicigniter`.
- Cron fires `zehn-operations-monitor-v2` hourly at minute 15.
- Recent gateway log shows hourly `zehn-main` summary write attempts; some of
  the close pairs include a failed invalid-schema attempt followed by a
  corrected retry:
  - `2026-05-17T17:23:46+05:00`
  - `2026-05-17T18:17:07+05:00`
  - `2026-05-17T19:17:36+05:00`
  - `2026-05-17T20:18:37+05:00`
  - `2026-05-17T21:24:48+05:00`
  - `2026-05-17T22:16:41+05:00`
  - `2026-05-17T22:16:56+05:00`
  - `2026-05-17T23:19:53+05:00`
  - `2026-05-18T00:27:00+05:00`
  - `2026-05-18T01:18:15+05:00`
  - `2026-05-18T02:25:53+05:00`
  - `2026-05-18T03:23:14+05:00`
  - `2026-05-18T04:18:50+05:00`
  - `2026-05-18T05:21:00+05:00`
  - `2026-05-18T17:22:08+05:00`

Impact:

- This is likely too noisy for long-term durable memory.
- It should be a rolling profile, local log, or only write when state changes.

### F-007: Local fallback memory grows with transient Yaad outages

Evidence:

- `li-architect` local memory includes several fallback records with:
  "Yaad unavailable during review/write: MCP connection closed..."
- `li-docs` local memory includes similar fallback records.
- Gateway log from 2026-05-17 morning shows many `MCP client is closing`
  failures for Yaad tools.

Impact:

- Local memory files become long operational archives.
- Once Yaad returns, there is no visible automatic local-fallback flush and
  compaction policy.

### F-008: Some active prompt text still says "write Yaad every terminal outcome"

Evidence:

- Multiple `AGENT.md` files contain the shared doctrine:
  "Write durable Yaad memory after every terminal outcome..."
- `LOGICIGNITER_YAAD_SCHEMA_CONTRACT.md` says every terminal outcome must
  produce one durable Yaad write.

Impact:

- This is correct for real terminal outcomes, but too broad without an
  idempotency rule.
- In practice, repeated review/re-review/meeting/heartbeat terminal states can
  each produce separate durable records for the same issue/PR unless the agent
  first checks for an existing terminal record.

## Immediate Fix Plan

1. Tighten memory-write policy in the schema contract.
   - Add a "durable write selectivity" section.
   - Terminal issue/PR outcomes: write once per repo/issue/PR/stage.
   - Re-reviews: update existing memory or skip if no material state changed.
   - No-work scans: do not write Yaad unless the queue state changed, a blocker
     appeared/disappeared, or an action was taken.

2. Change COO heartbeat behavior from append-only memory to rolling state.
   - Use a stable local state file/profile for current queue state.
   - Store full scan artifacts locally with retention.
   - Write Yaad only on material transitions.

3. Change Zehn operations monitor behavior to rolling summary.
   - Hourly monitor should not write a new Yaad summary just because an existing
     known blocker still exists.
   - Write Yaad only for new failure class, resolved failure, new blocked owner,
     restart, outage, or material progress.

4. Add explicit idempotency instructions to all active role prompts.
   - Before writing terminal memory, query by repo/issue/PR and outcome key.
   - If an equivalent memory exists, reference it instead of adding another.
   - If the state changed, update/supersede rather than duplicate.

5. Fix the delegation durable-memory scope design.
   - Current config cannot make built-in delegation memory use
     `organization:logicigniter`.
   - Options:
     - add configurable primary scope type/external key in code, or
     - disable built-in delegation Yaad write for LogicIgniter and require
       explicit org-scoped terminal writes, or
     - keep `project:zehn` but treat it as Zehn-system provenance, not
       LogicIgniter company memory.

6. Add a fallback-local-memory flush/retention policy.
   - Local fallback entries should be short queue entries.
   - Once Yaad write succeeds, mark local fallback flushed or superseded.
   - Avoid appending full review reports into `MEMORY.md`.

7. Add a runtime monitor check for invalid Yaad schema calls.
   - Count invalid `scope_type`, `memory_class`, and `binding_mode` errors.
   - If any appear in the last operating window, report exact agent/prompt path
     instead of writing another broad health summary.

## Recommended First Implementation Order

1. Prompt/config-only cleanup first:
   - memory write selectivity;
   - no-work scan write suppression;
   - operations monitor write-on-change;
   - idempotency-before-write instruction.
2. Observe one operating window.
3. Then decide whether code is needed for delegation memory configurable scopes.

Do not start by changing Go code unless prompt/config changes cannot stop the
clutter. The only finding that clearly points at a code/config limitation is
F-005.
