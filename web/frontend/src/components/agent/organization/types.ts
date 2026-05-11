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

export type AgentDetailTab = AgentWorkbenchSection

export type AgentActivityShortcut = "inbox" | "outbox" | "meetings" | "errors"

export interface AgentSelectedActivityRecord {
  type: string
  recordID: string
  sourceSection: AgentWorkbenchSection
  title?: string
  peerAgentIDs?: string[]
}

export interface OrganizationSelectionState {
  selectedAgentID: string | null
  workbenchSection: AgentWorkbenchSection
  selectedRecord: AgentSelectedActivityRecord | null
}
