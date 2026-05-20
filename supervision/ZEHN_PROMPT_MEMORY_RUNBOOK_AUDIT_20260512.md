# Zehn Prompt, Memory, Runbook, And Cron Audit - 2026-05-12

Status: evidence pass 1 complete; remediation not applied.

Purpose: audit Zehn local Markdown operating content, agent files, memory files,
runbooks, and cron payloads for stale facts, contradictions, invalid commands,
unnecessary permission barriers, missing authority, over-broad authority, and
workflow gaps that prevent Zehn agents from producing useful LogicIgniter work.

Ground rules for this audit:

- No runtime restart or reload is part of this audit.
- Findings must cite file paths and evidence.
- Proposed fixes are documented only; no prompt/config cleanup is performed in
  this pass unless Ali explicitly requests it later.
- The audit document is updated as evidence is collected so findings do not
  depend on conversational memory.

## Inventory Snapshot

Collected with:

```bash
find .picoclaw -type f \( -name '*.md' -o -name 'AGENT.md' -o -name 'IDENTITY.md' -o -name 'SOUL.md' -o -name 'USER.md' \) | sort
find operations supervision workspace -type f \( -name '*.md' -o -name 'SKILL.md' \) 2>/dev/null | sort
jq -r '.jobs[] | ...' .picoclaw/workspace/cron/jobs.json
```

Initial scope includes:

- `.picoclaw/workspace/operating-prompts/*.md`
- `.picoclaw/workspace/memory/*.md`
- `.picoclaw/workspace*/AGENT.md`, `IDENTITY.md`, `SOUL.md`, `USER.md`
- `.picoclaw/workspace*/memory/*.md`
- `.picoclaw/workspace*/evidence/*.md`
- `supervision/*.md`
- `supervision/zehn_feature_tasks/*.md`
- `operations/*.sh` where referenced by prompts/runbooks
- cron job payloads from `.picoclaw/workspace/cron/jobs.json`

Out-of-scope unless referenced by Zehn operating prompts:

- third-party bundled skill docs not authored for this LogicIgniter operating
  system;
- historical session JSONL transcripts, except when needed as evidence for
  whether a prompt produced failures.

Inventory size:

- Active Markdown/agent identity inventory under `.picoclaw/workspace*`:
  246 files.
- Current cron jobs in `.picoclaw/workspace/cron/jobs.json`: 6.
- Current LogicIgniter agent workspaces with `AGENT.md`: 40.
- Current old persistent app-agent workspaces: 0.
- Current bundle-owner workspaces: 10.

This pass focused first on files that can actively steer runtime behavior:
cron payloads, operating prompts, current agent boot files, current memory
policy files, setup/readiness planning files, and runbooks/scripts referenced
from prompts. Historical evidence files were sampled and classified as a class
of risk rather than rewritten line by line.

## Initial High-Risk Findings

### F-001: Cron payload inventory changed after the earlier audit

Evidence:

- Earlier cron extraction from `.picoclaw/workspace/cron/jobs.json` listed 11
  scheduled jobs, including old operating-check and specialist queue jobs.
- Current extraction lists 6 scheduled jobs:
  - `li-weekly-plan`
  - `li-daily-synthesis`
  - `zehn-operations-monitor`
  - `li-weekly-review`
  - `li-ceo-daily-sync`
  - `li-nonexec-weekly-pulse`
- The old `logicigniter-ceo-operating-check` and
  `logicigniter-engineering-check` prompts now live under
  `.picoclaw/workspace/operating-prompts/archive/`.

Impact:

- The earlier 11-job finding is no longer current evidence.
- The active scheduler should be evaluated against the current 6-job operating
  model, not against archived operating-check prompts.

Required follow-up:

- Keep archived prompt references out of current-runtime conclusions unless an
  active cron job, current prompt, or dispatch rule still points to them.
- Continue verifying current jobs with
  `operations/verify-logicigniter-cron-routing.sh`.

### F-002: Several cron payloads say "Report HEARTBEAT_OK only when no claimable issue exists" but omit active PR review work

Evidence:

- Cron payloads for specialist queues focus on "unclaimed `zehn:ready`" issues.
- Archived scheduler/worker prompt files partially fixed this:
  `.picoclaw/workspace/operating-prompts/archive/logicigniter-specialist-work-check.md`
  and
  `.picoclaw/workspace/operating-prompts/archive/logicigniter-specialist-worker-check.md`
  both contain an active PR review queue requirement.
