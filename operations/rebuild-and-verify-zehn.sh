#!/usr/bin/env bash
set -euo pipefail

ROOT="${ZEHN_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
VERIFY_ONLY="${VERIFY_ONLY:-0}"

cd "$ROOT"

printf '== source ==\n'
git rev-parse HEAD
git status --short --branch

printf '\n== targeted tests ==\n'
go test ./pkg/constants ./pkg/agent -run 'TestIsInternalChannel|TestShouldPublishToolFeedback|TestRunAgentDelegation|TestRunAgentMeeting' -count=1

if [[ "$VERIFY_ONLY" != "1" ]]; then
	printf '\n== build picoclaw ==\n'
	make build

	printf '\n== build launcher ==\n'
	make build-launcher
fi

printf '\n== build sync verification ==\n'
PICOCLAW_HOME="${PICOCLAW_HOME:-$ROOT/.picoclaw}" \
	CHECK_RUNNING=0 \
	"$ROOT/operations/verify-running-build-sync.sh"
