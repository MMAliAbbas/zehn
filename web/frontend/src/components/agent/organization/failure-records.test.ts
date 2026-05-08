/// <reference types="node" />
import assert from "node:assert/strict"

import test from "node:test"

import type { AgentOrganizationActivityRecord } from "@/api/agents"

import { buildFailureDrilldownRecords } from "./formatting.ts"

test("uses the recent failure list when multiple failures exist", () => {
  const current: AgentOrganizationActivityRecord = {
    type: "delegation",
    record_id: "delegation-active",
    status: "running",
    role: "target",
    updated_at: "2026-05-07T11:00:00Z",
  }
  const lastFailure: AgentOrganizationActivityRecord = {
    type: "meeting",
    record_id: "meeting-failed",
    status: "failed",
    role: "participant",
    updated_at: "2026-05-07T10:00:00Z",
  }
  const recentFailures: AgentOrganizationActivityRecord[] = [
    {
      type: "meeting",
      record_id: "meeting-failed",
      status: "failed",
      role: "participant",
      updated_at: "2026-05-07T10:00:00Z",
    },
    {
      type: "delegation",
      record_id: "delegation-failed",
      status: "failed",
      role: "target",
      updated_at: "2026-05-07T09:00:00Z",
    },
  ]

  assert.deepEqual(
    buildFailureDrilldownRecords(current, lastFailure, recentFailures).map(
      (record) => `${record.type}:${record.record_id}`,
    ),
    ["meeting:meeting-failed", "delegation:delegation-failed"],
  )
})

test("falls back to current and last failure summaries while data is unavailable", () => {
  const current: AgentOrganizationActivityRecord = {
    type: "delegation",
    record_id: "delegation-current-failed",
    status: "failed",
    role: "target",
  }
  const lastFailure: AgentOrganizationActivityRecord = {
    type: "delegation",
    record_id: "delegation-current-failed",
    status: "failed",
    role: "target",
  }

  assert.deepEqual(
    buildFailureDrilldownRecords(current, lastFailure, undefined).map(
      (record) => `${record.type}:${record.record_id}`,
    ),
    ["delegation:delegation-current-failed"],
  )
})
