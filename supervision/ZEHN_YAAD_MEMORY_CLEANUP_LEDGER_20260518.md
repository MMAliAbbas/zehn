# Zehn Yaad Memory Cleanup Ledger - 2026-05-18

Status: read-only candidate ledger; no Yaad data changed.

Related plan:
`supervision/ZEHN_YAAD_MEMORY_CLEANUP_PLAN_20260518.md`

Canonical scope:

```json
{"scope_type":"organization","external_key":"logicigniter"}
```

## Read-Only Evidence

Read-only Yaad ledger pass:

- Time: 2026-05-18 17:16-17:17 local.
- Tool path: Zehn CLI through configured Yaad MCP.
- Mutations requested: none.
- Mutations observed: none.
- Result: compact candidate ledger returned by Yaad read/list/query tools.
- Runtime behavior note: the pass still required tool discovery and five LLM
  iterations, but was materially better than the earlier broad audit.

Earlier read-only audit confirmed:

- `organization:logicigniter` exists:
  `c3434c10-2330-4ee3-a85c-50dd5fdd7b60`.
- Valid scope types: `global`, `organization`, `project`, `agent`,
  `user_persona`, `agent_persona`.
- Yaad MCP exposes 16 tools.

## Candidate Ledger

| ID | Title / Scope | Class | Created By | Issue Type | Classification | Evidence | Proposed Action |
|---|---|---|---|---|---|---|---|
| `901cf6d9-beab-4030-81a8-033dc4be5087` | LogicIgniter Repo Access Doctrine | `decision` | unknown | current repo/path doctrine | `KEEP_CURRENT` | Says `/Users/aliai/logicigniter` is the live repo source of truth. | Keep current; useful anchor for path confusion. |
| `1a1aaa3c-a3f7-41db-bbe0-8fa4577df476` | LogicIgniter Zehn hybrid agent model is authoritative | `decision` | `codex` | old `li-app/app-owner` correction | `KEEP_CURRENT` | Says 51 apps are portfolio context, not persistent `li-app-* app-owner` agents. | Keep as authoritative routing memory. |
| `a99f0cc9-d5cf-463c-9186-023c94ceef1a` | Zehn specialist GitHub issue routing policy | `decision` | `zehn-main` | current specialist routing | `KEEP_CURRENT` | Defines `zehn:ready`, specialist area labels, and claim flow. | Keep current. |
| `cd493160-cc3a-4dfd-b9c0-7750e1e42ceb` | Zehn executable GitHub work policy | `decision` | `zehn-main` | current GitHub execution policy | `KEEP_CURRENT` | Defines issue labels, branch/verify/PR/review rules, and no dirty repos. | Keep current. |
| `10be32dd-82c0-48ca-9f42-9db028b30556` | INIT-20260517-ignite-videoedit-studio-m1-engine-planning | `decision` | `li-ceo` | current active initiative | `KEEP_CURRENT` | Highest-priority Ignite Videoedit Studio M1 plan; references issues #25-#34, PR #35, and delegated #26. | Keep current until a newer initiative status supersedes it. |
| `790987bd-0b24-40da-ae34-e01a132d8a65` | INIT-20260517-ignite-videoedit-studio-m0-planning | `decision` | `li-ceo` | superseded initiative phase | `KEEP_HISTORICAL` | M0/M0.5 issues #1-#11; later M1 memory exists. | Keep as historical phase record; reduce retrieval priority if possible. |
| `2035455c-b3ef-46c1-8011-f90c889a9591` | LogicIgniter P0 company focus on MCP/auth final-readiness blocker | `decision` | `li-ceo` | aging directive | `KEEP_HISTORICAL` | Ali directive to focus company on MCP/auth blocker; later initiative focus shifted to Videoedit M1. | Keep historical unless a newer P0 directive explicitly supersedes it. |
| `bafe4973-3a3f-453e-80e8-6e11ee5e2b4f` | Current-root MCP repo absence blocks LogicIgniter final launch-readiness claim | `decision` | `li-ceo` | stale path and `svc-services-mcp missing` | `SUPERSEDE` | Says `/Users/aliai/logicigniter/svc-services-mcp` is missing and references old `/Users/ali/projects/logicigniter`; later checks report service exists/reachable. | Update as superseded or lower trust/importance; keep only as historical incident evidence. |
| `b820f7b7-0d2e-48a0-8620-3462e07db676` | LogicIgniter current-root MCP reconciliation blocks final launch claim | `decision` | unknown | stale `svc-services-mcp missing` | `SUPERSEDE` | Says `/Users/aliai/logicigniter/svc-services-mcp` is missing. | Supersede or merge into historical MCP reconciliation cluster. |
| `995a17b1-9857-4f44-b11e-ada7a3f59007` | LogicIgniter CEO Operating Check - Current-root MCP blocker persists 2026-05-10 07:05 | `decision` | `li-ceo` | duplicate stale operating check | `MERGE` | Lists inspected paths and says missing `/Users/aliai/logicigniter/svc-services-mcp`. | Merge into one historical May 10 MCP-path incident summary. |
| `db6ffcba-9b0a-48f9-9a46-aa966c81d89e` | Current-root final launch readiness blocked by missing svc-services-mcp | `decision` | unknown | duplicate stale operating check | `MERGE` | Says final verification hard-requires missing `svc-services-mcp`. | Merge into same historical MCP-path incident cluster. |
| `eced46b0-dc19-43cf-8a0c-b4f6586d2431` | LogicIgniter CEO operating check - current-root MCP blocker persists | `fact` | unknown | duplicate stale operating check | `MERGE` | Says `list_dir` confirmed `svc-services-mcp` missing. | Merge/deprioritize; do not treat as current. |
| `aac6dc01-8418-4ffe-b310-57bd1e51e645` | Current-root MCP reconciliation blocks LogicIgniter final launch readiness | `fact` | `li-cto` | historical evidence listing | `KEEP_HISTORICAL` | Lists files inspected for May 10 final-readiness evidence. | Keep if provenance is useful; ensure it does not rank as current blocker. |
| `560d01bc-a8eb-42d1-ab23-3fff454805fd` | QA evidence policy for P0 MCP/auth/final-readiness | `decision` | unknown | final-readiness evidence rule | `KEEP_HISTORICAL` | Says old `FINAL_READINESS_AUTOMATION_STATUS` is not enough final evidence; PR/test status from May 12. | Keep as historical QA evidence policy; refresh separately if still active. |
| `7235c318-23cc-4b51-bb7b-1ed3a9f0e190` | Yaad MCP read failure triage probe 2026-05-18 | `summary` | `li-cdao` | invalid schema/scope-type finding | `KEEP_HISTORICAL` | Records `scope_type company` is invalid and valid type is `organization`; Yaad read/upsert probe. | Keep as recent Yaad tooling triage; avoid treating as product fact. |
| `4f72b066-4968-4d00-a0a1-c285f0d22d5a` | Current-root MCP reconciliation blocks final launch-readiness claims | `decision` | `li-cto` | duplicate MCP blocker / access limitation | `MERGE` | Read final-readiness docs; GitHub tools unavailable; pinned May 10 blocker context. | Merge into historical MCP-readiness incident cluster and reduce noise. |
| `0037b467-a97f-4cee-9346-89ca0e247831` | Scope: LogicIgniter workspace | scope | n/a | stale scope/path `/Users/ali/projects/logicigniter` | `SCOPE_FIX` | Scope external key is old path; current doctrine says `/Users/aliai/logicigniter`. | Candidate for stale scope description/deprecation. Do not use for current operations. |
| `b2d23971-9eb6-4580-855a-536e2d198ec9` | Scope: `project:logicigniter` | scope | n/a | ambiguous duplicate project scope | `SCOPE_FIX` | Exists beside canonical `organization:logicigniter`. | Review whether needed; do not use for org-wide company facts. |
| unresolved | COO heartbeat / no-claimable cluster | unknown | mixed | possible low-value repeated operational summaries | `MERGE` or `DEACTIVATE` after exact IDs | Query produced oversized results or no inline candidate in this pass; gateway logs prove repeated COO heartbeat writes exist. | Run a narrower follow-up by labels/profile kinds, then merge repetitive no-op summaries into rolling profile. |
| unresolved | Invalid schema memory spam | unknown | mixed | invalid `memory_class` / invalid `scope_type` diagnostics | `KEEP_HISTORICAL` for one diagnostic, `DEACTIVATE` for duplicates | Concrete diagnostic row is `7235c318...`; logs also show failed writes with invalid classes `event` and `episodic`, but failed writes should not exist as Yaad memories unless later captured by summaries. | Keep one diagnostic memory; remove/deprioritize duplicates only after exact IDs are retrieved. |

