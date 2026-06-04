#!/usr/bin/env python3
"""Build a deterministic LogicIgniter work-queue snapshot.

The scanner is intentionally label-driven. It must not infer executable work
from broad text matches because that creates false blocker counts and status
loops. Agents may use the resulting JSON as the COO control-plane input.
"""

from __future__ import annotations

import argparse
import datetime as dt
import json
import subprocess
import sys
import time
from dataclasses import dataclass
from typing import Any


ORG = "logicigniter"
GH_COMMAND_ATTEMPTS = 3
GH_COMMAND_RETRY_SECONDS = 1.5
AREA_PREFIX = "area:"
STATUS_LABELS = {
    "zehn:ready",
    "zehn:claimed",
    "zehn:in-progress",
    "zehn:blocked",
    "zehn:retrying",
    "approval:ali-required",
    "type:continuation",
}
AREA_LABELS = {
    "area:backend",
    "area:frontend",
    "area:ux",
    "area:integration",
    "area:data-ai",
    "area:architecture",
    "area:devops",
    "area:qa",
    "area:security",
    "area:docs",
    "area:product",
    "area:finance",
    "area:legal",
    "area:revenue",
    "area:marketing",
    "area:cco",
}


@dataclass(frozen=True)
class ItemRef:
    repo: str
    number: int
    title: str
    url: str
    labels: tuple[str, ...]
    updated_at: str
    kind: str
    comments: tuple[dict[str, str], ...] = ()
    comments_error: str = ""

    def as_dict(self) -> dict[str, Any]:
        primary_owner, supporting_owners = owners_for(self.labels)
        data: dict[str, Any] = {
            "repo": self.repo,
            "number": self.number,
            "title": self.title,
            "url": self.url,
            "labels": list(self.labels),
            "updated_at": self.updated_at,
            "kind": self.kind,
            "primary_owner": primary_owner,
        }
        if supporting_owners:
            data["supporting_owners"] = supporting_owners
        rework = detect_rework_path(self.comments)
        if rework:
            data["rework_path"] = rework
        if self.comments_error:
            data["source_warning"] = {
                "repo": self.repo,
                "number": self.number,
                "kind": self.kind,
                "source": "comments",
                "error": self.comments_error,
            }
        return data


def run_json(args: list[str]) -> Any:
    errors: list[str] = []
    for attempt in range(1, GH_COMMAND_ATTEMPTS + 1):
        proc = subprocess.run(args, check=False, text=True, capture_output=True)
        if proc.returncode == 0:
            break
        errors.append(f"attempt {attempt}: {proc.stderr.strip()}")
        if attempt < GH_COMMAND_ATTEMPTS:
            time.sleep(GH_COMMAND_RETRY_SECONDS)
    else:
        raise RuntimeError(
            f"command failed after {GH_COMMAND_ATTEMPTS} attempts: {' '.join(args)}\n"
            + "\n".join(errors)
        )
    if not proc.stdout.strip():
        return []
    return json.loads(proc.stdout)


def label_names(raw_labels: Any) -> tuple[str, ...]:
    names: list[str] = []
    for label in raw_labels or []:
        if isinstance(label, str):
            names.append(label)
        elif isinstance(label, dict) and label.get("name"):
            names.append(str(label["name"]))
    return tuple(sorted(set(names)))


def comment_refs(raw_comments: Any) -> tuple[dict[str, str], ...]:
    refs: list[dict[str, str]] = []
    for comment in raw_comments or []:
        if not isinstance(comment, dict):
            continue
        refs.append(
            {
                "id": str(comment.get("id") or ""),
                "url": str(comment.get("url") or ""),
                "body": str(comment.get("body") or ""),
            }
        )
    return tuple(refs)


def repo_name(raw_repo: Any) -> str:
    if isinstance(raw_repo, str):
        return raw_repo.rsplit("/", 1)[-1]
    if isinstance(raw_repo, dict):
        name = raw_repo.get("name") or raw_repo.get("nameWithOwner") or ""
        return str(name).rsplit("/", 1)[-1]
    return "unknown"


