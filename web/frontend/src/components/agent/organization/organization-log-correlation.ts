import { findAgentLogReferenceFields } from "../../../lib/agent-log-filter.ts"

export type OrganizationLogCorrelationMode =
  | "all"
  | "selected-agent"
  | "selected-record"

export interface OrganizationLogCorrelationTarget {
  selectedAgentID: string
  selectedRecordID?: string
  peerAgentIDs?: string[]
}

const ANSI_PATTERN = new RegExp(String.raw`\u001B\[[0-9;]*m`, "g")

function stripAnsi(input: string): string {
  return input.replace(ANSI_PATTERN, "")
}

function addFields(target: Set<string>, fields: string[]) {
  fields.forEach((field) => target.add(field))
}

export function findOrganizationLogCorrelationFields(
  line: string,
  target: OrganizationLogCorrelationTarget,
): string[] {
  const matches = new Set<string>()

  addFields(matches, findAgentLogReferenceFields(line, target.selectedAgentID))

  if (
    target.selectedRecordID &&
    stripAnsi(line).includes(target.selectedRecordID)
  ) {
    matches.add("record_id")
  }

  for (const peerAgentID of target.peerAgentIDs ?? []) {
    addFields(matches, findAgentLogReferenceFields(line, peerAgentID))
  }

  return Array.from(matches)
}

export function filterOrganizationLogLines(
  logs: string[],
  mode: OrganizationLogCorrelationMode,
  target: OrganizationLogCorrelationTarget,
): string[] {
  if (mode === "all") {
    return logs
  }

  const scopedTarget =
    mode === "selected-agent"
      ? { selectedAgentID: target.selectedAgentID }
      : target

  return logs.filter(
    (line) =>
      findOrganizationLogCorrelationFields(line, scopedTarget).length > 0,
  )
}
