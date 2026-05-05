# Current Agent Collaboration Audit

> Back to [README](../README.md)

This audit records the current PicoClaw/Zehn behavior before adding durable
target-agent delegation or chaired agent meetings. It is based on the current
source in `pkg/agent`, `pkg/tools`, `pkg/session`, `pkg/config`, and
`pkg/channels`, plus the Zehn delegation and meeting memory notes.

## Summary

PicoClaw already supports multiple configured agents for inbound channel
routing. A routed inbound message resolves to one real `AgentInstance`, receives
a routed session scope, and then runs through the normal agent loop.

PicoClaw also supports `spawn` and `subagent` as ephemeral helper work. Those
paths use subturns: child turns are linked to a parent turn, run in an
ephemeral in-memory session, and shallow-copy the parent agent. They do not
create durable delegation records, do not run as a real target agent identity,
and do not use Discord as an internal transport.

Current Zehn target-agent delegation and meetings are therefore missing. They
must be added as separate primitives and tools rather than overloading
`spawn`, `subagent`, session routing, or Discord self-messages.

## Current Inbound Routing

`pkg/channels/discord/discord.go` receives Discord user messages in
`DiscordChannel.handleMessage`. It ignores messages authored by the bot itself
at lines 523-530, enforces the sender allowlist at lines 532-550, applies group
trigger and mention filtering at lines 559-581, builds a `bus.InboundContext`
with Discord channel, chat, sender, message, mention, guild, and reply metadata
at lines 671-686, and calls `HandleInboundContext` at line 688.

The agent loop normalizes inbound messages and routes non-system messages in
`AgentLoop.processMessage` (`pkg/agent/agent_message.go` lines 105-203).
System messages bypass normal routing and go to `processSystemMessage` at
lines 135-138.

Normal routing is:

1. `resolveMessageRoute` calls `AgentRegistry.ResolveRoute` with the normalized
   inbound context (`pkg/agent/agent_message.go` lines 206-219).
2. `AgentRegistry.ResolveRoute` delegates to `routing.RouteResolver`
   (`pkg/agent/registry.go` lines 68-70).
3. `routing.RouteResolver.ResolveRoute` matches dispatch rules or falls back to
   the default agent (`pkg/routing/route.go` lines 39-61).
4. `AgentRegistry.GetAgent` resolves the selected configured agent
   (`pkg/agent/registry.go` lines 59-65).
5. `AgentLoop.processMessage` runs `al.runAgentLoop(ctx, agent, opts)` using
   that real configured agent (`pkg/agent/agent_message.go` line 203).

If a dispatch rule names an unknown agent, `RouteResolver.pickAgentID` falls
back to the default agent (`pkg/routing/route.go` lines 64-80). This is inbound
channel routing, not delegation.

## Current Session Allocation

`AgentLoop.allocateRouteSession` calls `session.AllocateRouteSession` with the
resolved route agent ID, normalized inbound context, and route session policy
(`pkg/agent/agent_message.go` lines 222-227).

`session.AllocateRouteSession` builds a structured `SessionScope`, canonical
session key, legacy session aliases, and main-session alias
(`pkg/session/allocator.go` lines 30-43). The scope includes normalized agent
ID, channel, account, and configured dimensions such as `space`, `chat`,
`topic`, and `sender` (`pkg/session/allocator.go` lines 45-112).

`resolveScopeKey` preserves an explicit caller-provided session key and
otherwise uses the routed session key (`pkg/agent/agent_utils.go` lines
324-329). `processMessage` attaches session aliases, inbound context, route
snapshot, and cloned session scope to the dispatch request
(`pkg/agent/agent_message.go` lines 170-179).

This means current durable session allocation belongs to inbound routed turns.
Subturn helper work intentionally does not use that durable route session.

## Current Subturn Spawning

The core subturn path is `spawnSubTurn` in `pkg/agent/subturn.go` lines
264-494.

Important current behavior:

- `AgentLoopSpawner.SpawnSubTurn` only converts `tools.SubTurnConfig` into
  `agent.SubTurnConfig` and delegates to `spawnSubTurn`
  (`pkg/agent/subturn.go` lines 209-237).
- `spawnSubTurn` enforces subturn concurrency and depth limits
  (`pkg/agent/subturn.go` lines 270-319).
- The child context is created from `context.Background()` with its own timeout,
  not from the parent context (`pkg/agent/subturn.go` lines 321-331).
- The child agent is a shallow copy of the parent turn's agent, falling back to
  the default agent only when the parent has no agent
  (`pkg/agent/subturn.go` lines 335-352).
- The child uses a new `ephemeralSession` as `agent.Sessions` and `childTS`
  session (`pkg/agent/subturn.go` lines 345-347 and 392-393).
- The child dispatch uses generated session key `subturn-N`, inherits the
  parent inbound context, disables history, disables summary, and suppresses
  outbound response sending (`pkg/agent/subturn.go` lines 354-371).
