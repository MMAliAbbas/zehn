#!/usr/bin/env bash
set -euo pipefail

ROOT="${ZEHN_ROOT:-/Users/aliai/zehn}"
HOME_DIR="${PICOCLAW_HOME:-/Users/aliai/.picoclaw-zehn}"
OUT_ROOT="${ZEHN_BASELINE_DIR:-/tmp/zehn-runtime-baseline}"
TS="$(date -u +%Y%m%dT%H%M%SZ)"
OUT="$OUT_ROOT/$TS"

mkdir -p "$OUT"

{
  echo "timestamp_utc=$TS"
  echo "root=$ROOT"
  echo "home=$HOME_DIR"
  echo "user=$(id -un 2>/dev/null || true)"
  echo "hostname=$(hostname 2>/dev/null || true)"
  echo "pwd=$(pwd)"
} > "$OUT/meta.txt"

if command -v git >/dev/null 2>&1 && [ -d "$ROOT/.git" ]; then
  git -C "$ROOT" status --short --branch > "$OUT/git-status.txt" 2>&1 || true
  git -C "$ROOT" log -n 25 --format='%h %ad %s' --date=iso > "$OUT/git-log.txt" 2>&1 || true
  git -C "$ROOT" diff --stat > "$OUT/git-diff-stat.txt" 2>&1 || true
fi

if [ -f "$HOME_DIR/config.json" ]; then
  shasum -a 256 "$HOME_DIR/config.json" > "$OUT/config.sha256" 2>&1 || true
fi

if [ -x "$ROOT/build/picoclaw" ]; then
  shasum -a 256 "$ROOT/build/picoclaw" > "$OUT/picoclaw.sha256" 2>&1 || true
fi

if [ -x "$ROOT/build/picoclaw-launcher" ]; then
  shasum -a 256 "$ROOT/build/picoclaw-launcher" > "$OUT/picoclaw-launcher.sha256" 2>&1 || true
fi

curl -fsS http://127.0.0.1:18790/health > "$OUT/health.json" 2>&1 || true
curl -fsS http://127.0.0.1:18790/ready > "$OUT/ready.json" 2>&1 || true
lsof -nP -iTCP:18790 -sTCP:LISTEN > "$OUT/listener-18790.txt" 2>&1 || true
ps -axo pid,ppid,lstart,stat,command > "$OUT/processes.txt" 2>&1 || true

if [ -f "$HOME_DIR/logs/gateway.log" ]; then
  tail -n 400 "$HOME_DIR/logs/gateway.log" > "$OUT/gateway-tail.log" 2>&1 || true
fi

if [ -f "$HOME_DIR/logs/gateway_panic.log" ]; then
  tail -n 200 "$HOME_DIR/logs/gateway_panic.log" > "$OUT/gateway-panic-tail.log" 2>&1 || true
fi

if [ -f "$HOME_DIR/workspace/heartbeat.log" ]; then
  tail -n 250 "$HOME_DIR/workspace/heartbeat.log" > "$OUT/heartbeat-tail.log" 2>&1 || true
fi

if [ -f "$HOME_DIR/workspace/heartbeat/state.json" ]; then
  cp "$HOME_DIR/workspace/heartbeat/state.json" "$OUT/heartbeat-state.json" 2>/dev/null || true
fi

if [ -f "$HOME_DIR/workspace/cron/jobs.json" ]; then
  cp "$HOME_DIR/workspace/cron/jobs.json" "$OUT/cron-jobs.json" 2>/dev/null || true
fi

if [ -d "$HOME_DIR/workspace/delegations" ]; then
  find "$HOME_DIR/workspace/delegations" -type f -name '*.json' -mtime -2 -print \
    | sort > "$OUT/recent-delegation-files.txt" 2>&1 || true
  while IFS= read -r f; do
    [ -f "$f" ] || continue
    printf '\n== %s ==\n' "$f"
    sed -n '1,220p' "$f"
  done < "$OUT/recent-delegation-files.txt" > "$OUT/recent-delegations.jsonl" 2>&1 || true
fi

if [ -d "$HOME_DIR/workspace/sessions" ]; then
  find "$HOME_DIR/workspace/sessions" -type f -mtime -1 -print \
    | sort > "$OUT/recent-session-files.txt" 2>&1 || true
fi

{
  echo "Zehn runtime baseline collected at:"
  echo "$OUT"
} | tee "$OUT/README.txt"
