# Zehn Autonomy Control-Plane Repair Plan - 2026-05-11

Purpose: define the next Zehn repair work from evidence, not optimism. This plan uses the runtime audit and performance review to repair the autonomous operating loop so Zehn itself can create, prepare, claim, execute, verify, review, and report work without Ali manually shepherding every step.

This is a planning document only. No runtime/config/code changes are included here.

## Evidence Base

Primary sources:

- `.picoclaw/workspace/reports/ZEHN_RUNTIME_AUDIT_20260511.md`
- `.picoclaw/workspace/reports/ZEHN_IMPLEMENTATION_PERFORMANCE_REVIEW_20260511.md`
- `.picoclaw/workspace/cron/jobs.json`
- `.picoclaw/workspace/operating-prompts/*.md`
- `.picoclaw/workspace/memory/LOGICIGNITER_GITHUB_CONTROL_PLANE.md`
- `.picoclaw/workspace/memory/LOGICIGNITER_OPERATING_CADENCE.md`
- `.picoclaw/workspace/memory/LOGICIGNITER_REPO_ACCESS_DOCTRINE.md`
- `.picoclaw/workspace/memory/LOGICIGNITER_ENGINEERING_QUALITY_DOCTRINE.md`
- `.picoclaw/config.json`
- gateway logs under `.picoclaw/logs/`
- sampled GitHub state for `logicigniter/business`

Verified facts:

- Cron jobs exist and fire.
- Cron payloads are `agent_turn` messages with `channel` and `to`; there is no native cron `agent_id`.
- Discord dispatch rules route channels to agents. For example, the engineering Discord channel routes to `li-engineering`.
- Specialist cron jobs currently deliver to the engineering channel, so the first receiving agent is the routed channel agent, not necessarily the intended specialist.
- The current specialist prompt says the current agent should delegate to the target specialist.
- Logs show self-delegation patterns such as `li-backend-developer:li-backend-developer`.
- Labels exist in sampled repos, but active issues are not labeled with `zehn:ready` and `area:*`.
- The organization Project exists with useful fields, but it contains only old setup-era items.
- Yaad is reachable, but agents still invent invalid `memory_class` and `binding_mode` values.
- Some verification ran on existing PRs, but no full issue-claim-implementation-review-merge loop completed.

## Core Diagnosis

The failure is not that Zehn lacks instructions. The failure is that the instructions do not form a reliable control plane.

The current setup has four disconnected layers:

1. Operating intent in memories and prompts.
2. Cron jobs that send Discord-channel messages.
3. GitHub issue/project policy documents.
4. Specialist agents that can act if a valid issue reaches them.

The missing substrate is the bridge between these layers:

- no reliable queue-reconciliation loop;
- no direct issue/project preparation duty;
- no hard separation between scheduler/router prompts and worker prompts;
- no verified machine-readable issue contract;
- no reliable durable memory schema guidance;
- no runtime evidence gate that says an operating cycle truly succeeded.

## Design Principle

Zehn should not depend on Ali to prepare the board.

If the queue is empty or malformed, the responsible Zehn role must do one of these:

1. create a correct executable issue;
2. fix labels/project metadata on an existing issue;
3. mark the issue blocked with a specific reason;
4. request Ali approval only when the work crosses an approval boundary.

`HEARTBEAT_OK` is valid only after the agent proves that no safe useful action exists.

## Target Operating Loop

The repaired loop should be:

```text
CEO/CTO/Product/Ops discovers priority
  -> creates or selects GitHub issue
  -> reconciler ensures issue labels/project fields/body contract
  -> specialist queue finds matching issue
  -> specialist claims with lease comment
  -> issue-linked branch created
  -> dedicated execution session performs work
  -> verification runs and records evidence
  -> repo ends clean
  -> normal PR opens
  -> internal reviewers inspect
  -> Codex GitHub review requested
  -> approval signal detected
  -> merge only if policy permits
  -> issue/project/Yaad/Discord updated
```

## Repair Workstreams

### Workstream 1: Split Scheduler Prompts From Worker Prompts

Problem:

The current specialist prompt begins with scheduler/router behavior:

```text
Delegate this check to the target specialist named in the scheduler message...
```

If the target specialist receives that same text, it can delegate to itself. The audit found self-delegation in logs.

Required change:

- Keep a scheduler/router prompt for routed channel agents.
- Create a separate specialist worker prompt that contains only worker instructions.
- Cron messages should tell the channel-routed agent:
  - target specialist ID;
  - target labels;
  - send the worker prompt content/task to that specialist;
  - do not ask the target to delegate again.
