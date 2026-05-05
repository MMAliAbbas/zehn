#!/usr/bin/env bash
# Host-side loop for Zehn delegation/meeting feature tasks.

set -uo pipefail

ROOT="${ZEHN_ROOT:-/Users/aliai/zehn}"
RUNNER="${ZEHN_FEATURE_ONE_TASK_RUNNER:-$ROOT/operations/run-one-zehn-feature-task.sh}"
RUN_DIR="${ZEHN_FEATURE_RUN_DIR:-/tmp/zehn-feature-loop}"
LOG_DIR="${ZEHN_FEATURE_LOG_DIR:-/tmp/zehn-feature-loop}"
LOCK_FILE="$RUN_DIR/zehn-feature-loop.lock"
INTERVAL_SECONDS="${INTERVAL_SECONDS:-300}"
MAX_RUNS=""
CONTINUE_ON_FAILURE="${CONTINUE_ON_FAILURE:-1}"
PASS_ARGS=()
LOOP_LOG=""

usage() {
  cat <<'USAGE'
usage: operations/run-zehn-feature-loop.sh [options] [-- runner-options]

Options:
  --interval-seconds <n>  Wait this many seconds between completed tasks. Default: 300.
  --max-runs <n>          Stop after n runner executions.
  --once                  Run exactly one task.
  --continue-on-failure   Keep moving after failed tasks. Default.
  --stop-on-failure       Stop after the first failed task.
  -h, --help              Show this help.

Any arguments after -- are passed to run-one-zehn-feature-task.sh.
USAGE
}

while [ $# -gt 0 ]; do
  case "$1" in
    --interval-seconds)
      INTERVAL_SECONDS="${2:-}"
      shift 2
      ;;
    --max-runs)
      MAX_RUNS="${2:-}"
      shift 2
      ;;
    --once)
      MAX_RUNS="1"
      shift
      ;;
    --continue-on-failure)
      CONTINUE_ON_FAILURE=1
      shift
      ;;
    --stop-on-failure)
      CONTINUE_ON_FAILURE=0
      shift
      ;;
    --)
      shift
      PASS_ARGS=("$@")
      break
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
  line="$(printf '[%s] %s\n' "$(date '+%H:%M:%S')" "$*")"
  printf '%s\n' "$line"
  if [ -n "$LOOP_LOG" ]; then
    printf '%s\n' "$line" >> "$LOOP_LOG"
  fi
}

die() {
  log "ERROR: $*"
  exit 1
}

is_positive_integer() {
  case "$1" in
    ''|*[!0-9]*) return 1 ;;
    *) [ "$1" -gt 0 ] ;;
  esac
}

acquire_lock() {
  mkdir -p "$RUN_DIR"
  if [ -f "$LOCK_FILE" ]; then
    lock_pid="$(cat "$LOCK_FILE" 2>/dev/null || true)"
    if [ -n "$lock_pid" ] && kill -0 "$lock_pid" 2>/dev/null; then
      die "another Zehn feature loop is active: pid $lock_pid"
    fi
    rm -f "$LOCK_FILE"
  fi
  echo $$ > "$LOCK_FILE"
}

cleanup() {
  rm -f "$LOCK_FILE"
}

[ -x "$RUNNER" ] || die "missing executable runner: $RUNNER"
is_positive_integer "$INTERVAL_SECONDS" || die "--interval-seconds must be a positive integer"
if [ -n "$MAX_RUNS" ]; then
  is_positive_integer "$MAX_RUNS" || die "--max-runs must be a positive integer"
fi

mkdir -p "$RUN_DIR" "$LOG_DIR" || die "cannot create loop dirs: $RUN_DIR $LOG_DIR"
LOOP_LOG="$LOG_DIR/zehn-feature-loop-$(date +%Y%m%d%H%M%S).log"
touch "$LOOP_LOG" || die "cannot write loop log: $LOOP_LOG"

acquire_lock
trap cleanup EXIT INT TERM

log "loop log: $LOOP_LOG"
log "runner: $RUNNER"
log "interval seconds: $INTERVAL_SECONDS"
log "continue on failure: $CONTINUE_ON_FAILURE"

run_count=0
while true; do
  if [ -n "$MAX_RUNS" ] && [ "$run_count" -ge "$MAX_RUNS" ]; then
    log "max runs reached: $MAX_RUNS"
    exit 0
  fi

  run_count=$((run_count + 1))
  iteration_log="$LOG_DIR/zehn-feature-iteration-$run_count-$(date +%Y%m%d%H%M%S).log"
  iteration_runner="$RUN_DIR/run-one-zehn-feature-task-$run_count-$(date +%Y%m%d%H%M%S).sh"
  log "starting Zehn feature runner iteration $run_count"
  cp "$RUNNER" "$iteration_runner" || die "cannot snapshot runner: $RUNNER"
  chmod +x "$iteration_runner" || die "cannot make runner snapshot executable: $iteration_runner"

  if [ "${#PASS_ARGS[@]}" -gt 0 ]; then
    "$iteration_runner" "${PASS_ARGS[@]}" > "$iteration_log" 2>&1
  else
    "$iteration_runner" > "$iteration_log" 2>&1
  fi
  status=$?
  sed 's/^/[runner] /' "$iteration_log" | tee -a "$LOOP_LOG"

  if [ "$status" -ne 0 ]; then
    if grep -q "no eligible Zehn feature task found" "$iteration_log"; then
      log "no eligible Zehn feature task found; loop complete"
      exit 0
    fi
    log "runner failed with exit code $status"
    log "iteration log: $iteration_log"
    if [ "$CONTINUE_ON_FAILURE" -ne 1 ]; then
      log "stopping because --stop-on-failure was requested"
      exit "$status"
    fi
    log "continuing after failed task; sleeping $INTERVAL_SECONDS seconds"
    sleep "$INTERVAL_SECONDS"
    continue
  fi

  log "runner iteration $run_count completed"
  log "iteration log: $iteration_log"
  if [ -n "$MAX_RUNS" ] && [ "$run_count" -ge "$MAX_RUNS" ]; then
    log "max runs reached: $MAX_RUNS"
    exit 0
  fi
  log "sleeping $INTERVAL_SECONDS seconds before next task"
  sleep "$INTERVAL_SECONDS"
done
