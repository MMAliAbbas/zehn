# Zehn Runtime Audit - 2026-05-11

Audit window: approximately `2026-05-10 23:27 PKT` through `2026-05-11 11:31 PKT`, with a second-pass spot check through approximately `2026-05-11 12:40 PKT`.

Purpose: determine why Zehn is not operating as expected as an autonomous personal and LogicIgniter company assistant. This report is evidence-first and does not propose or apply fixes.

## Sources Checked

- Zehn gateway logs:
  - `.picoclaw/logs/gateway.log`
  - `.picoclaw/logs/gateway_panic.log`
  - `.picoclaw/logs/launcher.log`
- Runtime config:
  - `.picoclaw/config.json`
  - `.picoclaw/workspace/cron/jobs.json`
- Current operating prompts:
  - `.picoclaw/workspace/operating-prompts/logicigniter-ceo-operating-check.md`
  - `.picoclaw/workspace/operating-prompts/logicigniter-engineering-check.md`
  - `.picoclaw/workspace/operating-prompts/personal-operating-check.md`
  - `.picoclaw/workspace/operating-prompts/logicigniter-specialist-work-check.md`
  - `.picoclaw/workspace/operating-prompts/zehn-operations-monitor.md`
  - Correction from the first pass: `.picoclaw/workspace/cron/messages/...` does not exist. The schedule and embedded dispatch messages live in `cron/jobs.json`; canonical prompt files live under `operating-prompts/`.
- Local memories modified during the audit window:
  - `.picoclaw/workspace/memory/MEMORY.md`
  - `.picoclaw/workspace/memory/LOGICIGNITER_GITHUB_CONTROL_PLANE.md`
  - `.picoclaw/workspace/memory/LOGICIGNITER_OPERATING_CADENCE.md`
  - `.picoclaw/workspace-li-ceo/memory/MEMORY.md`
  - `.picoclaw/workspace-li-cto/memory/MEMORY.md`
  - `.picoclaw/workspace-li-cto/memory/202605/20260511.md`
  - `.picoclaw/workspace-li-devops/memory/MEMORY.md`
  - `.picoclaw/workspace-li-integration-engineer/memory/MEMORY.md`
  - `.picoclaw/workspace-li-qa/memory/202605/20260511.md`
- GitHub state sampled with `gh` for:
  - `logicigniter/business`
  - `logicigniter/integration_tests`
  - `logicigniter/scripts`
  - `logicigniter/operations`
  - `logicigniter/svc-services-mcp`
- GitHub Project state sampled for organization project:
  - `logicigniter` project #1, `LogicIgniter Operating System`
- GitHub organization-wide issue search for items created inside the exact 12-hour audit window.
- Yaad read-only query through Zehn CLI, scope `organization:logicigniter`.

## Executive Summary

Zehn was busy during the audit window, but the activity was mostly inspection, delegation chatter, and failed or partial attempts. It did not behave like a reliable autonomous execution system.

The strongest failure pattern is not one single crash. It is a control-plane mismatch:

- Cron can schedule messages, but it does not natively target an agent by `agent_id`; it routes through the channel/default agent and relies on prompt-following delegation.
- The prompts tell agents to inspect GitHub and work on labeled issues, but the active issues are not labeled with the required workflow labels.
- The organization project exists, but it contains only four setup-era items while the active follow-up issues and PRs are not present there.
- Agents frequently claim `HEARTBEAT_OK` while logs show blockers, failed tool calls, stale memory, invalid Yaad writes, and GitHub command misuse.
- Several agents still reason from stale or contradictory memory, especially around `svc-services-mcp` and the MCP final-readiness blocker.
- Gateway health/readiness is inconsistent: `lsof` shows the gateway listening, one `/health` probe returned `200 OK`, `/ready` returned `503`, and repeated curl probes also failed to connect during the same audit pass.
- Delegation is active but not disciplined: self-delegation and duplicate specialist turns were observed.
- Some useful verification did happen, but it did not close the issue-claim-implementation-review-merge loop.

This means the system is running pieces of the intended operating model, but it is not yet closing the loop from issue discovery to claim, implementation, verification, PR review, and merge.

## Current Runtime State

Current gateway state is inconsistent from the shell probes:

```text
lsof -nP -iTCP:18790 -sTCP:LISTEN
picoclaw- 44779 ... TCP 127.0.0.1:18790 (LISTEN)

lsof -nP -iTCP:18800 -sTCP:LISTEN
picoclaw- 44778 ... TCP 127.0.0.1:18800 (LISTEN)
```

One health probe succeeded:

```text
HTTP/1.1 200 OK
{"status":"ok","uptime":"20h52m23.779483735s","pid":44779}
```

But repeated `curl` probes also failed with:

```text
curl: (7) Failed to connect to 127.0.0.1 port 18790
```

And `/ready` returned `503 Service Unavailable` at least once.

Conclusion: the process appears to be listening, but readiness is not healthy and direct probing was inconsistent. This should be treated as a live-operability finding, not proof of a fully healthy gateway.

## Runtime Config Facts

