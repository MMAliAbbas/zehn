#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCANNER="$ROOT/operations/logicigniter-work-queue-scan.py"
FIXTURE="$ROOT/operations/logicigniter-work-queue-fixture.json"

require() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "missing required command: $1" >&2
    exit 1
  }
}

require jq

tmp="$(mktemp)"
trap 'rm -f "$tmp" "$tmp.no-pr"' EXIT

"$SCANNER" --fixture "$FIXTURE" --today 2026-05-25 >"$tmp"

[ "$(jq -r '.next_action.type' "$tmp")" = "REVIEW_PR" ]
[ "$(jq -r '.counts.ready' "$tmp")" = "1" ]
[ "$(jq -r '.counts.blocked' "$tmp")" = "1" ]
[ "$(jq -r '.counts.approval_gated' "$tmp")" = "1" ]
[ "$(jq -r '.counts.malformed' "$tmp")" = "1" ]
[ "$(jq -r '.counts.unblock_candidates' "$tmp")" = "2" ]

jq '.prs[0].labels=[{"name":"approval:ali-required"},{"name":"area:docs"}]' "$FIXTURE" >"$tmp.no-pr"
"$SCANNER" --fixture "$tmp.no-pr" --today 2026-05-25 >"$tmp"

[ "$(jq -r '.next_action.type' "$tmp")" = "APPROVAL_REQUEST" ]
[ "$(jq -r '.next_action.owner' "$tmp")" = "li-ceo" ]

jq '.prs[0].labels=[{"name":"approval:ali-required"},{"name":"area:docs"}] |
    .prs[0].comments=[{
      "id":"fixture-comment-1",
      "url":"https://github.com/logicigniter/config/pull/50#issuecomment-fixture",
      "body":"Do not merge PR #50 as-is.\nBounded merge condition: PR #50 may merge if revised so that:\n1. LICENSE remains Copyright (c) 2026 Logic Igniter LLC.\n2. AGENTS.md distinguishes the LogicIgniter brand from the Logic Igniter LLC legal entity.\n3. No customer-facing legal claim is changed.\nNext step: revise PR."
    }]' "$FIXTURE" >"$tmp.no-pr"
"$SCANNER" --fixture "$tmp.no-pr" --today 2026-05-25 >"$tmp"

[ "$(jq -r '.next_action.type' "$tmp")" = "REWORK_BLOCKER" ]
[ "$(jq -r '.next_action.owner' "$tmp")" = "li-docs" ]
[ "$(jq -r '.next_action.target.rework_path.source_comment_id' "$tmp")" = "fixture-comment-1" ]
jq -e '.next_action.target.rework_path.conditions
  | index("LICENSE remains Copyright (c) 2026 Logic Igniter LLC.")
  and index("AGENTS.md distinguishes the LogicIgniter brand from the Logic Igniter LLC legal entity.")
  and index("No customer-facing legal claim is changed.")' "$tmp" >/dev/null

jq '.prs[0].labels=[{"name":"approval:ali-required"},{"name":"area:docs"}] |
    .prs[0].comments_error="gh api timed out"' "$FIXTURE" >"$tmp.no-pr"
"$SCANNER" --fixture "$tmp.no-pr" --today 2026-05-25 >"$tmp"

[ "$(jq -r '.next_action.type' "$tmp")" = "SOURCE_UNAVAILABLE" ]
jq -e '.source_warnings[0].error == "gh api timed out"' "$tmp" >/dev/null

jq '.prs=[
      {
        "repository":{"name":"old-repo"},
        "number":41,
        "title":"Old PR",
        "url":"https://github.com/logicigniter/old-repo/pull/41",
        "updatedAt":"2026-05-24T00:00:00Z",
        "labels":[]
      },
      {
        "repository":{"name":"new-repo"},
        "number":42,
        "title":"New PR",
        "url":"https://github.com/logicigniter/new-repo/pull/42",
        "updatedAt":"2026-05-25T00:00:00Z",
        "labels":[]
      }
    ]' "$FIXTURE" >"$tmp.no-pr"
"$SCANNER" --fixture "$tmp.no-pr" --today 2026-05-25 >"$tmp"

[ "$(jq -r '.next_action.target.number' "$tmp")" = "42" ]

jq '.prs=[]' "$FIXTURE" >"$tmp.no-pr"
"$SCANNER" --fixture "$tmp.no-pr" --today 2026-05-25 >"$tmp"

[ "$(jq -r '.next_action.type' "$tmp")" = "CLAIM_READY" ]
[ "$(jq -r '.next_action.owner' "$tmp")" = "li-backend-developer" ]

jq '.issues |= map(select(.number != 10 and .number != 30)) | .prs=[]' "$FIXTURE" >"$tmp.no-pr"
"$SCANNER" --fixture "$tmp.no-pr" --today 2026-05-25 >"$tmp"

[ "$(jq -r '.next_action.type' "$tmp")" = "APPROVAL_REQUEST" ]
[ "$(jq -r '.next_action.owner' "$tmp")" = "li-ceo" ]

jq '.issues |= map(select(.number != 10 and .number != 20)) | .prs=[]' "$FIXTURE" >"$tmp.no-pr"
"$SCANNER" --fixture "$tmp.no-pr" --today 2026-05-25 >"$tmp"

[ "$(jq -r '.next_action.type' "$tmp")" = "UNBLOCK_DISPATCHED" ]
[ "$(jq -r '.next_action.owner' "$tmp")" = "li-devops" ]

jq '.issues=[
      {
        "repository":{"name":"svc-logicigniter-web"},
        "number":124,
        "title":"Confirm legal entity wording before public web copy update",
        "url":"https://github.com/logicigniter/svc-logicigniter-web/issues/124",
        "updatedAt":"2026-05-25T00:00:00Z",
        "labels":[
          {"name":"zehn:blocked"},
          {"name":"area:frontend"},
          {"name":"area:docs"},
          {"name":"area:product"}
        ]
      }
    ] | .prs=[]' "$FIXTURE" >"$tmp.no-pr"
"$SCANNER" --fixture "$tmp.no-pr" --today 2026-05-25 >"$tmp"

[ "$(jq -r '.next_action.owner' "$tmp")" = "li-docs" ]
jq -e '.next_action.target.supporting_owners
  | index("li-frontend-developer")
  and index("li-cpo")' "$tmp" >/dev/null

echo "logicigniter work queue scanner verification passed"