def normalize_issue(raw: dict[str, Any]) -> ItemRef:
    return ItemRef(
        repo=repo_name(raw.get("repository")),
        number=int(raw["number"]),
        title=str(raw.get("title", "")),
        url=str(raw.get("url", "")),
        labels=label_names(raw.get("labels")),
        updated_at=str(raw.get("updatedAt") or raw.get("updated_at") or ""),
        kind="issue",
        comments=comment_refs(raw.get("comments")),
        comments_error=str(raw.get("comments_error") or ""),
    )


def normalize_pr(raw: dict[str, Any]) -> ItemRef:
    return ItemRef(
        repo=repo_name(raw.get("repository")),
        number=int(raw["number"]),
        title=str(raw.get("title", "")),
        url=str(raw.get("url", "")),
        labels=label_names(raw.get("labels")),
        updated_at=str(raw.get("updatedAt") or raw.get("updated_at") or ""),
        kind="pr",
        comments=comment_refs(raw.get("comments")),
        comments_error=str(raw.get("comments_error") or ""),
    )


def fetch_issue_comments(repo: str, number: int) -> list[dict[str, Any]]:
    return run_json(
        [
            "gh",
            "api",
            f"repos/{ORG}/{repo}/issues/{number}/comments",
            "--paginate",
        ]
    )


def needs_comment_enrichment(item: dict[str, Any]) -> bool:
    labels = set(label_names(item.get("labels")))
    return bool(labels & {"approval:ali-required", "zehn:blocked"})


def enrich_comments(items: list[dict[str, Any]]) -> list[dict[str, Any]]:
    enriched: list[dict[str, Any]] = []
    for item in items:
        copy = dict(item)
        if needs_comment_enrichment(copy):
            repo = repo_name(copy.get("repository"))
            try:
                copy["comments"] = fetch_issue_comments(repo, int(copy["number"]))
            except Exception as exc:
                copy["comments_error"] = str(exc)
        enriched.append(copy)
    return enriched


def fetch_live(limit: int) -> dict[str, list[dict[str, Any]]]:
    issue_query = "state:open"
    pr_query = "state:open"
    issue_fields = "repository,title,number,labels,assignees,updatedAt,url"
    pr_fields = "repository,title,number,labels,updatedAt,url"
    return {
        "issues": enrich_comments(
            run_json(
                [
                    "gh",
                    "search",
                    "issues",
                    "--owner",
                    ORG,
                    issue_query,
                    "--json",
                    issue_fields,
                    "--limit",
                    str(limit),
                    "--sort",
                    "updated",
                    "--order",
                    "desc",
                ]
            )
        ),
        "prs": enrich_comments(
            run_json(
                [
                    "gh",
                    "search",
                    "prs",
                    "--owner",
                    ORG,
                    pr_query,
                    "--json",
                    pr_fields,
                    "--limit",
                    str(limit),
                    "--sort",
                    "updated",
                    "--order",
                    "desc",
                ]
            )
        ),
    }


def detect_rework_path(comments: tuple[dict[str, str], ...]) -> dict[str, str] | None:
    """Return a documented safe rework path from issue/PR comments if present."""
    for comment in reversed(comments):
        body = comment.get("body", "")
        text = " ".join(body.lower().split())
        if not text:
            continue
        has_rework_signal = any(
            phrase in text
            for phrase in (
                "if pr #20 is revised",
                "may merge under",
                "may merge if",
                "merge condition",
                "bounded merge condition",
                "next step: revise",
                "revise pr",
                "revised so that",
                "safe path",
                "safe rework",
            )
        )
        has_hold_signal = any(
            phrase in text
            for phrase in (
                "do not merge as-is",
                "do not merge pr",
                "not merge pr",
                "keep blocked",
                "blocked/unmerged",
            )
        )
        if has_rework_signal and has_hold_signal:
            return {
                "source_comment_id": comment.get("id", ""),
                "source_url": comment.get("url", ""),
                "summary": summarize_rework(body),
                "conditions": extract_rework_conditions(body),
            }
    return None


