# LogicIgniter Heartbeat Operating Loop Plan - 2026-05-19

Status: implementation contract for the next Zehn operating artifact update.

Purpose: make heartbeat activate LogicIgniter as a functioning company instead
of passively checking stale local state, without making the system specific to
one repo, product, or app.

## Non-Negotiable Principles

- Heartbeat is the trigger, not the CEO.
- `zehn-main` owns Zehn runtime health and routing.
- `li-ceo` owns company priority, initiative direction, and executive
  delegation.
- `li-coo` owns execution flow, WIP control, stuck work, stale claims, and
  GitHub/project queue movement.
- `li-cto` owns technical delivery quality and engineering coherence.
- `li-cpo` owns product continuity, initiative shape, acceptance criteria, and
  successor-work discipline.
- GitHub issues, PRs, labels, and Projects are the live execution control
  plane.
- Yaad is durable memory, not the live work queue.
- The design must be LogicIgniter-wide and initiative-based. It must not name
  one app, one repo, or one product as the heartbeat's special case.
- `HEARTBEAT_OK` is valid only after the required operating checks complete
  successfully and no action is needed.

## Required Implementation Shape

This plan must implement the exact operating model Ali approved:

1. Define the company control model.
2. Rewrite heartbeat so `zehn-main` routes to CEO instead of acting as CEO.
3. Rewrite the CEO operating prompt.
4. Rewrite the COO work-selection prompt.
5. Update CTO/CPO/specialist prompts only where needed and without dumping
   repeated policy text.
6. Define one canonical GitHub work contract.
7. Define heartbeat acceptance criteria and verification scenarios.
8. Consider Go code only after prompt/config/artifact implementation proves the
   runtime cannot support the required behavior.

## Current Source Audit

Before implementation, the current source state shows:

- `HEARTBEAT.md` contains the core contradiction: it says to delegate to
  `li-coo` only when a concrete signal already exists, while the desired model
  requires a live company operating check to discover those signals.
- `LOGICIGNITER_OPERATING_CADENCE.md` already says heartbeat-triggered work
  selection is the continuous execution path and cron is not the issue-
  resolution path.
- `LOGICIGNITER_WORK_SELECTION.md` already defines the executable queue,
  filtering, ranking, claim, review, merge, and reconcile flow.
- `LOGICIGNITER_GITHUB_CONTROL_PLANE.md` already contains many work-contract
  details, but it is named as a control-plane artifact and not the single
  canonical contract requested here.
- Active `logicigniter-ceo-operating-check.md`,
  `logicigniter-coo-work-selection.md`, and
  `LOGICIGNITER_ACTIVE_INITIATIVES.md` are absent before this pass and need to
  be created.

## Files To Change

1. `.picoclaw/workspace/HEARTBEAT.md`
   - Replace the "concrete signal first" model with a company operating loop.
   - Route LogicIgniter operating checks to `li-ceo` every heartbeat cycle.
   - Keep Zehn runtime hygiene in `zehn-main`.
   - Forbid single-repo special casing.
   - Define when `HEARTBEAT_OK` is allowed.

2. `.picoclaw/workspace/memory/LOGICIGNITER_ACTIVE_INITIATIVES.md`
   - Create the canonical initiative registry if absent.
   - Define initiative fields and operating rules.
   - Include the required fields: initiative ID, title, priority, CEO owner,
     operating owner, product/technical owner, GitHub scope, active repos/
     projects, current state, WIP floor, last verified timestamp, next required
     decision, and escalation rules.
   - Keep entries company-level and initiative-level, not hardcoded to one
     implementation repo.

3. `.picoclaw/workspace/operating-prompts/logicigniter-ceo-operating-check.md`
   - Create active CEO heartbeat prompt if absent.
   - CEO reads the active initiative registry and Yaad first.
   - CEO delegates company operating control to COO and consults CTO/CPO/etc.
     only when relevant.
   - CEO must produce action, terminal routing, or a structured no-action
     report.

4. `.picoclaw/workspace/operating-prompts/logicigniter-coo-work-selection.md`
   - Create active COO heartbeat work-selection prompt if absent.
   - COO performs live organization-wide GitHub issue/PR/project inspection
     bounded by active initiatives and company queues.
   - COO may use bounded `gh` read/label/comment/project operations when MCP
     GitHub tools are unavailable and the task is company execution control.
   - COO dispatches role-matched specialists when claimable work exists.

5. `.picoclaw/workspace-li-ceo/AGENT.md`
   - Add a short pointer to the active initiative registry and CEO operating
     check prompt if missing.
   - Do not duplicate large policy text.

6. `.picoclaw/workspace-li-coo/AGENT.md`
   - Add a short pointer to the COO work-selection prompt and registry if
     missing.
   - Do not duplicate large policy text.

7. `.picoclaw/workspace-li-cto/AGENT.md`
   - Audit for alignment with the operating model.
   - Add only a short pointer if CTO does not already clearly own technical
     delivery, architecture, engineering quality, and specialist routing.

8. `.picoclaw/workspace-li-cpo/AGENT.md`
   - Audit for alignment with the operating model.
   - Add only a short pointer if CPO does not already clearly own product
     continuity, acceptance criteria, and successor-work discipline.

9. Specialist role files, only if audit proves a gap:
   - QA/Security/DevOps must own review and release gates.
   - Backend/Frontend/UX/Integration/Data/Architecture/Docs specialists must
     claim and execute role-matched issues.
   - Updates must be short pointers to the canonical work contract, not large
     repeated blocks.

