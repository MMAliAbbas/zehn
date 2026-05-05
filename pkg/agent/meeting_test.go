package agent

import (
	"context"
	"encoding/json"
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
