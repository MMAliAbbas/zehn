// Zehn-fork-only file. See supervision/zehn_feature_tasks/task_057-github-artifact-repo-routing.md.

package agent

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	integrationtools "github.com/sipeed/picoclaw/pkg/tools/integration"
)

func setupFakeLIRoot(t *testing.T, repoNames, nonRepoNames []string) string {
	t.Helper()
	root := t.TempDir()
	for _, name := range repoNames {
		if err := os.MkdirAll(filepath.Join(root, name, ".git"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range nonRepoNames {
		if err := os.MkdirAll(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestRepoResolver_DiscoverFiltersWorktreesAndDotfiles(t *testing.T) {
	root := setupFakeLIRoot(t,
		[]string{"svc-billing", "supervision", "svc-logicigniter-web", "li-shared-ui"},
		[]string{"svc-logicigniter-web-issue83", ".post-merge-evidence", ".worktrees", "regular-non-git-dir"},
	)
	repos := discoverLogicIgniterRepos(root)
	sort.Strings(repos)
	want := []string{"li-shared-ui", "supervision", "svc-billing", "svc-logicigniter-web"}
	if !reflect.DeepEqual(repos, want) {
		t.Errorf("discoverLogicIgniterRepos:\n got:  %v\n want: %v", repos, want)
	}
}

func TestRepoResolver_DiscoverGitAsFileSkipped(t *testing.T) {
	// Worktree clones have .git as a file pointing at the main worktree's
	// gitdir, not as a directory. Confirm we filter those out.
	root := setupFakeLIRoot(t, []string{"real-repo"}, nil)
	worktreeDir := filepath.Join(root, "worktree-clone")
	if err := os.MkdirAll(worktreeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(worktreeDir, ".git"), []byte("gitdir: ../real-repo/.git/worktrees/clone\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	repos := discoverLogicIgniterRepos(root)
	sort.Strings(repos)
	want := []string{"real-repo"}
	if !reflect.DeepEqual(repos, want) {
		t.Errorf("got %v, want %v", repos, want)
	}
}

func TestRepoResolver_ExplicitRepoWins(t *testing.T) {
	r := newZehnGitHubRepoResolver(setupFakeLIRoot(t, []string{"svc-contentaudit-grpc"}, nil), "logicigniter/supervision")
	got := r.resolveRepo(integrationtools.GitHubIssueRequest{
		Repo:  "logicigniter/svc-billing",
		Title: "would otherwise match svc-contentaudit-grpc",
	})
	if got != "logicigniter/svc-billing" {
		t.Errorf("explicit Repo should win, got %s", got)
	}
}

func TestRepoResolver_BareNameNormalizedToOwner(t *testing.T) {
	r := newZehnGitHubRepoResolver(setupFakeLIRoot(t, nil, nil), "logicigniter/supervision")
	got := r.resolveRepo(integrationtools.GitHubIssueRequest{Repo: "svc-billing"})
	if got != "logicigniter/svc-billing" {
		t.Errorf("bare name should normalize to owner/name, got %s", got)
	}
}

func TestRepoResolver_BodyMarkerExtracted(t *testing.T) {
	r := newZehnGitHubRepoResolver(setupFakeLIRoot(t, nil, nil), "logicigniter/supervision")
	got := r.resolveRepo(integrationtools.GitHubIssueRequest{
		Body: "do the thing\nTarget repo: logicigniter/svc-billing\nmore body",
	})
	if got != "logicigniter/svc-billing" {
		t.Errorf("body marker should be honored, got %s", got)
	}
}

func TestRepoResolver_TitleMentionMatched(t *testing.T) {
	r := newZehnGitHubRepoResolver(
		setupFakeLIRoot(t, []string{"svc-contentaudit-grpc", "supervision", "svc-billing"}, nil),
		"logicigniter/supervision",
	)
	got := r.resolveRepo(integrationtools.GitHubIssueRequest{
		Title: "Stage 1 -> 2: complete svc-contentaudit-grpc proto and handlers",
	})
	if got != "logicigniter/svc-contentaudit-grpc" {
		t.Errorf("got %s", got)
	}
}

func TestRepoResolver_BodyMentionMatched(t *testing.T) {
	r := newZehnGitHubRepoResolver(
		setupFakeLIRoot(t, []string{"svc-mainttriage-grpc", "supervision"}, nil),
		"logicigniter/supervision",
	)
	got := r.resolveRepo(integrationtools.GitHubIssueRequest{
		Title: "Backend task",
		Body:  "Implement the handler in svc-mainttriage-grpc package and add tests.",
	})
	if got != "logicigniter/svc-mainttriage-grpc" {
		t.Errorf("got %s", got)
	}
}

func TestRepoResolver_TitleBeatsBody(t *testing.T) {
	r := newZehnGitHubRepoResolver(
		setupFakeLIRoot(t, []string{"svc-contentaudit-grpc", "svc-billing"}, nil),
		"logicigniter/supervision",
	)
	got := r.resolveRepo(integrationtools.GitHubIssueRequest{
		Title: "fix svc-billing",
		Body:  "background reference to svc-contentaudit-grpc",
	})
	if got != "logicigniter/svc-billing" {
		t.Errorf("title match should take precedence over body, got %s", got)
	}
}

func TestRepoResolver_LongestMatchWins(t *testing.T) {
	r := newZehnGitHubRepoResolver(
		setupFakeLIRoot(t, []string{"svc-billing", "svc-billing-grpc"}, nil),
		"logicigniter/supervision",
	)
	got := r.resolveRepo(integrationtools.GitHubIssueRequest{
		Title: "fix svc-billing-grpc",
	})
	if got != "logicigniter/svc-billing-grpc" {
		t.Errorf("expected longer name to win, got %s", got)
	}
}

func TestRepoResolver_WordBoundary_NoSubstringHit(t *testing.T) {
	r := newZehnGitHubRepoResolver(
		setupFakeLIRoot(t, []string{"svc-bill"}, nil),
		"logicigniter/supervision",
	)
	got := r.resolveRepo(integrationtools.GitHubIssueRequest{
		Title: "fix svc-billing-grpc",
	})
	// svc-bill is a substring of svc-billing-grpc but the trailing 'i' is
	// a word char; word-boundary match must reject this.
	if got != "logicigniter/supervision" {
		t.Errorf("substring should not match across word boundary, got %s", got)
	}
}

func TestRepoResolver_AreaLabelMapping(t *testing.T) {
	r := newZehnGitHubRepoResolver(
		setupFakeLIRoot(t, []string{"svc-logicigniter-web", "supervision"}, nil),
		"logicigniter/supervision",
	)
	got := r.resolveRepo(integrationtools.GitHubIssueRequest{
		Title:  "hook BFF",
		Labels: []string{"zehn:ready", "Area:Frontend"},
	})
	if got != "logicigniter/svc-logicigniter-web" {
		t.Errorf("expected svc-logicigniter-web from area:frontend, got %s", got)
	}
}

func TestRepoResolver_AreaLabelOnlyMatchesIfRepoExists(t *testing.T) {
	r := newZehnGitHubRepoResolver(
		setupFakeLIRoot(t, []string{"supervision"}, nil), // no svc-logicigniter-web
		"logicigniter/supervision",
	)
	got := r.resolveRepo(integrationtools.GitHubIssueRequest{
		Labels: []string{"area:frontend"},
	})
	if got != "logicigniter/supervision" {
		t.Errorf("missing target repo should not be selected; got %s", got)
	}
}

func TestRepoResolver_UnmappedAreaLabelFallsThrough(t *testing.T) {
	r := newZehnGitHubRepoResolver(
		setupFakeLIRoot(t, []string{"supervision"}, nil),
		"logicigniter/supervision",
	)
	got := r.resolveRepo(integrationtools.GitHubIssueRequest{
		Labels: []string{"area:backend", "area:architecture", "area:product"},
	})
	if got != "logicigniter/supervision" {
		t.Errorf("unmapped area labels should fall through to default, got %s", got)
	}
}

func TestRepoResolver_FallsBackToDefault(t *testing.T) {
	r := newZehnGitHubRepoResolver(
		setupFakeLIRoot(t, []string{"supervision"}, nil),
		"logicigniter/supervision",
	)
	got := r.resolveRepo(integrationtools.GitHubIssueRequest{
		Title:  "generic title",
		Body:   "generic body",
		Labels: []string{"unrelated"},
	})
	if got != "logicigniter/supervision" {
		t.Errorf("got %s", got)
	}
}

func TestRepoResolver_MissingLIRootSurvives(t *testing.T) {
	r := newZehnGitHubRepoResolver("/nonexistent/path", "logicigniter/supervision")
	got := r.resolveRepo(integrationtools.GitHubIssueRequest{
		Title: "would otherwise match svc-billing",
	})
	if got != "logicigniter/supervision" {
		t.Errorf("missing LI root should not panic and should fall back, got %s", got)
	}
}

func TestRepoResolver_NilSafe(t *testing.T) {
	var r *zehnGitHubRepoResolver
	got := r.resolveRepo(integrationtools.GitHubIssueRequest{Title: "anything"})
	if got != "" {
		t.Errorf("nil resolver should return empty, got %s", got)
	}
}
