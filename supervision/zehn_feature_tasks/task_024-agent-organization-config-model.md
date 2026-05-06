# Task 024: Agent Organization Config Model

Slug: `024-agent-organization-config-model`

Docs-only allowed: no

## Goal

Add a generic optional configuration model for agent organization hierarchy that
is separate from delegation permissions.

## Allowed repos/files

- `pkg/config/config.go`
- `pkg/config/defaults.go`
- `pkg/config/config_test.go`
- `pkg/config/*organization*_test.go`
- `docs/reference/**`
- `supervision/zehn_feature_tasks/**`

## Required reading

- `pkg/config/config.go`
- `pkg/config/defaults.go`
- `pkg/config/config_test.go`
- `docs/reference/agent-delegation-meetings.md`
- `supervision/ZEHN_AGENT_ORGANIZATION_UI_PLAN.md`

## Work

- Add an optional `agents.organization` config section.
- Model explicit roots and nodes without changing existing agent loading when
  the section is absent.
- Keep reporting hierarchy separate from `subagents.allow_agents`.
- Add validation/helper behavior for duplicate nodes, unknown agents, unknown
  parents, and cycles.
- Add tests for parsing, empty config compatibility, deterministic sorting, and
  invalid hierarchy cases.
- Add generic reference documentation explaining reporting hierarchy versus
  delegation permissions.

## Acceptance criteria

- Existing configs without organization metadata continue to load.
- Organization metadata can represent multiple roots and stable sibling order.
- Invalid hierarchy data fails deterministically with useful errors.
- No runtime behavior changes are introduced outside config parsing/helpers.
- Documentation avoids private deployment details and stays generic.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./pkg/config -run 'Agent|Organization|Config|Hierarchy' -count=1
go test ./pkg/config -run 'Agent|Organization|Config|Hierarchy' -race
operations/audit-zehn-feature-task.sh 024-agent-organization-config-model
```

