import { launcherFetch } from "@/api/http"

export type AgentOrganizationStatus =
  | "idle"
  | "working"
  | "delegating"
  | "meeting"
  | "failed"
  | string

export interface AgentOrganizationActivityRecord {
  type: string
  record_id: string
  status: string
  role?: string
  agent_id?: string
  updated_at?: string
}

export interface AgentOrganizationRecentEvent {
  source: string
  agent_id: string
  level?: string
  event?: string
  message: string
  timestamp?: string
}

export interface AgentOrganizationAgentActivity {
  inbox_count: number
  outbox_count: number
  meeting_count: number
  failure_count: number
  recent_events?: AgentOrganizationRecentEvent[]
  current?: AgentOrganizationActivityRecord
  last_failure?: AgentOrganizationActivityRecord
  last_updated_at?: string
}

export interface AgentOrganizationAgent {
  id: string
  name?: string
  label?: string
  group?: string
  workspace?: string
  status: AgentOrganizationStatus
  activity: AgentOrganizationAgentActivity
}

export interface AgentOrganizationNode extends AgentOrganizationAgent {
  children?: AgentOrganizationNode[]
}

export interface AgentOrganizationActivitySummary {
  delegation_count: number
  meeting_count: number
  failure_count: number
  active_count: number
  recent?: AgentOrganizationActivityFeed[]
}

export interface AgentOrganizationActivityFeed {
  type: string
  agent_id?: string
  record_id?: string
  status?: string
  summary?: string
  timestamp?: string
}

export interface AgentOrganizationSnapshotMetadata {
  source: string
  generated_at: string
  has_hierarchy: boolean
}

export interface AgentOrganizationSnapshot {
  roots?: AgentOrganizationNode[]
  agents?: Record<string, AgentOrganizationAgent>
  activity: AgentOrganizationActivitySummary
  metadata: AgentOrganizationSnapshotMetadata
}

export interface AgentActivityListResponse<TRecord> {
  agent_id: string
  kind: string
  limit: number
  records: TRecord[]
}

export interface AgentDelegationActivityRecord {
  delegation_id: string
  status: string
  parent_agent_id: string
  target_agent_id: string
  requester_id?: string
  role: string
  mode?: string
  priority?: string
  artifact_refs?: string[]
  created_at: string
  updated_at: string
  started_at?: string
  completed_at?: string
}

export interface AgentMeetingActivityRecord {
  meeting_id: string
  status: string
  title?: string
  sponsor_agent_id: string
  chair_agent_id: string
  participants?: string[]
  role: string
  artifact_refs?: string[]
  created_at: string
  updated_at: string
  completed_at?: string
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await launcherFetch(path, options)
  if (!res.ok) {
    throw new Error(await extractErrorMessage(res))
  }
  return res.json() as Promise<T>
}

export async function getAgentOrganization(): Promise<AgentOrganizationSnapshot> {
  return request<AgentOrganizationSnapshot>("/api/agents/organization")
}

export async function getAgentActivity(
  agentID: string,
): Promise<AgentOrganizationAgent> {
  return request<AgentOrganizationAgent>(
    `/api/agents/${encodeURIComponent(agentID)}/activity`,
  )
}

export async function getAgentInbox(
  agentID: string,
  limit?: number,
): Promise<AgentActivityListResponse<AgentDelegationActivityRecord>> {
  return request<AgentActivityListResponse<AgentDelegationActivityRecord>>(
    agentActivityPath(agentID, "inbox", limit),
  )
}

export async function getAgentOutbox(
  agentID: string,
  limit?: number,
): Promise<AgentActivityListResponse<AgentDelegationActivityRecord>> {
  return request<AgentActivityListResponse<AgentDelegationActivityRecord>>(
    agentActivityPath(agentID, "outbox", limit),
  )
}

export async function getAgentMeetings(
  agentID: string,
  limit?: number,
): Promise<AgentActivityListResponse<AgentMeetingActivityRecord>> {
  return request<AgentActivityListResponse<AgentMeetingActivityRecord>>(
    agentActivityPath(agentID, "meetings", limit),
  )
}

function agentActivityPath(agentID: string, kind: string, limit?: number) {
  const path = `/api/agents/${encodeURIComponent(agentID)}/${kind}`
  if (!limit) {
    return path
  }
  return `${path}?limit=${encodeURIComponent(String(limit))}`
}

async function extractErrorMessage(res: Response): Promise<string> {
  try {
    const raw = await res.text()
    if (raw.trim() === "") {
      return `API error: ${res.status} ${res.statusText}`
    }
    try {
      const body = JSON.parse(raw) as {
        error?: string
        errors?: string[]
      }
      if (Array.isArray(body.errors) && body.errors.length > 0) {
        return body.errors.join("; ")
      }
      if (typeof body.error === "string" && body.error.trim() !== "") {
        return body.error
      }
    } catch {
      return raw.trim()
    }
  } catch {
    // ignore invalid body
  }
  return `API error: ${res.status} ${res.statusText}`
}
