/// <reference types="node" />

import assert from "node:assert/strict"
import test from "node:test"

import {
  applyGatewayLogsResponse,
  canPollGatewayLogs,
  createGatewayLogsState,
  isGatewayLogsStale,
} from "./gateway-logs-state.ts"

test("applies incremental gateway log responses without replacing existing lines", () => {
  const initial = {
    ...createGatewayLogsState(),
    logs: ["first"],
    logOffset: 1,
    logRunID: 10,
  }

  assert.deepEqual(
    applyGatewayLogsResponse(initial, {
      logs: ["second", "third"],
      log_total: 3,
      log_run_id: 10,
    }),
    {
      logs: ["first", "second", "third"],
      logOffset: 3,
      logRunID: 10,
      error: null,
      lastUpdatedAt: 0,
    },
  )
})

test("replaces the visible gateway log buffer when the run id changes", () => {
  const initial = {
    ...createGatewayLogsState(),
    logs: ["old run"],
    logOffset: 1,
    logRunID: 10,
  }

  assert.deepEqual(
    applyGatewayLogsResponse(initial, {
      logs: ["new run"],
      log_total: 1,
      log_run_id: 11,
    }),
    {
      logs: ["new run"],
      logOffset: 1,
      logRunID: 11,
      error: null,
      lastUpdatedAt: 0,
    },
  )
})

test("classifies active gateway states as pollable", () => {
  assert.equal(canPollGatewayLogs("running"), true)
  assert.equal(canPollGatewayLogs("restarting"), true)
  assert.equal(canPollGatewayLogs("stopped"), false)
  assert.equal(canPollGatewayLogs("error"), false)
})

test("marks gateway logs stale after the configured threshold", () => {
  const state = {
    ...createGatewayLogsState(),
    lastUpdatedAt: 1000,
  }

  assert.equal(isGatewayLogsStale(state, 1600, 500), true)
  assert.equal(isGatewayLogsStale(state, 1400, 500), false)
  assert.equal(isGatewayLogsStale(createGatewayLogsState(), 1600, 500), false)
})