def summarize_rework(body: str) -> str:
    lines: list[str] = []
    for raw_line in body.splitlines():
        line = raw_line.strip()
        lower = line.lower()
        if not line:
            continue
        if any(
            phrase in lower
            for phrase in (
                "license",
                "agents.md",
                "merge condition",
                "may merge",
                "next step",
                "revise",
                "safe",
                "do not merge",
            )
        ):
            lines.append(line)
        if len(lines) >= 5:
            break
    if not lines:
        return "Approval-gated item contains a documented bounded rework path."
    return " ".join(lines)[:900]


def extract_rework_conditions(body: str) -> list[str]:
    conditions: list[str] = []
    collecting = False
    for raw_line in body.splitlines():
        line = raw_line.strip()
        lower = line.lower()
        if not line:
            if collecting:
                break
            continue
        if (
            "bounded merge condition" in lower
            or "may merge if" in lower
            or "revised so that" in lower
        ):
            collecting = True
            tail = condition_tail(line)
            if tail:
                conditions.append(tail)
            continue
        if not collecting:
            continue
        if lower.startswith(("next step", "do not merge", "source:", "evidence:")):
            break
        cleaned = clean_condition_line(line)
        if cleaned:
            conditions.append(cleaned)
    return dedupe_preserve_order(conditions)


def condition_tail(line: str) -> str:
    lower = line.lower()
    markers = ("revised so that:", "revised so that", "may merge if", "merge condition:")
    for marker in markers:
        idx = lower.find(marker)
        if idx == -1:
            continue
        tail = line[idx + len(marker) :].strip(" :-")
        if tail:
            return clean_condition_line(tail)
    return ""


def clean_condition_line(line: str) -> str:
    return line.lstrip("-*0123456789. )").strip()


def dedupe_preserve_order(values: list[str]) -> list[str]:
    seen: set[str] = set()
    result: list[str] = []
    for value in values:
        if value in seen:
            continue
        seen.add(value)
        result.append(value)
    return result


def parse_retry_date(labels: tuple[str, ...]) -> str | None:
    for label in labels:
        if label.startswith("retry:"):
            return label.split(":", 1)[1]
    return None


def retry_due(labels: tuple[str, ...], today: dt.date) -> bool:
    retry = parse_retry_date(labels)
    if not retry:
        return True
    try:
        return dt.date.fromisoformat(retry) <= today
    except ValueError:
        return True


def classify(items: list[ItemRef], prs: list[ItemRef], today: dt.date) -> dict[str, Any]:
    ready: list[dict[str, Any]] = []
    in_progress: list[dict[str, Any]] = []
    blocked: list[dict[str, Any]] = []
    approval_gated: list[dict[str, Any]] = []
    malformed: list[dict[str, Any]] = []
    continuation: list[dict[str, Any]] = []
    unblock_candidates: list[dict[str, Any]] = []

    for item in sort_refs(items):
        labels = set(item.labels)
        has_area = any(label.startswith(AREA_PREFIX) for label in labels)
        data = item.as_dict()

        if "approval:ali-required" in labels:
            approval_gated.append(data)
            if data.get("rework_path"):
                unblock_candidates.append(
                    {
                        **data,
                        "unblock_type": "approval-safe-rework",
                        "required_owner": owner_for(item.labels),
                        "reason": "approval-gated item has a documented bounded rework path",
                    }
                )
            else:
                unblock_candidates.append(
                    {
                        **data,
                        "unblock_type": "approval-question",
                        "required_owner": "li-ceo",
                        "reason": "approval:ali-required label is present",
                    }
                )
            continue

        if "zehn:blocked" in labels:
            blocked.append(data)
            if retry_due(item.labels, today):
                unblock_candidates.append(
                    {
                        **data,
                        "unblock_type": "blocked-retry-due",
                        "required_owner": owner_for(item.labels),
                        "reason": "zehn:blocked is present and retry is due or missing",
                    }
                )
            continue

        if "zehn:in-progress" in labels or "zehn:claimed" in labels:
            in_progress.append(data)
            continue

        if "type:continuation" in labels:
            continuation.append(data)

        if "zehn:ready" in labels:
            if not has_area:
                malformed.append(
                    {
                        **data,
                        "malformed_reason": "zehn:ready issue has no area:* label",
                        "repair_owner": "li-coo",
                    }
                )
            else:
                ready.append(data)

    open_prs = [pr.as_dict() for pr in sort_refs(prs)]

    return {
        "ready": ready,
        "in_progress": in_progress,
        "open_prs": open_prs,
        "blocked": blocked,
        "approval_gated": approval_gated,
        "malformed": malformed,
        "continuation": continuation,
        "unblock_candidates": unblock_candidates,
    }


