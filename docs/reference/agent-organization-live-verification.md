# Agent Organization Live Verification

This guide verifies the launcher Agent > Organization page and its read-only
activity APIs against a running launcher and gateway.

## Scope

The organization page reads these launcher endpoints:

- `GET /api/agents/organization`
- `GET /api/agents/{id}/activity`
- `GET /api/agents/{id}/inbox`
- `GET /api/agents/{id}/outbox`
- `GET /api/agents/{id}/meetings`
- `GET /api/agents/{id}/failures`

These endpoints must not write config, delegation records, meeting records,
memory, channel state, GitHub artifacts, Discord messages, or other external
artifacts. They only read launcher config, local delegation and meeting record
directories, and bounded gateway log lines for best-effort recent-event
enrichment.

## Refresh Behavior

The Agent > Organization page is a near-live operational dashboard based on
conservative polling, not a guaranteed realtime event stream. While the page is
mounted, the organization snapshot refreshes every 15 seconds. When an agent
detail drawer is open, the inbox, outbox, meetings, and failures drill-downs
also refresh every 15 seconds, but only for the currently visible tab. Closed
drawers and hidden drill-down tabs do not continue polling their activity
endpoints.

Background refreshes keep the last visible content on screen. Initial loading
and full-page error states are reserved for the first request before any usable
snapshot or drill-down data has been loaded.

## Command Center Behavior

The Organization page is a read-only command center. Operators can select
agents, change workbench tabs, filter visible logs, and drill into record
summaries, but the page must not start, stop, retry, edit, delete, or publish
delegations, meetings, gateway state, memory, configuration, GitHub artifacts,
or Discord messages.

On desktop, selecting an agent card keeps the organization canvas visible and
opens the persistent Agent Workbench. On mobile, selecting a card opens the
detail sheet. The selected card is visually marked and exposes `aria-pressed`.
The Details button always opens Overview. Count pills act as shortcuts:

- Inbox opens the Inbox workbench tab or detail tab.
- Outbox opens the Outbox workbench tab or detail tab.
- Meetings opens the Meetings workbench tab or detail tab.
- Errors opens the Failures workbench tab or detail tab.

The workbench tabs are Overview, Inbox, Outbox, Meetings, Failures, Recent
Events, and Live Logs. Inbox, Outbox, Meetings, and Failures poll only while
their tab is visible. Overview, Recent Events, and Live Logs use the
already-loaded organization snapshot or the shared gateway log polling state.

The command header summarizes the organization snapshot: active work,
delegations, meetings, failures, hierarchy or flat mode, generated time,
refreshed time, and query state. A background refresh error after data has
loaded should show a stale query state without clearing the visible command
center.

Command header totals are scoped to the configured organization agents in the
loaded launcher config. Delegation and meeting records count only when at least
one requester, target, sponsor, chair, or participant is still a configured
agent. Historical records that reference only removed or renamed agent IDs are
omitted from organization totals and the global Recent Activity feed. If a
record has both configured and unknown agent IDs, the record can still count,
but the Recent Activity entry must point at a configured agent so the entry
opens a visible card and workbench tab.

The Recent Activity feed shows the newest organization-level structured
activity and gateway events, capped to the newest entries. Delegations open the
selected agent's Inbox, meetings open Meetings, failures open Failures, and
gateway events open Recent Events. Failed delegations and failed meetings both
appear as failure feed entries.

The Live Logs tab uses incremental gateway log polling. `All Logs` shows the
current browser-retained gateway lines. The browser buffer is capped to the
newest 2,000 lines so a long-running Organization page cannot grow memory
without bound. Incremental polling still follows the gateway-reported total log
offset after older visible lines are discarded, so operators should use the
gateway's source logs or captured artifacts when reviewing lines older than the
live buffer. `Selected Agent` shows only lines with explicit agent reference
fields, including `agent_id`, `target_agent_id`,
`parent_agent_id`, `requester_id`, `sponsor_agent_id`, `chair_agent_id`,
`child_agent_id`, `route_agent_id`, and `scope_agent_id`. Arbitrary message
text, partial substrings, tokens, and sensitive-looking fields must not count
as selected-agent references. When a detail record is selected, `Selected
Record` further narrows or highlights lines that reference the selected record
id or known peer agent IDs from the selected record row. It does not match the
selected agent by itself. If no retained live log lines match, the panel shows
a record-specific empty state without clearing the all-logs view.

