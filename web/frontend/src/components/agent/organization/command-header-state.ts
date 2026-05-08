import type { AgentOrganizationSnapshot } from "@/api/agents"

export type OrganizationHeaderMode = "hierarchy" | "flat"

export type OrganizationHeaderRefreshState =
  | "loading"
  | "refreshing"
  | "stale"
  | "live"

export interface OrganizationHeaderActivitySummary {
  activeWork: number
  delegations: number
  meetings: number
  failures: number
  mode: OrganizationHeaderMode
}

export function summarizeOrganizationHeaderActivity(
  snapshot: AgentOrganizationSnapshot | undefined,
): OrganizationHeaderActivitySummary {
  const activity = snapshot?.activity

  return {
    activeWork: activity?.active_count ?? 0,
    delegations: activity?.delegation_count ?? 0,
    meetings: activity?.meeting_count ?? 0,
    failures: activity?.failure_count ?? 0,
    mode: snapshot?.metadata?.has_hierarchy === true ? "hierarchy" : "flat",
  }
}

export function resolveOrganizationHeaderRefreshState({
  hasData,
  isError,
  isFetching,
}: {
  hasData: boolean
  isError: boolean
  isFetching: boolean
}): OrganizationHeaderRefreshState {
  if (!hasData && isFetching) {
    return "loading"
  }
  if (hasData && isError) {
    return "stale"
  }
  if (isFetching) {
    return "refreshing"
  }
  return "live"
}
