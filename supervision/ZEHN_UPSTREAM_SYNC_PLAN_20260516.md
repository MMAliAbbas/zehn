# Zehn ↔ Upstream PicoClaw Sync Plan

Date: 2026-05-16
Status: planning document, awaiting Ali approval. No code changes applied.
Trigger: PR #2 (tasks 056+057) merged with red Linter + Security CI; security failures are stdlib CVEs fixed in `go 1.25.10` which upstream already has, prompting a proper sync evaluation.

## Why this is a multi-day operation, not a one-session merge

Initial dry-run on 2026-05-16 03:00 PKT against `upstream/main` at `0df050ff`:

| Metric | Value |
| --- | --- |
| Zehn `main` ahead of `upstream/main` | 80 commits (entire zehn-features series 004-057) |
| `upstream/main` ahead of Zehn `main` | 189 commits (~10 days of community work) |
| Auto-merged files | 335 |
| Files with content conflicts | 8 |
| Architectural conflicts requiring design decisions | 2 |
| Tractable text-level conflicts | 6 |
| Last clean common ancestor | ~2026-05-06 (right before Zehn-features tasks started landing) |

The 8 conflict files split into three buckets. The two architectural conflicts are the multi-day cost; the rest are mechanical.

## Bucket 1 — Architectural (require explicit Ali decisions before any merge work)

### A1. Event-bus reconciliation

Two parallel event systems now exist in the codebases:

| Side | Package | Purpose | Origin |
| --- | --- | --- | --- |
| Zehn | `pkg/agent/eventbus.go` (`*EventBus`) | Used by delegation/meeting Yaad-traceable event logging — every "Agent event: turn_end" log line flows through this. Tasks 005-020 wired it into 8 files. | 2026-05-06 zehn-features |
| Upstream | `pkg/events/` (`runtimeevents.Bus`) — **new package, does not exist in Zehn** | Used by upstream's `feat: agent self evolution` (PR #2847) and lifecycle event publishing. | 2026-04-26 → 2026-05-12 upstream |

`pkg/agent/agent_init.go` is the conflict touchpoint — both sides extend the `AgentLoop` struct literal with different fields. The merge requires choosing one of three paths:

- **A1.a — Keep both, run side-by-side.** Add upstream's `runtimeEvents` field + initialization alongside Zehn's existing `eventBus`. The bridge code that subscribes to `runtimeEvents` is wired per upstream. Zehn's event publishers stay on `EventBus`.
  - Effort: ~4-6 hours
  - Risk: dual-firing for related events; the systems may evolve to disagree on what each emits
  - Long-term cost: every future sync has to keep them in sync manually
- **A1.b — Migrate Zehn's `EventBus` to wrap or use `runtimeevents.Bus`.** Zehn's existing pub/sub is refactored to publish through upstream's bus. Yaad logging stays intact (same data, different transport).
  - Effort: ~1-2 days of careful refactor across 8 Zehn files
  - Risk: subtle semantic changes in event ordering, observer behavior, or Yaad write timing
  - Long-term benefit: one bus, future syncs are mechanical
- **A1.c — Keep Zehn's, drop upstream's `runtimeevents` integration.** Refuse the upstream feature; resolve agent_init.go in favor of Zehn (`-Xours`). Self-evolution feature is unavailable in Zehn.
  - Effort: 1 hour
  - Risk: divergence grows; future syncs will keep hitting this conflict
  - Long-term cost: forever forked on this seam

**Recommendation: A1.b** if you want a long-lived sustainable fork. **A1.c** if you don't care about upstream's self-evolution feature. **A1.a** is a trap — pays the merge cost without solving the divergence problem.

### A2. `delegate_to_agent` tool reconciliation

Two independent implementations of the same conceptual feature now exist at the same file path.

| Side | File | Tool name | Commit | Date |
| --- | --- | --- | --- | --- |
| Zehn | `pkg/tools/delegate.go` | `delegate_to_agent` | `a25b1f52` task 004 | 2026-05-05 |
| Upstream | `pkg/tools/delegate.go` | `delegate` | PR #2531 merged commit `658961b7` (initial commit `484ef399`) | 2026-04-15 (opened), 2026-05-07 (merged) |

