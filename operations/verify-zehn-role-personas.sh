#!/usr/bin/env bash
set -u

ROOT="${ZEHN_ROOT:-/Users/aliai/zehn}"
HOME_DIR="$ROOT/.picoclaw"
failures=0

fail() {
  failures=$((failures + 1))
  printf 'FAIL: %s\n' "$1"
}

check_file() {
  path="$1"
  label="$2"
  if [ ! -f "$path" ]; then
    fail "$label missing: $path"
    return 1
  fi
  return 0
}

check_agent_workspace() {
  dir="$1"
  agent="$2"
  require_repo="$3"

  check_file "$dir/AGENT.md" "$agent AGENT.md" || return
  check_file "$dir/SOUL.md" "$agent SOUL.md" || return
  check_file "$dir/USER.md" "$agent USER.md" || return
  check_file "$dir/memory/MEMORY.md" "$agent memory/MEMORY.md" || return

  if ! grep -q 'Yaad' "$dir/memory/MEMORY.md"; then
    fail "$agent memory lacks Yaad posture"
  fi

  if ! grep -q 'Active Operating Doctrine' "$dir/memory/MEMORY.md"; then
    fail "$agent memory lacks Active Operating Doctrine"
  fi

  memory_lines="$(wc -l < "$dir/memory/MEMORY.md" | tr -d ' ')"
  if [ "$memory_lines" -gt 90 ]; then
    fail "$agent memory is too long for active boot context ($memory_lines lines)"
  fi

  if grep -Eiq 'Historical Working Notes|^## .*2026-|^- 2026-' "$dir/memory/MEMORY.md"; then
    fail "$agent memory contains append-style historical notes"
  fi

  if grep -Eiq 'groundwork|groundwork-only|CISO is advisory|51 app-owner' "$dir/AGENT.md" "$dir/SOUL.md" "$dir/USER.md" "$dir/IDENTITY.md" 2>/dev/null; then
    fail "$agent active files contain stale passive/setup language"
  fi

  if grep -Rq 'li-app-' "$dir/AGENT.md" "$dir/SOUL.md" "$dir/USER.md" "$dir/IDENTITY.md" "$dir/memory/MEMORY.md" 2>/dev/null; then
    fail "$agent still references old li-app-* ownership"
  fi

  if grep -Rql 'Ignite Messaging\|Ignite Commerce\|Ignite Compliance\|Ignite Workflow\|Ignite Contract\|Ignite People\|Ignite Property\|Ignite Media\|Ignite Security\|Ignite Revenue\|10 Ignite' \
    "$dir/AGENT.md" "$dir/SOUL.md" "$dir/USER.md" "$dir/IDENTITY.md" "$dir/memory/MEMORY.md" 2>/dev/null; then
    fail "$agent active files contain stale Ignite portfolio identity"
  fi

  if [ "$require_repo" = "yes" ]; then
    if ! grep -q 'LOGICIGNITER_COMPANY_OPERATING_CONTRACT' "$dir/AGENT.md"; then
      fail "$agent AGENT.md lacks company operating contract"
    fi
    if ! grep -Eq 'LOGICIGNITER_REPO_ACCESS_DOCTRINE|/Users/aliai/logicigniter' "$dir/AGENT.md"; then
      fail "$agent AGENT.md lacks repo access doctrine"
    fi
    if ! grep -q 'LOGICIGNITER_ENGINEERING_QUALITY_DOCTRINE' "$dir/AGENT.md"; then
      fail "$agent AGENT.md lacks engineering quality doctrine"
    fi
    if ! grep -Eq 'dirty repo|repo dirty|Never leave|never leave|LOGICIGNITER_COMPANY_OPERATING_CONTRACT' "$dir/AGENT.md" "$dir/memory/MEMORY.md"; then
      fail "$agent lacks no-dirty-repo posture"
    fi
  fi
}

if [ ! -d "$HOME_DIR" ]; then
  fail "PicoClaw home missing: $HOME_DIR"
else
  check_agent_workspace "$HOME_DIR/workspace" "zehn-main" "no"
  check_agent_workspace "$HOME_DIR/workspace-personal" "personal" "no"

  for dir in "$HOME_DIR"/workspace-li-*; do
    [ -d "$dir" ] || continue
    agent="$(basename "$dir" | sed 's/^workspace-//')"
    check_agent_workspace "$dir" "$agent" "yes"
  done

  active_current_files=(
    "$HOME_DIR/config.json"
    "$HOME_DIR/workspace/cron/jobs.json"
    "$HOME_DIR/workspace/memory/LOGICIGNITER_PORTFOLIO_REGISTRY_V1.md"
    "$HOME_DIR/workspace/memory/LOGICIGNITER_SOFTWARE_DELIVERY_SYSTEM.md"
    "$HOME_DIR/workspace/memory/LOGICIGNITER_SOLUTION_PORTFOLIO_PLAN.md"
    "$HOME_DIR/workspace/memory/ZEHN_CURRENT_STATE.md"
  )

  if grep -Rql 'Ignite Messaging\|Ignite Commerce\|Ignite Compliance\|Ignite Workflow\|Ignite Contract\|Ignite People\|Ignite Property\|Ignite Media\|Ignite Security\|Ignite Revenue\|10 Ignite' \
    "$HOME_DIR"/workspace-li-bundle-* "${active_current_files[@]}" 2>/dev/null; then
    fail "stale Ignite package identity remains in active bundle/config/current-state files"
  fi
fi

if [ -f "$HOME_DIR/config.json" ]; then
  jq empty "$HOME_DIR/config.json" >/dev/null || fail "config.json is not valid JSON"
else
  fail "config.json missing"
fi

if [ -f "$HOME_DIR/workspace/cron/jobs.json" ]; then
  jq empty "$HOME_DIR/workspace/cron/jobs.json" >/dev/null || fail "cron jobs.json is not valid JSON"
fi

if [ "$failures" -eq 0 ]; then
  printf 'PASS: Zehn role persona verification passed\n'
else
  printf 'Summary: %s failure(s)\n' "$failures"
fi

exit "$failures"
