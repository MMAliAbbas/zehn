# Zehn ↔ Upstream PicoClaw — Coexistence + Daily-Sync Plan

Date: 2026-05-16
Status: research + plan document. No code changes.
Companion to: `supervision/ZEHN_UPSTREAM_SYNC_PLAN_20260516.md` (one-time-sync plan written earlier today).

Author note: written autonomously while Ali is away. Investigation notes captured first; synthesized plan at the end.

## TL;DR

**Why Zehn fell 189 commits behind:** 22 upstream-shared files have Zehn modifications, plus two cases of parallel feature development (delegate-tool, runtime events). No system flags this in real time; no policy decides what to do when overlaps happen.

**Quantified conflict surface (today):**
- `pkg/config/config.go`: 12 upstream commits since divergence + Zehn's `+301/-290` diff → highest risk
- `pkg/agent/agent_init.go`: 3 upstream commits + Zehn's `+91/-54` → already conflicting
- ~20 other files at lower but real risk

**Proposal: five-phase migration.**

1. **One-time sync** (per companion `ZEHN_UPSTREAM_SYNC_PLAN_20260516.md`): catch up the 189-commit backlog. 2.5-3.5 days. **Requires A1 (event-bus) + A2 (delegate-tool) decisions from Ali first.**
2. **File rename pass:** migrate existing Zehn-only files in shared dirs to the `zehn_*` prefix convention (used by tasks 056/057). Prevents future add/add conflicts. 1-2 days.
3. **Modification surface refactor:** collect Zehn-only fields into nested sub-structs (e.g., `Zehn ZehnConfig` inside `config.Config`); reduces upstream-shared modification surface from 22 files to ~6-8. ~13 hours across weeks.
4. **Daily auto-sync GitHub Action:** runs at 02:00 UTC daily; attempts merge of upstream/main into a temp branch; opens PR on clean merge or issue on conflict. Catches divergence within 24h, never again 189 commits. 1 day.
5. **Ongoing discipline:** every new zehn-features task includes `upstream_modifications` justification + CI assertion that new files in shared dirs use `zehn_` prefix.

**Total upfront effort:** ~5-7 days of focused work spread over 2-4 weeks. After that, syncs become routine: daily PRs on clean weeks, surfaced issues on conflict weeks.

**Companion docs:**
- `ZEHN_UPSTREAM_SYNC_PLAN_20260516.md` — one-time bring-current sync (Phase 1 prereq)
- `ZEHN_AGENT_STATE_CACHE_PLAN_20260516.md` — independent: stops agents rescanning repos every heartbeat

## Problem statement (two parts)

**Part A.** Zehn has accumulated 80 commits since the last clean sync with `sipeed/picoclaw` upstream. Many of those commits modify upstream-clean code, creating a divergence surface that grows with every Zehn-features task. Today we're 189 commits behind. Without intervention, the divergence keeps growing and every future sync becomes harder.

**Part B.** When upstream and Zehn build the same conceptual feature in parallel (delegate-tool today; event-bus also today), Zehn ends up with two implementations. There is no policy for **picking, coexisting, or merging** when this happens, and no early-warning system to surface the duplication before it lands.

We need both:
- a **discipline** for how Zehn builds new features so future divergence is bounded;
- a **system** that runs daily, attempts the sync automatically, and escalates clearly when human judgment is needed.

## Investigation notes (live)

### Observation 1 — Zehn-features task inventory

57 task spec files under `supervision/zehn_feature_tasks/`. 80 commits since divergence (~2026-05-06). Tasks fall into three theme clusters:

- **Tasks 001–023 — Delegation + meeting + GitHub artifacts substrate.** Built the core platform Zehn agents use today: delegate tool, async executor, meeting framework, delegation/meeting records, GitHub artifact interface, Yaad MCP integration.
- **Tasks 024–055 — Agent Organization UI.** The web frontend "Organization Command Center" — config-model, snapshot API, live-log panel, filters, drilldowns, failure-reason rendering, accessibility. Heavy `web/frontend/` + `web/backend/` work.
- **Tasks 056–057 — GitHub artifact writer + repo routing.** The PR we just merged today.

Common commit shape: each task ships as TWO commits — a "add task spec" commit (touches only `supervision/zehn_feature_tasks/`) plus a "complete task" commit (touches the implementation files). This makes per-task file-touch analysis easy.

### Observation 2 — Files Zehn modified in upstream-clean paths (the conflict surface)

Computed by intersecting Zehn's `git log main --not upstream/main --name-only` with the set of files that exist on `upstream/main`. Then for each shared file, counted upstream commits since divergence + Zehn's net diff:

