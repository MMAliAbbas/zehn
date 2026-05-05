#!/usr/bin/env bash
# Host-side runner for one Zehn delegation/meeting feature task.

set -uo pipefail

export PATH="/opt/homebrew/bin:$HOME/go/bin:/usr/local/bin:$PATH"
export GIT_TERMINAL_PROMPT=0
export GIT_ASKPASS=/bin/false

ROOT="${ZEHN_ROOT:-/Users/aliai/zehn}"
TASK_DIR="${ZEHN_FEATURE_TASK_DIR:-$ROOT/supervision/zehn_feature_tasks}"
PROMPT_TEMPLATE="${ZEHN_FEATURE_PROMPT_TEMPLATE:-$ROOT/supervision/ZEHN_FEATURE_AUTOMATION_PROMPT.md}"
STATUS_DOC="${ZEHN_FEATURE_STATUS_DOC:-$ROOT/supervision/ZEHN_FEATURE_AUTOMATION_STATUS.md}"
FAILURE_DOC="${ZEHN_FEATURE_FAILURE_DOC:-$ROOT/supervision/ZEHN_FEATURE_AUTOMATION_FAILURES.md}"
QUALITY_AUDIT="${ZEHN_FEATURE_QUALITY_AUDIT:-$ROOT/operations/audit-zehn-feature-task.sh}"
RUN_DIR="${ZEHN_FEATURE_RUN_DIR:-/tmp/zehn-feature-loop}"
LOG_DIR="${ZEHN_FEATURE_LOG_DIR:-/tmp/zehn-feature-loop}"
LOCK_FILE="$RUN_DIR/zehn-feature-task.lock"

TASK=""
DRY_RUN=0
SKIP_CODEX=0
AUTO_COMMIT=0
RETRY_FAILED=0
CODEX_TIMEOUT="${CODEX_TIMEOUT:-7200}"
RUN_LOG=""

usage() {
  cat <<'USAGE'
usage: operations/run-one-zehn-feature-task.sh [options]

Options:
  --task <slug>   Run this task instead of selecting the next eligible task.
  --dry-run       Print selected task and rendered prompt only.
  --skip-codex    Run audits and verification without invoking Codex.
  --commit        Commit local changes after verification. Default: no commit.
  --retry-failed  Ignore the failure ledger when selecting the next task.
  -h, --help      Show this help.
USAGE
}

while [ $# -gt 0 ]; do
  case "$1" in
    --task)
      TASK="${2:-}"
      shift 2
      ;;
    --dry-run)
      DRY_RUN=1
      shift
      ;;
    --skip-codex)
      SKIP_CODEX=1
      shift
      ;;
    --commit)
      AUTO_COMMIT=1
      shift
      ;;
    --retry-failed)
      RETRY_FAILED=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage >&2
      exit 64
      ;;
  esac
done

log() {
  printf '[%s] %s\n' "$(date '+%H:%M:%S')" "$*"
}

die() {
  log "ERROR: $*"
  exit 1
}

with_timeout() {
  seconds="$1"
  shift
  if command -v timeout >/dev/null 2>&1; then
    timeout --kill-after=10s "${seconds}s" "$@"
  elif command -v gtimeout >/dev/null 2>&1; then
    gtimeout --kill-after=10s "${seconds}s" "$@"
  else
    perl -e 'alarm shift; exec @ARGV' "$seconds" "$@"
  fi
}

task_has_status() {
  task="$1"
  desired="$2"
  [ -f "$STATUS_DOC" ] || return 1
  awk -v task="$task" -v desired="$desired" '
    $0 ~ "^\\|[[:space:]]*" task "[[:space:]]*\\|[[:space:]]*" desired "[[:space:]]*\\|" { found=1 }
    END { exit found ? 0 : 1 }
  ' "$STATUS_DOC"
}

task_has_open_failure() {
  task="$1"
  [ "$RETRY_FAILED" -eq 0 ] || return 1
  [ -f "$FAILURE_DOC" ] || return 1
  awk -v task="$task" '
    $0 ~ "^\\|[[:space:]]*" task "[[:space:]]*\\|[[:space:]]*needs-review[[:space:]]*\\|" { found=1 }
    END { exit found ? 0 : 1 }
  ' "$FAILURE_DOC"
}

select_next_task() {
  for task_file in "$TASK_DIR"/task_*.md; do
    [ -f "$task_file" ] || continue
    slug="$(basename "$task_file" .md)"
    slug="${slug#task_}"
    task_has_status "$slug" "green" && continue
    task_has_open_failure "$slug" && continue
    printf '%s\n' "$slug"
    return 0
  done
  return 1
}