def owner_for(labels: tuple[str, ...]) -> str:
    return owners_for(labels)[0]


def owners_for(labels: tuple[str, ...]) -> tuple[str, list[str]]:
    area_to_owner = {
        "area:docs": "li-docs",
        "area:security": "li-security",
        "area:legal": "li-legal",
        "area:finance": "li-cfo",
        "area:revenue": "li-cro",
        "area:devops": "li-devops",
        "area:qa": "li-qa",
        "area:architecture": "li-architect",
        "area:backend": "li-backend-developer",
        "area:integration": "li-integration-engineer",
        "area:frontend": "li-frontend-developer",
        "area:ux": "li-ux-designer",
        "area:data-ai": "li-data-ai-engineer",
        "area:product": "li-cpo",
        "area:marketing": "li-marketing",
        "area:cco": "li-cco",
    }
    priority = tuple(area_to_owner)
    present = set(label for label in labels if label.startswith(AREA_PREFIX))
    ordered_areas = [area for area in priority if area in present]
    ordered_areas.extend(sorted(present - set(ordered_areas)))
    owners = dedupe_preserve_order(
        [area_to_owner.get(area, "li-coo") for area in ordered_areas]
    )
    if not owners:
        return "li-coo", []
    return owners[0], owners[1:]


def sort_refs(refs: list[ItemRef]) -> list[ItemRef]:
    return sorted(
        refs,
        key=lambda ref: (parse_updated(ref.updated_at), ref.repo, ref.number),
        reverse=True,
    )


def parse_updated(value: str) -> dt.datetime:
    if not value:
        return dt.datetime.min.replace(tzinfo=dt.timezone.utc)
    try:
        return dt.datetime.fromisoformat(value.replace("Z", "+00:00"))
    except ValueError:
        return dt.datetime.min.replace(tzinfo=dt.timezone.utc)