- The current 6-job scheduler no longer has those specialist queue entries, so
  this finding should be treated as historical unless a specialist queue is
  reintroduced.
- Recent GitHub evidence showed real active work in open PRs and
  `zehn:review-internal` / `zehn:in-progress`, including:
  - `logicigniter/scripts` PR #8
  - `logicigniter/integration_tests` PR #14
  - `logicigniter/scripts` PR #6
  - `logicigniter/business` PR #54
  - `logicigniter/integration_tests` PR #12
  - `logicigniter/scripts` PR #4
  - `logicigniter/integration_tests` PR #11
- Recent delegation summaries show specialists returning `HEARTBEAT_OK` or
  `NO_MATCHING_ISSUES` when no fresh `zehn:ready` issue exists, even while
  review-stage PRs are open.

Impact:

- Agents can appear healthy while not advancing the actual review/merge queue.
- Work stalls in PR review instead of moving through QA, DevOps, Security,
  Docs, and post-merge reconciliation.
- The current prompt files reduce the risk, but the cron summary and older
  delegation records can still bias agents toward issue-only inspection.

Required follow-up:

- Cron payloads and agent boot files need the same explicit "active PR review
  queue" wording already present in the worker prompt.
- This is a prompt/workflow defect, not a Go runtime defect.

### F-003: `verify-pr.sh` is referenced as a standard command but is not present

Evidence:

- `.picoclaw/workspace/operating-prompts/archive/logicigniter-specialist-worker-check.md`
  instructs agents to run:
  `/Users/aliai/logicigniter/scripts/verification/verify-pr.sh --repo <repo> --issue <issue>`
- Local check found no file at:
  `/Users/aliai/logicigniter/scripts/verification/verify-pr.sh`
- Direct filesystem check on 2026-05-12 showed
  `/Users/aliai/logicigniter/scripts/verification/` exists and contains many
  verification scripts, but not `verify-pr.sh`.
- `git -C /Users/aliai/logicigniter/scripts status --short --branch` showed
  the checked-out scripts repo is on
  `chore/7-align-local-bff-auth-discovery`, not the older
  `chore/3-standard-pr-verification` branch that agent memories say contained
  `verify-pr.sh`.

Impact:

- Agents following the prompt will repeatedly hit a missing-command path.
- This can create unnecessary blockers and encourage inconsistent fallback
  verification.

Required follow-up:

- Either create the standard `verify-pr.sh` in LogicIgniter, or update Zehn
  prompts to reference the actual current verification commands per repo until
  the standard wrapper exists.
- Until it exists on the active scripts branch or main, prompts should not make
  it an unconditional gate without naming the fallback verification path.

### F-004: Engineering prompt still pushes broad CTO investigation and can hit tool iteration limits

Evidence:

- Recent delegation records include repeated CTO completions with:
  `I've reached max_tool_iterations without a final response...`
- The archived prompt
  `.picoclaw/workspace/operating-prompts/archive/logicigniter-engineering-check.md`
  asks CTO to inspect many surfaces: GitHub Projects/issues, PRs, failed checks,
  recent repo changes, business plans, integration readiness, quality gates,
  security/devops blockers, service makeover status, local repo state, and more.

Impact:

- CTO checks consume tool budget and sometimes produce no final answer.
- Cron marks the outer job `ok` because a turn completed, but the business
  outcome is degraded.

Required follow-up:

- Keep CTO checks bounded to a small number of active PRs/issues per run.
- Make existing open PRs the default first-class queue.
- Stop before the tool budget is exhausted with a terminal partial status.

### F-005: Readiness audit contains stale exec capability facts

Evidence:

- Current config values:

```text
.tools.exec.allow_remote = true
.tools.cron.allow_command = false
.tools.cron.exec_timeout_minutes = 12
.agents.defaults.max_tool_iterations = 50
```

- `.picoclaw/workspace/memory/ZEHN_READINESS_AUDIT.md` still says:
  - `Exec: enabled, remote use blocked by allow_remote: false`
  - `Exec is enabled, but remote exec is blocked by allow_remote: false`
  - `Exec | enabled, remote blocked`

Impact:

- Agents reading the readiness audit may incorrectly believe remote-capable exec
  is disabled.
- Security posture decisions become ambiguous: the current runtime is more
  permissive than the stale audit says.

Required follow-up:

- Update or archive stale readiness audit sections.
- Keep one canonical runtime capability summary instead of several copied
  status tables.