acquire_lock() {
  mkdir -p "$RUN_DIR"
  if [ -f "$LOCK_FILE" ]; then
    lock_pid="$(cat "$LOCK_FILE" 2>/dev/null || true)"
    if [ -n "$lock_pid" ] && kill -0 "$lock_pid" 2>/dev/null; then
      die "another Zehn feature runner is active: pid $lock_pid"
    fi
    rm -f "$LOCK_FILE"
  fi
  echo $$ > "$LOCK_FILE"
}

release_lock() {
  rm -f "$LOCK_FILE"
}

assert_repo_clean() {
  status="$(git -C "$ROOT" status --short)"
  [ -z "$status" ] || {
    log "repo has dirty tracked/managed changes before $TASK"
    printf '%s\n' "$status"
    return 1
  }
}

extract_allowed_paths() {
  TASK_FILE="$TASK_DIR/task_$TASK.md" python3 <<'PY'
from pathlib import Path
import os

task_file = Path(os.environ["TASK_FILE"])
text = task_file.read_text()
in_scope = False
for line in text.splitlines():
    if line.strip() == "## Allowed repos/files":
        in_scope = True
        continue
    if in_scope and line.startswith("## "):
        break
    if in_scope and line.lstrip().startswith("- `") and "`" in line[4:]:
        print(line.split("`", 2)[1].rstrip("/"))
PY
}

assert_scoped_changes() {
  allowed=()
  while IFS= read -r line; do
    allowed+=("$line")
  done < <(extract_allowed_paths)

  changed=()
  while IFS= read -r line; do
    changed+=("$line")
  done < <({
    git -C "$ROOT" diff --name-only
    git -C "$ROOT" diff --cached --name-only
    git -C "$ROOT" ls-files --others --exclude-standard
  } | sort -u)

  bad=()
  for rel in "${changed[@]}"; do
    [ -n "$rel" ] || continue
    ok=0
    for item in "${allowed[@]}"; do
      item="${item%/}"
      case "$item" in
        */**)
          prefix="${item%/**}"
          [[ "$rel" == "$prefix" || "$rel" == "$prefix/"* ]] && ok=1
          ;;
        *\*)
          prefix="${item%\*}"
          [[ "$rel" == "$prefix"* ]] && ok=1
          ;;
        *)
          [[ "$rel" == "$item" || "$rel" == "$item/"* ]] && ok=1
          ;;
      esac
      [ "$ok" -eq 1 ] && break
    done
    [ "$ok" -eq 1 ] || bad+=("$rel")
  done

  if [ "${#bad[@]}" -gt 0 ]; then
    printf 'Unscoped changes detected:\n'
    printf 'UNSCOPED %s\n' "${bad[@]}"
    return 1
  fi
}

assert_staged_changes_scoped() {
  allowed=()
  while IFS= read -r line; do
    allowed+=("$line")
  done < <(extract_allowed_paths)

  staged=()
  while IFS= read -r line; do
    staged+=("$line")
  done < <(git -C "$ROOT" diff --cached --name-only | sort -u)

  bad=()
  for rel in "${staged[@]}"; do
    [ -n "$rel" ] || continue
    ok=0
    for item in "${allowed[@]}"; do
      item="${item%/}"
      case "$item" in
        */**)
          prefix="${item%/**}"
          [[ "$rel" == "$prefix" || "$rel" == "$prefix/"* ]] && ok=1
          ;;
        *\*)
          prefix="${item%\*}"
          [[ "$rel" == "$prefix"* ]] && ok=1
          ;;
        *)
          [[ "$rel" == "$item" || "$rel" == "$item/"* ]] && ok=1
          ;;
      esac
      [ "$ok" -eq 1 ] && break
    done
    [ "$ok" -eq 1 ] || bad+=("$rel")
  done

  if [ "${#bad[@]}" -gt 0 ]; then
    printf 'Unscoped staged changes detected:\n'
    printf 'UNSCOPED-STAGED %s\n' "${bad[@]}"
    return 1
  fi
}

render_prompt() {
  prompt_file="$1"
  task_file="$TASK_DIR/task_$TASK.md"
  {
    cat "$PROMPT_TEMPLATE"
    printf '\n## Selected Task\n\n'
    printf 'Task file: `%s`\n\n' "$task_file"
    cat "$task_file"
  } > "$prompt_file"
}

extract_verification_script() {
  script_file="$1"
  TASK_FILE="$TASK_DIR/task_$TASK.md" SCRIPT_FILE="$script_file" python3 <<'PY'
from pathlib import Path
import os
import re
import sys

task_file = Path(os.environ["TASK_FILE"])
script_file = Path(os.environ["SCRIPT_FILE"])
text = task_file.read_text()
match = re.search(r"## Verification commands\s+```bash\n(.*?)\n```", text, re.S)
if not match:
    print(f"missing verification bash block in {task_file}", file=sys.stderr)
    sys.exit(1)
script_file.write_text("set -euo pipefail\n" + match.group(1).strip() + "\n")
PY
}

