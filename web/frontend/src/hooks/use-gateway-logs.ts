import { useAtomValue } from "jotai"
import { useEffect, useRef, useState } from "react"

import { clearGatewayLogs, getGatewayLogs } from "@/api/gateway"
import { gatewayAtom } from "@/store/gateway"

import {
  applyGatewayLogsError,
  applyGatewayLogsResponse,
  canPollGatewayLogs,
  createGatewayLogsState,
  isGatewayLogsStale,
} from "./gateway-logs-state"

export function useGatewayLogs() {
  const [logState, setLogState] = useState(createGatewayLogsState)
  const [now, setNow] = useState(() => Date.now())
  const [clearing, setClearing] = useState(false)
  const logStateRef = useRef(logState)
  const syncTokenRef = useRef(0)

  const gateway = useAtomValue(gatewayAtom)

  useEffect(() => {
    logStateRef.current = logState
  }, [logState])

  useEffect(() => {
    const interval = setInterval(() => {
      setNow(Date.now())
    }, 1000)

    return () => {
      clearInterval(interval)
    }
  }, [])

  const clearLogs = async () => {
    setClearing(true)
    try {
      const data = await clearGatewayLogs()
      syncTokenRef.current += 1
      setLogState((current) => ({
        ...current,
        logs: [],
        logOffset: data.log_total ?? 0,
        logRunID: data.log_run_id ?? current.logRunID,
        error: null,
        lastUpdatedAt: Date.now(),
      }))
    } catch {
      // Ignore clear failures silently to avoid noisy transient errors.
    } finally {
      setClearing(false)
    }
  }

  useEffect(() => {
    let mounted = true
    let timeout: ReturnType<typeof setTimeout>

    const fetchLogs = async () => {
      if (!mounted || !canPollGatewayLogs(gateway.status)) {
        if (mounted) {
          timeout = setTimeout(fetchLogs, 1000)
        }
        return
      }

      try {
        const requestToken = syncTokenRef.current
        const requestOffset = logStateRef.current.logOffset
        const requestRunId = logStateRef.current.logRunID
        const data = await getGatewayLogs({
          log_offset: requestOffset,
          log_run_id: requestRunId,
        })

        if (!mounted || requestToken !== syncTokenRef.current) {
          return
        }

        const receivedAt = Date.now()
        setNow(receivedAt)
        setLogState((current) =>
          applyGatewayLogsResponse(current, data, receivedAt),
        )
      } catch (error) {
        if (mounted) {
          setLogState((current) => applyGatewayLogsError(current, error))
        }
      } finally {
        if (mounted) {
          timeout = setTimeout(fetchLogs, 1000)
        }
      }
    }

    fetchLogs()

    return () => {
      mounted = false
      clearTimeout(timeout)
    }
  }, [gateway.status])

  return {
    clearLogs,
    clearing,
    error: logState.error,
    gatewayStatus: gateway.status,
    logs: logState.logs,
    stale: isGatewayLogsStale(logState, now),
  }
}