Relevant config facts verified from `.picoclaw/config.json`:

```text
heartbeat.enabled=true
heartbeat.interval=30
tools.cron.enabled=true
tools.exec.enabled=true
tools.exec.timeout_seconds=720
tools.exec.enable_deny_patterns=true
tools.exec.allow_remote=true
tools.cron.exec_timeout_minutes=12
tools.cron.allow_command=false
tools.mcp.enabled=true
tools.mcp.servers.yaad.enabled=true
find_skills disabled
install_skill disabled
```

Relevant path posture:

```text
allow_read_paths includes /Users/aliai/logicigniter, /Users/aliai/projects, and Zehn memory paths
allow_write_paths includes /Users/aliai/logicigniter and /Users/aliai/projects
```

This matches the intended "full authority on local LogicIgniter/project work" posture, but it does not by itself make agents use the correct repo, branch, issue, or verification workflow.

## Activity Observed

Gateway log turn-end counts for the audit window:

```text
li-architect             completed 18
li-backend-developer    completed 17
li-ceo                  completed 13
li-cto                  completed 36
li-data-ai-engineer     completed 17
li-devops               completed 23, error 1
li-frontend-developer   completed 20
li-integration-engineer completed 25
li-qa                   completed 16
li-ux-designer          completed 10
personal                completed 13
zehn-main               completed 139
```

There were `210` delegation files created or updated under `.picoclaw/workspace/delegations` during the audit window.

This proves the system was active. It does not prove the system was productive.

Delegation status aggregation:

```text
li-cto                  completed 36
li-integration-engineer completed 25
li-devops               completed 23
li-frontend-developer   completed 20
li-architect            completed 18
li-data-ai-engineer     completed 17
li-backend-developer    completed 17
li-qa                   completed 16
personal                completed 13
li-ceo                  completed 13
li-ux-designer          completed 11
li-integration-engineer running 1
li-cto                  running 1
li-devops               failed 1
```

There were also `92` specialist/session JSONL files modified during the audit window.

## End-to-End Automation Flow Recheck

This pass traced the intended autonomous path from scheduler entry point to execution outcome.

### 1. Launcher and Gateway

Observed process state:

```text
launcher: 127.0.0.1:18800 listening
gateway:  127.0.0.1:18790 listening
```

The gateway was active enough to run cron and delegation turns, but readiness was not clean because `/ready` returned `503` at least once and direct curl probes were inconsistent. This should not be interpreted as "gateway down"; it should be interpreted as "runtime active but not ready by its own contract."

### 2. Cron Entry Point

Cron jobs are configured in `.picoclaw/workspace/cron/jobs.json`. Verified active schedules:

```text
logicigniter-ceo-operating-check         5 * * * *      discord li_ceo channel
personal-operating-check                 9 * * * *      discord personal channel
logicigniter-engineering-check           */30 * * * *   discord engineering channel
logicigniter-architect-work-queue        12 * * * *     discord engineering channel
logicigniter-backend-work-queue          17 * * * *     discord engineering channel
logicigniter-frontend-work-queue         22 * * * *     discord engineering channel
logicigniter-ux-work-queue               27 * * * *     discord engineering channel
logicigniter-integration-work-queue      32 * * * *     discord engineering channel
logicigniter-data-ai-work-queue          37 * * * *     discord engineering channel
```

Each cron payload is an `agent_turn` message with `channel: discord` and a Discord `to` channel. There is no hard `agent_id` field in the cron payload.

Implication: cron is a message scheduler, not a direct worker scheduler. The cron message must be routed by channel/default-agent behavior, then the receiving agent must follow the embedded instruction and delegate.

### 3. Prompt Acquisition

Canonical prompts are under `.picoclaw/workspace/operating-prompts/`.

Evidence shows prompt reads were not uniformly reliable across all contexts:

```text
2026-05-09T23:05:10 read_file failed:
path escapes workspace: /Users/aliai/.picoclaw-zehn/workspace/operating-prompts/logicigniter-ceo-operating-check.md
```

Later internal/delegated turns show successful reads of files under the same prompt area. Therefore the correct finding is not "prompts are always inaccessible." The correct finding is:

```text
Prompt-file access has been context-sensitive. The embedded cron message is a critical fallback, and any drift between embedded message text and prompt file content can change behavior.
```

### 4. Delegation Routing

The main scheduling model is:

```text
cron -> Discord channel message -> default/main agent turn -> delegate_to_agent(target specialist) -> specialist work
```

This works often enough to create many delegation records, but it is not disciplined enough yet. Logs show duplicate and self-delegation patterns, including:

```text
scope_delegation="li-backend-developer:li-backend-developer:default"
scope_delegation="li-backend-developer:li-backend-developer:logicigniter-specialist-work-check-li-backend"
scope_delegation="personal:personal:scheduled-personal-operating-check"
```

Self-delegation can be valid for some wrapper patterns, but the backend and frontend specialist traces show it also functions as repeated work expansion. It wastes turns and can hide whether the original scheduled intent actually completed.

