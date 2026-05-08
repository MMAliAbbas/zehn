# Zehn Agent Organization Command Center Plan

## Goal

Turn the launcher Organization page from a mostly static agent hierarchy into a
live operating console for configured agents.

The page should help an operator answer:

- Which agents are active right now?
- Who is delegating to whom?
- Which inbox, outbox, meeting, or failure records need attention?
- What is happening in the gateway logs now?
- What changed recently for the selected agent or for the org overall?

## Current State

- The backend exposes a read-only organization snapshot at
  `/api/agents/organization`.
- The backend exposes per-agent inbox, outbox, and meeting record list APIs.
- The frontend renders an organization page with cards, summary metrics, and a
  details sheet.
- The gateway log page already uses incremental polling through
  `/api/gateway/logs`.
- Agent recent events are currently a parsed snapshot from gateway logs, not an
  interactive live log view.

## Target Experience

### Command Header

The top of the page should show a compact live status bar:

- active/running work count
- delegations
- meetings
- failures
- live/refresh state
- latest update timestamp

### Organization Canvas

Cards should remain visible as the main hierarchy, but become more interactive:

- clicking a card selects an agent
- selected agent has a clear active visual state
- inbox/outbox/meetings/errors count pills act as shortcuts
- cards show concise current status and last activity

### Persistent Workbench

On desktop, selected agent details should appear in a persistent side workbench
instead of only a temporary sheet. On mobile, the existing sheet pattern can
remain.

Workbench sections:

- Overview
- Inbox
- Outbox
- Meetings
- Failures
- Recent events
- Live logs

### Live Activity

Use existing incremental gateway log polling first. Avoid adding WebSocket/SSE
until polling proves insufficient.

The live log panel should:

- follow the current gateway run id and offset
- auto-scroll only when the user is near the bottom
- preserve readable mono formatting
- filter or highlight selected-agent references where practical
- fail quietly with a visible stale/error state instead of breaking the page

### Failure Clarity

Old failures should not be visually indistinguishable from active failures.
Failure drilldown should show:

- failed record id
- peer agent
- status
- created/updated/completed timestamps
- whether newer activity exists
- artifact references, if present

## Implementation Principles

- Keep the first version read-only.
- Prefer existing APIs and polling before adding new backend surfaces.
- Reuse existing log polling and record list components where possible.
- Keep UI dense and operational, not decorative.
- Do not mutate config, records, memory, channels, or external artifacts.
- Preserve mobile usability with the existing sheet pattern.
- Add tests for new state/model helpers and any backend endpoints added.

## Task Sequence

1. Workbench state and selection model.
2. Clickable activity shortcuts.
3. Desktop persistent workbench with mobile sheet fallback.
4. Live gateway log panel embedded in organization workbench.
5. Agent-scoped log filtering/highlighting.
6. Org-wide activity feed.
7. Failure drilldown and stale failure clarity.
8. Command header live status improvements.
9. Final verification and operator docs.
