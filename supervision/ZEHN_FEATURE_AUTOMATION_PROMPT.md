# Zehn Feature Automation Prompt

You are closing one Zehn delegation/meeting feature task. Do exactly one task
from `supervision/zehn_feature_tasks/` and do not switch tasks.

## Source Of Truth

Use current PicoClaw/Zehn implementation evidence first:

1. Current source under `pkg/agent`, `pkg/tools`, `pkg/session`, `pkg/config`,
   and `pkg/channels`.
2. Current tests next to the code being changed.
3. `.picoclaw/workspace/memory/AGENT_DELEGATION_SYSTEM.md`.
4. `.picoclaw/workspace/memory/AGENT_MEETING_SYSTEM.md`.
5. Upstream PicoClaw issue context only as background.

## Product Decisions Already Made

- Discord is the human visibility layer, not the internal delegation bus.
- GitHub Project is the tracker, not the company brain.
- Yaad plus curated business docs are durable memory.
- Existing `spawn` and `subagent` behavior must not be overloaded for durable
  target-agent delegation.
- Department heads may chair meetings inside their own domains.
- Meeting output defaults to one consolidated recommendation from the chair.

## Operating Rules

- Respect repo-local `CONTRIBUTING.md`.
- Keep changes scoped to the selected task's `Allowed repos/files`.
- Keep upstream-clean code generic. Zehn/Yaad/GitHub/Discord-specific behavior
  belongs in narrow adapters or private configuration paths.
- Do not commit secrets or local-only config.
- Add tests in the same package style as nearby tests.
- Run the task's verification commands.
- Leave unrelated user changes alone.

## Done Means

- The selected task's acceptance criteria are met.
- The selected task's verification commands pass or have a documented blocker.
- New behavior is covered with deterministic tests.
- Existing spawn, subagent, routing, session, and Discord behavior is not
  regressed.
- The final response names changed files, verification evidence, and remaining
  risks.
