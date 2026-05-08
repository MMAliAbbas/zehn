import type { AgentWorkbenchSection, OrganizationSelectionState } from "./types"

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