run_verification() {
  verification_script="$LOG_DIR/verify-$TASK-$(date +%Y%m%d%H%M%S).sh"
  verification_log="$LOG_DIR/verify-$TASK-$(date +%Y%m%d%H%M%S).log"
  extract_verification_script "$verification_script" || return 1
  log "running task verification commands"
  if ! env -u PICOCLAW_HOME -u PICOCLAW_CONFIG bash "$verification_script" > "$verification_log" 2>&1; then
    sed 's/^/[verify] /' "$verification_log" | tail -220
    return 1
  fi
  sed 's/^/[verify] /' "$verification_log" | tail -120
}

run_codex_task() {
  prompt_file="$LOG_DIR/prompt-$TASK-$(date +%Y%m%d%H%M%S).md"
  codex_log="$LOG_DIR/codex-$TASK-$(date +%Y%m%d%H%M%S).log"
  render_prompt "$prompt_file"
  log "running Codex host task for $TASK"
  log "codex log: $codex_log"
  if ! with_timeout "$CODEX_TIMEOUT" \
    codex exec "$(cat "$prompt_file")" \
      --dangerously-bypass-approvals-and-sandbox \
      -C "$ROOT" > "$codex_log" 2>&1; then
    log "Codex task failed or was interrupted; see $codex_log"
    return 1
  fi
}

record_failure_status() {
  exit_code="$1"
  mkdir -p "$(dirname "$FAILURE_DOC")"
  TASK="$TASK" RUN_LOG="$RUN_LOG" EXIT_CODE="$exit_code" FAILURE_DOC="$FAILURE_DOC" python3 <<'PY'
from datetime import datetime
from pathlib import Path
import os
import re

task = os.environ["TASK"]
run_log = os.environ["RUN_LOG"]
exit_code = os.environ["EXIT_CODE"]
failure_doc = Path(os.environ["FAILURE_DOC"])
rows = {}
if failure_doc.exists():
    for line in failure_doc.read_text().splitlines():
        match = re.match(r"\|\s*([^|]+?)\s*\|\s*needs-review\s*\|\s*`?([^`|]+)`?\s*\|\s*`?([^`|]+)`?\s*\|\s*([^|]+)\|", line)
        if match:
            rows[match.group(1).strip()] = (match.group(2).strip(), match.group(3).strip(), match.group(4).strip())
rows[task] = (Path(run_log).name, "not-archived", f"exit {exit_code}")
lines = [
    "# Zehn Feature Automation Failures",
    "",
    f"Updated: {datetime.now().astimezone().isoformat(timespec='seconds')}",
    "",
    "This ledger is host-runner owned. `needs-review` tasks should be skipped by an unattended loop so one red task cannot block the remaining Zehn feature work.",
    "",
    "| Task | Status | Host runner evidence | Archived diff | Reason |",
    "| --- | --- | --- | --- | --- |",
]
for key in sorted(rows):
    evidence, archive, reason = rows[key]
    lines.append(f"| {key} | needs-review | `{evidence}` | `{archive}` | {reason} |")
failure_doc.write_text("\n".join(lines) + "\n")
PY
}

record_green_status() {
  mkdir -p "$(dirname "$STATUS_DOC")"
  TASK="$TASK" EVIDENCE="$(basename "$RUN_LOG")" STATUS_DOC="$STATUS_DOC" TASK_DIR="$TASK_DIR" python3 <<'PY'
from datetime import datetime
from pathlib import Path
import os
import re

task = os.environ["TASK"]
evidence = os.environ["EVIDENCE"]
status_doc = Path(os.environ["STATUS_DOC"])
task_dir = Path(os.environ["TASK_DIR"])
rows = {}
if status_doc.exists():
    for line in status_doc.read_text().splitlines():
        match = re.match(r"\|\s*([^|]+?)\s*\|\s*green\s*\|\s*`?([^`|]+)`?\s*\|\s*([^|]*)\|", line)
        if match:
            rows[match.group(1).strip()] = (match.group(2).strip(), match.group(3).strip())
rows[task] = (evidence, "host verified")
tasks = sorted(p.stem.removeprefix("task_") for p in task_dir.glob("task_*.md"))
not_green = [item for item in tasks if item not in rows]
lines = [
    "# Zehn Feature Automation Status",
    "",
    f"Updated: {datetime.now().astimezone().isoformat(timespec='seconds')}",
    "",
    "This ledger is host-runner owned. A task is green only after its verification commands pass and related changes are reviewed according to the Zehn feature automation process.",
    "",
    "## Green Tasks",
    "",
    "| Task | Status | Host runner evidence | Notes |",
    "| --- | --- | --- | --- |",
]
for key in sorted(rows):
    evidence_value, notes = rows[key]
    lines.append(f"| {key} | green | `{evidence_value}` | {notes} |")
lines.extend(["", f"Total green: {len(rows)} / {len(tasks)}", "", "## Not Green In This Ledger", ""])
if not_green:
    lines.append("`" + "`, `".join(not_green) + "`")
else:
    lines.append("All Zehn feature tasks have host-runner green evidence.")
status_doc.write_text("\n".join(lines) + "\n")
PY
}

