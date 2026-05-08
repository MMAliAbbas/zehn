import type { AgentOrganizationNode } from "@/api/agents"

export interface OrderedNode extends AgentOrganizationNode {
  children?: OrderedNode[]
}

export type AgentWorkbenchSection =
  | "overview"
  | "inbox"
  | "outbox"
  | "meetings"
  | "failures"
  | "recent"
  | "live-logs"

export const AGENT_WORKBENCH_SECTIONS = [
  "overview",
  "inbox",
  "outbox",
  "meetings",
  "failures",
  "recent",
  "live-logs",
] as const satisfies readonly AgentWorkbenchSection[]

export type AgentDetailTab = Exclude<
  AgentWorkbenchSection,
  "failures" | "live-logs"
>

export interface OrganizationSelectionState {
  selectedAgentID: string | null
  workbenchSection: AgentWorkbenchSection
}
