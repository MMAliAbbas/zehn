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
