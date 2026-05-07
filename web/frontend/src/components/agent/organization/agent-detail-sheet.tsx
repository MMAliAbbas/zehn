import { useQuery } from "@tanstack/react-query"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import type { AgentOrganizationAgent } from "@/api/agents"
import { getAgentInbox, getAgentMeetings, getAgentOutbox } from "@/api/agents"
import { Button } from "@/components/ui/button"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"

import {
  AGENT_DETAIL_LIMIT,
  AGENT_DETAIL_REFRESH_INTERVAL_MS,
} from "./constants"
import {
  AgentOverviewPanel,
  DelegationRecordsPanel,
  MeetingRecordsPanel,
  RecentEventsPanel,
} from "./detail-panels"
import { displayAgentName } from "./formatting"
import type { AgentDetailTab } from "./types"

export function AgentDetailSheet({
  agent,
  open,
  onOpenChange,
}: {
  agent: AgentOrganizationAgent
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  const { t } = useTranslation()
  const [activeTab, setActiveTab] = useState<AgentDetailTab>("overview")
  const displayName = displayAgentName(agent)

  const inboxQuery = useQuery({
    queryKey: ["agents", agent.id, "inbox", AGENT_DETAIL_LIMIT],
    queryFn: () => getAgentInbox(agent.id, AGENT_DETAIL_LIMIT),
    enabled: open && activeTab === "inbox",
    refetchInterval:
      open && activeTab === "inbox" ? AGENT_DETAIL_REFRESH_INTERVAL_MS : false,
  })
  const outboxQuery = useQuery({
    queryKey: ["agents", agent.id, "outbox", AGENT_DETAIL_LIMIT],
    queryFn: () => getAgentOutbox(agent.id, AGENT_DETAIL_LIMIT),
    enabled: open && activeTab === "outbox",
    refetchInterval:
      open && activeTab === "outbox" ? AGENT_DETAIL_REFRESH_INTERVAL_MS : false,
  })
  const meetingsQuery = useQuery({
    queryKey: ["agents", agent.id, "meetings", AGENT_DETAIL_LIMIT],
    queryFn: () => getAgentMeetings(agent.id, AGENT_DETAIL_LIMIT),
    enabled: open && activeTab === "meetings",
    refetchInterval:
      open && activeTab === "meetings"
        ? AGENT_DETAIL_REFRESH_INTERVAL_MS
        : false,
  })

  const tabs: Array<{ key: AgentDetailTab; label: string; count?: number }> = [
    {
      key: "overview",
      label: t("pages.agent.organization.detail.overview", "Overview"),
    },
    {
      key: "inbox",
      label: t("pages.agent.organization.inbox", "Inbox"),
      count: agent.activity.inbox_count,
    },
    {
      key: "outbox",
      label: t("pages.agent.organization.outbox", "Outbox"),
      count: agent.activity.outbox_count,
    },
    {
      key: "meetings",
      label: t("pages.agent.organization.meetings", "Meetings"),
      count: agent.activity.meeting_count,
    },
    {
      key: "recent",
      label: t("pages.agent.organization.detail.recent", "Recent Events"),
    },
  ]

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

        <div className="border-border/70 overflow-x-auto border-b px-4 py-2">
          <div
            className="flex min-w-max gap-1"
            role="tablist"
            aria-label={t(
              "pages.agent.organization.detail.tabs_label",
              "Agent activity sections",
            )}
          >
            {tabs.map((tab) => (
              <Button
                key={tab.key}
                type="button"
                variant={activeTab === tab.key ? "secondary" : "ghost"}
                size="sm"
                role="tab"
                aria-selected={activeTab === tab.key}
                onClick={() => setActiveTab(tab.key)}
              >
                {tab.label}
                {typeof tab.count === "number" && tab.count > 0 ? (
                  <span className="text-muted-foreground tabular-nums">
                    {tab.count}
                  </span>
                ) : null}
              </Button>
            ))}
          </div>
        </div>

        <div className="min-h-0 flex-1 overflow-auto px-4 py-4">
          {activeTab === "overview" ? (
            <AgentOverviewPanel agent={agent} />
          ) : activeTab === "inbox" ? (
            <DelegationRecordsPanel
              agentID={agent.id}
              label={t("pages.agent.organization.inbox", "Inbox")}
              query={inboxQuery}
            />
          ) : activeTab === "outbox" ? (
            <DelegationRecordsPanel
              agentID={agent.id}
              label={t("pages.agent.organization.outbox", "Outbox")}
              query={outboxQuery}
            />
          ) : activeTab === "meetings" ? (
            <MeetingRecordsPanel query={meetingsQuery} />
          ) : (
            <RecentEventsPanel agent={agent} />
          )}
        </div>
      </SheetContent>
    </Sheet>
  )
}
