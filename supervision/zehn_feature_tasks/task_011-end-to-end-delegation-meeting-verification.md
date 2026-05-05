# Task 011: End-To-End Delegation And Meeting Verification

Slug: `011-end-to-end-delegation-meeting-verification`

Docs-only allowed: no

## Goal

Add end-to-end verification proving the complete Zehn delegation and meeting
workflow works without regressing existing PicoClaw behavior.

## Allowed repos/files

- `pkg/agent/**`
- `pkg/tools/**`
- `pkg/config/**`
- `pkg/session/**`
- `pkg/channels/discord/**`
- `docs/architecture/**`
- `docs/reference/**`
- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`
- `.picoclaw/workspace/memory/AGENT_MEETING_SYSTEM.md`
- `supervision/zehn_feature_tasks/**`

## Required reading

- all previous Zehn feature task outputs
- `pkg/agent/agent_test.go`
- `pkg/agent/subturn_test.go`
- `pkg/tools/spawn_test.go`
- `pkg/tools/subagent_tool_test.go`
- `pkg/channels/discord/discord.go`
- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`
- `.picoclaw/workspace/memory/AGENT_MEETING_SYSTEM.md`

## Work

- Add integration-style tests with fake providers and fake external adapters.
- Verify Ali-to-CEO objective, CEO-to-domain-head delegation, domain meeting,
  consolidated recommendation, and approval-needed output.
- Verify no GitHub issue is created for non-executable discussion.
- Verify GitHub issue/comment paths for executable work.
- Verify Yaad/local persistence fallback behavior.
- Verify existing `spawn`, `subagent`, routing, and Discord self-message
  behavior are not regressed.
- Update docs with the final supported workflow.

## Acceptance criteria

- Full delegation and meeting path has deterministic test coverage.
- Existing single-agent behavior remains unchanged.
- Existing `spawn`/`subagent` behavior remains unchanged except where explicitly
  documented and tested.
- Meeting output includes consolidated recommendation, participants, timeline,
  risks, and follow-ups.
- Verification does not require live Discord, live GitHub, or live Yaad.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./pkg/agent ./pkg/tools ./pkg/config ./pkg/session ./pkg/channels/discord -count=1
go test ./pkg/agent ./pkg/tools ./pkg/config ./pkg/session ./pkg/channels/discord -race
make generate
make test
```
