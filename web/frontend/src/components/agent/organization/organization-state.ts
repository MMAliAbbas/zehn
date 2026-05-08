import type {
  AgentOrganizationAgent,
  AgentOrganizationNode,
  AgentOrganizationSnapshot,
} from "@/api/agents"

import type {
  AgentActivityShortcut,
  AgentDetailTab,
  AgentWorkbenchSection,
  OrganizationSelectionState,
} from "./types"

export const DEFAULT_WORKBENCH_SECTION: AgentWorkbenchSection = "overview"

export function createOrganizationSelectionState(): OrganizationSelectionState {
  return {
    selectedAgentID: null,
    workbenchSection: DEFAULT_WORKBENCH_SECTION,
  }
}

export function selectOrganizationAgent(
  current: OrganizationSelectionState,
  agentID: string,
  section: AgentWorkbenchSection = current.workbenchSection,
): OrganizationSelectionState {
  return {
    selectedAgentID: agentID,
    workbenchSection: section,
  }
}

export function resolveActivityShortcut(
  shortcut: AgentActivityShortcut,
): {
  workbenchSection: AgentWorkbenchSection
  detailTab: AgentDetailTab
} {
  if (shortcut === "errors") {
    return {
      workbenchSection: "failures",
      detailTab: "recent",
    }
  }

  return {
    workbenchSection: shortcut,
    detailTab: shortcut,
  }
}

export function detailTabForWorkbenchSection(
  section: AgentWorkbenchSection,
): AgentDetailTab {
  if (section === "failures" || section === "live-logs") {
    return "recent"
  }

  return section
}

export function resolveSelectedOrganizationAgent(
  snapshot: AgentOrganizationSnapshot | undefined,
  selectedAgentID: string | null,
): AgentOrganizationAgent | null {
  if (!snapshot || !selectedAgentID) {
    return null
  }

  const indexedAgent = snapshot.agents?.[selectedAgentID]
  if (indexedAgent) {
    return indexedAgent
  }

  for (const root of snapshot.roots ?? []) {
    const match = findAgentInTree(root, selectedAgentID)
    if (match) {
      return match
    }
  }

  return null
}

function findAgentInTree(
  node: AgentOrganizationNode,
  agentID: string,
): AgentOrganizationAgent | null {
  if (node.id === agentID) {
    return node
  }

  for (const child of node.children ?? []) {
    const match = findAgentInTree(child, agentID)
    if (match) {
      return match
    }
  }

  return null
}
