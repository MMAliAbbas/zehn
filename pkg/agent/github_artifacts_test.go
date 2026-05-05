package agent

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/providers"
	"github.com/sipeed/picoclaw/pkg/tools"
	integrationtools "github.com/sipeed/picoclaw/pkg/tools/integration"
)

type recordingGitHubMeetingProvider struct {
	mu sync.Mutex
}

func (p *recordingGitHubMeetingProvider) Chat(
	ctx context.Context,
	messages []providers.Message,
	tools []providers.ToolDefinition,
	model string,
	opts map[string]any,
) (*providers.LLMResponse, error) {
	content := strings.Join(messageContents(messages), "\n")
	switch {
	case strings.Contains(content, "Target agent: cmo"):
		return &providers.LLMResponse{Content: "Position: Launch a focused campaign.\nRaw internal transcript: SECRET_RAW_DEBATE", FinishReason: "stop"}, nil
	case strings.Contains(content, "Target agent: cfo"):
		return &providers.LLMResponse{Content: "Risk: margin pressure.\nDependency: CFO approval.\nAcceptance criteria: daily margin review.\nRaw internal transcript: SECRET_RAW_DEBATE", FinishReason: "stop"}, nil
	case strings.Contains(content, "Consolidate this chaired meeting"):
		return &providers.LLMResponse{Content: "Recommendation: run a tracked sales sprint.\nTimeline: two weeks.\nRisks: margin pressure.\nFollow-ups: CMO owns campaign; CFO approves discount.", FinishReason: "stop"}, nil
	default:
		return &providers.LLMResponse{Content: "advisory response", FinishReason: "stop"}, nil
	}
}

func (p *recordingGitHubMeetingProvider) GetDefaultModel() string {
	return "provider-default"
}

type advisoryGitHubMeetingProvider struct{}

func (p *advisoryGitHubMeetingProvider) Chat(
	ctx context.Context,
	messages []providers.Message,
	tools []providers.ToolDefinition,
	model string,
	opts map[string]any,
) (*providers.LLMResponse, error) {
	content := strings.Join(messageContents(messages), "\n")
	if strings.Contains(content, "Consolidate this chaired meeting") {
		return &providers.LLMResponse{Content: "Recommendation: keep monitoring the situation.\nRisks: none.", FinishReason: "stop"}, nil
	}
	return &providers.LLMResponse{Content: "Position: no execution needed.", FinishReason: "stop"}, nil
}

func (p *advisoryGitHubMeetingProvider) GetDefaultModel() string {
	return "provider-default"
}

type fakeGitHubArtifactWriter struct {
	mu         sync.Mutex
	fail       bool
	issues     []integrationtools.GitHubIssueRequest
	comments   []integrationtools.GitHubCommentRequest
	nextNumber int
}

func (w *fakeGitHubArtifactWriter) CreateIssue(ctx context.Context, req integrationtools.GitHubIssueRequest) (integrationtools.GitHubIssueArtifact, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.fail {
		return integrationtools.GitHubIssueArtifact{}, errors.New("github unavailable")
	}
	w.nextNumber++
	w.issues = append(w.issues, req)
	return integrationtools.GitHubIssueArtifact{
		Number: w.nextNumber,
		URL:    "https://github.example.test/org/repo/issues/" + strings.TrimSpace(req.SourceID),
	}, nil
}

func (w *fakeGitHubArtifactWriter) CreateComment(ctx context.Context, req integrationtools.GitHubCommentRequest) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.fail {
		return errors.New("github unavailable")
	}
	w.comments = append(w.comments, req)
	return nil
}

func (w *fakeGitHubArtifactWriter) snapshot() ([]integrationtools.GitHubIssueRequest, []integrationtools.GitHubCommentRequest) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return append([]integrationtools.GitHubIssueRequest(nil), w.issues...),
		append([]integrationtools.GitHubCommentRequest(nil), w.comments...)
}

