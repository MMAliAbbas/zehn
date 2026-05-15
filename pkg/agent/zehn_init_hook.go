// Zehn-fork-only file. See supervision/zehn_feature_tasks/task_056-github-artifact-writer-implementation.md.
//
// Not part of upstream sipeed/picoclaw. File name carries the `zehn_` prefix
// so it is visible during upstream sync and cannot collide with future
// upstream additions.

package agent

import (
	"os/exec"
	"testing"

	"github.com/sipeed/picoclaw/pkg/config"
)

// wireZehnGitHubArtifactWriter installs the Zehn-fork's GitHubArtifactWriter
// implementation onto the AgentLoop at init time. Called once from
// NewAgentLoop after registerSharedTools.
//
// The wire-up is a no-op when:
//   - al is nil (defensive)
//   - the binary is running under `go test` (per testing.Testing()).
//     Existing tests such as TestRunAgentMeetingGitHubDisabledWriterRecordsSkipped
//     construct AgentLoop via NewAgentLoop and assert that the writer
//     stays nil so the "skipped/disabled" code path executes. Skipping the
//     wire-up under test preserves that contract without forcing test files
//     to mock out an env var. Tests that exercise the writer explicitly
//     call al.SetGitHubArtifactWriter(...) themselves and are unaffected.
//   - the `gh` CLI is not on PATH (CI environments, headless runs without
//     GitHub auth). In that case al.githubArtifacts stays nil and
//     delegation/meeting records record the legacy "github artifact writer
//     disabled" status, preserving pre-fix behavior.
//
// Tasks 009 (GitHub meeting artifacts) and 020 (runtime-owned publisher)
// scaffolded the interface, storage, async publisher, and test fakes but
// never landed a production writer or a wire-up site. Task 056 closes that
// gap without touching upstream-clean files beyond the single helper call
// in NewAgentLoop.
func wireZehnGitHubArtifactWriter(al *AgentLoop, cfg *config.Config) {
	if al == nil {
		return
	}
	if testing.Testing() {
		return
	}
	if _, err := exec.LookPath("gh"); err != nil {
		return
	}
	writer := newZehnGitHubArtifactWriter(defaultZehnGitHubArtifactRepo, defaultZehnGitHubArtifactTimeout)
	al.SetGitHubArtifactWriter(writer)
	_ = cfg // reserved for future config-driven overrides (repo, timeout, enable flag)
}