- Async results are delivered to the parent turn's `pendingResults` channel
  only when `cfg.Async` is true (`pkg/agent/subturn.go` lines 447-450 and
  499-517).

So subturns are isolated nested helper turns. They are not durable private
agent-to-agent work items.

## Current `spawn` Tool

`pkg/tools/spawn.go` defines the current `spawn` tool. Its schema exposes
`task`, optional `label`, and optional `agent_id` (`pkg/tools/spawn.go` lines
44-62).

The current direct path proves that `agent_id` does not select the target agent:

1. `SpawnTool.execute` reads `agent_id` at line 94.
2. If `agent_id` is present, it is passed only to `allowlistCheck` at lines
   96-101.
3. The subturn config created for `t.spawner.SpawnSubTurn` uses
   `Model: t.defaultModel`, `Tools: nil`, and the generated subagent prompt at
   lines 121-132.
4. `agent_id` is not copied into `tools.SubTurnConfig`.
5. `AgentLoopSpawner.SpawnSubTurn` has no target-agent field to receive
   (`pkg/agent/subturn.go` lines 221-234).
6. `spawnSubTurn` selects `baseAgent := parentTS.agent`, not a target agent
   (`pkg/agent/subturn.go` lines 335-352).

Therefore, in the registered `SpawnTool` path, `agent_id` currently gates
permission only. It does not run the target agent, does not use target
workspace, target prompt files, target tools, target sessions, or target memory.

There is a legacy `SubagentManager.Spawn` path that stores an `AgentID` on an
in-memory `SubagentTask` (`pkg/tools/subagent.go` lines 135-159) and passes it
to the manager spawner (`pkg/tools/subagent.go` lines 203-214). The manager
spawner registered in `pkg/agent/agent_init.go` can use that target ID to
select only a target model (`pkg/agent/agent_init.go` lines 282-287), then still
calls `spawnSubTurn` (`pkg/agent/agent_init.go` lines 290-301). That fallback
also does not run as the real target `AgentInstance`.

## Current `subagent` Tool

`pkg/tools/subagent.go` defines the synchronous `subagent` tool. Its schema only
accepts `task` and optional `label` (`pkg/tools/subagent.go` lines 370-384).
There is no `agent_id` argument.

When a spawner is available, `SubagentTool.Execute` calls
`t.spawner.SpawnSubTurn` with `Async: false`, the parent-derived default model,
and a generic subagent prompt (`pkg/tools/subagent.go` lines 395-422). The
result is returned synchronously to the caller (`pkg/tools/subagent.go` lines
427-450).

This is useful for bounded helper analysis, but it is not target-agent
delegation.

## Current Tool Registration And Permission Checks

`registerSharedTools` registers shared tools for every configured agent
(`pkg/agent/agent_init.go` lines 80-331).

For `spawn` and `subagent`:

- A `SubagentManager` is created per registered agent when `spawn` or
  `spawn_status` is enabled and `subagent` is enabled
  (`pkg/agent/agent_init.go` lines 228-233).
- The manager default model and workspace come from the current registering
  agent (`pkg/agent/agent_init.go` lines 233-234).
- The manager tool set is a clone of the current agent's tools before
  `spawn`/`spawn_status` are added (`pkg/agent/agent_init.go` lines 304-308).
- The registered `spawn` tool receives `NewSubTurnSpawner(al)`, not the
  manager's target-aware legacy spawner (`pkg/agent/agent_init.go` lines
  310-311).
- The `spawn` allowlist checker calls
  `registry.CanSpawnSubagent(currentAgentID, targetAgentID)`
  (`pkg/agent/agent_init.go` lines 312-315).

`AgentRegistry.CanSpawnSubagent` enforces `subagents.allow_agents`: it rejects
missing parents, missing allowlists, and non-matching target IDs, while allowing
`*` (`pkg/agent/registry.go` lines 84-103). This permission helper should be
reused for delegation, but the current `spawn` behavior should not be redefined
as durable delegation.

## Current `message` Tool

The `message` tool is user-channel output, not internal agent routing.

`MessageTool.Execute` accepts `content`, optional `channel`, optional `chat_id`,
and optional `reply_to_message_id` (`pkg/tools/integration/message.go` lines
39-61). If channel or chat ID is omitted, it uses the current tool context
(`pkg/tools/integration/message.go` lines 100-115). It calls the configured
send callback and records per-session targets for final-response suppression
(`pkg/tools/integration/message.go` lines 121-142).

The callback installed by `registerSharedTools` publishes a `bus.OutboundMessage`
with outbound context and current turn metadata (`pkg/agent/agent_init.go` lines
135-159). It does not enqueue an inbound message and does not select another
agent.

## Discord Self-Message Handling

Discord is currently a human-facing channel. It is not an internal delegation
bus.

The Discord adapter drops bot-authored messages immediately
(`pkg/channels/discord/discord.go` lines 523-530). If an agent sends a Discord
message using the `message` tool or a final response, that outbound message is
not reprocessed as an inbound user message because self-messages are ignored.

