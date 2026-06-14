# Zehn Setup Planning Ledger

Historical setup log. Do not treat older runtime snapshots, PIDs, agent counts,
draft-PR preferences, mention-policy notes, or setup-phase delegation examples
in this file as current truth.

For current operating state, read:

`/Users/aliai/.picoclaw-zehn/workspace/memory/ZEHN_CURRENT_STATE.md`

For current GitHub execution policy, read:

`/Users/aliai/.picoclaw-zehn/workspace/memory/LOGICIGNITER_GITHUB_CONTROL_PLANE.md`

For current approval policy, read:

`/Users/aliai/.picoclaw-zehn/workspace/memory/LOGICIGNITER_APPROVAL_ESCALATION_MATRIX.md`

This file records the design discussion for configuring Zehn/PicoClaw as Ali's
always-on personal and company assistant.

## Baseline

- Machine: Ali's local MacBook Pro.
- Runtime posture: local-first.
- Launcher: Web launcher on loopback.
- CLI command: `zehn` alias/symlink to `picoclaw`.
- Provider: OpenAI/ChatGPT OAuth-backed provider for Zehn runtime. Codex CLI is
  not the intended Zehn LLM provider.
- Secrets: `.security.yml`, not `config.json`.
- External channels: disabled until individually designed and allowlisted.
- Yaad: private MCP integration is now the preferred durable memory path.
- Yaad scope correction: LogicIgniter-wide durable memory must use
  `scope_type: organization` with `external_key: logicigniter`. `company` is
  not a valid Yaad scope type. Agents should call `scope_type_list` when scope
  validity is uncertain instead of inventing new scope types.
- Powerful tools: exec remote disabled, command cron disabled, MCP and external
  channel exposure enabled only by explicit staged rollout.

## Discussion Protocol

- Ali will briefly explain desired needs and scenarios.
- Codex will ask MCQs to drive decisions.
- Each answer and resulting decision should be recorded here.
- Temporary ideas remain in this ledger until promoted into `AGENT.md`,
  `SOUL.md`, `USER.md`, `memory/MEMORY.md`, config, skills, or MCP setup.

## Open Design Areas

- Assistant identity and personality.
- Personal vs company mode boundaries.
- Privacy and escalation policy.
- Allowed tools and confirmation rules.
- Memory model and Yaad integration.
- Channel rollout order and allowlists.
- Automation, reminders, cron, and heartbeat.
- Company workflows and integrations.
- Backup, observability, upgrade, and recovery posture.

## Current Status Update - 2026-05-06

This planning ledger is still the canonical local record of the original MCQ
and setup design discussion. It has not been deleted. Newer focused docs were
created during implementation, but they do not replace this ledger.

Implementation completed in the Zehn fork:

- Native private target-agent delegation via `delegate_to_agent`.
- `delegation_status` and `delegation_inbox` with caller-scoped visibility.
- Local redacted delegation records.
- Bounded async delegation executor.
- Terminal-state delegation memory persistence through Yaad MCP when available.
- Idempotent delegation memory behavior.
- Chaired meeting v1 via `start_agent_meeting`.
- Local redacted meeting records.
- Required participant failure policy for meeting v1.
- Redacted GitHub artifact publishing with runtime-owned async publisher.
- Discord visibility summaries without using Discord as the internal
  delegation bus.
- Deterministic end-to-end delegation/meeting tests.
- Focused normal and race verification passed.

Implementation automation status:

- Feature tasks `001` through `023` are green.
- The implementation is acceptable for private staged Zehn rollout.
- Upstream/public publishing is intentionally parked until later because the
  local branch contains private skills and supervision history. A clean
  upstream branch must be rebuilt separately when publishing becomes a priority.

Remaining Zehn setup work:

1. Rebuild launcher and CLI from the current source.
2. Restart through the launcher UI so gateway lifecycle remains standard.
3. Confirm the active runtime uses the intended `PICOCLAW_HOME` and config.
4. Confirm delegation/meeting tools are enabled only for intended agents.
5. Keep GitHub artifacts disabled during first local live tests.
6. Run one local CLI/Web sync delegation.
7. Run one local CLI/Web async delegation and inspect status/inbox.
8. Run one local chaired meeting with two required participants.
9. Enable Yaad MCP and verify one terminal delegation memory write.
10. Run one narrow Discord command from Ali in an allowlisted channel.
11. Enable GitHub artifacts last with one low-risk approval/follow-up test.
12. Later: add business-repo meeting document persistence, operating cadence
    cron/heartbeat, broader GitHub Project automation, meeting v2 debate, and
    clean upstream-publishable branch splitting.

## Product Model Update - 2026-05-08

LogicIgniter's current product strategy is now captured in
`/Users/aliai/logicigniter/supervision/IN_HOUSE_THIRD_PARTY_PRIORITIZATION.md`.
This replaces the earlier launch-dependency-only framing.

Zehn should retain the original 51-app launch-readiness discipline and use the
canonical 10-suite portfolio names Ali supplied. The current product model is:

1. 51 services as the atomic product/API layer.
2. First-party LogicIgniter platform products created from third-party
   dependency categories where practical.
3. 10 canonical LogicIgniter suites as the first marketable packaging layer.

The current suite names are SaaS Growth & Retention, E-commerce Operations,
Finance & Revenue Intelligence, Legal & Compliance, Developer & DevOps, Content
& Marketing, HR & Workforce, Professional Services, Real Estate & Property, and
Education. Zehn should treat provider-backed delivery as an implementation
stage, not as the product identity.

## Decisions

- Long-term memory will be handled by an external memory system called Yaad,
  hosted on another machine. Details, skill instructions, and integration
  specifics are pending from Ali.
- Until Yaad is specified and integrated, PicoClaw local workspace memory should
  be treated as bootstrap/runtime context, not the final durable long-term
  memory system.
- Yaad will be the canonical source of truth for Zehn's durable long-term
  memory. Local PicoClaw workspace memory should function as boot context,
  runtime notes, cache, and fallback only.
- Before Yaad is connected, new durable memories should be written as local
  candidate memories for later review/import into Yaad, not treated as canonical
  long-term memory.
- Primary intended use: Zehn should serve two major purposes:
  1. Ali's personal assistant.
  2. A complete LogicIgniter management and operating system.
- LogicIgniter is Ali's SaaS-based services company.
- Ali's external consulting job exists but is intentionally out of scope for
  Zehn for now.
- LogicIgniter mode should eventually cover an organization-wide role/agent
  model, including CEO, CTO, product owner, development team, sales, marketing,
  and other major company functions.
- Architecture decision: Zehn should be the top-level coordinator with separate
  internal assistants/agents underneath:
  1. Personal Assistant for Ali's personal projects and personal operating needs.
  2. LogicIgniter operating system for company management and execution.
- LogicIgniter should then contain its own role/agent structure for company
  functions.
- Context routing decision: when a request could belong to both Personal and
  LogicIgniter contexts, Zehn should infer the likely context from content, but
  ask for confirmation when there is risk around memory boundaries, company
  data, external actions, permissions, tools, or irreversible decisions.
- LogicIgniter target-state: Zehn should support the architecture of a large
  scale US-based software organization.
- LogicIgniter product scope: 51 SaaS applications organized into 10 bundles.
- Current company reality: most pieces exist but are currently used in bits and
  pieces rather than as a unified operating system.
- Current operating phase: development phase. Individual services need makeover
  work, so the Engineering department will be the most active near-term area.
- Other departments should begin groundwork now, but full-scale operation should
  activate when LogicIgniter goes live.
- Department rollout decision: define the full LogicIgniter target org chart now,
  but tag each department/role with an activation state such as active,
  groundwork, or future. Engineering is active now; other departments should
  mostly perform groundwork until launch/live operations.
- Department/role activation states:
  - `active`: participates in regular execution now.
  - `groundwork`: prepares assets, plans, systems, and research but does not
    drive daily operations.
  - `advisory`: reviews plans and flags risks/requirements without owning daily
    execution.
  - `future`: defined for target architecture but mostly inactive until a later
    phase.
- Current active departments: Engineering, Product, and Operations.
- Current Engineering surface should include relevant sub-functions such as QA
  and DevOps/SRE as needed for service makeover work.
- Other departments should be assigned groundwork or advisory states unless a
  specific near-term need activates them.
