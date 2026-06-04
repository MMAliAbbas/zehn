#!/usr/bin/env python3
"""Apply audited Zehn runtime artifact fixes with backups.

This script intentionally edits only runtime artifacts under PICOCLAW_HOME. It
does not edit Go/source files and does not restart Zehn.
"""

from __future__ import annotations

import argparse
import datetime as dt
import json
import re
import shutil
from pathlib import Path


def backup_file(path: Path, home: Path, backup_root: Path) -> None:
    if not path.exists():
        return
    target = backup_root / path.relative_to(home)
    target.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(path, target)


def prune_internal_cron(home: Path, backup_root: Path) -> tuple[int, int]:
    jobs_path = home / "workspace" / "cron" / "jobs.json"
    backup_file(jobs_path, home, backup_root)
    data = json.loads(jobs_path.read_text(encoding="utf-8"))
    jobs = data.get("jobs", [])
    kept = [
        job
        for job in jobs
        if not (
            job.get("enabled") is True
            and job.get("payload", {}).get("channel") == "internal"
            and "svc-webhookrouter-grpc" in job.get("payload", {}).get("message", "")
        )
    ]
    data["jobs"] = kept
    jobs_path.write_text(json.dumps(data, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    return len(jobs), len(kept)


def trim_role_memories(home: Path, backup_root: Path) -> int:
    count = 0
    for memory_path in sorted(home.glob("workspace*/memory/MEMORY.md")):
        text = memory_path.read_text(encoding="utf-8")
        if not memory_needs_trim(text):
            continue
        write_trimmed_memory(home, memory_path, backup_root)
        count += 1
    return count


def memory_needs_trim(text: str) -> bool:
    if len(text.splitlines()) > 90:
        return True
    return bool(re.search(r"Historical Working Notes|^## .*2026-|^- 2026-", text, re.I | re.M))


def write_trimmed_memory(home: Path, memory_path: Path, backup_root: Path) -> None:
    backup_file(memory_path, home, backup_root)
    workspace = memory_path.parents[1]
    agent_id = workspace.name.removeprefix("workspace-")
    if workspace.name == "workspace":
        agent_id = "zehn-main"
        role_line = "supervise Zehn runtime health, route company checks, and keep Personal and LogicIgniter responsibilities separated."
        scope_line = "Do not act as LogicIgniter CEO or as an implementing worker; delegate to the correct configured agent."
    elif workspace.name == "workspace-personal":
        agent_id = "personal"
        role_line = "serve as Ali's personal assistant while keeping consulting and company contexts separated unless Ali explicitly bridges them."
        scope_line = "Use personal memory only for personal-assistant context; do not create company artifacts from personal context."
    elif workspace.name.startswith("workspace-li-"):
        role_line = "execute the role defined in this workspace's AGENT.md, SOUL.md, and USER.md for LogicIgniter."
        scope_line = "Use `/Users/aliai/logicigniter` as the live company repo home when making LogicIgniter claims or doing repo work."
    else:
        role_line = "execute the role defined in this workspace's AGENT.md, SOUL.md, and USER.md."
        scope_line = "Use live source evidence before making operational claims."

    memory_path.write_text(
        f"""# {agent_id} Memory

## Active Operating Doctrine

- Role: {role_line}
- Scope: {scope_line}
- Yaad posture: Yaad is the canonical durable memory system. For LogicIgniter-wide facts, use `scope_type=organization`, `external_key=logicigniter`.
- Local memory posture: this file is boot/runtime fallback only, not a historical ledger.
- Evidence posture: prefer live GitHub, live repo state, current Yaad entries, and canonical workspace memory docs over old local summaries.
- Repo hygiene: never leave a LogicIgniter repo dirty; end clean, committed on an issue branch, or report the exact cleanup blocker.
- Operating style: if blocked, state the exact limitation, assign the right owner, and choose the next useful role-appropriate action.

## Canonical References

- This workspace's `AGENT.md`, `SOUL.md`, and `USER.md`
- `/Users/aliai/.picoclaw-zehn/workspace/memory/LOGICIGNITER_COMPANY_OPERATING_CONTRACT.md`
- `/Users/aliai/.picoclaw-zehn/workspace/memory/LOGICIGNITER_ACTIVE_INITIATIVES.md`
- `/Users/aliai/.picoclaw-zehn/workspace/memory/LOGICIGNITER_REPO_ACCESS_DOCTRINE.md`
- `/Users/aliai/.picoclaw-zehn/workspace/memory/LOGICIGNITER_ENGINEERING_QUALITY_DOCTRINE.md`
- `/Users/aliai/.picoclaw-zehn/workspace/memory/LOGICIGNITER_YAAD_SCHEMA_CONTRACT.md`

## Historical Records

Historical records were archived by timestamped runtime backup before this file was condensed. Do not reconstruct current operating state from old local memory dumps when Yaad and live GitHub state are available.
""",
        encoding="utf-8",
    )


def archive_old_scoreboards(home: Path, backup_root: Path) -> int:
    scoreboard_dir = home / "workspace" / "memory" / "scoreboard"
    archive_dir = home / "workspace" / "memory" / "archive" / "scoreboard-history-20260527"
    archive_dir.mkdir(parents=True, exist_ok=True)
    count = 0
    for path in sorted(scoreboard_dir.glob("202605*.md")):
        if path.name >= "20260526.md":
            continue
        backup_file(path, home, backup_root)
        shutil.move(str(path), str(archive_dir / path.name))
        count += 1
    return count


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--home", default="/Users/aliai/.picoclaw-zehn")
    args = parser.parse_args()

    home = Path(args.home).expanduser().resolve()
    if not (home / "config.json").exists():
        raise SystemExit(f"missing Zehn config: {home / 'config.json'}")

    stamp = dt.datetime.now(dt.timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    backup_root = home / "recovery-backups" / f"{stamp}-runtime-artifact-fixes"
    backup_root.mkdir(parents=True, exist_ok=True)

    before, after = prune_internal_cron(home, backup_root)
    trimmed_memories = trim_role_memories(home, backup_root)
    archived_scoreboards = archive_old_scoreboards(home, backup_root)

    print(f"backup={backup_root}")
    print(f"cron_jobs_before={before}")
    print(f"cron_jobs_after={after}")
    print(f"role_memories_trimmed={trimmed_memories}")
    print(f"scoreboards_archived={archived_scoreboards}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
