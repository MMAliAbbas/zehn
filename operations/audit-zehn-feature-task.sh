#!/usr/bin/env bash
# Static guard for Zehn feature task files.

set -uo pipefail

ROOT="${ZEHN_ROOT:-/Users/aliai/zehn}"
TASK="${1:-}"
SCRIPT_PATH="$(cd "$(dirname "$0")" && pwd)/$(basename "$0")"

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

warn() {
  printf 'WARN: %s\n' "$*"
}

contains() {
  pattern="$1"
  grep -E "$pattern" "$TASK_FILE" >/dev/null 2>&1
}

allowed_paths_contain() {
  pattern="$1"
  awk -v pattern="$pattern" '
    $0 == "## Allowed repos/files" { in_scope=1; next }
    in_scope && /^## / { in_scope=0 }
    in_scope && $0 ~ pattern { found=1 }
    END { exit found ? 0 : 1 }
  ' "$TASK_FILE"
}

print_indented() {
  sed 's/^/  - /'
}

filter_publishability_paths() {
  grep -E '^(workspace/skills/|\.picoclaw/workspace/skills/|supervision/zehn_feature_tasks/|supervision/ZEHN_FEATURE_|supervision/ZEHN_UPSTREAM_PUBLISHING_CHECKLIST\.md)' || true
}

filter_private_skill_paths() {
  grep -E '^(workspace/skills/|\.picoclaw/workspace/skills/)' || true
}

filter_supervision_artifact_paths() {
  grep -E '^(supervision/zehn_feature_tasks/|supervision/ZEHN_FEATURE_|supervision/ZEHN_UPSTREAM_PUBLISHING_CHECKLIST\.md)' || true
}

audit_publishability() {
  history_limit="${ZEHN_PUBLISHABILITY_HISTORY_LIMIT:-50}"

  printf 'Publishability advisory audit\n'
  printf 'History window: last %s commit(s) reachable from HEAD\n' "$history_limit"

  current_skill_paths="$(git -C "$ROOT" ls-files | filter_private_skill_paths)"
  if [ -n "$current_skill_paths" ]; then
    warn "tracked private/local skill paths in current tree"
    printf '%s\n' "$current_skill_paths" | print_indented
  else
    pass "no tracked private/local skill paths in current tree"
  fi

  history_skill_paths="$(git -C "$ROOT" log --name-only --format= -n "$history_limit" -- 2>/dev/null | filter_private_skill_paths | sort -u)"
  if [ -n "$history_skill_paths" ]; then
    warn "private/local skill paths in recent history"
    printf '%s\n' "$history_skill_paths" | print_indented
  else
    pass "no private/local skill paths found in recent history"
  fi

  current_supervision_paths="$(git -C "$ROOT" ls-files | filter_supervision_artifact_paths)"
  if [ -n "$current_supervision_paths" ]; then
    warn "private supervision automation artifacts in current tree"
    printf '%s\n' "$current_supervision_paths" | print_indented
  else
    pass "no private supervision automation artifacts in current tree"
  fi

  history_supervision_paths="$(git -C "$ROOT" log --name-only --format= -n "$history_limit" -- 2>/dev/null | filter_supervision_artifact_paths | sort -u)"
  if [ -n "$history_supervision_paths" ]; then
    warn "private supervision automation artifacts in recent history"
    printf '%s\n' "$history_supervision_paths" | print_indented
  else
    pass "no private supervision automation artifacts found in recent history"
  fi

  history_publishability_paths="$(git -C "$ROOT" log --name-only --format= -n "$history_limit" -- 2>/dev/null | filter_publishability_paths | sort -u)"
  if [ -n "$history_publishability_paths" ]; then
    warn "branch history contains private automation paths; rebuild or split upstream branches only with explicit operator approval"
  else
    pass "recent branch history has no known private automation paths"
  fi

  printf 'Publishability advisory summary: warnings are non-blocking in this task audit; clean them before upstream push or PR.\n'
}

