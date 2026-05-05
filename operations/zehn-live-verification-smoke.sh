#!/usr/bin/env bash
# Deterministic local smoke checks for the Zehn staged live verification plan.
#
# This script intentionally avoids live Discord, GitHub, and Yaad side effects.

set -uo pipefail

ROOT="${ZEHN_ROOT:-/Users/aliai/zehn}"
PLAN="$ROOT/supervision/ZEHN_FEATURE_LIVE_VERIFICATION.md"

failures=0

fail() {
  failures=$((failures + 1))
  printf 'FAIL: %s\n' "$*"
}

pass() {
  printf 'PASS: %s\n' "$*"
}

require_file() {
  path="$1"
  [ -f "$path" ] && pass "file exists: ${path#$ROOT/}" || fail "missing file: ${path#$ROOT/}"
}

require_pattern() {
  pattern="$1"
  description="$2"
  if grep -E -i "$pattern" "$PLAN" >/dev/null 2>&1; then
    pass "$description"
  else
    fail "$description"
  fi
}

run_check() {
  description="$1"
  shift
  if "$@"; then
    pass "$description"
  else
    fail "$description"
  fi
}

require_file "$PLAN"

if [ -f "$PLAN" ]; then
  require_pattern 'local CLI only' 'plan includes local CLI-only stage'
  require_pattern 'local gateway only' 'plan includes local gateway-only stage'
  require_pattern 'Yaad' 'plan names Yaad stage'
  require_pattern 'GitHub' 'plan names GitHub stage'
  require_pattern 'Discord' 'plan names Discord stage'
  require_pattern 'operator gate|operator gates' 'plan includes operator gates'
  require_pattern 'rollback|disable' 'plan includes rollback or disable steps'
  require_pattern 'ZEHN_LIVE_DISCORD_CONFIRM' 'plan names Discord live gate'
  require_pattern 'ZEHN_LIVE_GITHUB_CONFIRM' 'plan names GitHub live gate'
  require_pattern 'ZEHN_LIVE_YAAD_CONFIRM' 'plan names Yaad live gate'
  require_pattern 'must not send live Discord messages' 'plan keeps default Discord side effects disabled'
  require_pattern 'must not write live GitHub' 'plan keeps default GitHub side effects disabled'
  require_pattern 'must not write live Yaad' 'plan keeps default Yaad side effects disabled'
fi

run_check 'runner script has valid Bash syntax' bash -n "$ROOT/operations/run-one-zehn-feature-task.sh"
run_check 'smoke script has valid Bash syntax' bash -n "$0"

printf 'Summary: %s failure(s)\n' "$failures"
[ "$failures" -eq 0 ]
