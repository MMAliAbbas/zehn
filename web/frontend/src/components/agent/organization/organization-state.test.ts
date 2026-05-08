/// <reference types="node" />
import assert from "node:assert/strict"

import test from "node:test"

import {
  DEFAULT_WORKBENCH_SECTION,
  createOrganizationSelectionState,
  detailTabForWorkbenchSection,
  resolveActivityShortcut,
  resolveSelectedOrganizationAgent,
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

test("resolves error shortcuts directly to the failures detail tab", () => {
  assert.deepEqual(resolveActivityShortcut("errors"), {
    workbenchSection: "failures",
    detailTab: "failures",
  })
})

test("resolves failures workbench selection to the failures detail tab", () => {
  assert.equal(detailTabForWorkbenchSection("failures"), "failures")
})

test("resolves live logs workbench selection to the live logs detail tab", () => {
  assert.equal(detailTabForWorkbenchSection("live-logs"), "live-logs")
})

test("resolves a selected agent from snapshot agents before walking roots", () => {
  const snapshot = {
    agents: {
      "li-engineering": {
        id: "li-engineering",
        label: "Engineering",
        status: "working",
        activity: {
          inbox_count: 1,
          outbox_count: 0,
          meeting_count: 0,
          failure_count: 0,
        },
      },
    },
    roots: [
      {
        id: "li-cto",
        label: "CTO",
        status: "idle",
        activity: {
          inbox_count: 0,
          outbox_count: 0,
          meeting_count: 0,
          failure_count: 0,
        },
        children: [
          {
            id: "li-engineering",
            label: "Engineering from tree",
            status: "idle",
            activity: {
              inbox_count: 0,
              outbox_count: 0,
              meeting_count: 0,
              failure_count: 0,
            },
          },
        ],
      },
    ],
    activity: {
      delegation_count: 1,
      meeting_count: 0,
      failure_count: 0,
      active_count: 1,
    },
    metadata: {
      source: "test",
      generated_at: "2026-05-09T00:00:00Z",
      has_hierarchy: true,
    },
  }

  assert.equal(
    resolveSelectedOrganizationAgent(snapshot, "li-engineering")?.label,
    "Engineering",
  )
})
