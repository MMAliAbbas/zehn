import { useQuery } from "@tanstack/react-query"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import type { AgentOrganizationAgent } from "@/api/agents"
import {
  getAgentFailures,
  getAgentInbox,
  getAgentMeetings,
  getAgentOutbox,
} from "@/api/agents"
import { Button } from "@/components/ui/button"

import {
  AGENT_DETAIL_LIMIT,
  AGENT_DETAIL_REFRESH_INTERVAL_MS,
} from "./constants"
import {
  AgentOverviewPanel,
  DelegationRecordsPanel,
  FailureRecordsPanel,
  ActivityRecordDetailPanel,
  LiveLogsPanel,
  MeetingRecordsPanel,
  RecentEventsPanel,
} from "./detail-panels"
import { detailTabForWorkbenchSection } from "./organization-state"
import type {
  AgentDetailTab,
  AgentSelectedActivityRecord,
  AgentWorkbenchSection,
} from "./types"

export function AgentDetailContent({
  agent,
  activeSection,
  enabled,
  selectedRecord,
  onSectionChange,
  onSelectedRecordChange,
}: {
  agent: AgentOrganizationAgent
  activeSection: AgentWorkbenchSection
  enabled: boolean
  selectedRecord: AgentSelectedActivityRecord | null
  onSectionChange: (section: AgentWorkbenchSection) => void
  onSelectedRecordChange: (record: AgentSelectedActivityRecord | null) => void
}) {
  const { t } = useTranslation()
  const activeTab = detailTabForWorkbenchSection(activeSection)

  const inboxQuery = useQuery({
    queryKey: ["agents", agent.id, "inbox", AGENT_DETAIL_LIMIT],
    queryFn: () => getAgentInbox(agent.id, AGENT_DETAIL_LIMIT),
    enabled: enabled && activeTab === "inbox",
    refetchInterval:
      enabled && activeTab === "inbox"
        ? AGENT_DETAIL_REFRESH_INTERVAL_MS
        : false,
  })
  const outboxQuery = useQuery({
    queryKey: ["agents", agent.id, "outbox", AGENT_DETAIL_LIMIT],
    queryFn: () => getAgentOutbox(agent.id, AGENT_DETAIL_LIMIT),
    enabled: enabled && activeTab === "outbox",
    refetchInterval:
      enabled && activeTab === "outbox"
        ? AGENT_DETAIL_REFRESH_INTERVAL_MS
        : false,
  })
  const meetingsQuery = useQuery({
    queryKey: ["agents", agent.id, "meetings", AGENT_DETAIL_LIMIT],
    queryFn: () => getAgentMeetings(agent.id, AGENT_DETAIL_LIMIT),
    enabled: enabled && activeTab === "meetings",
    refetchInterval:
      enabled && activeTab === "meetings"
        ? AGENT_DETAIL_REFRESH_INTERVAL_MS
        : false,
  })
  const failuresQuery = useQuery({
    queryKey: ["agents", agent.id, "failures", AGENT_DETAIL_LIMIT],
    queryFn: () => getAgentFailures(agent.id, AGENT_DETAIL_LIMIT),
    enabled: enabled && activeTab === "failures",
    refetchInterval:
      enabled && activeTab === "failures"
        ? AGENT_DETAIL_REFRESH_INTERVAL_MS
        : false,
  })

  const tabs = agentDetailTabs(agent, t)

  return (
    <>
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
              onClick={() => onSectionChange(tab.key)}
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
        <ActivityRecordDetailPanel
          agentID={agent.id}
          selectedRecord={selectedRecord}
          onClear={() => onSelectedRecordChange(null)}
        />
        {activeTab === "overview" ? (
          <AgentOverviewPanel agent={agent} />
        ) : activeTab === "inbox" ? (
          <DelegationRecordsPanel
            agentID={agent.id}
            label={t("pages.agent.organization.inbox", "Inbox")}
            query={inboxQuery}
            sourceSection="inbox"
            onSelectRecord={onSelectedRecordChange}
          />
        ) : activeTab === "outbox" ? (
          <DelegationRecordsPanel
            agentID={agent.id}
            label={t("pages.agent.organization.outbox", "Outbox")}
            query={outboxQuery}
            sourceSection="outbox"
            onSelectRecord={onSelectedRecordChange}
          />
        ) : activeTab === "meetings" ? (
          <MeetingRecordsPanel
            query={meetingsQuery}
            onSelectRecord={onSelectedRecordChange}
          />
        ) : activeTab === "failures" ? (
          <FailureRecordsPanel
            agent={agent}
            query={failuresQuery}
            onSelectRecord={onSelectedRecordChange}
          />
        ) : activeTab === "live-logs" ? (
          <LiveLogsPanel agent={agent} />
        ) : (
          <RecentEventsPanel agent={agent} />
        )}
      </div>
    </>
  )
}

export function StatefulAgentDetailContent({
  agent,
  enabled,
  initialTab = "overview",
}: {
  agent: AgentOrganizationAgent
  enabled: boolean
  initialTab?: AgentDetailTab
}) {
  const [activeSection, setActiveSection] =
    useState<AgentWorkbenchSection>(initialTab)
  const [selectedRecord, setSelectedRecord] =
    useState<AgentSelectedActivityRecord | null>(null)

  useEffect(() => {
    if (enabled) {
      setActiveSection(initialTab)
      setSelectedRecord(null)
    }
  }, [agent.id, enabled, initialTab])

  const handleSectionChange = (section: AgentWorkbenchSection) => {
    setActiveSection(section)
    setSelectedRecord(null)
  }

  const handleSelectedRecordChange = (
    record: AgentSelectedActivityRecord | null,
  ) => {
    setSelectedRecord(record)
    if (record) {
      setActiveSection(record.sourceSection)
    }
  }

  return (
    <AgentDetailContent
      agent={agent}
      activeSection={activeSection}
      enabled={enabled}
      selectedRecord={selectedRecord}
      onSectionChange={handleSectionChange}
      onSelectedRecordChange={handleSelectedRecordChange}
    />
  )
}

function agentDetailTabs(
  agent: AgentOrganizationAgent,
  t: ReturnType<typeof useTranslation>["t"],
) {
  return [
    {
      key: "overview" as const,
      label: t("pages.agent.organization.detail.overview", "Overview"),
    },
    {
      key: "inbox" as const,
      label: t("pages.agent.organization.inbox", "Inbox"),
      count: agent.activity.inbox_count,
    },
    {
      key: "outbox" as const,
      label: t("pages.agent.organization.outbox", "Outbox"),
      count: agent.activity.outbox_count,
    },
    {
      key: "meetings" as const,
      label: t("pages.agent.organization.meetings", "Meetings"),
      count: agent.activity.meeting_count,
    },
    {
      key: "failures" as const,
      label: t("pages.agent.organization.detail.failures", "Failures"),
      count: agent.activity.failure_count,
    },
    {
      key: "recent" as const,
      label: t("pages.agent.organization.detail.recent", "Recent Events"),
    },
    {
      key: "live-logs" as const,
      label: t("pages.agent.organization.detail.live_logs", "Live Logs"),
    },
  ]
}
