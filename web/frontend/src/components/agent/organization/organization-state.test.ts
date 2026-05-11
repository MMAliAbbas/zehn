/// <reference types="node" />
import assert from "node:assert/strict"

import test from "node:test"

import {
  filterOrganizationLogLines,
  findOrganizationLogCorrelationFields,
} from "./organization-log-correlation.ts"
import {
  DEFAULT_WORKBENCH_SECTION,
  clearSelectedOrganizationRecord,
  createOrganizationSelectionState,
  detailTabForWorkbenchSection,
  resolveActivityShortcut,
  resolveAgentCardShortcut,
  resolveSelectableActivityRecord,
  resolveSelectedOrganizationAgent,
  selectOrganizationActivityRecord,
  selectOrganizationAgent,
  selectOrganizationWorkbenchSection,
} from "./organization-state.ts"
import { AGENT_WORKBENCH_SECTIONS } from "./types.ts"

test("creates empty organization selection state with overview workbench", () => {
  assert.deepEqual(createOrganizationSelectionState(), {
    selectedAgentID: null,
    workbenchSection: DEFAULT_WORKBENCH_SECTION,
    selectedRecord: null,
  })
})

test("selects an agent while preserving the current workbench section by default", () => {
  const current = {
    selectedAgentID: "li-cto",
    workbenchSection: "meetings" as const,
    selectedRecord: null,
  }

  assert.deepEqual(selectOrganizationAgent(current, "li-engineering"), {
    selectedAgentID: "li-engineering",
    workbenchSection: "meetings",
    selectedRecord: null,
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
      selectedRecord: null,
    },
  )
})

test("selects an activity record and moves the workbench to its source section", () => {
  const current = selectOrganizationAgent(
    createOrganizationSelectionState(),
    "li-engineering",
    "overview",
  )

  assert.deepEqual(
    selectOrganizationActivityRecord(current, {
      type: "delegation",
      recordID: "delegation-123",
      sourceSection: "failures",
      title: "Current failure",
    }),
    {
      selectedAgentID: "li-engineering",
      workbenchSection: "failures",
      selectedRecord: {
        type: "delegation",
        recordID: "delegation-123",
        sourceSection: "failures",
        title: "Current failure",
      },
    },
  )
})

test("clears the selected activity record without changing the selected agent or section", () => {
  const current = selectOrganizationActivityRecord(
    selectOrganizationAgent(
      createOrganizationSelectionState(),
      "li-engineering",
      "inbox",
    ),
    {
      type: "delegation",
      recordID: "delegation-123",
      sourceSection: "inbox",
    },
  )

  assert.deepEqual(clearSelectedOrganizationRecord(current), {
    selectedAgentID: "li-engineering",
    workbenchSection: "inbox",
    selectedRecord: null,
  })
})

test("preserves a selected activity record when switching to live logs", () => {
  const current = selectOrganizationActivityRecord(
    selectOrganizationAgent(
      createOrganizationSelectionState(),
      "li-engineering",
    ),
    {
      type: "delegation",
      recordID: "delegation-123",
      sourceSection: "failures",
    },
  )

  assert.deepEqual(selectOrganizationWorkbenchSection(current, "live-logs"), {
    selectedAgentID: "li-engineering",
    workbenchSection: "live-logs",
    selectedRecord: {
      type: "delegation",
      recordID: "delegation-123",
      sourceSection: "failures",
    },
  })
  assert.equal(
    selectOrganizationWorkbenchSection(current, "overview").selectedRecord,
    null,
  )
})

test("only resolves selectable activity records when detail inspection is available", () => {
  const record = {
    type: "meeting",
    recordID: "meeting-123",
    sourceSection: "meetings" as const,
    title: "Planning meeting",
  }

  assert.equal(resolveSelectableActivityRecord(record, false), null)
  assert.equal(resolveSelectableActivityRecord(record, undefined), null)
  assert.deepEqual(resolveSelectableActivityRecord(record, true), record)
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

test("resolves card shortcuts to their independently focusable destinations", () => {
  assert.deepEqual(resolveAgentCardShortcut("details"), {
    workbenchSection: "overview",
    detailTab: "overview",
  })
  assert.deepEqual(resolveAgentCardShortcut("inbox"), {
    workbenchSection: "inbox",
    detailTab: "inbox",
  })
  assert.deepEqual(resolveAgentCardShortcut("outbox"), {
    workbenchSection: "outbox",
    detailTab: "outbox",
  })
  assert.deepEqual(resolveAgentCardShortcut("meetings"), {
    workbenchSection: "meetings",
    detailTab: "meetings",
  })
  assert.deepEqual(resolveAgentCardShortcut("errors"), {
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

test("correlates live logs with the selected activity record and known peers", () => {
  const logs = [
    "level=info agent_id=li-engineering msg=start",
    "level=info delegation_id=delegation-123 target_agent_id=li-engineering",
    "level=info requester_id=li-cto msg=peer update",
    "level=info agent_id=li-cfo msg=global",
  ]
  const target = {
    selectedAgentID: "li-engineering",
    selectedRecordID: "delegation-123",
    peerAgentIDs: ["li-cto"],
  }

  assert.deepEqual(findOrganizationLogCorrelationFields(logs[1], target), [
    "target_agent_id",
    "record_id",
  ])
  assert.deepEqual(
    filterOrganizationLogLines(logs, "selected-record", target),
    [
      "level=info agent_id=li-engineering msg=start",
      "level=info delegation_id=delegation-123 target_agent_id=li-engineering",
      "level=info requester_id=li-cto msg=peer update",
    ],
  )
  assert.deepEqual(filterOrganizationLogLines(logs, "all", target), logs)
})
