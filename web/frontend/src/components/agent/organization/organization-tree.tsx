import { cn } from "@/lib/utils"

import { AgentCard } from "./agent-card"
import type { OrderedNode } from "./types"

export function OrganizationBranch({
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