The Failures tab fetches the selected agent's recent visible failed delegation
and meeting records. The current and last-failure summaries remain fallback
context while records are loading or unavailable, but the fetched recent failure
list is authoritative once loaded. If newer activity exists, the tab labels
older failures as historical and keeps the newer current activity visible in
Overview and Recent Events. Failure drilldown is limited to record type, record
id, peer agent, role, status, created, updated, completed, and artifact
references.

## Badge Meanings

- `Idle`: the agent is configured and has no active or failed structured
  activity selected as current. A completed record can still appear as the
  current activity while the badge is idle, but it must not replace any active
  visible work or meeting for the same agent.
- `Working`: the current selected record is an active delegation targeting this
  agent.
- `Delegating`: the current selected record is an active delegation requested
  by this agent.
- `Meeting`: the current selected record is a started meeting where this agent
  is sponsor, chair, or participant.
- `Failed`: the current selected record is the newest failed delegation or
  meeting involving this agent and no newer active or completed record has
  superseded it.

Current activity selection gives active visible work and active meetings
precedence over completed records before comparing timestamps. This keeps an
agent in `Working`, `Delegating`, or `Meeting` while an older active record is
still open, even if a newer completed record exists for the same agent. For all
other structured records, selection uses record `updated_at` first, then a
stable tie-break: failed, meeting, working, delegating, completed, then idle.
Failed records are retained separately through `last_failure` and error
counters, so an old failure remains visible without permanently masking newer
operational work. A newer failure can still become current when no newer active
or completed record has superseded it.

Recent gateway events can appear in the detail drawer as secondary evidence,
but they do not change badge status, counters, current activity, or
structured-record selection.

## Preparation

1. Start the launcher with the config under test.
2. Start the gateway from the launcher and wait until it reports running.
3. Open the launcher UI and navigate to Agent > Organization.
4. Identify one configured agent ID for config-only checks and the agent IDs
   used for delegation and meeting checks.

Set these shell variables for API checks:

```bash
export LAUNCHER_URL="http://127.0.0.1:8080"
export AGENT_ID="main"
```

If dashboard authentication is enabled, include the same cookie or auth header
that the browser uses for launcher API requests.

## Staged Checks

### 1. Config-Only State

Use a config with agents but no local delegation or meeting records. If the
record directories do not exist, leave them absent for this check.

```bash
curl -fsS "$LAUNCHER_URL/api/agents/organization"
curl -fsS "$LAUNCHER_URL/api/agents/$AGENT_ID/activity"
curl -fsS "$LAUNCHER_URL/api/agents/$AGENT_ID/inbox"
curl -fsS "$LAUNCHER_URL/api/agents/$AGENT_ID/outbox"
curl -fsS "$LAUNCHER_URL/api/agents/$AGENT_ID/meetings"
curl -fsS "$LAUNCHER_URL/api/agents/$AGENT_ID/failures"
```

Expected result:

- The page loads in flat mode when `agents.organization` is absent.
- The command header shows zero active work, delegations, meetings, and
  failures.
- Desktop shows an empty Agent Workbench prompt until a card is selected.
- The selected agent shows `Idle`.
- Inbox, outbox, meetings, and errors are zero.
- Drill-down tabs show empty record lists.
- Missing `delegations` or `meetings` directories return empty activity rather
  than page or API errors.

### 2. Active Delegation

Create or trigger a delegation record where one configured agent requests work
from another and the record is `requested` or `running`.

Expected result:

