# Zehn Company Operating Model And Role Persona Remediation Todo - 2026-05-12

Status: role remediation complete; controlled live drill still pending.

Source audit:
`supervision/ZEHN_COMPANY_OPERATING_MODEL_ROLE_AUDIT_20260512.md`

Goal: convert Zehn's LogicIgniter setup from passive monitoring into a
company-like operating system where roles have clear authority, measurable
outputs, terminal outcomes, and role-specific judgment.

Non-negotiables:

- Do not change Go code unless a fact-checked runtime limitation remains after
  prompt/config/role remediation.
- Do not restart or reload Zehn until Ali asks.
- Keep external side effects approval-gated.
- Preserve approved internal software delivery authority for LogicIgniter repo
  work.
- Never leave repos dirty.
- Keep Yaad as canonical durable memory.
- Every role must know `/Users/aliai/logicigniter` is the live repo home.
- Every role must respect LogicIgniter architecture and prefer proper fixes
  over patches.
- Every role must support the company objective: maximize profit through
  portfolio and volume, not price alone.

## Phase 0 - Research And Guardrails

- [x] Re-open and summarize the current config sections for agents, cron,
  tools, GitHub/MCP, and delegation limits before editing role files.
- [x] Re-open current operating prompts and identify which ones must be kept in
  sync with role files.
- [x] Confirm whether runtime loads `AGENT.md`, `SOUL.md`, `USER.md`, and
  `memory/MEMORY.md` per turn or requires restart/reload for changes.
- [x] Capture a before/after inventory of every active role file.
- [x] Do not edit `.picoclaw` files until this phase is complete.

Research summary:

- Structured agents load `AGENT.md`, `SOUL.md`, `USER.md`, and
  `memory/MEMORY.md`; those paths are prompt-cache tracked.
- Current defaults allow broad local repo access:
  `allow_read_outside_workspace=true`, `restrict_to_workspace=false`,
  `max_tool_iterations=50`, `async_delegation.max_concurrent=9`.
- Current cron coverage is broad enough to run the company loop. The remaining
  issue is role/persona/operating-contract quality, not missing scheduled jobs.
- Config organization labels for bundle owners still use Ignite package names;
  bundle role rewrites must also decide how to present canonical suite names in
  `agents.organization`.

## Phase 1 - Canonical Company Operating Contract

- [x] Create or update a canonical internal operating contract for all
  LogicIgniter agents.
- [x] Define terminal outcomes:
  - merged;
  - reviewed/approved;
  - blocked with owner, blocker, and retry date;
  - escalated with precise approval question;
  - delegated with expected evidence;
  - deferred with reason and review date;
  - replaced by a better issue/plan with stale item handled.
- [x] Define "changed-state reporting": every scheduled run must say what
  changed, what remained blocked, and what will be checked next.
- [x] Define "real company mode": internal execution is expected; only external
  commitments and high-risk changes are gated.
- [x] Replace passive setup posture with development-phase internal
  deliverables.

Evidence:

- Added
  `.picoclaw/workspace/memory/LOGICIGNITER_COMPANY_OPERATING_CONTRACT.md`.
- Updated
  `.picoclaw/workspace/memory/LOGICIGNITER_ORGANIZATION_TREE.md` to reference
  the active company operating contract and replace passive groundwork framing
  with active internal-department behavior.

## Phase 2 - Executive Role Rewrite

Handcraft all four files where present: `AGENT.md`, `SOUL.md`, `USER.md`,
`memory/MEMORY.md`.

- [x] `li-ceo`: company outcomes, prioritization, approval routing, no endless
  diagnosis, delegates with terminal expectations.
- [x] `li-coo`: throughput owner, WIP aging, stale item cleanup, handoff
  quality, operations scoreboard.
- [x] `li-cto`: technical strategy, architecture confidence, bounded
  diagnostics, specialist routing, proper-fix doctrine.
- [x] `li-cpo`: product portfolio readiness, all-51 launch coherence, bundle
  ownership, requirements quality.
- [x] `li-ciso`: security governance, auth/secrets/data risk, exception
  handling, review discipline.
- [x] `li-cmo`: positioning, demand-generation systems, launch narrative,
  content and campaign readiness without unapproved publishing.
- [x] `li-cro`: sales strategy, pipeline design, qualification, volume-profit
  operating model, no unapproved outreach.
- [x] `li-cfo`: pricing economics, margin, runway, portfolio profitability,
  approval-gated financial actions.
- [x] `li-legal`: contracts, privacy/compliance posture, policy gates,
  risk registers, no legal commitments.
- [x] `li-chro`: hiring system, role design, interview process, team capacity.
- [x] `li-cco`: onboarding, support readiness, retention, customer health.
- [x] `li-cdao`: Yaad/memory/data/AI governance, analytics, evaluation.

Evidence:

- Rewrote executive `AGENT.md`, `SOUL.md`, and `USER.md` files for `li-ceo`,
  `li-coo`, `li-cto`, `li-cpo`, `li-ciso`, `li-cmo`, `li-cro`, `li-cfo`,
  `li-legal`, `li-chro`, `li-cco`, and `li-cdao`.