## Proposed Mutation Batch

Do not apply until reviewed.

Batch A - Safe Current-State Profile:

- Upsert profile `logicigniter_current_operating_state` under
  `organization:logicigniter`.
- Content should include:
  - current repo root `/Users/aliai/logicigniter`;
  - current org model: executive, department, bundle, specialist roles;
  - current active initiative: Ignite Videoedit Studio M1, unless superseded;
  - current GitHub issue/PR work policy;
  - valid Yaad schema constraints.

Batch B - Supersede Stale MCP Missing Records:

- Update `bafe4973-3a3f-453e-80e8-6e11ee5e2b4f`.
- Update `b820f7b7-0d2e-48a0-8620-3462e07db676`.
- Mark title/content as `SUPERSEDED`.
- Lower trust/importance if supported.
- Preserve evidence and original date.

Batch C - Merge Duplicate MCP Incident Records:

- Consolidate `995a17b1-9857-4f44-b11e-ada7a3f59007`,
  `db6ffcba-9b0a-48f9-9a46-aa966c81d89e`,
  `eced46b0-dc19-43cf-8a0c-b4f6586d2431`,
  `4f72b066-4968-4d00-a0a1-c285f0d22d5a`.
- Preferred action: add one historical summary and deactivate or lower
  priority on duplicates.

