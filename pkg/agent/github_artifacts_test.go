package agent

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

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

type secretGitHubDelegationProvider struct {
	secret string
}

func (p *secretGitHubDelegationProvider) Chat(
	ctx context.Context,
	messages []providers.Message,
	tools []providers.ToolDefinition,
	model string,
	opts map[string]any,
) (*providers.LLMResponse, error) {
	return &providers.LLMResponse{Content: "completed result with " + p.secret, FinishReason: "stop"}, nil
}

func (p *secretGitHubDelegationProvider) GetDefaultModel() string {
	return "provider-default"
}

type secretGitHubMeetingProvider struct {
	secret string
}

func (p *secretGitHubMeetingProvider) Chat(
	ctx context.Context,
	messages []providers.Message,
	tools []providers.ToolDefinition,
	model string,
	opts map[string]any,
) (*providers.LLMResponse, error) {
	content := strings.Join(messageContents(messages), "\n")
	switch {
	case strings.Contains(content, "Target agent: cmo"):
		return &providers.LLMResponse{Content: "Position: campaign can use " + p.secret + "\nRisk: leak " + p.secret, FinishReason: "stop"}, nil
	case strings.Contains(content, "Consolidate this chaired meeting"):
		return &providers.LLMResponse{Content: "Recommendation: approve plan with " + p.secret + "\nTimeline: launch window uses " + p.secret + "\nRisks: margin exposure " + p.secret + "\nFollow-ups: CMO replaces " + p.secret, FinishReason: "stop"}, nil
	default:
		return &providers.LLMResponse{Content: "advisory response", FinishReason: "stop"}, nil
	}
}

func (p *secretGitHubMeetingProvider) GetDefaultModel() string {
	return "provider-default"
}

type fakeGitHubArtifactWriter struct {
	mu         sync.Mutex
	fail       bool
	issues     []integrationtools.GitHubIssueRequest
	comments   []integrationtools.GitHubCommentRequest
	nextNumber int
}

type blockingGitHubArtifactWriter struct {
	startOnce sync.Once
	started   chan struct{}
	release   chan struct{}
}

func (w *blockingGitHubArtifactWriter) CreateIssue(ctx context.Context, req integrationtools.GitHubIssueRequest) (integrationtools.GitHubIssueArtifact, error) {
	w.startOnce.Do(func() { close(w.started) })
	select {
	case <-w.release:
		return integrationtools.GitHubIssueArtifact{
			Number: 101,
			URL:    "https://github.example.test/org/repo/issues/" + strings.TrimSpace(req.SourceID),
		}, nil
	case <-ctx.Done():
		return integrationtools.GitHubIssueArtifact{}, ctx.Err()
	}
}