- LogicIgniter role model: executive-led hierarchy.
- Zehn coordinates the LogicIgniter company system. Under it, a CEO agent owns
  company-level coordination. C-suite roles report into/coordinate through the
  CEO, and each C-suite role owns departments, teams, or project pods.
- LogicIgniter decision authority: CEO agent has final internal decision
  authority within the AI company system for routine and reversible decisions.
  The CEO agent should escalate to Ali for high-impact, risky, irreversible,
  financial, legal, brand-sensitive, personnel, or external-facing decisions.
- CEO agent operating style: strategic operator. It should set priorities,
  resolve tradeoffs, drive weekly execution, maintain alignment across roles,
  and escalate only when authority/risk boundaries require Ali's decision.
- Target LogicIgniter C-suite roles:
  - CEO
  - CTO
  - CPO
  - COO
  - CMO
  - CRO
  - CFO
  - CLO / Legal
  - CHRO
  - CISO
  - Chief Customer Officer
  - Chief Data/AI Officer
- These roles are target architecture roles; each still needs an activation
  state assignment for the current development phase.
- Current C-suite activation:
  - `active`: CEO, CTO, CPO, COO.
  - `advisory`: CISO.
  - `groundwork` or `future`: CMO, CRO, CFO, CLO/Legal, CHRO, Chief Customer
    Officer, Chief Data/AI Officer, unless a specific near-term need changes
    their state.
- Current Engineering organization: hybrid model.
  - Shared functions: platform/architecture, DevOps/SRE, security, QA standards,
    documentation standards, and engineering process.
  - Temporary execution pods: service or bundle makeover pods focused on specific
    SaaS apps or bundles.
- Engineering execution authority: Engineering agents should be able to perform
  autonomous execution in trusted LogicIgniter repositories, including editing,
  testing, and preparing commits, while escalating risky changes.
- Autonomous engineering requires explicit trusted-repo boundaries, testing
  requirements, rollback/commit discipline, and escalation rules before broad use.
- Trusted engineering repo boundary: repositories under the LogicIgniter GitHub
  organization are trusted for autonomous engineering by default unless
  explicitly excluded.
- Trusted-repo design still needs an exclusion list for sensitive, archived,
  third-party, client, or out-of-scope repositories.
- Autonomous engineering Git rule: branch + PR only. Agents should create
  branches and pull requests for LogicIgniter work and should not push directly
  to main/protected branches.
- Engineering escalation rule: escalate before implementation when either risk
  category or blast radius warrants it.
  - Sensitive categories include auth, payments, data deletion, security, infra,
    migrations, production config, legal/compliance, public APIs, customer data,
    and external integrations.
  - Blast-radius triggers include shared libraries, multiple services,
    deployment/release paths, data models, user-facing behavior, or anything
    difficult to roll back.
- Product management model: portfolio hierarchy.
  - Company portfolio level.
  - 10 canonical LogicIgniter suites.
  - 51 individual services/apps as the atomic product/API layer.
- Product should track rollups at portfolio and bundle level while preserving
  app-level ownership, backlog, quality state, makeover needs, and launch
  readiness.
- Product inventory source of truth during setup: structured local registry
  first. This should capture the 10 canonical LogicIgniter suites, 51
  services/apps, product surfaces, and provider-backed status in a versioned local
  file/database in the Zehn workspace, then later sync/promote into Yaad and/or
  company systems when available.
- Initial SaaS app registry fields should support launch readiness:
  - name
  - bundle
  - repo(s)
  - status
  - makeover priority
  - owner role
  - current issues
  - next action
  - target customer
  - pricing status
  - docs status
  - QA status
  - deployment status
  - security status
- Operations role during development phase: both cadence/tracking and company
  systems setup.
  - Cadence/tracking includes weekly planning, status reports, decision logs,
    risks, blockers, and follow-ups.
  - Company systems setup includes docs structure, registries, templates, SOPs,
    dashboards, repo/project hygiene, and operating routines.
- Current LogicIgniter operating cadence: startup operating cadence.
  - Weekly planning.
  - Daily standup/check-in.
  - Midweek risk review.
  - Friday status report.
  - Monthly roadmap review.
- Cadence interaction model: proactive drafts. Zehn should prepare draft plans,
  reports, check-ins, and reviews on schedule, then ask Ali to review, adjust,
  or approve them.
- First external channel target: Discord.
- Ali has already added a Discord bot token.
- Discord should use multiple domain-specific channels, including Personal and
  LogicIgniter role/department channels such as CEO and CTO.
- Zehn needs a planned Discord channel list. Ali can provide required channel
  IDs after channels are created or selected.
- Discord channel rollout decision: hybrid. Create top-level Personal,
  executive, department, approvals, and system channels now, while keeping
  sub-team channels lean and adding more only when needed.
- Discord proactive posting policy: proactive posts only in key operating
  channels at first. Other channels should be mention/respond-only or receive
  targeted output when explicitly routed.
- Initial proactive Discord channels should include personal inbox/reminders,
  LogicIgniter CEO, Engineering, Operations cadence, approvals, and bot health.
- Discord access policy: use both channel allowlists and user/role allowlists.
  Start with only Ali's Discord user ID allowed. Add team members or roles only
  intentionally later.
- Discord memory/session model: hybrid. LogicIgniter channels should share
  company-level context while retaining per-channel working memory for
  role-specific details. Personal channels remain separate from LogicIgniter.
- Discord cross-posting policy: summary-only. Detailed work should stay in the
  source channel. Zehn may post concise summaries, decisions, risks, or action
  items to CEO, Operations, approvals, or decision-log channels when relevant.
- Discord approval routing: approvals should be requested in the source channel
  where the context exists, and LogicIgniter-impacting approvals should also be
  copied or summarized into the central `li-approvals` channel.
- Approval threshold: Zehn must get explicit approval before any external side
  effect or high-risk internal action.
  - External side effects include sending external messages, creating PRs,
    opening issues, changing configs, scheduling jobs, publishing content,
    contacting people, or changing third-party systems.
  - High-risk internal actions include financial, legal, production, security,
    public-facing, customer-data, irreversible, or broad-blast-radius changes.
- Discord role-output format: named role sections inside a single Zehn response
  by default, such as CEO, CTO, CPO, COO, CISO/Risk, Decision, and Next Actions.
  Avoid noisy separate messages per role unless explicitly requested.
- Internal role disagreement policy: show role debate/disagreement regularly,
  not only when material. Zehn should make competing views visible and then show
  how the CEO/decision layer resolves or escalates the decision.
- Role debate detail policy: short debate by default, with full council debate
  available on demand. Default responses should keep Discord usable while still
  exposing role disagreement.
- Full role debate trigger phrase: `full council`.
- LogicIgniter response template policy: adaptive.
  - Use a decision memo for strategic/tradeoff decisions: Context, Role Debate,
    Decision, Risks, Next Actions, Approval Needed.
  - Use an execution brief for implementation/operating work: Goal, Plan, Owners,
    Tasks, Timeline, Blockers, Approval Needed.
- Personal Assistant behavior: hybrid. Quiet and lightweight by default for
  reminders, notes, planning, and personal project tracking, but able to invoke
  specialist roles for personal projects or complex decisions when useful.
- Personal project model: mini portfolio. Each personal project should track
  goal, repo, status, priority, next action, risks, and notes, without using the
  full LogicIgniter app readiness model.
- Consulting job boundary: hard exclude. Zehn should not store, organize, route,
  or integrate consulting-job details unless Ali explicitly overrides this later.
  Consulting work should not mix into Personal or LogicIgniter systems.
- Consulting accidental-mention behavior: ask before proceeding. If content
  appears consulting-related, Zehn should ask whether to ignore it, answer once
  without storing memory, or explicitly override the exclusion.
- Memory capture posture: broad capture. Zehn should save most useful context
  unless it is sensitive, temporary, noisy, explicitly excluded, or belongs to
  the consulting-job boundary. Before Yaad is connected, these should go into
  local candidate memory for review/import.
- Candidate memory confirmation policy: auto-capture low-risk useful facts as
  candidates, but ask before saving sensitive, personal, company-critical,
  boundary-crossing, ambiguous, or potentially harmful memories.
- Candidate memory review cadence: daily review digest. Once Discord is enabled,
  Zehn should post candidate memories to `zehn-memory-review` for Ali to approve,
  reject, correct, or classify before Yaad import.
