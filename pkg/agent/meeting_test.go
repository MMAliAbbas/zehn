package agent

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/providers"
	"github.com/sipeed/picoclaw/pkg/tools"
)

type recordingMeetingProvider struct {
	mu    sync.Mutex
	tasks []string
}

func (p *recordingMeetingProvider) Chat(
	ctx context.Context,
	messages []providers.Message,
	tools []providers.ToolDefinition,
	model string,
	opts map[string]any,
) (*providers.LLMResponse, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	content := strings.Join(messageContents(messages), "\n")
	p.tasks = append(p.tasks, content)
	switch {
	case strings.Contains(content, "Target agent: cmo"):
		return &providers.LLMResponse{Content: "CMO: Use a focused campaign and existing testimonials.", FinishReason: "stop"}, nil
	case strings.Contains(content, "Target agent: cfo"):
		return &providers.LLMResponse{Content: "CFO: Keep discounting capped and review margin daily.", FinishReason: "stop"}, nil
	case strings.Contains(content, "Consolidate this chaired meeting"):
		return &providers.LLMResponse{Content: "Recommendation: run a two-week sales sprint.\nTimeline: daily pipeline review.\nRisks: margin pressure.\nFollow-ups: CRO owns execution.", FinishReason: "stop"}, nil
	default:
		return &providers.LLMResponse{Content: "unexpected meeting turn", FinishReason: "stop"}, nil
	}
}

func (p *recordingMeetingProvider) GetDefaultModel() string {
	return "provider-default"
}

func (p *recordingMeetingProvider) calls() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]string(nil), p.tasks...)
}

func TestRunAgentMeeting_DelegatesParticipantsAndPersistsConsolidatedRecommendation(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "ceo", Subagents: &config.SubagentsConfig{AllowAgents: []string{"cro"}}},
		{ID: "cro", Subagents: &config.SubagentsConfig{AllowAgents: []string{"cmo", "cfo"}}},
		{ID: "cmo"},
		{ID: "cfo"},
	})
	provider := &recordingMeetingProvider{}
	al := NewAgentLoop(cfg, bus.NewMessageBus(), provider)

	result, err := al.StartAgentMeeting(context.Background(), tools.MeetingExecutionRequest{
		Title:               "Two-week sales lift",
		SponsorAgentID:      "ceo",
		ChairAgentID:        "cro",
		ParticipantAgentIDs: []string{"cmo", "cfo"},
		Goal:                "Increase sales by 30% in two weeks.",
		Constraints:         []string{"No customer-facing commitments without approval."},
		Notes:               "Focus on existing pipeline.",
		ArtifactRefs:        []string{"objective:sales-lift"},
	})
	if err != nil {
		t.Fatalf("StartAgentMeeting() error = %v", err)
	}
	if result.MeetingID == "" {
		t.Fatal("MeetingID is empty")
	}
	if !strings.Contains(result.Recommendation, "two-week sales sprint") {
		t.Fatalf("Recommendation = %q", result.Recommendation)
	}
	if len(result.Timeline) == 0 {
		t.Fatal("Timeline should be populated")
	}
	if len(result.Risks) == 0 {
		t.Fatal("Risks should be populated")
	}
	if len(result.FollowUps) == 0 {
		t.Fatal("FollowUps should be populated")
	}

	rec, err := al.meetingRecords.Get(context.Background(), result.MeetingID)
	if err != nil {
		t.Fatalf("meetingRecords.Get() error = %v", err)
	}
	if rec.SponsorAgentID != "ceo" || rec.ChairAgentID != "cro" {
		t.Fatalf("record sponsor/chair = %q/%q, want ceo/cro", rec.SponsorAgentID, rec.ChairAgentID)
	}
	if got := strings.Join(rec.Participants, ","); got != "cmo,cfo" {
		t.Fatalf("Participants = %q, want cmo,cfo", got)
	}
	if len(rec.ParticipantTurns) != 2 {
		t.Fatalf("ParticipantTurns = %d, want 2", len(rec.ParticipantTurns))
	}
	if !strings.Contains(rec.ParticipantTurns[0].Response, "CMO:") {
		t.Fatalf("first participant response = %q", rec.ParticipantTurns[0].Response)
	}
	if !strings.Contains(rec.Recommendation, "two-week sales sprint") {
		t.Fatalf("record recommendation = %q", rec.Recommendation)
	}
	if len(rec.Timeline) == 0 || len(rec.Risks) == 0 || len(rec.FollowUps) == 0 {
		t.Fatalf("record missing timeline/risks/follow-ups: %#v", rec)
	}
	if len(rec.ArtifactRefs) != 1 || rec.ArtifactRefs[0] != "objective:sales-lift" {
		t.Fatalf("ArtifactRefs = %v", rec.ArtifactRefs)
	}

	data, err := os.ReadFile(filepath.Join(al.meetingRecords.dir, rec.Filename()))
	if err != nil {
		t.Fatalf("ReadFile(record) error = %v", err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal(record) error = %v", err)
	}
	if _, ok := raw["participant_turns"]; !ok {
		t.Fatalf("persisted record missing participant_turns: %s", data)
	}

	calls := provider.calls()
	if len(calls) != 3 {
		t.Fatalf("provider calls = %d, want 3", len(calls))
	}
	for _, want := range []string{"Target agent: cmo", "Target agent: cfo", "Consolidate this chaired meeting"} {
		if !strings.Contains(strings.Join(calls, "\n"), want) {
			t.Fatalf("provider calls missing %q:\n%s", want, strings.Join(calls, "\n---\n"))
		}
	}
}

