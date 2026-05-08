import {
  IconAlertTriangle,
  IconCircleCheck,
  IconFileDescription,
  IconLoader2,
} from "@tabler/icons-react"
import type { ReactNode } from "react"
import { useTranslation } from "react-i18next"

import { cn } from "@/lib/utils"

import { isProblemStatus } from "./formatting"

export function TabState({
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

export function ActivityRecordFrame({
  status,
  tone = "auto",
  children,
}: {
  status: string
  tone?: "auto" | "muted"
  children: ReactNode
}) {
  const isProblem = tone !== "muted" && isProblemStatus(status)
  return (
    <article
      className={cn(
        "border-border/70 rounded-lg border px-3 py-3 text-sm",
        isProblem && "border-destructive/30 bg-destructive/3",
        tone === "muted" && "bg-muted/20",
      )}
    >
      {children}
    </article>
  )
}

export function RecordFact({ label, value }: { label: string; value: string }) {
  return (
    <div className="min-w-0">
      <div className="text-muted-foreground text-[11px] leading-4">{label}</div>
      <div className="truncate text-xs leading-5" title={value}>
        {value}
      </div>
    </div>
  )
}

export function ArtifactSummary({ refs }: { refs?: string[] }) {
  const { t } = useTranslation()
  const artifactRefs = refs ?? []
  const count = artifactRefs.length
  return (
    <div className="text-muted-foreground mt-3 min-w-0 text-xs">
      <div className="flex min-w-0 items-center gap-1.5">
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
      {count > 0 ? (
        <div className="mt-2 space-y-1">
          {artifactRefs.map((ref) => (
            <div key={ref} className="truncate font-mono" title={ref}>
              {ref}
            </div>
          ))}
        </div>
      ) : null}
    </div>
  )
}
