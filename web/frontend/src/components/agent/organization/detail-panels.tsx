import type { UseQueryResult } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"

import type {
  AgentDelegationActivityRecord,
  AgentMeetingActivityRecord,
  AgentOrganizationAgent,
  AgentOrganizationRecentEvent,
} from "@/api/agents"
import { getAgentInbox, getAgentMeetings } from "@/api/agents"
import { LogsPanel } from "@/components/logs/logs-panel"
import { useGatewayLogs } from "@/hooks/use-gateway-logs"
import { useLogWrapColumns } from "@/hooks/use-log-wrap-columns"

import {
  compactActivityEvents,
  errorMessage,
  formatTimestamp,
  shortRecordID,
  summarizeActivity,
} from "./formatting"
import {
  ActivityRecordFrame,
  ArtifactSummary,
  RecordFact,
  TabState,
} from "./record-components"
import { StatusBadge } from "./status-components"

export function AgentOverviewPanel({
  agent,
}: {
  agent: AgentOrganizationAgent
}) {
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

export function LiveLogsPanel() {
  const { t } = useTranslation()
  const { contentRef, measureRef, wrapColumns } = useLogWrapColumns()
  const { error, gatewayStatus, logs, stale } = useGatewayLogs()

  const statusText = error
    ? t("pages.agent.organization.detail.live_logs_error", {
        defaultValue: "Log polling error: {{message}}",
        message: error,
      })
    : stale
      ? t(
          "pages.agent.organization.detail.live_logs_stale",
          "Log polling is stale",
        )
      : gatewayStatus === "stopped"
        ? t(
            "pages.agent.organization.detail.live_logs_stopped",
            "Gateway is stopped",
          )
        : gatewayStatus === "error"
          ? t(
              "pages.agent.organization.detail.live_logs_gateway_error",
              "Gateway is in an error state",
            )
          : t("pages.agent.organization.detail.live_logs_status", {
              defaultValue: "Gateway {{status}}",
              status: t(
                `pages.agent.organization.status.${gatewayStatus}`,
                gatewayStatus,
              ),
            })

  return (
    <div className="flex min-h-[26rem] flex-col gap-3">
      <div
        className={
          error || stale || gatewayStatus === "error"
            ? "border-destructive/30 bg-destructive/10 text-destructive rounded-md border px-3 py-2 text-xs"
            : "border-border/70 text-muted-foreground rounded-md border px-3 py-2 text-xs"
        }
      >
        {statusText}
      </div>
      <div className="min-h-0 flex-1">
        <LogsPanel
          logs={logs}
          wrapColumns={wrapColumns}
          contentRef={contentRef}
          measureRef={measureRef}
        />
      </div>
    </div>
  )
}

export function DelegationRecordsPanel({
  agentID,
  label,
  query,
}: {
  agentID: string
  label: string
  query: UseQueryResult<Awaited<ReturnType<typeof getAgentInbox>>, Error>
}) {
  const { t } = useTranslation()
  if (query.isLoading && !query.data) {
    return (
      <TabState
        loading
        title={t("pages.agent.organization.detail.loading", "Loading records")}
      />
    )
  }
  if (query.error && !query.data) {
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

export function MeetingRecordsPanel({
  query,
}: {
  query: UseQueryResult<Awaited<ReturnType<typeof getAgentMeetings>>, Error>
}) {
  const { t } = useTranslation()
  if (query.isLoading && !query.data) {
    return (
      <TabState
        loading
        title={t("pages.agent.organization.detail.loading", "Loading records")}
      />
    )
  }
  if (query.error && !query.data) {
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

export function RecentEventsPanel({
  agent,
}: {
  agent: AgentOrganizationAgent
}) {
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
