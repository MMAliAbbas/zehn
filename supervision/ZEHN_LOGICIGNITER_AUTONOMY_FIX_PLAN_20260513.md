# Zehn LogicIgniter Autonomy Fix Plan

Date: 2026-05-13
Status: planning document, no runtime changes applied

## Purpose

This plan corrects the current LogicIgniter automation path so Zehn can move
from monitoring and repeated review loops into a complete autonomous delivery
cycle:

issue -> claim -> branch -> implementation -> verification -> PR -> review ->
merge when approved -> post-merge reconcile -> memory/project update.

The plan also corrects an important architecture boundary: MCP and final
readiness must prove behavior through runtime APIs, not by requiring direct
access to the identity database.

## Current Evidence

- `scripts/local-preview/start-mcp-runtime-proof.sh` is now host-native and no
  longer starts Docker.
- `scripts/local-preview/start-real-stack.sh` and `stop-real-stack.sh` are
  still Docker Compose based.
- `scripts/local-preview/README.md` still tells agents to use
  `start-real-stack.sh --detach`, which is wrong for this machine.
- GitHub API reports zero workflows for:
  - `logicigniter/integration_tests`
  - `logicigniter/scripts`
- Open PRs in those repos therefore show empty `statusCheckRollup` arrays.
- Zehn logs show provider warnings from successful Codex stream
  reconstruction. These are noisy but not currently functional failures.
- Zehn logs show `Unknown channel for outbound message` for internal
  delegation traffic. This is noisy and should be cleaned later, but it is not
  the primary LogicIgniter delivery blocker.
- Logs and reports still do not prove one complete autonomous
  issue-to-merge-to-reconcile cycle after the latest restart.

## Correct DB Boundary

Final readiness and MCP runtime proof must not depend on direct identity
database access.

Correct primary readiness contract:

1. Keycloak realm is reachable through the configured local/public URL.
2. `svc-identity` is reachable through its service API.
3. BFF auth and subscribe paths work through service APIs.
4. MCP endpoint can list and call permitted tools using the expected auth path.
5. Failures report the exact API-level symptom and the responsible service
   boundary.

Incorrect primary readiness contract:

- Requiring final readiness, MCP runtime proof, or normal automation to inspect
  `svc_identity` tables directly.
- Treating `IDENTITY_DB_DSN` as a required input for normal MCP readiness.
- Seeding tenants, API keys, or bundle subscriptions by direct DB writes in
  normal integration flow.

Allowed exception:

- A separate local-only diagnostic script may inspect the identity DB when an
  API-level readiness check fails and deeper diagnosis is requested. That
  script must be named and documented as diagnostic-only, not as the readiness
  contract.

## Revised Fix Plan

### 1. Retire Docker-First Local Preview Guidance

Goal: stop agents and scripts from choosing Docker paths on this machine.

Required changes:

- Replace `scripts/local-preview/README.md` typical flow with host-native
  Keycloak/Postgres guidance.
- Mark `start-real-stack.sh` as deprecated or convert it into a host-native
  wrapper that delegates to the current local-preview scripts.
- Update `stop-real-stack.sh` so it does not assume Docker Compose.
- Audit integration runner references to `start-real-stack.sh` and replace
  with host-native equivalents.

Acceptance:

- `rg "docker compose|start-real-stack" scripts/local-preview operations`
  shows no active recommended path that agents should use for local final
  readiness.
- Any remaining Docker mention is clearly marked legacy/manual-only.

### 2. Replace DB-Coupled MCP Preflight With API-Level Readiness

Goal: validate MCP readiness through real service boundaries.

Required changes:

- Add or update a host-native readiness wrapper that checks:
  - Keycloak realm URL.
  - Identity service health/API readiness.
  - BFF health/API readiness.
  - BFF subscribe/auth path for the expected local persona/API-key flow.
  - MCP tool catalog and one safe read-only MCP call.
- Remove `IDENTITY_DB_DSN` as a normal prerequisite for MCP/final-readiness
  proof unless the specific test is explicitly a migration/seed diagnostic.
- Keep direct DB checks only in a separate diagnostic command, for example:
  `scripts/local-preview/diagnose-identity-db.sh`.

Acceptance:

- A failed readiness run identifies the API boundary that failed.
- Normal final readiness can run without direct DB inspection.
- Diagnostic DB access is opt-in and clearly separate.

### 3. Establish PR Verification For Repos Without GitHub Checks

Goal: prevent agents from stalling on empty GitHub check state.

Required changes:

- Add minimal GitHub Actions workflows for repos where automated checks should
  exist, starting with:
  - `integration_tests`
  - `scripts`
- Until workflows exist, define empty `statusCheckRollup` as neither pass nor
  fail; it requires local `verify-pr` evidence and review approval.
- Teach agent prompts/runbooks that Codex `COMMENTED` is not approval and
  `eyes` is only review-started.

Acceptance:

- PRs either have GitHub checks or an explicit local verification artifact.
- Agents stop treating “no checks reported” as an unresolved mystery.

### 4. Split Long Verification Into Staged Evidence

Goal: avoid one giant command reaching the 12-minute limit without useful
evidence.

Required changes:

- Define a staged verification sequence:
  1. host infrastructure readiness;
  2. service readiness;
  3. BFF/MCP auth path;
  4. targeted integration package;
  5. final readiness.
- Each stage writes an evidence file or PR comment summary.
- Agents should run the smallest stage that matches the issue scope before
  running full final readiness.

Acceptance:

- A timeout leaves a stage-level failure, not an ambiguous whole-system failure.
- Agents know exactly which stage to retry or delegate.

### 5. Clean Stale PR/Issue State

Goal: remove stale branches and memory from earlier failed assumptions.

Required changes:

- Reconcile open PRs whose fixes already landed on `main`.
- Close or update superseded PRs.
- Remove stale `zehn:in-progress`, `zehn:review-internal`, or blocker labels
  when the source issue is already fixed.
- Add a short Yaad/company memory update only after verified reconciliation.

Acceptance:

- Open LogicIgniter PRs represent current work only.
- Agents stop reasoning from obsolete MCP/auth blockers.

### 6. Prove One Complete Autonomous Delivery Cycle

Goal: verify Zehn can operate like a real execution system, not only a
monitoring system.

Required changes:

- Select or create one low-risk, well-scoped `zehn:ready` issue.
- Let the matching specialist claim it.
- Require branch creation, implementation, verification, PR creation, review,
  merge if approved, post-merge reconcile, project update, and memory update.
- Record the evidence in Discord/GitHub/Yaad.

Acceptance:

- One complete cycle exists with links to issue, branch, PR, verification,
  review signal, merge/reconcile result, and memory update.

## Do Not Do

- Do not make MCP readiness depend on direct DB access.
- Do not direct-write tenant/API-key/subscription rows as the normal setup
  mechanism.
- Do not treat disabled/missing GitHub checks as passed.
- Do not merge based on Codex `COMMENTED` or `eyes` alone.
- Do not let agents leave any LogicIgniter repo dirty.
- Do not patch around auth by weakening validation.

## Implementation Order

1. Fix local-preview documentation and remove Docker-first guidance.
2. Add API-level readiness wrapper and separate DB diagnostic script.
3. Add or document PR verification policy for repos with no GitHub workflows.
4. Reconcile stale open PRs/issues/memory.
5. Run one controlled autonomous delivery proof.

This order avoids another loop where agents keep retrying old Docker or DB
assumptions while the actual delivery process remains unproven.