Batch D - Scope Cleanup:

- Mark `project:/Users/ali/projects/logicigniter` stale if Yaad supports safe
  scope metadata update.
- Decide whether `project:logicigniter` has a useful current purpose.

Batch E - Heartbeat Noise Follow-Up:

- Query exact profile kinds and labels for COO heartbeat records.
- Keep one rolling COO heartbeat/current queue profile.
- Deactivate or merge repeated no-op summaries older than the operating window.

## Risks

- Over-cleaning could remove useful incident provenance. Prefer supersede/merge
  over delete.
- Existing agents may still write noisy memory unless prompts/runtime checks are
  fixed.
- If profile reads remain broad, a clean profile can still become too large.
  Keep current-state profile short and deterministic.

## Verification After Approved Cleanup

- Query `svc-services-mcp missing`; stale blockers must be marked superseded or
  historical.
- Query `/Users/ali/projects/logicigniter`; old path must be stale-labeled.
- Query `current LogicIgniter operating state`; result must return current root,
  current agent model, current active initiative, and GitHub work policy.
- Watch one heartbeat cycle; no new invalid Yaad schema errors should appear.
- Watch one COO work-selection run; it should read current profile first and not
  write a new timestamped profile unless something materially changed.

## Cleanup Execution - 2026-05-18

Execution path: Zehn CLI session using Yaad MCP tools only. No deletes were
performed. No filesystem, GitHub, Discord, shell, or repository mutations were
requested inside the Zehn cleanup session.

Applied changes:

- Upserted profile `logicigniter_current_operating_state` under
  `organization:logicigniter`.
  - Profile ID: `0c2c2592-2f09-4135-8242-c8e9a06ac480`
  - Version: `1`
  - Trust: `0.95`
