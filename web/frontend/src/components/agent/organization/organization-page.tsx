import {
  IconAlertTriangle,
  IconCalendarStats,
  IconCircleCheck,
  IconClock,
  IconInbox,
  IconLoader2,
  IconNetwork,
  IconSend,
} from "@tabler/icons-react"
import { useQuery } from "@tanstack/react-query"
import type { TFunction } from "i18next"
import { type ComponentType, type ReactNode, useMemo } from "react"
import { useTranslation } from "react-i18next"

import {
  type AgentOrganizationActivityRecord,
  type AgentOrganizationAgent,
  type AgentOrganizationNode,
  type AgentOrganizationSnapshot,
  getAgentOrganization,
} from "@/api/agents"
import { PageHeader } from "@/components/page-header"
import { Badge } from "@/components/ui/badge"
import {
  Card,
  CardAction,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { cn } from "@/lib/utils"

interface OrderedNode extends AgentOrganizationNode {
  children?: OrderedNode[]
}

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
        <CardAction>
          <StatusBadge status={agent.status} />
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
  )
}

function StatusBadge({ status }: { status: string }) {
  const { t } = useTranslation()
  const variant = status === "failed" ? "destructive" : "outline"

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
