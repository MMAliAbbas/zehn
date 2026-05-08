export type AgentLogScopeMode = "all" | "selected"

export const AGENT_LOG_REFERENCE_FIELDS = [
  "agent_id",
  "target_agent_id",
  "parent_agent_id",
  "requester_id",
  "sponsor_agent_id",
  "chair_agent_id",
] as const

const AGENT_LOG_REFERENCE_FIELD_SET = new Set<string>(
  AGENT_LOG_REFERENCE_FIELDS,
)
const ANSI_PATTERN = new RegExp(String.raw`\u001B\[[0-9;]*m`, "g")
const KEY_VALUE_PATTERN =
  /\b(agent_id|target_agent_id|parent_agent_id|requester_id|sponsor_agent_id|chair_agent_id)=(?:"([^"]*)"|'([^']*)'|([^\s,}]+))/g

type AgentReferenceField = (typeof AGENT_LOG_REFERENCE_FIELDS)[number]

function stripAnsi(input: string): string {
  return input.replace(ANSI_PATTERN, "")
}

function valueMatchesAgentID(value: unknown, agentID: string): boolean {
  if (typeof value === "string") {
    return value === agentID
  }

  if (Array.isArray(value)) {
    return value.some((item) => valueMatchesAgentID(item, agentID))
  }

  return false
}

function collectJSONMatches(
  value: unknown,
  agentID: string,
  matches: Set<AgentReferenceField>,
) {
  if (!value || typeof value !== "object") {
    return
  }

  if (Array.isArray(value)) {
    value.forEach((item) => collectJSONMatches(item, agentID, matches))
    return
  }

  Object.entries(value as Record<string, unknown>).forEach(([key, item]) => {
    if (
      AGENT_LOG_REFERENCE_FIELD_SET.has(key) &&
      valueMatchesAgentID(item, agentID)
    ) {
      matches.add(key as AgentReferenceField)
    }

    if (item && typeof item === "object") {
      collectJSONMatches(item, agentID, matches)
    }
  })
}

function findJSONPayload(input: string): unknown {
  const start = input.indexOf("{")
  const end = input.lastIndexOf("}")

  if (start < 0 || end <= start) {
    return null
  }

  try {
    return JSON.parse(input.slice(start, end + 1))
  } catch {
    return null
  }
}

export function findAgentLogReferenceFields(
  line: string,
  agentID: string,
): AgentReferenceField[] {
  if (!agentID) {
    return []
  }

  const normalized = stripAnsi(line)
  const matches = new Set<AgentReferenceField>()
  collectJSONMatches(findJSONPayload(normalized), agentID, matches)

  KEY_VALUE_PATTERN.lastIndex = 0
  let match: RegExpExecArray | null
  while ((match = KEY_VALUE_PATTERN.exec(normalized)) !== null) {
    const field = match[1] as AgentReferenceField
    const value = match[2] ?? match[3] ?? match[4] ?? ""
    if (value === agentID) {
      matches.add(field)
    }
  }

  return AGENT_LOG_REFERENCE_FIELDS.filter((field) => matches.has(field))
}

export function filterAgentLogLines(
  logs: string[],
  agentID: string,
  mode: AgentLogScopeMode,
): string[] {
  if (mode === "all" || !agentID) {
    return logs
  }

  return logs.filter(
    (line) => findAgentLogReferenceFields(line, agentID).length > 0,
  )
}
