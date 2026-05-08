# Task 046: Organization Card Accessibility Controls

Slug: `046-organization-card-accessibility-controls`

Docs-only allowed: no

## Goal

Fix the agent card interaction model so the organization tree remains keyboard
and screen-reader friendly without nesting interactive controls inside a
container that also acts as a button.

## Allowed repos/files

- `web/frontend/src/components/agent/organization/agent-card.tsx`
- `web/frontend/src/components/agent/organization/status-components.tsx`
- `web/frontend/src/components/agent/organization/organization-state.ts`
- `web/frontend/src/components/agent/organization/organization-state.test.ts`
- `docs/reference/**`
- `supervision/**`

## Required reading

- `web/frontend/src/components/agent/organization/agent-card.tsx`
- `web/frontend/src/components/agent/organization/status-components.tsx`
- `web/frontend/src/components/agent/organization/organization-state.ts`
- `docs/reference/agent-organization-live-verification.md`

## Work

- Replace the outer `role="button"` card pattern with valid interactive
  structure.
- Keep the same desktop workbench and mobile detail-sheet behavior.
- Keep Details, inbox, outbox, meetings, and failures shortcuts independently
  focusable.
- Preserve keyboard support for selecting a card and opening shortcuts.
- Add or update tests for shortcut resolution if behavior changes.
- Document the expected manual keyboard check.

## Acceptance criteria

- Agent cards do not contain nested interactive elements inside a fake button.
- Keyboard users can select an agent and open each shortcut predictably.
- Screen-reader labels remain meaningful for card selection and shortcut
  controls.
- Visual layout does not regress.
- Frontend build passes.

## Verification commands

```bash
cd /Users/aliai/zehn
go test ./web/backend/api -run 'Agent|Organization' -count=1
cd /Users/aliai/zehn/web/frontend
node --test --experimental-strip-types src/components/agent/organization/organization-state.test.ts
pnpm lint
pnpm build
cd /Users/aliai/zehn
operations/audit-zehn-feature-task.sh 046-organization-card-accessibility-controls
```
