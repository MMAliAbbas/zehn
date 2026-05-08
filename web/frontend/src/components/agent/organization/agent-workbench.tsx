import { useTranslation } from "react-i18next"

import type { AgentOrganizationAgent } from "@/api/agents"

import { AgentDetailContent } from "./agent-detail-content"
import { displayAgentName } from "./formatting"
import type { AgentWorkbenchSection } from "./types"

export function AgentWorkbench({
  agent,
  activeSection,
  onSectionChange,
}: {
  agent: AgentOrganizationAgent | null
  activeSection: AgentWorkbenchSection
  onSectionChange: (section: AgentWorkbenchSection) => void
}) {
  const { t } = useTranslation()

  return (
    <aside className="border-border/70 bg-background sticky top-0 hidden max-h-[calc(100vh-8rem)] min-h-[32rem] min-w-0 flex-col overflow-hidden rounded-lg border lg:flex">
      {agent ? (
        <>
          <div className="border-border/70 border-b px-4 py-4">
            <div className="text-muted-foreground text-xs font-medium tracking-wide uppercase">
              {t(
                "pages.agent.organization.workbench.title",
                "Agent Workbench",
              )}
            </div>
            <h2 className="mt-1 truncate text-base font-medium">
              {displayAgentName(agent)}
            </h2>
            <div className="text-muted-foreground mt-1 truncate font-mono text-xs">
              {agent.id}
            </div>
          </div>

          <AgentDetailContent
            agent={agent}
            activeSection={activeSection}
            enabled
            onSectionChange={onSectionChange}
          />
        </>
      ) : (
        <div className="flex h-full min-h-[32rem] flex-col justify-center px-4 py-6">
          <div className="text-sm font-medium">
            {t(
              "pages.agent.organization.workbench.empty",
              "Select an agent",
            )}
          </div>
          <div className="text-muted-foreground mt-1 text-sm">
            {t(
              "pages.agent.organization.workbench.empty_detail",
              "Choose a card in the organization canvas to inspect activity without leaving the hierarchy.",
            )}
          </div>
        </div>
      )}
    </aside>
  )
}
