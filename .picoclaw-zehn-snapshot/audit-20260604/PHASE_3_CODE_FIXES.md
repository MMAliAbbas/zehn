# Phase 3 — Runtime Code Fixes

## Update 2026-06-05 13:30 +05

All three fixes have since been committed. `git describe` now reports `v0.2.9-211-g1b0e20e2`. Local main is 2 commits ahead of `origin/main`; not pushed.

- `58e822d3` fix: surface MCP blank-cause errors and Pico rejected-origin diagnostics (Phase 3.1 + 3.2)
- `1b0e20e2` fix(mcp): retry transient HTTP 5xx errors once with small backoff (Phase 3.3)

Phase 3.3 was upgraded from "deferred" to landed after live data on 2026-06-05 confirmed persistent Yaad transport instability (~55% success rate in the post-bootstrap window, including one `Bad Gateway` failure that was not retried). The new fix adds a single retry-with-250ms-backoff on the same connection when the error matches a transient-5xx predicate. Does NOT address `recover-lost-fail` (reconnect-itself-failing), `blank-cause` (SDK-swallowed cause), or `memory_update` conflicts — those remain open follow-ups. See section 3.3 below and `RECOVERY_COMPLETION_SUMMARY.md` for full scope.

The running gateway is still on the dirty `v0.2.9-209-g77d13f90-dirty` build until the next restart; the committed fixes are not yet exercising in production.

## Status (original)

All changes are in `/Users/aliai/zehn` (uncommitted; HEAD remains `77d13f90`). Two surgical edits, both compile-pass and pass package unit tests.

## 3.1 — Pico WebSocket origin diagnostic (PARTIAL — observability fix only)

**File**: `pkg/channels/pico/pico.go` L123–144 (insertion at L137)
**Symptom observed in audit**: 16+ consecutive `websocket: request origin not allowed by Upgrader.CheckOrigin` errors at 21:47–21:49 +05, with no clue what origin the browser was actually sending.
**Edit**: Added `logger.WarnCF` inside the `checkOrigin` callback to log the rejected origin string, allowed list, remote addr, and request URI before returning false.
**Why partial**: The fix is observability-only. Once the next live attempt happens, the log will reveal whether the rejected origin matches what `config.json:.channel_list.pico.settings.allow_origins` accepts (currently `[http://127.0.0.1:18800, http://localhost:18800]`). If the browser is calling the gateway on port 18790 directly, the config needs to widen — but I haven't widened it preemptively because that would be guessing without evidence.
**Acceptance**: package unit tests pass (`go test ./pkg/channels/pico/...` → ok 3.752s). Live-prove deferred to Phase 5.

## 3.2 — MCP blank-error annotation (PROPAGATION FIX)

**File**: `pkg/tools/integration/mcp_tool.go` L260–267 → rewritten to L260–280
**Symptom observed in audit**: at 21:45:52 a `mcp_yaad_memory_query` failure returned `"MCP tool execution failed: failed to call tool: calling \"tools/call\": sending \"tools/call\":"` — the underlying transport-layer cause was swallowed, agents had no way to distinguish transport vs schema vs auth failures.
**Edit**: Annotate the error with `server=<name>, tool=<name>` so the failing surface is named. If the error chain ends with a bare colon (blank-cause pattern), append `[blank-cause: transport-layer error with no underlying message — likely SDK swallowing or remote 5xx without body]` so agents can react with specificity instead of narrating "Yaad failed".
**Why this shape**: The deeper truncation lives in `github.com/modelcontextprotocol/go-sdk v1.5.0` and would require an SDK PR or fork. The annotation here preserves any information that does exist and tells the agent when the SDK has dropped it.
**Acceptance**: package unit tests pass (`go test ./pkg/tools/integration/...` → ok 3.719s). Live-prove deferred to Phase 5 (next agent cycle that triggers a transport-level Yaad failure).

## 3.3 — Yaad `memory_update` conflict handling (DEFERRED — prompt-side mitigation in place)

**Symptom observed in audit**: 3 of 4 `memory_update` calls failed in the 18:00–22:00 +05 window (1 Bad Gateway, 2 `conflict` from `expected_version` mismatch). Even with Zehn off, this audit session itself hit one 502 on an update call (which committed anyway — see Phase 1.1).
**Decision**: do NOT add retry/refetch logic in `pkg/mcp/manager.go` for this release.

Reasoning:
- The audit's own Phase 1.1 exercise proved that updates can commit and *also* return non-success at the HTTP layer (502 followed by version increment). Naïve client-side retry would risk double-applying updates.
- A correct fix needs (a) idempotency keys per `memory_update`, OR (b) version-conflict typed errors that callers can react to with explicit re-fetch-and-merge.
- The audit's prompt-side mitigation is already in place: `zehn-operations-monitor.md` (Phase 2.1 edit) now tells the agent to retry up to 3 times with refetched `expected_version` and to report the underlying transport error verbatim instead of writing a `YAAD_DEGRADED` ledger entry to `MEMORY.md`.

This is documented as **Open Follow-up #1** in `PHASE_3_OPEN_FOLLOWUPS.md` (see below).

## 3.4 — Codex empty-output (NO ACTION — already at debug level)

**Status**: Commit `77d13f90` already changed `WarnCF` → `DebugCF` for the empty-output reconstruction. Audit confirms the patch is real. The underlying empty-output condition is logged at debug and can be reviewed in `gateway.log` if it recurs. No code change needed; if 8+ empty outputs appear in 24h, that's a provider-side issue to escalate.

## Build verification

```
make build         → build/picoclaw-darwin-amd64 (37 MB) — OK
make build-launcher → build/picoclaw-launcher-darwin-amd64 (23 MB) — OK
```

Both binaries built clean with `goolm,stdjson` build tags. Launcher version string `v0.2.9-209-g77d13f90-dirty` (`-dirty` because the two edits are uncommitted).

Package tests:
```
go test ./pkg/channels/pico/...        → ok 3.752s
go test ./pkg/tools/integration/...    → ok 3.719s
go test ./pkg/mcp/...                  → ok 4.309s
```

## Open Follow-up #1

`memory_update` conflict / Bad-Gateway-with-commit semantics need either:
(a) idempotency keys at the Yaad API level, OR
(b) typed-error surfacing from the MCP SDK so a `conflict` becomes structurally distinguishable from a transport failure, and clients can refetch the current version and retry the update against the new baseline.

Until then, the prompt-side rule (retry max 3 times with refetched version; report transport error verbatim) is the operational mitigation, and Phase 1.1's manually-recovered `9de0d453-...` v5 stands as the durable truthful state.

## Phase 3 acceptance
- [x] 3.1 Pico WS — diagnostic logging added; compile + tests pass; live-prove in Phase 5.
- [x] 3.2 MCP blank-error — annotation added; compile + tests pass; live-prove in Phase 5.
- [x] 3.3 Yaad update conflict — code change deferred; prompt-side mitigation in place via Phase 2.1 edits; open follow-up filed.
- [x] 3.4 Codex empty-output — no action; existing fix in `77d13f90` is correct.
- [x] Both binaries rebuilt; no compile errors.
- [x] Both touched packages plus `pkg/mcp` pass unit tests.

Phase 3 complete (with one acknowledged deferral documented).