- Candidate memory classification before Yaad import:
  - Scope: personal, logicigniter, system, excluded, or future additional scopes.
  - Sensitivity: public, internal, confidential, secret.
  - Lifespan: permanent, long-term, temporary, or expires-on-date.
- Research posture: proactive research agent. Zehn should eventually be able to
  schedule and run research digests for markets, competitors, SaaS trends,
  company groundwork, and related strategic inputs. This needs source-quality,
  citation, cadence, and channel-routing rules before broad activation.
- Initial proactive research focus: balanced digest covering competitive
  intelligence, product/market strategy, and engineering/technology watch, with
  deeper focused research available when a department requests it.
- Proactive research cadence: daily light scan/signals plus weekly synthesis.
- Proactive research Discord routing: use a central `li-research-digest` channel
  as the archive, and route short summaries to relevant channels such as CEO,
  Product, Engineering, CISO, Marketing, or Sales when applicable.
- Research source standard: tiered confidence. Zehn may include useful signals
  from imperfect sources, but must label source quality/confidence and prefer
  official docs, primary sources, credible reports, and public company pages
  when available.
- GitHub management posture for LogicIgniter: create/update with approval.
  Zehn may draft and prepare issues, PRs, labels, milestones, project-board
  updates, and repo-management changes, but must get approval before applying
  them because these are external side effects.
- Engineering work item source of truth: hybrid.
  - GitHub Issues/Projects are canonical for repo-specific engineering tasks.
  - Zehn registry is canonical for portfolio/company planning, bundle/app
    readiness, and cross-repo initiatives.
- Calendar/email target posture: design for eventual calendar and email
  integration, but do not enable until Discord and GitHub workflows are stable.
  Calendar/email access should be treated as high-risk external side effects and
  require explicit approval policies.
- Secrets posture: Yaad/private memory may store secret metadata and references,
  but not raw secrets. Actual secrets should remain in `.security.yml` for the
  local setup or move to a proper secret manager later.
- Production/customer data policy: no direct access. Zehn should only work with
  summaries, synthetic examples, or sanitized exports of production/customer
  data unless Ali explicitly revisits and changes this policy later.
- Customer/lead/user contact policy: Zehn may contact customers, leads, or users
  only after explicit approval of the exact message and exact recipient/audience.
  Otherwise Zehn should draft only.
- Financial decision policy: Zehn may analyze, recommend, prepare budgets,
  purchase recommendations, invoices, or financial plans, but any financial
  execution requires explicit approval.
- Legal/compliance policy: Zehn may maintain compliance checklists, draft
  policies/contracts/checklists, track obligations, and flag risks, but must not
  represent outputs as legal advice. Formal, external, or legally significant
  use requires human/legal review and approval.
- Documentation policy: Zehn should own and maintain internal company knowledge
  structures, SOPs, templates, decision logs, and operating manuals. External or
  public documentation can be drafted by Zehn but requires approval before
  publishing or external use.
- Company docs storage model: hybrid.
  - Local Zehn workspace for runtime notes, drafts, and working materials.
  - LogicIgniter GitHub docs/knowledge-base repo for approved company docs.
  - Yaad for durable memory and long-term recall.
- Reporting style: layered reports. Zehn should lead with a short executive
  summary focused on decisions, risks, and next actions, with department-level
  detail available on request.
- Decision tracking: important decisions should be logged with date, context,
  owner, rationale, and revisit/validation date.
- Risk tracking: department-specific risk registers with an executive rollup.
  Risks should include severity, likelihood, impact, owner, mitigation, status,
  and review date.
- Security posture in development phase: continuous security advisory function.
  CISO should maintain a security risk register, watch dependencies/advisories,
  review architecture, and gate PRs or plans for sensitive changes such as auth,
  data, infra, dependencies, releases, and security-sensitive code.
- Launch requirement: all 51 SaaS applications must launch together.
- App makeover prioritization should use a readiness/risk score, but only to
  sequence makeover work, reduce risk, and manage dependencies toward a
  simultaneous launch of all 51 apps. The score must not be used to choose a
  subset of apps for launch.
- Simultaneous launch readiness model: portfolio launch gate.
  - Per-app gates: every app must meet baseline quality, security, docs, QA,
    deployment, and product-readiness requirements.
  - Per-bundle gates: every bundle must meet product story, positioning, pricing,
    support, documentation, and readiness requirements.
  - Company-level go/no-go gate: executive review of portfolio readiness, risks,
    launch operations, support, legal/compliance, and approvals.
- Launch blocking rule: any failed app readiness gate blocks the entire 51-app
  launch. No partial launch, hidden launch, or limited launch exception is
  assumed.
- Launch readiness strictness: professional SaaS launch standard. Gates should
  include QA pass, documentation, onboarding, pricing, monitoring, support
  readiness, deployment readiness, and security review.
- Launch readiness tracking: both checklist gates and numeric progress scoring.
  Checklist gates determine pass/fail blocking status; scores visualize progress
  and help prioritize work.
- Bundle operating model: each of the 10 canonical LogicIgniter suites should have a mini-GM /
  bundle owner role responsible for readiness, positioning, dependencies, risks,
  and coordination across apps in that bundle.
- Bundle owner reporting: dual report to CPO and COO. CPO owns product direction;
  COO owns operational readiness. CTO participates for engineering dependencies
  and implementation risk.
- App-level ownership: every one of the 51 services/apps should have a named app
  owner role accountable for readiness, risks, next actions, and coordination
  with the bundle owner and engineering pod.
- App owner implementation model: hybrid. Bundle owners should be persistent
  roles/agents. App owners should be registry-driven virtual roles generated
  from app registry data when needed.
- Multi-role work model: workflow templates plus dynamic role selection.
  Repeatable work such as bundle launch readiness, app makeover, weekly planning,
  risk review, and research digest should use predefined workflow templates.
  Zehn may dynamically add roles when a request requires expertise not included
  in the default workflow.
- First formal workflow template to design: App makeover workflow.
- Phase 1 real PicoClaw agents:
  - `zehn-main` / coordinator
  - `personal`
  - `li-ceo`
  - `li-cto`
  - `li-cpo`
  - `li-coo`
  - `li-ciso`
- Other target C-suite roles, bundle owners, virtual app owners, and temporary pods can
  remain virtual, registry-driven, or future until activated.
- Discord-to-agent routing direction: CEO as LogicIgniter company default. Ali
  prefers to communicate primarily with the CEO agent, which should delegate to
  proper agents/roles as needed. Role-specific channels may still exist, but the
  main operating interface should be CEO-led.
- CEO delegation policy: delegate when needed. CEO should answer simple company
  topics directly, and consult/delegate to CTO, CPO, COO, CISO, bundle owners,
  virtual app owners, or workflow-specific roles for decisions, plans, risks, and
  cross-functional work.
- Delegation architecture should use Yaad and GitHub appropriately:
  - Yaad for canonical organizational memory, role context, decisions, durable
    facts, operating history, and cross-agent continuity.
  - GitHub Issues/Projects/Wiki or docs repos for concrete engineering/product
    execution artifacts, approved tasks, PRs, project tracking, and documented
    plans.
  - Zehn workflow templates for deciding when/how the CEO delegates, gathers
    perspectives, synthesizes disagreement, and asks Ali for approval.
- Delegation/execution artifact source: GitHub Projects should be the canonical
  system for LogicIgniter delegation and execution at scale. Issues and PRs
  should be linked underneath GitHub Projects. Zehn must get approval before
  creating/updating these external artifacts.
- GitHub Projects / repo structure:
  - LogicIgniter has a `business` repo for high-level company-wide artifacts.
    A GitHub Project may be created in or attached to this repo for company-wide
    portfolio, executive, and operating artifacts.
  - Each individual service/application repo should have its own project or
    project tracking to manage issues and work.
  - Scope includes 51 SaaS app repos plus shared/platform repos such as infra,
    Go packages, web, portal, proto, BFF, and similar shared components.
  - Bundle-level and executive rollups should aggregate from these repo/project
    sources rather than replacing them.
- Bundle-level rollups should live in both:
  - `business` repo for durable bundle docs, decisions, strategy, readiness
    narratives, and historical records.
  - GitHub Project views for live task status, execution tracking, issues, PRs,
    blockers, and readiness progress.
