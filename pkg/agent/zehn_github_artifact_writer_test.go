// Zehn-fork-only file. See supervision/zehn_feature_tasks/task_056-github-artifact-writer-implementation.md.

package agent

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	integrationtools "github.com/sipeed/picoclaw/pkg/tools/integration"
)

func TestZehnGitHubArtifactWriter_CreateIssue_HappyPath(t *testing.T) {
	var capturedName string
	var capturedArgs []string
	w := newZehnGitHubArtifactWriter("logicigniter/test", time.Second)
	// Use an empty LI root so the resolver's discovery returns no known
	// repos and falls back to the default repo. Keeps this test
	// hermetic across machines.
	w.resolver = newZehnGitHubRepoResolver(t.TempDir(), "logicigniter/test")
	w.execCmd = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		capturedName = name
		capturedArgs = append([]string(nil), args...)
		return []byte("https://github.com/logicigniter/test/issues/42\n"), nil
	}

	art, err := w.CreateIssue(context.Background(), integrationtools.GitHubIssueRequest{
		SourceType: "delegation",
		SourceID:   "delegation-xyz",
		Title:      "Test issue",
		Body:       "Body content",
		Labels:     []string{"zehn:ready", "area:backend", "  ", "risk:medium"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedName != "gh" {
		t.Errorf("expected command name gh, got %q", capturedName)
	}
	want := []string{
		"issue", "create",
		"--repo", "logicigniter/test",
		"--title", "Test issue",
		"--body", "Body content",
		"--label", "zehn:ready",
		"--label", "area:backend",
		"--label", "risk:medium",
	}
	if !reflect.DeepEqual(capturedArgs, want) {
		t.Errorf("args mismatch:\n got:  %v\n want: %v", capturedArgs, want)
	}
	if art.Number != 42 {
		t.Errorf("expected issue number 42, got %d", art.Number)
	}
	if art.URL != "https://github.com/logicigniter/test/issues/42" {
		t.Errorf("URL mismatch: got %q", art.URL)
	}
}

func TestZehnGitHubArtifactWriter_CreateIssue_EmptyTitleRejected(t *testing.T) {
	w := newZehnGitHubArtifactWriter("logicigniter/test", time.Second)
	w.execCmd = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		t.Fatalf("execCmd should not be called when title is empty")
		return nil, nil
	}
	for _, title := range []string{"", "   ", "\t\n"} {
		_, err := w.CreateIssue(context.Background(), integrationtools.GitHubIssueRequest{
			Title: title,
			Body:  "body",
		})
		if err == nil {
			t.Errorf("expected error for empty title %q", title)
		}
	}
}

func TestZehnGitHubArtifactWriter_CreateIssue_PropagatesGhError(t *testing.T) {
	w := newZehnGitHubArtifactWriter("logicigniter/test", time.Second)
	ghErr := errors.New("gh: not authenticated")
	w.execCmd = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return nil, ghErr
	}
	_, err := w.CreateIssue(context.Background(), integrationtools.GitHubIssueRequest{
		Title: "t", Body: "b",
	})
	if err == nil {
		t.Fatal("expected error from gh")
	}
	if !strings.Contains(err.Error(), "gh: not authenticated") {
		t.Errorf("error should wrap gh stderr, got %v", err)
	}
}

func TestZehnGitHubArtifactWriter_CreateIssue_EmptyGhOutputRejected(t *testing.T) {
	w := newZehnGitHubArtifactWriter("logicigniter/test", time.Second)
	w.execCmd = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return []byte("   \n"), nil
	}
	_, err := w.CreateIssue(context.Background(), integrationtools.GitHubIssueRequest{
		Title: "t", Body: "b",
	})
	if err == nil {
		t.Fatal("expected error when gh returns empty output")
	}
}

func TestZehnGitHubArtifactWriter_CreateComment_HappyPath_URL(t *testing.T) {
	var capturedArgs []string
	w := newZehnGitHubArtifactWriter("logicigniter/test", time.Second)
	w.execCmd = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		capturedArgs = append([]string(nil), args...)
		return nil, nil
	}
	err := w.CreateComment(context.Background(), integrationtools.GitHubCommentRequest{
		IssueURL: "https://github.com/logicigniter/test/issues/42",
		Body:     "comment body",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"issue", "comment", "https://github.com/logicigniter/test/issues/42", "--body", "comment body"}
	if !reflect.DeepEqual(capturedArgs, want) {
		t.Errorf("args mismatch:\n got:  %v\n want: %v", capturedArgs, want)
	}
}

func TestZehnGitHubArtifactWriter_CreateComment_HappyPath_NumberAndRepo(t *testing.T) {
	var capturedArgs []string
	w := newZehnGitHubArtifactWriter("logicigniter/test", time.Second)
	w.execCmd = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		capturedArgs = append([]string(nil), args...)
		return nil, nil
	}
	err := w.CreateComment(context.Background(), integrationtools.GitHubCommentRequest{
		IssueNumber: 7,
		Body:        "ack",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"issue", "comment", "7", "--repo", "logicigniter/test", "--body", "ack"}
	if !reflect.DeepEqual(capturedArgs, want) {
		t.Errorf("args mismatch:\n got:  %v\n want: %v", capturedArgs, want)
	}
}