### F-006: Portfolio/bundle owner identity is internally inconsistent

Evidence:

- Earlier `LOGICIGNITER_PORTFOLIO_REGISTRY_V1.md` content redefined the
  original 10 MCQ/user bundle names into alternate market-package labels.
- Bundle agent IDs remained old suite slugs while role files presented alternate
  package labels. This has since been remediated to canonical suite names.
- User-provided original bundle names included:
  - SaaS Growth & Retention Suite
  - E-commerce Operations Suite
  - Finance & Revenue Intelligence Suite
  - Legal & Compliance Suite
  - Developer & DevOps Suite
  - Content & Marketing Suite
  - HR & Workforce Suite
  - Professional Services Suite
  - Real Estate & Property Suite
  - Education Suite

Impact:

- Agents can confuse product-market bundle names with internal agent IDs.
- Product, docs, and route planning may drift from the user's original
  portfolio taxonomy without an explicit approval record.
- Existing bundle-owner workspaces are still active, but their names no longer
  mean what their IDs say.

Required follow-up:

- Decide whether "Ignite ..." bundles are approved canonical product packaging
  or only a draft hypothesis.
- If draft, mark them as provisional and keep original 10 suite names as
  canonical until Ali approves replacement packaging.
- If approved, create a clear mapping table from original suite -> new market
  bundle -> agent ID -> hosted route.

### F-007: Software delivery system blocks repo/GitHub creation too strongly for future autonomous custom projects

Evidence:

- `LOGICIGNITER_SOFTWARE_DELIVERY_SYSTEM.md` says Ali approval is required
  before creating GitHub repositories, issues, projects, branches, PRs, or repo
  creation unless a later policy explicitly grants automation.
- Current LogicIgniter operating policy grants standing authority for private
  setup work to create issues, issue-linked branches, commits, pushes, normal
  PRs, and review requests inside policy.

Impact:

- For new custom/internal projects, agents may stall at the first GitHub issue
  or branch step even when the work is internal and low-risk.
- The rule may be valid for new repository creation and external commitments,
  but it is too broad if it blocks ordinary issue/branch/PR work after Ali has
  approved a project lane.

Required follow-up:

- Split "new repo creation" from "issue/branch/PR work in an approved repo".
- Add a project-specific standing-authority model:
  - local draft/intake allowed;
  - issue/branch/PR allowed after Ali approves the project lane and repo;
  - new repo/external/customer/prod commitments still require Ali approval.

### F-008: Standard verification policy is correct in principle but currently impossible in practice

Evidence:

- `LOGICIGNITER_GITHUB_CONTROL_PLANE.md` requires:
  `/Users/aliai/logicigniter/scripts/verification/verify-pr.sh --repo <repo> --issue <issue>`
- Multiple agent memory entries also record that `verify-pr.sh` is absent or
  unavailable on the relevant branches.

Impact:

- Agents face a contradiction: policy says execution requires the standard
  wrapper, but the wrapper is unavailable.
- This encourages repeated "blocked" reporting or inconsistent fallback
  verification.

Required follow-up:

- Treat creation/merge of `verify-pr.sh` as a P0 control-plane prerequisite, or
  temporarily downgrade the policy to "use `verify-pr.sh` when present, else
  use documented repo-specific verification and file an issue to add/port the
  wrapper."

### F-009: Safety guard blocks local process control, but DevOps prompts expect restarts

Evidence:

- DevOps memory recorded that local BFF restart could not be executed because
  `kill`, `pkill`, Python/Ruby signal paths were blocked by exec safety guard.
- Current goal requires agents to update local repo and restart affected local
  LogicIgniter services after PR merge.
- The newly added post-merge script uses service launcher scripts but does not
  solve the already-running-process stop case for every service.

Impact:

- Agents may correctly identify that a service must restart but remain unable
  to stop the old process.
- This produces "working" analysis without actual runtime adoption of merged
  code.

Required follow-up:

- Provide trusted service-control scripts that avoid blocked command shapes and
  are explicitly allowed by Zehn's exec guard, or adjust allow patterns for
  narrow local service-control scripts only.
- Do not ask agents to manually `kill` processes if the guard blocks it.

### F-010: Historical evidence files are mixed with live runtime instruction files

Evidence:

- `.picoclaw/workspace-li-devops/evidence/*.md`,
  `.picoclaw/workspace-li-cto/memory/202605/*.md`, and
  `.picoclaw/workspace-li-qa/memory/202605/*.md` contain highly specific
  incident evidence: branch names, PIDs, commit SHAs, current ports, and
  current failure states.
