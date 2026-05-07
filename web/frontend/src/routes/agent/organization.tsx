import { createFileRoute } from "@tanstack/react-router"

import { OrganizationPage } from "@/components/agent/organization/organization-page"

export const Route = createFileRoute("/agent/organization")({
  component: AgentOrganizationRoute,
})

function AgentOrganizationRoute() {
  return <OrganizationPage />
}
