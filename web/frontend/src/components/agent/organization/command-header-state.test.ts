/// <reference types="node" />
import assert from "node:assert/strict"

import test from "node:test"

import {
  resolveOrganizationHeaderRefreshState,
  summarizeOrganizationHeaderActivity,
} from "./command-header-state.ts"

test("summarizes organization command header activity from snapshot totals", () => {
  assert.deepEqual(
    summarizeOrganizationHeaderActivity({
      activity: {
        active_count: 3,
        delegation_count: 5,
        meeting_count: 2,
        failure_count: 1,
      },
      metadata: {
        source: "test",
        generated_at: "2026-05-09T00:00:00Z",
        has_hierarchy: true,
      },
    }),
    {
      activeWork: 3,
      delegations: 5,
      meetings: 2,
      failures: 1,
      mode: "hierarchy",
    },
  )
})

test("defaults missing organization command header totals to flat idle state", () => {
  assert.deepEqual(summarizeOrganizationHeaderActivity(undefined), {
    activeWork: 0,
    delegations: 0,
    meetings: 0,
    failures: 0,
    mode: "flat",
  })
})

test("classifies command header refresh state from organization query state", () => {
  assert.equal(
    resolveOrganizationHeaderRefreshState({
      hasData: false,
      isError: false,
      isFetching: true,
    }),
    "loading",
  )
  assert.equal(
    resolveOrganizationHeaderRefreshState({
      hasData: true,
      isError: false,
      isFetching: true,
    }),
    "refreshing",
  )
  assert.equal(
    resolveOrganizationHeaderRefreshState({
      hasData: true,
      isError: true,
      isFetching: false,
    }),
    "stale",
  )
  assert.equal(
    resolveOrganizationHeaderRefreshState({
      hasData: true,
      isError: false,
      isFetching: false,
    }),
    "live",
  )
})