The Zehn delegation note also states that delegation should not depend on
Discord self-messages re-entering the bot and that Discord may receive
visibility summaries but is not the internal delegation bus
(`.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`). The meeting note
matches this: Discord should post concise summaries, while detailed records live
in meeting documents, Yaad, or GitHub artifacts
(`.picoclaw/workspace/memory/AGENT_MEETING_SYSTEM.md`).

## Current System Messages

`processSystemMessage` handles `Channel == "system"` messages separately
(`pkg/agent/agent_message.go` lines 230-301). It parses an origin channel and
chat ID from `msg.ChatID`, strips a `Result:\n` prefix when present, logs and
returns for internal origin channels, otherwise uses the default agent and main
session to produce a user-facing response.

This is not a target-agent delegation path. It does not resolve a target agent
from a delegation request and does not preserve target workspace, prompt files,
tools, memory, or durable task state.

## What Exists Today

- Multi-agent inbound routing through `AgentRegistry`, `RouteResolver`, and
  route-scoped session allocation.
- Per-agent tool registration, including `message`, `spawn`, `subagent`, and
  `spawn_status` when enabled.
- Permission checking for optional `spawn.agent_id` through
  `subagents.allow_agents`.
- Ephemeral child turns through `spawnSubTurn`, with parent-child turn tracking,
  depth/concurrency limits, and async result delivery to the parent turn.
- Discord inbound filtering, allowlist checks, group trigger checks, and
  self-message ignore behavior.

## What Is Missing

- No generic `RunAgentDelegation` primitive that runs a real target
  `AgentInstance`.
- No `delegate_to_agent` tool.
- No durable delegation record schema or local delegation store.
- No Yaad persistence for delegation request/status/result.
- No async delegation status, inbox, or completion tools.
- No meeting record schema or `start_agent_meeting` capability.
- No chaired participant-turn workflow that uses real target-agent delegation.
- No GitHub issue/project integration for delegation or meeting artifacts.
- No Discord visibility adapter for delegation or meeting summaries beyond
  ordinary channel messaging.

## What Must Not Be Overloaded

- Do not redefine `spawn` as durable target-agent delegation. Existing `spawn`
  is an async ephemeral subturn helper.
- Do not redefine `subagent` as target-agent delegation. Existing `subagent` is
  a synchronous generic helper.
- Do not use Discord self-messages as the internal transport. The Discord
  adapter intentionally ignores bot-authored messages.
- Do not make inbound dispatch rules stand in for private delegation. Inbound
  routing maps external channel messages to agents; delegation is internal
  agent-to-agent work.
- Do not make GitHub Project the company brain. Zehn decisions say GitHub
  Project is a tracker, while Yaad plus curated business docs are durable
  memory.

## Files That Must Change For Target-Agent Delegation

Task 003 should add or change the generic source-level delegation primitive in:

- `pkg/agent/delegation*.go`: new request/result types and
  `RunAgentDelegation`-style implementation.
- `pkg/agent/delegation*_test.go`: deterministic coverage for success,
  permission denial, missing targets, target workspace/prompt/session identity,
  and failure paths.
- `pkg/agent/registry.go`: reuse or narrowly extend target lookup and
  `CanSpawnSubagent` permission behavior if descriptors or clearer errors are
  needed.
- `pkg/agent/agent_init.go`: later registration of new delegation tools once
  the primitive exists.
- `pkg/session/**`: only if a private internal delegation session scope or key
  helper is needed.

Task 004 and later should add the tool and durable layers in:

- `pkg/tools/delegate*.go` and `pkg/tools/delegate*_test.go`.
- `pkg/agent/delegation*.go` for local records and async state.
- Narrow persistence adapters for Yaad, GitHub, and Discord summaries in later
  tasks.

## Files That Should Remain Unchanged For Upstream Behavior

Preserve these existing behavior surfaces unless a later task has a specific,
tested compatibility fix:

- `pkg/tools/spawn.go`: keep current async ephemeral `spawn` semantics.
- `pkg/tools/subagent.go`: keep current synchronous generic helper semantics.
- `pkg/agent/subturn.go`: keep subturn isolation, shallow parent-agent copy,
  ephemeral session store, and parent result delivery semantics.
- `pkg/channels/discord/discord.go`: keep Discord self-message ignore and
  human-channel behavior.
- `pkg/tools/integration/message.go`: keep `message` as outbound user-channel
  publishing.
- `pkg/routing/route.go`: keep inbound dispatch semantics separate from
  internal delegation.
- `pkg/session/allocator.go`: keep route session allocation semantics for
  inbound messages.

## Upstream References

The local Zehn task sequence records these upstream PicoClaw references as
background for multi-agent discovery and delegation work:

- PicoClaw issue `#1934`
- PicoClaw issue `#2148`
- PicoClaw PR `#2158`

They are background context only for this audit. The source evidence above is
the source of truth for current behavior in this workspace.

