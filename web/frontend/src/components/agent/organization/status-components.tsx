import {
  IconActivity,
  IconAlertTriangle,
  IconCalendarEvent,
  IconCircleCheck,
  IconClipboardList,
  IconTerminal2,
} from "@tabler/icons-react"
import type { ComponentType, MouseEventHandler, ReactNode } from "react"
import { useTranslation } from "react-i18next"

import type {
  AgentOrganizationActivityFeed,
  AgentOrganizationSnapshot,
} from "@/api/agents"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

import { displayAgentName, formatTimestamp, isProblemStatus } from "./formatting"
import type { AgentWorkbenchSection } from "./types"

export function SnapshotSummary({
  snapshot,
}: {
  snapshot: AgentOrganizationSnapshot | undefined
}) {
  const { t } = useTranslation()
  const activity = snapshot?.activity
  const hasHierarchy = snapshot?.metadata?.has_hierarchy === true

  return (
    <div className="border-border/70 bg-card grid gap-3 rounded-lg border px-4 py-3 text-sm shadow-xs sm:grid-cols-4">
      <SummaryMetric
        label={t("pages.agent.organization.active", "Active")}
        value={activity?.active_count ?? 0}
      />
      <SummaryMetric
        label={t("pages.agent.organization.delegations", "Delegations")}
        value={activity?.delegation_count ?? 0}
      />
      <SummaryMetric
        label={t("pages.agent.organization.meetings", "Meetings")}
        value={activity?.meeting_count ?? 0}
      />
      <SummaryMetric
        label={t("pages.agent.organization.mode", "Mode")}
        value={
          hasHierarchy
            ? t("pages.agent.organization.hierarchy", "Hierarchy")
            : t("pages.agent.organization.flat", "Flat")
        }
      />
    </div>
  )
}

function SummaryMetric({
  label,
  value,
}: {
  label: string
  value: number | string
}) {
  return (
    <div className="min-w-0">
      <div className="text-muted-foreground text-xs">{label}</div>
      <div className="truncate text-base font-medium">{value}</div>
    </div>
  )
}

export function StatusBadge({ status }: { status: string }) {
  const { t } = useTranslation()
  const variant = isProblemStatus(status) ? "destructive" : "outline"

  return (
    <Badge
      variant={variant}
      className={cn(
        "max-w-28 truncate rounded-md capitalize",
        status === "meeting" && "border-blue-500/30 text-blue-700",
        status === "working" && "border-emerald-500/30 text-emerald-700",
        status === "delegating" && "border-amber-500/30 text-amber-700",
      )}
      title={status}
    >
      {t(`pages.agent.organization.status.${status}`, status)}
    </Badge>
  )
}

export function CountPill({
  icon: Icon,
  label,
  value,
  ariaLabel,
  onClick,
}: {
  icon: ComponentType<{ className?: string }>
  label: string
  value: number
  ariaLabel?: string
  onClick?: MouseEventHandler<HTMLButtonElement>
}) {
  const content = (
    <>
      <Icon className="text-muted-foreground size-3.5 shrink-0" />
      <span className="text-muted-foreground truncate">{label}</span>
      <span className="font-medium">{value}</span>
    </>
  )

  if (onClick) {
    return (
      <Button
        type="button"
        variant="ghost"
        size="xs"
        className="border-border/70 bg-muted/30 hover:bg-muted/70 focus-visible:ring-ring/50 h-6 max-w-full gap-1 border px-1.5"
        title={`${label}: ${value}`}
        aria-label={ariaLabel ?? `${label}: ${value}`}
        onClick={onClick}
        onKeyDown={(event) => event.stopPropagation()}
      >
        {content}
      </Button>
    )
  }

  return (
    <span
      className="border-border/70 bg-muted/30 inline-flex h-6 max-w-full items-center gap-1 rounded-md border px-1.5 text-xs"
      title={`${label}: ${value}`}
    >
      {content}
    </span>
  )
}

export function StatePanel({
  icon,
  title,
  detail,
  destructive = false,
}: {
  icon: ReactNode
  title: string
  detail: string
  destructive?: boolean
}) {
  return (
    <div
      className={cn(
        "border-border/70 bg-card flex items-start gap-3 rounded-lg border px-4 py-4 text-sm shadow-xs",
        destructive && "border-destructive/30 text-destructive",
      )}
    >
      <div className="mt-0.5 shrink-0">{icon}</div>
      <div className="min-w-0">
        <div className="font-medium">{title}</div>
        <div
          className={cn(
            "text-muted-foreground mt-1 break-words",
            destructive && "text-destructive/80",
          )}
        >
          {detail}
        </div>
      </div>
    </div>
  )
}