- Specialist worker instructions must explicitly say:
  - if you are the named target specialist, do not delegate this same queue check to yourself;
  - only delegate to another role for review or a genuinely different specialty.

Acceptance:

- Logs for one backend or frontend work-queue run show one parent delegation into the specialist, not a self-delegation chain.
- If no issue is claimable, the target specialist returns a status explaining queue state, not a self-delegated `HEARTBEAT_OK`.

### Workstream 2: Add a GitHub Control-Plane Reconciler Duty

Problem:

The GitHub Project and labels exist, but current work is not prepared for autonomous specialists. Active issues are unlabeled and not represented in the Project.

Required change:

- Define a recurrent `li-operations` or `li-coo` control-plane reconciler duty.
- The reconciler should inspect key repos and classify open issues/PRs.
- For issues that are clearly Zehn-executable and inside standing authority, it should:
  - add `zehn:ready`;
  - add one or more `area:*` labels;
  - add `risk:*`;
  - add approval label if needed;
  - add/update Project item fields;
  - comment with missing acceptance criteria if body is incomplete.
- For issues that are not executable, it should leave them unready and document what is missing.

Important:

The reconciler must not convert every issue to `zehn:ready`. It must enforce readiness, not bypass it.

Acceptance:

- At least one existing suitable issue becomes claimable by a specialist without Ali manually editing labels.
- Project item count increases to include active executable issues.
- A non-executable issue is left unready with a clear blocker comment or report.

### Workstream 3: Define a Machine-Readable Executable Issue Contract

Problem:

Prompts require issue body fields, but there is no strongly reusable template or checker that agents can use before claiming.

Required change:

- Define the required issue body sections:
  - Goal
  - Repo
  - Owner Agent
  - Area Labels
  - Risk
  - Approval Required
  - Scope
  - Non-Goals
  - Acceptance Criteria
  - Verification Command
  - Sensitive Areas
  - Review Requirements
  - Dirty Repo Rule
- Add issue template guidance where appropriate.
- Add a checklist that the reconciler and specialists both use.

Acceptance:

- A specialist can reject an issue as "not claimable" with the exact missing fields.
- A reconciler can prepare an issue to claimable state without inventing missing technical detail beyond its authority.

### Workstream 4: Turn Empty Queue Into Action, Not `HEARTBEAT_OK`

Problem:

Specialists can inspect an empty or malformed queue and return `HEARTBEAT_OK`, even when the real problem is missing labels/project metadata or missing issue preparation.

Required change:

- CEO/CTO/Engineering prompt must treat "no claimable issues" as a control-plane signal:
  - If real work exists but is unlabeled, delegate reconciler work.
  - If no issue exists for a known blocker, create one or delegate issue creation.
  - If a specialist cannot find work, it reports queue health and recommends/requests a reconciler action.
- Specialist worker prompt must distinguish:
  - no matching issue exists;
  - matching issues exist but are not claimable;
  - GitHub inspection failed;
  - repo is dirty;
  - approval boundary blocks execution.

Acceptance:

- A run with unlabeled active issues returns "queue malformed" plus a concrete action, not `HEARTBEAT_OK`.
- A run with no issues and no known blockers returns `HEARTBEAT_OK` only after checking GitHub successfully.

### Workstream 5: Standardize Safe Shell Patterns For Agents

Problem:

Logs show many exec safety-guard blocks. Agents use arrays, command substitution, compound shell, and complex loops that hit deny patterns.

Required change:

- Add safe command examples to operating prompts and specialist worker prompt.
- Prefer explicit single commands:
  - `gh issue list -R logicigniter/business --state open --label zehn:ready --label area:backend --json number,title,labels,url`
  - `git -C /Users/aliai/logicigniter/business status --short --branch`
- For multi-repo scanning, either:
  - run separate simple commands; or
  - use a trusted checked-in helper script after it is created through issue/PR flow.
- Do not let agents create ad hoc temp scripts in workspaces unless the output path and cleanup rule are explicit.

Acceptance:

- A specialist queue run completes without safety-guard blocks.
- If a safety guard blocks a command, the response reports the exact blocked command class and the simpler retry used.

### Workstream 6: Add Yaad Schema Discipline To Prompts And Memory

Problem:

Agents invent invalid Yaad values such as `memory_class: event`, `operating_state`, `project_note`, and binding modes such as `shared`.

Required change:

- Add a short Yaad schema contract to shared memory and active prompts:
  - approved scope: `{"scope_type":"organization","external_key":"logicigniter"}`
  - approved memory classes for Zehn operating work: `fact`, `decision`, `summary`, `note`, `runbook`, `best_practice`, `anti_pattern`, `architecture_decision`
  - omit `binding_mode` unless the Yaad tool schema explicitly requires it;
  - call `scope_type_list` when uncertain;
  - if write fails, retry once with a valid class, then report failure.
