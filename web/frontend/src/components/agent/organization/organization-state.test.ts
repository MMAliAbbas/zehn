/// <reference types="node" />

import assert from "node:assert/strict"
import test from "node:test"

import {
  createOrganizationSelectionState,
  DEFAULT_WORKBENCH_SECTION,
  resolveActivityShortcut,
  selectOrganizationAgent,
} from "./organization-state.ts"
import { AGENT_WORKBENCH_SECTIONS } from "./types.ts"

test("creates empty organization selection state with overview workbench", () => {
  assert.deepEqual(createOrganizationSelectionState(), {
    selectedAgentID: null,
    workbenchSection: DEFAULT_WORKBENCH_SECTION,
  })
})

test("selects an agent while preserving the current workbench section by default", () => {
  const current = {
    selectedAgentID: "li-cto",
    workbenchSection: "meetings" as const,
  }

  assert.deepEqual(selectOrganizationAgent(current, "li-engineering"), {
    selectedAgentID: "li-engineering",
    workbenchSection: "meetings",
  })
})

test("selects an agent and records an explicit workbench section", () => {
  assert.deepEqual(
    selectOrganizationAgent(
      createOrganizationSelectionState(),
      "li-cro",
      "outbox",
    ),
    {
      selectedAgentID: "li-cro",
      workbenchSection: "outbox",
    },
  )
})

test("declares all planned organization workbench sections", () => {
  assert.deepEqual(AGENT_WORKBENCH_SECTIONS, [
    "overview",
    "inbox",
    "outbox",
    "meetings",
    "failures",
    "recent",
    "live-logs",
  ])
})

test("resolves activity shortcuts to deterministic visible detail tabs", () => {
  assert.deepEqual(resolveActivityShortcut("inbox"), {
    workbenchSection: "inbox",
    detailTab: "inbox",
  })
  assert.deepEqual(resolveActivityShortcut("outbox"), {
    workbenchSection: "outbox",
    detailTab: "outbox",
  })
  assert.deepEqual(resolveActivityShortcut("meetings"), {
    workbenchSection: "meetings",
    detailTab: "meetings",
  })
})

test("resolves error shortcuts to the failures workbench with recent events as the visible fallback", () => {
  assert.deepEqual(resolveActivityShortcut("errors"), {
    workbenchSection: "failures",
    detailTab: "recent",
  })
})
