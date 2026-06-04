# LogicIgniter Company Utilization Contract

Status: canonical operating contract.
Owner: `li-ceo` for company accountability; `li-coo` for dispatch hygiene.

This contract prevents LogicIgniter from becoming a CEO/COO status loop. If
LogicIgniter has active initiatives, the organization must maintain useful work
or an explicit reason for idleness across departments, bundle owners, and
specialists.

## Utilization State

Every visible LogicIgniter role must be classified during utilization checks:

- `active-owner`: owns a current issue, PR, delegation, meeting output, or
  dated deliverable.
- `ready-for-dispatch`: has a safe next lane tied to an active initiative.
- `approval-blocked`: cannot proceed without Ali approval or an external
  authority action.
- `intentionally-idle`: idle by design, with dated reason and review date.
- `not-applicable-now`: role is outside the current initiative scope.

Unknown is not an acceptable durable state. If state cannot be checked because
GitHub, Yaad, or local evidence is unavailable, report `SOURCE_UNAVAILABLE`.

## Automatic Operating Requirement

Heartbeat must cause `li-ceo` to check company utilization automatically. Ali
must not need to ask CEO to "use the team" when active initiatives exist.

Each CEO operating cycle must do two bounded things:

1. Select or inspect the highest-priority operating action through COO.
2. Ensure the organization has a current utilization state.

The utilization check is not a request to wake every agent. It is a bounded
dispatch pass:

- dispatch at most five new role assignments per cycle;
- prefer roles that are idle, stale, or directly relevant to active P0/P1
  initiatives;
- do not duplicate an active delegation, issue claim, or PR review;
- do not invent busywork;
- every assignment must point to a GitHub issue, PR, Yaad decision, local
  artifact, or explicit issue spec.

## COO Dispatch Board Duties

`li-coo` owns the dispatch board and stale-idle detection. The board can live in
GitHub issues/project state plus Yaad/local summaries; it does not require a new
database.

During a utilization pass, COO must identify:

- active departments and roles;
- active initiatives they support;
- current owner/evidence link;
- stale roles with no movement;
- claimable ready work by `area:*`;
- bundle owners lacking a readiness lane;
- completed work with missing successor issues;
- approval-blocked roles with exact approval question.

COO must dispatch or propose work for the highest-value idle/stale roles,
bounded by the five-assignment limit.

## Role Activation Lanes

- `li-cpo` activates product and bundle-owner readiness lanes.
- `li-cto` and `li-engineering` activate architecture and specialist execution
  lanes.
- `li-coo` activates operations, stale-WIP, dispatch-board, and GitHub hygiene
  lanes.
- `li-cro`, `li-cmo`, and `li-cco` activate GTM, sales, onboarding, and
  customer-readiness lanes.
- `li-cfo`, `li-legal`, `li-ciso`, `li-security`, and `li-cdao` activate
  finance, legal, security, compliance, evidence, data, and memory lanes.
- Bundle owners activate bundle-specific readiness, demo, acceptance, and risk
  packets through `li-cpo`.

## Heartbeat Acceptance

`HEARTBEAT_OK` is invalid when:

- active initiatives exist and no utilization state is current;
- more than five active roles are unknown/idle without reason;
- any P0 initiative has no active owner, blocker owner, or next dispatch
  candidate;
- bundle owners have no readiness lane while portfolio launch is active;
- a utilization pass created issues but no follow-up claim/dispatch happened;
- a previous utilization delegation is stale or still running past its
  checkpoint without evidence.

Valid quiet state requires every active role to have work, a blocker, a dated
defer reason, or a clear not-applicable classification.
