import {
  IconAlertTriangle,
  IconCalendarStats,
  IconClock,
  IconInbox,
  IconInfoCircle,
  IconSend,
} from "@tabler/icons-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import type { AgentOrganizationAgent } from "@/api/agents"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardAction,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

import { AgentDetailSheet } from "./agent-detail-sheet"
import { displayAgentName, summarizeActivity } from "./formatting"
import { CountPill, EmptyActivityLine, StatusBadge } from "./status-components"

export function AgentCard({ agent }: { agent: AgentOrganizationAgent }) {
  const { t } = useTranslation()
  const [detailOpen, setDetailOpen] = useState(false)
  const displayName = displayAgentName(agent)
  const activity = summarizeActivity(agent.activity.current, t)
  const counts = [
    {
      key: "inbox",
      icon: IconInbox,
      label: t("pages.agent.organization.inbox", "Inbox"),
      value: agent.activity.inbox_count,
    },
    {
      key: "outbox",
      icon: IconSend,
      label: t("pages.agent.organization.outbox", "Outbox"),
      value: agent.activity.outbox_count,
    },
    {
      key: "meetings",
      icon: IconCalendarStats,
      label: t("pages.agent.organization.meetings", "Meetings"),
      value: agent.activity.meeting_count,
    },
    {
      key: "errors",
      icon: IconAlertTriangle,
      label: t("pages.agent.organization.errors", "Errors"),
      value: agent.activity.failure_count,
    },
  ].filter((item) => item.value > 0)

  return (
    <>
      <Card size="sm" className="min-w-0 rounded-lg py-3">
        <CardHeader className="grid-cols-[minmax(0,1fr)_auto] gap-3 px-3">
          <CardTitle className="min-w-0">
            <div className="truncate text-sm leading-5" title={displayName}>
              {displayName}
            </div>
            <div
              className="text-muted-foreground truncate font-mono text-[11px] leading-4"
              title={agent.id}
            >
              {agent.id}
            </div>
          </CardTitle>
          <CardAction className="flex items-center gap-1.5">
            <StatusBadge status={agent.status} />
            <Button
              type="button"
              variant="outline"
              size="xs"
              onClick={() => setDetailOpen(true)}
            >
              <IconInfoCircle />
              {t("pages.agent.organization.details", "Details")}
            </Button>
          </CardAction>
        </CardHeader>
        <CardContent className="space-y-2 px-3">
          <div
            className="text-muted-foreground flex min-w-0 items-center gap-1.5 text-xs"
            title={activity}
          >
            <IconClock className="size-3.5 shrink-0" />
            <span className="truncate">{activity}</span>
          </div>
          {counts.length > 0 ? (
            <div className="flex min-w-0 flex-wrap gap-1.5">
              {counts.map((item) => (
                <CountPill
                  key={item.key}
                  icon={item.icon}
                  label={item.label}
                  value={item.value}
                />
              ))}
            </div>
          ) : (
            <EmptyActivityLine
              label={t(
                "pages.agent.organization.no_activity",
                "No active records",
              )}
            />
          )}
        </CardContent>
      </Card>
      <AgentDetailSheet
        agent={agent}
        open={detailOpen}
        onOpenChange={setDetailOpen}
      />
    </>
  )
}
