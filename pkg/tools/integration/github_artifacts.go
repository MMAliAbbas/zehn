package integrationtools

import "context"

type GitHubArtifactWriter interface {
	CreateIssue(ctx context.Context, req GitHubIssueRequest) (GitHubIssueArtifact, error)
	CreateComment(ctx context.Context, req GitHubCommentRequest) error
}

type GitHubIssueRequest struct {
	SourceType string
	SourceID   string
	Title      string
	Body       string
	Labels     []string
	// Repo is an optional target repository in "owner/name" or bare "name"
	// form. When set, the writer routes the issue to this repo verbatim
	// (caller wins). When empty, the Zehn writer's resolver inspects the
	// title, body, and labels to pick a target repo from the live
	// /Users/aliai/logicigniter/ inventory, falling back to a default.
	// Field is optional so existing callers compile unchanged.
	Repo string
}

type GitHubIssueArtifact struct {
	Number int
	URL    string
}

type GitHubCommentRequest struct {
	IssueNumber   int
	IssueURL      string
	SourceType    string
	SourceID      string
	AuthorAgentID string
	Body          string
}
