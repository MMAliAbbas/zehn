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
