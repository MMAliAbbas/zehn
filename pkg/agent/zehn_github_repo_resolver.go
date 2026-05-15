// Zehn-fork-only file. See supervision/zehn_feature_tasks/task_057-github-artifact-repo-routing.md.
//
// Not part of upstream sipeed/picoclaw. File name carries the `zehn_` prefix
// so it is visible during upstream sync and cannot collide with future
// upstream additions.

package agent

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	integrationtools "github.com/sipeed/picoclaw/pkg/tools/integration"
)

const (
	defaultLogicIgniterRoot = "/Users/aliai/logicigniter"
	defaultLogicIgniterOwner = "logicigniter"
)

// zehnGitHubRepoResolver picks a target GitHub repo for a delegation- or
// meeting-derived issue. Resolution order:
//
//  1. req.Repo (caller wins; passed through verbatim after normalization to
//     `owner/name`).
//  2. `Target repo: <name>` line in req.Body (agents can explicitly steer).
//  3. Longest exact-name match of a known LogicIgniter repo in req.Title.
//  4. Longest exact-name match of a known LogicIgniter repo in req.Body.
//  5. area:* label mapping (area:frontend → svc-logicigniter-web, etc.).
//  6. defaultRepo.
//
// "Known repo" is discovered once per resolver instance by listing entries
// under liRoot and keeping the ones that contain a real `.git` directory.
// This filter excludes worktree clones such as `svc-logicigniter-web-issue83`
// (which has `.git` as a file pointing at the parent worktree, not a dir)
// and dotfile evidence directories.
type zehnGitHubRepoResolver struct {
	liRoot      string
	defaultRepo string
	owner       string

	once       sync.Once
	mu         sync.Mutex
	knownRepos []string
}

func newZehnGitHubRepoResolver(liRoot, defaultRepo string) *zehnGitHubRepoResolver {
	if liRoot == "" {
		liRoot = defaultLogicIgniterRoot
	}
	if defaultRepo == "" {
		defaultRepo = defaultLogicIgniterOwner + "/supervision"
	}
	owner := defaultLogicIgniterOwner
	if i := strings.Index(defaultRepo, "/"); i > 0 {
		owner = defaultRepo[:i]
	}
	return &zehnGitHubRepoResolver{
		liRoot:      liRoot,
		defaultRepo: defaultRepo,
		owner:       owner,
	}
}

func (r *zehnGitHubRepoResolver) resolveRepo(req integrationtools.GitHubIssueRequest) string {
	if r == nil {
		return ""
	}
	if explicit := strings.TrimSpace(req.Repo); explicit != "" {
		return r.normalize(explicit)
	}
	if marked := extractRepoFromBodyMarker(req.Body); marked != "" {
		return r.normalize(marked)
	}

	known := r.loadKnownRepos()
	if name := matchKnownRepoName(req.Title, known); name != "" {
		return r.owner + "/" + name
	}
	if name := matchKnownRepoName(req.Body, known); name != "" {
		return r.owner + "/" + name
	}
	if name := mapAreaLabelToRepo(req.Labels, known); name != "" {
		return r.owner + "/" + name
	}
	return r.defaultRepo
}

func (r *zehnGitHubRepoResolver) normalize(repo string) string {
	repo = strings.TrimSpace(repo)
	if strings.Contains(repo, "/") {
		return repo
	}
	return r.owner + "/" + repo
}

func (r *zehnGitHubRepoResolver) loadKnownRepos() []string {
	r.once.Do(func() {
		repos := discoverLogicIgniterRepos(r.liRoot)
		r.mu.Lock()
		r.knownRepos = repos
		r.mu.Unlock()
	})
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.knownRepos...)
}

// discoverLogicIgniterRepos returns LI subdirectory names that look like real
// git repos (have a `.git` *directory*). Worktree clones whose `.git` is a
// file are intentionally skipped — they are not standalone repos and writing
// an issue against their basename would 404 on github.com.
func discoverLogicIgniterRepos(liRoot string) []string {
	entries, err := os.ReadDir(liRoot)
	if err != nil {
		return nil
	}
	var repos []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") || !e.IsDir() {
			continue
		}
		gitDir := filepath.Join(liRoot, name, ".git")
		st, err := os.Stat(gitDir)
		if err != nil || !st.IsDir() {
			continue
		}
		repos = append(repos, name)
	}
	return repos
}

var bodyRepoMarkerRE = regexp.MustCompile(`(?m)^Target repo:\s*([^\s]+)`)

func extractRepoFromBodyMarker(body string) string {
	m := bodyRepoMarkerRE.FindStringSubmatch(body)
	if len(m) == 2 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

// matchKnownRepoName returns the longest known repo name that appears as a
// whole word in text. Word boundaries are non-alphanumeric/non-underscore
// characters; hyphens are explicitly non-word so `svc-contentaudit-grpc`
// matches cleanly inside prose like "fix svc-contentaudit-grpc tests".
func matchKnownRepoName(text string, known []string) string {
	if text == "" || len(known) == 0 {
		return ""
	}
	var best string
	for _, name := range known {
		if name == "" {
			continue
		}
		if !wordBoundaryContains(text, name) {
			continue
		}
		if len(name) > len(best) {
			best = name
		}
	}
	return best
}

func wordBoundaryContains(text, needle string) bool {
	idx := 0
	for {
		rel := strings.Index(text[idx:], needle)
		if rel < 0 {
			return false
		}
		start := idx + rel
		end := start + len(needle)
		leftOK := start == 0 || !isWordByte(text[start-1])
		rightOK := end == len(text) || !isWordByte(text[end])
		if leftOK && rightOK {
			return true
		}
		idx = start + 1
	}
}

func isWordByte(b byte) bool {
	return b == '_' ||
		(b >= 'A' && b <= 'Z') ||
		(b >= 'a' && b <= 'z') ||
		(b >= '0' && b <= '9')
}

// mapAreaLabelToRepo translates an area:* label to a known LI repo when the
// mapping has a target and the target actually exists in the discovered set.
// Unknown areas (`area:backend`, `area:product`, `area:architecture`) have
// no single-repo target and are intentionally left for fall-through to the
// default repo.
func mapAreaLabelToRepo(labels []string, known []string) string {
	knownSet := make(map[string]struct{}, len(known))
	for _, k := range known {
		knownSet[k] = struct{}{}
	}
	areaMap := map[string]string{
		"area:frontend":    "svc-logicigniter-web",
		"area:portal":      "svc-logicigniter-portal",
		"area:supervision": "supervision",
		"area:integration": "integration_tests",
		"area:devops":      "operations",
		"area:ops":         "operations",
		"area:operations":  "operations",
		"area:docs":        "supervision",
		"area:proto":       "proto",
		"area:infra":       "infra",
		"area:business":    "business",
		"area:shared-ui":   "li-shared-ui",
	}
	for _, l := range labels {
		key := strings.ToLower(strings.TrimSpace(l))
		repo, ok := areaMap[key]
		if !ok {
			continue
		}
		if _, exists := knownSet[repo]; exists {
			return repo
		}
	}
	return ""
}