- The target agent shows `Working`.
- The requester shows `Delegating`.
- Clicking the target card selects it in the persistent workbench on desktop or
  opens the detail sheet on mobile.
- The target Inbox count pill opens Inbox and shows the delegation ID.
- The requester Outbox count pill opens Outbox and shows the same delegation ID.
- The target inbox includes the delegation ID.
- The requester outbox includes the same delegation ID.
- The command header active work and delegation counts increase.
- The Recent Activity feed includes the delegation and opens the related
  agent's Inbox when clicked.
- Raw prompts, provider messages, and private failure text are not present in
  the organization page or API responses.

### 3. Active Meeting

Create or trigger a started meeting with a sponsor, chair, and participant.

Expected result:

- Sponsor, chair, and participant show `Meeting`.
- Each related agent's meeting drill-down includes the meeting ID.
- The Meetings count pill opens the Meetings workbench tab or detail tab.
- The command header active work and meeting counts increase.
- The Recent Activity feed includes the meeting and opens the chair or related
  agent's Meetings tab when clicked.
- The chair remains the visible owner of the consolidated recommendation.
- Raw meeting goals, notes, and participant turn text are not present in
  launcher API responses.

### 4. Failed Record

Create or retain a failed delegation or meeting record for a configured agent.

Expected result:

- The related agent shows `Failed`.
- The command header failure count increases and uses the failure styling.
- The failed record is selected as current when it is the newest relevant
  structured record.
- The Errors count pill opens the Failures tab.
- The errors count increases for the related agent.
- If a newer running delegation, started meeting, or completed record exists
  for the same agent, the badge follows that newer current activity while the
  failed record remains visible as `last_failure`.
- The Failures tab marks old failures as historical when newer current activity
  exists.
- Failed delegation and failed meeting records appear as failure entries in the
  Recent Activity feed and open the Failures tab.
- The Failures tab lists recent visible failed delegation and meeting records
  when more than one failure contributes to the selected agent's error count.
- Failure details are redacted to record status and identifiers; private error
  strings do not appear in launcher responses.

### 4a. Diagnostic Drilldown Path

Use local fixture records or retained test records that include one failed
delegation and one failed meeting for the same configured agent. Prefer copying
known-safe fixture JSON into a disposable workspace over triggering new runtime
work when the goal is UI verification.

Check the full operator path without pressing any runtime action:

```bash
curl -fsS "$LAUNCHER_URL/api/agents/organization"
curl -fsS "$LAUNCHER_URL/api/agents/$AGENT_ID/failures"
curl -fsS "$LAUNCHER_URL/api/agents/$AGENT_ID/activity/delegation/$DELEGATION_ID"
curl -fsS "$LAUNCHER_URL/api/agents/$AGENT_ID/activity/meeting/$MEETING_ID"
```

Expected result:

- The affected card shows `Failed` only when the failure is the current
  structured record. If newer running, started, or completed activity exists,
  the card follows that newer activity while the failure remains visible as a
  historical failure.
- The card shows a short failure reason when the current record has diagnostic
  fields. Old records without diagnostic fields still render with stable
  fallback copy instead of blank or broken UI.
- The Errors count opens the Failures tab. The tab labels current failures as
  current and older failures as historical, shows reason, source, severity,
  peer, role, timestamps, artifact count, and a Details action when detail is
  available.
- Record detail shows bounded summaries for request, context, result, memory
  errors, artifact errors, and participant status. Long task, result, and error
  text wraps and scrolls inside its detail area rather than expanding the
  command center indefinitely.
- Detail APIs return only records visible to the selected configured agent.
  A record for another agent returns `403`, an unknown record returns `404`,
  and neither response includes private task, goal, result, or error content.
- Live Logs `Selected Agent` matches only explicit structured agent reference
  fields. `Selected Record` further narrows or highlights lines that reference
  the selected record ID or known peer IDs from the selected record. Plain text
  substrings, partial IDs, tokens, and unrelated records must not match.