func (w *blockingGitHubArtifactWriter) CreateComment(ctx context.Context, req integrationtools.GitHubCommentRequest) error {
	return nil
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

func TestRunAgentDelegationGitHubPublishDoesNotBlockCompletedResult(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "parent", Subagents: &config.SubagentsConfig{AllowAgents: []string{"target"}}},
		{ID: "target"},
	})
	al := NewAgentLoop(cfg, bus.NewMessageBus(), &recordingDelegationProvider{})
	al.githubArtifactPublisher = newGitHubArtifactPublisher(1, time.Second)
	writer := &blockingGitHubArtifactWriter{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	al.SetGitHubArtifactWriter(writer)

	type delegationResult struct {
		result AgentDelegationResult
		err    error
	}
	done := make(chan delegationResult, 1)
	go func() {
		result, err := al.RunAgentDelegation(context.Background(), AgentDelegationRequest{
			ParentAgentID:    "parent",
			TargetAgentID:    "target",
			Task:             "Prepare the approval package for launch.",
			ThreadKey:        "launch-approval",
			ApprovalRequired: true,
		})
		done <- delegationResult{result: result, err: err}
	}()

	var completed delegationResult
	select {
	case got := <-done:
		completed = got
	case <-writer.started:
		select {
		case got := <-done:
			completed = got
		case <-time.After(500 * time.Millisecond):
			close(writer.release)
			got := <-done
			if got.err != nil {
				t.Fatalf("RunAgentDelegation() returned late with error = %v", got.err)
			}
			t.Fatalf("RunAgentDelegation() waited for blocked GitHub publishing")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("RunAgentDelegation() did not complete or start GitHub publishing")
	}
	if completed.err != nil {
		t.Fatalf("RunAgentDelegation() error = %v", completed.err)
	}
	if completed.result.DelegationID == "" {
		t.Fatal("DelegationID is empty")
	}
	rec, err := al.delegationRecords.Get(context.Background(), completed.result.DelegationID)
	if err != nil {
		t.Fatalf("delegationRecords.Get() error = %v", err)
	}
	if rec.Status != AgentDelegationStatusCompleted {
		t.Fatalf("delegation status = %q, want completed", rec.Status)
	}
	if rec.GitHubArtifact == nil || rec.GitHubArtifact.Status != AgentGitHubArtifactStatusPending {
		t.Fatalf("GitHubArtifact = %#v, want pending status while writer is blocked", rec.GitHubArtifact)
	}

	select {
	case <-writer.started:
	default:
	}
	close(writer.release)
	rec = waitForDelegationGitHubStatus(t, al, completed.result.DelegationID, AgentGitHubArtifactStatusCreated)
	if rec.GitHubArtifact.IssueURL == "" {
		t.Fatalf("GitHubArtifact = %#v, want issue URL", rec.GitHubArtifact)
	}
}

func TestRunAgentDelegationGitHubPublisherRecordsCapacityFailure(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "parent", Subagents: &config.SubagentsConfig{AllowAgents: []string{"target"}}},
		{ID: "target"},
	})
	al := NewAgentLoop(cfg, bus.NewMessageBus(), &recordingDelegationProvider{})
	al.githubArtifactPublisher = newGitHubArtifactPublisher(1, time.Second)
	writer := &blockingGitHubArtifactWriter{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	al.SetGitHubArtifactWriter(writer)

	first, err := al.RunAgentDelegation(context.Background(), AgentDelegationRequest{
		ParentAgentID:    "parent",
		TargetAgentID:    "target",
		Task:             "Prepare the first approval package.",
		ThreadKey:        "first",
		ApprovalRequired: true,
	})
	if err != nil {
		t.Fatalf("first RunAgentDelegation() error = %v", err)
	}
	<-writer.started
	waitForDelegationGitHubStatus(t, al, first.DelegationID, AgentGitHubArtifactStatusPending)

	second, err := al.RunAgentDelegation(context.Background(), AgentDelegationRequest{
		ParentAgentID:    "parent",
		TargetAgentID:    "target",
		Task:             "Prepare the second approval package.",
		ThreadKey:        "second",
		ApprovalRequired: true,
	})
	if err != nil {
		t.Fatalf("second RunAgentDelegation() error = %v", err)
	}
	rec := waitForDelegationGitHubStatus(t, al, second.DelegationID, AgentGitHubArtifactStatusFailed)
	if rec.GitHubArtifact == nil || !strings.Contains(rec.GitHubArtifact.Error, "capacity") {
		t.Fatalf("GitHubArtifact = %#v, want capacity failure", rec.GitHubArtifact)
	}

	close(writer.release)
	waitForDelegationGitHubStatus(t, al, first.DelegationID, AgentGitHubArtifactStatusCreated)
}

func TestRunAgentDelegationGitHubPublisherTimeoutRecordsFailure(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "parent", Subagents: &config.SubagentsConfig{AllowAgents: []string{"target"}}},
		{ID: "target"},
	})
	al := NewAgentLoop(cfg, bus.NewMessageBus(), &recordingDelegationProvider{})
	al.githubArtifactPublisher = newGitHubArtifactPublisher(1, 10*time.Millisecond)
	writer := &blockingGitHubArtifactWriter{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	al.SetGitHubArtifactWriter(writer)

	result, err := al.RunAgentDelegation(context.Background(), AgentDelegationRequest{
		ParentAgentID:    "parent",
		TargetAgentID:    "target",
		Task:             "Prepare an approval package that times out in GitHub.",
		ThreadKey:        "timeout",
		ApprovalRequired: true,
	})
	if err != nil {
		t.Fatalf("RunAgentDelegation() error = %v", err)
	}
	<-writer.started
	rec := waitForDelegationGitHubStatus(t, al, result.DelegationID, AgentGitHubArtifactStatusFailed)
	if rec.GitHubArtifact == nil || !strings.Contains(rec.GitHubArtifact.Error, context.DeadlineExceeded.Error()) {
		t.Fatalf("GitHubArtifact = %#v, want deadline exceeded failure", rec.GitHubArtifact)
	}
}

