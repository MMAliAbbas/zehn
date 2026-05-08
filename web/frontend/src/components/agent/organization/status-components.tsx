import {
  IconActivity,
  IconAlertTriangle,
  IconCalendarEvent,
  IconCircleCheck,
  IconClipboardList,
  IconClock,
  IconHierarchy2,
  IconRefresh,
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

import {
  type OrganizationHeaderRefreshState,
  resolveOrganizationHeaderRefreshState,
  summarizeOrganizationHeaderActivity,
} from "./command-header-state"
import {
  displayAgentName,
  formatTimestamp,
  isProblemStatus,
} from "./formatting"
import type { AgentWorkbenchSection } from "./types"

export function OrganizationCommandHeader({
  snapshot,
  isFetching,
  isError,
  dataUpdatedAt,
}: {
  snapshot: AgentOrganizationSnapshot | undefined
  isFetching: boolean
  isError: boolean
  dataUpdatedAt: number
}) {
  const { t } = useTranslation()
  const summary = summarizeOrganizationHeaderActivity(snapshot)
  const refreshState = resolveOrganizationHeaderRefreshState({
    hasData: Boolean(snapshot),
    isError,
    isFetching,
  })
  const generatedAt = snapshot?.metadata?.generated_at
  const refreshedAt =
    dataUpdatedAt > 0 ? new Date(dataUpdatedAt).toISOString() : undefined

  return (
    <section className="border-border/70 bg-card rounded-lg border px-3 py-2.5 text-sm shadow-xs">
      <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div className="grid min-w-0 flex-1 grid-cols-2 gap-2 sm:grid-cols-3 xl:grid-cols-6">
          <CommandHeaderMetric
            label={t(
              "pages.agent.organization.command_header.active_work",
              "Active Work",
            )}
            value={summary.activeWork}
          />
          <CommandHeaderMetric
            label={t("pages.agent.organization.delegations", "Delegations")}
            value={summary.delegations}
          />
          <CommandHeaderMetric
            label={t("pages.agent.organization.meetings", "Meetings")}
            value={summary.meetings}
          />
          <CommandHeaderMetric
            label={t("pages.agent.organization.detail.failures", "Failures")}
            value={summary.failures}
            problem={summary.failures > 0}
          />
          <CommandHeaderMetric
            icon={<IconHierarchy2 className="size-3.5" />}
            label={t("pages.agent.organization.mode", "Mode")}
            value={
              summary.mode === "hierarchy"
                ? t("pages.agent.organization.hierarchy", "Hierarchy")
                : t("pages.agent.organization.flat", "Flat")
            }
          />
          <CommandHeaderRefreshBadge state={refreshState} />
        </div>

        <div className="text-muted-foreground grid min-w-0 gap-1 text-xs sm:grid-cols-2 lg:w-auto lg:max-w-md lg:grid-cols-1 xl:grid-cols-2">
          <CommandHeaderTime
            label={t(
              "pages.agent.organization.command_header.generated",
              "Generated",
            )}
            value={formatTimestamp(generatedAt, t)}
          />
          <CommandHeaderTime
            label={t(
              "pages.agent.organization.command_header.refreshed",
              "Refreshed",
            )}
            value={formatTimestamp(refreshedAt, t)}
          />
        </div>
      </div>
    </section>
  )
}

function CommandHeaderMetric({
  icon,
  label,
  value,
  problem = false,
}: {
  icon?: ReactNode
  label: string
  value: number | string
  problem?: boolean
}) {
  return (
    <div className="min-w-0 px-1.5 py-0.5">
      <div className="text-muted-foreground flex min-w-0 items-center gap-1 text-xs">
        {icon ? <span className="shrink-0">{icon}</span> : null}
        <span className="truncate">{label}</span>
      </div>
      <div
        className={cn(
          "truncate text-base leading-tight font-semibold",
          problem && "text-destructive",
        )}
      >
        {value}
      </div>
    </div>
  )
}

function CommandHeaderRefreshBadge({
  state,
}: {
  state: OrganizationHeaderRefreshState
}) {
  const { t } = useTranslation()
  const labels: Record<OrganizationHeaderRefreshState, string> = {
    loading: t("pages.agent.organization.command_header.loading", "Loading"),
    refreshing: t(
      "pages.agent.organization.command_header.refreshing",
      "Refreshing",
    ),
    stale: t("pages.agent.organization.command_header.stale", "Stale"),
    live: t("pages.agent.organization.command_header.live", "Live"),
  }

  return (
    <div className="min-w-0 px-1.5 py-0.5">
      <div className="text-muted-foreground flex min-w-0 items-center gap-1 text-xs">
        <IconRefresh className="size-3.5 shrink-0" />
        <span className="truncate">
          {t("pages.agent.organization.command_header.query", "Query")}
        </span>
      </div>
      <div className="flex min-w-0 items-center gap-1.5 text-base leading-tight font-semibold">
        <span
          className={cn(
            "size-2 shrink-0 rounded-full",
            state === "live" && "bg-emerald-500",
            state === "refreshing" && "bg-blue-500",
            state === "loading" && "bg-muted-foreground",
            state === "stale" && "bg-destructive",
          )}
          aria-hidden="true"
        />
        <span className="truncate">{labels[state]}</span>
      </div>
    </div>
  )
}

function CommandHeaderTime({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex min-w-0 items-center gap-1.5">
      <IconClock className="size-3.5 shrink-0" />
      <span className="shrink-0">{label}</span>
      <span className="text-foreground/80 truncate font-medium">{value}</span>
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
        {t(`pages.agent.organization.activity_type.${entry.type}`, entry.type)}
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