**Chronology matters:** upstream's PR was open 3 weeks *before* Zehn's task 004 shipped. Zehn (Codex-authored) likely used upstream's open PR as a structural template — both share package, imports, and base struct fields. Zehn's version then added model-aware extensions (`defaultModel`, `maxTokens`, `temperature`, `targetExists`, `targetModel`) that upstream's doesn't have.

This is an `add/add` conflict — both forks created files at the same path with different content. Resolution options:

- **A2.a — Adopt upstream's `delegate` as the base, migrate Zehn's extensions onto it.** Renames the in-system tool from `delegate_to_agent` → `delegate`. All Zehn role prompts and operating prompts that reference `delegate_to_agent` need updating. Zehn-specific features (`targetModel`, etc.) become optional fields/methods on upstream's struct.
  - Effort: ~1 day
  - Risk: tool-rename breaks any cached LLM behavior expecting `delegate_to_agent`; need to test li-ceo/li-coo/li-cto delegation flow
  - Benefit: future syncs are mechanical; aligned with upstream
- **A2.b — Keep Zehn's `delegate_to_agent`, refuse upstream's.** Resolve in favor of Zehn (`-Xours`). Zehn agents continue to use the richer model-aware tool.
  - Effort: 1 hour
  - Risk: cumulative divergence; upstream's delegate-tool will continue evolving, every sync hits the same wall
  - Benefit: no behavior change for current Zehn operations
- **A2.c — Manually merge: keep `delegate_to_agent` name + Zehn's structure, port upstream's improvements where helpful.** Hand-roll the file.
  - Effort: 0.5-1 day
  - Risk: hybrid file owned by neither fork; cherry-picks needed forever
  - Benefit: keeps Zehn behavior + selectively absorbs upstream changes

**Recommendation: A2.a** for long-term hygiene. **A2.b** if `delegate_to_agent` is too embedded in Zehn doctrine to rename.

## Bucket 2 — Tractable text-level conflicts (resolvable once Bucket 1 is decided)

| File | Conflict shape | Expected resolution |
| --- | --- | --- |
| `pkg/agent/agent_test.go` | Test updates flowing from runtime-events changes upstream + delegation/meeting tests added by Zehn | Mostly mergeable once A1 lands — test fixtures don't conflict, just need both sets present |
| `pkg/agent/registry.go` | Registry method additions from both sides | Both adding methods, not rewriting; ~30 min |
| `pkg/agent/registry_test.go` | Same as above for tests | ~30 min |
| `pkg/config/config.go` | Both added config struct sections | Additive merge; ~1 hour |
| `pkg/tools/delegate_test.go` | `add/add` like delegate.go | Resolved same direction as A2 |
| `web/frontend/src/i18n/locales/en.json` | Translation key additions; values diverge for keys both touched | ~30 min text merge |

## Bucket 3 — Auto-merged in the dry-run (zero work, listed for awareness)

335 files merged cleanly. Highlights worth knowing:

- New upstream packages absorbed without conflict: `pkg/events/`, `integration/`, MQTT channel docs, agent self-evolution docs, several locale additions.
- Go toolchain bump: `go.mod` directive becomes `go 1.25.10` — closes all 3 stdlib CVEs (`GO-2026-4976`, `GO-2026-4971`, `GO-2026-4918`).
- `golang.org/x/net` bumped 0.53.0 → 0.54.0.
- Multiple dep bumps (slack-go, telego, gronx, sqlite, systray, tailwindcss).

## Prerequisites before the sync starts