func TestRunAgentMeeting_DepartmentHeadCanChairOwnDomainMeeting(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "cto", Subagents: &config.SubagentsConfig{AllowAgents: []string{"devops"}}},
		{ID: "devops"},
	})
	al := NewAgentLoop(cfg, bus.NewMessageBus(), &recordingMeetingProvider{})

	result, err := al.StartAgentMeeting(context.Background(), tools.MeetingExecutionRequest{
		Title:               "Deployment risk",
		SponsorAgentID:      "cto",
		ChairAgentID:        "cto",
		ParticipantAgentIDs: []string{"devops"},
		Goal:                "Review deployment risk.",
	})
	if err != nil {
		t.Fatalf("StartAgentMeeting() error = %v", err)
	}
	if result.MeetingID == "" {
		t.Fatal("MeetingID is empty")
	}
}

func TestRunAgentMeeting_PublishesEventBasedDiscordVisibilitySummaries(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "ceo", Subagents: &config.SubagentsConfig{AllowAgents: []string{"cro"}}},
		{ID: "cro", Subagents: &config.SubagentsConfig{AllowAgents: []string{"cmo", "cfo"}}},
		{ID: "cmo"},
		{ID: "cfo"},
	})
	cfg.Channels = discordVisibilityChannels(t, true)
	msgBus := bus.NewMessageBus()
	al := NewAgentLoop(cfg, msgBus, &recordingMeetingProvider{})
	al.SetGitHubArtifactWriter(&fakeGitHubArtifactWriter{})

	result, err := al.StartAgentMeeting(context.Background(), tools.MeetingExecutionRequest{
		Title:               "Two-week sales lift",
		SponsorAgentID:      "ceo",
		ChairAgentID:        "cro",
		ParticipantAgentIDs: []string{"cmo", "cfo"},
		Goal:                "Increase sales by 30% in two weeks.",
		Approvals:           []string{"Ali approval before customer-facing discounts."},
		Notes:               "PRIVATE_RAW_MEETING_NOTES must stay private.",
	})
	if err != nil {
		t.Fatalf("StartAgentMeeting() error = %v", err)
	}

	messages := collectOutboundMessages(t, msgBus, 11)
	contents := make([]string, 0, len(messages))
	for _, msg := range messages {
		contents = append(contents, msg.Content)
	}
	joined := strings.Join(contents, "\n")
	for _, want := range []string{
		"Meeting opened",
		"Delegation created",
		"Recommendation ready",
		"Approval needed",
		"Issue created",
		"Meeting completed",
		result.MeetingID,
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("visibility summaries missing %q:\n%s", want, joined)
		}
	}
	if strings.Contains(joined, "PRIVATE_RAW_MEETING_NOTES") || strings.Contains(joined, "CMO:") || strings.Contains(joined, "CFO:") {
		t.Fatalf("visibility summaries leaked internal meeting material:\n%s", joined)
	}
}

