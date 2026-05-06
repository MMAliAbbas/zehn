# Zehn Agent Organization UI Plan

Updated: 2026-05-07

## Goal

Add a launcher-visible organization view for configured agents that shows a
proper hierarchy, each agent's current operating state, and drill-down access to
inbox, outbox, meetings, and recent runtime events.

## Key Modeling Decision

Reporting structure must be modeled separately from delegation permission.

`subagents.allow_agents` answers: "which agents may this agent call?"

Organization hierarchy answers: "where does this agent sit in the operating
structure?"

Those are related but not identical. A department head may consult a peer, a
specialist, or a shared service without that peer reporting to them. Conversely,
an org chart child may not always be an allowed delegation target if the child
should only receive work through a different runtime path.

## Proposed Generic Config Model

Add an optional `agents.organization` section:

```json
{
  "agents": {
    "organization": {
      "roots": ["main"],
      "nodes": [
        {
          "agent_id": "main",
          "parent_agent_id": "",
          "label": "Main Coordinator",
          "group": "executive",
          "sort": 10
        }
      ]
    }
  }
}
```

Suggested fields:

- `roots`: ordered root agent IDs. If absent, derive roots from nodes without
  parents.
- `nodes[].agent_id`: configured agent ID.
- `nodes[].parent_agent_id`: optional reporting parent.
- `nodes[].label`: optional display override; fallback to agent name.
- `nodes[].group`: optional display grouping such as executive, department,
  bundle, app, personal, or support.
- `nodes[].sort`: optional stable sibling ordering.

Validation:

- Every node must reference an existing configured agent.
- Every parent must reference an existing configured agent or be empty.
- Duplicate node entries are invalid.
- Cycles are invalid.
- Missing organization config should not break existing installs.

Fallback behavior:

- If `agents.organization` is absent, the API should return a flat or lightly
  inferred view from `agents.list`.
- It should never infer reporting lines from `subagents.allow_agents` and label
  them as official hierarchy.

## Activity Model

Agent activity should come from structured runtime records first:

- Delegation records for inbox and outbox.
- Meeting records for meetings chaired by, sponsored by, or involving an agent.
- Config/registry data for registered versus configured state.

Logs should be secondary:

- Use logs for recent events and troubleshooting hints.
- Do not use log parsing as the only way to determine whether an agent is idle
  or working.

Suggested state precedence:

1. `failed`: most recent active/recent record failed.
2. `meeting`: active meeting chaired by or involving the agent.
3. `working`: active delegation targeted to the agent.
4. `delegating`: active delegation requested by the agent.
5. `idle`: configured and no active work.
6. `not_registered`: configured but missing from runtime registration, if the
   backend can observe that distinction.

## API Shape

Recommended endpoints:

- `GET /api/agents/organization`
- `GET /api/agents/{id}/activity`
- `GET /api/agents/{id}/inbox`
- `GET /api/agents/{id}/outbox`
- `GET /api/agents/{id}/meetings`

The first endpoint should be enough for the initial page. Drill-down endpoints
can support richer drawers/tabs without making the main payload too large.

## UI Shape

Add a launcher page under the Agent section:

- Sidebar item: Organization
- Route: `/agent/organization`
- Main page:
  - tree or grouped hierarchy
  - compact cards per agent
  - state badge
  - current activity summary
  - quick counts for inbox, outbox, meetings, and errors
- Agent detail drawer:
  - Overview
  - Inbox
  - Outbox
  - Meetings
  - Recent events

Keep the UI dense and operational. This is a command-center page, not a landing
page.

## Implementation Sequence

1. Add generic organization config structs and validation helpers.
2. Add backend snapshot/read model from config plus delegation/meeting stores.
3. Add inbox/outbox/meeting API drill-down endpoints.
4. Add frontend API client and organization page.
5. Add detail drawer with inbox/outbox/meeting tabs.
6. Add optional recent event enrichment from gateway logs.
7. Add end-to-end verification and operator documentation.

## Non-Goals For First Version

- Drag-and-drop org editing.
- Agent start/stop controls from org cards.
- Live bidirectional event stream.
- Company-specific hardcoding in source code.
- Treating delegation permissions as reporting lines.

