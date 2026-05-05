#!/usr/bin/env bash
# Static guard for Zehn feature task files.

set -uo pipefail

ROOT="${ZEHN_ROOT:-/Users/aliai/zehn}"
TASK="${1:-}"

if [ -z "$TASK" ]; then
  echo "usage: operations/audit-zehn-feature-task.sh <task-slug>" >&2
  exit 64
fi

TASK_FILE="$ROOT/supervision/zehn_feature_tasks/task_$TASK.md"

failures=0

fail() {
  failures=$((failures + 1))
  printf 'FAIL: %s\n' "$*"
}

pass() {
  printf 'PASS: %s\n' "$*"
}

contains() {
  pattern="$1"
  grep -E "$pattern" "$TASK_FILE" >/dev/null 2>&1
}

printf 'Zehn feature task audit for %s\n' "$TASK"
printf 'Task file: %s\n' "$TASK_FILE"

[ -f "$TASK_FILE" ] || {
  fail "missing task file"
  printf 'Summary: %s failure(s)\n' "$failures"
  exit 1
}

contains "^Slug: \`$TASK\`$" && pass "slug matches file name" || fail "slug does not match task file name"
contains "^Docs-only allowed: (yes|no)$" && pass "declares docs-only policy" || fail "missing docs-only policy"
contains "^## Goal$" && pass "has goal section" || fail "missing goal section"
contains "^## Allowed repos/files$" && pass "has scope section" || fail "missing allowed repos/files section"
contains "^## Required reading$" && pass "has required reading" || fail "missing required reading"
contains "^## Work$" && pass "has work section" || fail "missing work section"
contains "^## Acceptance criteria$" && pass "has acceptance criteria" || fail "missing acceptance criteria"
contains "^## Verification commands$" && pass "has verification commands" || fail "missing verification commands"
contains '^```bash$' && pass "has executable bash verification block" || fail "missing bash verification block"

if grep -E "logicigniter|LogicIgniter|/Users/ali/projects|/Users/aliai/logicigniter|MCP v1|gRPC|li_" "$TASK_FILE" >/dev/null 2>&1; then
  fail "task contains LogicIgniter/final-readiness leakage"
else
  pass "no LogicIgniter leakage"
fi

if contains '^Docs-only allowed: no$'; then
  if contains 'go test|make test|make check|make generate'; then
    pass "non-doc task has code verification"
  else
    fail "non-doc task lacks code verification command"
  fi
fi

printf 'Summary: %s failure(s)\n' "$failures"
[ "$failures" -eq 0 ]
