#!/usr/bin/env bash
set -euo pipefail

HOME_DIR="${PICOCLAW_HOME:-/Users/aliai/.picoclaw-zehn}"
if [[ ! -f "$HOME_DIR/config.json" && -f /Users/aliai/.picoclaw-zehn/config.json ]]; then
  HOME_DIR="/Users/aliai/.picoclaw-zehn"
fi
failures=0

fail() {
  printf 'FAIL: %s\n' "$1" >&2
  failures=$((failures + 1))
}

pass() {
  printf 'PASS: %s\n' "$1"
}

require_file() {
  local path="$1"
  if [[ -f "$path" ]]; then
    pass "file exists: $path"
  else
    fail "missing file: $path"
  fi
}

require_no_grep() {
  local pattern="$1"
  local path="$2"
  local label="$3"
  if [[ ! -e "$path" ]]; then
    fail "cannot inspect missing path for $label: $path"
    return
  fi
  if rg -n "$pattern" "$path" >/tmp/zehn-remediation-rg.txt 2>/dev/null; then
    cat /tmp/zehn-remediation-rg.txt >&2
    fail "$label"
  else
    pass "$label"
  fi
}

require_file "$HOME_DIR/config.json"
require_file "$HOME_DIR/workspace/HEARTBEAT.md"
require_file "$HOME_DIR/workspace/cron/jobs.json"
require_file "$HOME_DIR/workspace/operating-prompts/logicigniter-ceo-operating-check.md"
require_file "$HOME_DIR/workspace/operating-prompts/logicigniter-coo-work-selection.md"
require_file "$HOME_DIR/workspace/memory/LOGICIGNITER_ACTIVE_INITIATIVES.md"
require_file "$HOME_DIR/workspace/memory/LOGICIGNITER_OPERATING_CYCLE_LEDGER.md"
require_file "$HOME_DIR/workspace/memory/LOGICIGNITER_WORK_QUEUE_SCANNER_CONTRACT.md"

jq empty "$HOME_DIR/config.json" >/dev/null && pass "config JSON is valid"
jq empty "$HOME_DIR/workspace/cron/jobs.json" >/dev/null && pass "cron JSON is valid"

if jq -e '.heartbeat.enabled == true and (.heartbeat.interval | tonumber) == 30' "$HOME_DIR/config.json" >/dev/null; then
  pass "heartbeat is enabled at 30 minutes"
else
  fail "heartbeat is not enabled at 30 minutes"
fi

if jq -e '.agents.defaults.restrict_to_workspace == false and .agents.defaults.allow_read_outside_workspace == true' "$HOME_DIR/config.json" >/dev/null; then
  pass "agents may inspect LogicIgniter repo home"
else
  fail "agents are still workspace-restricted"
fi

if jq -e '[.jobs[] | select(.enabled and .payload.channel == "internal")] | length <= 1' "$HOME_DIR/workspace/cron/jobs.json" >/dev/null; then
  pass "internal cron retry jobs are bounded"
else
  fail "too many enabled internal cron retry jobs"
fi

require_no_grep 'logicigniter-company-snapshot|recommended_heartbeat_outcome' \
  "$HOME_DIR/workspace" \
  "runtime does not reference removed snapshot workflow"

require_no_grep '/Users/aliai/zehn/\.picoclaw' \
  "$HOME_DIR/workspace/HEARTBEAT.md" \
  "heartbeat does not reference old repo-local home"

require_no_grep '/Users/aliai/zehn/\.picoclaw' \
  "$HOME_DIR/workspace/operating-prompts" \
  "operating prompts do not reference old repo-local home"

if [[ "$failures" -ne 0 ]]; then
  printf 'Summary: %d failure(s)\n' "$failures" >&2
  exit 1
fi

printf 'Summary: all current-runtime prompt/memory remediation checks passed\n'