### 5. GitHub Queue Discovery

The specialist prompt requires specialists to claim only issues with:

```text
zehn:ready
matching area:* label
not claimed/in-progress/blocked/approval-gated
enough scope, acceptance criteria, verification, risk, and review detail
```

Current sampled issue state does not satisfy this. Example verified query:

```text
gh issue list -R logicigniter/business --state open --label zehn:ready
[]
```

Implication: specialists can follow the prompt correctly and still find no claimable work.

### 6. GitHub Project Control Plane

The organization project exists:

```text
logicigniter project #1: LogicIgniter Operating System
items.totalCount: 4
fields.totalCount: 18
```

Fields include `Status`, `Department`, `Bundle`, `Risk`, `App`, `Priority`, `Approval Required`, `Owner Agent`, `Target Date`, and `Linked pull requests`.

But the project has only four items. The currently active follow-up issues and PRs observed in `business`, `integration_tests`, `scripts`, `operations`, and `supervision` are not represented as a live project board. This makes the GitHub Project a weak operating surface even though the field model exists.

Additional nuance: the Project `Bundle` field still uses the original suite names. That may be intentional historical naming, but it does not line up cleanly with newer Ignite bundle display names and can confuse reporting.

### 7. Execution and Verification

The intended flow is documented as:

```text
issue -> claim labels/comment -> issue branch -> dedicated Codex execution -> verification script -> commit -> push -> internal review -> normal PR -> @codex review -> merge only after approval signal
```

Observed reality:

- No claimable `zehn:ready` issue queue was found in sampled repos.
- No full specialist claim-to-PR cycle was observed during the exact audit window.
- Some verification did run on existing PRs. A delegation summary reported standard verification passed for:
  - `business` PR #54 with evidence under `/Users/aliai/logicigniter/var/verification/issue-53/business-20260510T203058Z`
  - `scripts` PR #4 with evidence under `/Users/aliai/logicigniter/var/verification/issue-3/scripts-20260510T203059Z`
  - `integration_tests` PR #12 with evidence under `/Users/aliai/logicigniter/var/verification/issue-10/integration_tests-20260510T203100Z`
- Those PRs still did not become merge-ready because the review/check approval gate remained unmet.

### 8. Review and Merge Gate

The workflow says Codex `eyes` means review started and post-review thumbs-up/formal approval means approval. Observed state included `COMMENTED`, no GitHub checks, and no clear thumbs-up/formal approval signal.

Implication: the merge gate is policy-defined, but not yet machine-closed. Agents can identify that nothing is merge-ready, but they are not progressing PRs to an approved state.

### 9. Memory and Reporting

Yaad is reachable and some writes succeed, but agents still invent invalid schema values:

```text
invalid memory_class "operating_state"
invalid memory_class "event"
invalid binding_mode "shared"
```

Later retries often succeed when the agent falls back to valid classes such as `fact` or `note`.

Implication: durable memory is available, but schema discipline is not yet reliable. A failed memory write can make an agent believe it recorded state when it did not.

## Major Failures

### 1. Gateway Health/Readiness Is Inconsistent

Severity: blocking for live operation.

Evidence:

```text
lsof shows PID 44779 listening on 127.0.0.1:18790
one /health probe returned 200 OK
/ready returned 503 Service Unavailable
later curl probes returned connection failures
```

Impact:

- The gateway may be alive enough to run cron/delegations but not ready by its own readiness contract.
- Health/ready status is not currently giving a simple reliable operator answer.
- Any "lastStatus ok" in `jobs.json` is historical job status, not proof of full runtime readiness.

### 2. Cron Does Not Natively Target Specialist Agents

Severity: high.

Evidence from runtime code inspection:

- `pkg/cron/service.go` `CronPayload` contains `kind`, `message`, `command`, `channel`, and `to`.
- It does not contain `agent_id`.
- Cron job execution routes messages by channel/chat, then the gateway routes to the default agent unless channel routing resolves otherwise.

Evidence from prompts:

- Specialist cron prompts ask `zehn-main` to delegate to a specialist.
- That means specialist execution depends on the model choosing the right `delegate_to_agent` call.

Impact:

- Scheduled specialist work is prompt-mediated, not a hard routing guarantee.
- A weak or distracted main-agent turn can return `HEARTBEAT_OK` without the intended specialist ever doing useful work.

### 3. `HEARTBEAT_OK` Is Masking Problems

Severity: high.

Evidence:

- `zehn-main Response: HEARTBEAT_OK` appeared `89` times during the audit window.
- The same window includes:
  - `213` exec safety-guard blocks.
  - Yaad write failures due invalid `memory_class` and `binding_mode`.
  - GitHub CLI misuse.
  - Missing file/path errors.
  - Internal channel outbound warnings.
  - Final-readiness blocker drift.

Prompt contradiction:

- The specialist prompt says `HEARTBEAT_OK` should only be returned when queue inspection succeeds and there is no matching claimable issue, blocker, stale claim, dirty repo, or needed triage.
- The logs show repeated failures and blockers existing in the same window.

Impact:

- The status signal is not trustworthy.
- Discord reports may look calm while execution is failing underneath.

### 4. GitHub Work Queue Is Not Actually Claimable By the Specialist Rules

Severity: high.

Evidence:

- Workflow labels exist on sampled repos, including `zehn:ready`, `zehn:claimed`, `zehn:in-progress`, `zehn:blocked`, and area labels.
- Current sampled open issues are unlabeled.

Examples:

```text
logicigniter/business
- #53 Codify Zehn GitHub issue execution policy: no labels
- #51 Track current-root MCP final-readiness reconciliation: no labels
- #49 Define LogicIgniter operating system v1: no labels

logicigniter/integration_tests
- #10 Add integration impact verification wrapper: no labels
- #9 Align MCP persona login identifiers with local Keycloak seed: no labels
- #7 Fix MCP final-readiness identity/BFF auth parity failures: no labels

logicigniter/scripts
- #3 Add standard PR verification entrypoint: no labels
- #1 Preflight local MCP runtime proof prerequisites: no labels

logicigniter/operations
- #3 Harden final-readiness ledger validation for task 026: no labels
```

The specialist prompt tells agents to claim only issues with `zehn:ready` plus matching `area:*`.

Impact:

- The queue can appear empty to every specialist even when real work exists.
- This directly explains why scheduled agents may inspect, then do nothing.

### 5. No New GitHub Issues Were Created During the Exact 12-Hour Audit Window

Severity: high.

Evidence:

Organization-wide GitHub issue search:

```text
gh search issues --owner logicigniter --created '>2026-05-10T18:27:00Z'
[]
```

Impact:

- After the system was expected to operate autonomously, it did not create new executable issue artifacts in the 12-hour window.
- Existing issue and PR activity around 2026-05-10 17:35-18:35 UTC appears to predate the audit window.

### 6. GitHub Project Exists But Is Not the Live Operating Board Yet

Severity: high.

Evidence:

```text
gh project list --owner logicigniter --limit 20 --format json
project #1: LogicIgniter Operating System
items.totalCount: 4
fields.totalCount: 18
```

Project fields are well-shaped for an operating control plane:

```text
Status
Department
Bundle
Risk
App
Priority
Approval Required
Owner Agent
Target Date
Linked pull requests
```

But the project has only four items, all from the setup period:

```text
business issue #49
business PR #50
operations issue #1
operations PR #2
```

Current active issues and PRs such as `business` #51/#53/#54, `integration_tests` #7/#9/#10/#11/#12, `scripts` #1/#2/#3/#4, `operations` #3/#4, and `supervision` #1/#2 are not represented as live project items in the sampled Project state.

Impact:

- The GitHub Project looks structurally ready but operationally stale.
- Agents that rely on project state will not see the live work queue.
- Agents that rely on raw issues will bypass the project fields and lose status/owner/priority/approval metadata.

### 7. Exec Safety Guard Blocked Generated Commands 213 Times

Severity: high.

Evidence:

The gateway log contained `213` instances of:

```text
Command blocked by safety guard (dangerous pattern detected)
```

Example pattern:

- An agent tried a multi-line shell command with `set -u`, arrays, substitutions, or compound shell syntax.
- The guard blocked it.
- The agent sometimes retried with a simpler command.

Impact:

- Agents are not reliably adapting their command style to Zehn's safety guard.
- Time is wasted on blocked commands.
- Some tasks fail before reaching real execution.

### 8. Internal Channel Outbound Warnings Are Extremely Frequent

Severity: medium-high.

Evidence:

The gateway log contained `2047` warnings:

```text
Unknown channel for outbound message
```

Impact:

- Internal delegation may be producing outbound messages that cannot be delivered to a real channel.
- This may not always break tool execution, but it creates noisy telemetry and can hide real delivery problems.

### 9. Yaad Writes Are Not Schema-Safe

Severity: high.

Evidence from gateway logs:

Invalid memory classes:

```text
invalid memory_class "episodic"
invalid memory_class "event"
invalid memory_class "operating_state"
invalid memory_class "semantic"
invalid memory_class "operating_status"
invalid memory_class "operational_finding"
```

Invalid binding modes:

```text
invalid binding_mode "primary"
invalid binding_mode "explicit"
invalid binding_mode "strong"
```

Impact:

- Agents are inventing Yaad schema values instead of using only valid Yaad scope and memory fields.
- Top-tier agents may think they saved durable facts when the write actually failed.

Counter-evidence:

- Later in the log, `li-integration-engineer` successfully wrote a Yaad memory with `memory_class: "fact"` and scope `organization:logicigniter`.

Conclusion:

- Yaad itself is reachable.
- The failure is inconsistent agent usage of the Yaad schema, not a total Yaad outage.

### 10. Stale Local Memory Is Contradicting Current Reality

Severity: high.

Evidence:

The CEO memory still contains stale statements that `svc-services-mcp` was missing or that MCP final-readiness was blocked by missing repo evidence.

Current evidence contradicts that:

