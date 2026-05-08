import type { GatewayLogsResponse } from "@/api/gateway"
import type { GatewayState } from "@/store/gateway"

const POLLABLE_GATEWAY_STATES = new Set<GatewayState>([
  "running",
  "starting",
  "restarting",
  "stopping",
])

export const GATEWAY_LOGS_STALE_AFTER_MS = 5000

export interface GatewayLogsState {
  logs: string[]
  logOffset: number
  logRunID: number
  error: string | null
  lastUpdatedAt: number
}

export function createGatewayLogsState(): GatewayLogsState {
  return {
    logs: [],
    logOffset: 0,
    logRunID: -1,
    error: null,
    lastUpdatedAt: 0,
  }
}

export function canPollGatewayLogs(status: GatewayState): boolean {
  return POLLABLE_GATEWAY_STATES.has(status)
}

export function isGatewayLogsStale(
  state: Pick<GatewayLogsState, "lastUpdatedAt">,
  now: number,
  thresholdMs = GATEWAY_LOGS_STALE_AFTER_MS,
): boolean {
  return state.lastUpdatedAt > 0 && now - state.lastUpdatedAt > thresholdMs
}

export function applyGatewayLogsResponse(
  state: GatewayLogsState,
  data: GatewayLogsResponse,
  now = state.lastUpdatedAt,
): GatewayLogsState {
  const nextRunID = data.log_run_id ?? state.logRunID
  const nextLogs = data.logs ?? []

  if (nextRunID !== state.logRunID) {
    return {
      logs: nextLogs,
      logOffset: data.log_total ?? nextLogs.length,
      logRunID: nextRunID,
      error: null,
      lastUpdatedAt: now,
    }
  }

  if (nextLogs.length === 0) {
    return {
      ...state,
      error: null,
      lastUpdatedAt: now,
    }
  }

  return {
    logs: [...state.logs, ...nextLogs],
    logOffset: data.log_total ?? state.logOffset + nextLogs.length,
    logRunID: state.logRunID,
    error: null,
    lastUpdatedAt: now,
  }
}

export function applyGatewayLogsError(
  state: GatewayLogsState,
  error: unknown,
): GatewayLogsState {
  return {
    ...state,
    error: error instanceof Error ? error.message : "Failed to load logs",
  }
}
