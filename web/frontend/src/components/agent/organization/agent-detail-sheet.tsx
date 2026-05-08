import type { AgentOrganizationAgent } from "@/api/agents"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"

import { StatefulAgentDetailContent } from "./agent-detail-content"
import { displayAgentName } from "./formatting"
import type { AgentDetailTab } from "./types"

export function AgentDetailSheet({
  agent,
  open,
  initialTab = "overview",
  onOpenChange,
}: {
  agent: AgentOrganizationAgent
  open: boolean
  initialTab?: AgentDetailTab
  onOpenChange: (open: boolean) => void
}) {
  const displayName = displayAgentName(agent)

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent
        className="w-full gap-0 p-0 sm:max-w-2xl"
        aria-describedby={`agent-detail-${agent.id}-description`}
      >
        <SheetHeader className="border-border/70 border-b px-4 py-4 pr-12">
          <SheetTitle className="min-w-0 pr-2">
            <span className="block truncate">{displayName}</span>
          </SheetTitle>
          <SheetDescription
            id={`agent-detail-${agent.id}-description`}
            className="min-w-0"
          >
            <span className="block truncate font-mono text-xs">{agent.id}</span>
          </SheetDescription>
        </SheetHeader>

        <StatefulAgentDetailContent
          agent={agent}
          enabled={open}
          initialTab={initialTab}
        />
      </SheetContent>
    </Sheet>
  )
}