- Runtime prompts and agent memories coexist in the same broad workspace tree.

Impact:

- Agents can accidentally treat stale evidence as current fact.
- Repeated incident notes can reinforce outdated blockers even after code,
  branches, or runtime state changes.

Required follow-up:

- Mark dated evidence files as historical evidence, not current state.
- Add a rule: before using any incident evidence older than the current run,
  re-check GitHub/repo/runtime state and cite fresh evidence.

### F-011: Active cron coverage omits several operational specialist roles

Evidence:

- `.picoclaw/workspace/cron/jobs.json` currently has 11 jobs.
- The scheduled specialist queues are:
  - `logicigniter-architect-work-queue`
  - `logicigniter-backend-work-queue`
  - `logicigniter-frontend-work-queue`
  - `logicigniter-ux-work-queue`
  - `logicigniter-integration-work-queue`
  - `logicigniter-data-ai-work-queue`
- There are no dedicated scheduled work queues for:
  - `li-devops`
  - `li-qa`
  - `li-security`
  - `li-docs`
- Yet the current work model repeatedly requires DevOps runtime reconciliation,
  QA verification, Security review, and Docs/evidence handling.

Impact:

- DevOps/QA/Security/Docs may only act when CEO/CTO/specialists delegate to
  them, not as active scheduled queues.
- This can leave PR review, runtime restart evidence, and review signoff waiting
  even when implementation specialists are active.

Required follow-up:

- Decide whether DevOps, QA, Security, and Docs need their own bounded scheduled
  queues.
- If yes, add them deliberately with role-specific labels or PR-review
  responsibilities, not generic broad scans.

### F-012: Specialist AGENT.md files still describe issue-only queue ownership

Evidence:

- Specialist boot files such as:
  - `.picoclaw/workspace-li-backend-developer/AGENT.md`
  - `.picoclaw/workspace-li-frontend-developer/AGENT.md`
  - `.picoclaw/workspace-li-architect/AGENT.md`
  - `.picoclaw/workspace-li-integration-engineer/AGENT.md`
  - `.picoclaw/workspace-li-data-ai-engineer/AGENT.md`
  - `.picoclaw/workspace-li-ux-designer/AGENT.md`
  say they "actively watch for `zehn:ready` GitHub issues" in their area.
- Those boot files do reference the worker prompt for scheduled queue checks,
  but their own role summary does not explicitly state that matching open PRs,
  internal review queues, merge blockers, and post-merge reconciliation are
  also part of the job.

Impact:

- When an agent relies on its own boot profile rather than the scheduler prompt,
  it can still see its job as new issue intake only.
- This matches observed behavior where agents reported no matching issues while
  open PRs were still waiting for review or runtime evidence.

Required follow-up:

- Update specialist role files to say their queue includes:
  - claimable `zehn:ready` issues;
  - matching open PRs requiring review/verification;
  - stale blocker cleanup;
  - post-merge reconciliation handoff when applicable.

### F-013: Archived CEO and engineering prompts still contain stricter `verify-pr.sh` wording

Evidence:

- `.picoclaw/workspace/operating-prompts/archive/logicigniter-ceo-operating-check.md`
  instructs the archived CEO flow to run:
  `/Users/aliai/logicigniter/scripts/verification/verify-pr.sh --repo <repo> --issue <issue>`.
- The same prompt says low-risk merges require "`verify-pr.sh` passed".
- `.picoclaw/workspace/operating-prompts/archive/logicigniter-engineering-check.md`
  has the same archived standard flow and merge gate.
- Current cron messages say "run `verify-pr.sh` when available" or
  "verify-pr.sh or documented repo verification", which is more accurate than
  the deeper prompt text.

Impact:

- If archived prompts are accidentally reactivated or copied into current
  prompts, agents may block on a missing wrapper even when a repo-specific
  fallback is acceptable.

Required follow-up:

- Keep the current active policy consistent:
  - preferred: `verify-pr.sh` when present on the active branch;
  - fallback: documented repo-specific verification plus issue/PR evidence.
- If archived prompts are restored, update them before activation.

### F-014: Setup planning still contains stale runtime snapshots and old counts

Evidence:

- `.picoclaw/workspace/memory/ZEHN_SETUP_PLANNING.md` contains old runtime
  verification details such as launcher/gateway PIDs, old post-restart state,
  and old agent-count milestones.