- Added role-specific active operating doctrine sections to each executive
  `memory/MEMORY.md`.
- Updated stale executive `IDENTITY.md` wording for CISO, CMO, CRO, and CHRO
  where it still described active roles as advisory/groundwork.

## Phase 3 - Department And Specialist Rewrite

- [x] `li-engineering`: technical execution manager, assigns specialists,
  enforces repo hygiene and verification.
- [x] `li-product`: product operating lane, acceptance criteria, service/app
  context quality.
- [x] `li-operations`: operating control plane, GitHub project hygiene,
  process compliance.
- [x] `li-marketing`: internal marketing production lane and launch assets.
- [x] `li-sales`: sales operating artifacts, ICP, pipeline, qualification,
  commercial feedback loop.
- [x] `li-finance`: financial analyst/controller lane under CFO.
- [x] `li-research`: market, competitor, buyer, and product research lane.
- [x] `li-customer-success`: support/onboarding/retention playbooks.
- [x] `li-docs`: user/admin/operator documentation lane and PR doc review.
- [x] `li-security`: implementation security review lane under CISO.
- [x] `li-devops`: local/service operations, post-merge reconcile, deployment
  and runtime hygiene.
- [x] `li-qa`: verification, evidence, acceptance, and release confidence.
- [x] `li-architect`: architecture memory file plus boundary/risk ownership.
- [x] `li-backend-developer`: backend issue/PR ownership and verification.
- [x] `li-frontend-developer`: frontend memory file plus UI issue/PR ownership.
- [x] `li-ux-designer`: UX memory file plus flow/copy/usability ownership.
- [x] `li-integration-engineer`: integration evidence and cross-repo runtime.
- [x] `li-data-ai-engineer`: data/AI memory file plus MCP/Yaad/eval ownership.

Evidence:

- Rewrote department `AGENT.md`, `SOUL.md`, and `USER.md` files for
  `li-engineering`, `li-product`, `li-operations`, `li-marketing`, `li-sales`,
  `li-finance`, `li-research`, `li-customer-success`, `li-docs`,
  `li-security`, `li-devops`, and `li-qa`.
- Rewrote specialist `AGENT.md`, `SOUL.md`, and `USER.md` files for
  `li-architect`, `li-backend-developer`, `li-frontend-developer`,
  `li-ux-designer`, `li-integration-engineer`, and `li-data-ai-engineer`.
- Added missing specialist memory files for `li-architect`,
  `li-frontend-developer`, `li-ux-designer`, and `li-data-ai-engineer`.
- Added active operating doctrine sections to all Phase 3 department and
  specialist memory files.
- Updated stale department identity wording for marketing and sales, and
  removed stale advisory wording from security memory.
- Verification passed for all Phase 3 roles: required boot files exist,
  `AGENT.md` references the company operating contract and live repo doctrine,
  memory has active operating doctrine, and active role files no longer contain
  stale `groundwork`, `advisory`, `passive`, or old `li-app-*` language.

## Phase 4 - Bundle Owner Rewrite

For each bundle owner, explicitly define:

- canonical original suite name;
- app/service coverage;
- provisional Ignite/package naming, if any;
- customer/persona/job-to-be-done;
- readiness gates;
- default collaboration with CPO, CTO, CRO, CMO, QA, Security, Docs, and
  specialists;
- all-51 launch dependency.

Bundle owners:

- [x] SaaS Growth & Retention Suite
- [x] E-commerce Operations Suite
- [x] Finance & Revenue Intelligence Suite
- [x] Legal & Compliance Suite
- [x] Developer & DevOps Suite
- [x] Content & Marketing Suite
- [x] HR & Workforce Suite
- [x] Professional Services Suite
- [x] Real Estate & Property Suite
- [x] Education Suite

Evidence:

- Rewrote bundle `AGENT.md`, `SOUL.md`, `USER.md`, `IDENTITY.md`, and
  `memory/MEMORY.md` files for all 10 canonical suites using Ali's original
  suite names and app lists.
- Removed stale `Ignite ...` package identities from bundle workspaces.
- Corrected `.picoclaw/config.json` agent names and organization labels for all
  bundle owners so registration and organization UI use the canonical suite
  names.
- Verification passed: config JSON parses, no stale Ignite bundle strings remain
  in config or bundle workspaces, each bundle agent references the company
  operating contract and live repo doctrine, and each bundle memory file has an
  active operating doctrine.

## Phase 5 - Personal And Main Zehn Roles

- [x] `zehn-main`: Zehn system monitor and maintainer only; not the
  LogicIgniter CEO substitute.
- [x] `personal`: Ali personal assistant; keep personal and LogicIgniter memory
  separated.

Evidence:

- Rewrote `zehn-main` `AGENT.md` and `memory/MEMORY.md` to define it as Zehn
  system monitor/router for runtime health, logs, MCP, providers, cron,
  heartbeat, channels, delegation, meeting tools, and automation quality.