| File | Upstream commits since divergence | Zehn net diff | Conflict risk |
|---|---|---|---|
| `pkg/config/config.go` | **12** | `+301/-290` | **HIGHEST** — both sides heavily modified config registry |
| `pkg/agent/instance.go` | 4 | `+9/-98` | Medium |
| `pkg/config/defaults.go` | 4 | `+16/-21` | Medium |
| `pkg/agent/agent_init.go` | 3 | `+91/-54` | **High** — already conflicting (event-bus issue) |
| `pkg/agent/registry.go` | 3 | `+102/-34` | High |
| `pkg/channels/manager.go` | 3 | `+20/-154` | Medium |
| `pkg/agent/agent.go` | 2 | `+28/-78` | Medium |
| `pkg/agent/agent_test.go` | 2 | `+96/-107` | Medium |
| `pkg/agent/prompt_contributors.go` | 2 | `+65/-41` | Medium |
| `pkg/config/config_test.go` | 2 | `+98/-395` | Medium |
| `pkg/mcp/manager_test.go` | 2 | `+21/-220` | Medium |
| `pkg/agent/prompt.go` | 1 | `+8/-8` | Low |
| `pkg/agent/steering_test.go` | 1 | `+16/-482` | Low |
| `pkg/agent/turn_state.go` | 1 | `+26/-192` | Low |
| `pkg/mcp/manager.go` | 1 | `+33/-54` | Low |

Plus files Zehn modified that upstream hasn't touched **yet** (zero risk for *this* sync, but the surface area for future syncs):
- `pkg/agent/delegation.go` (+405) — Zehn-only file in shared dir
- `pkg/agent/delegation_visibility.go` (+192) — Zehn-only file in shared dir
- `pkg/agent/prompt_test.go`, `pkg/agent/registry_test.go` — Zehn additions only
- `pkg/channels/discord/discord_test.go` — Zehn additions only
- `pkg/providers/oauth/codex_provider.go` + test — Zehn touched, upstream not (yet)

**Bottom line:** 13 files in `pkg/agent/` + 4 in `pkg/config/` + 2 in `pkg/channels/` + 2 in `pkg/mcp/` + 2 in `pkg/providers/oauth/` = **~22 files in upstream-clean directories that Zehn has modified**, of which 5–7 are at meaningful conflict risk on any sync.

### Observation 3 — Files Zehn has *added* (zehn-only namespace)

Far larger surface, but each file is zero-conflict because upstream has no version. These are the "safe" Zehn additions:

- `pkg/agent/eventbus.go`, `eventbus_test.go`, `events.go`, `agent_event.go` — Zehn's event bus (conflicts conceptually with upstream's new `pkg/events/`, see companion sync plan)
- `pkg/agent/delegation_store.go`, `delegation_memory.go`, `delegation_executor.go`, `delegation_meeting_e2e_test.go`, `delegation_test.go` — Zehn's delegation persistence
- `pkg/agent/meeting.go`, `meeting_store.go`, `meeting_test.go`, `meeting_v2.go`, `meeting_v2_test.go` — Zehn's meeting framework
- `pkg/agent/github_artifacts.go`, `github_artifacts_test.go` — GitHub artifact glue (tasks 009/020)
- `pkg/agent/organization.go`, `organization_test.go` — agent organization model
- `pkg/agent/zehn_github_artifact_writer.go`, `zehn_github_repo_resolver.go`, `zehn_init_hook.go` (+ tests) — today's tasks 056/057
- `pkg/tools/delegate.go`, `delegate_test.go`, `delegate_status.go`, `delegation_status_test.go` — Zehn's delegate tool (conflicts with upstream's parallel implementation)
- `pkg/tools/meeting.go`, `meeting_test.go` — Zehn's meeting tool
- `pkg/tools/integration/github_artifacts.go` — Zehn's GitHub artifact interface (modified by 056/057)
- `web/backend/api/organization.go`, `organization_test.go` — backend for the Organization UI
- 100+ `web/frontend/src/components/agent/organization/**` files — the UI itself

The `zehn_` filename prefix convention started recently (tasks 056/057). Many earlier Zehn files don't use it — names like `delegation.go`, `meeting.go`, `eventbus.go` are ambiguous with potential future upstream files of the same name.

### Observation 4 — Upstream features that arrived in parallel with Zehn-features

The two known overlapping features (full detail in the companion sync plan):

