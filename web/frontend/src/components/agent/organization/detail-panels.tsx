import { IconListDetails, IconX } from "@tabler/icons-react"
import { useQuery, type UseQueryResult } from "@tanstack/react-query"
import type { ReactNode } from "react"
import { useMemo, useState } from "react"
import { useTranslation } from "react-i18next"

import type {
  AgentOrganizationActivityDetail,
  AgentDelegationActivityRecord,
  AgentMeetingActivityRecord,
  AgentOrganizationActivityRecord,
  AgentOrganizationAgent,
  AgentOrganizationRecentEvent,
} from "@/api/agents"
import {
  ApiRequestError,
  getAgentActivityDetail,
  getAgentFailures,
  getAgentInbox,
  getAgentMeetings,
} from "@/api/agents"
import { LogsPanel } from "@/components/logs/logs-panel"
import { Button } from "@/components/ui/button"
import { useGatewayLogs } from "@/hooks/use-gateway-logs"
import { useLogWrapColumns } from "@/hooks/use-log-wrap-columns"
import type { AgentLogScopeMode } from "@/lib/agent-log-filter"
import {
  filterAgentLogLines,
  findAgentLogReferenceFields,
} from "@/lib/agent-log-filter"

import {
  buildFailureDrilldownRecords,
  compactActivityEvents,
  errorMessage,
  formatDiagnosticFreshness,
  formatDiagnosticReason,
  formatDiagnosticReasonSource,
  formatDiagnosticSeverity,
  formatTimestamp,
  shortRecordID,
  summarizeActivity,
} from "./formatting"
import { resolveSelectableActivityRecord } from "./organization-state"
import {
  ActivityRecordFrame,
  ArtifactSummary,
  RecordFact,
  TabState,
} from "./record-components"
import { StatusBadge } from "./status-components"
import type { AgentSelectedActivityRecord } from "./types"

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

export function LiveLogsPanel({ agent }: { agent: AgentOrganizationAgent }) {
  const { t } = useTranslation()
  const [scopeMode, setScopeMode] = useState<AgentLogScopeMode>("all")
  const { contentRef, measureRef, wrapColumns } = useLogWrapColumns()
  const { error, gatewayStatus, logs, stale } = useGatewayLogs()
  const visibleLogs = useMemo(
    () => filterAgentLogLines(logs, agent.id, scopeMode),
    [agent.id, logs, scopeMode],
  )
  const referencedLogCount = useMemo(
    () =>
      logs.filter(
        (line) => findAgentLogReferenceFields(line, agent.id).length > 0,
      ).length,
    [agent.id, logs],
  )
  const selectedAgentEmptyMessage = t(
    "pages.agent.organization.detail.live_logs_selected_empty",
    {
      defaultValue: "No live logs reference {{agent}} yet.",
      agent: agent.label || agent.name || agent.id,
    },
  )

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
      <div className="flex flex-col gap-2 lg:flex-row lg:items-center lg:justify-between">
        <div
          className={
            error || stale || gatewayStatus === "error"
              ? "border-destructive/30 bg-destructive/10 text-destructive rounded-md border px-3 py-2 text-xs"
              : "border-border/70 text-muted-foreground rounded-md border px-3 py-2 text-xs"
          }
        >
          {statusText}
        </div>
        <div
          className="border-border/70 inline-flex w-fit rounded-md border p-1"
          role="group"
          aria-label={t(
            "pages.agent.organization.detail.live_logs_scope_label",
            "Live log scope",
          )}
        >
          <Button
            type="button"
            size="sm"
            variant={scopeMode === "all" ? "secondary" : "ghost"}
            onClick={() => setScopeMode("all")}
          >
            {t("pages.agent.organization.detail.live_logs_all", "All Logs")}
          </Button>
          <Button
            type="button"
            size="sm"
            variant={scopeMode === "selected" ? "secondary" : "ghost"}
            onClick={() => setScopeMode("selected")}
          >
            {t("pages.agent.organization.detail.live_logs_selected", {
              defaultValue: "Selected Agent",
            })}
            <span className="text-muted-foreground tabular-nums">
              {referencedLogCount}
            </span>
          </Button>
        </div>
      </div>
      <div className="min-h-0 flex-1">
        <LogsPanel
          logs={visibleLogs}
          emptyMessage={
            scopeMode === "selected" ? selectedAgentEmptyMessage : undefined
          }
          getLineReferenceFields={(line) =>
            findAgentLogReferenceFields(line, agent.id)
          }
          wrapColumns={wrapColumns}
          contentRef={contentRef}
          measureRef={measureRef}
        />
      </div>
    </div>
  )
}

