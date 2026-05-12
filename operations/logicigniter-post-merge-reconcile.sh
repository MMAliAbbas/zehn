#!/usr/bin/env bash
# Reconcile a merged LogicIgniter PR into the local runtime.
#
# Intended caller: Zehn li-devops, after a merging agent delegates post-merge
# service reconciliation. This script is deliberately allowlist-based: it
# refuses dirty repos, unmerged PRs, non-main merges, and unmapped restart paths.

set -euo pipefail

ROOT="${LOGICIGNITER_ROOT:-/Users/aliai/logicigniter}"

usage() {
  cat <<'USAGE'
Usage:
  operations/logicigniter-post-merge-reconcile.sh --repo <repo> --pr <number>

Examples:
  operations/logicigniter-post-merge-reconcile.sh --repo svc-services-mcp --pr 12
  operations/logicigniter-post-merge-reconcile.sh --repo svc-identity --pr 34

The PR must be merged into main in logicigniter/<repo>. The local repo must be
clean before checkout/pull. The restart path must be known in this script.
USAGE
}

die() {
  printf 'RESULT=FAIL\nREASON=%s\n' "$*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "missing required command: $1"
}

repo=""
pr=""

while [ "$#" -gt 0 ]; do
  case "$1" in
    --repo)
      [ "$#" -ge 2 ] || die "--repo requires a value"
      repo="$2"
      shift 2
      ;;
    --pr)
      [ "$#" -ge 2 ] || die "--pr requires a value"
      pr="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      die "unknown argument: $1"
      ;;
  esac
done

[ -n "$repo" ] || die "--repo is required"
[ -n "$pr" ] || die "--pr is required"

case "$repo" in
  business|operations|supervision|scripts|integration_tests|keycloak|config|infra|proto|go-packages|svc-*)
    ;;
  *)
    die "repo is not in the trusted LogicIgniter allowlist: $repo"
    ;;
esac

repo_dir="$ROOT/$repo"
[ -d "$repo_dir/.git" ] || die "local repo is missing or not a git repo: $repo_dir"

require_cmd git
require_cmd gh
require_cmd jq
require_cmd curl

pr_json="$(gh pr view "$pr" -R "logicigniter/$repo" --json number,title,state,mergedAt,baseRefName,mergeCommit,url)"

state="$(printf '%s' "$pr_json" | jq -r '.state')"
merged_at="$(printf '%s' "$pr_json" | jq -r '.mergedAt // ""')"
base_ref="$(printf '%s' "$pr_json" | jq -r '.baseRefName')"
merge_oid="$(printf '%s' "$pr_json" | jq -r '.mergeCommit.oid // ""')"
pr_url="$(printf '%s' "$pr_json" | jq -r '.url')"
pr_title="$(printf '%s' "$pr_json" | jq -r '.title')"

[ "$state" = "MERGED" ] || die "PR $pr in logicigniter/$repo is not merged; state=$state"
[ -n "$merged_at" ] || die "PR $pr in logicigniter/$repo has no mergedAt timestamp"
[ "$base_ref" = "main" ] || die "PR $pr was not merged into main; base=$base_ref"

dirty="$(git -C "$repo_dir" status --porcelain)"
[ -z "$dirty" ] || die "local repo is dirty before post-merge reconcile: $repo_dir"

before_branch="$(git -C "$repo_dir" branch --show-current || true)"
before_head="$(git -C "$repo_dir" rev-parse --short HEAD)"

git -C "$repo_dir" fetch --prune origin
git -C "$repo_dir" switch main
git -C "$repo_dir" pull --ff-only origin main

after_head="$(git -C "$repo_dir" rev-parse --short HEAD)"
after_dirty="$(git -C "$repo_dir" status --porcelain)"
[ -z "$after_dirty" ] || die "local repo became dirty after checkout/pull: $repo_dir"

run_root_script() {
  local script="$1"
  shift
  [ -x "$ROOT/$script" ] || [ -f "$ROOT/$script" ] || die "missing restart script: $ROOT/$script"
  (cd "$ROOT" && bash "$script" "$@")
}

wait_http() {
  local name="$1"
  local url="$2"
  local attempts="${3:-30}"
  local delay="${4:-2}"

  for _ in $(seq 1 "$attempts"); do
    if curl -sf -m 3 "$url" >/dev/null 2>&1; then
      printf 'HEALTH_OK name=%s url=%s\n' "$name" "$url"
      return 0
    fi
    sleep "$delay"
  done
  die "health check failed for $name at $url"
}

restart_kind=""
health_summary=""

case "$repo" in
  svc-identity)
    restart_kind="identity"
    run_root_script scripts/local-preview/start-identity.sh
    wait_http svc-identity http://localhost:8090/healthz
    health_summary="http://localhost:8090/healthz"
    ;;
  svc-billing)
    restart_kind="billing"
    run_root_script scripts/local-preview/start-billing.sh
    wait_http svc-billing http://localhost:8092/healthz
    health_summary="http://localhost:8092/healthz"
    ;;
  svc-services-bff)
    restart_kind="bff"
    run_root_script scripts/local-preview/launch-bff-for-audit.sh
    wait_http svc-services-bff http://localhost:8091/healthz
    health_summary="http://localhost:8091/healthz"
    ;;
  svc-services-mcp)
    restart_kind="mcp"
    run_root_script scripts/local-preview/start-services-mcp.sh
    wait_http svc-services-mcp http://localhost:8093/healthz
    health_summary="http://localhost:8093/healthz"
    ;;
  svc-*-grpc)
    restart_kind="grpc"
    slug="${repo#svc-}"
    slug="${slug%-grpc}"
    run_root_script scripts/local-preview/start-all-grpc.sh stop "$slug"
    run_root_script scripts/local-preview/start-all-grpc.sh all "$slug"
    run_root_script scripts/local-preview/verify-services-up.sh "$slug"
    health_summary="verify-services-up.sh $slug"
    ;;
  scripts|integration_tests|proto|go-packages|config|infra|business|operations|supervision|keycloak)
    restart_kind="no-runtime-restart"
    health_summary="no direct service restart mapped; repo synced only"
    ;;
  *)
    die "no post-merge restart mapping for repo: $repo"
    ;;
esac

final_dirty="$(git -C "$repo_dir" status --porcelain)"
[ -z "$final_dirty" ] || die "local repo is dirty after post-merge reconcile: $repo_dir"

cat <<REPORT
RESULT=OK
REPO=$repo
PR=$pr
PR_URL=$pr_url
PR_TITLE=$pr_title
MERGED_AT=$merged_at
MERGE_COMMIT=$merge_oid
PREVIOUS_BRANCH=$before_branch
HEAD_BEFORE=$before_head
HEAD_AFTER=$after_head
RESTART_KIND=$restart_kind
HEALTH=$health_summary
DIRTY_REPO=false
REPORT