- Current config has 40 LogicIgniter agent workspaces, 0 old
  `workspace-li-app-*` directories, and 10 bundle owner workspaces.
- Current `jobs.json` contains 6 jobs and heartbeat is enabled.

Impact:

- Agents reading the full setup planning file can blend historical setup facts
  with current runtime truth.
- PIDs, job counts, and old verification state should not be used as current
  operational evidence.

Required follow-up:

- Add a "historical setup log, not current state" header to old planning
  sections.
- Create one short canonical current-state runtime summary that agents read
  before older setup notes.

### F-015: Portfolio naming conflict was found and remediated

Evidence:

- Earlier config and role files mixed the original 10 suite names with alternate
  market-package labels.
- Current config now names all 10 bundle agents with the original user-approved
  suite taxonomy:
  - SaaS Growth & Retention Suite
  - E-commerce Operations Suite
  - Finance & Revenue Intelligence Suite
  - Legal & Compliance Suite
  - Developer & DevOps Suite
  - Content & Marketing Suite
  - HR & Workforce Suite
  - Professional Services Suite
  - Real Estate & Property Suite
  - Education Suite
- The original user-provided taxonomy was 10 named suites:
  SaaS Growth & Retention, E-commerce Operations, Finance & Revenue
  Intelligence, Legal & Compliance, Developer & DevOps, Content & Marketing,
  HR & Workforce, Professional Services, Real Estate & Property, and Education.
- Bundle `AGENT.md`, `IDENTITY.md`, and `USER.md` files now use the canonical
  suite names rather than the earlier alternate labels.

Impact:

- This was a live config/prompt identity issue before remediation.
- Current impact is residual audit risk only: older documents may still mention
  the alternate packaging labels and should not be treated as current canon.

Required follow-up:

- Keep the original 10 suite names as the canonical user-supplied taxonomy.
- Reject future prompt/config changes that rename bundle owners unless Ali
  explicitly approves a taxonomy change.

### F-016: Post-merge reconciliation script is useful but may not be executable by Zehn under current tool safety

Evidence:

- `operations/logicigniter-post-merge-reconcile.sh` exists and is allowlist
  based.
- It runs `git fetch`, `git switch main`, `git pull --ff-only`, then starts or
  restarts mapped services by calling LogicIgniter scripts.
- It uses shell command shapes inside a script, which is good for avoiding
  arbitrary ad hoc commands in prompts.
- Current config allows exec remote and write paths for LogicIgniter/projects,
  but Zehn's exec safety guard previously blocked `git push`, `kill`, `pkill`,
  Python/Ruby signal paths, and some process-control operations.
- The post-merge script relies on downstream LogicIgniter launcher scripts to
  replace or reuse running processes. If those scripts internally need blocked
  process-control commands, this script can still fail.

Impact:

- The design is safer than letting agents improvise restart commands, but it is
  not yet proven end-to-end from inside Zehn.
- A merged PR may be pulled locally while the running service remains stale if
  restart scripts cannot replace existing processes.

Required follow-up:

- Run a non-destructive dry-run or controlled test after this audit, with Zehn
  using the exact command form agents will use.
- Add a clear failure-report requirement: if restart cannot replace a running
  process, DevOps must report the exact service, PID/port evidence when safe,
  script error, and manual/approved remediation path.

### F-017: The operating cadence doc intentionally says heartbeat must not use shell exec

Evidence:

- `.picoclaw/workspace/memory/ZEHN_OPERATING_CADENCE.md` says the built-in
  PicoClaw heartbeat is a single `zehn-main` supervisor loop.
- It explicitly says the heartbeat must not:
  - perform department work directly;
  - use shell execution during heartbeat;
  - create GitHub artifacts or mutate repos;
  - wake every agent.
- Current config has heartbeat enabled every 30 minutes.

Impact:

- This is not necessarily wrong, but it matters operationally: if Ali expects
  `zehn-main` heartbeat itself to inspect logs, GitHub, and repos with tools,
  the current written doctrine forbids that.
- The intended design instead uses cron jobs for tool-using work and heartbeat
  for supervisor/routing health.

Required follow-up:

- Confirm the desired contract:
  - heartbeat remains low-tool/no-shell supervisor only; or
  - heartbeat may run bounded diagnostics with shell access.
- If no-shell heartbeat remains, ensure all required actual monitoring exists
  as cron jobs. Currently Zehn operations monitoring is documented but not
  present as a visible cron job in `jobs.json`.

