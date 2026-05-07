# Agent Delegation And Meetings

This reference describes the supported private agent-to-agent delegation,
meeting v1 chaired sequential workflow, and the planned meeting v2 debate
workflow.

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

## Agent Organization Metadata

Configured agents may optionally include `agents.organization` metadata for
displaying a reporting hierarchy. This section is separate from delegation
permissions:

- `subagents.allow_agents` defines which target agents an agent may call.
- `agents.organization` defines where agents sit in a reporting or operating
  structure.

Do not infer official reporting lines from `subagents.allow_agents`. An agent
may be allowed to consult peers, specialists, or shared services that do not
report to it. Likewise, a reporting child may require a different runtime path
or approval flow before it can receive delegated work.

The organization section is optional. Configurations without it continue to load
using the normal `agents.list` behavior. When present, it supports ordered roots
and node metadata:

```json
{
  "agents": {
    "list": [
      { "id": "main", "name": "Main Coordinator" },
      { "id": "operations", "name": "Operations" }
    ],
    "organization": {
      "roots": ["main"],
      "nodes": [
        {
          "agent_id": "main",
          "label": "Main Coordinator",
          "group": "executive",
          "sort": 10
        },
        {
          "agent_id": "operations",
          "parent_agent_id": "main",
          "label": "Operations",
          "group": "department",
          "sort": 20
        }
      ]
    }
  }
}
```

Fields:

- `roots`: optional ordered root agent IDs. If omitted, roots are derived from
  nodes without `parent_agent_id`.
- `nodes[].agent_id`: configured agent ID for this organization node.
- `nodes[].parent_agent_id`: optional reporting parent agent ID.
- `nodes[].label`: optional display label override.
- `nodes[].group`: optional display grouping for UIs.
- `nodes[].sort`: optional stable numeric ordering within siblings and derived
  roots. Ties are ordered by `agent_id`.

Validation rejects duplicate nodes, unknown root agents, unknown node agents,
unknown parents, and reporting cycles. Organization metadata does not enable
delegation by itself and does not change spawn, subagent, routing, session, or
channel behavior.

Organization activity status is derived from structured delegation and meeting
records. Launcher gateway logs may be included as bounded recent-event
enrichment for drill-down troubleshooting, but log-derived events are secondary
evidence only: they do not override agent status, counts, current activity, or
failure precedence. The API only exposes redacted, truncated summaries from
known structured log fields such as `agent_id`, `target_agent_id`,
`parent_agent_id`, `chair_agent_id`, or `sponsor_agent_id`; malformed or
unrelated log lines are ignored.

The launcher organization page is read-only. It uses `GET
/api/agents/organization` for the initial hierarchy, `GET
/api/agents/{id}/activity` for a single agent's page-equivalent status and
counters, and the `inbox`, `outbox`, and `meetings` drill-down endpoints for
record lists. Loading these endpoints must not mutate launcher config, local
records, durable memory, channel state, or external artifacts. See [Agent
Organization Live Verification](agent-organization-live-verification.md) for
operator checks and badge meanings.

## Meeting V1 Workflow

`start_agent_meeting` starts meeting v1: a private chaired sequential meeting.
The sponsor names a chair and participants. The chair gathers private
participant turns through `delegate_to_agent` one participant at a time, then
returns one consolidated recommendation.

Meeting v1 is not live real-time debate. Participants do not see each other's
turns while those turns are being collected, and the current implementation does
not run multi-round back-and-forth discussion. The chair sees the collected
participant positions during synthesis and owns the final recommendation.

Meeting v1 uses a conservative required-participant failure policy:

- Every participant listed in `participant_agent_ids` is required.
- There is no implicit optional participant or partial-completion mode in v1.
- A participant provider, tool, delegation, or permission failure records a
  failed participant turn, marks the meeting failed, and stops before chair
  synthesis.
- A chair provider, tool, delegation, or internal synthesis failure marks the
  meeting failed after preserving completed participant turns.
- Context cancellation marks an already-created meeting record cancelled; if
  cancellation happens before the record is created, the call returns the
  cancellation error without a meeting record.