- Top-level company command center: `business` repo. Discord is the
  conversational interface, GitHub Projects are live execution/status views, and
  Yaad is canonical long-term memory.
- `business` repo update policy: Zehn may autonomously prepare draft PRs for
  internal company docs and operating artifacts. Approval is required before
  merge, public use, or external use.
- `business` repo confidentiality: private internal only. Treat all contents as
  confidential by default.
- GitHub Project mutation policy: Zehn may autonomously update internal
  LogicIgniter GitHub Projects. Approval is still required for public/external
  effects, high-risk actions, or changes that create external commitments.
- GitHub Issue creation policy: Zehn may create/update related issues in private
  LogicIgniter repos after Ali approves the underlying plan. It should not need
  separate approval per issue once the plan is approved.
- GitHub PR creation policy: after an approved plan and implementation in a
  trusted LogicIgniter repo, Zehn may open draft PRs autonomously. PRs remain
  reviewable and should not be merged without approval.
- GitHub PR merge policy: Zehn may merge PRs only after explicit approval from
  Ali/human reviewer.
- CI/deployment policy: Zehn may trigger/check CI and may trigger deployments
  only after explicit approval. Production deployments are high-risk and always
  approval-gated.
- Shell/exec policy: Zehn may run tests, builds, and development commands in
  trusted LogicIgniter workspaces/repos for engineering work. Remote
  Discord-triggered exec still needs guardrails, allowlists, and escalation for
  risky/destructive commands.
- Development-phase codebase control: Ali wants to give Zehn full control over
  trusted LogicIgniter codebases for now. This should be revised/tightened before
  actual launch.
- Even with full codebase control, production/customer data access remains
  prohibited and production deployments remain approval-gated.
- Full codebase control includes dependency upgrades, major refactors, broad
  code changes, and service makeover work in trusted LogicIgniter repositories
  during the development phase.
- Major codebase work still requires a written plan first, even when explicit
  approval is not required under development-phase full-control policy.
- Definition of major codebase work: size-based or impact-based. Either can make
  work major.
  - Size-based triggers include multi-file, multi-package, multi-service, or
    broad codebase changes.
  - Impact-based triggers include architecture, public APIs, data models, auth,
    deployment, dependencies, shared libraries, major UI/UX, or anything with
    broad blast radius.

## Raw Notes

- Ali: "we will be using a memory system hosted on another machine for long term
  memory for zehn. it's called 'yaad' I'll share skill and details later."
- MCQ decision: Yaad role = canonical memory store.
- MCQ decision: pre-Yaad durable memory handling = conservative local queue of
  candidate memories.
- Ali wants to explain the intended Zehn usage model before continuing with more
  MCQs.
- Ali is a software engineer working on personal projects.
- Ali has a company called LogicIgniter, a SaaS-based services company.
- Ali also has a consulting job for another company, but that should not be
  added to Zehn for now.
- Ali's real focus for Zehn is personal assistance and a complete LogicIgniter
  management system.
- Desired LogicIgniter system should have agents/roles able to run an entire
  organization: CEO, CTO, product owner, development team, sales, marketing, and
  every major operational aspect in an automated way.
- MCQ decision: relationship between personal and LogicIgniter systems = separate
  internal agents coordinated by Zehn.
- MCQ decision: ambiguous personal/company requests = infer context, but confirm
  when risky.
- Ali clarified that LogicIgniter needs the full architecture of a large-scale
  US-based software company offering 51 SaaS applications in 10 bundles.
- Ali clarified the current problem: the company has many pieces in place, but
  they are currently used as bits and pieces rather than one integrated operating
  system.
- Ali clarified the current phase: active development and service makeover work,
  with heavy Engineering usage now. Other departments should start groundwork,
  and full-scale architecture should be used once live.
- MCQ decision: LogicIgniter department rollout = target org chart now with
  phased activation states.
- MCQ decision: department/role activation states = active, groundwork,
  advisory, future.
- MCQ decision: current active departments = Engineering, Product, Operations.
- MCQ decision: LogicIgniter top-level role model = executive-led hierarchy.
- MCQ decision: internal LogicIgniter final decision authority = CEO agent,
  bounded by escalation rules to Ali.
- MCQ decision: CEO agent primary operating style = strategic operator.
- MCQ decision: target C-suite = expanded enterprise model.
- MCQ decision: current active C-suite = CEO, CTO, CPO, COO; CISO advisory.
- MCQ decision: current Engineering structure = hybrid shared functions plus
  temporary service/bundle makeover pods.
- MCQ decision: Engineering agent authority = autonomous execution in trusted
  LogicIgniter repositories, with risk escalation.
- MCQ decision: trusted autonomous engineering scope = LogicIgniter GitHub org
  allowlist, with exclusions.
- MCQ decision: autonomous engineering Git behavior = branch + PR only.
- MCQ decision: engineering risk escalation = both category-based and
  blast-radius-based.
- MCQ decision: Product structure = portfolio hierarchy from company portfolio
  to bundles to individual apps.
- MCQ decision: initial source of truth for bundles/apps inventory = structured
  local registry first, later promoted/synced to Yaad.
- MCQ decision: initial SaaS app registry schema = launch-readiness fields.
- MCQ decision: Operations development-phase responsibility = both operating
  cadence/tracking and company systems setup.
- MCQ decision: current operating cadence = startup cadence.
- MCQ decision: cadence interaction = proactive drafts.
- Channel direction: Discord first, using separate domain-specific channels for
  Personal and LogicIgniter operating streams.
- MCQ decision: Discord channel count = hybrid rollout.
- MCQ decision: Discord proactive behavior = proactive only in key channels.
- MCQ decision: Discord access restriction = channel allowlist plus user/role
  allowlist, Ali-only initially.
- MCQ decision: Discord memory/session scoping = shared LogicIgniter context plus
  per-channel working memory; separate Personal memory.
- MCQ decision: Discord cross-posting = summaries only, not full duplicate
  routing.
- MCQ decision: Discord approvals = source-channel request plus central
  `li-approvals` summary/copy for LogicIgniter-impacting approvals.
- MCQ decision: approval threshold = both external side effects and high-risk
  internal actions require explicit approval.
- MCQ decision: Discord role-output format = named role sections in one response.
- MCQ decision: internal role disagreement visibility = regularly show debate.
- MCQ decision: role debate detail = short by default, full council debate on
  demand.
- MCQ decision: full debate trigger phrase = `full council`.
- MCQ decision: LogicIgniter response template = adaptive decision memo or
  execution brief.
- MCQ decision: Personal Assistant behavior = quiet by default with optional
  specialist roles.
- MCQ decision: personal projects = mini portfolio model.
- MCQ decision: consulting job context = hard excluded from Zehn for now.
- MCQ decision: accidental consulting-job content = ask before proceeding.
- MCQ decision: memory capture scope = broad capture with exclusions/sensitivity
  safeguards and pre-Yaad candidate queue.
- MCQ decision: candidate memory capture = auto-capture low-risk, ask for
  sensitive or boundary-crossing memories.
- MCQ decision: candidate memory review = daily digest in `zehn-memory-review`.
- MCQ decision: candidate memory classification = scope + sensitivity +
  lifespan.
- MCQ decision: web/search posture for company groundwork = proactive research
  agent.
- MCQ decision: proactive research focus = balanced digest.
- MCQ decision: proactive research cadence = daily light scan plus weekly
  synthesis.
- MCQ decision: proactive research posting = central `li-research-digest` plus
  relevant routed summaries.
- MCQ decision: research source standard = tiered confidence.
- MCQ decision: GitHub issue/project management = create/update with approval.
- MCQ decision: engineering work item source of truth = GitHub for repo work,
  Zehn registry for portfolio/company and cross-repo planning.
- MCQ decision: calendar/email integration = target capability later, deferred
  until Discord and GitHub are stable.
- MCQ decision: secrets/memory posture = memory stores references/metadata only,
  raw secrets stay in `.security.yml` or future secret manager.
- MCQ decision: production/customer data access = no direct access; sanitized
  summaries/exports only.
- MCQ decision: customer/lead/user contact = allowed only with exact-message and
  exact-recipient/audience approval.
- MCQ decision: financial actions = prepare/recommend; execute only after
  approval.
- MCQ decision: legal/compliance = full compliance assistant, never legal advice,
  review-gated for formal/external use.