### F-018: Zehn operations monitor prompt exists but is not scheduled

Evidence:

- `.picoclaw/workspace/operating-prompts/zehn-operations-monitor.md` exists and
  points to `/Users/aliai/.picoclaw-zehn/logs/gateway.log`.
- The log path exists and `gateway.log` is large and active.
- `.picoclaw/workspace/cron/jobs.json` does not include a
  `zehn-operations-monitor` or equivalent bot-health scheduled job. The only
  `bot_health` mapping appears in dispatch rules, not as a cron payload.

Impact:

- The user expectation that `zehn-main` watches Zehn logs/issues/improvements
  is not fully implemented as an active scheduled task.
- The monitor prompt is currently a document, not guaranteed autonomous
  behavior.

Required follow-up:

- Add a bounded Zehn operations monitor cron if this remains a requirement.
- Keep it read-only by default: logs, config drift, failed delegations, MCP
  failures, channel failures, and restart-required status, with no self-edit or
  restart unless explicitly approved.

### F-019: Historical delegation records preserve old prompt text and should not be used as current instructions

Evidence:

- Old delegation JSON files include copied prompt bodies that say things like:
  - "open draft PRs" for execution work;
  - "do not create GitHub artifacts" during early simulation;
  - "app owners" as delegation targets;
  - specialist work checks that omit active PR review before `HEARTBEAT_OK`.
- Current policy now prefers normal PRs, not drafts, for Codex review.
- Current organization model demotes 51 app records from persistent execution
  agents to product context.

Impact:

- If agents mine old delegation records as examples, they can resurrect stale
  rules.
- This is likely because delegation records are both audit evidence and
  easy-to-read prompt examples.

Required follow-up:

- Add a top-level rule in current memory: delegation/meeting JSON records are
  historical evidence only and must not override current AGENT.md, operating
  prompt, config, GitHub, or Yaad facts.

### F-020: The current live agent count is lower and cleaner than older "87 agents" language

Evidence:

- Current config lists 40 LogicIgniter-related workspaces plus `zehn-main` and
  `personal`.
- Filesystem check found:
  - `40` `workspace-li-*` directories with `AGENT.md`;
  - `0` `workspace-li-app-*` directories;
  - `10` `workspace-li-bundle-*` directories.
- Old planning/memory text still mentions the earlier 87-agent model.

Impact:

- Old scale assumptions can drive wrong concurrency expectations, wrong
  delegation fanout, or concern about 51 app-owner agents that no longer exist.

Required follow-up:

- Replace old "87 agents" references in live instruction files with the current
  model:
  executives/departments/bundle owners/specialists, with 51 services as product
  context.

### F-021: Old draft-PR policy remains in setup history and delegation records

Evidence:

- Current live policy in:
  - `.picoclaw/workspace/AGENT.md`
  - `.picoclaw/workspace/memory/LOGICIGNITER_GITHUB_CONTROL_PLANE.md`
  - `.picoclaw/workspace/memory/LOGICIGNITER_APPROVAL_ESCALATION_MATRIX.md`
  - `.picoclaw/workspace/memory/LOGICIGNITER_OPERATING_CADENCE.md`
  says Zehn-executable work should use normal PRs, not drafts.
- `.picoclaw/workspace/memory/ZEHN_SETUP_PLANNING.md` still contains older MCQ
  and implementation notes saying Zehn may autonomously open draft PRs.
- Older delegation/meeting JSON records also contain early "no GitHub
  artifacts" and "open draft PR" instructions from simulation/setup phases.

Impact:

- Current policy is clear in the live operating docs, but old planning and
  historical records can mislead agents if they search memory broadly.

Required follow-up:

- Mark `ZEHN_SETUP_PLANNING.md` as historical plus current-status-indexed.
- Keep current GitHub execution policy as the canonical authority for PR type.

### F-022: `LOGICIGNITER_SOLUTION_PORTFOLIO_PLAN.md` still says app owners remain responsible

Evidence:

- The same file correctly says "Bundle owners and app owners created earlier
  under the old suite names should not be renamed casually."
- It also says under "Zehn Operating Implications":
  "App owners remain responsible for atomic service readiness even when a
  service participates in more than one solution bundle."
- Current organization config and filesystem state show no persistent
  `workspace-li-app-*` agents remain.
- Other current docs say the 51 services are product context, not persistent
  app-owner agents.

Impact:

