# Zehn Yaad Memory Cleanup Plan - 2026-05-18

Status: proposed; no Yaad data changed by this plan.

## Purpose

Clean LogicIgniter durable memory so Zehn uses Yaad as a useful operating
memory instead of a noisy archive of stale blockers, repeated heartbeat scans,
and schema mistakes.

Canonical Yaad scope for LogicIgniter:

```json
{"scope_type":"organization","external_key":"logicigniter"}
```

Primary local contract:
`.picoclaw/workspace/memory/LOGICIGNITER_YAAD_SCHEMA_CONTRACT.md`

## Evidence From 2026-05-18 Read-Only Audit

Connectivity:

- Yaad MCP is reachable from Zehn with 16 tools.
- `organization:logicigniter` exists:
  `c3434c10-2330-4ee3-a85c-50dd5fdd7b60`.
- Direct read/write operations have recently succeeded from Zehn agents.

Stale or ambiguous scopes:

- `project:/Users/ali/projects/logicigniter` exists and references the old
  workspace path. It is stale for current operations.
- `project:logicigniter` exists and is ambiguous unless a current project-level
  purpose is explicitly assigned.

High-value current memories observed:

- `10be32dd-82c0-48ca-9f42-9db028b30556` -
  `INIT-20260517-ignite-videoedit-studio-m1-engine-planning`.
- `1a1aaa3c-a3f7-41db-bbe0-8fa4577df476` -
  `LogicIgniter Zehn hybrid agent model is authoritative`.
- `a99f0cc9-d5cf-463c-9186-023c94ceef1a` -
  `Zehn specialist GitHub issue routing policy`.
- `cd493160-cc3a-4dfd-b9c0-7750e1e42ceb` -
  `Zehn executable GitHub work policy`.

Stale or superseded memory candidates:

- `bafe4973-3a3f-453e-80e8-6e11ee5e2b4f` -
  `Current-root MCP repo absence blocks LogicIgniter final launch-readiness
  claim`; stale or partially superseded after `svc-services-mcp` became present.
- `b820f7b7-0d2e-48a0-8620-3462e07db676` -
  `LogicIgniter current-root MCP reconciliation blocks final launch claim`;
  stale or partially superseded for the same reason.
- May 10-11 `LogicIgniter CEO Operating Check` / final-readiness memories are
  historically useful, but noisy as current operating context unless marked
  superseded.

Runtime schema violations seen in gateway logs:

- Invalid `scope_type`: `company`.
- Invalid `memory_class`: `event`.
- Invalid `memory_class`: `episodic`.

Misuse patterns:

- Timestamped heartbeat records were written as profiles, for example
  `coo_heartbeat_20260518_0349`. Profiles should be stable deterministic
  summaries, not event logs.
- Repeated "no claimable work" summaries are being added to memory. These are
  useful for audit only when compacted; repeated raw entries dilute retrieval.
- Large unfiltered profile reads can produce payloads over 60KB and get omitted
  from model context. Agents must use narrow profile kinds or targeted memory
  queries instead.

## Cleanup Classification

Classify every candidate before changing it:

- `KEEP_CURRENT`: authoritative current policy, active initiative, live runbook,
  durable company fact, or current decision.
- `KEEP_HISTORICAL`: accurate historical evidence that should remain active only
  if it is clearly marked historical and does not rank above current facts.
- `SUPERSEDE`: once-accurate memory now contradicted by later verified evidence.
- `DEACTIVATE`: duplicate/noisy/obsolete memory that actively harms retrieval.
- `MERGE`: multiple repeated operational summaries that should become one compact
  rolling summary/profile.
- `SCOPE_FIX`: scope exists but external key or purpose is stale/ambiguous.

## Proposed Actions

1. Read-only inventory.
   - List all scopes matching `logicigniter`.
   - Browse active `organization:logicigniter` memories in pages.
   - Query for:
     - `svc-services-mcp missing`
     - `/Users/ali/projects/logicigniter`
     - `li-app-`
     - `app-owner`
     - `scope_type company`
     - `memory_class event`
     - `memory_class episodic`
     - `COO heartbeat`
     - `no claimable work`

2. Produce a candidate ledger.
   - For every candidate, record Yaad ID, title, class, created_by, scope,
     classification, evidence, and proposed action.
   - Do not modify Yaad during ledger creation.

3. Create one current operating profile.
   - Scope: `organization:logicigniter`.
   - Profile kind: `logicigniter_current_operating_state`.
   - Content should be short and stable:
     - canonical repo root: `/Users/aliai/logicigniter`;
     - Zehn org model: executive/department/bundle/specialist, not 51 app-owner
       agents;
     - current high-priority initiative: Ignite Videoedit Studio unless
       superseded;
     - current GitHub work routing policy;
     - current Yaad schema rules.
   - This profile should replace timestamped heartbeat profiles as the first
     retrieval target.

4. Supersede stale blocker memories.
   - For stale MCP/path memories, either:
     - update the memory title/content to prepend `SUPERSEDED`, or
     - deactivate it if Yaad supports active=false safely for the principal.
   - Preferred for auditability: update title/content plus lower trust/importance
     before deleting anything.

5. Clean stale scopes only after confirmation.
   - `project:/Users/ali/projects/logicigniter` should be deactivated/renamed
     only if Yaad supports safe scope lifecycle operations.
   - If scope deletion is not supported or risky, add a scope description:
     `STALE: historical old local path; do not use for current Zehn operations`.

6. Consolidate heartbeat summaries.
   - Keep at most one rolling COO/operations profile plus a small number of
     latest terminal summaries.
   - Old repeated "no claimable work" memories should become historical or
     inactive once their evidence is older than the operating window.

7. Prevent recurrence.
   - Add verification that live prompts/runbooks do not contain:
     - `scope_type: company`
     - `memory_class: event`
     - `memory_class: episodic`
   - Ensure agents use only approved classes from
     `LOGICIGNITER_YAAD_SCHEMA_CONTRACT.md`.
   - Require targeted `memory_context` / `memory_query` before broad profile
     reads.

## Verification Criteria

After cleanup:

- `memory_context` for current LogicIgniter operating state returns current
  root, current agent model, current work policy, and active initiative without
  stale MCP-missing claims.
- Query for `svc-services-mcp missing` returns only superseded/historical
  entries, not active current blockers.
- Query for `/Users/ali/projects/logicigniter` returns only stale-labeled
  historical references.
- Query for `li-app-* app-owner agents` returns the corrective hybrid-agent
  policy before any old app-owner delegation memories.
- Gateway logs over the next operating window show no new Yaad schema errors for
  invalid scope type or memory class.
- Heartbeat/COO writes are compact and do not create a new timestamped profile
  for every run.

## Safety Rules

- Do not delete memory first. Prefer update/supersede/deactivate.
- Do not change secrets, tokens, Yaad server config, or MCP config.
- Do not clean personal/user persona memory during this LogicIgniter pass.
- Do not treat historical incident evidence as current operating truth.
- Every Yaad mutation must be logged with memory ID, old classification, new
  classification, and evidence.

## Recommended Execution Order

1. Run read-only inventory and create `ZEHN_YAAD_MEMORY_CLEANUP_LEDGER`.
2. Review candidate ledger with Ali.
3. Apply only approved supersede/deactivate/profile changes.
4. Run retrieval verification.
5. Observe one heartbeat cycle and one CEO/COO work-selection cycle.

