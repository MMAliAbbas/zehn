import { AgentCard } from "./agent-card"
import type { OrderedNode } from "./types"

export function OrganizationBranch({
  node,
  depth,
}: {
  node: OrderedNode
  depth: number
}) {
  const levels = buildLevels(node)

  return (
    <div className="min-w-0 space-y-3">
      {levels.map((level, index) => (
        <div
          key={`${node.id}:level:${index}`}
          className={index > 0 ? "border-border/60 border-t pt-3" : undefined}
        >
          <div className={gridClassForLevel(index + depth, level.length)}>
            {level.map((levelNode) => (
              <AgentCard key={levelNode.id} agent={levelNode} />
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}

function buildLevels(root: OrderedNode): OrderedNode[][] {
  const levels: OrderedNode[][] = []
  let currentLevel: OrderedNode[] = [root]

  while (currentLevel.length > 0) {
    levels.push(currentLevel)
    currentLevel = currentLevel.flatMap((node) => node.children ?? [])
  }

  return levels
}

function gridClassForLevel(depth: number, itemCount: number): string {
  if (itemCount === 1) {
    return "grid min-w-0 grid-cols-[minmax(0,1fr)] gap-2"
  }

  if (depth === 0) {
    return "grid min-w-0 grid-cols-[minmax(0,1fr)] gap-2"
  }

  return "grid min-w-0 gap-2 md:grid-cols-2 xl:grid-cols-4"
}
