# Zehn Exec Safety Guard Investigation - 2026-05-12

## Scope

This investigation covers the recurring `Command blocked by safety guard
(dangerous pattern detected)` failures observed in the Zehn gateway logs. No
exec safety configuration or Go code was changed as part of this investigation.

## Current Config

From `.picoclaw/config.json`:

- `tools.exec.enabled`: `true`
- `tools.exec.enable_deny_patterns`: `true`
- `tools.exec.allow_remote`: `true`
- `tools.exec.timeout_seconds`: `720`
- `tools.exec.custom_deny_patterns`: `null`
- `tools.exec.custom_allow_patterns`: `null`

Therefore the active behavior is the built-in PicoClaw deny-pattern list in
`pkg/tools/shell.go`.

## Built-In Deny Patterns That Matter Operationally

The default guard blocks, among other things:

- command substitution: `$()`
- variable expansion form: `${...}`
- heredocs: `<<EOF`
- `kill`, `killall`, `pkill`
- `git push`
- `sudo`
- `chmod 755`-style numeric chmod
- `docker run` and `docker exec`
- destructive commands such as `rm -rf`

## Evidence From Recent Logs

Recent blocked commands were not simple `gh` or `git status` calls. They were
mostly complex shell fragments or process-control attempts, for example:

- `ps eww -p $(lsof -tiTCP:8091 ...) ... | grep ...`
  - blocked because of `$()` and pipes/complex shell form.
- multi-line evidence script using `TS=$(date ...)`, `${TS}`, loops, `tee`,
  and temp files;
  - blocked because of command substitution, `${...}`, and script-like shell.
- `kill 16518`
  - blocked because `kill` is explicitly denied.
- `for item in ...; do ... ${item%#*} ...; done`
  - blocked because of shell loop and `${...}` expansion.
- heredoc temp-script creation:
  - `cat > /tmp/li_qa_pr_check.sh <<'EOF' ...`
  - blocked because heredocs are denied.

## Important Finding

The safety guard is not merely blocking dangerous destructive commands. It is
also blocking normal autonomous-company operations that our prompts currently
expect agents to perform:

- branch push after verified issue work;
- local service restart through `kill`/`pkill` style process control;
- compact multi-step evidence collection;
- temporary verification helper scripts.

This creates a direct contradiction:

- LogicIgniter operating prompts require agents to create issue branches, push
  branches, open PRs, and sometimes request DevOps restart/reconciliation.
- The default exec deny list blocks `git push` and local process-control
  commands.

## Current Risk

If we only keep telling agents to use simpler commands, some failures will
reduce, but the core contradiction remains. Agents can inspect, review, and
comment, but they cannot reliably complete the full autonomous delivery loop
when it requires pushing a branch or restarting a local service.

## Recommendation

Do not blindly loosen all exec safety.

The proper fix should be a narrow, explicit allow policy for the exact
LogicIgniter operations Zehn is expected to perform, such as:

- allow `git push origin <issue-number-branch>` or agreed issue branch pattern;
- allow project-owned restart scripts rather than raw `kill`/`pkill`;
- allow approved verification scripts under `/Users/aliai/logicigniter/scripts`
  and repo-local verification commands;
- keep destructive patterns blocked.

Until that is implemented, Zehn should report these failures as tooling-policy
blockers, not as engineering blockers inside LogicIgniter.
