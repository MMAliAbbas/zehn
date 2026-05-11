import {
  IconAlertTriangle,
  IconCalendarStats,
  IconClock,
  IconInbox,
  IconInfoCircle,
  IconSend,
} from "@tabler/icons-react"
import { useEffect, useState } from "react"
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
import { cn } from "@/lib/utils"

import { AgentDetailSheet } from "./agent-detail-sheet"
import {
  displayAgentName,
  formatDiagnosticReason,
  isProblemStatus,
  summarizeActivity,
} from "./formatting"
import {
  type AgentCardShortcut,
  resolveAgentCardShortcut,
} from "./organization-state"
import { CountPill, EmptyActivityLine, StatusBadge } from "./status-components"
import type {
  AgentDetailTab,
  AgentWorkbenchSection,
} from "./types"

export function AgentCard({
  agent,
  selected,
  onSelect,
}: {
  agent: AgentOrganizationAgent
  selected: boolean
  onSelect: (agentID: string, section?: AgentWorkbenchSection) => void
}) {
  const { t } = useTranslation()
  const [detailOpen, setDetailOpen] = useState(false)
  const [detailInitialTab, setDetailInitialTab] =
    useState<AgentDetailTab>("overview")
  const displayName = displayAgentName(agent)
  const activity = summarizeActivity(agent.activity.current, t)
  const currentFailureReason =
    agent.activity.current &&
    isProblemStatus(agent.activity.current.status) &&
    agent.activity.current.reason?.trim()
      ? formatDiagnosticReason(agent.activity.current, t)
      : ""
  const desktopWorkbench = useDesktopWorkbenchLayout()
  const selectAgent = () => {
    onSelect(agent.id)
    if (!desktopWorkbench) {
      setDetailInitialTab("overview")
      setDetailOpen(true)
    }
  }
  const counts = [
    {
      key: "inbox" as const,
      icon: IconInbox,
      label: t("pages.agent.organization.inbox", "Inbox"),
      value: agent.activity.inbox_count,
    },
    {
      key: "outbox" as const,
      icon: IconSend,
      label: t("pages.agent.organization.outbox", "Outbox"),
      value: agent.activity.outbox_count,
    },
    {
      key: "meetings" as const,
      icon: IconCalendarStats,
      label: t("pages.agent.organization.meetings", "Meetings"),
      value: agent.activity.meeting_count,
    },
    {
      key: "errors" as const,
      icon: IconAlertTriangle,
      label: t("pages.agent.organization.errors", "Errors"),
      value: agent.activity.failure_count,
    },
  ].filter((item) => item.value > 0)

  const openCardShortcut = (shortcut: AgentCardShortcut) => {
    const target = resolveAgentCardShortcut(shortcut)
    onSelect(agent.id, target.workbenchSection)
    setDetailInitialTab(target.detailTab)
    if (!desktopWorkbench) {
      setDetailOpen(true)
    }
  }

  return (
    <>
      <Card
        size="sm"
        className={cn(
          "relative min-w-0 rounded-lg py-3 transition-colors",
          selected
            ? "bg-primary/5 ring-primary/40"
            : "has-focus-visible:ring-ring/50",
        )}
      >
        <button
          type="button"
          className={cn(
            "focus-visible:ring-ring/50 absolute inset-0 z-0 rounded-lg text-left transition-colors focus-visible:ring-[3px] focus-visible:outline-none",
            !selected && "hover:bg-muted/40",
          )}
          aria-pressed={selected}
          aria-label={t(
            "pages.agent.organization.select_agent",
            "Select {{agent}}",
            { agent: displayName },
          )}
          onClick={selectAgent}
        >
          <span className="sr-only">
            {t("pages.agent.organization.select_agent", "Select {{agent}}", {
              agent: displayName,
            })}
          </span>
        </button>
        <CardHeader className="pointer-events-none relative z-10 grid-cols-[minmax(0,1fr)_auto] gap-3 px-3">
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
          <CardAction className="pointer-events-none flex items-center gap-1.5">
            <StatusBadge status={agent.status} />
            <Button
              type="button"
              variant="outline"
              size="xs"
              className="pointer-events-auto"
              aria-label={t(
                "pages.agent.organization.details_shortcut_label",
                "Open Details for {{agent}}",
                { agent: displayName },
              )}
              onClick={() => openCardShortcut("details")}
            >
              <IconInfoCircle />
              {t("pages.agent.organization.details", "Details")}
            </Button>
          </CardAction>
        </CardHeader>
        <CardContent className="pointer-events-none relative z-10 space-y-2 px-3">
          <div
            className="text-muted-foreground flex min-w-0 items-center gap-1.5 text-xs"
            title={activity}
          >
            <IconClock className="size-3.5 shrink-0" />
            <span className="truncate">{activity}</span>
          </div>
          {currentFailureReason ? (
            <div
              className="text-destructive flex min-w-0 items-start gap-1.5 text-xs leading-4"
              title={currentFailureReason}
            >
              <IconAlertTriangle className="mt-0.5 size-3.5 shrink-0" />
              <span className="line-clamp-2 break-words">
                {currentFailureReason}
              </span>
            </div>
          ) : null}
          {counts.length > 0 ? (
            <div className="pointer-events-auto flex min-w-0 flex-wrap gap-1.5">
              {counts.map((item) => (
                <CountPill
                  key={item.key}
                  icon={item.icon}
                  label={item.label}
                  value={item.value}
                  ariaLabel={t(
                    "pages.agent.organization.activity_shortcut_label",
                    "Open {{label}} for {{agent}}",
                    { label: item.label, agent: displayName },
                  )}
                  onClick={() => openCardShortcut(item.key)}
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
        initialTab={detailInitialTab}
        onOpenChange={setDetailOpen}
      />
    </>
  )
}

function useDesktopWorkbenchLayout() {
  const [desktopWorkbench, setDesktopWorkbench] = useState(() =>
    typeof window === "undefined"
      ? false
      : window.matchMedia("(min-width: 1024px)").matches,
  )

  useEffect(() => {
    const mediaQuery = window.matchMedia("(min-width: 1024px)")
    const updateLayout = () => setDesktopWorkbench(mediaQuery.matches)

    updateLayout()
    mediaQuery.addEventListener("change", updateLayout)
    return () => mediaQuery.removeEventListener("change", updateLayout)
  }, [])

  return desktopWorkbench
}
