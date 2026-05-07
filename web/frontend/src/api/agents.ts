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

export interface AgentOrganizationAgentActivity {
  inbox_count: number
  outbox_count: number
  meeting_count: number
  failure_count: number
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