export function ActivityRecordDetailPanel({
  agentID,
  selectedRecord,
  onClear,
}: {
  agentID: string
  selectedRecord: AgentSelectedActivityRecord | null
  onClear: () => void
}) {
  const { t } = useTranslation()
  const detailQuery = useQuery({
    queryKey: [
      "agents",
      agentID,
      "activity-detail",
      selectedRecord?.type,
      selectedRecord?.recordID,
    ],
    queryFn: () =>
      getAgentActivityDetail(
        agentID,
        selectedRecord?.type ?? "",
        selectedRecord?.recordID ?? "",
      ),
    enabled: Boolean(selectedRecord),
  })

  if (!selectedRecord) {
    return null
  }

  const title =
    selectedRecord.title ||
    t("pages.agent.organization.detail.record_details", "Record details")

  return (
    <div className="mb-4 rounded-lg border border-border/70 bg-muted/10">
      <div className="flex min-w-0 items-start justify-between gap-3 border-b border-border/70 px-3 py-3">
        <div className="min-w-0">
          <div className="text-muted-foreground text-[11px] font-medium tracking-wide uppercase">
            {t(
              "pages.agent.organization.detail.selected_record",
              "Selected record",
            )}
          </div>
          <div className="mt-0.5 truncate text-sm font-medium">{title}</div>
          <div className="text-muted-foreground mt-0.5 truncate font-mono text-xs">
            {selectedRecord.recordID}
          </div>
        </div>
        <Button
          type="button"
          size="icon-sm"
          variant="ghost"
          aria-label={t(
            "pages.agent.organization.detail.close_record_details",
            "Close record details",
          )}
          onClick={onClear}
        >
          <IconX className="size-4" />
        </Button>
      </div>
      <div className="space-y-3 px-3 py-3">
        {detailQuery.isLoading ? (
          <TabState
            loading
            title={t(
              "pages.agent.organization.detail.loading_record_details",
              "Loading record details",
            )}
          />
        ) : detailQuery.error ? (
          <TabState
            destructive
            title={recordDetailErrorTitle(detailQuery.error, t)}
            detail={errorMessage(detailQuery.error)}
          />
        ) : detailQuery.data ? (
          <ActivityRecordDetailContent detail={detailQuery.data} />
        ) : (
          <TabState
            title={t(
              "pages.agent.organization.detail.no_record_details",
              "No record details",
            )}
            detail={t(
              "pages.agent.organization.detail.no_record_details_detail",
              "The selected record did not return diagnostic detail.",
            )}
          />
        )}
      </div>
    </div>
  )
}

