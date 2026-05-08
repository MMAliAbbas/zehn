/// <reference types="node" />

import assert from "node:assert/strict"
import test from "node:test"

import {
  filterAgentLogLines,
  findAgentLogReferenceFields,
} from "./agent-log-filter.ts"

test("matches selected agents in structured JSON log fields", () => {
  const line = JSON.stringify({
    level: "info",
    event: "delegation.completed",
    agent_id: "li-cto",
    target_agent_id: "li-engineering",
    message: "done",
  })

  assert.deepEqual(findAgentLogReferenceFields(line, "li-engineering"), [
    "target_agent_id",
  ])
  assert.deepEqual(findAgentLogReferenceFields(line, "li-cto"), ["agent_id"])
})

test("matches selected agents in deterministic key-value log fields", () => {
  const line =
    'ts=2026-05-09T00:00:00Z requester_id=li-cro chair_agent_id="li-cto" msg="meeting started"'

  assert.deepEqual(findAgentLogReferenceFields(line, "li-cro"), [
    "requester_id",
  ])
  assert.deepEqual(findAgentLogReferenceFields(line, "li-cto"), [
    "chair_agent_id",
  ])
})

test("filters only selected-agent lines without mutating the source buffer", () => {
  const logs = [
    "level=info agent_id=li-cto msg=start",
    "level=info target_agent_id=li-engineering msg=delegated",
    "level=info agent_id=li-cfo msg=ignored",
  ]

  assert.deepEqual(filterAgentLogLines(logs, "li-engineering", "selected"), [
    "level=info target_agent_id=li-engineering msg=delegated",
  ])
  assert.deepEqual(filterAgentLogLines(logs, "li-engineering", "all"), logs)
})

test("does not match arbitrary message text or sensitive-looking fields", () => {
  const logs = [
    "level=info msg=li-engineering",
    "token=li-engineering requester_id=li-cro",
  ]

  assert.deepEqual(findAgentLogReferenceFields(logs[0], "li-engineering"), [])
  assert.deepEqual(findAgentLogReferenceFields(logs[1], "li-engineering"), [])
})