commit_changes() {
  [ "$AUTO_COMMIT" -eq 1 ] || {
    log "auto-commit disabled"
    return 0
  }

  allowed=()
  while IFS= read -r line; do
    allowed+=("$line")
  done < <(extract_allowed_paths)

  if [ "${#allowed[@]}" -eq 0 ]; then
    die "no allowed paths found for $TASK; refusing to auto-stage"
  fi

  for item in "${allowed[@]}"; do
    item="${item%/}"
    if git -C "$ROOT" check-ignore -q -- "$item"; then
      log "skipping ignored allowed path during auto-stage: $item"
      continue
    fi
    case "$item" in
      */**)
        prefix="${item%/**}"
        git -C "$ROOT" add -- "$prefix" || return 1
        ;;
      *\*)
        matches=()
        while IFS= read -r path; do
          matches+=("$path")
        done < <(git -C "$ROOT" ls-files --modified --deleted --others --exclude-standard -- "$item")
        if [ "${#matches[@]}" -gt 0 ]; then
          git -C "$ROOT" add -- "${matches[@]}" || return 1
        fi
        ;;
      *)
        if [ -e "$ROOT/$item" ] || git -C "$ROOT" ls-files --error-unmatch "$item" >/dev/null 2>&1; then
          git -C "$ROOT" add -- "$item" || return 1
        fi
        ;;
    esac
  done

  assert_staged_changes_scoped || return 1

  if git -C "$ROOT" diff --cached --quiet; then
    log "nothing staged"
    return 0
  fi
  git -C "$ROOT" commit -m "chore(zehn-features): complete $TASK"
}

cleanup() {
  code=$?
  if [ "$code" -ne 0 ] && [ -n "$TASK" ] && [ -n "$RUN_LOG" ]; then
    record_failure_status "$code" || true
  fi
  release_lock
  exit "$code"
}

main() {
  [ -d "$ROOT" ] || die "missing workspace: $ROOT"
  [ -f "$PROMPT_TEMPLATE" ] || die "missing prompt template: $PROMPT_TEMPLATE"
  [ -d "$TASK_DIR" ] || die "missing task dir: $TASK_DIR"

  if [ -z "$TASK" ]; then
    TASK="$(select_next_task)" || die "no eligible Zehn feature task found"
  fi

  [ -f "$TASK_DIR/task_$TASK.md" ] || die "missing task file: $TASK_DIR/task_$TASK.md"
  mkdir -p "$RUN_DIR" "$LOG_DIR"
  RUN_LOG="$LOG_DIR/runner-$TASK-$(date +%Y%m%d%H%M%S).log"

  if [ "$DRY_RUN" -eq 1 ]; then
    prompt_file="$LOG_DIR/prompt-$TASK-dry-run.md"
    render_prompt "$prompt_file"
    cat <<EOF
Dry run only.

Selected task: $TASK
Rendered prompt:
  - $prompt_file
Task file:
  - $TASK_DIR/task_$TASK.md
Quality audit:
  - $QUALITY_AUDIT $TASK
EOF
    exit 0
  fi

  touch "$RUN_LOG" || die "cannot write runner log: $RUN_LOG"
  exec > >(tee -a "$RUN_LOG") 2>&1
  log "runner log: $RUN_LOG"
  log "selected task: $TASK"

  acquire_lock
  trap cleanup EXIT INT TERM

  assert_repo_clean || die "repo is dirty before $TASK; commit or clean before running automation"
  "$QUALITY_AUDIT" "$TASK" || die "task audit failed for $TASK"

  if [ "$SKIP_CODEX" -eq 0 ]; then
    run_codex_task || die "Codex task failed for $TASK"
  else
    log "skipping Codex because --skip-codex was set"
  fi

  assert_scoped_changes || die "changes outside allowed scope for $TASK"
  "$QUALITY_AUDIT" "$TASK" || die "task audit failed after changes for $TASK"
  run_verification || die "verification failed for $TASK"
  record_green_status || die "failed to write green status for $TASK"
  commit_changes || die "commit failed for $TASK"
  log "complete for $TASK"
}

if [ "${BASH_SOURCE[0]}" = "$0" ]; then
  main "$@"
fi
