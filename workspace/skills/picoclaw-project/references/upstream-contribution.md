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

