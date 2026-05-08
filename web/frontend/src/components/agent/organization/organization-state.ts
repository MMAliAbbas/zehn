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
