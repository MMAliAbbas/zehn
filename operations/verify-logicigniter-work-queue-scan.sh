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

jq '.prs=[]' "$FIXTURE" >"$tmp.no-pr"
"$SCANNER" --fixture "$tmp.no-pr" --today 2026-05-25 >"$tmp"

[ "$(jq -r '.next_action.type' "$tmp")" = "CLAIM_READY" ]
[ "$(jq -r '.next_action.owner' "$tmp")" = "li-backend-developer" ]

jq '.issues |= map(select(.number != 10)) | .prs=[]' "$FIXTURE" >"$tmp.no-pr"
"$SCANNER" --fixture "$tmp.no-pr" --today 2026-05-25 >"$tmp"

[ "$(jq -r '.next_action.type' "$tmp")" = "APPROVAL_REQUEST" ]
[ "$(jq -r '.next_action.owner' "$tmp")" = "li-ceo" ]

jq '.issues |= map(select(.number != 10 and .number != 20)) | .prs=[]' "$FIXTURE" >"$tmp.no-pr"
"$SCANNER" --fixture "$tmp.no-pr" --today 2026-05-25 >"$tmp"

[ "$(jq -r '.next_action.type' "$tmp")" = "UNBLOCK_DISPATCHED" ]
[ "$(jq -r '.next_action.owner' "$tmp")" = "li-devops" ]

echo "logicigniter work queue scanner verification passed"
