# Zehn Company Operating Model And Role Persona Audit - 2026-05-12

Status: evidence pass complete; role remediation and no-dump cleanup applied.

Purpose: explain why the current LogicIgniter Zehn setup behaves like a
monitoring/reconciliation loop instead of a real operating company, and ground
the next role-file rewrite in file evidence rather than assumptions.

Rules for this audit:

- Do not restart, reload, or change runtime behavior during this audit.
- Do not edit agent persona files until the inventory and todo are explicit.
- Treat `.picoclaw/workspace-*` as runtime-local boot context, not Git-tracked
  product source.
- Findings must cite concrete paths or observed runtime behavior.
- Yaad remains the canonical durable memory target for durable company facts.

## Evidence Sources

Local files inspected:

- `.picoclaw/workspace*/AGENT.md`
- `.picoclaw/workspace*/SOUL.md`
- `.picoclaw/workspace*/USER.md`
- `.picoclaw/workspace*/memory/MEMORY.md`
- `.picoclaw/workspace/memory/LOGICIGNITER_ORGANIZATION_TREE.md`
- `.picoclaw/workspace/cron/jobs.json`
- Existing prompt/runbook audit:
  `supervision/ZEHN_PROMPT_MEMORY_RUNBOOK_AUDIT_20260512.md`
- Runtime loader code:
  - `pkg/agent/definition.go`
  - `pkg/agent/context.go`
- Current config:
  - `.picoclaw/config.json`

Recent runtime behavior previously observed from logs:

- Zehn generated many turns, tool calls, delegations, and `HEARTBEAT_OK`
  responses, but little terminal business progress.
- Many active GitHub items were already `zehn:in-progress` or
  `zehn:review-internal`, so queue checks often inspected and skipped rather
  than advancing work.
- PR review/check metadata was often empty or inconclusive, leaving work stuck
  between implementation, review, and merge.
- CEO/CTO/engineering checks repeatedly performed broad diagnostics instead of
  enforcing terminal outcomes.

## Inventory Snapshot

Workspace count:

- 42 active workspace directories matching `.picoclaw/workspace*`.

Role-file inventory:

- Every active workspace has `AGENT.md`, `SOUL.md`, and `USER.md`.
- All active workspaces now have `memory/MEMORY.md`.

File-size signal:

```text
.picoclaw/workspace-li-sales                 AGENT 2161  SOUL 134  USER 155  MEMORY 2394
.picoclaw/workspace-li-marketing             AGENT 2101  SOUL 140  USER 107  MEMORY 2391
.picoclaw/workspace-li-finance               AGENT 2161  SOUL 129  USER 148  MEMORY 2401
.picoclaw/workspace-li-cfo                   AGENT 2305  SOUL 151  USER 159  MEMORY 2408
.picoclaw/workspace-li-cro                   AGENT 2211  SOUL 137  USER 163  MEMORY 2427
.picoclaw/workspace-li-legal                 AGENT 2366  SOUL 162  USER 175  MEMORY 2477
.picoclaw/workspace-li-chro                  AGENT 2248  SOUL 147  USER 149  MEMORY 2411
.picoclaw/workspace-li-customer-success      AGENT 2206  SOUL 147  USER 157  MEMORY 2465
.picoclaw/workspace-li-architect             AGENT 1822  SOUL 309  USER 233  MEMORY missing
.picoclaw/workspace-li-frontend-developer    AGENT 1692  SOUL 224  USER 169  MEMORY missing
.picoclaw/workspace-li-data-ai-engineer      AGENT 1741  SOUL 186  USER 155  MEMORY missing
.picoclaw/workspace-li-ux-designer           AGENT 1774  SOUL 215  USER 176  MEMORY missing
```

This snapshot captured the pre-remediation problem. The current verifier now
checks active role memory length, stale setup language, Yaad posture, live repo
awareness, and no-dirty-repo posture.

Runtime-loading facts:

- `pkg/agent/definition.go` prefers structured `AGENT.md`.
- For structured agents it loads paired `SOUL.md` and workspace `USER.md`.
- `pkg/agent/context.go` tracks `AGENT.md`, `SOUL.md`, `USER.md`, and
  `memory/MEMORY.md` for system-prompt cache invalidation.
- Existing tests confirm `USER.md` changes invalidate the structured prompt
  cache.

Config facts:

- `.picoclaw/config.json` has `agents.defaults.max_tool_iterations = 50`.
- `.picoclaw/config.json` has
  `agents.defaults.async_delegation.max_concurrent = 9`.
- `.picoclaw/config.json` has
  `agents.defaults.allow_read_outside_workspace = true`.
- `.picoclaw/config.json` has `agents.defaults.restrict_to_workspace = false`.
- `.picoclaw/config.json` contains `agents.organization` with roots:
  `zehn-main`, `personal`, and `li-ceo`.
- Current org labels and bundle-owner roles use Ali's canonical 10-suite
  taxonomy.

Prompt and cron facts:

- Active operating prompt files exist under
  `.picoclaw/workspace/operating-prompts/`.
- Current cron jobs include CEO hourly, personal hourly, engineering every 30
  minutes, GitHub control-plane hourly, specialist queues, and Zehn operations
  monitor.
- Specialist cron payloads already mention open PR review before
  `HEARTBEAT_OK`.
- The remaining issue is not absence of scheduled checks. It is that role
  personas and operating expectations do not yet force company-like movement.

## Findings

### F-001: Many business roles are written as passive groundwork functions

Evidence:

- `.picoclaw/workspace-li-sales/AGENT.md`
- `.picoclaw/workspace-li-marketing/AGENT.md`
- `.picoclaw/workspace-li-finance/AGENT.md`
- `.picoclaw/workspace-li-cfo/AGENT.md`
- `.picoclaw/workspace-li-chro/AGENT.md`
- `.picoclaw/workspace-li-customer-success/memory/MEMORY.md`

Examples found in role files and memory summaries include repeated
`groundwork`, `prepare`, and planning-only wording.

Impact:

- Agents correctly avoid external side effects, but they also avoid behaving
  like accountable departments.
- Sales, marketing, finance, customer success, HR, legal, and research do not
  have enough concrete development-phase deliverables.
- The company behaves like a readiness monitor, not an operating company.

Correct target:

- Development-phase roles should still produce measurable internal artifacts:
  pipeline hypotheses, pricing models, launch narratives, qualification
  criteria, support playbooks, compliance checklists, retention models, hiring
  plans, and review-ready GitHub/business artifacts when permitted.

### F-002: Role files lack a shared terminal-outcome contract

Evidence:

- Existing prompts and runtime behavior allow `HEARTBEAT_OK`,
  `NO_MATCHING_ISSUES`, and broad status summaries even when open PRs or
  blockers remain.
- The prior audit fixed several prompt-level instances, but role boot files do
  not consistently require work to end in one of a few terminal states.

Impact:

- Agents can inspect the system repeatedly without moving an issue, PR,
  meeting, or decision to a terminal state.
- CEO and CTO can continue to synthesize and re-diagnose without assigning a
  concrete closure owner.

Correct target:

Every role should treat useful work as one of:

- merged;
- reviewed and explicitly approved;
- blocked with named blocker, named owner, and next retry/check date;
- escalated to Ali with a precise approval question;
- delegated with a due/verification expectation;
- deferred with reason and review date;
- replaced by a better issue/plan with the stale item closed or marked stale.

### F-003: COO is not strong enough as company throughput owner

Evidence:

- `.picoclaw/workspace-li-coo/AGENT.md` is better than many roles, but still
  centers planning, operating cadence, and reports.
- Recent behavior shows active queues stuck in review/in-progress while CEO and
  CTO continue diagnostics.

Impact:

- No role is clearly accountable for stale WIP, idle PRs, blocked items, and
  repeated non-terminal heartbeats.
- Work can be "owned" by many agents while no one forces movement.

Correct target:

- `li-coo` should own the operating scoreboard, WIP aging, stuck-item cleanup,
  stale queue reconciliation, handoff quality, and "what changed since last
  cycle" discipline.

### F-004: CEO and CTO were overloaded with diagnostic loops

Evidence:

- Before cleanup, CEO and CTO memory files were huge relative to most roles.
  They have now been replaced with concise active operating doctrine files.
- Earlier log reviews found CTO max-tool-iteration failures and broad
  multi-surface investigations.

Impact:

- CEO/CTO spend too much time rediscovering context and too little time making
  bounded decisions.
- Other roles wait for CEO/CTO synthesis rather than owning their lane.

Correct target:

- CEO should choose outcomes and enforce decision/approval boundaries.
- CTO should own technical direction and remove technical ambiguity, then route
  execution to specialist agents with bounded checks.

