# Testing And Verification

## Expected Style

Go tests are colocated as `*_test.go`. Shared behavior generally uses table-driven tests. Match the local style in the package being changed, especially around gateway and launcher tests.

Use `gofmt` for Go code. Frontend code uses TypeScript, React, ESLint, Prettier, sorted imports, no semicolons, and Tailwind class sorting. Do not hand-edit generated frontend route trees.

## Preferred Commands

- `make generate`: verify generation.
- `make test`: authoritative repository Go test command.
- `make check`: broader pre-PR verification.
- `go test -run TestName -v ./pkg/...`: targeted iteration only.
- `cd web/frontend && pnpm build && pnpm lint`: frontend verification.

## Delegation And Meeting Verification

For delegation/meeting changes, run focused normal and race tests before claiming the feature is safe:

```bash
go test ./pkg/agent ./pkg/tools ./pkg/config -run 'Delegation|Delegate|Meeting|GitHub|Artifact|Memory|Status|Inbox|Publisher|Participant|Failure|Cancel' -count=1
go test ./pkg/agent ./pkg/tools ./pkg/config -run 'Delegation|Delegate|Meeting|GitHub|Artifact|Memory|Status|Inbox|Publisher|Participant|Failure|Cancel' -race
```

For broader confidence without launcher/backend listener dependencies:

```bash
go test ./pkg/agent ./pkg/tools ./pkg/config ./pkg/channels/discord -count=1
```

Review tests should cover: permission denial, missing target, sync/async success, async capacity/cancellation/shutdown, status/inbox visibility, local record redaction, GitHub artifact redaction, publisher timeout/failure, Yaad unavailable/strict/idempotent behavior, meeting participant failure, chair failure, cancellation, and the end-to-end sponsor-chair-participant recommendation path.

## Current Sandbox Findings

In this environment, `make generate` completed successfully.

`make test` failed in the launcher/backend phase for environment reasons that need outside-sandbox confirmation:

- `web/backend` could not download `fyne.io/systray` because network access to `proxy.golang.org` was blocked.
- `web/backend/api` tests using local listeners and health probes failed with `operation not permitted`.
- Direct `go list ./...` without Makefile cache settings failed because the default Go build cache path was outside the writable sandbox.

Do not claim the suite passes until these are rerun in an environment with module cache access and local listener permission.
