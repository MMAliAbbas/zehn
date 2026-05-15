// Zehn-fork-only file. See supervision/zehn_feature_tasks/task_056-github-artifact-writer-implementation.md.
//
// Not part of upstream sipeed/picoclaw. File name carries the `zehn_` prefix
// so it is visible during upstream sync and cannot collide with future
// upstream additions.

package agent

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	integrationtools "github.com/sipeed/picoclaw/pkg/tools/integration"
)

const (
	defaultZehnGitHubArtifactRepo    = "logicigniter/supervision"
	defaultZehnGitHubArtifactTimeout = 30 * time.Second
)

// zehnGitHubArtifactWriter shells out to the `gh` CLI to file GitHub issues
// and comments for delegation and meeting records. The Zehn fork's
// AgentLoop.SetGitHubArtifactWriter receives an instance of this from
// wireZehnGitHubArtifactWriter at init time.
type zehnGitHubArtifactWriter struct {
	defaultRepo string
	timeout     time.Duration
	resolver    *zehnGitHubRepoResolver

	// execCmd is injectable so unit tests can substitute a fake without
	// touching os/exec. Production callers leave this as the default,
	// which invokes `gh` via exec.CommandContext.
	execCmd func(ctx context.Context, name string, args ...string) ([]byte, error)
}

func newZehnGitHubArtifactWriter(defaultRepo string, timeout time.Duration) *zehnGitHubArtifactWriter {
	if defaultRepo == "" {
		defaultRepo = defaultZehnGitHubArtifactRepo
	}
	if timeout <= 0 {
		timeout = defaultZehnGitHubArtifactTimeout
	}
	return &zehnGitHubArtifactWriter{
		defaultRepo: defaultRepo,
		timeout:     timeout,
		resolver:    newZehnGitHubRepoResolver(defaultLogicIgniterRoot, defaultRepo),
		execCmd: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return exec.CommandContext(ctx, name, args...).Output()
		},
	}
}

// CreateIssue invokes `gh issue create --repo R --title T --body B [--label L]...`
// and returns the parsed issue number and URL on success.
func (w *zehnGitHubArtifactWriter) CreateIssue(ctx context.Context, req integrationtools.GitHubIssueRequest) (integrationtools.GitHubIssueArtifact, error) {
	if w == nil {
		return integrationtools.GitHubIssueArtifact{}, errors.New("zehn github artifact writer is nil")
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return integrationtools.GitHubIssueArtifact{}, errors.New("github issue title is empty")
	}

	repo := w.resolver.resolveRepo(req)
	if repo == "" {
		repo = w.defaultRepo
	}

	args := []string{"issue", "create", "--repo", repo, "--title", title, "--body", req.Body}
	for _, label := range req.Labels {
		label = strings.TrimSpace(label)
		if label == "" {
			continue
		}
		args = append(args, "--label", label)
	}

	cctx, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()

	out, err := w.execCmd(cctx, "gh", args...)
	if err != nil {
		return integrationtools.GitHubIssueArtifact{}, fmt.Errorf("gh issue create: %w", err)
	}
	url := strings.TrimSpace(string(out))
	if url == "" {
		return integrationtools.GitHubIssueArtifact{}, errors.New("gh issue create returned empty output")
	}
	return integrationtools.GitHubIssueArtifact{
		Number: parseIssueNumberFromURL(url),
		URL:    url,
	}, nil
}

// CreateComment invokes `gh issue comment <number-or-url> --body B [--repo R]`
// and treats a clean exit as success. The IssueURL takes precedence over
// IssueNumber when both are set, because `gh` can derive the repo from a
// full URL and we avoid a stale-repo edge case.
func (w *zehnGitHubArtifactWriter) CreateComment(ctx context.Context, req integrationtools.GitHubCommentRequest) error {
	if w == nil {
		return errors.New("zehn github artifact writer is nil")
	}
	if req.IssueURL == "" && req.IssueNumber <= 0 {
		return errors.New("github comment requires issue number or URL")
	}

	var args []string
	if req.IssueURL != "" {
		args = []string{"issue", "comment", req.IssueURL, "--body", req.Body}
	} else {
		args = []string{"issue", "comment", strconv.Itoa(req.IssueNumber), "--repo", w.defaultRepo, "--body", req.Body}
	}

	cctx, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()

	if _, err := w.execCmd(cctx, "gh", args...); err != nil {
		return fmt.Errorf("gh issue comment: %w", err)
	}
	return nil
}

// parseIssueNumberFromURL extracts the trailing integer from a canonical
// GitHub issue URL such as "https://github.com/owner/repo/issues/123".
// Returns 0 when the URL is malformed or the trailing segment is not an int.
func parseIssueNumberFromURL(url string) int {
	trimmed := strings.TrimRight(strings.TrimSpace(url), "/")
	idx := strings.LastIndex(trimmed, "/")
	if idx < 0 || idx == len(trimmed)-1 {
		return 0
	}
	n, err := strconv.Atoi(trimmed[idx+1:])
	if err != nil {
		return 0
	}
	return n
}
