# Runtime Architecture

## Gateway Startup

`pkg/gateway.Run` is the main runtime entry. It initializes panic/file logging, loads config, prechecks the gateway port, opens listeners, writes a singleton PID file with token metadata, creates the provider or `startupBlockedProvider`, creates the message bus and `AgentLoop`, then starts cron, heartbeat, media, channels, health, device services, optional voice, the agent loop, and config hot reload.

Gateway reload stops and recreates most services while preserving the channel manager. Changing the gateway listen address requires a full restart because the PID file and listener identity are not rewritten by reload.

## Message Bus

`pkg/bus.MessageBus` owns inbound, outbound, outbound media, audio, and voice channels. Defaults are intentionally buffered. Streaming channels are decoupled through `StreamDelegate`, so do not assume every output is a simple request/response.

## Turn Lifecycle

`pkg/agent.AgentLoop.Run` consumes inbound bus messages, resolves session and route scope, prevents duplicate active turns per session, and applies `agents.defaults.max_parallel_turns` with a worker semaphore.

The pipeline is split across:

- `pipeline_setup.go`: history, summary, memory, restore point, media, compression, user message persistence, provider candidates.
- `pipeline_llm.go`: hooks, fallback candidates, rate limiting, web search, thinking-level gating, timeout retry, media fallback, context compression retry.
- `pipeline_execute.go`: tool approvals, tool execution, async callbacks, media delivery, sensitive filtering, steering checkpoints.
- `pipeline_finalize.go`: assistant persistence, memory ingestion, session save, compaction.

## Long-Running Behavior

Long-running sessions and active turns are first-class. New user input can steer an active turn. Graceful interrupt asks the model to stop scheduling tools and summarize; hard abort cancels provider/tool work. Avoid changes that treat inactive network silence as runtime failure without tracing active-turn, streaming, WebSocket, and channel semantics.

## Durable Agent Delegation

Target-agent delegation is implemented as a real configured-agent turn, not as a chat-channel loop and not as the legacy `spawn`/`subagent` helper. The public tool layer lives in `pkg/tools/delegate.go` and `pkg/tools/delegate_status.go`; the runtime path lives in `pkg/agent/delegation.go`.

Important properties:

- Parent and target agent IDs are normalized and checked through `AgentRegistry.CanSpawnSubagent`, which is backed by `subagents.allow_agents`.
- The delegated turn runs on an internal `delegation` session scope for the target agent.
- Local delegation records are written under the agent workspace, redacted through config sensitive filtering, and stored with `0600` atomic writes.
- Async delegation is accepted only through a bounded executor owned by `AgentLoop`; executor close cancels/drains accepted async work.
- `delegation_status` and `delegation_inbox` require caller agent identity and filter by parent, target, requester, or explicit visible participants. Missing identity must fail closed.

## Chaired Meeting V1

`start_agent_meeting` is a private chaired sequential meeting flow. It is intentionally not live debate. A sponsor selects a chair; the chair consults required participants one at a time via durable delegation; the chair then synthesizes one recommendation with timeline, risks, approvals, and follow-ups.

Participant failures are required-participant failures by default: the meeting records the failed participant turn, marks the meeting failed or cancelled, and does not silently synthesize around missing required input. Meeting v2 debate is a future design area and should not be implied by v1 tool text or docs.

## GitHub Artifacts

Meeting/delegation GitHub artifacts are tracker outputs only. They should be created from redacted local records, not raw prompts or provider outputs. Artifact publishing is non-blocking through an `AgentLoop`-owned publisher with bounded capacity, timeout, and close/drain behavior. Publishing errors update local records as `failed`; disabled writers record `skipped` when executable work would otherwise need a tracker artifact.

## Providers

Provider selection is model-list driven. The explicit provider wins; otherwise PicoClaw infers by model prefix and defaults plain names to OpenAI-compatible. Multi-key model configs expand into virtual fallback entries. Fallback cooldowns are in-memory and distinguish standard, billing, context, format, timeout, network, and overload classes.