1. **Ali decision on A1** (event-bus). Without this, agent_init.go cannot be resolved.
2. **Ali decision on A2** (delegate-tool). Without this, delegate.go and delegate_test.go cannot be resolved, AND the rename (if A2.a) may need a follow-up doctrine sweep through `.picoclaw/workspace*/AGENT.md` files.
3. **Clean working tree on the merge host.** The pre-existing untracked supervision/operations files from prior sessions must be either committed, archived, or git-ignored before the sync to avoid confusion.
4. **Yaad token rotation** (still pending from earlier in this session) — sync work doesn't require it, but the rotation should happen before any sync-related Yaad writes from the rebuilt gateway.
5. **Linter cleanup** is recommended but not required: ~25 pre-existing files have `golines/gci/gofumpt` formatting failures, and 21 govet shadow warnings exist in Zehn-side test files. None of these are introduced by the sync itself, but they will remain red until addressed separately. Can be cleaned up before, during, or after the sync.

## Recommended order of operations once Ali approves A1 + A2

1. **Capture today's state**: tag `pre-sync-2026-05-16` on main so rollback is one command.
2. **Re-fetch upstream**: `git fetch upstream --no-tags`. Upstream may have moved since this plan was written.
3. **Create sync branch**: `sync/upstream-2026-05-NN` from main.
4. **Merge in stages** rather than one big `git merge upstream/main`:
   - Stage 1: `git merge -s ours <upstream-pre-events-commit>` to absorb dep bumps + new packages without touching the architectural files
   - Stage 2: hand-resolve A1 and A2 per the chosen direction
   - Stage 3: complete the merge with `git merge upstream/main`
5. **Run the full local CI parity check**: build, test, vet, govulncheck. Fix any newly-surfaced issues.
6. **Run formatters** (gofumpt, golines, gci) on Zehn-side files to clear the long-standing lint debt while the diff is already large.
7. **Fix the 21 govet shadow warnings** (mechanical, ~20-30 min).
8. **Open the sync PR.** Body documents A1 and A2 resolution explicitly. Required-internal-reviews include `li-architect` and `li-backend-developer`.
9. **Ali reviews and merges.**
10. **Restart gateway** to load new binary. Validate end-to-end that delegation, meeting, and heartbeat flows still work.

## Time budget (with Ali decisions in hand)

| Phase | Estimate |
| --- | --- |
| Bucket 1 architectural resolutions (both A1 + A2) | 1-2 days |
| Bucket 2 text conflicts | 0.5 day |
| Local CI parity (build + test + lint cleanup + vet + vuln) | 0.5 day |
| Buffer for unexpected regressions | 0.5 day |
| **Total** | **2.5-3.5 days** of focused work |

Without prerequisites #1 and #2, work cannot start. With them, this is a contiguous multi-day operation that should NOT be interrupted with feature work or other PRs — the merge surface is too large to babysit while doing other things.

## What this plan deliberately doesn't include

- **A "just sync security" path** — bumping `go.mod` to `1.25.10` and `x/net` to `0.54.0` independently of the broader sync. That's a 10-minute task that could be done now and would close the security CI gate without committing to the larger sync. Tracked separately as a possible task 058. Listed here for awareness but explicitly excluded from this plan's scope.
- **Lint debt cleanup as a separate operation** — the ~25 formatter failures + 21 shadow warnings could be cleaned up in a small standalone PR (no sync needed). The sync plan above absorbs this work; doing it separately first is also valid.
- **Rebasing in-flight feature work onto post-sync main** — any zehn-features tasks that land between sync-PR-merge and "now" need their own rebase plans.

## Open questions for Ali

1. **A1 decision**: A1.a, A1.b, or A1.c?
2. **A2 decision**: A2.a, A2.b, or A2.c?
3. **Should task 058 (security-only quick PR) ship first, independently of the full sync?** Would close the security CI gate within an hour and is risk-free.
4. **Lint debt — pre, during, or post sync?**
5. **Timing**: when is a 2.5-3.5 day uninterrupted window realistic? The work doesn't tolerate parallel feature changes.

## Files in scope of this plan

| Tracked here | Path |
| --- | --- |
| The plan itself | `supervision/ZEHN_UPSTREAM_SYNC_PLAN_20260516.md` |
| Eventual sync PR | branch `sync/upstream-2026-05-NN` (not yet created) |

No code changes accompany this document. Implementation waits on Ali decisions on A1 and A2.
