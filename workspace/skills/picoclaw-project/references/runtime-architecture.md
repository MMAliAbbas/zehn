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

## Providers

Provider selection is model-list driven. The explicit provider wins; otherwise PicoClaw infers by model prefix and defaults plain names to OpenAI-compatible. Multi-key model configs expand into virtual fallback entries. Fallback cooldowns are in-memory and distinguish standard, billing, context, format, timeout, network, and overload classes.