Operators can infer which persisted delegation or meeting record is currently
driving a card, the sanitized reason source used for the badge and list, and
which local record, artifact reference, or retained log line to inspect next.
Operators cannot infer that a failure was retried, repaired, externally
published, or written to memory unless the visible persisted record or artifact
status says so. The Organization page is a read-only diagnostic surface; it
does not restart agents, retry work, write memory, publish artifacts, mutate
Discord, or edit record JSON.

### 5. Missing Record Directory

Stop the gateway if needed, move the local `delegations` or `meetings`
directory aside in the test workspace, then reload the page.

Expected result:

- The organization page still loads.
- Missing stores are treated as empty stores.
- The page does not create replacement record files merely by loading.

Restore the directories before resuming normal operation.

### 6. Recent-Event Enrichment

Run a normal gateway turn or delegation that emits structured gateway log lines
or launcher-captured text log lines with explicit agent key/value fields such
as `agent_id`, `target_agent_id`, `parent_agent_id`, `chair_agent_id`, or
`sponsor_agent_id`.

Expected result:

- Matching events appear under the selected agent's Recent Events tab.
- Matching lines appear in Live Logs when the scope is `Selected Agent`; all
  gateway lines remain visible when the scope is `All Logs`.
- Malformed or unrelated log lines are ignored.
- Agent matches come only from explicit key/value fields, not arbitrary message
  text or partial substrings.
- Sensitive tokens and long messages are redacted or truncated.
- Recent events do not change an otherwise `Idle`, `Working`, `Delegating`,
  `Meeting`, or `Failed` badge.

### 7. Command Center Interaction Pass

Run this pass once with each staged state above: no records, historical failure
superseded by newer work, active delegation, active meeting, and gateway logs.

Expected result:

- Each agent card exposes a focused card-selection control followed by separate
  focused shortcut controls. Tab order must not trap focus or place shortcut
  controls inside another interactive control.
- Enter and Space on the focused card-selection control select the card without
  navigating away from the Organization page.
- The selected visual state follows the last selected card.
- Details is independently focusable and opens Overview.
- Inbox, Outbox, Meetings, and Errors shortcut pills open their matching
  workbench tabs on desktop and matching detail tabs on mobile.
- Switching workbench tabs does not change records, config, gateway status, or
  external artifacts.
- Live Logs reports stopped, stale, or polling errors as panel state instead of
  breaking the page.
- The organization canvas, command header, activity feed, and selected
  workbench remain usable after a background refresh.

## Read-Only Confirmation

Before and after loading the page and calling all five endpoints, compare the
config and workspace record files:

```bash
find "$PICOCLAW_WORKSPACE" -type f | sort | xargs shasum -a 256 > /tmp/org-before.sha
curl -fsS "$LAUNCHER_URL/api/agents/organization" >/dev/null
curl -fsS "$LAUNCHER_URL/api/agents/$AGENT_ID/activity" >/dev/null
curl -fsS "$LAUNCHER_URL/api/agents/$AGENT_ID/inbox" >/dev/null
curl -fsS "$LAUNCHER_URL/api/agents/$AGENT_ID/outbox" >/dev/null
curl -fsS "$LAUNCHER_URL/api/agents/$AGENT_ID/meetings" >/dev/null
find "$PICOCLAW_WORKSPACE" -type f | sort | xargs shasum -a 256 > /tmp/org-after.sha
diff -u /tmp/org-before.sha /tmp/org-after.sha
```

Expected result: no diff. If gateway activity is running concurrently, pause
new turns first or limit the comparison to config, `delegations`, and
`meetings` record directories for the staged test workspace.

## Local Verification Commands

Run the task verification set after changing this area:

```bash
go test ./web/backend/api ./pkg/agent ./pkg/config -run 'Agent|Organization|Activity|Inbox|Outbox|Meeting|Event|Failure' -count=1
cd web/frontend && pnpm lint && pnpm build
cd ../..
operations/audit-zehn-feature-task.sh 043-organization-command-center-verification
```
