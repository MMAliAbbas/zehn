import type { AgentOrganizationNode } from "@/api/agents"

export interface OrderedNode extends AgentOrganizationNode {
  children?: OrderedNode[]
}

export type AgentDetailTab =
  | "overview"
  | "inbox"
  | "outbox"
  | "meetings"
  | "recent"
