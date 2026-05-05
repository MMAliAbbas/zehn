# Agent Delegation And Meetings

This reference describes the supported private agent-to-agent delegation and
chaired meeting workflow.

## Runtime Boundaries

`delegate_to_agent` is the durable target-agent delegation path. It runs the
configured target `AgentInstance` with that agent's workspace, prompt files,
model, tools, memory, and private internal delegation session scope.

`spawn` and `subagent` remain ephemeral helper tools. They use subturns for
bounded helper work and must not be treated as durable target-agent delegation.

Discord is a human visibility layer only. Delegation and meetings do not depend
on Discord self-messages re-entering the bot; the Discord adapter continues to
ignore bot-authored inbound messages. Optional visibility summaries are concise
status updates, not raw internal debate.

## Delegation Workflow

Enable `delegate_to_agent` only for agents that should perform durable private
delegation. Delegation is constrained by `subagents.allow_agents`.

Supported delegation modes:

- `sync`: returns the target agent result to the caller and records the
  request/result locally.
- `async`: returns a delegation ID immediately, records the task locally, and
  lets `delegation_status` or `delegation_inbox` inspect progress.

Delegation records include parent agent, target agent, task, status, session
scope, result, local artifact references, optional GitHub artifact status, and
durable memory write status. Yaad writes are preferred when available; when Yaad
is unavailable and strict mode is not enabled, local records preserve the result
and record the Yaad failure.

GitHub issues are created only for executable or approval-tracked delegation
work, such as async execution or `approval_required` work. Advisory exchanges do
not create issues.

## Meeting Workflow

`start_agent_meeting` starts a private chaired meeting. The sponsor names a
chair and participants. The chair gathers private participant turns through
`delegate_to_agent`, then returns one consolidated recommendation.

Meeting records include:

- meeting ID, title, sponsor, chair, and participants
- goal, constraints, approvals, and artifact refs
- participant turns and chair synthesis
- consolidated recommendation
- timeline, risks, and follow-ups
- optional GitHub artifact status

The tool output includes the meeting ID, artifact refs, participants,
consolidated recommendation, timeline, risks, approval-needed text, and
follow-ups. Raw debate and internal prompts are kept out of GitHub issues and
Discord visibility summaries.

GitHub issues are created for meetings with executable follow-ups or explicit
approval tracking. Issue bodies contain curated meeting artifacts. Participant
comments are added only for material positions, risks, commitments,
dependencies, acceptance criteria, or follow-ups.

## Zehn Sales-Growth Pattern

A complete Zehn sales-growth workflow follows this shape:

1. Ali gives an objective to the CEO agent.
2. The CEO opens the executive objective and delegates the domain meeting to
   the responsible department head, such as CRO for sales growth.
3. The department head chairs the working meeting with relevant participants.
4. Participants provide private domain positions through target-agent
   delegation.
5. The chair returns one consolidated recommendation with participants,
   timeline, risks, approvals, and follow-ups.
6. The CEO reviews before asking Ali for approval when the plan has external,
   financial, customer-facing, production, legal, compliance, security, or
   irreversible effects.
7. Executable work is tracked in GitHub issues or project artifacts; durable
   memory remains in local records and the configured memory system.

The deterministic end-to-end verification for this workflow uses fake LLM
providers plus fake GitHub and Yaad adapters. It does not require live Discord,
GitHub, or Yaad.
