import {
  IconAlertTriangle,
  IconLoader2,
  IconNetwork,
} from "@tabler/icons-react"
import { useQuery } from "@tanstack/react-query"
import { useCallback, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"

import { getAgentOrganization } from "@/api/agents"
import { PageHeader } from "@/components/page-header"

import { ORGANIZATION_REFRESH_INTERVAL_MS } from "./constants"
import { buildOrderedRoots } from "./formatting"
import {
  createOrganizationSelectionState,
  selectOrganizationAgent,
} from "./organization-state"
import { OrganizationBranch } from "./organization-tree"
import { SnapshotSummary, StatePanel } from "./status-components"
import type { AgentWorkbenchSection } from "./types"

export function OrganizationPage() {
  const { t } = useTranslation()
  const [selection, setSelection] = useState(createOrganizationSelectionState)
  const organizationQuery = useQuery({
    queryKey: ["agents", "organization"],
    queryFn: getAgentOrganization,
    refetchInterval: ORGANIZATION_REFRESH_INTERVAL_MS,
  })

  const roots = useMemo(
    () => buildOrderedRoots(organizationQuery.data),
    [organizationQuery.data],
  )

  const handleSelectAgent = useCallback(
    (agentID: string, section?: AgentWorkbenchSection) => {
      setSelection((current) =>
        selectOrganizationAgent(current, agentID, section),
      )
    },
    [],
  )

  return (
    <div className="bg-background flex h-full flex-col">
      <PageHeader title={t("navigation.organization", "Organization")} />

      <div className="flex-1 overflow-auto px-6 py-6 pb-20">
        <div className="mx-auto w-full max-w-7xl space-y-4">
          {organizationQuery.isLoading && !organizationQuery.data ? (
            <StatePanel
              icon={<IconLoader2 className="size-4 animate-spin" />}
              title={t(
                "pages.agent.organization.loading",
                "Loading organization",
              )}
              detail={t(
                "pages.agent.organization.loading_detail",
                "Reading configured agents and current activity.",
              )}
            />
          ) : organizationQuery.error && !organizationQuery.data ? (
            <StatePanel
              icon={<IconAlertTriangle className="size-4" />}
              title={t(
                "pages.agent.organization.error",
                "Failed to load organization",
              )}
              detail={
                organizationQuery.error instanceof Error
                  ? organizationQuery.error.message
                  : t(
                      "pages.agent.organization.error_detail",
                      "The organization snapshot is unavailable.",
                    )
              }
              destructive
            />
          ) : roots.length === 0 ? (
            <StatePanel
              icon={<IconNetwork className="size-4" />}
              title={t(
                "pages.agent.organization.empty",
                "No configured agents",
              )}
              detail={t(
                "pages.agent.organization.empty_detail",
                "Add agents to the launcher configuration to populate this view.",
              )}
            />
          ) : (
            <section className="space-y-4">
              <SnapshotSummary snapshot={organizationQuery.data} />
              <div className="space-y-3">
                {roots.map((node) => (
                  <OrganizationBranch
                    key={node.id}
                    node={node}
                    depth={0}
                    selectedAgentID={selection.selectedAgentID}
                    onSelectAgent={handleSelectAgent}
                  />
                ))}
              </div>
            </section>
          )}
        </div>
      </div>
    </div>
  )
}