- This sentence can resurrect the removed 51 persistent app-owner model.
- The intended replacement appears to be: app/service records remain product
  context; specialist agents and bundle/product/CTO roles own execution and
  readiness by specialty.

Required follow-up:

- Replace "App owners remain responsible" with a role-neutral service-readiness
  ownership model, for example:
  "Atomic service readiness is tracked through service records and owned by the
  relevant bundle/product/technical specialists, not persistent app-owner
  agents."

### F-023: Software delivery system approval barrier is still broader than the current standing authority

Evidence:

- `.picoclaw/workspace/memory/LOGICIGNITER_SOFTWARE_DELIVERY_SYSTEM.md`
  says GitHub issues, projects, branches, PRs, or repo creation require Ali
  approval unless later policy grants automation.
- `.picoclaw/workspace/memory/LOGICIGNITER_APPROVAL_ESCALATION_MATRIX.md`
  now grants standing authority for private setup/development work to create
  issues, issue-linked branches, commits, pushes, and normal PRs inside trusted
  LogicIgniter repos.
- This contradiction is especially important for future `/Users/aliai/projects`
  custom/internal project work.

Impact:

- Agents may unnecessarily stop before creating issues/branches/PRs for an
  approved internal project lane.
- Repo creation should remain approval-gated, but issue/branch/PR work inside
  an already-approved project/repo should not be blocked by a stale broad rule.

Required follow-up:

- Split the policy into:
  - new repo creation: Ali approval required;
  - local project workspace creation: allowed after CEO intake classification
    inside configured filesystem scope;
  - issue/branch/PR work in an approved trusted repo: allowed under standing
    authority and normal risk gates.

### F-024: Discord mention policy in planning does not match current config

Evidence:

- `.picoclaw/workspace/memory/ZEHN_SETUP_PLANNING.md` says the Discord server is
  configured with "mention-only group behavior."
- Current `.picoclaw/config.json` shows:
  - `channel_list.discord.allow_from` contains Ali's user ID and
    `discord:<id>` form.
  - `channel_list.discord.settings.mention_only = false`.
- The user previously asked to stop the "mention" requirement.

Impact:

- The current config appears aligned with the user's later preference, but the
  planning file says otherwise.
- Agents reading setup history may describe Discord behavior incorrectly.

Required follow-up:

- Update planning/status text to say current Discord is Ali allowlisted and not
  mention-only at the global Discord channel config level.
- Preserve any per-channel/group trigger details separately if they exist.

### F-025: Zehn operations monitor is documented as scheduled but only the prompt exists

Evidence:

- `.picoclaw/workspace/operating-prompts/zehn-operations-monitor.md` says:
  "You are `zehn-main` running a scheduled Zehn operations monitor check."
- Current `.picoclaw/workspace/cron/jobs.json` has no job that references this
  prompt.
- Current `channel_list.discord` includes bot-health dispatch mapping, but that
  is not the same as a scheduled monitor job.

Impact:

- This creates false confidence that Zehn is autonomously reviewing its own
  logs, failures, and useful-work output.
- The prompt is good, but unused unless manually invoked or wired into cron.

Required follow-up:

- Either add a scheduled bot-health monitor job or rename the prompt to "manual
  monitor template" until scheduled.

### F-026: Current audit itself created new untracked files that must be handled after review

Evidence:

- `git status --short` currently shows:
  - `?? operations/logicigniter-post-merge-reconcile.sh`
  - `?? supervision/ZEHN_PROMPT_MEMORY_RUNBOOK_AUDIT_20260512.md`
- `bash -n operations/logicigniter-post-merge-reconcile.sh` passes.

Impact:

- This is acceptable during the audit because the audit file is being
  incrementally written as requested.
- It must not be forgotten. The hard repo hygiene rule says we should end with
  a deliberate state: committed branch/PR, explicitly left untracked with
  explanation, or user-directed cleanup.

Required follow-up:

- After Ali reviews the audit, decide whether to commit the audit and
  post-merge script to an issue-linked branch/PR or keep/remove them.

### F-027: Agent workspaces contain stale issue/PR body draft files beside live prompts

Evidence:

- File inventory found draft body files such as:
  - `.picoclaw/workspace-li-ceo/tmp-pr-7-body.md`
  - `.picoclaw/workspace-li-operations/business-issue-53-body.md`
  - `.picoclaw/workspace-li-operations/issue-7-body.md`
  - `.picoclaw/workspace-li-operations/issue-9-body.md`
  - `.picoclaw/workspace-li-operations/issue-10-body.md`
  - `.picoclaw/workspace-li-operations/scripts-issue-1-body.md`