- Updated `zehn-main` identity to say it is not the LogicIgniter CEO.
- Rewrote `personal` `AGENT.md` and `memory/MEMORY.md` to keep personal
  assistance separate from LogicIgniter company execution and consulting-job
  context.
- Removed active stale `Ignite ...`, CISO-advisory, and department-groundwork
  phrasing from `zehn-main` memory.
- Removed LogicIgniter execution doctrine blocks from `personal` active memory
  so it routes company execution back to `li-ceo` instead of acting as a company
  worker.

## Phase 6 - Prompt, Cron, And Memory Alignment

- [x] Update operating prompts to match the new terminal-outcome contract.
- [x] Ensure scheduled jobs ask for movement, not inspection-only reports.
- [x] Make DevOps/QA/Security/Docs review queues explicit.
- [x] Ensure CEO and CTO prompts are bounded.
- [x] Ensure specialist prompts select work by role labels, PR review need,
  stale blocked state, and repo hygiene.
- [x] Ensure all durable company facts are written or queued to Yaad using:
  `scope_type=organization`, `external_key=logicigniter`.

Evidence:

- Inspected active operating prompts and `.picoclaw/workspace/cron/jobs.json`.
- Confirmed cron jobs point at operating prompt files and ask for meaningful
  status, issue/PR inspection, specialist claim/execute behavior, review queues,
  dirty-repo discipline, and terminal reporting instead of passive reports.
- Corrected the CEO operating prompt from stale alternate-package language to
  the canonical 10-suite model using Ali's original suite names.
- Verified `.picoclaw/config.json` and `.picoclaw/workspace/cron/jobs.json`
  parse as valid JSON.
- Verified no stale alternate-package, setup-passive, advisory-only, or
  old persistent app-agent language remains in active operating prompts,
  cron jobs, or config.

## Phase 7 - Verification

- [x] Re-audit Phase 2 executive role files after Phase 3/4 rewrites to catch
  contradictions, stale wording, over-broad restrictions, missing operating
  authority, or accidental drift introduced during earlier edits.
- [x] Re-audit all newly rewritten Phase 3 department/specialist files after
  completion before treating the role rewrite as live-ready.
- [x] Add a role-persona verification script that fails on:
  - missing `AGENT.md`, `SOUL.md`, or `USER.md`;
  - missing specialist memory files;
  - stale setup-passive language in active business roles;
  - old 51 app-owner agent targets;
  - missing Yaad posture;
  - missing repo access doctrine;
  - missing terminal-outcome contract;
  - missing no-dirty-repo rule for execution roles.
- [x] Run the verification script locally.
- [ ] Run one controlled CEO operating drill:
  - CEO selects one outcome;
  - COO tracks movement;
  - CTO/CPO route technical/product work;
  - one specialist executes or reports a precise blocker;
  - QA/DevOps/Security/Docs review path is visible.
- [ ] Inspect logs after the drill for loops, capacity errors, skipped PRs, and
  non-terminal responses.

Evidence:

- Added `operations/verify-zehn-role-personas.sh`.
- Fixed verifier false positives so it checks stale setup language without
  rejecting intentional phrasing such as "not passive advisory".

## Phase 8 - No-Dump Active Memory Cleanup

- [x] Replace oversized active memory files instead of appending corrections.
- [x] Remove stale Ignite package naming from active portfolio references.
- [x] Convert CEO, CTO, CPO, COO, Product, Engineering, QA, DevOps, Operations,
  Backend, and Integration memory files into concise active operating doctrine.
- [x] Correct `ZEHN_SETUP_PLANNING.md` sections that had become active
  contradictions while preserving the planning record.
- [x] Strengthen `operations/verify-zehn-role-personas.sh` to fail on:
  - active memory over 90 lines;
  - append-style historical notes inside active memory;
  - stale Ignite portfolio identity in active boot/current-state files;
  - old persistent app-agent references in active role files.
- [x] Verify `.picoclaw/config.json` and `.picoclaw/workspace/cron/jobs.json`
  still parse as JSON.
- Added missing Yaad posture lines to backend and integration specialist memory.
- Ran `operations/verify-zehn-role-personas.sh`; result:
  `PASS: Zehn role persona verification passed`.

## Phase 8 - Yaad Update

- [ ] Write a concise organization-level Yaad memory that the role-persona
  rewrite was completed and names the canonical company operating contract.
- [ ] Do not dump all role files into Yaad; store durable facts and pointers.

## Done Criteria

- All 42 active workspaces have intentional, role-specific files.
- No active role is only a generic title plus safety disclaimer.
- Business roles produce internal deliverables, not just planning summaries.
- Specialists can pick, work, verify, and hand off role-appropriate work.
- CEO delegates with expected terminal outcomes.
- COO owns throughput and stuck-work cleanup.
- CTO and CPO stop broad unbounded diagnostics.
- Bundle owners use clear original-suite and packaging taxonomy.
- Cron prompts and role files agree.
- Verification passes.
- A controlled live drill proves movement instead of monitoring.
