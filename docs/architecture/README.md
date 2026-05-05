# Architecture

Internal architecture notes for major runtime mechanisms and subsystem design.

- [Steering](steering.md): injecting messages into a running agent loop between tool calls.
- [SubTurn Mechanism](subturn.md): sub-agent coordination, concurrency control, and lifecycle handling.
- [Current Agent Collaboration Audit](current-agent-collaboration-audit.md): current multi-agent routing, subturn, tool, session, and Discord behavior before durable delegation.
- [Session System](session-system.md): session scope allocation, JSONL persistence, alias compatibility, and migration. ([ZH](session-system.zh.md))
- [Routing System](routing-system.md): agent dispatch, session policy selection, and light/heavy model routing. ([ZH](routing-system.zh.md))
- [Hook System Guide](hooks/README.md): current hook architecture and protocol details.
- [Agent Refactor](agent-refactor/README.md): notes and checkpoints for the agent refactor work.

For proposal-style or exploratory docs, also see [`../design/`](../design/).