- Add stale-memory rule:
  - when a current fact supersedes a previous blocker, write a new memory that names the superseded blocker and says it is stale.

Acceptance:

- Next CEO/CTO/specialist durable memory write uses a valid memory class.
- Failed Yaad writes are reported and retried safely instead of silently treated as success.

### Workstream 7: Add Review/Merge State Closure

Problem:

Verification can run, but PRs remain stuck because Codex review/check/approval state is not machine-closed.

Required change:

- Define the exact `gh` queries agents should use to inspect:
  - PR draft state;
  - checks;
  - review decision;
  - review comments;
  - reactions or formal approval signal;
  - labels such as `zehn:merge-ready`.
- Define who adds `zehn:merge-ready` and when.
- Define what agents do when Codex only `COMMENTED`.

Acceptance:

- A PR can be classified as:
  - blocked by checks;
  - blocked by internal review;
  - blocked by Codex comments;
  - blocked by missing approval signal;
  - merge-ready;
  - approval-gated for Ali.

### Workstream 8: Add Runtime Success Criteria And Observation Checklist

Problem:

Past work judged progress by activity. The new standard must judge progress by completed state transitions.

Required change:

Create a runtime observation checklist:

```text
cron fired
correct routed agent received message
router delegated once to correct target
target did not self-delegate
GitHub queue inspected
issue created/prepared or valid no-work conclusion reached
issue claimed if claimable
repo status checked before mutation
branch created before mutation
verification ran
repo clean at end
PR opened or blocker reported
Yaad write succeeded or failure reported
Discord report included meaningful status
```

Acceptance:

- Every autonomy repair run must be judged against this checklist.

## Implementation Sequence

Do not implement all workstreams at once.

### Step 1: Planning And Artifacts Only

Create or update docs/prompts only:

- split scheduler and worker prompt design;
- define executable issue contract;
- define reconciler responsibilities;
- define Yaad schema contract;
- define observation checklist.

No Go code yet.

### Step 2: Prompt And Memory Repair

Apply only prompt/memory changes:

- CEO prompt;
- CTO/Engineering prompt;
- specialist scheduler prompt;
- new specialist worker prompt;
- GitHub control-plane memory;
- operating cadence memory;
- Yaad schema memory.

Restart required after prompt/config changes only if the gateway caches prompt content or cron embedded messages are changed. Prompt files read at runtime may not require restart, but `cron/jobs.json` embedded message changes do.

### Step 3: GitHub Control-Plane Repair

Using GitHub tools, let Zehn-relevant control plane be prepared:

- ensure labels exist across target repos;
- add issue templates where appropriate;
- add current active issues to Project;
- label only issues that meet the executable contract;
- mark incomplete issues as not ready with missing fields.

This should be done by an agent-controlled task only after Step 2 prompts can describe the work correctly.

### Step 4: One Autonomous Cycle Observation

Let one scheduled control-plane/reconciler cycle and one specialist cycle run.

Do not judge by whether an issue was manually completed. Judge by the checklist in Workstream 8.

### Step 5: Runtime Code Only If Needed

Only consider Go code if evidence proves prompt/config/workflow cannot solve a boundary.

Candidate code changes, if needed later:

- cron payload support for `agent_id`;
- first-class internal scheduled agent turn, bypassing Discord routing;
- self-delegation guard for `delegate_to_agent`;
- structured queue status tool;
- Project item reconciliation helper.

These must be separate upstream-aware tasks with tests.

## What Not To Do

- Do not create another manual proving issue as the main solution.
- Do not add more agents.
- Do not add UI features before the control loop works.
- Do not broadly edit Go code before prompt/config/GitHub workflow boundaries are tested.
- Do not claim autonomy from cron activity alone.
- Do not accept `HEARTBEAT_OK` when a tool failed, queue is malformed, memory write failed, repo is dirty, or a known blocker exists.

## Recommended First Concrete Repair

Start with Workstream 1 and Workstream 4 together:

1. Split the specialist prompt into scheduler/router and worker prompts.
2. Update scheduler messages so target specialists receive worker instructions, not another scheduler prompt.
3. Make empty/malformed queues produce reconciler action instead of `HEARTBEAT_OK`.

Reason:

This addresses the most direct runtime failure: specialist schedules currently create activity but can collapse into self-delegation and passive status. Without this fix, GitHub project/label cleanup will still not produce reliable specialist execution.

## Readiness To Implement

Ready to implement only after Ali approves this plan or asks for changes.

Implementation should be done as small, reviewable steps with verification after each step.
