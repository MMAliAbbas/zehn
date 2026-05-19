# Zehn Stopped Runtime Investigation - 2026-05-13

## Scope

Zehn was stopped before investigation. This file investigates only the first two runtime problems found after the 2026-05-12 restart:

1. LogicIgniter MCP runtime proof falls back to Docker even though this machine is host-native.
2. MCP integration tests reported `identity DB preflight failed: required table "tenants" is missing`.

Other observed problems are captured only as a todo list at the end.

## Stop Confirmation

- Command used: `launchctl bootout gui/$(id -u) /Users/aliai/Library/LaunchAgents/io.picoclaw.launcher.plist`
- Verification:
  - `lsof -nP -iTCP:18790 -sTCP:LISTEN` returned no listener.
  - `lsof -nP -iTCP:18800 -sTCP:LISTEN` returned no listener.

## Problem 1: Docker Fallback In MCP Runtime Proof

### Evidence

Post-restart Zehn logs contained this failure:

```text
starting local infrastructure, identity, billing, and direct gRPC services...
==> Starting Keycloak + Postgres via Docker Compose...
/Users/aliai/logicigniter/scripts/local-preview/start-real-stack.sh: line 69: docker: command not found
```

Relevant files inspected:

- `/Users/aliai/logicigniter/scripts/local-preview/start-mcp-runtime-proof.sh`
- `/Users/aliai/logicigniter/scripts/local-preview/start-real-stack.sh`
- `/Users/aliai/logicigniter/scripts/local-preview/run-all-services.sh`
- `/Users/aliai/logicigniter/scripts/local-preview/start-identity.sh`

`start-mcp-runtime-proof.sh` currently checks readiness like this:

```bash
curl -sf -m 3 "$KEYCLOAK_URL/realms/master" >/dev/null 2>&1 &&
  docker exec -e PGPASSWORD=postgres "${PG_CONTAINER:-li-localpreview-postgres-1}" \
    pg_isready -U postgres >/dev/null 2>&1
```

If that check fails, it runs:

```bash
COMPOSE_PROJECT_NAME="${COMPOSE_PROJECT_NAME:-li-localpreview}" \
  bash "$SCRIPT_DIR/start-real-stack.sh" --detach
```

`start-real-stack.sh` is explicitly Docker-based:

- Starts Keycloak and Postgres with `docker compose up -d`.
- Runs `docker compose exec -T postgres pg_isready`.
- Creates DBs through `docker compose exec -T postgres psql`.

### Root Cause

`start-mcp-runtime-proof.sh` has not been fully adapted to the host-native local environment. Even though direct host-native scripts exist (`start-identity.sh`, `start-billing.sh`, `run-all-services.sh`, `start-all-grpc.sh`), the MCP runtime proof still uses Docker both in its readiness probe and fallback startup path.

### Required Fix Direction

Do not install Docker just to satisfy this path. The correct fix is to make the MCP runtime proof host-native aware:

- Detect host-native Keycloak at `http://localhost:8180`.
- Detect host-native Postgres using `pg_isready` / `psql` directly, not `docker exec`.
- If infra is not ready, fail with a precise host-native prerequisite message or call the approved host-native Keycloak/Postgres startup path.
- Keep Docker support optional only when Docker is explicitly available.

## Problem 2: Identity DB Preflight Missing `tenants`

### Evidence

Post-restart Zehn logs contained this integration test failure:

```text
identity DB preflight failed: required table "tenants" is missing.
Set IDENTITY_DB_DSN to the host-native identity database
(expected svc_identity for local preview) before running MCP integration tests.
```

Postgres logs also show repeated real SQL failures from 2026-05-12:

```text
ERROR: relation "tenants" does not exist
STATEMENT: INSERT INTO tenants (id, keycloak_user_id, company_name) ...
```

and service-domain fixture failures:

```text
ERROR: relation "accounts" does not exist
STATEMENT: INSERT INTO accounts ...
```

Current read-only DB check shows the situation has changed:

```text
database: svc_identity
public.tenants: present
public.api_keys: present
public.tenant_bundle_subscriptions: present
```

Focused current test:

```bash
IDENTITY_URL=http://localhost:8090 \
BFF_URL=http://localhost:8091 \
KEYCLOAK_URL=http://localhost:8180 \
DB_DSN=postgres://postgres:postgres@localhost:5432/logicigniter_local \
SERVICE_SCHEMA_PREFIX=svc_ \
IDENTITY_DB_DSN=postgres://postgres:postgres@localhost:5432/svc_identity \
go test -tags integration ./tests/mcp -run TestMCPStaticPrerequisites -count=1 -v
```

Result:

```text
PASS
ok github.com/logicigniter/integration_tests/tests/mcp
```

Current service reachability:

- `http://localhost:8090/healthz`: not reachable
- `http://localhost:8091/healthz`: not reachable
- `http://localhost:8093/healthz`: not reachable
- `http://localhost:8180/realms/master`: not reachable

### Root Cause

The logged failure was real at the time, but the current database schema no longer reproduces the missing `tenants` condition. The stronger root cause is local runtime state inconsistency:

- MCP tests expect the host-native identity DB at `svc_identity`.
- Earlier test runs connected to a DB state where identity migrations had not populated `tenants`.
- The current DB has the required identity tables, but the local identity/BFF/MCP/Keycloak services are down.
- Therefore the system lacks one canonical host-native "prepare and verify local MCP runtime" entrypoint that starts services, applies migrations, verifies schemas, then runs tests.

### Required Fix Direction

Fix this as a local-runtime orchestration problem, not as a test workaround:

- Before final MCP runtime tests, verify `svc_identity` exists and contains `tenants`, `api_keys`, and `tenant_bundle_subscriptions`.
- Verify `logicigniter_local` service schemas and required app tables before running runtime tests.
- Make the startup proof script run identity migrations before tests and report exact missing schemas/tables.
- Keep `IDENTITY_DB_DSN=postgres://postgres:postgres@localhost:5432/svc_identity` explicit in runtime proof output.
- Ensure local services are actually up before reporting MCP runtime ready.

## Todo: Remaining Runtime Issues

These were not investigated or fixed in this pass.

1. Investigate PR check readiness.
   - Symptom: `gh pr checks` reports `no checks reported` on active branches.
   - Goal: determine whether this is expected for repos without Actions, missing workflow triggers, wrong branch state, or delayed checks.

2. Investigate 12-minute command timeout.
   - Symptom: one Zehn tool execution hit `Command timed out after 12m0s`.
   - Goal: classify whether long verification should be split, moved to a background runner, or given explicit longer command timeouts.

3. Investigate provider streaming warning.
   - Symptom: frequent `Codex completed response had empty output; reconstructed output from streamed output_item.done events`.
   - Goal: confirm whether this is harmless provider behavior, schema/stream parsing drift, or a reliability risk.

4. Investigate internal outbound channel warning.
   - Symptom: frequent `Unknown channel for outbound message` for internal delegation flows.
   - Goal: confirm whether internal delegation responses should be suppressed, routed to an inbox, or handled as a first-class internal channel.

5. Investigate complete autonomous delivery loop.
   - Symptom: agents inspect/delegate/review, but there is not yet evidence of a full issue-to-branch-to-PR-to-review-to-merge-to-post-merge-restart cycle completing after the restart.
   - Goal: verify the end-to-end process on one low-risk issue before scaling.

6. Investigate stale or dirty repo blockers.
   - Symptom: earlier PR checkout was blocked by dirty local files.
   - Goal: enforce "never leave repo dirty" and define what agents should do when a repo is dirty before checkout or merge.

7. Investigate GitHub field usage.
   - Symptom: `Unknown JSON field: "reviewDecision"` occurred on an issue command.
   - Goal: separate issue JSON fields from PR JSON fields in prompts/scripts.

8. Investigate LogicIgniter repo branch posture.
   - Symptom: some repos are on active task branches, e.g. `scripts` on `chore/3-standard-pr-verification`.
   - Goal: decide whether Zehn should operate from main only, worktree per issue, or current branch with strict dirty-state rules.

9. Investigate local runtime service authority.
   - Symptom: local LogicIgniter services are not currently reachable while DB tables exist.
   - Goal: define one canonical host-native start/status/stop command set for Keycloak, Postgres, identity, billing, BFF, MCP, and 51 gRPC services.

