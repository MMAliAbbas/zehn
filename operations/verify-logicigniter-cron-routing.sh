#!/usr/bin/env bash
set -euo pipefail

root="${ZEHN_ROOT:-/Users/aliai/zehn}"
config="$root/.picoclaw/config.json"
jobs="$root/.picoclaw/workspace/cron/jobs.json"

if [[ ! -f "$config" ]]; then
  echo "missing config: $config" >&2
  exit 1
fi

if [[ ! -f "$jobs" ]]; then
  echo "missing cron jobs: $jobs" >&2
  exit 1
fi

jq empty "$config"
jq empty "$jobs"

check_route() {
  local chat="$1"
  local expected="$2"
  local name="$3"
  local actual

  actual="$(
    jq -r --arg chat "direct:$chat" '
      .agents.dispatch.rules[]
      | select(.when.channel == "discord" and .when.chat == $chat and .when.sender == "cron")
      | .agent
    ' "$config" | head -n 1
  )"

  if [[ "$actual" != "$expected" ]]; then
    printf 'FAIL route %-42s expected=%s actual=%s\n' "$name" "$expected" "${actual:-MISSING}" >&2
    return 1
  fi

  printf 'OK   route %-42s %s\n' "$name" "$actual"
}

check_job_target() {
  local job="$1"
  local expected_to="$2"
  local actual_to

  actual_to="$(
    jq -r --arg job "$job" '
      .jobs[]
      | select(.name == $job)
      | .payload.to
    ' "$jobs" | head -n 1
  )"

  if [[ "$actual_to" != "$expected_to" ]]; then
    printf 'FAIL job   %-42s expected_to=%s actual_to=%s\n' "$job" "$expected_to" "${actual_to:-MISSING}" >&2
    return 1
  fi

  printf 'OK   job   %-42s to=%s\n' "$job" "$actual_to"
}

check_route "1487893480511377580" "personal" "cron-discord-personal-main"
check_route "1487902195734024353" "li-ceo" "cron-discord-li-ceo"
check_route "1487902422310326474" "li-cto" "cron-discord-li-engineering-to-cto"
check_route "1487902530070642738" "li-operations" "cron-discord-li-operations"
check_route "1500307780613963907" "li-coo" "cron-discord-li-coo"
check_route "1487902583074062530" "li-security" "cron-discord-li-security"
check_route "1500307930321260564" "li-docs" "cron-discord-li-docs"
check_route "1500308004539469845" "li-qa" "cron-discord-li-qa"
check_route "1500308052220444852" "li-devops" "cron-discord-li-devops"
check_route "1488120554048458814" "zehn-main" "cron-discord-bot-health"

check_job_target "li-weekly-plan" "1487902195734024353"
check_job_target "li-daily-synthesis" "1500307780613963907"
check_job_target "zehn-operations-monitor" "1488120554048458814"
check_job_target "li-weekly-review" "1487902195734024353"
check_job_target "li-ceo-daily-sync" "1487902195734024353"
check_job_target "li-nonexec-weekly-pulse" "1487902195734024353"

echo "LogicIgniter cron routing verification passed."
