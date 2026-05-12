# Zehn Prompt, Memory, Runbook, And Cron Remediation Todo - 2026-05-12

Status: remediated; verification passed locally on 2026-05-12.

Source audit:
`supervision/ZEHN_PROMPT_MEMORY_RUNBOOK_AUDIT_20260512.md`

Rules for this remediation pass:

- Do not restart or reload Zehn during this pass.
- Fix canonical files first, then repeated agent snippets.
- Keep runtime behavior local-first and Ali-allowlisted.
- Do not weaken approval boundaries for production, customer data, secrets,
  auth/payment/billing/migrations, broad infrastructure, external/public
  commitments, or direct main pushes.
- Make drift detectable with a verification script before declaring complete.

## Todo

- [x] F-001: Add or explicitly decide DevOps, QA, Security, and Docs scheduled
  queues.
- [x] F-002: Update cron payloads to include active PR review queue before
  `HEARTBEAT_OK`.
- [x] F-003: Resolve missing `verify-pr.sh` policy by using a consistent
  "preferred when present, documented fallback when absent" rule until the
  wrapper exists on the active branch/main.
- [x] F-004: Bound CTO/engineering checks around active PRs/issues and avoid
  tool-budget exhaustion.
- [x] F-005: Mark old readiness audit stale and point agents to current state.
- [x] F-006: Clarify original suites vs Ignite packaging taxonomy.
- [x] F-007: Fix software-delivery GitHub authority for approved internal
  project lanes.
- [x] F-008: Align standard verification policy everywhere.
- [x] F-009: Clarify trusted service-control/post-merge path and blocked
  process-control behavior.
- [x] F-010: Mark historical evidence files as historical and require fresh
  state checks.
- [x] F-011: Add DevOps, QA, Security, Docs operational scheduled queues if
  still required.
- [x] F-012: Update specialist `AGENT.md` role summaries to include active PR
  review, stale blocker cleanup, and post-merge handoff.
- [x] F-013: Fix CEO/engineering unconditional `verify-pr.sh` wording.
- [x] F-014: Add a canonical current-state summary and mark setup planning as
  historical.
- [x] F-015: Resolve live bundle naming conflict with provisional mapping.
- [x] F-016: Record post-merge script as unproven until controlled test.
- [x] F-017: Clarify heartbeat contract versus cron/tooling monitors.
- [x] F-018: Schedule Zehn operations monitor or mark it manual.
- [x] F-019: Add rule that delegation/meeting records are historical evidence,
  not current instructions.
- [x] F-020: Remove live-instruction reliance on 87-agent language.
- [x] F-021: Mark old draft-PR planning as historical; keep normal PR policy
  canonical.
- [x] F-022: Replace app-owner responsibility wording in solution portfolio
  plan.
- [x] F-023: Split new repo approval from approved repo issue/branch/PR
  standing authority.
- [x] F-024: Correct Discord mention-only planning text against current config.
- [x] F-025: Wire or relabel Zehn operations monitor.
- [x] F-026: End with deliberate git state for audit/script files.
- [x] F-027: Mark issue/PR body draft files as drafts requiring GitHub
  revalidation.
- [x] F-028: Add stale-phrase verification script.

## Verification Evidence

Passed locally:

```bash
jq empty .picoclaw/workspace/cron/jobs.json
bash -n operations/logicigniter-post-merge-reconcile.sh
bash -n operations/verify-zehn-prompt-memory-remediation.sh
operations/verify-zehn-prompt-memory-remediation.sh
```

Residual live-test note: the post-merge reconciliation script is the trusted
path, but remains marked trusted-but-not-live-proven until Zehn runs it from the
active runtime after a real merged PR and reports successful health/repo-hygiene
evidence.

## Completion Criteria

- Current-state memory exists and is referenced by live prompts.
- Cron JSON validates with `jq`.
- Scheduled queue messages mention active PR review before `HEARTBEAT_OK`.
- Canonical docs share the same verification fallback rule.
- Specialist boot files no longer imply issue-only ownership.
- Stale historical sources are clearly labeled as non-authoritative.
- Verification script fails on known stale live-instruction phrases.