type failingMeetingProvider struct {
	err            error
	cancelAfterCMO context.CancelFunc
	cancelOnChair  context.CancelFunc
	calls          []string
	mu             sync.Mutex
}

func (p *failingMeetingProvider) Chat(
	ctx context.Context,
	messages []providers.Message,
	tools []providers.ToolDefinition,
	model string,
	opts map[string]any,
) (*providers.LLMResponse, error) {
	content := strings.Join(messageContents(messages), "\n")
	p.mu.Lock()
	p.calls = append(p.calls, content)
	p.mu.Unlock()
	switch {
	case strings.Contains(content, "Target agent: cmo"):
		if p.cancelAfterCMO != nil {
			p.cancelAfterCMO()
		}
		return &providers.LLMResponse{Content: "CMO: use customer proof.", FinishReason: "stop"}, nil
	case strings.Contains(content, "Target agent: cfo"):
		return nil, p.err
	case strings.Contains(content, "Consolidate this chaired meeting"):
		if p.cancelOnChair != nil {
			p.cancelOnChair()
			return &providers.LLMResponse{Content: "Recommendation: proceed.", FinishReason: "stop"}, nil
		}
		return nil, p.err
	default:
		return &providers.LLMResponse{Content: "unexpected meeting turn", FinishReason: "stop"}, nil
	}
}

func (p *failingMeetingProvider) GetDefaultModel() string {
	return "provider-default"
}

func (p *failingMeetingProvider) callText() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return strings.Join(p.calls, "\n---\n")
}

func TestRunAgentMeeting_ParticipantFailureStopsBeforeChairSynthesis(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "ceo", Subagents: &config.SubagentsConfig{AllowAgents: []string{"cro"}}},
		{ID: "cro", Subagents: &config.SubagentsConfig{AllowAgents: []string{"cmo", "cfo"}}},
		{ID: "cmo"},
		{ID: "cfo"},
	})
	providerErr := errors.New("cfo unavailable")
	provider := &failingMeetingProvider{err: providerErr}
	al := NewAgentLoop(cfg, bus.NewMessageBus(), provider)

	_, record, err := al.RunAgentMeeting(context.Background(), AgentMeetingRequest{
		Title:               "Sales lift",
		SponsorAgentID:      "ceo",
		ChairAgentID:        "cro",
		ParticipantAgentIDs: []string{"cmo", "cfo"},
		Goal:                "Review plan.",
	})
	if !errors.Is(err, providerErr) {
		t.Fatalf("RunAgentMeeting() error = %v, want provider error", err)
	}
	rec, err := al.meetingRecords.Get(context.Background(), record.MeetingID)
	if err != nil {
		t.Fatalf("meetingRecords.Get() error = %v", err)
	}
	if rec.Status != AgentMeetingStatusFailed {
		t.Fatalf("meeting status = %q, want failed", rec.Status)
	}
	if len(rec.ParticipantTurns) != 2 {
		t.Fatalf("ParticipantTurns = %d, want successful and failed turn", len(rec.ParticipantTurns))
	}
	if rec.ParticipantTurns[1].AgentID != "cfo" || rec.ParticipantTurns[1].Status != "failed" {
		t.Fatalf("failed participant turn = %#v", rec.ParticipantTurns[1])
	}
	if strings.Contains(provider.callText(), "Consolidate this chaired meeting") {
		t.Fatalf("chair synthesis ran after required participant failure:\n%s", provider.callText())
	}
}

