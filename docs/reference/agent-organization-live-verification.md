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

These endpoints must not write config, delegation records, meeting records,
memory, channel state, GitHub artifacts, Discord messages, or other external
artifacts. They only read launcher config, local delegation and meeting record
directories, and bounded gateway log lines for recent-event enrichment.

## Badge Meanings

- `Idle`: the agent is configured and has no active or failed structured
  activity selected as current.
- `Working`: the newest highest-priority active delegation targets this agent.
- `Delegating`: the newest highest-priority active delegation was requested by
  this agent.
- `Meeting`: the agent is sponsor, chair, or participant in an active meeting.
- `Failed`: the newest highest-priority delegation or meeting record involving
  the agent failed. Failure takes precedence over active work.

Recent gateway events can appear in the detail drawer, but they do not change
badge status, counters, current activity, or failure precedence.

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
```

Expected result:

- The page loads in flat mode when `agents.organization` is absent.
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
- The target inbox includes the delegation ID.
- The requester outbox includes the same delegation ID.
- Raw prompts, provider messages, and private failure text are not present in
  the organization page or API responses.

### 3. Active Meeting

Create or trigger a started meeting with a sponsor, chair, and participant.

Expected result:

- Sponsor, chair, and participant show `Meeting`.
- Each related agent's meeting drill-down includes the meeting ID.
- The chair remains the visible owner of the consolidated recommendation.
- Raw meeting goals, notes, and participant turn text are not present in
  launcher API responses.

### 4. Failed Record

Create or retain a failed delegation or meeting record for a configured agent.

Expected result:

- The related agent shows `Failed`.
- The failed record is selected as current when it is the highest-priority
  recent structured record.
- The errors count increases for the related agent.
- Failure details are redacted to record status and identifiers; private error
  strings do not appear in launcher responses.

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
with `agent_id`, `target_agent_id`, `parent_agent_id`, `chair_agent_id`, or
`sponsor_agent_id`.

Expected result:

- Matching events appear under the selected agent's Recent Events tab.
- Malformed or unrelated log lines are ignored.
- Sensitive tokens and long messages are redacted or truncated.
- Recent events do not change an otherwise `Idle`, `Working`, `Delegating`,
  `Meeting`, or `Failed` badge.

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
go test ./web/backend/api ./pkg/agent ./pkg/config -run 'Agent|Organization|Activity|Inbox|Outbox|Meeting|Event' -count=1
cd web/frontend && pnpm build
cd ../..
operations/audit-zehn-feature-task.sh 030-agent-organization-live-verification
```