function ActivityRecordDetailContent({
  detail,
}: {
  detail: AgentOrganizationActivityDetail
}) {
  const { t } = useTranslation()
  return (
    <div className="space-y-3">
      <DetailSection
        title={t("pages.agent.organization.detail.identity", "Identity")}
      >
        <div className="grid gap-2 sm:grid-cols-2">
          <RecordFact
            label={t("pages.agent.organization.detail.record_type", "Type")}
            value={t(
              `pages.agent.organization.activity_type.${detail.type}`,
              detail.type,
            )}
          />
          <RecordFact
            label={t("pages.agent.organization.detail.status", "Status")}
            value={t(
              `pages.agent.organization.status.${detail.status}`,
              detail.status,
            )}
          />
          <RecordFact
            label={t("pages.agent.organization.role_label", "Role")}
            value={
              detail.role
                ? t(`pages.agent.organization.role.${detail.role}`, detail.role)
                : t("common.notAvailable", "Unavailable")
            }
          />
          <RecordFact
            label={t("pages.agent.organization.detail.peer_agent_short", "Peer")}
            value={detail.peer_agent_id || detail.agent_id || unavailable(t)}
          />
          <RecordFact
            label={t("pages.agent.organization.detail.created", "Created")}
            value={formatTimestamp(detail.created_at, t)}
          />
          <RecordFact
            label={t("pages.agent.organization.detail.updated", "Updated")}
            value={formatTimestamp(detail.updated_at, t)}
          />
          <RecordFact
            label={t("pages.agent.organization.detail.completed", "Completed")}
            value={formatTimestamp(detail.completed_at, t)}
          />
        </div>
      </DetailSection>

      <DetailSection title={t("pages.agent.organization.detail.reason", "Reason")}>
        <DetailText value={formatDiagnosticReason(detail, t)} />
        <div className="mt-3 grid gap-2 sm:grid-cols-3">
          <RecordFact
            label={t(
              "pages.agent.organization.detail.reason_source_label",
              "Reason source",
            )}
            value={formatDiagnosticReasonSource(detail, t)}
          />
          <RecordFact
            label={t(
              "pages.agent.organization.detail.severity_label",
              "Severity",
            )}
            value={formatDiagnosticSeverity(detail, t)}
          />
          <RecordFact
            label={t(
              "pages.agent.organization.detail.current_status",
              "Current status",
            )}
            value={formatDiagnosticFreshness(detail, detail.current === true, t)}
          />
        </div>
      </DetailSection>

      <DetailSection
        title={t(
          "pages.agent.organization.detail.request_context",
          "Request and context",
        )}
      >
        <DetailText
          label={t(
            "pages.agent.organization.detail.request_summary",
            "Request summary",
          )}
          value={detail.request_summary}
        />
        <DetailText
          label={t(
            "pages.agent.organization.detail.context_summary",
            "Context summary",
          )}
          value={detail.context_summary}
        />
      </DetailSection>

      <DetailSection
        title={t("pages.agent.organization.detail.result_summary", "Result")}
      >
        <DetailText value={detail.result_summary} />
      </DetailSection>

      <DetailSection
        title={t("pages.agent.organization.detail.memory_status", "Memory")}
      >
        <div className="grid gap-2 sm:grid-cols-2">
          <RecordFact
            label={t("pages.agent.organization.detail.provider", "Provider")}
            value={detail.memory?.provider || unavailable(t)}
          />
          <RecordFact
            label={t("pages.agent.organization.detail.status", "Status")}
            value={detail.memory?.status || unavailable(t)}
          />
          <RecordFact
            label={t("pages.agent.organization.detail.memory_id", "Memory ID")}
            value={detail.memory?.memory_id || unavailable(t)}
          />
          <RecordFact
            label={t("pages.agent.organization.detail.updated", "Updated")}
            value={formatTimestamp(detail.memory?.updated_at, t)}
          />
        </div>
        <DetailText
          label={t("pages.agent.organization.detail.error", "Error")}
          value={detail.memory?.error}
        />
      </DetailSection>

      <DetailSection
        title={t("pages.agent.organization.detail.artifact_status", "Artifact")}
      >
        <div className="grid gap-2 sm:grid-cols-2">
          <RecordFact
            label={t("pages.agent.organization.detail.status", "Status")}
            value={detail.artifact?.status || unavailable(t)}
          />
          <RecordFact
            label={t("pages.agent.organization.detail.issue_id", "Issue ID")}
            value={
              typeof detail.artifact?.issue_id === "number"
                ? String(detail.artifact.issue_id)
                : unavailable(t)
            }
          />
          <RecordFact
            label={t("pages.agent.organization.detail.issue_url", "Issue URL")}
            value={detail.artifact?.issue_url || unavailable(t)}
          />
          <RecordFact
            label={t("pages.agent.organization.detail.updated", "Updated")}
            value={formatTimestamp(detail.artifact?.updated_at, t)}
          />
        </div>
        <DetailText
          label={t("pages.agent.organization.detail.error", "Error")}
          value={detail.artifact?.error}
        />
      </DetailSection>

      {detail.participants && detail.participants.length > 0 ? (
        <DetailSection
          title={t(
            "pages.agent.organization.detail.participant_status",
            "Participant status",
          )}
        >
          <div className="space-y-2">
            {detail.participants.map((participant, index) => (
              <div
                key={`${participant.agent_id ?? "participant"}:${index}`}
                className="rounded-md border border-border/70 px-3 py-2"
              >
                <div className="grid gap-2 sm:grid-cols-2">
                  <RecordFact
                    label={t(
                      "pages.agent.organization.detail.agent_id",
                      "Agent ID",
                    )}
                    value={participant.agent_id || unavailable(t)}
                  />
                  <RecordFact
                    label={t("pages.agent.organization.detail.status", "Status")}
                    value={participant.status || unavailable(t)}
                  />
                  <RecordFact
                    label={t(
                      "pages.agent.organization.detail.delegation_id",
                      "Delegation ID",
                    )}
                    value={participant.delegation_id || unavailable(t)}
                  />
                  <RecordFact
                    label={t("pages.agent.organization.detail.created", "Created")}
                    value={formatTimestamp(participant.created_at, t)}
                  />
                </div>
                <DetailText
                  label={t(
                    "pages.agent.organization.detail.summary",
                    "Summary",
                  )}
                  value={participant.summary}
                />
              </div>
            ))}
          </div>
        </DetailSection>
      ) : null}

      <DetailSection
        title={t(
          "pages.agent.organization.detail.artifact_references",
          "Artifact references",
        )}
      >
        <ArtifactSummary refs={detail.artifact_refs} />
      </DetailSection>
    </div>
  )
}