func TestRunAgentMeeting_ChairFailureMarksMeetingFailed(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "ceo", Subagents: &config.SubagentsConfig{AllowAgents: []string{"cro"}}},
		{ID: "cro", Subagents: &config.SubagentsConfig{AllowAgents: []string{"cmo"}}},
		{ID: "cmo"},
	})
	providerErr := errors.New("chair synthesis failed")
	al := NewAgentLoop(cfg, bus.NewMessageBus(), &failingMeetingProvider{err: providerErr})

	_, record, err := al.RunAgentMeeting(context.Background(), AgentMeetingRequest{
		Title:               "Sales lift",
		SponsorAgentID:      "ceo",
		ChairAgentID:        "cro",
		ParticipantAgentIDs: []string{"cmo"},
		Goal:                "Review plan.",
	})
	if !errors.Is(err, providerErr) {
		t.Fatalf("RunAgentMeeting() error = %v, want provider error", err)
	}
	rec, err := al.meetingRecords.Get(context.Background(), record.MeetingID)
	if err != nil {
		t.Fatalf("meetingRecords.Get() error = %v", err)
	}
	if rec.Status != AgentMeetingStatusFailed {
		t.Fatalf("meeting status = %q, want failed", rec.Status)
	}
	if rec.ChairTurn != nil {
		t.Fatalf("ChairTurn = %#v, want nil on chair failure", rec.ChairTurn)
	}
	if len(rec.ParticipantTurns) != 1 || rec.ParticipantTurns[0].Status != string(TurnEndStatusCompleted) {
		t.Fatalf("ParticipantTurns = %#v, want completed participant preserved", rec.ParticipantTurns)
	}
}

func TestRunAgentMeeting_CancelDuringParticipantPersistenceMarksMeetingCancelled(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "ceo", Subagents: &config.SubagentsConfig{AllowAgents: []string{"cro"}}},
		{ID: "cro", Subagents: &config.SubagentsConfig{AllowAgents: []string{"cmo"}}},
		{ID: "cmo"},
	})
	ctx, cancel := context.WithCancel(context.Background())
	al := NewAgentLoop(cfg, bus.NewMessageBus(), &failingMeetingProvider{cancelAfterCMO: cancel})

	_, record, err := al.RunAgentMeeting(ctx, AgentMeetingRequest{
		Title:               "Sales lift",
		SponsorAgentID:      "ceo",
		ChairAgentID:        "cro",
		ParticipantAgentIDs: []string{"cmo"},
		Goal:                "Review plan.",
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RunAgentMeeting() error = %v, want context.Canceled", err)
	}
	rec, err := al.meetingRecords.Get(context.Background(), record.MeetingID)
	if err != nil {
		t.Fatalf("meetingRecords.Get() error = %v", err)
	}
	if rec.Status != AgentMeetingStatusCancelled {
		t.Fatalf("meeting status = %q, want cancelled", rec.Status)
	}
	if !strings.Contains(rec.Error, context.Canceled.Error()) {
		t.Fatalf("meeting error = %q, want cancellation", rec.Error)
	}
}

func TestRunAgentMeeting_CancelDuringCompletionMarksMeetingCancelled(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "ceo", Subagents: &config.SubagentsConfig{AllowAgents: []string{"cro"}}},
		{ID: "cro", Subagents: &config.SubagentsConfig{AllowAgents: []string{"cmo"}}},
		{ID: "cmo"},
	})
	ctx, cancel := context.WithCancel(context.Background())
	al := NewAgentLoop(cfg, bus.NewMessageBus(), &failingMeetingProvider{cancelOnChair: cancel})

	_, record, err := al.RunAgentMeeting(ctx, AgentMeetingRequest{
		Title:               "Sales lift",
		SponsorAgentID:      "ceo",
		ChairAgentID:        "cro",
		ParticipantAgentIDs: []string{"cmo"},
		Goal:                "Review plan.",
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RunAgentMeeting() error = %v, want context.Canceled", err)
	}
	rec, err := al.meetingRecords.Get(context.Background(), record.MeetingID)
	if err != nil {
		t.Fatalf("meetingRecords.Get() error = %v", err)
	}
	if rec.Status != AgentMeetingStatusCancelled {
		t.Fatalf("meeting status = %q, want cancelled", rec.Status)
	}
	if rec.ChairTurn != nil {
		t.Fatalf("ChairTurn = %#v, want nil when completion record write fails", rec.ChairTurn)
	}
	if len(rec.ParticipantTurns) != 1 {
		t.Fatalf("ParticipantTurns = %d, want completed participant preserved", len(rec.ParticipantTurns))
	}
}