- MCQ decision: documentation = Zehn owns internal docs; external docs require
  approval before publishing/use.
- MCQ decision: company docs storage = local Zehn workspace for drafts/runtime,
  GitHub docs repo for approved docs, Yaad for durable memory.
- MCQ decision: reporting style = layered executive summary first, expandable
  department detail.
- MCQ decision: decision tracking = decision log with revisit dates.
- MCQ decision: risk tracking = department-specific registers plus executive
  rollup.
- MCQ decision: current security posture = continuous CISO advisory function.
- MCQ decision: app makeover prioritization = readiness/risk scoring for work
  sequencing, with all 51 apps launching together.
- MCQ decision: simultaneous launch readiness = portfolio launch gate with
  per-app, per-bundle, and company-level go/no-go gates.
- MCQ decision: failed app readiness = blocks entire launch.
- MCQ decision: launch readiness strictness = professional SaaS launch standard.
- MCQ decision: launch readiness tracking = both checklist gates and progress
  score.
- MCQ decision: SaaS bundle ownership = mini-GM/bundle owner per bundle.
- MCQ decision: bundle owner reporting = dual report to CPO and COO, with CTO
  involvement for engineering dependencies.
- MCQ decision: app-level ownership = one accountable app context per app.
- Current implementation decision: persistent bundle owners plus
  registry-driven virtual app owners; do not use persistent `li-app-*` agents.
- MCQ decision: multi-role work = workflow templates plus dynamic role additions.
- Ali asked whether this can be set up correctly in Zehn; answer direction:
  yes, if implemented in layers rather than assuming every company-OS concept is
  native to PicoClaw.
- MCQ decision: first workflow template = App makeover workflow.
- MCQ decision: phase 1 real agents = active leadership agents: coordinator,
  personal, CEO, CTO, CPO, COO, CISO.
- MCQ decision: Discord/LogicIgniter default routing = CEO as company default,
  with CEO delegation to relevant agents/roles.
- MCQ decision: CEO delegation autonomy = delegate when needed.
- Ali suggested using Yaad memory and GitHub Projects/Issues/Wiki for delegation;
  design direction is to use Yaad for memory/context, GitHub for execution
  artifacts, and Zehn workflows for delegation logic.
- MCQ decision: delegation artifacts = GitHub Projects canonical, with Issues/PRs
  linked underneath and approval before mutation.
- Structure correction from Ali: GitHub work tracking should center on the
  `business` repo for high-level company artifacts and per-service/shared-repo
  projects for work tracking, with rollups above them.
- MCQ decision: bundle-level rollups = both `business` repo durable docs and
  GitHub Project live views.
- MCQ decision: top-level company command center = `business` repo.
- MCQ decision: `business` repo updates = autonomous draft PRs for internal docs,
  approval before merge/public/external use.
- MCQ decision: `business` repo visibility/content posture = private internal
  only, confidential by default.
- MCQ decision: GitHub Project mutations = autonomous for internal projects;
  approval for public/external/high-risk effects.
- MCQ decision: GitHub Issues = create/update after approved plan, no per-issue
  approval required.
- MCQ decision: GitHub PR creation = autonomous draft PRs after approved plan and
  implementation in trusted repos.
- MCQ decision: GitHub PR merging = only after explicit approval.
- MCQ decision: CI/deployment actions = deployments only after explicit
  approval; production is high-risk.
- MCQ decision: shell/exec = allowed in trusted workspaces for engineering work,
  with remote-trigger guardrails.
- Policy revision: replace destructive-only command approval framing with
  development-phase full codebase control in trusted repos; revisit/tighten
  before launch.
- MCQ decision: full codebase control includes dependency upgrades and major
  refactors during development.
- MCQ decision: major codebase work = written plan first.
- MCQ decision: major codebase work definition = both size and impact triggers.
- Yaad skill status: `yaad-memory` exists locally at
  `/Users/aliai/.codex/skills/yaad-memory/SKILL.md`, with valid skill
  frontmatter and reference files under `references/`.
- Operational note: if `yaad-memory` does not appear in Codex's available skill
  list in a session, restart/open a new Codex session so skill discovery reloads.
  The file can still be read directly from disk for planning in the current
  workspace.
- Yaad skill caveat: bundled runtime paths mention `/Users/ali/apps/yaad`; on
  this machine/user and with Yaad hosted on another machine, treat those as
  examples until the actual Yaad host, endpoint, MCP command, and token flow are
  provided.
- Yaad HTTP MCP local setup: Zehn is configured to use
  `https://yaad.mmaliabbas.com/mcp` as an HTTP MCP server named `yaad`, with the
  `Authorization` header set to `Bearer ${YAAD_AGENT_TOKEN}`.
- PicoClaw local patch: MCP HTTP/SSE header values now expand environment
  variables before being sent, so Yaad tokens can stay in
  `.picoclaw/secrets/yaad-zehn-mbp-i7.env` instead of `config.json`.
- Verification status: `go test ./pkg/mcp` passes after the header expansion
  patch. After DNS propagation/cache refresh, `picoclaw mcp test yaad` using
  Zehn's config and `/Users/aliai/.picoclaw-zehn/secrets/yaad-zehn-mbp-i7.env`
  succeeds. Yaad HTTP MCP is reachable and lists 14 tools.
- Historical agent setup update: Ali provided the real 10 LogicIgniter bundle
  names and 51 app names. Zehn originally created slugged bundle and app-owner
  agent IDs from `.picoclaw/agents/portfolio.json`. Current active execution
  uses bundle owners plus discipline specialists; the old persistent
  app-owner agents are not active.
- Discord routing update: Discord server `1487893479555203193` is configured
  with Ali-only allowlist user `1050544532216877136` and mention-only group
  behavior. Dispatch rules route the provided Discord channel IDs to Personal,
  CEO, CTO, CPO, COO, CISO, Engineering, Product, Operations, Security,
  Research, Docs, QA, DevOps, memory review, approvals, and bot health agents.
  A focused route check confirmed all configured channel IDs resolve to the
  intended agents.
- Shell env permanence update: `/Users/aliai/.zshrc` now sources
  `/Users/aliai/.picoclaw-zehn/secrets/yaad-zehn-mbp-i7.env` inside the existing
  Zehn/PicoClaw block, so new interactive terminals load `YAAD_AGENT_TOKEN`
  without storing the token in `config.json`. Verification with `zsh -ic`
  confirmed `PICOCLAW_HOME`, `PICOCLAW_CONFIG`, and `YAAD_AGENT_TOKEN` are
  present. Existing terminals need `source ~/.zshrc` or a new terminal.
- Autostart standard path: the UI-created LaunchAgent
  `/Users/aliai/Library/LaunchAgents/io.picoclaw.launcher.plist` now keeps the
  official `io.picoclaw.launcher` label and `RunAtLoad`, but its
  `ProgramArguments` point to `/Users/aliai/.picoclaw-zehn/bin/zehn-launcher-run`.
  That wrapper loads `PICOCLAW_HOME`, `PICOCLAW_CONFIG`, `PICOCLAW_BINARY`, and
  `/Users/aliai/.picoclaw-zehn/secrets/yaad-zehn-mbp-i7.env`, then execs
  `/Users/aliai/zehn/build/picoclaw-launcher -no-browser
  /Users/aliai/.picoclaw-zehn/config.json`. Verification after `launchctl`
  reload showed launcher PID `29319` on `127.0.0.1:18800`, gateway PID `29320`
  on `127.0.0.1:18790`, and Yaad MCP connected with 14 tools.
- OpenAI/ChatGPT OAuth investigation: Ali completed `picoclaw auth login
  --provider openai`, and `picoclaw auth status` shows provider `openai`,
  method `oauth`, active. The external `codex-cli` model entry was removed from
  `.picoclaw/config.json` and `.picoclaw/.security.yml`; Zehn now has only the
  `primary-openai` model configured as provider `openai`, model `gpt-5.5`,
  `auth_method: oauth`, enabled. In this PicoClaw fork, source inspection shows
  `provider=openai` plus `auth_method=oauth` is still routed internally through
  `createCodexAuthProvider()` and the ChatGPT/Codex backend. A config-only CLI
  smoke test reaches Yaad and OpenAI OAuth but returns an empty parsed response,
  so removing `codex-cli` fixed the external CLI-provider problem but not the
  OAuth response parsing/runtime issue.