- `/Users/aliai/logicigniter/svc-services-mcp` exists locally.
- GitHub repo `logicigniter/svc-services-mcp` exists.
- `gh issue list` and `gh pr list` for `logicigniter/svc-services-mcp` both succeed, with no issues or PRs.
- CTO memory later says `svc-services-mcp main 6bc3c94` and host-native endpoints are reachable.

Impact:

- CEO-level planning can continue to repeat an obsolete blocker.
- Agents are not reliably invalidating stale memory when the world changes.

Yaad evidence confirms the same problem exists in durable memory, not only local files:

```text
bafe4973-3a3f-453e-80e8-6e11ee5e2b4f
Current-root MCP repo absence blocks LogicIgniter final launch-readiness claim
Mentions svc-services-mcp missing, final-readiness MCP failure, and invalid_credentials.

eced46b0-dc19-43cf-8a0c-b4f6586d2431
LogicIgniter CEO operating check — current-root MCP blocker persists
Mentions svc-services-mcp missing and current-root MCP/final-readiness blocker.
```

Yaad also contains newer contradictory memories:

```text
ccfa4a08-9954-41a2-8c0f-bfe2c8cd730f
LogicIgniter CEO Operating Check — 2026-05-11 08:08
Mentions svc-services-mcp reachable but final-readiness MCP still blocked by auth/persona issues.

504f64b5-00fd-4336-ba8f-63bedfdfef4c
LogicIgniter CEO Operating Check — 2026-05-11 11:05
Mentions svc-services-mcp reachable and latest final-readiness MCP failure remaining.
```

Conclusion:

- Active Yaad memory contains both obsolete and current facts.
- Agents need a way to supersede or demote stale memories; otherwise retrieval may surface the wrong blocker.

### 11. Final-Readiness Blocker Has Drifted Without a Clean Canonical State

Severity: high.

Observed blocker sequence:

1. Missing `svc-services-mcp` repo.
2. Missing or renamed final-readiness docs path.
3. Persona login identifier mismatch.
4. Keycloak seed / integration credential mismatch.
5. Current QA memory says BFF auth parity issue remains:
   - local identity issues JWT with issuer `https://identity.logicigniter.com/realms/logicigniter`
   - BFF middleware expects `http://localhost:8180/realms/logicigniter`

Impact:

- Agents are chasing different versions of "the blocker."
- The canonical current blocker is not cleanly promoted above stale blockers.

### 12. Agents Are Confused About Git Working Directories

Severity: medium-high.

Evidence:

Errors include:

```text
fatal: not a git repository
failed to run git: not a git repository
/Users/aliai/logicigniter is not a git repo
/Users/aliai/.picoclaw-zehn/workspace-li-qa is not a git repo
/Users/aliai/.picoclaw-zehn/workspace-li-backend-developer is not a git repo
```

Impact:

- Some agents run `git` inside their Zehn workspace or the aggregate `/Users/aliai/logicigniter` directory instead of a specific repo.
- This prevents reliable branch, diff, status, verification, and PR workflows.

### 13. GitHub CLI Usage Is Error-Prone

Severity: medium-high.

Evidence:

Errors include:

```text
gh pr diff: unknown flag --stat
gh pr list: unknown JSON field "checks"
expected OWNER/REPO, got /Users/aliai/logicigniter/scripts
GraphQL error for aliai/logicigniter-scripts
gh issue/pr command accepted at most 1 arg, received 4
```

Impact:

- Agents are mixing local paths, wrong owner names, and unsupported gh flags.
- This prevents reliable issue/PR automation.

### 14. Local Repos Are Clean But Some Are Left on Work Branches

Severity: medium.

Evidence from sampled local repo status:

```text
business          chore/53-zehn-github-execution-policy
integration_tests chore/10-integration-impact-verification
operations        chore/3-final-readiness-ledger-validation
scripts           chore/3-standard-pr-verification
supervision       chore/1-final-readiness-ledger-current-root-caveat
.github           chore/1-zehn-executable-work-templates
```

Most service repos are on `main`.

Impact:

- "Clean" is not the same as "on main and ready for the next autonomous task."
- Agents may inspect or build from feature branches and misunderstand current production truth.

### 15. Codex Provider Emits Empty-Output Reconstruction Warnings Constantly

Severity: medium.

Evidence:

`2194` warnings:

```text
Codex completed response had empty output; reconstructed output from streamed output_item.done events
```

Impact:

- The provider often recovers, but the system is relying on reconstruction.
- This increases risk of empty, partial, or confusing final messages.

### 16. Some Agent Turns Are Too Long for Responsive Operations

Severity: medium.

Evidence:

Maximum observed turn durations:

```text
li-qa       max 824243 ms
li-cto      max 463949 ms
zehn-main   max 484595 ms
```

There were `16` proactive compression warnings:

```text
Proactive compression: context budget exceeded before LLM call
```

Impact:

- Long turns reduce operational responsiveness.
- Context compression may cause loss of important instructions in long-running autonomous operation.

### 17. Discord Delivery Had Transient Failures

Severity: low-medium.

Evidence:

Two Discord send failures:

```text
temporary failure
```

Impact:

- Not the main failure mode, but it means Discord cannot be treated as a perfect audit sink.

### 18. `svc-services-mcp` Exists But Has No GitHub Work Items

Severity: medium.

Evidence:

```text
gh issue list -R logicigniter/svc-services-mcp
[]

gh pr list -R logicigniter/svc-services-mcp
[]
```

Impact:

- The repo exists, but the autonomous work queue has not been populated there.
- If MCP is central to final readiness, there is no direct issue/PR control plane in that repo yet.

### 19. Bundle Agent IDs Use Historical Handles While Display Names Use Newer Bundle Names

Severity: medium-low as a direct runtime defect; medium as an operating UX/control-plane risk.

Evidence from runtime registration:

```text
agent_id=li-bundle-saas-growth-and-retention-suite
name="LogicIgniter Bundle Owner: Ignite Messaging"

agent_id=li-bundle-e-commerce-operations-suite
name="LogicIgniter Bundle Owner: Ignite Workflow Webhooks"

agent_id=li-bundle-finance-and-revenue-intelligence-suite
name="LogicIgniter Bundle Owner: Ignite Commerce Ops"
```

Correction from the first audit pass:

- This should not be treated as a proven routing bug by itself.
- Some agent files document old IDs as stable internal handles while newer names are product-facing bundle identities.

Remaining impact:

- The underlying `agent_id` still encodes the old 10-suite names, while the display name reflects newer bundle naming.
- The GitHub Project `Bundle` field still uses the original suite names.
- This can confuse human review, labels, memories, and project reporting even if runtime routing is technically stable.
- The safer wording is: this is a documented handle/product-name mismatch that needs explicit mapping, not proof that delegation is broken.

### 20. Personal Assistant Cadence Has No Useful Inputs Yet

Severity: medium.

Evidence from the `personal` cron response:

```text
Personal operating check:
- No tracked personal projects, personal reminders, or explicitly allowed pending personal...
```

Impact:

- The personal agent is running, but it has no structured personal backlog or reminder source to act on.
- It cannot become useful without either a personal task source, approved memory scope, or explicit personal operating artifacts.

### 21. Active Cron Jobs Overlap Long Turns

Severity: medium.

Evidence from `gateway_panic.log`:

- Engineering checks are scheduled every 30 minutes.
- Some engineering checks took:
  - `485054ms` at `08:00`
  - `365656ms` at `10:00`
  - `324413ms` at `11:00`
- The next engineering run can start while prior delegated follow-up work is still active or shortly after it completes.

Impact:

- Zehn can stack expensive checks and specialist follow-ups.
- This increases context pressure, internal-channel warnings, and the chance of stale or duplicate investigations.

### 22. Some Delegations Remain Running

Severity: medium.

Evidence:

Delegation aggregation found:

```text
li-integration-engineer running 1
li-cto                  running 1
li-devops               failed 1
```

Impact:

- The system had not reached a fully settled state at audit time.
- Open/running delegation files can confuse later status summaries if they are not finalized or timed out cleanly.

### 23. Duplicate and Self-Delegation Waste Turns and Obscure Completion

Severity: high.

Evidence:

Logs show specialist agents delegating to themselves, including backend and personal examples:

```text
scope_delegation="li-backend-developer:li-backend-developer:default"
scope_delegation="li-backend-developer:li-backend-developer:logicigniter-specialist-work-check-li-backend"
scope_delegation="personal:personal:scheduled-personal-operating-check"
```

Impact:

- Some scheduled work becomes "agent asks itself to do the same work" instead of a clean worker turn.
- This consumes extra model/tool turns.
- It makes status harder to interpret because both the parent and child self-turn may return `HEARTBEAT_OK`.
- It increases the chance of duplicate queue scans and stale conclusions.

### 24. Prompt Reads Are Context-Sensitive, So Embedded Cron Messages Can Drift From Canonical Prompts

Severity: high.

Evidence:

At least one scheduled context failed to read the canonical prompt:

```text
path escapes workspace: /Users/aliai/.picoclaw-zehn/workspace/operating-prompts/logicigniter-ceo-operating-check.md
```

Later delegated/internal contexts successfully read prompt and memory files under Zehn workspaces.

Impact:

- This is not a simple missing-file problem; the files exist.
- Behavior depends on tool file-access context.
- If the cron embedded message and prompt file diverge, the agent may execute stale instructions even though the prompt file looks correct.

### 25. Verification Work Happened, But It Is Not Yet Connected to a Full Autonomous Merge Loop

Severity: high.

Evidence:

A delegation summary reported standard verification passes:

```text
business #54 passed
scripts #4 passed
integration_tests #12 passed
```

Evidence paths were reported under:

```text
/Users/aliai/logicigniter/var/verification/issue-53/business-20260510T203058Z
/Users/aliai/logicigniter/var/verification/issue-3/scripts-20260510T203059Z
/Users/aliai/logicigniter/var/verification/issue-10/integration_tests-20260510T203100Z
```

