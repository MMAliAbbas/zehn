# Memory, Sessions, Context, And Yaad

## Workspace Memory

Workspace prompt memory is read from `<workspace>/memory/MEMORY.md` and daily notes under `<workspace>/memory/YYYYMM/YYYYMMDD.md`. These are injected by the context builder.

## Sessions

Conversation history lives under `<workspace>/sessions`. The current JSONL store appends per-session events and stores summaries, aliases, and scope metadata in `.meta.json` files. A legacy JSON session store exists and can migrate older sessions.

## Context Managers

The default `legacy` context manager reads session history, performs emergency compression, and summarizes asynchronously. A `seahorse` context manager exists behind supported build behavior and uses SQLite-backed retrieval plus retrieval tools.

## Turn Context

During setup, the pipeline assembles history, summaries, memory, media references, and provider candidates. On context pressure it can proactively compress, retry after context overflow errors, and persist updated summaries.

## Yaad Positioning

For Zehn, keep Yaad private unless explicitly designing an upstream-neutral extension point. The safest first integration is Yaad as a private MCP server or private prompt/context contributor configured outside upstream defaults. Avoid upstream PRs that mention or require Yaad.

A future private Zehn branch could add a Yaad-backed `ContextManager`, but that should happen only after the MCP/private-config path proves stable and the session lifecycle is fully understood.