- Record-store failures are returned to the caller. When a meeting record
  already exists, PicoClaw attempts to mark it failed or cancelled before
  returning the store error.

Meeting records include:

- meeting ID, title, sponsor, chair, and participants
- goal, constraints, approvals, and artifact refs
- participant turns and chair synthesis
- failed participant turns when required participant collection fails
- consolidated recommendation
- timeline, risks, and follow-ups
- optional GitHub artifact status

The tool output includes the meeting ID, artifact refs, participants,
consolidated recommendation, timeline, risks, approval-needed text, and
follow-ups. Raw internal prompts and participant turn text are kept out of
GitHub issues and Discord visibility summaries by default.

GitHub issues are created for meetings with executable follow-ups or explicit
approval tracking. Issue bodies contain curated meeting artifacts. Participant
comments are added only for material positions, risks, commitments,
dependencies, acceptance criteria, or follow-ups.

## Meeting V2 Debate Design

Meeting v2 should add an explicit multi-round debate loop without changing the
meaning of meeting v1. It should be a new execution mode or versioned tool path
so existing `start_agent_meeting` callers keep the chaired sequential behavior.

Turn order:

- The chair opens with the goal, constraints, decision rules, participant list,
  maximum rounds, and per-turn budget.
- Round 1 collects initial positions in deterministic participant order from
  the request after normalization and de-duplication.
- Later rounds use the same order unless the chair explicitly inserts a focused
  intervention, such as asking Finance to respond to a margin risk.
- The chair synthesis turn happens after the debate loop stops and remains the
  single default user-facing recommendation.

Participant visibility:

- In round 1, participants receive only the meeting context, their role, and
  chair instructions.
- In later rounds, each participant receives a curated transcript summary of
  prior turns, including speaker, round, position, risks, objections,
  commitments, and open questions.
- Participants should not receive hidden chain-of-thought, raw provider
  messages, secrets, or unredacted private data.
- The chair can redact or summarize sensitive prior material before it becomes
  visible to later participants.

Chair interventions:

- The chair may add a short intervention between turns or rounds to narrow the
  question, resolve ambiguity, call for evidence, ask for dissent, or stop a
  tangent.
- Interventions are recorded as chair turns with round number, reason, and
  target audience.
- Chair interventions must not silently rewrite participant positions; they can
  summarize, challenge, or request a response.

Stopping criteria:

- Stop when the chair determines consensus or a clear recommendation exists.
- Stop when the configured maximum round count is reached.
- Stop when no participant adds material new risks, alternatives, commitments,
  or objections in a full round.
- Stop immediately on cancellation, policy violation, missing required approval
  boundary, or repeated participant failure.
- If unresolved dissent remains, the chair must include it in risks or
  follow-ups instead of extending indefinitely.

Token limits:

- Each participant turn gets an explicit maximum response budget.
- The rolling transcript is summarized before it exceeds the configured meeting
  context budget.
- Summaries must preserve decisions, disagreements, assumptions, risks,
  approvals, dependencies, and follow-ups.
- If summarization cannot fit the budget, the chair stops the debate and emits
  a recommendation with a token-limit risk.

Failure handling:

- A participant failure records a failed turn with error text redacted through
  the existing record filter.
- The chair may continue if the failed participant is optional and the remaining
  quorum is enough for the decision.
- Required participant failure stops the meeting and records a failed status.
- Repeated provider/tool failures stop the debate, preserve all completed
  turns, and publish only a concise blocker summary to visibility channels.
- Partial results must never be presented as complete consensus.

Audit trail:

- Meeting v2 records need version, mode, round count, turn order, chair
  interventions, participant-visible summaries, raw private turn references,
  stop reason, token-budget decisions, failures, chair synthesis, final
  recommendation, timeline, risks, approvals, follow-ups, artifact refs, and
  optional GitHub artifact status.
- GitHub and Discord continue to receive curated summaries only.
- Durable memory should store final decisions and curated summaries, not raw
  transcript text by default.

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