def choose_next_action(snapshot: dict[str, Any]) -> dict[str, Any]:
    if snapshot.get("source_warnings"):
        return {
            "type": "SOURCE_UNAVAILABLE",
            "owner": "li-coo",
            "target": snapshot["source_warnings"][0],
            "reason": "required GitHub source data could not be loaded",
        }
    priority_pr = choose_priority_pr(snapshot["open_prs"])
    if priority_pr:
        pr = priority_pr
        labels = tuple(pr["labels"])
        if pr.get("rework_path"):
            return {
                "type": "REWORK_BLOCKER",
                "owner": owner_for(labels),
                "target": pr,
                "reason": "approval-gated PR has a documented bounded rework path",
            }
        if "approval:ali-required" in labels:
            return {
                "type": "APPROVAL_REQUEST",
                "owner": "li-ceo",
                "target": pr,
                "reason": "open PR is approval-gated and cannot move by review alone",
            }
        if "zehn:blocked" in labels:
            return {
                "type": "UNBLOCK_DISPATCHED",
                "owner": owner_for(labels),
                "target": pr,
                "reason": "open PR is blocked and requires blocker-removal work",
            }
        return {
            "type": "REVIEW_PR",
            "owner": "li-coo",
            "target": pr,
            "reason": "workflow-labeled open PR requires review, merge, or reconcile",
        }
    if snapshot["ready"]:
        item = snapshot["ready"][0]
        return {
            "type": "CLAIM_READY",
            "owner": owner_for(tuple(item["labels"])),
            "target": item,
            "reason": "claimable ready issue with canonical area label",
        }
    if snapshot["unblock_candidates"]:
        item = snapshot["unblock_candidates"][0]
        if item["unblock_type"] == "approval-question":
            action_type = "APPROVAL_REQUEST"
        elif item["unblock_type"] == "approval-safe-rework":
            action_type = "REWORK_BLOCKER"
        else:
            action_type = "UNBLOCK_DISPATCHED"
        return {
            "type": action_type,
            "owner": item["required_owner"],
            "target": item,
            "reason": item["reason"],
        }
    if snapshot["malformed"]:
        item = snapshot["malformed"][0]
        return {
            "type": "NORMALIZE_ISSUE",
            "owner": item["repair_owner"],
            "target": item,
            "reason": item["malformed_reason"],
        }
    return {
        "type": "NO_CHANGED_STATE",
        "owner": "li-coo",
        "target": None,
        "reason": "no canonical ready issue, workflow-actionable PR, unblock candidate, approval request, or malformed work found",
    }


def choose_priority_pr(open_prs: list[dict[str, Any]]) -> dict[str, Any] | None:
    """Return a PR that is explicitly workflow-actionable.

    Generic unlabeled open PRs are still reported in the snapshot, but they do
    not outrank ready issues or due unblock work. That prevents one stale PR
    from starving the broader company queue.
    """
    for pr in open_prs:
        labels = set(pr["labels"])
        if pr.get("rework_path"):
            return pr
        if labels & {"approval:ali-required", "zehn:blocked", "zehn:ready"}:
            return pr
    return None


def build_snapshot(raw: dict[str, Any], today: dt.date) -> dict[str, Any]:
    issues = [normalize_issue(item) for item in raw.get("issues", [])]
    prs = [normalize_pr(item) for item in raw.get("prs", [])]
    queues = classify(issues, prs, today)
    source_warnings = collect_source_warnings(queues)
    result = {
        "schema_version": 1,
        "generated_at": dt.datetime.now(dt.timezone.utc).isoformat(),
        "organization": ORG,
        "label_contract": sorted(STATUS_LABELS | AREA_LABELS),
        "counts": {name: len(value) for name, value in queues.items()},
        "source_warnings": source_warnings,
        **queues,
    }
    result["next_action"] = choose_next_action(result)
    return result


def collect_source_warnings(queues: dict[str, Any]) -> list[dict[str, Any]]:
    warnings: list[dict[str, Any]] = []
    queue_names = (
        "ready",
        "in_progress",
        "open_prs",
        "blocked",
        "approval_gated",
        "malformed",
        "continuation",
        "unblock_candidates",
    )
    for queue_name in queue_names:
        for item in queues.get(queue_name, []):
            warning = item.get("source_warning")
            if isinstance(warning, dict) and warning not in warnings:
                warnings.append(warning)
    return warnings


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--fixture", help="read fixture JSON instead of GitHub")
    parser.add_argument("--limit", type=int, default=500)
    parser.add_argument("--today", default=dt.date.today().isoformat())
    args = parser.parse_args()

    try:
        today = dt.date.fromisoformat(args.today)
    except ValueError:
        print(f"invalid --today date: {args.today}", file=sys.stderr)
        return 2

    try:
        if args.fixture:
            with open(args.fixture, encoding="utf-8") as handle:
                raw = json.load(handle)
        else:
            raw = fetch_live(args.limit)
        print(json.dumps(build_snapshot(raw, today), indent=2, sort_keys=True))
    except Exception as exc:
        print(json.dumps({"error": str(exc), "next_action": {"type": "SOURCE_UNAVAILABLE"}}))
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