func TestRunAgentMeetingGitHubSkipsIssueWithoutExecutableWork(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "ceo", Subagents: &config.SubagentsConfig{AllowAgents: []string{"cro"}}},
		{ID: "cro", Subagents: &config.SubagentsConfig{AllowAgents: []string{"cmo"}}},
		{ID: "cmo"},
	})
	al := NewAgentLoop(cfg, bus.NewMessageBus(), &advisoryGitHubMeetingProvider{})
	writer := &fakeGitHubArtifactWriter{}
	al.SetGitHubArtifactWriter(writer)

	result, err := al.StartAgentMeeting(context.Background(), tools.MeetingExecutionRequest{
		Title:               "Market pulse",
		SponsorAgentID:      "ceo",
		ChairAgentID:        "cro",
		ParticipantAgentIDs: []string{"cmo"},
		Goal:                "Decide whether a campaign needs execution.",
	})
	if err != nil {
		t.Fatalf("StartAgentMeeting() error = %v", err)
	}

	issues, comments := writer.snapshot()
	if len(issues) != 0 || len(comments) != 0 {
		t.Fatalf("GitHub artifacts = %d issues/%d comments, want none", len(issues), len(comments))
	}
	rec, err := al.meetingRecords.Get(context.Background(), result.MeetingID)
	if err != nil {
		t.Fatalf("meetingRecords.Get() error = %v", err)
	}
	if rec.Status != AgentMeetingStatusCompleted {
		t.Fatalf("meeting status = %q, want completed", rec.Status)
	}
	if rec.GitHubArtifact != nil {
		t.Fatalf("GitHubArtifact = %#v, want nil", rec.GitHubArtifact)
	}
}

func TestRunAgentMeetingGitHubCreatesIssueAndCuratedParticipantComments(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "ceo", Subagents: &config.SubagentsConfig{AllowAgents: []string{"cro"}}},
		{ID: "cro", Subagents: &config.SubagentsConfig{AllowAgents: []string{"cmo", "cfo"}}},
		{ID: "cmo"},
		{ID: "cfo"},
	})
	al := NewAgentLoop(cfg, bus.NewMessageBus(), &recordingGitHubMeetingProvider{})
	writer := &fakeGitHubArtifactWriter{}
	al.SetGitHubArtifactWriter(writer)

	result, err := al.StartAgentMeeting(context.Background(), tools.MeetingExecutionRequest{
		Title:               "Tracked sales sprint",
		SponsorAgentID:      "ceo",
		ChairAgentID:        "cro",
		ParticipantAgentIDs: []string{"cmo", "cfo"},
		Goal:                "Create executable work for a sales sprint.",
		Approvals:           []string{"approval required for discounting"},
	})
	if err != nil {
		t.Fatalf("StartAgentMeeting() error = %v", err)
	}

	issues, comments := writer.snapshot()
	if len(issues) != 1 {
		t.Fatalf("issues = %d, want 1", len(issues))
	}
	if !strings.Contains(issues[0].Title, "Tracked sales sprint") {
		t.Fatalf("issue title = %q, want meeting title", issues[0].Title)
	}
	for _, want := range []string{"run a tracked sales sprint", "CMO owns campaign", "GitHub is a tracker"} {
		if !strings.Contains(issues[0].Body, want) {
			t.Fatalf("issue body missing %q:\n%s", want, issues[0].Body)
		}
	}
	if strings.Contains(issues[0].Body, "SECRET_RAW_DEBATE") {
		t.Fatalf("issue body included raw transcript marker:\n%s", issues[0].Body)
	}
	if len(comments) != 2 {
		t.Fatalf("comments = %d, want 2", len(comments))
	}
	joinedComments := comments[0].Body + "\n" + comments[1].Body
	for _, want := range []string{"Position: Launch a focused campaign.", "Risk: margin pressure.", "Acceptance criteria: daily margin review."} {
		if !strings.Contains(joinedComments, want) {
			t.Fatalf("comments missing %q:\n%s", want, joinedComments)
		}
	}
	if strings.Contains(joinedComments, "SECRET_RAW_DEBATE") {
		t.Fatalf("comments included raw transcript marker:\n%s", joinedComments)
	}
	if len(result.ArtifactRefs) == 0 || !strings.Contains(strings.Join(result.ArtifactRefs, ","), "github.example.test") {
		t.Fatalf("ArtifactRefs = %v, want GitHub issue URL", result.ArtifactRefs)
	}
}