- OpenAI/ChatGPT OAuth fix: upstream issue
  `https://github.com/sipeed/picoclaw/issues/2674` exactly matched Zehn's empty
  response symptom. Upstream PR `#2581` / commit `b72d1b83` was applied locally
  without committing. The fix accumulates `response.output_item.done` stream
  items and hydrates the final completed response only when `response.output` is
  empty. Verification: `go test ./pkg/providers/oauth/...` passes when local
  `httptest` listener permission is available; `make build` produced
  `build/picoclaw`; CLI smoke test with `primary-openai` model `gpt-5.5` and
  Yaad env sourced returned `oauth-ok`. Runtime launcher/gateway still need a
  UI restart before they use the rebuilt binary.
- User-facing identity preference: Ali wants all assistant communications to
  identify the system as `Zehn`, not `picoclaw` or `PicoClaw`. Agent roles
  should be expressed as specializations under Zehn, for example `Zehn,
  LogicIgniter CEO` or `Zehn CTO`, while keeping the underlying PicoClaw runtime
  name only for technical setup/debug context.
- Agent prompt implementation update: Zehn workspace `AGENT.md` files have
  been rewritten from generic repeated prompts into role-specific operating
  prompts. The runtime loads `AGENT.md`, `SOUL.md`, and `USER.md`;
  `IDENTITY.md` is not loaded when `AGENT.md` exists, so important
  identity/authority content was promoted into `AGENT.md`.
- Agent prompt coverage:
  - Core agents: `zehn-main`, `personal`, `li-ceo`.
  - Active leadership: CTO, CPO, COO, and CISO/security leadership.
  - Active/support departments: Engineering, Product, Operations, Security,
    Research, Docs, QA, DevOps/SRE.
  - Business departments: CMO, CRO, CFO, Legal, CHRO, CCO, CDAO, Marketing,
    Sales, Finance, Customer Success.
  - Bundle owners: all 10 bundles now have mini-GM prompts with app lists,
    rollup duties, and simultaneous-launch blocking language.
  - App context: all 51 app records are portfolio context, not persistent
    `li-app-*` execution agents. Use bundle owners and discipline specialists
    for active work.
- Agent prompt verification: checked active `AGENT.md` role coverage, bundle
  launch-blocking prompts, and specialist execution prompts. `go test
  ./pkg/agent` passed during that implementation period. CLI smoke test against
  `zehn-main` correctly returned Zehn's operating role and LogicIgniter
  escalation policy using the rewritten coordinator prompt.
- Companion prompt implementation update: the loaded runtime companion files
  have been populated alongside `AGENT.md`. Core, Personal, CEO, C-suite,
  active departments, business departments, and all 10 bundle-owner workspaces
  now have role-specific `SOUL.md`, `USER.md`, and `memory/MEMORY.md` content.
  App-specific detail belongs in the portfolio registry, Yaad, GitHub issues,
  and repo docs, not in persistent `li-app-*` execution workspaces.
- Historical Yaad memory policy note, superseded by
  `LOGICIGNITER_YAAD_SCHEMA_CONTRACT.md`: every Zehn workspace
  `memory/MEMORY.md` file declares Yaad as the preferred durable memory system.
  Agents should prefer Yaad reads and use selective/idempotent Yaad writes for
  material durable facts, decisions, preferences, runbooks, or architecture
  choices. Local `MEMORY.md` files are boot and
  runtime fallback context, not the final source of truth.
- Memory policy verification: 87 `memory/MEMORY.md` files exist under
  `.picoclaw`, all 87 contain `## Yaad Policy`, and all 87 contain the rule to
  update or queue durable local memory for Yaad.
- Runtime validation update: Ali confirmed the gateway was restarted from the
  launcher UI, Discord routing was tested, and Yaad behavior was checked. Local
  operation is now validated at the launcher, Discord, and Yaad MCP behavior
  layers.
- Next setup layer: create Zehn's first durable operating artifacts for
  LogicIgniter: org chart, role activation/ownership, portfolio registry,
  approval/escalation matrix, operating cadence, and GitHub/project mapping.
  These should live as reviewable local workspace artifacts first, then be
  promoted into Yaad and/or the LogicIgniter `business` repo after Ali approves
  the structure.
- LogicIgniter repo home decision: `/Users/aliai/logicigniter` will be the flat
  local home for LogicIgniter GitHub repositories. Each direct child directory
  should be a GitHub-backed repo. Do not introduce local-only duplicate
  workspaces or physical category folders like `apps/`, `core/`, or `platform/`;
  logical classification belongs in the `business` repo registry.
- GitHub workflow decision: work should be issue-first, branch-per-issue,
  PR-based, and GitHub-backed. Branch names should use the issue number only
  where practical, and must not contain `codex`, `claude`, `agent`, or
  `feature`. Do not push directly to `main`.
- Portfolio scope decision: the 51-app launch excludes the extra service repos
  `svc-contentaudit-grpc`, `svc-maintenanceroi-grpc`, `svc-mentormatch-grpc`,
  and `svc-socialspark-grpc`. These are post-launch candidate/legacy repos to
  review after the 51-app launch succeeds.
- LogicIgniter workspace implementation: created `/Users/aliai/logicigniter` as
  the flat local GitHub repo home and cloned all 73 `logicigniter/*` repos into
  it. Verification found 73 direct child Git repos.
- Business operating-system v1 issue/PR: created `logicigniter/business#49`
  titled `Define LogicIgniter operating system v1`, created branch `49`, and
  opened draft PR `logicigniter/business#50`.
- Business operating-system v1 docs added in draft PR #50:
  `systems/zehn-operating-model.md`, `systems/github-control-plane.md`,
  `systems/repo-registry.yaml`, `organization/operating-organization.md`,
  `governance/approval-escalation-policy.md`, and
  `portfolio/launch-portfolio-v1.md`, plus README links.
- Registry verification for PR #50: YAML parsed successfully with 73 total repo
  entries, 51 launch app repos, and 4 post-launch candidate repos.
- GitHub Project setup: created the single organization-level Project
  `LogicIgniter Operating System` at
  `https://github.com/orgs/logicigniter/projects/1`.
- GitHub Project fields added: `Department`, `Bundle`, `App`, `Priority`,
  `Risk`, `Approval Required`, `Owner Agent`, and `Target Date`. GitHub default
  fields already provide title, assignees, status, labels, linked PRs,
  milestone, repository, reviewers, parent issue, and sub-issue progress.
- GitHub Project initial items: added `logicigniter/business#49` and draft PR
  `logicigniter/business#50`. Set the issue item to CEO/P0/Medium/Ali/li-ceo
  and the PR item to CEO/li-ceo.
- GitHub Development-link correction: branch names should remain issue-led but
  use an issue number plus short slug, such as `49-operating-system-v1`, rather
  than a pure numeric branch like `49`. GitHub does not infer issue Development
  links from branch names alone. The reliable automatic flow is to create the
  branch through `gh issue develop <issue> --name <issue-short-slug> --base
  main --checkout`, then create the PR from that linked branch with a closing
  keyword in the PR body.
- Operations cadence v1: after PR #50 was merged, created
  `logicigniter/operations#1`, created linked development branch
  `1-operating-cadence-v1` through `gh issue develop`, drafted and pushed
  operations cadence docs, and opened draft PR `logicigniter/operations#2`.
  Verification showed PR #2 has `closingIssuesReferences` for operations issue
  #1, confirming the corrected Development-link flow works.
- Operations cadence v1 merge: Ali merged `logicigniter/operations#2` on
  2026-05-03. Local `operations/main` was fetched, pruned, and fast-forwarded
  to merge commit `ae4652be135de0390e830c4b9bd375c7fcf292f3`.

## Agent-To-Agent Delegation Research

- Source audit conclusion: current Zehn/PicoClaw supports multi-agent inbound
  routing and ephemeral subturn helpers, but it does not yet provide true
  persistent private agent-to-agent delegation.
- Inbound routing is external-context driven. `ProcessDirectWithChannel` creates
  a normal inbound message, and `resolveMessageRoute` chooses the agent from
  dispatch rules based on channel/chat/sender context. There is no direct
  target-agent parameter in this path.
- The `spawn` tool accepts `agent_id`, but in the current direct subturn path
  that value is used only for the parent allowlist check. It is not passed into
  `SubTurnConfig`, and the spawned turn does not switch to the target agent's
  workspace, prompt, memory, or sessions.