| Feature | Upstream commit / PR | Date | Zehn equivalent | Date | Outcome |
|---|---|---|---|---|---|
| `delegate` tool | `484ef399`, merged `658961b7` (PR #2531) | opened 2026-04-15, merged 2026-05-07 | `delegate_to_agent` task 004 (`a25b1f52`) | 2026-05-05 | **Two implementations, same filename, `add/add` merge conflict** |
| Runtime events / `pkg/events/` | `eedebabb`, lifecycle events `e613258f` (PR #2847 et al.) | 2026-04-26 onwards | `pkg/agent/eventbus.go` ecosystem (tasks 005–020) | 2026-05-06 | **Two parallel event systems, both alive, no consolidation** |

Probable other overlaps to check during any sync:
- Agent self-evolution: upstream PR #2847. Zehn doesn't have an equivalent; this is upstream-only.
- Channel-system refactor branch (upstream `refactor/channel-system`) — Zehn modifies `pkg/channels/manager.go`; upstream's refactor may rewrite it.
- Telegram / WeChat / Discord channel updates upstream — Zehn touched `pkg/channels/discord/discord_test.go`; benign additions likely.

### Observation 5 — Existing upstream-clean discipline in current zehn doctrine

`supervision/ZEHN_FEATURE_AUTOMATION_PROMPT.md` already states the rule:

> "Keep upstream-clean code generic. Zehn/Yaad/GitHub/Discord-specific behavior belongs in narrow adapters or private configuration paths."

And `supervision/ZEHN_UPSTREAM_PUBLISHING_CHECKLIST.md` exists (we haven't read it yet — it's untracked in the supervision/ untracked set), implying there's prior thinking about publishability back to upstream.

In practice: the discipline has been imperfectly applied. Recent tasks (042–055, the organization UI work) heavily modified upstream files because the UI is hooked into the agent runtime. Tasks 056/057 are a positive example — they used `zehn_*` prefixed filenames and an additive 3-line edit to `agent_init.go`.

The discipline needs **enforcement mechanisms**, not just doctrine. The auto-sync workflow below provides that enforcement.

### Observation 6 — GitHub Actions workflow patterns viable for daily sync

`zehn/.github/workflows/nightly.yml` already runs a daily-build workflow at `cron: '0 0 * * *'` (midnight UTC). It uses `actions/checkout@v6` with `fetch-depth: 0` (full history), `actions/setup-go`, and produces nightly artifacts. The pattern is exactly what a daily sync workflow needs.

Existing workflow inventory:

```
build.yml         — push/PR build
create-tag.yml    — tag-driven release
create_dmg.yml    — macOS DMG packaging
docker-build.yml  — Docker image push
nightly.yml       — daily build (model for daily sync)
pr.yml            — the lint + vulncheck + test we've been fighting
release.yml       — full release
stale.yml         — stale issue/PR closer
upload-tos.yml    — ToS asset upload
```

None of them currently do upstream sync. The daily sync workflow would be a new file, e.g. `.github/workflows/upstream-sync.yml`, modeled on `nightly.yml`'s scheduling block.

### Observation 7 — Available tooling

For workflow-side work:

- **`gh` CLI** — already used heavily; GitHub Actions runners have it preinstalled with auth scoped to `GITHUB_TOKEN`.
- **`git`** — for the merge attempt, conflict detection, branch creation.
- **`actions/create-pull-request`** or `gh pr create` — to open the auto-sync PR.

For local-side work (when conflicts occur):

- **`git mergetool`** — manual conflict resolution.
- **`gofumpt`, `golines`, `gci`** — already discussed for lint cleanup; same tools needed during sync for files Zehn formatted differently than upstream.

For divergence tracking and reporting:

- **`git log --left-right`** — count ahead/behind.
- **`git rev-list --left-right --count`** — quick numeric summary.
- **`git log --not upstream/main --name-only`** — Zehn-only file touches.

### Observation 8 — Per-commit file-touch profile of recent work

For the 10 most recent Zehn commits, the upstream-shared file ratio is alarming for organization-* tasks:

| Commit | Files touched | Zehn-only | Upstream-shared |
|---|---|---|---|
| 8c4e710f (056+057) | 9 | 7 | 2 (the additive `agent_init.go` line + the Zehn-only-but-in-shared-dir `pkg/tools/integration/github_artifacts.go`) |
| 08a4abac (055 hardening) | 9 | 2 | **7** |
| 38542675 (053 log correlation) | 11 | 2 | **9** |
| fbc16f62 (052 drilldown) | 11 | 1 | **10** |
| cdb6703b (051 failure-reason) | 7 | 1 | **6** |

The organization-* tasks are the dominant contributors to divergence. Each task touches 6-10 upstream-shared files (mostly `web/frontend/`). Tasks 056/057 are an outlier in the opposite direction — mostly Zehn-only files, additive edits elsewhere.

**Lesson:** the organization-* feature cluster sits in territory where Zehn-only-namespace discipline is hard to enforce (the UI is necessarily integrated with the runtime). Future work in this area should either be re-architected to use upstream-clean extension points (hooks into existing UI), or accepted as the price of having a custom UI.

## Coexistence strategy — three enforced pillars

The 22 upstream-shared files Zehn has modified are the entire problem. The fix is a discipline that shrinks the modification surface AND prevents new modifications from being committed without a deliberate decision. Three pillars, each with an enforcement mechanism (not just doctrine):

### Pillar 1 — File-prefix namespacing (the `zehn_` convention)

**Rule:** Every Zehn-only Go file in a shared directory (`pkg/agent/`, `pkg/tools/`, `pkg/config/`, etc.) MUST have a `zehn_` filename prefix.

Tasks 056/057 set this precedent. Older Zehn-only files in shared dirs (`pkg/agent/delegation.go`, `eventbus.go`, `meeting.go`, `organization.go`, etc.) do NOT follow the convention and are at perpetual risk of upstream creating a file with the same name and triggering an `add/add` conflict (exactly what happened with `pkg/tools/delegate.go`).

**Enforcement:**
- A pre-commit hook (or CI check) that fails the build if a new file is created in a shared directory without the `zehn_` prefix UNLESS it's an explicit allow-list (e.g., `agent_init.go` modifications are allowed because that's the wire-up site).
- A scheduled rename task to migrate existing offenders: one task per file, each a small focused PR. Suggested phasing: rename the highest-conflict-risk files first (those upstream is most likely to create — `delegation.go`, `eventbus.go`, `meeting.go`, `organization.go`).

**Cost of the rename pass:** medium — each rename touches every file that imports the renamed package member. Several hundred lines of mechanical change. One-time work.

### Pillar 2 — Upstream-clean modification budget

**Rule:** Every Zehn-features task that modifies an upstream-shared file must include an `upstream_modifications` section in the task spec listing exactly which files and why. The reviewer's job is to push back when an upstream-clean extension point would have worked instead.

**Enforcement:**
- The task spec template (`supervision/zehn_feature_tasks/task_NNN-*.md`) gains a new required section.
- A divergence dashboard (described under Pillar 3 / Auto-sync) reports the total modification surface and trend.
- Quarterly review: which upstream-shared files have grown vs. shrunk in Zehn divergence?

**Goal:** the modification surface trends DOWN over time, not up. Today: 22 files. Target by 2026-07: under 10.

**Specific reductions to target:**
- `pkg/agent/agent_init.go` — already minimal (1 added call line); keep it that way.
- `pkg/agent/agent.go` — currently +28/-78 of struct edits. Refactor the Zehn-only fields into a sub-struct `Zehn ZehnAgentLoopExtensions` so all Zehn additions live in one block.
- `pkg/config/config.go` — currently +301/-290. This is the largest target. Move Zehn-side config into a `zehn` subsection (`config.Zehn ZehnConfig`) so upstream's other config additions don't interleave.
- `pkg/agent/registry.go` — Zehn's organization-aware registration could be a registered callback rather than an inline modification.

These are **architectural refactors**, not just renames. Each is a multi-hour focused PR; estimated 1-3 days of total work across all four.

### Pillar 3 — Extension points instead of modifications

**Rule:** When Zehn needs a new behavior that touches the runtime, prefer adding an EXTENSION POINT to upstream (one line, generic interface) and providing the Zehn implementation in a `zehn_*` file.

This was the pattern that worked beautifully for tasks 056/057:
- Added `wireZehnGitHubArtifactWriter(al, cfg)` call in `agent_init.go` (1 logical line + 2 lines of comment + blank)
- All the actual code lives in `pkg/agent/zehn_github_artifact_writer.go` + `pkg/agent/zehn_init_hook.go`
- Future upstream changes to `agent_init.go` only conflict with that single call line — minimal surface

**Pattern formalized:** for every Zehn-side feature that needs to hook into upstream runtime, add ONE call site (with a `wireZehn*` or `zehnHook*` function name) and put the implementation in `zehn_*` files. The call site is the only contract surface.

**Where this DOESN'T work:** the frontend (`web/frontend/src/components/agent/organization/**`). The Organization UI is necessarily integrated with the upstream React app structure. Two options for that surface:
- Accept the divergence; treat the entire `web/frontend/src/components/agent/organization/**` subtree as Zehn-only (which it effectively is) and document that.
- Build a "Zehn frontend extension" layer — a separate React mount point or a separate page route that doesn't share components with upstream. Larger refactor, possibly not worth it.

**Recommendation for the frontend:** accept current divergence (it's net-additive — Zehn ADDS organization views, doesn't modify upstream's views). Audit periodically.

## Daily auto-sync architecture

A new GitHub Actions workflow runs daily, attempts to merge upstream into a temporary branch, and surfaces outcomes as PRs or issues for human triage. The goal is **early detection** of upstream changes that affect Zehn-modified files — not automatic merging of conflicts.

### Architecture overview

```
┌────────────────────────────────────────────────────────────┐
│  .github/workflows/upstream-sync.yml                       │
│  Schedule: '0 2 * * *' (daily 02:00 UTC = 07:00 PKT)       │
└────────────────────────────────────────────────────────────┘
                          │
                          ▼
   ┌─────────────────────────────────────────────────┐
   │  Step 1: Fetch upstream/main (full history)     │
   └─────────────────────────────────────────────────┘
                          │
                          ▼
   ┌─────────────────────────────────────────────────┐
   │  Step 2: Compute divergence + write report      │
   │   - ahead/behind counts                         │
   │   - file-level conflict prediction              │
   │   - new upstream files matching Zehn-only names │
   └─────────────────────────────────────────────────┘
                          │
                          ▼
   ┌─────────────────────────────────────────────────┐
   │  Step 3: Attempt merge on temp branch           │
   │  branch: auto-sync/upstream-YYYYMMDD            │
   └─────────────────────────────────────────────────┘
                          │
              ┌───────────┼────────────┐
              ▼           ▼            ▼
       [clean merge]  [merge with    [unmergeable]
              │       conflicts]         │
              │           │              ▼
              ▼           ▼      Open/update issue
       Run lint+test  Don't push   "Sync conflict report"
       Open PR with   Open/update
       label          issue
       auto-sync:     auto-sync:
       clean          conflict
```

### Workflow YAML sketch (the file would land as `.github/workflows/upstream-sync.yml`)

```yaml
name: Upstream Sync (Daily)

on:
  schedule:
    - cron: '0 2 * * *'   # 02:00 UTC daily
  workflow_dispatch: {}    # manual trigger from Actions UI

permissions:
  contents: write
  pull-requests: write
  issues: write

jobs:
  attempt-sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0
          token: ${{ secrets.GITHUB_TOKEN }}

      - name: Configure git
        run: |
          git config user.name "zehn-auto-sync"
          git config user.email "noreply@mmaliabbas.com"

      - name: Add and fetch upstream
        run: |
          git remote add upstream https://github.com/sipeed/picoclaw.git || true
          git fetch upstream main --no-tags

      - name: Compute divergence
        id: divergence
        run: |
          AHEAD=$(git rev-list --count upstream/main..main)
          BEHIND=$(git rev-list --count main..upstream/main)
          echo "ahead=$AHEAD" >> $GITHUB_OUTPUT
          echo "behind=$BEHIND" >> $GITHUB_OUTPUT

      - name: Predict conflicts
        id: predict
        run: |
          # Files touched on both sides since divergence
          git log main --not upstream/main --name-only --pretty=format: \
              | sort -u > /tmp/zehn_files.txt
          git log upstream/main --not main --name-only --pretty=format: \
              | sort -u > /tmp/upstream_files.txt
          comm -12 /tmp/zehn_files.txt /tmp/upstream_files.txt > /tmp/shared.txt
          echo "shared_count=$(wc -l </tmp/shared.txt)" >> $GITHUB_OUTPUT

      - name: Detect add/add overlaps
        id: overlap
        run: |
          # New upstream files matching Zehn-only filenames
          git diff --name-only --diff-filter=A upstream/main main \
              | grep -v '^supervision/' | grep -v '^.picoclaw/' \
              > /tmp/zehn_added.txt
          OVERLAPS=$(while read f; do
              if git cat-file -e upstream/main:"$f" 2>/dev/null; then
                  echo "$f"
              fi
          done </tmp/zehn_added.txt | wc -l)
          echo "overlap_count=$OVERLAPS" >> $GITHUB_OUTPUT

      - name: Attempt merge
        id: merge
        run: |
          DATE=$(date -u +%Y%m%d)
          BRANCH="auto-sync/upstream-${DATE}"
          git checkout -b "$BRANCH"
          if git merge upstream/main --no-edit; then
              echo "result=clean" >> $GITHUB_OUTPUT
          else
              echo "result=conflict" >> $GITHUB_OUTPUT
              git diff --name-only --diff-filter=U > /tmp/conflicts.txt
              git merge --abort
          fi
          echo "branch=$BRANCH" >> $GITHUB_OUTPUT

      - name: On clean merge — run tests
        if: steps.merge.outputs.result == 'clean'
        id: tests
        run: |
          set +e
          go test -tags goolm,stdjson ./... > /tmp/test.log 2>&1
          rc=$?
          echo "result=$rc" >> "$GITHUB_OUTPUT"
          {
              echo "log<<EOF"
              tail -200 /tmp/test.log
              echo "EOF"
          } >> "$GITHUB_OUTPUT"
          exit 0
        continue-on-error: true

      - name: On clean merge — push and open PR
        if: steps.merge.outputs.result == 'clean'
        run: |
          git push origin "${{ steps.merge.outputs.branch }}"
          LABEL="auto-sync:clean"
          if [ "${{ steps.tests.outputs.result }}" != "0" ]; then
              LABEL="auto-sync:test-failure"
          fi
          gh pr create \
              --title "auto-sync: upstream/main → main ($(date -u +%Y-%m-%d))" \
              --body "Automated daily sync from sipeed/picoclaw.\n\nAhead: ${{ steps.divergence.outputs.ahead }}\nBehind: ${{ steps.divergence.outputs.behind }}\nShared-file changes: ${{ steps.predict.outputs.shared_count }}\nAdd/add overlaps: ${{ steps.overlap.outputs.overlap_count }}\n\nTests: ${{ steps.tests.outputs.result == '0' && 'PASS' || 'FAIL' }}\n\nTest log tail:\n\`\`\`\n${{ steps.tests.outputs.log }}\n\`\`\`\n\nNote: PRs created with GITHUB_TOKEN may not trigger every normal pull_request workflow in all repository configurations. Until a GitHub App token or PAT-based bot token is chosen, treat this sync workflow's own test result as the authoritative automated check." \
              --label "$LABEL,approval:ali-required"
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: On conflict — open or update sync-status issue
        if: steps.merge.outputs.result == 'conflict'
        run: |
          ISSUE_TITLE="Upstream sync blocked — conflicts in $(date -u +%Y-%m-%d) attempt"
          BODY=$(cat <<EOF
          Daily auto-sync attempt could not auto-merge upstream/main.

          ## Divergence
          - ahead: ${{ steps.divergence.outputs.ahead }}
          - behind: ${{ steps.divergence.outputs.behind }}
          - shared-file changes since last sync: ${{ steps.predict.outputs.shared_count }}
          - new upstream files conflicting with Zehn add/add: ${{ steps.overlap.outputs.overlap_count }}

          ## Conflict files
          \`\`\`
          $(cat /tmp/conflicts.txt)
          \`\`\`

          This issue stays open until a sync PR merges. Daily attempts append to this issue until resolved.
          EOF
          )
          # Find existing open auto-sync conflict issue; create or comment
          EXISTING=$(gh issue list --label auto-sync:conflict --state open --json number --jq '.[0].number')
          if [ -n "$EXISTING" ] && [ "$EXISTING" != "null" ]; then
              gh issue comment "$EXISTING" --body "$BODY"
          else
              gh issue create --title "$ISSUE_TITLE" --body "$BODY" --label "auto-sync:conflict,approval:ali-required"
          fi
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### Behaviors

| Outcome | Action | Surfaces as |
|---|---|---|
| Upstream is 0 commits ahead (we're already current) | No-op (workflow exits early) | Nothing |
| Clean merge + tests pass | PR opened, labeled `auto-sync:clean` + `approval:ali-required` | One PR per day on clean days; Ali reviews and merges or rejects |
| Clean merge + tests fail | PR opened, labeled `auto-sync:test-failure` + body includes test output | One PR per day; Ali decides whether to fix tests in the PR or revert |
| Merge conflict (the painful case) | Issue opened (or commented if one is already open), labeled `auto-sync:conflict` + `approval:ali-required`; no PR | One durable issue per conflict episode; daily attempts append to it until resolved |
| Add/add overlap detected (new upstream file matches a Zehn-only filename) | Even if the merge is clean, flag in the PR/issue body | Human attention drawn to the duplication |

### Observability

A small **divergence dashboard** runs in the same workflow (or as a sibling job) and writes to a permanent tracking issue (e.g., `#999 — Upstream Divergence Dashboard`):

- Ahead/behind counts over time (a multi-week table)
- Shared-file modification count over time
- Days since last clean sync
- Days since last conflict
- Top 10 most-touched Zehn-side files in last 30 days

This issue is updated daily by the workflow with a one-line append + edit. Lets Ali see trend at a glance.

## Decision framework — pick / coexist / merge / drop

When upstream and Zehn both deliver something at the same conceptual layer, this decision tree applies:

```
Q1. Does upstream's implementation already merge or look on track to merge?
    NO → Zehn proceeds independently. Continue to track upstream.
    YES → Q2

Q2. Does Zehn need features upstream's version doesn't have?
    NO → ADOPT upstream's, delete Zehn's, update references.
    YES → Q3

Q3. Can Zehn's extensions be expressed as additions to upstream's API,
     without modifying upstream's code?
    YES → MERGE: adopt upstream's base, port Zehn's extensions onto it
          in a `zehn_*` file or a thin wrapper.
    NO → Q4

Q4. Is the feature surface stable enough that diverging permanently is
    acceptable (the feature won't keep evolving upstream)?
    YES → COEXIST: rename Zehn's file to `zehn_<name>`, document
          the divergence in `supervision/ZEHN_UPSTREAM_DIVERGENCE.md`.
    NO → ESCALATE to Ali for explicit architectural call.
```

Applied to the two known overlaps (from companion sync plan):

| Overlap | Q1 | Q2 | Q3 | Decision |
|---|---|---|---|---|
| delegate-tool | yes (merged 2026-05-07) | yes (Zehn has `targetModel`, `maxTokens`, allowlist checker) | yes (can wrap upstream's tool) | **MERGE** — adopt upstream's `delegate` as base, port Zehn's extensions |
| runtime events | yes (merged) | partially (Zehn uses for Yaad logging; upstream for self-evolution) | borderline | **MERGE** (preferred) or **COEXIST** (acceptable) — companion sync plan's A1.b vs A1.a tradeoff |

### Anti-patterns to avoid

- **Silent re-implementation of an upstream PR**: discovering after the fact that you built what was already in flight upstream. The daily-sync workflow's "add/add overlap detection" catches this within 24h.
- **Re-merging the same conflict from scratch every sync**: each conflict resolution should record the decision (in `supervision/ZEHN_UPSTREAM_DIVERGENCE.md`) so future syncs don't re-litigate.
- **Allowing modification-surface growth without a justification**: every Zehn task touching an upstream file must explain why in the spec. CI check enforces this.

## Migration roadmap

Five phases. Each is independently shippable.

### Phase 1 — One-time bring-current sync

Per the companion document `supervision/ZEHN_UPSTREAM_SYNC_PLAN_20260516.md`. Resolve current 189-commit backlog with explicit decisions on A1 (event-bus) and A2 (delegate-tool). 2.5-3.5 days.

This is the prerequisite: the daily sync starts from a clean baseline.

### Phase 2 — File rename pass (Pillar 1)

Migrate the existing offenders to `zehn_*` prefix. Each rename is a small focused PR:

- Suggested order (highest collision risk first):
  1. `pkg/agent/delegation.go` → `pkg/agent/zehn_delegation.go`
  2. `pkg/agent/eventbus.go` + ecosystem → `pkg/agent/zehn_eventbus.go` etc.
  3. `pkg/agent/meeting.go` + ecosystem → `pkg/agent/zehn_meeting.go` etc.
  4. `pkg/agent/organization.go` + test → `pkg/agent/zehn_organization.go`
  5. `pkg/tools/delegate.go` → handled by A2 decision in companion plan (if adopted upstream's, this becomes a delete + zehn_-prefixed wrapper for extensions)
  6. `pkg/tools/meeting.go` → `pkg/tools/zehn_meeting.go`

Each rename is 1-2 hours including import updates and test fixes. Total: 1-2 days of focused work, parallelizable across PRs.

### Phase 3 — Modification surface refactor (Pillar 2)

Highest-value targets:

1. **`pkg/config/config.go`** — refactor Zehn-side config into a nested `Zehn ZehnConfig` struct. Single source of upstream-file modification becomes ONE line (the embed) instead of 300 interleaved lines. Estimated: 4 hours.
2. **`pkg/agent/agent.go`** — collect Zehn-only `AgentLoop` fields into a nested `Zehn *ZehnAgentLoopExtensions`. Same idea. Estimated: 3 hours.
3. **`pkg/agent/registry.go`** — make Zehn's registration paths use a hook pattern rather than inline mods. Estimated: 4 hours.
4. **`pkg/channels/manager.go`** — same hook pattern. Estimated: 2 hours.

After this phase, the upstream-shared modification surface drops from 22 files to ~6-8 files, all with minimal Zehn footprint.

### Phase 4 — Daily auto-sync workflow activation

Land `.github/workflows/upstream-sync.yml`. Watch first run. Tune thresholds (conflict labels, escalation cadence). One PR.

Activation isn't risky because the workflow doesn't auto-merge — it just opens PRs and issues. Worst case: the workflow itself has a bug and posts nothing; no harm.

Estimated: 1 day to write, test, and iterate.

### Phase 5 — Ongoing discipline + extension-point pattern

Every new zehn-features task gets:

- An `upstream_modifications` section in the task spec
- A reviewer check (or CI assertion) that new files in shared dirs use `zehn_` prefix
- A reviewer check that any upstream-file modification is justified

This is cultural enforcement. The daily workflow + dashboard make backsliding visible.

## Open questions for Ali

1. **Approval ordering**: Phase 1 (one-time sync) needs to land before Phase 2-5 are useful. The companion sync plan needs A1+A2 decisions to start. Recommend: decide A1 and A2 this week, schedule the 3-day sync window next week. Phase 2-5 happen incrementally after.

2. **Daily-sync schedule**: 02:00 UTC = 07:00 PKT. Reasonable? Alternative: 22:00 UTC = 03:00 PKT (overnight in Pakistan, no chance of conflicting with active work). The cron expression in the YAML is one-line to change.

3. **Auto-sync PR auto-merge?**: Should clean merges with passing tests auto-merge after a 48-hour review window, or always require Ali approval? Recommend: always require approval initially; revisit after 4 successful clean syncs.

4. **PR workflow token model**: should the daily sync use default `GITHUB_TOKEN`, a GitHub App token, or a PAT-backed bot token? `GITHUB_TOKEN` is simplest, but PRs it creates may not trigger every normal PR workflow depending on repo settings. Recommend: start with `GITHUB_TOKEN` and make the sync workflow's own test output authoritative; move to a GitHub App token after the first clean week if normal PR checks are required.

5. **Divergence dashboard scope**: post to a tracking GitHub issue, or to Discord, or both? Recommend: GitHub issue (durable, queryable), with a daily Discord ping linking to the issue.

6. **Phase 2 rename pass cost**: ~1-2 days. Worth pursuing now, or defer until after Phase 1 completes? Recommend: defer — renaming files in the middle of a sync makes everything harder.

7. **Phase 3 refactor cost**: 13 hours total across four files. Worth doing as one push, or split into per-file PRs? Recommend: per-file PRs, one per week. Each refactor is independent.

8. **Frontend divergence**: accept it (audit-only) or invest in a separation layer? Recommend: accept it; the Organization UI is fundamentally Zehn-specific.

9. **Existing zehn supervision artifact `supervision/ZEHN_UPSTREAM_PUBLISHING_CHECKLIST.md` is untracked** — should be reviewed for prior thinking before Phase 4 lands.

## Cost summary

| Phase | Effort | Risk | Blocker |
|---|---|---|---|
| Phase 1 — one-time sync | 2.5-3.5 days | Medium (A1+A2 decisions) | Companion sync plan |
| Phase 2 — file renames | 1-2 days | Low (mechanical) | Phase 1 |
| Phase 3 — modification refactor | ~13 hours across weeks | Medium (touches hot files) | Phase 1 |
| Phase 4 — daily auto-sync | 1 day | Low (additive, observe-only) | None (could land in parallel with Phase 1) |
| Phase 5 — ongoing discipline | continuous, ~30 min per future task | Low | None |
| **Total upfront** | **~5-7 days of focused work**, spread over 2-4 weeks |

After this investment, future upstream syncs become routine: daily auto-PRs for clean weeks, surfaced issues for conflict weeks, drastically reduced surprise.

## What this plan deliberately doesn't address

- **Backporting Zehn features to upstream** (`supervision/ZEHN_UPSTREAM_PUBLISHING_CHECKLIST.md` territory). Goal here is to keep the fork sustainable, not to contribute back. Backporting is a separate decision per-feature.
- **Vendoring upstream as a submodule**. Considered and rejected: too invasive a refactor for the gain.
- **Forking the fork**: i.e., zehn becomes a vanilla picoclaw fork plus a separate Zehn-features overlay distributed via a different repo. Could be a long-term direction; not in scope now.
- **The agent rescan problem** (companion document `supervision/ZEHN_AGENT_STATE_CACHE_PLAN_20260516.md`). Independent of sync work, can proceed in parallel.
