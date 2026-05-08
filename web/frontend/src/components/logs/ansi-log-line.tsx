import { Fragment, useMemo } from "react"

import { parseAnsiSegments, wrapLogLine } from "@/lib/ansi-log"
import { cn } from "@/lib/utils"

type AnsiLogLineProps = {
  referenceFields?: string[]
  line: string
  wrapColumns: number
}

export function AnsiLogLine({
  referenceFields = [],
  line,
  wrapColumns,
}: AnsiLogLineProps) {
  const segments = useMemo(() => {
    return parseAnsiSegments(wrapLogLine(line, wrapColumns))
  }, [line, wrapColumns])
  const hasReference = referenceFields.length > 0

  return (
    <div
      className={cn(
        "break-normal whitespace-pre-wrap",
        hasReference &&
          "border-l-2 border-amber-300/80 bg-amber-300/10 pl-2",
      )}
      title={
        hasReference
          ? `Selected agent reference: ${referenceFields.join(", ")}`
          : undefined
      }
    >
      {segments.map((segment, index) => (
        <Fragment key={`${index}-${segment.text.length}`}>
          <span style={segment.style}>{segment.text}</span>
        </Fragment>
      ))}
    </div>
  )
}