export function OrganizationActivityFeed({
  snapshot,
  onSelectAgent,
}: {
  snapshot: AgentOrganizationSnapshot | undefined
  onSelectAgent: (agentID: string, section?: AgentWorkbenchSection) => void
}) {
  const { t } = useTranslation()
  const entries = snapshot?.activity?.recent ?? []
  const agents = snapshot?.agents ?? {}

  return (
    <section className="border-border/70 bg-card rounded-lg border px-4 py-3 shadow-xs">
      <div className="mb-2 flex items-center justify-between gap-3">
        <div className="flex min-w-0 items-center gap-2">
          <IconActivity className="text-muted-foreground size-4 shrink-0" />
          <h2 className="truncate text-sm font-medium">
            {t("pages.agent.organization.feed.title", "Recent Activity")}
          </h2>
        </div>
        <Badge variant="outline" className="rounded-md">
          {entries.length}
        </Badge>
      </div>
      {entries.length === 0 ? (
        <EmptyActivityLine
          label={t(
            "pages.agent.organization.feed.empty",
            "No recent activity records",
          )}
        />
      ) : (
        <div className="grid gap-2 md:grid-cols-2 xl:grid-cols-4">
          {entries.slice(0, 8).map((entry, index) => (
            <ActivityFeedEntry
              key={`${entry.type}:${entry.record_id ?? entry.agent_id ?? index}:${entry.timestamp ?? index}`}
              entry={entry}
              agentLabel={
                entry.agent_id && agents[entry.agent_id]
                  ? displayAgentName(agents[entry.agent_id])
                  : entry.agent_id
              }
              onSelectAgent={
                entry.agent_id && agents[entry.agent_id]
                  ? () =>
                      onSelectAgent(
                        entry.agent_id as string,
                        activityFeedSection(entry),
                      )
                  : undefined
              }
            />
          ))}
        </div>
      )}
    </section>
  )
}

function ActivityFeedEntry({
  entry,
  agentLabel,
  onSelectAgent,
}: {
  entry: AgentOrganizationActivityFeed
  agentLabel?: string
  onSelectAgent?: () => void
}) {
  const { t } = useTranslation()
  const content = (
    <>
      <div className="flex min-w-0 items-center gap-2">
        {activityFeedIcon(entry.type)}
        <span className="truncate text-xs font-medium">
          {agentLabel ??
            t("pages.agent.organization.feed.org_scope", "Organization")}
        </span>
        {entry.status ? (
          <Badge
            variant={isProblemStatus(entry.status) ? "destructive" : "outline"}
            className="max-w-24 shrink-0 truncate rounded-md capitalize"
          >
            {t(`pages.agent.organization.status.${entry.status}`, entry.status)}
          </Badge>
        ) : null}
      </div>
      <div className="mt-1 truncate text-xs">
        {t(
          `pages.agent.organization.activity_type.${entry.type}`,
          entry.type,
        )}
        {entry.summary ? ` / ${entry.summary}` : ""}
      </div>
      <div className="text-muted-foreground mt-1 truncate text-xs">
        {formatTimestamp(entry.timestamp, t)}
      </div>
    </>
  )

  if (onSelectAgent) {
    return (
      <button
        type="button"
        className="border-border/70 bg-background hover:bg-muted/60 focus-visible:ring-ring/50 min-w-0 rounded-md border px-3 py-2 text-left text-sm transition focus-visible:ring-2 focus-visible:outline-hidden"
        onClick={onSelectAgent}
      >
        {content}
      </button>
    )
  }

  return (
    <div className="border-border/70 bg-background min-w-0 rounded-md border px-3 py-2 text-sm">
      {content}
    </div>
  )
}

function activityFeedIcon(type: string) {
  const className = "text-muted-foreground size-3.5 shrink-0"
  switch (type) {
    case "delegation":
      return <IconClipboardList className={className} />
    case "meeting":
      return <IconCalendarEvent className={className} />
    case "failure":
      return <IconAlertTriangle className={className} />
    case "event":
      return <IconTerminal2 className={className} />
    default:
      return <IconActivity className={className} />
  }
}

function activityFeedSection(
  entry: AgentOrganizationActivityFeed,
): AgentWorkbenchSection {
  switch (entry.type) {
    case "delegation":
      return "inbox"
    case "meeting":
      return "meetings"
    case "failure":
      return "failures"
    case "event":
      return "recent"
    default:
      return "overview"
  }
}

export function EmptyActivityLine({ label }: { label: string }) {
  return (
    <div className="text-muted-foreground flex items-center gap-1.5 text-xs">
      <IconCircleCheck className="size-3.5" />
      {label}
    </div>
  )
}