func TestZehnGitHubArtifactWriter_CreateComment_RejectsMissingTarget(t *testing.T) {
	w := newZehnGitHubArtifactWriter("logicigniter/test", time.Second)
	w.execCmd = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		t.Fatalf("execCmd should not be called when both number and URL are missing")
		return nil, nil
	}
	err := w.CreateComment(context.Background(), integrationtools.GitHubCommentRequest{Body: "x"})
	if err == nil {
		t.Fatal("expected error when no issue target is provided")
	}
}

func TestZehnGitHubArtifactWriter_CreateIssue_RoutesToServiceRepoFromTitle(t *testing.T) {
	// End-to-end check that resolver-driven repo routing reaches the gh
	// command line for an issue whose title names a real LI service.
	var capturedArgs []string
	liRoot := setupFakeLIRoot(t, []string{"svc-contentaudit-grpc", "supervision"}, nil)

	w := newZehnGitHubArtifactWriter("logicigniter/supervision", time.Second)
	w.resolver = newZehnGitHubRepoResolver(liRoot, "logicigniter/supervision")
	w.execCmd = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		capturedArgs = append([]string(nil), args...)
		return []byte("https://github.com/logicigniter/svc-contentaudit-grpc/issues/7\n"), nil
	}

	if _, err := w.CreateIssue(context.Background(), integrationtools.GitHubIssueRequest{
		Title: "Stage 1 -> 2: complete svc-contentaudit-grpc proto",
		Body:  "Per CTO ladder sweep recommendation.",
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Args should carry --repo logicigniter/svc-contentaudit-grpc, NOT the
	// default supervision repo.
	wantRepoArg := "logicigniter/svc-contentaudit-grpc"
	foundRepoArg := false
	for i, a := range capturedArgs {
		if a == "--repo" && i+1 < len(capturedArgs) && capturedArgs[i+1] == wantRepoArg {
			foundRepoArg = true
			break
		}
	}
	if !foundRepoArg {
		t.Errorf("expected --repo %s in args, got %v", wantRepoArg, capturedArgs)
	}
}

func TestZehnGitHubArtifactWriter_CreateIssue_ExplicitRepoOverridesResolver(t *testing.T) {
	var capturedArgs []string
	liRoot := setupFakeLIRoot(t, []string{"svc-contentaudit-grpc"}, nil)
	w := newZehnGitHubArtifactWriter("logicigniter/supervision", time.Second)
	w.resolver = newZehnGitHubRepoResolver(liRoot, "logicigniter/supervision")
	w.execCmd = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		capturedArgs = append([]string(nil), args...)
		return []byte("https://github.com/logicigniter/svc-billing/issues/1\n"), nil
	}
	if _, err := w.CreateIssue(context.Background(), integrationtools.GitHubIssueRequest{
		Title: "fix svc-contentaudit-grpc",
		Repo:  "logicigniter/svc-billing", // caller wins
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, a := range capturedArgs {
		if a == "--repo" && i+1 < len(capturedArgs) {
			if capturedArgs[i+1] != "logicigniter/svc-billing" {
				t.Errorf("expected explicit repo to win, got %s", capturedArgs[i+1])
			}
			return
		}
	}
	t.Errorf("--repo flag missing from %v", capturedArgs)
}

func TestZehnGitHubArtifactWriter_DefaultsApplied(t *testing.T) {
	w := newZehnGitHubArtifactWriter("", 0)
	if w.defaultRepo != defaultZehnGitHubArtifactRepo {
		t.Errorf("empty repo should default to %q, got %q", defaultZehnGitHubArtifactRepo, w.defaultRepo)
	}
	if w.timeout != defaultZehnGitHubArtifactTimeout {
		t.Errorf("zero timeout should default to %v, got %v", defaultZehnGitHubArtifactTimeout, w.timeout)
	}
}

func TestParseIssueNumberFromURL(t *testing.T) {
	cases := map[string]int{
		"https://github.com/logicigniter/supervision/issues/123":  123,
		"https://github.com/foo/bar/issues/1":                     1,
		"https://github.com/foo/bar/issues/1/":                    1,
		"  https://github.com/foo/bar/issues/9999\n":              9999,
		"https://github.com/foo/bar/issues/":                      0,
		"":                                                        0,
		"not a url":                                               0,
		"https://github.com/foo/bar/issues/notanumber":            0,
	}
	for url, want := range cases {
		if got := parseIssueNumberFromURL(url); got != want {
			t.Errorf("parseIssueNumberFromURL(%q) = %d, want %d", url, got, want)
		}
	}
}