func TestRunAgentMeetingGitHubFailurePreservesMeetingRecord(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "ceo", Subagents: &config.SubagentsConfig{AllowAgents: []string{"cro"}}},
		{ID: "cro", Subagents: &config.SubagentsConfig{AllowAgents: []string{"cmo"}}},
		{ID: "cmo"},
	})
	al := NewAgentLoop(cfg, bus.NewMessageBus(), &recordingGitHubMeetingProvider{})
	al.SetGitHubArtifactWriter(&fakeGitHubArtifactWriter{fail: true})

	result, err := al.StartAgentMeeting(context.Background(), tools.MeetingExecutionRequest{
		Title:               "Tracked work with outage",
		SponsorAgentID:      "ceo",
		ChairAgentID:        "cro",
		ParticipantAgentIDs: []string{"cmo"},
		Goal:                "Create executable follow-up work.",
		Approvals:           []string{"approval required"},
	})
	if err != nil {
		t.Fatalf("StartAgentMeeting() error = %v", err)
	}
	rec, err := al.meetingRecords.Get(context.Background(), result.MeetingID)
	if err != nil {
		t.Fatalf("meetingRecords.Get() error = %v", err)
	}
	if rec.Status != AgentMeetingStatusCompleted {
		t.Fatalf("meeting status = %q, want completed", rec.Status)
	}
	if rec.GitHubArtifact == nil || rec.GitHubArtifact.Status != AgentGitHubArtifactStatusFailed {
		t.Fatalf("GitHubArtifact = %#v, want failed status", rec.GitHubArtifact)
	}
}

func TestRunAgentDelegationGitHubCreatesIssueForApprovalTrackedWork(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "parent", Subagents: &config.SubagentsConfig{AllowAgents: []string{"target"}}},
		{ID: "target"},
	})
	al := NewAgentLoop(cfg, bus.NewMessageBus(), &recordingDelegationProvider{})
	writer := &fakeGitHubArtifactWriter{}
	al.SetGitHubArtifactWriter(writer)

	result, err := al.RunAgentDelegation(context.Background(), AgentDelegationRequest{
		ParentAgentID:    "parent",
		TargetAgentID:    "target",
		Task:             "Prepare the approval package for launch.",
		ThreadKey:        "launch-approval",
		ApprovalRequired: true,
	})
	if err != nil {
		t.Fatalf("RunAgentDelegation() error = %v", err)
	}
	issues, comments := writer.snapshot()
	if len(issues) != 1 {
		t.Fatalf("issues = %d, want 1", len(issues))
	}
	if len(comments) != 0 {
		t.Fatalf("comments = %d, want no delegation comments", len(comments))
	}
	if !strings.Contains(issues[0].Body, "Approval required") {
		t.Fatalf("issue body missing approval marker:\n%s", issues[0].Body)
	}
	if len(result.ArtifactRefs) == 0 || !strings.Contains(strings.Join(result.ArtifactRefs, ","), "github.example.test") {
		t.Fatalf("ArtifactRefs = %v, want GitHub issue URL", result.ArtifactRefs)
	}
}

func TestRunAgentDelegationGitHubSkipsAdvisoryExchange(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "parent", Subagents: &config.SubagentsConfig{AllowAgents: []string{"target"}}},
		{ID: "target"},
	})
	al := NewAgentLoop(cfg, bus.NewMessageBus(), &recordingDelegationProvider{})
	writer := &fakeGitHubArtifactWriter{}
	al.SetGitHubArtifactWriter(writer)

	_, err := al.RunAgentDelegation(context.Background(), AgentDelegationRequest{
		ParentAgentID: "parent",
		TargetAgentID: "target",
		Task:          "Give concise advice on a decision.",
		ThreadKey:     "advice",
	})
	if err != nil {
		t.Fatalf("RunAgentDelegation() error = %v", err)
	}
	issues, comments := writer.snapshot()
	if len(issues) != 0 || len(comments) != 0 {
		t.Fatalf("GitHub artifacts = %d issues/%d comments, want none", len(issues), len(comments))
	}
}
