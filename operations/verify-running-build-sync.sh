#!/usr/bin/env bash
set -euo pipefail

ROOT="${ZEHN_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
PICOCLAW_BIN="${PICOCLAW_BIN:-$ROOT/build/picoclaw}"
LAUNCHER_BIN="${PICOCLAW_LAUNCHER_BIN:-$ROOT/build/picoclaw-launcher}"
PICOCLAW_HOME="${PICOCLAW_HOME:-$ROOT/.picoclaw}"
PID_FILE="${PICOCLAW_PID_FILE:-$PICOCLAW_HOME/.picoclaw.pid}"
STRICT_CLEAN="${STRICT_CLEAN:-0}"
CHECK_RUNNING="${CHECK_RUNNING:-1}"

cd "$ROOT"

source_head="$(git rev-parse HEAD)"
source_short="$(git rev-parse --short=8 HEAD)"

binary_revision() {
	local bin="$1"
	if [[ ! -x "$bin" ]]; then
		printf 'missing:%s' "$bin"
		return 0
	fi
	go version -m "$bin" | awk '$1 == "build" && $2 ~ /^vcs.revision=/ { sub(/^vcs.revision=/, "", $2); print $2; found=1 } END { if (!found) print "unknown" }'
}

binary_modified() {
	local bin="$1"
	if [[ ! -x "$bin" ]]; then
		printf 'unknown'
		return 0
	fi
	go version -m "$bin" | awk '$1 == "build" && $2 ~ /^vcs.modified=/ { sub(/^vcs.modified=/, "", $2); print $2; found=1 } END { if (!found) print "unknown" }'
}

picoclaw_revision="$(binary_revision "$PICOCLAW_BIN")"
launcher_revision="$(binary_revision "$LAUNCHER_BIN")"
picoclaw_modified="$(binary_modified "$PICOCLAW_BIN")"
launcher_modified="$(binary_modified "$LAUNCHER_BIN")"

failures=0

printf 'source_head=%s\n' "$source_head"
printf 'picoclaw_binary=%s\n' "$PICOCLAW_BIN"
printf 'picoclaw_revision=%s\n' "$picoclaw_revision"
printf 'picoclaw_modified=%s\n' "$picoclaw_modified"
printf 'launcher_binary=%s\n' "$LAUNCHER_BIN"
printf 'launcher_revision=%s\n' "$launcher_revision"
printf 'launcher_modified=%s\n' "$launcher_modified"

if [[ "$picoclaw_revision" != "$source_head" ]]; then
	printf 'FAIL: picoclaw binary revision does not match source HEAD\n' >&2
	failures=$((failures + 1))
fi

if [[ "$launcher_revision" != "$source_head" ]]; then
	printf 'FAIL: launcher binary revision does not match source HEAD\n' >&2
	failures=$((failures + 1))
fi

if [[ "$STRICT_CLEAN" == "1" ]]; then
	if [[ -n "$(git status --porcelain)" ]]; then
		printf 'FAIL: source worktree is dirty and STRICT_CLEAN=1\n' >&2
		failures=$((failures + 1))
	fi
	if [[ "$picoclaw_modified" == "true" || "$launcher_modified" == "true" ]]; then
		printf 'FAIL: binary was built from dirty source and STRICT_CLEAN=1\n' >&2
		failures=$((failures + 1))
	fi
fi

if [[ "$CHECK_RUNNING" != "1" ]]; then
	printf 'running_check=skipped\n'
elif [[ -f "$PID_FILE" ]]; then
	running_version="$(sed -n 's/.*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$PID_FILE" | head -n 1)"
	printf 'pid_file=%s\n' "$PID_FILE"
	printf 'running_version=%s\n' "${running_version:-unknown}"
	if [[ -n "${running_version:-}" && "$running_version" != *"$source_short"* ]]; then
		printf 'FAIL: running pid metadata does not include source short commit %s\n' "$source_short" >&2
		failures=$((failures + 1))
	fi
else
	printf 'pid_file=%s\n' "$PID_FILE"
	printf 'running_version=not-running-or-missing-pid-file\n'
fi

if [[ "$failures" -ne 0 ]]; then
	exit 1
fi

if [[ "$CHECK_RUNNING" == "1" ]]; then
	printf 'PASS: source, binaries, and running metadata are in sync\n'
else
	printf 'PASS: source and binaries are in sync; running metadata check skipped\n'
fi