- These are not current policy documents; they are generated work artifacts.

Impact:

- Agents can mistakenly read old issue-body drafts as current operating policy
  or current work state.
- This adds noise to workspace search and increases the chance of stale
  GitHub/task descriptions being reused.

Required follow-up:

- Move draft issue/PR body artifacts into a dated `archive/` or `drafts/`
  folder with clear historical labels, or delete them once the corresponding
  GitHub artifacts exist.
- Add a rule that `*-issue-*-body.md`, `tmp-pr-*-body.md`, and similar files are
  drafts only and must be revalidated against GitHub before reuse.

### F-028: Current review scope is large enough that future cleanup needs automation, not manual text edits

Evidence:

- The Markdown/agent-file inventory under `.picoclaw/workspace*` includes 246
  files matching Markdown or agent identity patterns.
- Many files repeat identical policy paragraphs for Yaad, repo access, and
  engineering quality doctrine.

Impact:

- Manual one-by-one cleanup is risky and likely to reintroduce divergence.
- Repeated text has already diverged in places: setup history says draft PRs,
  current control-plane says normal PRs; readiness audit says 87 agents, current
  config has the new specialist model; solution plan says app owners remain
  responsible, current org model removed persistent app agents.

Required follow-up:

- After this audit, fix canonical policy documents first.
- Then update repeated agent boot snippets mechanically from canonical sources,
  with a verification script that searches for forbidden stale phrases.
- Avoid freehand rewriting all 246 files without an explicit stale-phrase test.

## Open Audit Questions

These are no longer open as unknowns; they are now tracked findings:

- Runtime instructions versus historical evidence: F-010, F-019, F-027.
- Files stale enough to archive or mark historical: F-014, F-019, F-021,
  F-027.
- Permission gates that are accidental blockers: F-007, F-009, F-016, F-023.
- Old app-owner references after specialist-agent shift: F-020, F-022.
- Docker/host-native contradiction: F-010, plus current operating docs now
  correctly say not to assume Docker is required. The main residual problem is
  stale historical memory, not current prompt direction.
- `HEARTBEAT_OK` hiding actionable work: F-002, F-011, F-012, F-018, F-025.

## Verified Good Signals

- Current config gives Zehn the intended local work access posture:
  - `tools.exec.allow_remote = true`
  - `tools.exec.timeout_seconds = 720`
  - `tools.cron.exec_timeout_minutes = 12`
  - `agents.defaults.max_tool_iterations = 50`
  - `agents.defaults.restrict_to_workspace = false`
  - LogicIgniter and `/Users/aliai/projects` are in read/write path allowlists.
- Skill installation/search registries are disabled in config.
- DuckDuckGo web search is enabled and the other listed search providers are
  disabled.
- Current Discord global config is Ali-allowlisted and not mention-only.
- Current organization config no longer has 51 persistent app-owner agents.
- Current specialist worker prompt has the right active-PR review concept; the
  remaining gap is propagating that wording into cron summaries and AGENT boot
  files.
- Current approval matrix and GitHub control-plane docs correctly prefer normal
  PRs over draft PRs for Zehn-executable work.
- Current Yaad schema contract correctly uses
  `scope_type=organization`, `external_key=logicigniter`, and warns not to
  invent `company`, unsupported classes, or unsupported `binding_mode`.

## Recommended Remediation Order

1. Create or merge the standard `verify-pr.sh` path, or explicitly downgrade
   every prompt to the same documented fallback until it exists.
2. Decide and correct the bundle taxonomy: original 10 suites versus Ignite
   solution packaging.
3. Add/schedule missing operational queues: Zehn operations monitor, DevOps,
   QA, Security, and Docs, if Ali wants those roles autonomous.
4. Update current AGENT files and cron payloads to include active PR review,
   stale blocker cleanup, and post-merge handoff before `HEARTBEAT_OK`.
5. Mark setup/readiness/delegation/evidence files as historical and add a
   canonical current-state summary.
6. Fix software-delivery approval language so approved internal project work is
   not blocked while new repo/external/prod actions remain approval-gated.
7. Prove the post-merge reconciliation script from Zehn in a controlled
   non-destructive run before trusting it operationally.
8. Add a stale-phrase verification script so future prompt/memory cleanup is
   testable instead of manual.
