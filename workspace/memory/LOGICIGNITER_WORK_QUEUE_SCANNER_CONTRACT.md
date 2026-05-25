# LogicIgniter Work Queue Scanner Contract

Status: canonical COO input contract.
Owner: `li-coo`.

The COO must use the deterministic scanner before deciding whether work is
ready, blocked, approval-gated, malformed, or idle.

Scanner command:

```bash
/Users/aliai/zehn/operations/logicigniter-work-queue-scan.py
```

The scanner is label-driven. It must not count broad text-search matches as
blocked work. If a GitHub item is not labeled according to the contract, it is
not canonical queue state; it is a normalization problem.

## Canonical Labels

- `zehn:ready`
- `zehn:claimed`
- `zehn:in-progress`
- `zehn:blocked`
- `zehn:retrying`
- `approval:ali-required`
- `type:continuation`
- `area:backend`
- `area:frontend`
- `area:ux`
- `area:integration`
- `area:data-ai`
- `area:architecture`
- `area:devops`
- `area:qa`
- `area:security`
- `area:docs`
- `area:product`
- `area:finance`
- `area:legal`
- `area:revenue`
- `area:marketing`
- `area:cco`

Other `area:*` labels are allowed but route to `li-coo` until explicitly
mapped.

## Scanner Output

The scanner returns JSON with:

- `ready`
- `in_progress`
- `open_prs`
- `blocked`
- `approval_gated`
- `malformed`
- `continuation`
- `unblock_candidates`
- `next_action`

`next_action.type` is one of:

- `REVIEW_PR`
- `CLAIM_READY`
- `REWORK_BLOCKER`
- `UNBLOCK_DISPATCHED`
- `APPROVAL_REQUEST`
- `NORMALIZE_ISSUE`
- `NO_CHANGED_STATE`
- `SOURCE_UNAVAILABLE`

The COO must act on `next_action`. Reporting counts without acting on
`next_action` is an invalid operating cycle.

## Source Failure

If the scanner exits non-zero, the COO returns `SOURCE_UNAVAILABLE` with the
failed command and retry checkpoint. Do not replace a failed scanner with a
broad ad hoc search.
