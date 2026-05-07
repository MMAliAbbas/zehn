import {
  IconAlertTriangle,
  IconCalendarStats,
  IconCircleCheck,
  IconClock,
  IconFileDescription,
  IconInbox,
  IconInfoCircle,
  IconLoader2,
  IconNetwork,
  IconSend,
} from "@tabler/icons-react"
import { type UseQueryResult, useQuery } from "@tanstack/react-query"
import type { TFunction } from "i18next"
import { type ComponentType, type ReactNode, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"

import {
  type AgentDelegationActivityRecord,
  type AgentMeetingActivityRecord,
  type AgentOrganizationActivityRecord,
  type AgentOrganizationAgent,
  type AgentOrganizationNode,
  type AgentOrganizationRecentEvent,
  type AgentOrganizationSnapshot,
  getAgentInbox,
  getAgentMeetings,
  getAgentOrganization,
  getAgentOutbox,
} from "@/api/agents"
import { PageHeader } from "@/components/page-header"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardAction,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"
import { cn } from "@/lib/utils"

interface OrderedNode extends AgentOrganizationNode {
  children?: OrderedNode[]
}

type AgentDetailTab = "overview" | "inbox" | "outbox" | "meetings" | "recent"

const AGENT_DETAIL_LIMIT = 25

export function OrganizationPage() {
  const { t } = useTranslation()
  const organizationQuery = useQuery({
    queryKey: ["agents", "organization"],
    queryFn: getAgentOrganization,
  })

  const roots = useMemo(
    () => buildOrderedRoots(organizationQuery.data),
    [organizationQuery.data],
  )

  return (
    <div className="bg-background flex h-full flex-col">
      <PageHeader title={t("navigation.organization", "Organization")} />

      <div className="flex-1 overflow-auto px-6 py-6 pb-20">
        <div className="mx-auto w-full max-w-7xl space-y-4">
          {organizationQuery.isLoading ? (
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
          ) : organizationQuery.error ? (
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
                  <OrganizationBranch key={node.id} node={node} depth={0} />
                ))}
              </div>
            </section>
          )}
        </div>
      </div>
    </div>
  )
}

