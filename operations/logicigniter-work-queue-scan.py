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
from dataclasses import dataclass
from typing import Any


ORG = "logicigniter"
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

    def as_dict(self) -> dict[str, Any]:
        return {
            "repo": self.repo,
            "number": self.number,
            "title": self.title,
            "url": self.url,
            "labels": list(self.labels),
            "updated_at": self.updated_at,
            "kind": self.kind,
        }


def run_json(args: list[str]) -> Any:
    proc = subprocess.run(args, check=False, text=True, capture_output=True)
    if proc.returncode != 0:
        raise RuntimeError(
            f"command failed ({proc.returncode}): {' '.join(args)}\n{proc.stderr.strip()}"
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
    )


def fetch_live(limit: int) -> dict[str, list[dict[str, Any]]]:
    issue_query = "state:open"
    pr_query = "state:open"
    issue_fields = "repository,title,number,labels,assignees,updatedAt,url"
    pr_fields = "repository,title,number,labels,updatedAt,url"
    return {
        "issues": run_json(
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
            ]
        ),
        "prs": run_json(
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
            ]
        ),
    }


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

    for item in items:
        labels = set(item.labels)
        has_area = any(label.startswith(AREA_PREFIX) for label in labels)
        data = item.as_dict()

        if "approval:ali-required" in labels:
            approval_gated.append(data)
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

    open_prs = [pr.as_dict() for pr in prs]

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
    areas = [label for label in labels if label.startswith(AREA_PREFIX)]
    if not areas:
        return "li-coo"
    area = areas[0]
    return {
        "area:backend": "li-backend-developer",
        "area:frontend": "li-frontend-developer",
        "area:ux": "li-ux-designer",
        "area:integration": "li-integration-engineer",
        "area:data-ai": "li-data-ai-engineer",
        "area:architecture": "li-architect",
        "area:devops": "li-devops",
        "area:qa": "li-qa",
        "area:security": "li-security",
        "area:docs": "li-docs",
        "area:product": "li-cpo",
        "area:finance": "li-cfo",
        "area:legal": "li-legal",
        "area:revenue": "li-cro",
        "area:marketing": "li-marketing",
        "area:cco": "li-cco",
    }.get(area, "li-coo")


def choose_next_action(snapshot: dict[str, Any]) -> dict[str, Any]:
    if snapshot["open_prs"]:
        pr = snapshot["open_prs"][0]
        labels = tuple(pr["labels"])
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
            "reason": "open PRs outrank new issue claims because they block merge/reconcile",
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
        action_type = (
            "APPROVAL_REQUEST"
            if item["unblock_type"] == "approval-question"
            else "UNBLOCK_DISPATCHED"
        )
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
        "reason": "no canonical ready, PR, unblock, approval, or malformed work found",
    }


def build_snapshot(raw: dict[str, Any], today: dt.date) -> dict[str, Any]:
    issues = [normalize_issue(item) for item in raw.get("issues", [])]
    prs = [normalize_pr(item) for item in raw.get("prs", [])]
    queues = classify(issues, prs, today)
    result = {
        "schema_version": 1,
        "generated_at": dt.datetime.now(dt.timezone.utc).isoformat(),
        "organization": ORG,
        "label_contract": sorted(STATUS_LABELS | AREA_LABELS),
        "counts": {name: len(value) for name, value in queues.items()},
        **queues,
    }
    result["next_action"] = choose_next_action(result)
    return result


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
