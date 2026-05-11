import type { TFunction } from "i18next"

import type {
  AgentOrganizationActivityRecord,
  AgentOrganizationAgent,
  AgentOrganizationNode,
  AgentOrganizationSnapshot,
} from "@/api/agents"

import type { OrderedNode } from "./types"

export function buildOrderedRoots(
  snapshot: AgentOrganizationSnapshot | undefined,
): OrderedNode[] {
  if (!snapshot) {
    return []
  }
  if ((snapshot.roots?.length ?? 0) > 0) {
    return preserveNodeOrder(snapshot.roots ?? [])
  }
  return Object.values(snapshot.agents ?? {})
    .sort(compareAgents)
    .map((agent) => ({ ...agent, children: [] }))
}

function preserveNodeOrder(nodes: AgentOrganizationNode[]): OrderedNode[] {
  return nodes.map((node) => ({
    ...node,
    children: node.children ? preserveNodeOrder(node.children) : [],
  }))
}

function compareAgents(
  a: Pick<AgentOrganizationAgent, "id" | "label" | "name">,
  b: Pick<AgentOrganizationAgent, "id" | "label" | "name">,
) {
  return (
    displayAgentName(a).localeCompare(displayAgentName(b), undefined, {
      sensitivity: "base",
      numeric: true,
    }) || a.id.localeCompare(b.id, undefined, { sensitivity: "base" })
  )
}

export function displayAgentName(
  agent: Pick<AgentOrganizationAgent, "id" | "label" | "name">,
) {
  return agent.label?.trim() || agent.name?.trim() || agent.id
}

export function summarizeActivity(
  current: AgentOrganizationActivityRecord | undefined,
  t: TFunction,
) {
  if (!current) {
    return t("pages.agent.organization.idle_summary", "Idle")
  }
  const type = t(
    `pages.agent.organization.activity_type.${current.type}`,
    current.type,
  )
  const role = current.role
    ? t(`pages.agent.organization.role.${current.role}`, current.role)
    : ""
  const status = t(
    `pages.agent.organization.status.${current.status}`,
    current.status,
  )
  return [type, role, status].filter(Boolean).join(" / ")
}

export function formatDiagnosticReason(
  record: Pick<AgentOrganizationActivityRecord, "reason">,
  t: TFunction,
) {
  return (
    record.reason?.trim() ||
    t(
      "pages.agent.organization.detail.unknown_failure_reason",
      "No diagnostic reason available",
    )
  )
}

export function formatDiagnosticReasonSource(
  record: Pick<AgentOrganizationActivityRecord, "reason_source">,
  t: TFunction,
) {
  const source = record.reason_source?.trim()
  if (!source) {
    return t(
      "pages.agent.organization.detail.unknown_reason_source",
      "Unknown source",
    )
  }
  return t(
    `pages.agent.organization.detail.reason_source.${source}`,
    humanizeDiagnosticCode(source),
  )
}

export function formatDiagnosticSeverity(
  record: Pick<AgentOrganizationActivityRecord, "severity">,
  t: TFunction,
) {
  const severity = record.severity?.trim()
  if (!severity) {
    return t("pages.agent.organization.detail.unknown_severity", "Unknown")
  }
  return t(
    `pages.agent.organization.detail.severity.${severity}`,
    humanizeDiagnosticCode(severity),
  )
}

export function formatDiagnosticFreshness(
  record: Pick<AgentOrganizationActivityRecord, "current" | "stale">,
  current: boolean,
  t: TFunction,
) {
  if (record.stale) {
    return t("pages.agent.organization.detail.stale_diagnostic", "Stale")
  }
  if (record.current || current) {
    return t("pages.agent.organization.detail.current_diagnostic", "Current")
  }
  return t("pages.agent.organization.detail.stale_diagnostic", "Stale")
}

export function compactActivityEvents(
  current: AgentOrganizationActivityRecord | undefined,
  lastFailure: AgentOrganizationActivityRecord | undefined,
) {
  const records = [current, lastFailure].filter(
    Boolean,
  ) as AgentOrganizationActivityRecord[]
  const seen = new Set<string>()
  return records.filter((record) => {
    const key = `${record.type}:${record.record_id}`
    if (seen.has(key)) {
      return false
    }
    seen.add(key)
    return true
  })
}

export function buildFailureDrilldownRecords(
  current: AgentOrganizationActivityRecord | undefined,
  lastFailure: AgentOrganizationActivityRecord | undefined,
  recentFailures: AgentOrganizationActivityRecord[] | undefined,
) {
  if (recentFailures) {
    return dedupeActivityRecords(recentFailures)
  }
  const currentIsLastFailure =
    current?.type === lastFailure?.type &&
    current?.record_id === lastFailure?.record_id
  return compactActivityEvents(
    currentIsLastFailure ? current : undefined,
    lastFailure,
  )
}

function dedupeActivityRecords(records: AgentOrganizationActivityRecord[]) {
  const seen = new Set<string>()
  return records.filter((record) => {
    const key = `${record.type}:${record.record_id}`
    if (seen.has(key)) {
      return false
    }
    seen.add(key)
    return true
  })
}

export function formatTimestamp(value: string | undefined, t: TFunction) {
  if (!value) {
    return t("common.notAvailable", "Unavailable")
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date)
}

export function shortRecordID(id: string) {
  if (id.length <= 12) {
    return id
  }
  return id.slice(0, 12)
}

export function isProblemStatus(status: string) {
  const normalized = status.toLowerCase()
  return (
    normalized === "failed" ||
    normalized === "blocked" ||
    normalized === "error" ||
    normalized === "fatal"
  )
}

export function errorMessage(error: unknown) {
  return error instanceof Error ? error.message : String(error)
}

function humanizeDiagnosticCode(value: string) {
  const words = value
    .split(/[_\s-]+/)
    .filter(Boolean)
  if (words.length === 0) {
    return value
  }
  return [
    words[0].charAt(0).toUpperCase() + words[0].slice(1).toLowerCase(),
    ...words.slice(1).map((part) => part.toLowerCase()),
  ].join(" ")
}