func TestAgentLoopCloseDrainsAcceptedGitHubPublisherJobs(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "parent", Subagents: &config.SubagentsConfig{AllowAgents: []string{"target"}}},
		{ID: "target"},
	})
	al := NewAgentLoop(cfg, bus.NewMessageBus(), &recordingDelegationProvider{})
	al.githubArtifactPublisher = newGitHubArtifactPublisher(1, time.Second)
	writer := &blockingGitHubArtifactWriter{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	al.SetGitHubArtifactWriter(writer)

	result, err := al.RunAgentDelegation(context.Background(), AgentDelegationRequest{
		ParentAgentID:    "parent",
		TargetAgentID:    "target",
		Task:             "Prepare an approval package before shutdown.",
		ThreadKey:        "shutdown",
		ApprovalRequired: true,
	})
	if err != nil {
		t.Fatalf("RunAgentDelegation() error = %v", err)
	}
	<-writer.started

	closed := make(chan struct{})
	go func() {
		al.Close()
		close(closed)
	}()

	select {
	case <-closed:
		t.Fatal("AgentLoop.Close() returned before accepted GitHub job finished")
	case <-time.After(25 * time.Millisecond):
	}

	close(writer.release)
	select {
	case <-closed:
	case <-time.After(2 * time.Second):
		t.Fatal("AgentLoop.Close() did not drain accepted GitHub job")
	}
	waitForDelegationGitHubStatus(t, al, result.DelegationID, AgentGitHubArtifactStatusCreated)
}

func TestGitHubPublisherCapacityIsIsolatedBetweenAgentLoops(t *testing.T) {
	cfg1 := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "parent", Subagents: &config.SubagentsConfig{AllowAgents: []string{"target"}}},
		{ID: "target"},
	})
	al1 := NewAgentLoop(cfg1, bus.NewMessageBus(), &recordingDelegationProvider{})
	al1.githubArtifactPublisher = newGitHubArtifactPublisher(1, time.Second)
	writer1 := &blockingGitHubArtifactWriter{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	al1.SetGitHubArtifactWriter(writer1)

	cfg2 := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "parent", Subagents: &config.SubagentsConfig{AllowAgents: []string{"target"}}},
		{ID: "target"},
	})
	al2 := NewAgentLoop(cfg2, bus.NewMessageBus(), &recordingDelegationProvider{})
	al2.githubArtifactPublisher = newGitHubArtifactPublisher(1, time.Second)
	writer2 := &fakeGitHubArtifactWriter{}
	al2.SetGitHubArtifactWriter(writer2)

	first, err := al1.RunAgentDelegation(context.Background(), AgentDelegationRequest{
		ParentAgentID:    "parent",
		TargetAgentID:    "target",
		Task:             "Prepare the blocked approval package.",
		ThreadKey:        "blocked",
		ApprovalRequired: true,
	})
	if err != nil {
		t.Fatalf("first RunAgentDelegation() error = %v", err)
	}
	<-writer1.started
	waitForDelegationGitHubStatus(t, al1, first.DelegationID, AgentGitHubArtifactStatusPending)

	second, err := al2.RunAgentDelegation(context.Background(), AgentDelegationRequest{
		ParentAgentID:    "parent",
		TargetAgentID:    "target",
		Task:             "Prepare the isolated approval package.",
		ThreadKey:        "isolated",
		ApprovalRequired: true,
	})
	if err != nil {
		t.Fatalf("second RunAgentDelegation() error = %v", err)
	}
	waitForDelegationGitHubStatus(t, al2, second.DelegationID, AgentGitHubArtifactStatusCreated)

	close(writer1.release)
	waitForDelegationGitHubStatus(t, al1, first.DelegationID, AgentGitHubArtifactStatusCreated)
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

	rec := waitForMeetingGitHubStatus(t, al, result.MeetingID, AgentGitHubArtifactStatusCreated)
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
	if len(rec.ArtifactRefs) == 0 || !strings.Contains(strings.Join(rec.ArtifactRefs, ","), "github.example.test") {
		t.Fatalf("record ArtifactRefs = %v, want GitHub issue URL", rec.ArtifactRefs)
	}
}