- The `subagent` tool is synchronous but generic. It has no `agent_id`
  argument, uses a generic subagent prompt, and also runs as an isolated
  subturn rather than as a named persistent Zehn agent.
- `spawnSubTurn` intentionally uses an ephemeral session store and shallow-copies
  the parent agent. This is useful for private helper work during one active
  turn, but it is not a durable delegation record and does not create a
  persistent private conversation with CTO/Product/Operations/etc.
- Existing no-code options:
  - Use GitHub Issues/Projects and Yaad as a durable delegation ledger. This is
    persistent and auditable, but not a live private agent call.
  - Use Discord role channels for routed agent conversations. This is live and
    uses real configured agents, but it is not private internal delegation.
  - Use Yaad as an internal inbox plus cron/heartbeat polling. This is private
    and durable, but delayed and operationally heavier.
- Recommended Zehn direction: add a private built-in delegation capability
  rather than overloading the existing `spawn` semantics. The new capability
  should preserve the existing helper-subturn behavior while introducing a
  separate, auditable internal agent bus.
- Proposed private tool: `delegate_to_agent`.
  - Required arguments: `agent_id`, `task`.
  - Optional arguments: `mode` (`sync` or `async`), `thread_key`, `priority`,
    `due`, `artifact_refs`, `github_issue`, `yaad_memory_scope`.
  - Enforce the existing `subagents.allow_agents` relationship before running
    delegation.
  - Run the target as its real `AgentInstance`, with its real `AGENT.md`,
    `SOUL.md`, `USER.md`, workspace, memory policy, tools, and model config.
  - Use a private internal channel/session scope such as
    `internal:delegation:<parent_agent>:<target_agent>:<thread_key>`.
  - Persist the request, target response, status, and important decisions to
    Yaad when available, with local workspace memory/session fallback.
  - Support `async` mode for long-running work and `sync` mode for bounded
    advisory/debate requests.
  - Return concise results to the calling agent and keep a durable task record
    so results are not lost if the original user turn ends.
- Companion tools should follow once the first delegation path works:
  - `delegation_status` to inspect queued/running/completed tasks.
  - `delegation_inbox` for an agent to review its private assigned work.
  - `delegation_complete` or automatic completion recording for async work.
- LogicIgniter operating interpretation:
  - CEO remains Ali's main interface and primary company orchestrator.
  - CEO delegates privately to C-suite, departments, bundle owners, and app
    owners through the private delegation tool.
  - GitHub remains the execution control plane for real work; Yaad remains the
    durable memory and delegation ledger; Discord remains the human interface.
  - External side effects still require the approval policy already captured in
    this ledger.
- Implementation posture: this should be treated as Zehn-private first. If it
  later becomes upstreamable, upstream should receive a generic extension point
  for persistent internal agent delegation, not LogicIgniter-specific behavior.

## Delegated Agent Meeting System

- Delegation and meetings are distinct but connected systems:
  - Delegation assigns work or requests judgment from a target agent.
  - Meetings coordinate multiple agents around a problem, decision, plan, or
    review.
- Meeting chairing model: department heads may chair meetings inside their own
  domain. The CEO does not need to chair every meeting.
- CEO remains the company-level orchestrator and approval reviewer.
- Preferred example flow:
  1. Ali sends `li-ceo` a strategic request in Discord, for example:
     `We need to increase sales by 30% in the next 2 weeks`.
  2. `li-ceo` opens an executive objective and chairs the initial meeting with
     the relevant domain head, such as `li-cro` for sales.
  3. `li-ceo` gives direction, constraints, timeline, approval limits, and
     expected reporting cadence.
  4. The domain head chairs a working meeting with peers/subordinates, such as
     Sales, Marketing, Product, Finance, Customer Success, and relevant
     bundle owners, virtual app owners, and specialist roles.
  5. The domain head produces a consolidated recommendation, meeting notes,
     participant list, timeline, risks, and required approvals.
  6. `li-ceo` reviews the recommendation, challenges assumptions, requests
     revisions if needed, and then sends Ali a final approval request.
  7. After approval, execution tickets/issues are created where appropriate.
- Meeting output policy: default output is one consolidated recommendation from
  the chair. Detailed meeting notes, participants, discussion summary, timeline,
  options considered, dissent/risk notes, and follow-ups must be preserved for
  lookup.
- Meeting artifact path preference: curated meeting records should later live in
  the LogicIgniter `business` repo under a path such as
  `business/meetings/meeting_{id:datetime}.md`. During Zehn setup, local
  workspace planning copies may live under `.picoclaw/workspace/memory`.
- GitHub policy:
  - GitHub issues are for executable work, decisions requiring tracked action,
    or implementation follow-up.
  - Meeting notes may be summarized into issue comments after the meeting is
    complete.
  - Participating agents should add focused comments to the related issue when
    their domain position, commitment, risk, or acceptance criteria matter.
  - Raw meeting chatter should not be dumped into GitHub by default.
- GitHub Project policy: the organization project is the work tracker, not the
  company brain. It should track status, owner, priority, risk, approval need,
  due date, department, bundle/app, linked PRs, and repository. Yaad plus curated
  `business` repo docs remain the durable company brain.
- Business-domain records: sales, finance, legal, pricing, hiring, customer,
  partnership, and strategy discussions should be documented in the `business`
  repo and Yaad first. GitHub issues should be created only when execution,
  approval, or tracked follow-up is needed.
- Implementation docs created for this design:
  - `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`
  - `.picoclaw/workspace/memory/AGENT_MEETING_SYSTEM.md`

## 2026-05-06 Runtime Rollout Config Inspection And Patch

- Active config inspected: `/Users/aliai/.picoclaw-zehn/config.json`.
- Local-first posture remains in force:
  - Gateway host is `127.0.0.1`.
  - Launcher wrapper uses `/Users/aliai/.picoclaw-zehn` as `PICOCLAW_HOME`.
  - Launcher wrapper sources Yaad env from
    `/Users/aliai/.picoclaw-zehn/secrets/yaad-zehn-mbp-i7.env`.
  - Discord remains allowlisted to Ali only.
  - Telegram and Slack remain disabled.
- Delegation and meeting runtime tools were enabled in config:
  - `delegate_to_agent`
  - `delegation_status`
  - `delegation_inbox`
  - `start_agent_meeting`
- Delegation memory metadata was set for Zehn private runtime:
  - `project_key`: `zehn`
  - `labels`: `zehn`, `logicigniter`, `delegation`
  - `source`: `zehn-delegation`
- Initial department-head allowlists were added for practical LogicIgniter
  chaired-meeting flows:
  - CRO can coordinate Sales, Marketing, Customer Success, Finance, Product,
    Research, CMO, CFO, and CCO.
  - CMO can coordinate Marketing, Research, Docs, and Content/Marketing suite.
  - CFO can coordinate Finance and Finance/Revenue Intelligence suite.
  - Legal, CHRO, CCO, CDAO, Engineering, Product, Operations, Sales,
    Marketing, Finance, Security, Customer Success, and Research now have
    first-pass internal allowlists.
- GitHub artifact publishing remains intentionally out of the first live smoke
  path. Use Yaad plus local records first; enable GitHub issue/comment outputs
  only after a clean staged runtime test.
- A config backup was saved locally at:
  `/Users/aliai/.picoclaw-zehn/config.json.pre-delegation-rollout`.
- Binaries were rebuilt from current source:
  - `build/picoclaw`
  - `build/picoclaw-launcher`
- Verification passed:
  - `go test ./pkg/config -run 'Delegation|AgentConfig|LoadConfig' -count=1`
  - `go test ./pkg/agent ./pkg/tools ./pkg/config -run 'Delegation|Delegate|Meeting|GitHub|Artifact|Memory|Status|Inbox|Publisher|Participant|Failure|Cancel' -count=1`
  - `go test ./pkg/agent ./pkg/tools ./pkg/config ./pkg/channels/discord -count=1`
- Config-aware CLI status verified the rebuilt binary is using the right home
  when `PICOCLAW_HOME` and `PICOCLAW_CONFIG` are exported, and OpenAI OAuth is
  authenticated under that active runtime.