function SnapshotSummary({
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

function OrganizationBranch({
  node,
  depth,
}: {
  node: OrderedNode
  depth: number
}) {
  const hasChildren = (node.children?.length ?? 0) > 0

  return (
    <div className="min-w-0">
      <div
        className={cn(
          "grid min-w-0 grid-cols-[minmax(0,1fr)] gap-2",
          depth > 0 && "border-border/60 border-l pl-3 sm:pl-4",
        )}
      >
        <AgentCard agent={node} />
        {hasChildren && (
          <div className="space-y-2">
            {node.children?.map((child) => (
              <OrganizationBranch
                key={`${node.id}:${child.id}`}
                node={child}
                depth={depth + 1}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

function AgentCard({ agent }: { agent: AgentOrganizationAgent }) {
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
            <div className="text-muted-foreground flex items-center gap-1.5 text-xs">
              <IconCircleCheck className="size-3.5" />
              {t("pages.agent.organization.no_activity", "No active records")}
            </div>
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

function AgentDetailSheet({
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
  })
  const outboxQuery = useQuery({
    queryKey: ["agents", agent.id, "outbox", AGENT_DETAIL_LIMIT],
    queryFn: () => getAgentOutbox(agent.id, AGENT_DETAIL_LIMIT),
    enabled: open && activeTab === "outbox",
  })
  const meetingsQuery = useQuery({
    queryKey: ["agents", agent.id, "meetings", AGENT_DETAIL_LIMIT],
    queryFn: () => getAgentMeetings(agent.id, AGENT_DETAIL_LIMIT),
    enabled: open && activeTab === "meetings",
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

function AgentOverviewPanel({ agent }: { agent: AgentOrganizationAgent }) {
  const { t } = useTranslation()
  const activity = summarizeActivity(agent.activity.current, t)
  const metrics = [
    {
      label: t("pages.agent.organization.inbox", "Inbox"),
      value: agent.activity.inbox_count,
    },
    {
      label: t("pages.agent.organization.outbox", "Outbox"),
      value: agent.activity.outbox_count,
    },
    {
      label: t("pages.agent.organization.meetings", "Meetings"),
      value: agent.activity.meeting_count,
    },
    {
      label: t("pages.agent.organization.errors", "Errors"),
      value: agent.activity.failure_count,
    },
  ]

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center gap-2">
        <StatusBadge status={agent.status} />
        <span className="text-muted-foreground text-sm">{activity}</span>
      </div>
      <div className="grid gap-2 sm:grid-cols-4">
        {metrics.map((metric) => (
          <div
            key={metric.label}
            className="border-border/70 rounded-lg border px-3 py-2"
          >
            <div className="text-muted-foreground text-xs">{metric.label}</div>
            <div className="text-lg font-medium tabular-nums">
              {metric.value}
            </div>
          </div>
        ))}
      </div>
      <div className="space-y-2">
        <RecordFact
          label={t("pages.agent.organization.detail.workspace", "Workspace")}
          value={agent.workspace || t("common.notAvailable", "Unavailable")}
        />
        <RecordFact
          label={t("pages.agent.organization.detail.group", "Group")}
          value={agent.group || t("common.notAvailable", "Unavailable")}
        />
        <RecordFact
          label={t(
            "pages.agent.organization.detail.last_updated",
            "Last updated",
          )}
          value={formatTimestamp(agent.activity.last_updated_at, t)}
        />
      </div>
    </div>
  )
}

function DelegationRecordsPanel({
  agentID,
  label,
  query,
}: {
  agentID: string
  label: string
  query: UseQueryResult<Awaited<ReturnType<typeof getAgentInbox>>, Error>
}) {
  const { t } = useTranslation()
  if (query.isLoading) {
    return (
      <TabState
        loading
        title={t("pages.agent.organization.detail.loading", "Loading records")}
      />
    )
  }
  if (query.error) {
    return (
      <TabState
        destructive
        title={t(
          "pages.agent.organization.detail.load_error",
          "Failed to load records",
        )}
        detail={errorMessage(query.error)}
      />
    )
  }
  const records = query.data?.records ?? []
  if (records.length === 0) {
    return (
      <TabState
        title={t("pages.agent.organization.detail.empty", "No records")}
        detail={t(
          "pages.agent.organization.detail.empty_detail",
          "This section has no visible activity records.",
        )}
      />
    )
  }
  return (
    <div className="space-y-2" aria-label={label}>
      {records.map((record) => (
        <DelegationRecordItem
          key={record.delegation_id}
          agentID={agentID}
          record={record}
        />
      ))}
    </div>
  )
}

function DelegationRecordItem({
  agentID,
  record,
}: {
  agentID: string
  record: AgentDelegationActivityRecord
}) {
  const { t } = useTranslation()
  const peerAgent =
    record.role === "target"
      ? record.requester_id || record.parent_agent_id
      : record.target_agent_id

  return (
    <ActivityRecordFrame status={record.status}>
      <div className="flex min-w-0 flex-wrap items-start justify-between gap-2">
        <div className="min-w-0">
          <div className="truncate text-sm font-medium">
            {t("pages.agent.organization.detail.delegation_title", {
              defaultValue: "Delegation {{id}}",
              id: shortRecordID(record.delegation_id),
            })}
          </div>
          <div className="text-muted-foreground mt-0.5 truncate text-xs">
            {t("pages.agent.organization.detail.peer_agent", {
              defaultValue: "Peer: {{agent}}",
              agent: peerAgent || agentID,
            })}
          </div>
        </div>
        <StatusBadge status={record.status} />
      </div>
      <div className="mt-3 grid gap-2 sm:grid-cols-2">
        <RecordFact
          label={t("pages.agent.organization.role_label", "Role")}
          value={t(`pages.agent.organization.role.${record.role}`, record.role)}
        />
        <RecordFact
          label={t("pages.agent.organization.detail.mode", "Mode")}
          value={record.mode || t("common.notAvailable", "Unavailable")}
        />
        <RecordFact
          label={t("pages.agent.organization.detail.created", "Created")}
          value={formatTimestamp(record.created_at, t)}
        />
        <RecordFact
          label={t("pages.agent.organization.detail.updated", "Updated")}
          value={formatTimestamp(record.updated_at, t)}
        />
      </div>
      <ArtifactSummary refs={record.artifact_refs} />
    </ActivityRecordFrame>
  )
}

function MeetingRecordsPanel({
  query,
}: {
  query: UseQueryResult<Awaited<ReturnType<typeof getAgentMeetings>>, Error>
}) {
  const { t } = useTranslation()
  if (query.isLoading) {
    return (
      <TabState
        loading
        title={t("pages.agent.organization.detail.loading", "Loading records")}
      />
    )
  }
  if (query.error) {
    return (
      <TabState
        destructive
        title={t(
          "pages.agent.organization.detail.load_error",
          "Failed to load records",
        )}
        detail={errorMessage(query.error)}
      />
    )
  }
  const records = query.data?.records ?? []
  if (records.length === 0) {
    return (
      <TabState
        title={t("pages.agent.organization.detail.empty", "No records")}
        detail={t(
          "pages.agent.organization.detail.empty_detail",
          "This section has no visible activity records.",
        )}
      />
    )
  }
  return (
    <div className="space-y-2">
      {records.map((record) => (
        <MeetingRecordItem key={record.meeting_id} record={record} />
      ))}
    </div>
  )
}

function MeetingRecordItem({ record }: { record: AgentMeetingActivityRecord }) {
  const { t } = useTranslation()
  return (
    <ActivityRecordFrame status={record.status}>
      <div className="flex min-w-0 flex-wrap items-start justify-between gap-2">
        <div className="min-w-0">
          <div className="truncate text-sm font-medium">
            {record.title ||
              t("pages.agent.organization.detail.meeting_title", {
                defaultValue: "Meeting {{id}}",
                id: shortRecordID(record.meeting_id),
              })}
          </div>
          <div className="text-muted-foreground mt-0.5 truncate text-xs">
            {t("pages.agent.organization.detail.chair_agent", {
              defaultValue: "Chair: {{agent}}",
              agent: record.chair_agent_id,
            })}
          </div>
        </div>
        <StatusBadge status={record.status} />
      </div>
      <div className="mt-3 grid gap-2 sm:grid-cols-2">
        <RecordFact
          label={t("pages.agent.organization.role_label", "Role")}
          value={t(`pages.agent.organization.role.${record.role}`, record.role)}
        />
        <RecordFact
          label={t("pages.agent.organization.detail.sponsor", "Sponsor")}
          value={record.sponsor_agent_id}
        />
        <RecordFact
          label={t("pages.agent.organization.detail.created", "Created")}
          value={formatTimestamp(record.created_at, t)}
        />
        <RecordFact
          label={t("pages.agent.organization.detail.updated", "Updated")}
          value={formatTimestamp(record.updated_at, t)}
        />
      </div>
      <div className="mt-3">
        <RecordFact
          label={t(
            "pages.agent.organization.detail.participants",
            "Participants",
          )}
          value={
            (record.participants ?? []).join(", ") ||
            t("common.notAvailable", "Unavailable")
          }
        />
      </div>
      <ArtifactSummary refs={record.artifact_refs} />
    </ActivityRecordFrame>
  )
}

function RecentEventsPanel({ agent }: { agent: AgentOrganizationAgent }) {
  const { t } = useTranslation()
  const activityEvents = compactActivityEvents(
    agent.activity.current,
    agent.activity.last_failure,
  )
  const logEvents = agent.activity.recent_events ?? []
  if (activityEvents.length === 0 && logEvents.length === 0) {
    return (
      <TabState
        title={t("pages.agent.organization.detail.empty", "No records")}
        detail={t(
          "pages.agent.organization.detail.no_recent_detail",
          "No recent event summaries are available for this agent.",
        )}
      />
    )
  }
  return (
    <div className="space-y-2">
      {activityEvents.map((event) => (
        <ActivityRecordFrame
          key={`${event.type}:${event.record_id}`}
          status={event.status}
        >
          <div className="flex min-w-0 flex-wrap items-start justify-between gap-2">
            <div className="min-w-0">
              <div className="truncate text-sm font-medium">
                {summarizeActivity(event, t)}
              </div>
              <div className="text-muted-foreground mt-0.5 truncate font-mono text-xs">
                {event.record_id}
              </div>
            </div>
            <StatusBadge status={event.status} />
          </div>
          <div className="mt-3 grid gap-2 sm:grid-cols-2">
            <RecordFact
              label={t(
                "pages.agent.organization.detail.peer_agent_short",
                "Peer",
              )}
              value={event.agent_id || t("common.notAvailable", "Unavailable")}
            />
            <RecordFact
              label={t("pages.agent.organization.detail.updated", "Updated")}
              value={formatTimestamp(event.updated_at, t)}
            />
          </div>
        </ActivityRecordFrame>
      ))}
      {logEvents.map((event, index) => (
        <GatewayLogEventFrame
          key={`${event.source}:${event.timestamp ?? "untimed"}:${index}`}
          event={event}
        />
      ))}
    </div>
  )
}

function GatewayLogEventFrame({
  event,
}: {
  event: AgentOrganizationRecentEvent
}) {
  const { t } = useTranslation()
  const status = event.level || event.event || "info"
  return (
    <ActivityRecordFrame status={status}>
      <div className="flex min-w-0 flex-wrap items-start justify-between gap-2">
        <div className="min-w-0">
          <div className="truncate text-sm font-medium">{event.message}</div>
          <div className="text-muted-foreground mt-0.5 truncate font-mono text-xs">
            {event.event ||
              t("pages.agent.organization.detail.gateway_log", "gateway_log")}
          </div>
        </div>
        <StatusBadge status={status} />
      </div>
      <div className="mt-3 grid gap-2 sm:grid-cols-2">
        <RecordFact
          label={t("pages.agent.organization.detail.source", "Source")}
          value={event.source}
        />
        <RecordFact
          label={t("pages.agent.organization.detail.updated", "Updated")}
          value={formatTimestamp(event.timestamp, t)}
        />
      </div>
    </ActivityRecordFrame>
  )
}

function StatusBadge({ status }: { status: string }) {
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

function CountPill({
  icon: Icon,
  label,
  value,
}: {
  icon: ComponentType<{ className?: string }>
  label: string
  value: number
}) {
  return (
    <span
      className="border-border/70 bg-muted/30 inline-flex h-6 max-w-full items-center gap-1 rounded-md border px-1.5 text-xs"
      title={`${label}: ${value}`}
    >
      <Icon className="text-muted-foreground size-3.5 shrink-0" />
      <span className="text-muted-foreground truncate">{label}</span>
      <span className="font-medium">{value}</span>
    </span>
  )
}

function TabState({
  title,
  detail,
  loading = false,
  destructive = false,
}: {
  title: string
  detail?: string
  loading?: boolean
  destructive?: boolean
}) {
  return (
    <div
      role={destructive ? "alert" : "status"}
      className={cn(
        "border-border/70 flex items-start gap-3 rounded-lg border px-4 py-4 text-sm",
        destructive && "border-destructive/30 text-destructive",
      )}
    >
      <div className="mt-0.5 shrink-0">
        {loading ? (
          <IconLoader2 className="size-4 animate-spin" />
        ) : destructive ? (
          <IconAlertTriangle className="size-4" />
        ) : (
          <IconCircleCheck className="size-4" />
        )}
      </div>
      <div className="min-w-0">
        <div className="font-medium">{title}</div>
        {detail ? (
          <div
            className={cn(
              "text-muted-foreground mt-1 break-words",
              destructive && "text-destructive/80",
            )}
          >
            {detail}
          </div>
        ) : null}
      </div>
    </div>
  )
}

function ActivityRecordFrame({
  status,
  children,
}: {
  status: string
  children: ReactNode
}) {
  const isProblem = isProblemStatus(status)
  return (
    <article
      className={cn(
        "border-border/70 rounded-lg border px-3 py-3 text-sm",
        isProblem && "border-destructive/30 bg-destructive/3",
      )}
    >
      {children}
    </article>
  )
}

function RecordFact({ label, value }: { label: string; value: string }) {
  return (
    <div className="min-w-0">
      <div className="text-muted-foreground text-[11px] leading-4">{label}</div>
      <div className="truncate text-xs leading-5" title={value}>
        {value}
      </div>
    </div>
  )
}

function ArtifactSummary({ refs }: { refs?: string[] }) {
  const { t } = useTranslation()
  const count = refs?.length ?? 0
  return (
    <div className="text-muted-foreground mt-3 flex min-w-0 items-center gap-1.5 text-xs">
      <IconFileDescription className="size-3.5 shrink-0" />
      <span className="truncate">
        {count > 0
          ? t("pages.agent.organization.detail.artifact_count", {
              defaultValue: "{{count}} artifact reference",
              defaultValue_plural: "{{count}} artifact references",
              count,
            })
          : t(
              "pages.agent.organization.detail.no_artifacts",
              "No artifact references",
            )}
      </span>
    </div>
  )
}

function StatePanel({
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

function buildOrderedRoots(
  snapshot: AgentOrganizationSnapshot | undefined,
): OrderedNode[] {
  if (!snapshot) {
    return []
  }
  if ((snapshot.roots?.length ?? 0) > 0) {
    return preserveNodeOrder(snapshot.roots ?? [])
  }
  return Object.values(snapshot.agents ?? {})
    .sort(compareAgents)
    .map((agent) => ({ ...agent, children: [] }))
}

function preserveNodeOrder(nodes: AgentOrganizationNode[]): OrderedNode[] {
  return nodes.map((node) => ({
    ...node,
    children: node.children ? preserveNodeOrder(node.children) : [],
  }))
}

function compareAgents(
  a: Pick<AgentOrganizationAgent, "id" | "label" | "name">,
  b: Pick<AgentOrganizationAgent, "id" | "label" | "name">,
) {
  return (
    displayAgentName(a).localeCompare(displayAgentName(b), undefined, {
      sensitivity: "base",
      numeric: true,
    }) || a.id.localeCompare(b.id, undefined, { sensitivity: "base" })
  )
}

function displayAgentName(
  agent: Pick<AgentOrganizationAgent, "id" | "label" | "name">,
) {
  return agent.label?.trim() || agent.name?.trim() || agent.id
}

function summarizeActivity(
  current: AgentOrganizationActivityRecord | undefined,
  t: TFunction,
) {
  if (!current) {
    return t("pages.agent.organization.idle_summary", "Idle")
  }
  const type = t(
    `pages.agent.organization.activity_type.${current.type}`,
    current.type,
  )
  const role = current.role
    ? t(`pages.agent.organization.role.${current.role}`, current.role)
    : ""
  const status = t(
    `pages.agent.organization.status.${current.status}`,
    current.status,
  )
  return [type, role, status].filter(Boolean).join(" / ")
}

function compactActivityEvents(
  current: AgentOrganizationActivityRecord | undefined,
  lastFailure: AgentOrganizationActivityRecord | undefined,
) {
  const records = [current, lastFailure].filter(
    Boolean,
  ) as AgentOrganizationActivityRecord[]
  const seen = new Set<string>()
  return records.filter((record) => {
    const key = `${record.type}:${record.record_id}`
    if (seen.has(key)) {
      return false
    }
    seen.add(key)
    return true
  })
}

function formatTimestamp(value: string | undefined, t: TFunction) {
  if (!value) {
    return t("common.notAvailable", "Unavailable")
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date)
}

function shortRecordID(id: string) {
  if (id.length <= 12) {
    return id
  }
  return id.slice(0, 12)
}

function isProblemStatus(status: string) {
  const normalized = status.toLowerCase()
  return (
    normalized === "failed" ||
    normalized === "blocked" ||
    normalized === "error" ||
    normalized === "fatal"
  )
}

function errorMessage(error: unknown) {
  return error instanceof Error ? error.message : String(error)
}