10. `.picoclaw/workspace/memory/LOGICIGNITER_GITHUB_WORK_CONTRACT.md`
    - Create the canonical GitHub work contract requested by Ali.
    - It may reference `LOGICIGNITER_GITHUB_CONTROL_PLANE.md` and
      `LOGICIGNITER_WORK_SELECTION.md`, but it must clearly define:
      labels required for claimable work, specialist issue selection, claim
      flow, branch naming, PR title/body rules, review flow, merge/reconcile
      flow, successor issue rule, and dirty repo rule.

11. `.picoclaw/workspace/memory/LOGICIGNITER_HEARTBEAT_ACCEPTANCE_CRITERIA.md`
    - Create explicit test scenarios:
      ready issue exists with no active owner;
      PR green but unmerged;
      completed issue with no successor;
      GitHub unavailable;
      no work truly exists.

## Files Not To Change In This Pass

- Go source files.
- Config schema.
- Cron schedule.
- Channel configuration.
- Any LogicIgniter application repo.

## Expected Runtime Flow

1. Built-in heartbeat fires as `zehn-main`.
2. `zehn-main` performs Zehn runtime hygiene checks.
3. `zehn-main` delegates a bounded sync company operating check to `li-ceo`.
4. `li-ceo` reads `LOGICIGNITER_ACTIVE_INITIATIVES.md` and Yaad
   `organization:logicigniter`.
5. `li-ceo` determines whether company work needs COO execution control, CTO
   technical direction, CPO product continuity, or another role.
6. `li-ceo` delegates to `li-coo` for execution flow on active initiatives and
   company-wide queues.
7. `li-coo` inspects live GitHub/project state and dispatches role-matched
   specialists only when there is executable work.
8. Specialists claim/execute/review/PR/reconcile through
   `LOGICIGNITER_GITHUB_WORK_CONTRACT.md`,
   `LOGICIGNITER_WORK_SELECTION.md`, and
   `LOGICIGNITER_GITHUB_CONTROL_PLANE.md`.
9. If nothing needs action, the heartbeat may return `HEARTBEAT_OK` only after
   the above checks completed without tool/source failure.

## Implementation Order

1. Audit current heartbeat/CEO/COO/CTO/CPO/specialist prompts for
   contradictions.
2. Create active initiative registry.
3. Rewrite heartbeat to route to CEO.
4. Rewrite CEO operating check.
5. Rewrite COO work-selection.
6. Create canonical GitHub work contract.
7. Update CTO/CPO/specialist role docs only where the audit proves a gap.
8. Add heartbeat acceptance criteria.
9. Run static verification checks.
10. Observe one heartbeat cycle after Ali chooses to restart or wait for reload.
11. Review actual behavior before expanding.

## Verification

Run static checks after edits:

```bash
rg -n "Ignite Videoedit|apps-ignite-videoedit-studio|PR #55|PR #56|issue #57|issue #58|issue #59" \
  .picoclaw/workspace/HEARTBEAT.md \
  .picoclaw/workspace/memory/LOGICIGNITER_ACTIVE_INITIATIVES.md \
  .picoclaw/workspace/operating-prompts/logicigniter-ceo-operating-check.md \
  .picoclaw/workspace/operating-prompts/logicigniter-coo-work-selection.md
```

This must return no matches.

```bash
rg -n "li-ceo|li-coo|LOGICIGNITER_ACTIVE_INITIATIVES|HEARTBEAT_OK|organization:logicigniter" \
  .picoclaw/workspace/HEARTBEAT.md \
  .picoclaw/workspace/operating-prompts/logicigniter-ceo-operating-check.md \
  .picoclaw/workspace/operating-prompts/logicigniter-coo-work-selection.md
```

This must show the intended routing and memory references.

```bash
git diff -- .picoclaw/workspace/HEARTBEAT.md \
  .picoclaw/workspace/memory/LOGICIGNITER_ACTIVE_INITIATIVES.md \
  .picoclaw/workspace/memory/LOGICIGNITER_GITHUB_WORK_CONTRACT.md \
  .picoclaw/workspace/memory/LOGICIGNITER_HEARTBEAT_ACCEPTANCE_CRITERIA.md \
  .picoclaw/workspace/operating-prompts/logicigniter-ceo-operating-check.md \
  .picoclaw/workspace/operating-prompts/logicigniter-coo-work-selection.md \
  .picoclaw/workspace-li-ceo/AGENT.md \
  .picoclaw/workspace-li-coo/AGENT.md \
  .picoclaw/workspace-li-cto/AGENT.md \
  .picoclaw/workspace-li-cpo/AGENT.md
```

Review the diff manually for:

- no single-app special casing;
- no duplicate policy dump;
- no contradiction with `LOGICIGNITER_WORK_SELECTION.md`;
- no contradiction with `LOGICIGNITER_GITHUB_CONTROL_PLANE.md`;
- no instruction that makes heartbeat a shell worker;
- clear CEO/COO ownership split.
- CTO/CPO/specialist updates are short references, not repeated policy dumps.

## Restart Requirement

`AGENT.md`, `HEARTBEAT.md`, `USER.md`, `SOUL.md`, and workspace memory files
reload on the next request according to the local PicoClaw usage notes. A full
gateway restart is not required for prompt/memory changes, but the next
heartbeat cycle must be observed to verify behavior.
