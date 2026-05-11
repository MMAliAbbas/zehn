/// <reference types="node" />
import assert from "node:assert/strict"

import test from "node:test"

import type { AgentOrganizationActivityRecord } from "@/api/agents"

import {
  buildFailureDrilldownRecords,
  formatDiagnosticFreshness,
  formatDiagnosticReason,
  formatDiagnosticReasonSource,
  formatDiagnosticSeverity,
} from "./formatting.ts"

const t = ((key: string, fallback?: string) => fallback ?? key) as never

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

test("formats diagnostic reason metadata with stable fallbacks", () => {
  const record: AgentOrganizationActivityRecord = {
    type: "delegation",
    record_id: "delegation-failed",
    status: "failed",
    reason: "  provider timed out  ",
    reason_source: "record_error",
    severity: "error",
    current: true,
  }

  assert.equal(formatDiagnosticReason(record, t), "provider timed out")
  assert.equal(formatDiagnosticReasonSource(record, t), "Record error")
  assert.equal(formatDiagnosticSeverity(record, t), "Error")
  assert.equal(formatDiagnosticFreshness(record, false, t), "Current")
})

test("formats missing and stale diagnostic fields without broken copy", () => {
  const record: AgentOrganizationActivityRecord = {
    type: "meeting",
    record_id: "meeting-failed",
    status: "failed",
    stale: true,
  }

  assert.equal(
    formatDiagnosticReason(record, t),
    "No diagnostic reason available",
  )
  assert.equal(formatDiagnosticReasonSource(record, t), "Unknown source")
  assert.equal(formatDiagnosticSeverity(record, t), "Unknown")
  assert.equal(formatDiagnosticFreshness(record, true, t), "Stale")
})
