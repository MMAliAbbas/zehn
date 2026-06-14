# COO Daily Scoreboard Directory

## What lives here

- `YYYYMMDD.md` — one COO daily scoreboard per generation day, written by the `li-daily-synthesis-v2` cron at 08:30 +05.
- `LATEST.md` — **symlink** to the most recent truthful dated scoreboard (currently `20260604.md`).
- `SCOREBOARD_SCHEMA.md` — schema for the YYYYMMDD.md format.

## 2026-06-04 cleanup notes

- The pre-recovery `LATEST.md` was a *regular file* (not a symlink) that had been stamped with a "supersession notice" at 17:57 +05 with no newer dated scoreboard generated. It is preserved at `archive/SCOREBOARD_LATEST_20260604_pre_symlink.md` for the audit trail.
- `LATEST.md` is now a symlink to `20260604.md`. Future synthesis runs that produce a new dated file MUST `ln -sf <new>.md LATEST.md` so the two stay aligned.

## Known gaps in the dated record (2026-05-27 → 2026-06-02)

The following dates have no scoreboard file because the synthesis cron either did not fire, failed before producing a file, or its output was lost:

- 2026-05-27 (missing — between 20260526 and 20260528 in the directory listing)
- 2026-05-30
- 2026-05-31
- 2026-06-01
- 2026-06-02

These gaps will not be backfilled — the underlying inputs (live GitHub state, repo hygiene, queue snapshot) for those days are no longer reconstructable from durable memory alone. Treat them as missing-by-design rather than as "no events to report".

Cron `li-daily-synthesis-v2` is currently disabled pending Phase 6 of the recovery plan; gap will not grow until then.
