# LogicIgniter Terminal State Machine

Status: canonical operating-control rule.
Owner: `li-ceo` for company priority, `li-coo` for execution flow.

This artifact prevents heartbeat and operating checks from becoming status
loops. Every company operating tick must either change one piece of execution
state or return one terminal outcome with evidence.

## Terminal Outcomes

Use exactly one of these outcomes in CEO/COO operating summaries:

- `DISPATCHED`: one role-matched owner was assigned executable work.
- `UNBLOCK_DISPATCHED`: one role-matched owner was assigned blocker-removal
  work.
- `BLOCKER_ISSUE_CREATED`: one GitHub issue was created or repaired to make a
  blocker executable.
- `DEFERRED_WITH_RETRY`: work cannot move now, but has a concrete owner and
  retry date.
- `INVALID_BLOCKER_CLASSIFICATION`: a reported blocker was corrected because
  canonical scanner labels did not support the classification.
- `READY_TO_MERGE`: a PR is green and waiting only for approved merge action.
- `REVIEW_BLOCKED`: a PR needs specific reviewer action or PR-body/traceability
  repair before merge.
- `OWNER_BLOCKED`: work is blocked by a named owner/action/retry date.
- `ALI_APPROVAL_REQUIRED`: Ali must answer one precise approval question.
- `DIRTY_REPO_STOP`: repo state is dirty and must be reconciled before more
  work in that repo.
- `SOURCE_UNAVAILABLE`: GitHub, Yaad, provider, channel, or required local
  source could not be checked.
- `MERGED_RECONCILED`: PR was merged and post-merge reconcile completed.
- `NO_CHANGED_STATE`: live state was checked, no useful action exists, and
  quiet criteria are met.
- `ACTIVE_CHECK_RUNNING`: a prior async CEO/company cycle is still active and
  has a valid owner/checkpoint.

## One-Tick Rule

A heartbeat-triggered LogicIgniter check is a control tick, not an
implementation session.

Each tick may perform at most one changed-state action:

- dispatch one specialist;
- launch or inspect one async CEO/company operating cycle;
- repair one malformed issue or PR body;
- route one review/merge/reconcile action;
- ask one Ali approval question;
- write one changed-state Yaad summary;
- declare one terminal blocker with owner and retry date.

If more work is discovered, rank it and leave the rest for later ticks.

## Duplicate Suppression

Before posting a GitHub comment, Discord update, or Yaad summary, compare the
current work state to the last visible terminal update.

State identity should include the relevant subset of:

- initiative ID;
- repo and issue/PR number;
- labels and project status;
- PR head SHA;
- check conclusion;
- last terminal outcome;
- blocker owner and retry date;
- dirty repo status;
- last comment or Yaad memory ID when visible.

If the state is unchanged, do not post another status comment. Return
`NO_CHANGED_STATE` with the evidence pointer, or stay quiet when the caller
allows `HEARTBEAT_OK`.

## Active Work Rule

Do not start a second delegation for the same initiative, issue, PR, or repo
lane while an earlier matching delegation is still active or has no terminal
outcome. Report the active delegation ID, owner, and next checkpoint instead.

## Evidence Rule

Every non-quiet outcome must include:

- owner agent;
- target repo/issue/PR or initiative;
- evidence pointer;
- next checkpoint or retry date;
- whether Ali approval is needed.

Status without owner, evidence, and next checkpoint is not a valid operating
outcome.