function RecordDetailsAction({
  available,
  title,
  record,
  onSelectRecord,
}: {
  available?: boolean
  title: string
  record: Omit<AgentSelectedActivityRecord, "title">
  onSelectRecord: (record: AgentSelectedActivityRecord) => void
}) {
  const { t } = useTranslation()
  const selectableRecord = resolveSelectableActivityRecord(
    { ...record, title },
    available,
  )
  if (!selectableRecord) {
    return null
  }

  return (
    <div className="mt-3">
      <Button
        type="button"
        size="sm"
        variant="outline"
        onClick={() => onSelectRecord(selectableRecord)}
      >
        <IconListDetails className="size-4" />
        {t("pages.agent.organization.detail.details_action", "Details")}
      </Button>
    </div>
  )
}

function DetailSection({
  title,
  children,
}: {
  title: string
  children: ReactNode
}) {
  return (
    <section className="rounded-md border border-border/70 bg-background/70 px-3 py-3">
      <h3 className="text-muted-foreground text-[11px] font-medium tracking-wide uppercase">
        {title}
      </h3>
      <div className="mt-2">{children}</div>
    </section>
  )
}

function DetailText({ label, value }: { label?: string; value?: string }) {
  const { t } = useTranslation()
  const text = value?.trim() || unavailable(t)
  return (
    <div className={label ? "mt-2" : undefined}>
      {label ? (
        <div className="text-muted-foreground text-[11px] leading-4">
          {label}
        </div>
      ) : null}
      <div className="mt-0.5 whitespace-pre-wrap break-words text-xs leading-5">
        {text}
      </div>
    </div>
  )
}

function recordDetailErrorTitle(
  error: Error,
  t: ReturnType<typeof useTranslation>["t"],
) {
  if (error instanceof ApiRequestError && error.status === 404) {
    return t(
      "pages.agent.organization.detail.record_detail_not_found",
      "Record details not found",
    )
  }
  if (error instanceof ApiRequestError && error.status === 403) {
    return t(
      "pages.agent.organization.detail.record_detail_permission_denied",
      "Permission denied",
    )
  }
  return t(
    "pages.agent.organization.detail.record_detail_load_error",
    "Failed to load record details",
  )
}

function unavailable(t: ReturnType<typeof useTranslation>["t"]) {
  return t("common.notAvailable", "Unavailable")
}

export function DelegationRecordsPanel({
  agentID,
  label,
  query,
  sourceSection,
  onSelectRecord,
}: {
  agentID: string
  label: string
  query: UseQueryResult<Awaited<ReturnType<typeof getAgentInbox>>, Error>
  sourceSection: "inbox" | "outbox"
  onSelectRecord: (record: AgentSelectedActivityRecord) => void
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
          sourceSection={sourceSection}
          onSelectRecord={onSelectRecord}
        />
      ))}
    </div>
  )
}