func TestRunAgentMeetingGitHubArtifactsUseRedactedMeetingRecord(t *testing.T) {
	secret := "ghp_fake_meeting_secret_12345"
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "ceo", Subagents: &config.SubagentsConfig{AllowAgents: []string{"cro"}}},
		{ID: "cro", Subagents: &config.SubagentsConfig{AllowAgents: []string{"cmo"}}},
		{ID: "cmo"},
	})
	cfg.ModelList = append(cfg.ModelList, &config.ModelConfig{
		ModelName: "test-model",
		APIKeys:   config.SecureStrings{config.NewSecureString(secret)},
	})
	al := NewAgentLoop(cfg, bus.NewMessageBus(), &secretGitHubMeetingProvider{secret: secret})
	writer := &fakeGitHubArtifactWriter{}
	al.SetGitHubArtifactWriter(writer)

	result, err := al.StartAgentMeeting(context.Background(), tools.MeetingExecutionRequest{
		Title:               "Tracked sales sprint " + secret,
		SponsorAgentID:      "ceo",
		ChairAgentID:        "cro",
		ParticipantAgentIDs: []string{"cmo"},
		Goal:                "Create executable work around " + secret,
		Approvals:           []string{"approval required for " + secret},
	})
	if err != nil {
		t.Fatalf("StartAgentMeeting() error = %v", err)
	}
	waitForMeetingGitHubStatus(t, al, result.MeetingID, AgentGitHubArtifactStatusCreated)

	issues, comments := writer.snapshot()
	if len(issues) != 1 {
		t.Fatalf("issues = %d, want 1", len(issues))
	}
	if len(comments) != 1 {
		t.Fatalf("comments = %d, want 1", len(comments))
	}
	artifactText := issues[0].Title + "\n" + issues[0].Body + "\n" + comments[0].Body
	assertGitHubArtifactRedacted(t, artifactText, secret)
	for _, want := range []string{"approve plan", "Timeline", "Risks", "Follow-ups", "Focused participant note"} {
		if !strings.Contains(artifactText, want) {
			t.Fatalf("artifact text missing %q:\n%s", want, artifactText)
		}
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
	rec := waitForMeetingGitHubStatus(t, al, result.MeetingID, AgentGitHubArtifactStatusFailed)
	if rec.Status != AgentMeetingStatusCompleted {
		t.Fatalf("meeting status = %q, want completed", rec.Status)
	}
}

