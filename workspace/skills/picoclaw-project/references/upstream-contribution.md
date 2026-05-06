# Upstream Contribution Strategy

## Philosophy

Become upstream contributors first, then maintainers by demonstrated reliability. Start with small, high-quality PRs that fix real PicoClaw problems without pushing Zehn-specific agenda. Keep private memory and Yaad integration out of upstream unless there is a clean, generic extension point.

## Before Changing Code

Read `CONTRIBUTING.md`, `AGENTS.md`, nearby tests, and the relevant source path. Check upstream context before assuming a local fork is stale. Avoid branch names containing `codex`, `agent`, or `ai`; use `fix/...` or `feature/...`.

## Good First PR Types

- Documentation clarifications that reflect existing behavior.
- Test hardening for flaky or environment-sensitive launcher/gateway behavior.
- Small security improvements with narrow blast radius and clear tests.
- Bug fixes supported by failing tests and source evidence.
- Configuration UX improvements that preserve defaults and compatibility.

## Avoid

- Zehn-specific names, Yaad requirements, or private roadmap assumptions.
- Large rewrites before trust is earned.
- Changes that disrupt long-running sessions, WebSockets, streaming, steering, or gateway lifecycle without broad evidence.
- “Safety” patches that only address one narrow symptom while changing core semantics.

## Delegation/Meeting Publishability

The Zehn fork's delegation and meeting work can become upstreamable only after private automation is split away from generic PicoClaw code. Before publishing, run the local publishability audit and treat warnings for `workspace/skills/**`, `.picoclaw/**`, `supervision/zehn_feature_tasks/**`, and `supervision/ZEHN_FEATURE_*` as blockers for upstream-facing branches.

Upstreamable slices should be small and generic:

- Target-agent delegation primitive with allowlist checks and tests.
- Local redacted record stores and status/inbox visibility.
- Bounded async delegation executor.
- Meeting v1 with explicit chaired sequential semantics.
- Redacted, lifecycle-owned artifact publishing as a generic tracker adapter.

Keep Yaad-specific wording, private operating model, Discord channel map, and local automation ledgers out of upstream commits. Rebuild clean branches from upstream `main` rather than pushing mixed local history.