- Live CLI smoke was attempted with the same env file and rebuilt binary.
  Result: not passed yet because Yaad HTTP MCP is currently unreachable through
  the published Cloudflare endpoint. Direct checks to
  `https://yaad.mmaliabbas.com/healthz`, `/readyz`, and `/mcp` returned HTTP
  `530` for both IPv4 and IPv6 Cloudflare addresses. This points to the Yaad
  origin/tunnel side, not the Zehn config or token file. Do not broaden Discord
  testing until Yaad reachability is restored or a deliberate temporary
  MCP-disabled smoke path is chosen.

## 2026-05-07 Private LogicIgniter Operating Artifacts

- After partial live validation of Discord, Yaad MCP, delegation, and meeting
  flows, created the first reviewable Zehn-private LogicIgniter operating
  artifacts under `.picoclaw/workspace/memory/`.
- These files are private runtime/boot artifacts for review. They are not yet
  promoted to Yaad or the LogicIgniter `business` repo:
  - `LOGICIGNITER_ORGANIZATION_TREE.md`
  - `LOGICIGNITER_PORTFOLIO_REGISTRY_V1.md`
  - `LOGICIGNITER_APPROVAL_ESCALATION_MATRIX.md`
  - `LOGICIGNITER_OPERATING_CADENCE.md`
  - `LOGICIGNITER_GITHUB_CONTROL_PLANE.md`
- Current posture: use these files as the local operating draft set for Zehn.
  Promote stable facts to Yaad under `scope_type=organization` and
  `external_key=logicigniter` after Ali approves the content. Promote polished
  operating docs to the LogicIgniter `business` repo only through the approved
  issue/branch/PR workflow.
- Activation update: Ali wants configured LogicIgniter executive, department,
  bundle-owner, and specialist agents treated as active now. The current
  emphasis is CEO/CTO-led delegation across repos and the org, with business
  functions active for execution, readiness, process setup, and approval-gated
  decisions.
- LogicIgniter repo access update: active config now keeps
  `restrict_to_workspace=true` and `allow_read_outside_workspace=false`, while
  adding explicit file-tool allowlists for `/Users/aliai/logicigniter`:
  `allow_read_paths=["^/Users/aliai/logicigniter(/|$)"]` and
  `allow_write_paths=["^/Users/aliai/logicigniter(/|$)"]`. This is intended to
  let agents read/write trusted LogicIgniter repos without opening the rest of
  the machine. `exec.allow_remote=false` remains in force.
- Internal delegation policy update: Ali prefers not to restrict agent-to-agent
  communication. All 87 configured agents now have `subagents.allow_agents=["*"]`
  in the private runtime config. Relevance should be controlled by role
  discipline and task wording: agents may talk to any configured agent, but
  should consult only roles that materially improve the work.
- LogicIgniter business model clarification: LogicIgniter is not only the
  51-service SaaS/API portfolio and 10 canonical LogicIgniter suites. It is also a
  software development company capable of taking new requirements documents
  through intake, planning, architecture, implementation, QA, delivery, and
  support. New project work must be classified before execution and must not be
  automatically mixed into the 51-service portfolio.
- Software delivery workspace decision: project command workspaces live under
  `/Users/aliai/projects/{project-slug}`. These are shared project workspaces,
  not agent workspaces. Agent workspaces remain under `.picoclaw` for identity,
  memory, session context, and runtime records.
- Software delivery operating artifact: added
  `.picoclaw/workspace/memory/LOGICIGNITER_SOFTWARE_DELIVERY_SYSTEM.md` as the
  private local protocol for project intake, workspace setup, role ownership,
  repo/GitHub policy, Yaad project scopes, lifecycle, and approval boundaries.
  Promote stable durable rules to Yaad under `scope_type=organization`,
  `external_key=logicigniter`; use project-specific Yaad scope
  `scope_type=project`, `external_key=project:{project-slug}` once a project is
  approved.

## Future: Zehn Main Self-Monitoring And Improvement Loop

- Add this after the current organization dashboard, delegation, meeting,
  Yaad, Discord, and LogicIgniter hierarchy setup is stable.
- Goal: `zehn-main` becomes the runtime SRE/operator for Zehn itself.
- Operating posture: autonomous observation and diagnosis are allowed; material
  runtime changes require approval unless explicitly classified as safe
  auto-actions.
- Monitoring targets:
  - launcher and gateway health
  - `/health`, `/ready`, and `/reload` behavior
  - Discord/channel startup, connection, routing, and allowlist errors
  - Yaad MCP connection, auth, tool schema, scope, and memory-write failures
  - OpenAI/provider auth, model, rate-limit, timeout, and malformed-response
    failures
  - delegation and meeting failures, stuck async jobs, inbox backlogs, and
    stale running records
  - cron failures and scheduled-task drift
  - exec/tool-denial events and unsafe command attempts
  - skill install/use failures and stale local skill docs
  - config drift, local secret/env gaps, and restart-required state
  - frontend launcher errors and org-dashboard status anomalies
- Safe autonomous outputs:
  - health summaries to the operator channel
  - local diagnostic reports
  - Yaad observations under the correct organization/project/agent scope
  - GitHub issues for tracked maintenance work, after GitHub artifact policy is
    enabled
  - non-runtime status docs and runbook drafts
  - proposed remediation plans with evidence, risk, and rollback notes
- Approval-required actions:
  - editing `.picoclaw/config.json` or `.security.yml`
  - changing provider, MCP, channel, cron, exec, skill, or file-tool policy
  - restarting launcher/gateway or calling reload outside an approved procedure
  - installing packages or enabling remote skill registries
  - modifying Zehn source code
  - pushing branches, opening PRs, merging PRs, or deleting branches
  - changing Yaad scopes/tokens or memory schema
  - enabling new external channels or broadening `allow_from`
- Preferred implementation shape:
  1. Read-only health snapshot API or tool for gateway, launcher, channels,
     MCP, provider, records, cron, and recent logs.
  2. `zehn-main` cron heartbeat that runs a bounded diagnostic prompt and writes
     a concise health record.
  3. Failure classifier that groups known issues into auth, network, provider,
     config, channel, tool, memory, delegation, meeting, and frontend buckets.
  4. Maintenance queue backed by local records plus Yaad durable memory.
  5. Optional GitHub issue creation only for approved maintenance classes.
  6. Approved execution path through the existing task-loop/branch/PR process,
     not direct self-modification.
- Design constraint: this is a self-monitoring system, not unrestricted
  self-editing. Zehn should behave like a careful SRE: observe, diagnose,
  propose, request approval, then execute through a controlled path.

## Revision: Zehn Executable GitHub Work Policy

Date: 2026-05-10

This revises earlier draft-PR and explicit-merge-only setup notes for the
current private development phase.

- Executable LogicIgniter repo work should be represented by a GitHub issue
  using the Zehn executable-work structure: goal, owning repo, scope,
  acceptance criteria, verification command, integration mode, risk, sensitive
  areas, review required, and agent execution notes.
- Agents may autonomously pick up only issues labeled `zehn:ready` that are not
  already `zehn:claimed`, `zehn:in-progress`, `zehn:blocked`, or
  `approval:ali-required`.
- Claiming means adding `zehn:claimed` and `zehn:in-progress`, commenting the
  agent ID, timestamp, intended branch, and verification command, then
  re-reading the issue to confirm no newer claim exists.
- Branch names should start with the issue number and should not contain
  `codex`, `claude`, `agent`, or `feature`.
- Each execution pass should use a dedicated Codex execution session, run the
  repo verification command, commit scoped changes, push the branch, delegate
  the commit hash to required internal reviewers, apply review fixes, open a
  normal PR, and request `@codex review`.
- Do not use draft PRs for Zehn-executable work.
- Treat Codex 👀 as review started, not approval.
- Treat post-review 👍 or a formal approving Codex review as the Codex approval
  signal.
- Low-risk PRs may merge only when `zehn:merge-ready` is present,
  verification passed, required internal reviews passed, GitHub checks passed,
  Codex approval signal is present, and no `approval:ali-required`,
  `risk:high`, or `risk:critical` label exists.
- High-risk, critical-risk, auth, billing/payment, secrets, migrations/data,
  infra/deploy, security-sensitive, customer-facing, public, legal, financial,
  production, or otherwise approval-gated work still requires explicit Ali
  approval before merge.
- Hard rule remains unchanged: never leave any LogicIgniter child repo dirty.