### F-005: Specialist roles exist but are not complete as autonomous workers

Evidence:

- Specialist workspaces exist for architecture, backend, frontend, UX,
  integration, data/AI, DevOps, QA, security, docs, product, and engineering.
- Before remediation, four specialist workspaces lacked memory files and
  several specialist `AGENT.md` files were too shallow for their expected
  responsibility.

Impact:

- This has been remediated in active role files. Specialists are now expected
  to pick up appropriate work, verify it, document evidence, and leave repos
  clean.

Correct target:

- Each specialist needs explicit ownership boundaries, issue selection rules,
  verification expectations, repo hygiene, PR/review behavior, escalation
  criteria, Yaad write posture, and "find next useful task if blocked" rules.

### F-006: Bundle-owner naming was confusing

Evidence:

- Before remediation, bundle files used Ignite packaging names that did not
  directly match the 10 canonical suites from Ali's list.
- `LOGICIGNITER_ORGANIZATION_TREE.md` lists the 10 original bundle IDs as the
  active product ownership layer.

Impact:

- This has been remediated in active bundle files, config labels, and current
  portfolio references.

Correct target:

- Each bundle owner now uses the canonical suite name and app coverage. Older
  Ignite package names are not active product identity.

### F-007: Non-engineering roles lack current company objective alignment

Evidence:

- User direction: "Ultimate goal is to maximize profit by portfolio and volume,
  not price."
- Before remediation, many role files did not mention this economic objective
  directly.

Impact:

- This has been remediated in active role files where the objective is relevant.

Correct target:

- LogicIgniter business roles should continue to use the company objective:
  maximize profit through portfolio completeness, volume, operational quality,
  retention, and repeatable delivery, not by simply raising price.

### F-008: The org tree previously framed many departments too passively

Evidence:

- Before remediation, `.picoclaw/workspace/memory/LOGICIGNITER_ORGANIZATION_TREE.md`
  framed business functions too cautiously.

Impact:

- This has been remediated to internal-operating mode: no unapproved external
  commitments, but active internal execution is expected.

Correct target:

- Keep this posture in future edits.

### F-009: Missing-memory roles previously had weaker local boot context

Evidence:

- Before remediation, four active specialists had no local `memory/MEMORY.md`.

Impact:

- This has been remediated; all active workspaces now have local memory files
  and Yaad posture.

Correct target:

- Keep Yaad as canonical durable memory, with local memory serving only
  boot/runtime context.

### F-010: Current role files do not consistently define "real company behavior"

Evidence:

- Roles identify their function but often do not define:
  - standing operating cadence;
  - default outputs;
  - what to inspect first;
  - how to communicate with CEO/COO/peer agents;
  - how to handle blocked work;
  - what a good report looks like;
  - when to create issues or PRs;
  - when to write Yaad memory.

Impact:

- The runtime has agents, delegation, meetings, GitHub, Yaad, and cron, but the
  roles do not yet have enough operating doctrine to behave like departments.

Correct target:

- Rewrite each role as a handcrafted operator, not a title with generic safety
  language.

## Root Cause Hypothesis

The dominant issue is not a missing Go feature. The dominant issue is operating
model design:

1. The runtime can delegate, hold meeting v1, use GitHub, use Yaad, run tools,
   and run scheduled checks.
2. The prompts and roles over-emphasize inspection, readiness, safety, and
   broad diagnostics.
3. The roles under-specify accountable outputs, terminal outcomes, lane
   ownership, and cross-functional handoff rules.
4. Therefore agents keep "checking" and "summarizing" instead of operating like
   a company that converts work into completed artifacts.

This should be fixed first in role files, operating prompts, memory doctrine,
and verification. Code changes should be avoided unless a concrete runtime
limitation remains after those changes.

## Recommended Remediation Direction

- Handcraft role files in small batches, starting with company operating layer:
  `li-ceo`, `li-coo`, `li-cto`, `li-cpo`, `li-cro`, `li-cmo`, `li-cfo`,
  `li-ciso`, `li-legal`, `li-cdao`, `li-cco`, `li-chro`.
- Then rewrite specialist execution roles.
- Then rewrite bundle owners with explicit suite/app coverage and taxonomy.
- Then revise cron prompts to demand terminal outcomes and changed-state
  reporting.
- Then verify via a role-persona audit script and a controlled live drill.