func TestRunAgentMeetingGitHubDisabledWriterRecordsSkipped(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "ceo", Subagents: &config.SubagentsConfig{AllowAgents: []string{"cro"}}},
		{ID: "cro", Subagents: &config.SubagentsConfig{AllowAgents: []string{"cmo"}}},
		{ID: "cmo"},
	})
	al := NewAgentLoop(cfg, bus.NewMessageBus(), &recordingGitHubMeetingProvider{})

	result, err := al.StartAgentMeeting(context.Background(), tools.MeetingExecutionRequest{
		Title:               "Tracked work with disabled GitHub",
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
	if rec.GitHubArtifact == nil || rec.GitHubArtifact.Status != AgentGitHubArtifactStatusSkipped {
		t.Fatalf("GitHubArtifact = %#v, want skipped status", rec.GitHubArtifact)
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
	rec := waitForDelegationGitHubStatus(t, al, result.DelegationID, AgentGitHubArtifactStatusCreated)
	if rec.Result == nil || len(rec.Result.ArtifactRefs) == 0 || !strings.Contains(strings.Join(rec.Result.ArtifactRefs, ","), "github.example.test") {
		t.Fatalf("record result ArtifactRefs = %#v, want GitHub issue URL", rec.Result)
	}
}

func TestRunAgentDelegationGitHubArtifactUsesRedactedDelegationRecord(t *testing.T) {
	secret := "ghp_fake_delegation_secret_12345"
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "parent", Subagents: &config.SubagentsConfig{AllowAgents: []string{"target"}}},
		{ID: "target"},
	})
	cfg.ModelList = append(cfg.ModelList, &config.ModelConfig{
		ModelName: "test-model",
		APIKeys:   config.SecureStrings{config.NewSecureString(secret)},
	})
	al := NewAgentLoop(cfg, bus.NewMessageBus(), &secretGitHubDelegationProvider{secret: secret})
	writer := &fakeGitHubArtifactWriter{}
	al.SetGitHubArtifactWriter(writer)

	result, err := al.RunAgentDelegation(context.Background(), AgentDelegationRequest{
		ParentAgentID:    "parent",
		TargetAgentID:    "target",
		Task:             "prepare approval package with " + secret,
		ThreadKey:        "approval-thread-" + secret,
		Priority:         "high-" + secret,
		ApprovalRequired: true,
		ArtifactRefs:     []string{"input-ref-" + secret},
	})
	if err != nil {
		t.Fatalf("RunAgentDelegation() error = %v", err)
	}
	waitForDelegationGitHubStatus(t, al, result.DelegationID, AgentGitHubArtifactStatusCreated)

	issues, comments := writer.snapshot()
	if len(issues) != 1 {
		t.Fatalf("issues = %d, want 1", len(issues))
	}
	if len(comments) != 0 {
		t.Fatalf("comments = %d, want no delegation comments", len(comments))
	}
	artifactText := issues[0].Title + "\n" + issues[0].Body
	assertGitHubArtifactRedacted(t, artifactText, secret)
	for _, want := range []string{"Approval required", "Priority", "Current Result", "[FILTERED]"} {
		if !strings.Contains(artifactText, want) {
			t.Fatalf("artifact text missing %q:\n%s", want, artifactText)
		}
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

func assertGitHubArtifactRedacted(t *testing.T, artifactText, secret string) {
	t.Helper()
	if strings.Contains(artifactText, secret) {
		t.Fatalf("GitHub artifact leaked secret %q:\n%s", secret, artifactText)
	}
	if !strings.Contains(artifactText, "[FILTERED]") {
		t.Fatalf("GitHub artifact missing redaction placeholder:\n%s", artifactText)
	}
}

func waitForDelegationGitHubStatus(
	t *testing.T,
	al *AgentLoop,
	delegationID string,
	want AgentGitHubArtifactStatus,
) AgentDelegationRecord {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		rec, err := al.delegationRecords.Get(context.Background(), delegationID)
		if err != nil {
			t.Fatalf("delegationRecords.Get(%s) error = %v", delegationID, err)
		}
		if rec.GitHubArtifact != nil && rec.GitHubArtifact.Status == want {
			return rec
		}
		select {
		case <-deadline:
			t.Fatalf("delegation GitHub artifact = %#v, want %q", rec.GitHubArtifact, want)
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func waitForMeetingGitHubStatus(
	t *testing.T,
	al *AgentLoop,
	meetingID string,
	want AgentGitHubArtifactStatus,
) AgentMeetingRecord {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		rec, err := al.meetingRecords.Get(context.Background(), meetingID)
		if err != nil {
			t.Fatalf("meetingRecords.Get(%s) error = %v", meetingID, err)
		}
		if rec.GitHubArtifact != nil && rec.GitHubArtifact.Status == want {
			return rec
		}
		select {
		case <-deadline:
			t.Fatalf("meeting GitHub artifact = %#v, want %q", rec.GitHubArtifact, want)
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}