function DelegationRecordItem({
  agentID,
  record,
  sourceSection,
  onSelectRecord,
}: {
  agentID: string
  record: AgentDelegationActivityRecord
  sourceSection: "inbox" | "outbox"
  onSelectRecord: (record: AgentSelectedActivityRecord) => void
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
      <RecordDetailsAction
        available={record.detail_available}
        title={t("pages.agent.organization.detail.delegation_title", {
          defaultValue: "Delegation {{id}}",
          id: shortRecordID(record.delegation_id),
        })}
        record={{
          type: "delegation",
          recordID: record.delegation_id,
          sourceSection,
        }}
        onSelectRecord={onSelectRecord}
      />
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
  onSelectRecord,
}: {
  query: UseQueryResult<Awaited<ReturnType<typeof getAgentMeetings>>, Error>
  onSelectRecord: (record: AgentSelectedActivityRecord) => void
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
        <MeetingRecordItem
          key={record.meeting_id}
          record={record}
          onSelectRecord={onSelectRecord}
        />
      ))}
    </div>
  )
}

function MeetingRecordItem({
  record,
  onSelectRecord,
}: {
  record: AgentMeetingActivityRecord
  onSelectRecord: (record: AgentSelectedActivityRecord) => void
}) {
  const { t } = useTranslation()
  const title =
    record.title ||
    t("pages.agent.organization.detail.meeting_title", {
      defaultValue: "Meeting {{id}}",
      id: shortRecordID(record.meeting_id),
    })
  return (
    <ActivityRecordFrame status={record.status}>
      <div className="flex min-w-0 flex-wrap items-start justify-between gap-2">
        <div className="min-w-0">
          <div className="truncate text-sm font-medium">
            {title}
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
      <RecordDetailsAction
        available={record.detail_available}
        title={title}
        record={{
          type: "meeting",
          recordID: record.meeting_id,
          sourceSection: "meetings",
        }}
        onSelectRecord={onSelectRecord}
      />
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

export function FailureRecordsPanel({
  agent,
  query,
  onSelectRecord,
}: {
  agent: AgentOrganizationAgent
  query: UseQueryResult<Awaited<ReturnType<typeof getAgentFailures>>, Error>
  onSelectRecord: (record: AgentSelectedActivityRecord) => void
}) {
  const { t } = useTranslation()
  const current = agent.activity.current
  const lastFailure = agent.activity.last_failure
  const records = buildFailureDrilldownRecords(
    current,
    lastFailure,
    query.data?.records,
  )

  if (query.isLoading && !query.data && records.length === 0) {
    return (
      <TabState
        loading
        title={t("pages.agent.organization.detail.loading", "Loading records")}
      />
    )
  }

  if (query.error && !query.data && records.length === 0) {
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

  if (records.length === 0) {
    return (
      <TabState
        title={t("pages.agent.organization.detail.no_failures", "No failures")}
        detail={t(
          "pages.agent.organization.detail.no_failures_detail",
          "This agent has no visible failure records.",
        )}
      />
    )
  }

  return (
    <div className="space-y-2">
      {query.error && !query.data ? (
        <div className="border-destructive/30 bg-destructive/10 text-destructive rounded-md border px-3 py-2 text-xs">
          {t("pages.agent.organization.detail.load_error", {
            defaultValue: "Failed to load records: {{message}}",
            message: errorMessage(query.error),
          })}
        </div>
      ) : null}
      {agent.activity.failure_count > records.length ? (
        <div className="border-border/70 bg-muted/20 text-muted-foreground rounded-md border px-3 py-2 text-xs">
          {t("pages.agent.organization.detail.failure_list_partial", {
            defaultValue:
              "Showing {{shown}} of {{total}} visible failure records.",
            shown: records.length,
            total: agent.activity.failure_count,
          })}
        </div>
      ) : null}
      {lastFailure &&
      !records.some((record) => sameActivityRecord(record, lastFailure)) ? (
        <div className="border-border/70 bg-muted/20 text-muted-foreground rounded-md border px-3 py-2 text-xs">
          {current
            ? t("pages.agent.organization.detail.stale_failure_notice", {
                defaultValue:
                  "Last failure is historical. Newer current activity is {{activity}}.",
                activity: summarizeActivity(current, t),
              })
            : t(
                "pages.agent.organization.detail.last_failure_notice",
                "Last failure is not the current activity.",
              )}
        </div>
      ) : null}
      {records.map((record) => (
        <FailureRecordItem
          key={`${record.type}:${record.record_id}`}
          current={
            record.stale !== true &&
            (record.current === true || sameActivityRecord(record, current))
          }
          record={record}
          onSelectRecord={onSelectRecord}
        />
      ))}
    </div>
  )
}

function FailureRecordItem({
  current,
  record,
  onSelectRecord,
}: {
  current: boolean
  record: AgentOrganizationActivityRecord
  onSelectRecord: (record: AgentSelectedActivityRecord) => void
}) {
  const { t } = useTranslation()
  const reason = formatDiagnosticReason(record, t)
  const source = formatDiagnosticReasonSource(record, t)
  const severity = formatDiagnosticSeverity(record, t)
  const freshness = formatDiagnosticFreshness(record, current, t)
  return (
    <ActivityRecordFrame
      status={record.status}
      tone={current ? "auto" : "muted"}
    >
      <div className="flex min-w-0 flex-wrap items-start justify-between gap-2">
        <div className="min-w-0">
          <div className="truncate text-sm font-medium">
            {current
              ? t(
                  "pages.agent.organization.detail.current_failure",
                  "Current failure",
                )
              : t(
                  "pages.agent.organization.detail.historical_failure",
                  "Historical failure",
                )}
          </div>
          <div className="text-muted-foreground mt-0.5 truncate font-mono text-xs">
            {record.record_id}
          </div>
        </div>
        <StatusBadge status={record.status} />
      </div>
      <RecordDetailsAction
        available={record.detail_available}
        title={
          current
            ? t(
                "pages.agent.organization.detail.current_failure",
                "Current failure",
              )
            : t(
                "pages.agent.organization.detail.historical_failure",
                "Historical failure",
              )
        }
        record={{
          type: record.type,
          recordID: record.record_id,
          sourceSection: "failures",
        }}
        onSelectRecord={onSelectRecord}
      />
      <div className="mt-3 rounded-md border border-border/70 bg-background/60 px-3 py-2">
        <div className="text-muted-foreground text-[11px] leading-4">
          {t("pages.agent.organization.detail.failure_reason", "Reason")}
        </div>
        <div className="mt-0.5 line-clamp-2 text-xs leading-5 break-words">
          {reason}
        </div>
      </div>
      <div className="mt-3 grid gap-2 sm:grid-cols-2">
        <RecordFact
          label={t(
            "pages.agent.organization.detail.current_status",
            "Current status",
          )}
          value={freshness}
        />
        <RecordFact
          label={t("pages.agent.organization.detail.severity_label", "Severity")}
          value={severity}
        />
        <RecordFact
          label={t(
            "pages.agent.organization.detail.reason_source_label",
            "Reason source",
          )}
          value={source}
        />
        <RecordFact
          label={t("pages.agent.organization.detail.record_type", "Type")}
          value={t(
            `pages.agent.organization.activity_type.${record.type}`,
            record.type,
          )}
        />
        <RecordFact
          label={t("pages.agent.organization.role_label", "Role")}
          value={
            record.role
              ? t(`pages.agent.organization.role.${record.role}`, record.role)
              : t("common.notAvailable", "Unavailable")
          }
        />
        <RecordFact
          label={t("pages.agent.organization.detail.peer_agent_short", "Peer")}
          value={record.agent_id || t("common.notAvailable", "Unavailable")}
        />
        <RecordFact
          label={t("pages.agent.organization.detail.status", "Status")}
          value={t(
            `pages.agent.organization.status.${record.status}`,
            record.status,
          )}
        />
        <RecordFact
          label={t("pages.agent.organization.detail.created", "Created")}
          value={formatTimestamp(record.created_at, t)}
        />
        <RecordFact
          label={t("pages.agent.organization.detail.updated", "Updated")}
          value={formatTimestamp(record.updated_at, t)}
        />
        <RecordFact
          label={t("pages.agent.organization.detail.completed", "Completed")}
          value={formatTimestamp(record.completed_at, t)}
        />
      </div>
      <ArtifactSummary refs={record.artifact_refs} />
    </ActivityRecordFrame>
  )
}

function sameActivityRecord(
  a: AgentOrganizationActivityRecord | undefined,
  b: AgentOrganizationActivityRecord | undefined,
) {
  return Boolean(a && b && a.type === b.type && a.record_id === b.record_id)
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
