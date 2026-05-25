# LogicIgniter Blocker Remediation Contract

Status: canonical blocker handling contract.
Owner: `li-coo`.

Blocked work is executable operating work. A blocked count is not a terminal
outcome unless each blocker has an owner, next action, and retry or approval
path.

## Required Outcomes

When the scanner returns `UNBLOCK_DISPATCHED`, `APPROVAL_REQUEST`, or
`NORMALIZE_ISSUE`, COO must produce exactly one of:

- `UNBLOCK_DISPATCHED`: delegate the unblock task to the role named by the
  scanner.
- `ALI_APPROVAL_REQUIRED`: ask Ali one precise approval question with the
  GitHub link and consequence.
- `BLOCKER_ISSUE_CREATED`: create or repair one GitHub issue that makes the
  unblock path executable.
- `DEFERRED_WITH_RETRY`: record a concrete retry date and owner when the
  blocker cannot move now.
- `INVALID_BLOCKER_CLASSIFICATION`: correct labels or scanner inputs when a
  reported blocker is not a real blocker.

## Priority

When there is no claimable ready work, COO handles blockers in this order:

1. Open PRs that cannot merge.
2. Blockers on P0 active initiatives.
3. Approval-gated items with unclear approval questions.
4. Malformed work that can be normalized without Ali.
5. Stale blockers or stale claims.

## Invalid Outcomes

The following are invalid:

- returning only blocker counts;
- saying no work exists while `unblock_candidates` is non-empty;
- treating missing labels as real blocker state;
- asking Ali multiple unrelated approval questions in one cycle;
- creating duplicate comments when the blocker state has not changed.