write_minimal_task() {
  task_file="$1"
  slug="$2"
  allowed="$3"
  review="${4:-}"
  cat > "$task_file" <<EOF
# Task: $slug

Slug: \`$slug\`

Docs-only allowed: yes

## Goal

Exercise automation hygiene.

## Allowed repos/files

- \`$allowed\`

## Required reading

- \`operations/run-one-zehn-feature-task.sh\`

## Work

- Exercise audit behavior.

## Acceptance criteria

- Audit behavior is deterministic.

## Verification commands

\`\`\`bash
bash -n operations/run-one-zehn-feature-task.sh
\`\`\`
$review
EOF
}

run_runner_scope_self_test() {
  tmp="$(mktemp -d "${TMPDIR:-/tmp}/zehn-runner-scope.XXXXXX")" || return 1
  trap 'rm -rf "$tmp"' EXIT

  mkdir -p "$tmp/operations" "$tmp/supervision/zehn_feature_tasks" "$tmp/workspace/skills/local"
  cp "$ROOT/operations/run-one-zehn-feature-task.sh" "$tmp/operations/run-one-zehn-feature-task.sh"
  cp "$ROOT/operations/audit-zehn-feature-task.sh" "$tmp/operations/audit-zehn-feature-task.sh"

  git -C "$tmp" init -q || return 1
  git -C "$tmp" config user.email "zehn-runner-scope@example.invalid" || return 1
  git -C "$tmp" config user.name "Zehn Runner Scope Test" || return 1

  write_minimal_task "$tmp/supervision/zehn_feature_tasks/task_runner-scope.md" "runner-scope" "allowed.txt"
  printf 'base\n' > "$tmp/allowed.txt"
  printf 'base\n' > "$tmp/unscoped.txt"
  git -C "$tmp" add . || return 1
  git -C "$tmp" commit -q -m "seed runner scope self-test" || return 1

  ZEHN_ROOT="$tmp" \
  ZEHN_FEATURE_TASK_DIR="$tmp/supervision/zehn_feature_tasks" \
  ZEHN_FEATURE_PROMPT_TEMPLATE="$tmp/prompt.md" \
  ZEHN_FEATURE_STATUS_DOC="$tmp/status.md" \
  ZEHN_FEATURE_FAILURE_DOC="$tmp/failures.md" \
  ZEHN_FEATURE_QUALITY_AUDIT="$tmp/operations/audit-zehn-feature-task.sh" \
  ZEHN_FEATURE_RUN_DIR="$tmp/run" \
  ZEHN_FEATURE_LOG_DIR="$tmp/log" \
    . "$tmp/operations/run-one-zehn-feature-task.sh"

  TASK="runner-scope"
  AUTO_COMMIT=1
  printf 'allowed\n' > "$tmp/allowed.txt"
  printf 'dirty outside scope\n' > "$tmp/unscoped.txt"
  commit_changes >/dev/null || return 1
  if git -C "$tmp" show --name-only --format= HEAD | grep -qx 'unscoped.txt'; then
    printf 'FAIL: auto-commit included unscoped dirty file\n'
    return 1
  fi
  if ! git -C "$tmp" diff --name-only HEAD | grep -qx 'unscoped.txt'; then
    printf 'FAIL: unscoped dirty file did not remain outside the commit\n'
    return 1
  fi

  git -C "$tmp" reset --hard -q HEAD || return 1
  mkdir -p "$tmp/workspace/skills/local"
  printf 'private skill\n' > "$tmp/workspace/skills/local/SKILL.md"
  if assert_scoped_changes >/dev/null 2>&1; then
    printf 'FAIL: scope guard accepted unallowed workspace/skills change\n'
    return 1
  fi

  write_minimal_task "$tmp/supervision/zehn_feature_tasks/task_local-skill-no-review.md" "local-skill-no-review" "workspace/skills/**"
  if ZEHN_ROOT="$tmp" "$SCRIPT_PATH" local-skill-no-review >/dev/null 2>&1; then
    printf 'FAIL: audit accepted workspace/skills scope without reviewer acceptance\n'
    return 1
  fi

  write_minimal_task "$tmp/supervision/zehn_feature_tasks/task_local-skill-reviewed.md" "local-skill-reviewed" "workspace/skills/**" "Reviewer accepted local-skill scope: yes"
  if ! ZEHN_ROOT="$tmp" "$SCRIPT_PATH" local-skill-reviewed >/dev/null 2>&1; then
    printf 'FAIL: audit rejected reviewed workspace/skills scope\n'
    return 1
  fi

  printf 'PASS: runner scope self-test\n'
}

run_publishability_self_test() {
  tmp="$(mktemp -d "${TMPDIR:-/tmp}/zehn-publishability.XXXXXX")" || return 1
  trap 'rm -rf "$tmp"' EXIT

  mkdir -p "$tmp/workspace/skills/private" "$tmp/supervision/zehn_feature_tasks"
  git -C "$tmp" init -q || return 1
  git -C "$tmp" config user.email "zehn-publishability@example.invalid" || return 1
  git -C "$tmp" config user.name "Zehn Publishability Test" || return 1

  printf 'private skill\n' > "$tmp/workspace/skills/private/SKILL.md"
  printf 'private task\n' > "$tmp/supervision/zehn_feature_tasks/task_private.md"
  git -C "$tmp" add . || return 1
  git -C "$tmp" commit -q -m "seed private artifacts" || return 1
  git -C "$tmp" rm -q "$tmp/workspace/skills/private/SKILL.md" || return 1
  git -C "$tmp" commit -q -m "remove current private skill" || return 1
  mkdir -p "$tmp/workspace/skills/local"
  printf 'current private skill\n' > "$tmp/workspace/skills/local/SKILL.md"
  git -C "$tmp" add "$tmp/workspace/skills/local/SKILL.md" || return 1
  git -C "$tmp" commit -q -m "add current private skill" || return 1

  output="$(ROOT="$tmp" audit_publishability 2>&1)" || return 1
  printf '%s\n' "$output" | grep -q 'WARN: tracked private/local skill paths in current tree' || {
    printf 'FAIL: publishability audit did not report current private skill paths\n'
    return 1
  }
  printf '%s\n' "$output" | grep -q 'WARN: private/local skill paths in recent history' || {
    printf 'FAIL: publishability audit did not report historical private skill paths\n'
    return 1
  }
  printf '%s\n' "$output" | grep -q 'WARN: private supervision automation artifacts in current tree' || {
    printf 'FAIL: publishability audit did not report supervision artifacts\n'
    return 1
  }

  printf 'PASS: publishability self-test\n'
}

if [ "$TASK" = "--runner-scope-self-test" ]; then
  run_runner_scope_self_test
  exit $?
fi

if [ "$TASK" = "--publishability-self-test" ]; then
  run_publishability_self_test
  exit $?
fi

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

if allowed_paths_contain '^[[:space:]]*- `workspace/skills(/|`)'; then
  if contains '^Reviewer accepted local-skill scope: yes$'; then
    pass "workspace/skills scope has reviewer acceptance"
  else
    fail "workspace/skills scope requires reviewer acceptance"
  fi
else
  pass "no workspace/skills scope"
fi

if grep -E "logicigniter|LogicIgniter|/Users/ali/projects|/Users/aliai/logicigniter|MCP v1|gRPC|li_" "$TASK_FILE" >/dev/null 2>&1; then
  fail "task contains LogicIgniter/final-readiness leakage"
else
  pass "no LogicIgniter leakage"
fi

if [ "$TASK" = "021-upstream-publishability-audit" ]; then
  audit_publishability
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