- Updated `bafe4973-3a3f-453e-80e8-6e11ee5e2b4f`.
  - Title: `SUPERSEDED - Current-root MCP repo absence blocks LogicIgniter final launch-readiness claim`
  - Version: `2`
  - Trust: `0.35`
  - Importance: `0.2`
- Updated `b820f7b7-0d2e-48a0-8620-3462e07db676`.
  - Title: `SUPERSEDED - LogicIgniter current-root MCP reconciliation blocks final launch claim`
  - Version: `2`
  - Trust: `0.35`
  - Importance: `0.2`
- Added historical summary `d54e20e6-dc80-4105-b9f4-cc4ffd7b00f3`.
  - Title: `Historical MCP path incident cluster - May 2026`
  - Class: `summary`
  - Labels: `yaad-cleanup,mcp,historical`
- Updated duplicate MCP incident memories:
  - `995a17b1-9857-4f44-b11e-ada7a3f59007`
  - `db6ffcba-9b0a-48f9-9a46-aa966c81d89e`
  - `eced46b0-dc19-43cf-8a0c-b4f6586d2431`
  - `4f72b066-4968-4d00-a0a1-c285f0d22d5a`
  - Each title now starts with `HISTORICAL MERGED -`.
  - Each was lowered to trust `0.4` and importance `0.2`.
- Updated stale old local path project scope.
  - Scope ID: `0037b467-a97f-4cee-9346-89ca0e247831`
  - External key: `/Users/ali/projects/logicigniter`
  - Display name: `STALE LogicIgniter old local path`
- Updated ambiguous project scope.
  - Scope ID: `b2d23971-9eb6-4580-855a-536e2d198ec9`
  - External key: `logicigniter`
  - Display name: `LogicIgniter project scope - limited use`

## Cleanup Verification - 2026-05-18

Read-only verification was run through Zehn CLI with Yaad read/list tools only.
Result: pass.

Verified records:

- Profile `logicigniter_current_operating_state`: `Current LogicIgniter operating state`
- Memory `bafe4973-3a3f-453e-80e8-6e11ee5e2b4f`:
  `SUPERSEDED - Current-root MCP repo absence blocks LogicIgniter final launch-readiness claim`
- Memory `b820f7b7-0d2e-48a0-8620-3462e07db676`:
  `SUPERSEDED - LogicIgniter current-root MCP reconciliation blocks final launch claim`
- Memory `d54e20e6-dc80-4105-b9f4-cc4ffd7b00f3`:
  `Historical MCP path incident cluster - May 2026`
- Project scope `/Users/ali/projects/logicigniter`:
  `STALE LogicIgniter old local path`
- Project scope `logicigniter`:
  `LogicIgniter project scope - limited use`

## Local Prompt Guardrails - 2026-05-18

Implemented local runtime prompt/memory guardrails to prevent Zehn from
recreating the same Yaad clutter pattern:

- Replaced blanket "write Yaad after every terminal outcome" instructions with
  selective/idempotent write-back rules.
- Added query-before-add behavior and stable outcome-key guidance in
  `.picoclaw/workspace/memory/LOGICIGNITER_YAAD_SCHEMA_CONTRACT.md`.
- Updated heartbeat, CEO daily sync, operations monitor, work-selection,
  meeting, scoreboard, release-ladder, approval, and non-engineering
  deliverable instructions to skip unchanged no-work scans, unchanged blockers,
  and duplicate summaries.
- Updated active specialist `AGENT.md` files and local `memory/MEMORY.md`
  policies to prefer Yaad reads while writing only material changed state.

Verification:

- No active high-risk blanket-write instruction remains for:
  `Write durable Yaad memory after every terminal outcome`,
  `Every terminal outcome must produce one durable Yaad write`,
  `Each terminal state writes Yaad before returning`,
  `Every meeting should update Yaad`, or `one Yaad ... write`.
- Remaining broad-search hit is a historical local fallback event in
  `.picoclaw/workspace-li-ceo/memory/MEMORY.md`, not an active instruction.
- No Go/source/runtime config changes were made for this guardrail pass.
