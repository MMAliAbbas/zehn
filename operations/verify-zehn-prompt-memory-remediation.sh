#!/usr/bin/env bash
set -euo pipefail

ROOT="${ZEHN_ROOT:-/Users/aliai/zehn}"
cd "$ROOT"

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

require_grep() {
  local pattern="$1"
  local path="$2"
  local label="$3"
  if grep -Eq "$pattern" "$path"; then
    pass "$label"
  else
    fail "$label"
  fi
}

reject_live_phrase() {
  local pattern="$1"
  local label="$2"
  local matches

  matches="$(
    {
      printf '%s\n' .picoclaw/workspace/AGENT.md
      printf '%s\n' .picoclaw/workspace/HEARTBEAT.md
      printf '%s\n' .picoclaw/workspace/memory/MEMORY.md
      printf '%s\n' .picoclaw/workspace/memory/ZEHN_CURRENT_STATE.md
      find .picoclaw/workspace/memory -maxdepth 1 -type f -name 'LOGICIGNITER_*.md' -print
      find .picoclaw/workspace/operating-prompts -maxdepth 1 -type f -name '*.md' -print
      find .picoclaw -maxdepth 1 -type d -name 'workspace-li-*' -exec test -f '{}/AGENT.md' ';' -print
      printf '%s\n' .picoclaw/workspace-personal/AGENT.md
    } | while IFS= read -r candidate; do
      [[ -f "$candidate" ]] || continue
      grep -HEn "$pattern" "$candidate" || true
    done
  )"

  if [[ -n "$matches" ]]; then
    printf '%s\n' "$matches" >&2
    fail "$label"
  else
    pass "$label"
  fi
}

require_file ".picoclaw/workspace/memory/ZEHN_CURRENT_STATE.md"
require_file ".picoclaw/workspace/cron/jobs.json"
require_file ".picoclaw/workspace/operating-prompts/logicigniter-specialist-work-check.md"
require_file ".picoclaw/workspace/operating-prompts/logicigniter-specialist-worker-check.md"
require_file ".picoclaw/workspace/operating-prompts/logicigniter-post-merge-reconcile.md"

if jq empty .picoclaw/workspace/cron/jobs.json; then
  pass "cron JSON is valid"
else
  fail "cron JSON is invalid"
fi

for job in \
  logicigniter-architect-work-queue \
  logicigniter-backend-work-queue \
  logicigniter-frontend-work-queue \
  logicigniter-ux-work-queue \
  logicigniter-integration-work-queue \
  logicigniter-data-ai-work-queue \
  logicigniter-devops-work-queue \
  logicigniter-qa-work-queue \
  logicigniter-security-work-queue \
  logicigniter-docs-work-queue \
  zehn-operations-monitor; do
  count="$(jq --arg job "$job" '[.jobs[] | select(.name == $job)] | length' .picoclaw/workspace/cron/jobs.json)"
  if [[ "$count" == "1" ]]; then
    pass "cron job present exactly once: $job"
  else
    fail "cron job count for $job is $count, want 1"
  fi
done

if jq -e '[.jobs[] | select((.name | test("^logicigniter-(architect|backend|frontend|ux|integration|data-ai|devops|qa|security|docs)-work-queue$")) and (.payload.message | contains("matching open PRs")))] | length == 10' .picoclaw/workspace/cron/jobs.json >/dev/null; then
  pass "all specialist cron payloads include active PR inspection"
else
  fail "one or more specialist cron payloads lacks active PR inspection"
fi

if find .picoclaw -maxdepth 1 -type d -name 'workspace-li-app-*' | grep -q .; then
  find .picoclaw -maxdepth 1 -type d -name 'workspace-li-app-*' >&2
  fail "old li-app workspace directories must not exist"
else
  pass "no old li-app workspace directories"
fi

require_grep 'matching open PRs' .picoclaw/workspace/operating-prompts/logicigniter-specialist-work-check.md \
  "scheduler prompt requires active PR inspection"
require_grep 'matching open PRs' .picoclaw/workspace/operating-prompts/logicigniter-specialist-worker-check.md \
  "worker prompt requires active PR inspection"
require_grep 'trusted-but-not-live-proven' .picoclaw/workspace/operating-prompts/logicigniter-post-merge-reconcile.md \
  "post-merge script proof status is explicit"
require_grep 'ZEHN_CURRENT_STATE.md' .picoclaw/workspace/AGENT.md \
  "root agent references current-state authority"
require_grep 'After Ali approves a project lane' .picoclaw/workspace/memory/LOGICIGNITER_SOFTWARE_DELIVERY_SYSTEM.md \
  "project lane approval is separated from approved-repo execution"
require_grep 'product strategy overlay' .picoclaw/workspace/memory/LOGICIGNITER_SOLUTION_PORTFOLIO_PLAN.md \
  "suite taxonomy conflict is clarified"
require_grep 'product strategy overlay' .picoclaw/workspace/memory/LOGICIGNITER_PORTFOLIO_REGISTRY_V1.md \
  "portfolio registry names are provisional overlay"
require_grep 'documented repo-specific' .picoclaw/workspace/memory/LOGICIGNITER_GITHUB_CONTROL_PLANE.md \
  "control plane has verify-pr fallback"
require_grep 'documented repo-specific' .picoclaw/workspace/operating-prompts/logicigniter-ceo-operating-check.md \
  "CEO prompt has verify-pr fallback"
require_grep 'documented repo-specific' .picoclaw/workspace/operating-prompts/logicigniter-engineering-check.md \
  "engineering prompt has verify-pr fallback"

for agent in \
  .picoclaw/workspace-li-architect/AGENT.md \
  .picoclaw/workspace-li-backend-developer/AGENT.md \
  .picoclaw/workspace-li-data-ai-engineer/AGENT.md \
  .picoclaw/workspace-li-devops/AGENT.md \
  .picoclaw/workspace-li-docs/AGENT.md \
  .picoclaw/workspace-li-frontend-developer/AGENT.md \
  .picoclaw/workspace-li-integration-engineer/AGENT.md \
  .picoclaw/workspace-li-qa/AGENT.md \
  .picoclaw/workspace-li-security/AGENT.md \
  .picoclaw/workspace-li-ux-designer/AGENT.md; do
  require_file "$agent"
  require_grep 'open PRs|matching PRs' "$agent" "specialist watches active PRs: $agent"
done

reject_live_phrase 'Exec: enabled, remote use blocked by allow_remote: false' \
  "live instructions do not claim old exec allow_remote state"
reject_live_phrase 'Exec is enabled, but remote exec is blocked by allow_remote: false' \
  "live instructions do not repeat old exec allow_remote state"
reject_live_phrase '87 agents|87-agent' \
  "live instructions do not rely on the old 87-agent model"
reject_live_phrase 'open draft PRs' \
  "live instructions do not ask Zehn to open draft PRs"
reject_live_phrase 'App owners remain responsible' \
  "live instructions do not reference removed persistent app-owner agents"
reject_live_phrase 'configured with "mention-only group behavior"|Discord.*is mention-only|global Discord.*mention-only at the global' \
  "live instructions do not claim Discord is mention-only"

if [[ "$failures" -ne 0 ]]; then
  printf 'Summary: %d failure(s)\n' "$failures" >&2
  exit 1
fi

printf 'Summary: all remediation checks passed\n'
