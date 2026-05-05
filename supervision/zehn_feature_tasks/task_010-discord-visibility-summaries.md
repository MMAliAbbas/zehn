# Task 010: Discord Visibility Summaries

Slug: `010-discord-visibility-summaries`

Docs-only allowed: no

## Goal

Add optional Discord visibility summaries for delegation and meetings without
using Discord as the internal delegation or meeting bus.

## Allowed repos/files

- `pkg/agent/meeting*.go`
- `pkg/agent/delegation*.go`
- `pkg/tools/meeting*.go`
- `pkg/tools/delegate*.go`
- `pkg/tools/integration/message.go`
- `pkg/tools/integration/message_test.go`
- `pkg/channels/discord/discord.go`
- `pkg/channels/discord/*_test.go`
- `pkg/config/config.go`
- `pkg/config/defaults.go`
- `.picoclaw/workspace/memory/AGENT_MEETING_SYSTEM.md`
- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `pkg/tools/integration/message.go`
- `pkg/channels/discord/discord.go`
- `.picoclaw/workspace/memory/AGENT_MEETING_SYSTEM.md`
- `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`

## Work

- Add optional summary hooks for delegation created, meeting opened,
  recommendation ready, approval needed, issue created, blocker raised, and
  completion.
- Use existing outbound message mechanisms.
- Do not re-ingest self-authored Discord messages.
- Do not post raw internal transcripts by default.
- Respect configured channel allowlists and dispatch boundaries.

## Acceptance criteria

- Discord summaries can be disabled.
- Discord summaries are concise and event-based.
- Discord self-message behavior remains ignored by inbound handling.
- Internal delegation and meeting execution do not depend on Discord.
- Existing Discord channel tests still pass.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./pkg/channels/discord ./pkg/tools/integration -count=1
go test ./pkg/agent ./pkg/tools ./pkg/channels/discord
```