But the same summaries say these PRs are not merge-ready because Codex review state is `COMMENTED`, checks are absent, or the approval signal is missing.

Impact:

- The first-pass statement "no productive work" was too broad.
- Corrected statement: Zehn performed useful inspection and verification work, but did not complete the autonomous delivery loop.

### 26. Merge Approval Signal Is Policy-Defined But Not Machine-Closed

Severity: high.

Evidence:

The workflow requires a post-review Codex thumbs-up or formal approval. Observed PR review state includes `COMMENTED`, and agents explicitly reported no merge-ready PR.

Impact:

- Agents can identify that PRs are not approved.
- There is no verified end-to-end mechanism proving an agent can detect the approved signal, merge safely, close the issue, and update project/memory state.

### 27. Agents Create Temporary Local Scripts to Work Around Exec Guard Blocks

Severity: medium.

Evidence:

`li-frontend-developer` created:

```text
/Users/aliai/.picoclaw-zehn/workspace-li-frontend-developer/tmp_frontend_queue_check.sh
```

The temp script was created after direct shell patterns with arrays/loops were blocked by the safety guard.

Impact:

- This is understandable adaptation, but it creates local workspace artifacts.
- Those artifacts can become stale or misleading if not cleaned or clearly marked temporary.
- It also confirms agents do not yet have a standard safe command style for repeated repo/issue scans.

### 28. `max_tool_iterations` Was Seen Historically, But Is Not Reproduced After Current Config Tuning

Severity: medium.

Evidence:

The log contains a historical CEO check memory note about:

```text
delegate_to_agent(li-ceo, sync) returned: max_tool_iterations without final response
```

Current config shows:

```text
max_tool_iterations: 50
```

The current log pass did not find a fresh `I've reached max_tool_iterations` runtime failure after that tuning.

Impact:

- This should remain a watch item, not a current proven blocker.
- If long operating checks keep expanding through self-delegation or broad scans, the limit may still be hit again.

### 29. Current Canonical Final-Readiness Blocker Is Persona/Auth Parity, Not Repo Absence

Severity: high.

Evidence:

Latest inspected state says:

```text
svc-services-mcp exists locally and on GitHub
host-native endpoints are reachable in at least some probes
final-readiness-20260510T160629Z still has Failures: 1, Skips: 1
Mira invalid_credentials
Diego empty JWT
```

Impact:

- Stale "repo missing" and "Docker missing" memories must not drive current planning as the main blocker.
- The current blocker should be framed as final-readiness persona/auth/JWT parity in the host-native runtime path.
- Any next automation should work from that canonical state unless newer evidence supersedes it.

## Contradictions Found

### Contradiction A: `svc-services-mcp` Missing vs Present

Stale local memory and active Yaad memory say the repo was missing or unavailable.

Current evidence:

- Local repo exists.
- GitHub repo exists.
- CTO memory says host-native MCP endpoint is healthy.

Conclusion: stale memory remains active and is likely polluting agent decisions.

### Contradiction B: `HEARTBEAT_OK` vs Active Failures

Agents frequently report `HEARTBEAT_OK`, but logs in the same window show:

- blocked commands
- invalid Yaad writes
- GitHub command errors
- stale blockers
- missing path errors
- no newly created issues inside the audit window

Conclusion: `HEARTBEAT_OK` is not a reliable business or execution health signal.

### Contradiction C: "GitHub Execution Not Optional" vs Unlabeled Issues

Prompts require agents to use GitHub issues and labels.

Current open issues are mostly unlabeled, so specialists cannot claim them under their own rules.

Conclusion: the process exists on paper but the current GitHub queue is not prepared for the process.

### Contradiction D: "Full Authority" vs Safety Guard Blocks

The user allowed broad authority, and config permits broad read/write paths and long exec timeouts.

But the exec safety guard still blocked `213` generated commands.

Conclusion: authority is broad at the path/config level, but generated command style is still incompatible with Zehn guardrails.

### Contradiction E: "Repo Clean" vs "Repo Ready"

Some memories report repos clean.

Current status shows several repos are clean but still on feature branches.

Conclusion: agents need to distinguish clean working tree, branch posture, PR posture, and merge readiness.

### Contradiction F: Bundle Handles vs Product Names

Runtime agent IDs still refer to the original SaaS suite names, while runtime display names use newer "Ignite ..." bundle names.

Correction: this is not proof of broken runtime routing because the old IDs may be intentionally stable handles. It remains a control-plane ambiguity because GitHub Project bundle fields, human-facing docs, memories, and Discord reporting can refer to either vocabulary.

### Contradiction G: Yaad Is Canonical vs Local/Stale Memory Still Drives Decisions

The operating model says Yaad is canonical durable memory.

But active Yaad contains both stale and current versions of the same blocker, while local memory also contains stale notes.

Conclusion: "Yaad is canonical" is not enough unless memories have lifecycle semantics such as superseded, stale, current, or deprecated.

## What Zehn Actually Accomplished

Verified accomplishments during or immediately around the audit period:

- Cron scheduling fired repeatedly.
- Delegation machinery was active.
- Specialist agents were invoked many times.
- GitHub workflow labels exist in sampled repos.
- The organization project exists with useful fields for `Status`, `Department`, `Bundle`, `Risk`, `App`, `Priority`, `Approval Required`, `Owner Agent`, and linked PRs.
- Several PRs and issues exist from the setup period:
  - `business` PR #54, issue #53.
  - `integration_tests` PRs #8, #11, #12 and issues #7, #9, #10.
  - `scripts` PRs #2, #4 and issues #1, #3.
  - `operations` PR #4 and issue #3.
- Standard verification was reportedly rerun successfully for existing PRs:
  - `business` PR #54.
  - `scripts` PR #4.
  - `integration_tests` PR #12.
- Local memories captured important current blockers, especially the BFF issuer mismatch.
- The system identified at least one likely real technical blocker:
  - issuer mismatch between identity JWT issuer and BFF middleware expected issuer.
- It successfully wrote at least one later Yaad fact memory using the valid `memory_class: "fact"` shape.

## What Zehn Did Not Accomplish

- It did not create any new GitHub issue during the exact 12-hour audit window.
- It did not close the loop from issue selection to claim, branch, implementation, verification, PR, review, and merge.
- It did not make the GitHub Project a live operating board for current work.
- It did not populate sampled current issues with the labels required by the specialist work-queue prompt.
- It did not maintain a clean canonical current blocker state.
- It did not consistently save durable Yaad memories because schema values were invalid.
- It did not keep `/ready` healthy.
- It did not make specialist work queues claimable because open issues lack required labels.
- It did not resolve stale Yaad memories that directly contradict newer facts.

## Yaad Read-Only Findings

Read-only Yaad query under `organization:logicigniter` returned these relevant active memories:

```text
bafe4973-3a3f-453e-80e8-6e11ee5e2b4f
Current-root MCP repo absence blocks LogicIgniter final launch-readiness claim

eced46b0-dc19-43cf-8a0c-b4f6586d2431
LogicIgniter CEO operating check — current-root MCP blocker persists

ccfa4a08-9954-41a2-8c0f-bfe2c8cd730f
LogicIgniter CEO Operating Check — 2026-05-11 08:08

504f64b5-00fd-4336-ba8f-63bedfdfef4c
LogicIgniter CEO Operating Check — 2026-05-11 11:05

853ddc71-58a7-4e53-b8ed-251273ad05b3
LogicIgniter CEO Operating Check — 2026-05-10 20:06

5969aa92-59cb-4195-b8f7-0b0fce709c1b
LogicIgniter CEO Operating Check — 2026-05-11 10:06

d9e8d2df-fe9f-43ca-b4e9-562e77b4542e
LogicIgniter CEO Operating Check — 2026-05-10 22:05

bc7bf0c5-704d-4761-8263-828b33c2a532
LogicIgniter CEO Operating Check — 2026-05-10 16:05 MCP runtime blocker

c275f0bf-384a-45ea-8a9f-c5fb1f28c09c
Current-root final readiness MCP runtime proof blocked by missing Docker

9210fd68-185e-40d7-98c7-f04ed10dcae6
2026-05-10 Final readiness MCP persona auth parity finding

3f3a741b-5d8e-440b-a45d-f18c6b0e77f4
LogicIgniter CEO Operating Check — 2026-05-11 07:05

892d8b9c-0095-4743-964a-49e11da1b4c3
LogicIgniter engineering operating check 2026-05-10 11:30 current-root final gate blockers

0e99212b-4f91-46b7-9988-cb0755865666
LogicIgniter CEO Operating Check — 2026-05-11 05:05

69fdefca-7621-4847-a05c-bc5ebbe42911
LogicIgniter engineering operating check 2026-05-10: repo evidence update

cd493160-cc3a-4dfd-b9c0-7750e1e42ceb
Zehn executable GitHub work policy

a99f0cc9-d5cf-463c-9186-023c94ceef1a
Zehn specialist GitHub issue routing policy
```

Interpretation:

- Yaad contains the desired GitHub execution and specialist routing policies.
- Yaad also contains active stale memories about `svc-services-mcp` being missing.
- Newer memories corrected the repo status but did not supersede or deactivate the stale memories.

## Audit Conclusion

The system is not failing because "nothing is running." It is failing because the operating loop is not yet coherent enough:

- scheduler activity exists;
- delegation activity exists;
- memory writes partly work;
- GitHub labels exist;
- a GitHub Project exists;
- standard verification can run on existing PRs;
- repos are mostly clean;
- but the queue is not claimable, the Project is not live for current work, memory is contradictory, ready is unhealthy, command generation trips safety guards, self-delegation creates duplicate work, and agents often report `HEARTBEAT_OK` instead of escalating operational failure.

This should be treated as a control-plane and operating-contract failure, not simply a prompt wording problem. The next fixes should start at the operating flow boundaries: canonical issue/project readiness, hard evidence of claim-to-PR execution, memory schema discipline, and prevention of duplicate/self-delegation loops.
