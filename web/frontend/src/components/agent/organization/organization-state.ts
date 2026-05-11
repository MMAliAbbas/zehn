import type {
  AgentOrganizationAgent,
  AgentOrganizationNode,
  AgentOrganizationSnapshot,
} from "@/api/agents"

import type {
  AgentActivityShortcut,
  AgentDetailTab,
  AgentSelectedActivityRecord,
  AgentWorkbenchSection,
  OrganizationSelectionState,
} from "./types"

export const DEFAULT_WORKBENCH_SECTION: AgentWorkbenchSection = "overview"
export type AgentCardShortcut = "details" | AgentActivityShortcut

export function createOrganizationSelectionState(): OrganizationSelectionState {
  return {
    selectedAgentID: null,
    workbenchSection: DEFAULT_WORKBENCH_SECTION,
    selectedRecord: null,
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
    selectedRecord: null,
  }
}

export function selectOrganizationActivityRecord(
  current: OrganizationSelectionState,
  record: AgentSelectedActivityRecord,
): OrganizationSelectionState {
  return {
    ...current,
    workbenchSection: record.sourceSection,
    selectedRecord: record,
  }
}

export function clearSelectedOrganizationRecord(
  current: OrganizationSelectionState,
): OrganizationSelectionState {
  return {
    ...current,
    selectedRecord: null,
  }
}

export function selectOrganizationWorkbenchSection(
  current: OrganizationSelectionState,
  section: AgentWorkbenchSection,
): OrganizationSelectionState {
  return {
    ...current,
    workbenchSection: section,
    selectedRecord: section === "live-logs" ? current.selectedRecord : null,
  }
}

export function resolveSelectableActivityRecord(
  record: AgentSelectedActivityRecord,
  detailAvailable?: boolean,
): AgentSelectedActivityRecord | null {
  return detailAvailable === true ? record : null
}

export function resolveActivityShortcut(shortcut: AgentActivityShortcut): {
  workbenchSection: AgentWorkbenchSection
  detailTab: AgentDetailTab
} {
  if (shortcut === "errors") {
    return {
      workbenchSection: "failures",
      detailTab: "failures",
    }
  }

  return {
    workbenchSection: shortcut,
    detailTab: shortcut,
  }
}

export function resolveAgentCardShortcut(shortcut: AgentCardShortcut): {
  workbenchSection: AgentWorkbenchSection
  detailTab: AgentDetailTab
} {
  if (shortcut === "details") {
    return {
      workbenchSection: "overview",
      detailTab: "overview",
    }
  }

  return resolveActivityShortcut(shortcut)
}

export function detailTabForWorkbenchSection(
  section: AgentWorkbenchSection,
): AgentDetailTab {
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
