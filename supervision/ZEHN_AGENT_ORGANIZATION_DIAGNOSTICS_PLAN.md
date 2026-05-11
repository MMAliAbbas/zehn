# Zehn Agent Organization Diagnostics Plan

## Goal

Make the Organization page useful as an operator command center by explaining
why agents are marked failed, what record caused the status, whether the
failure is current or historical, and where to inspect supporting evidence.

This plan is intentionally read-only for the runtime. It must not mutate agent
records, config, memory, channels, GitHub artifacts, or external systems.

## Source Facts

- The Organization snapshot is served by `GET /api/agents/organization`.
- Per-agent record lists are served by:
  - `GET /api/agents/{id}/inbox`
  - `GET /api/agents/{id}/outbox`
  - `GET /api/agents/{id}/meetings`
  - `GET /api/agents/{id}/failures`
- The frontend already has a persistent workbench with Overview, Inbox,
  Outbox, Meetings, Failures, Recent Events, and Live Logs sections.
- Failure records shown by the UI currently include record identity, type,
  status, peer, role, timestamps, and artifact references.
- Delegation records already contain richer source data:
  - request task and metadata
  - result content/status
  - error message/type
  - durable-memory write status/error
  - GitHub artifact write status/error
- Meeting records already contain richer source data:
  - title, goal, constraints, notes
  - participant turns
  - chair turn
  - recommendation/timeline/risks/approvals/follow-ups
  - error
  - GitHub artifact write status/error

## Problem

The UI can show that an agent failed, but it often cannot explain why. That
forces the operator to leave the screen, inspect JSON files or logs manually,
and guess whether a failure is still actionable.

The missing operator answers are:

- What failed?
- Why did it fail?
- Was this the agent's current blocker or an old historical failure?
- Was the failure caused by a delegation turn, meeting turn, durable memory
  write, artifact publishing, capacity, permission, config, or another source?
- What is the next inspection target?
- Which supporting record, artifact, or log line should be opened next?

## Design Principles

- Use existing persisted records as the source of truth.
- Add read-only derived diagnostics; do not alter record storage semantics.
- Keep summaries short, sanitized, and bounded.
- Do not expose full prompt/task/result bodies in compact list views.
- Provide explicit detail views for deeper inspection.
- Keep failure visibility fail-closed: unknown or unreadable records should not
  be shown as healthy.
- Preserve existing routes and response fields for compatibility.
- Add tests before relying on behavior in the UI.
- Keep changes small enough for task-loop review and rollback.

## Diagnostic Model

Add derived fields to organization activity summaries rather than changing the
core delegation or meeting record schema first.

Recommended compact fields:

- `summary`: short human-readable record summary.
- `reason`: short failure or status reason when known.
- `reason_source`: where the reason came from, such as `record_error`,
  `memory_error`, `artifact_error`, `participant_turn`, or `status`.
- `severity`: derived operator severity such as `info`, `warning`, or
  `error`.
- `current`: whether this record is currently driving the agent card status.
- `stale`: whether this is historical and newer activity exists.
- `detail_available`: whether a safe detail endpoint can provide more context.

The compact model should be safe for cards, feeds, and record lists.

## Detail Model

Add a safe read-only detail shape for a selected record. The detail view can
include more context than list rows, but it must still avoid dumping large raw
task/result bodies by default.

Recommended detail sections:

- identity: record id, type, status, role, peer, created/updated/completed
- reason: source, type, message, severity
- request/context summary: bounded and sanitized
- result summary: bounded and sanitized
- memory/artifact status: provider, status, error, updated time
- meeting participant status, if applicable
- artifact references
- suggested inspection target, such as record JSON, related artifact, or logs

## UI Target

Cards should show a concise current-state line. If status is failed, the card
should show a short reason instead of only "failed".

The Failures tab should show:

- current vs historical badge
- reason
- source
- peer/role
- timestamps
- artifact count
- detail action

The detail pane should show the expanded diagnostic record without requiring
the operator to inspect JSON manually.

The Live Logs section should help inspect selected failures by filtering or
highlighting the selected agent and record id.

## Task Sequence

1. Backend diagnostic summary model.
2. Backend safe record detail endpoint.
3. Frontend failure reason rendering.
4. Frontend record detail drilldown.
5. Live-log correlation for selected diagnostic records.
6. Final diagnostics verification and documentation.

## Safety Gates

Every task must keep the runtime read-only and pass targeted backend/frontend
verification before being marked green.

No task should:

- restart the gateway or launcher
- mutate config
- mutate runtime records
- add external side effects
- expose secrets or large prompt/result bodies
- require operator downtime

