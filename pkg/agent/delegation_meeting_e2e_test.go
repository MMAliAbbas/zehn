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
	"github.com/sipeed/picoclaw/pkg/session"
	"github.com/sipeed/picoclaw/pkg/tools"
)

type e2eDelegationMeetingProvider struct {
	mu    sync.Mutex
	calls []string
}

func (p *e2eDelegationMeetingProvider) Chat(
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
	case strings.Contains(content, "Target agent: li-cmo"):
		return &providers.LLMResponse{Content: "Position: Launch a focused existing-pipeline campaign.\nRisk: message fatigue.\nFollow-up: CMO drafts campaign copy.", FinishReason: "stop"}, nil
	case strings.Contains(content, "Target agent: li-cfo"):
		return &providers.LLMResponse{Content: "Risk: margin pressure.\nDependency: CFO reviews discount guardrails.\nAcceptance criteria: daily margin review.", FinishReason: "stop"}, nil
	case strings.Contains(content, "Consolidate this chaired meeting"):
		return &providers.LLMResponse{Content: "Recommendation: run a two-week existing-pipeline sales sprint.\nTimeline: day 1 segment pipeline; days 2-10 execute outreach; days 11-14 close and review.\nRisks: margin pressure; message fatigue.\nFollow-ups: CRO owns sprint board; CMO drafts copy; CFO reviews discount guardrails.", FinishReason: "stop"}, nil
	case strings.Contains(content, "Ali objective: increase sales by 30%"):
		return &providers.LLMResponse{Content: "CEO objective opened. Delegate sales strategy to li-cro for a chaired domain meeting before asking Ali for approval.", FinishReason: "stop"}, nil
	default:
		return &providers.LLMResponse{Content: "advisory response", FinishReason: "stop"}, nil
	}
}

func (p *e2eDelegationMeetingProvider) GetDefaultModel() string {
	return "provider-default"
}

func (p *e2eDelegationMeetingProvider) snapshot() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]string(nil), p.calls...)
}

func TestEndToEndDelegationMeetingWorkflowUsesFakesAndPreservesBoundaries(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "zehn", Subagents: &config.SubagentsConfig{AllowAgents: []string{"li-ceo"}}},
		{ID: "li-ceo", Subagents: &config.SubagentsConfig{AllowAgents: []string{"li-cro"}}},
		{ID: "li-cro", Subagents: &config.SubagentsConfig{AllowAgents: []string{"li-cmo", "li-cfo"}}},
		{ID: "li-cmo"},
		{ID: "li-cfo"},
	})
	cfg.Channels = discordVisibilityChannels(t, true)
	cfg.Tools.Delegate.Enabled = true
	cfg.Tools.Meeting.Enabled = true
	cfg.Tools.Spawn.Enabled = true
	cfg.Tools.Subagent.Enabled = true

	msgBus := bus.NewMessageBus()
	provider := &e2eDelegationMeetingProvider{}
	al := NewAgentLoop(cfg, msgBus, provider)
	github := &fakeGitHubArtifactWriter{}
	al.SetGitHubArtifactWriter(github)

	restoreWriter := overrideDelegationMemoryWriterForTest(t, &recordingDelegationMemoryWriter{err: errors.New("yaad unavailable")})
	defer restoreWriter()
	restoreStrict := overrideDelegationMemoryStrictForTest(t, false)
	defer restoreStrict()

	ceoResult, err := al.RunAgentDelegation(context.Background(), AgentDelegationRequest{
		ParentAgentID: "zehn",
		TargetAgentID: "li-ceo",
		Task:          "Ali objective: increase sales by 30% in the next two weeks.",
		ThreadKey:     "sales-growth-objective",
		RequestedBy:   "ali",
	})
	if err != nil {
		t.Fatalf("RunAgentDelegation(ceo objective) error = %v", err)
	}
	if !strings.Contains(ceoResult.Content, "Delegate sales strategy to li-cro") {
		t.Fatalf("CEO delegation result = %q", ceoResult.Content)
	}
	if ceoResult.SessionScope == nil || ceoResult.SessionScope.AgentID != "li-ceo" {
		t.Fatalf("CEO delegation scope = %#v, want li-ceo", ceoResult.SessionScope)
	}
	issues, comments := github.snapshot()
	if len(issues) != 0 || len(comments) != 0 {
		t.Fatalf("advisory CEO delegation created GitHub artifacts: %d issues/%d comments", len(issues), len(comments))
	}

	meetingTool := tools.NewMeetingTool(al)
	toolCtx := tools.WithToolSessionContext(context.Background(), "li-ceo", "test-session", &session.SessionScope{AgentID: "li-ceo"})
	toolResult := meetingTool.Execute(toolCtx, map[string]any{
		"title":                 "Two-week sales growth sprint",
		"chair_agent_id":        "li-cro",
		"participant_agent_ids": []any{"li-cmo", "li-cfo"},
		"goal":                  "Produce one recommendation for increasing sales by 30% in the next two weeks.",
		"constraints":           []any{"No customer-facing discounts without approval.", "Use existing pipeline first."},
		"approvals":             []any{"Ali approval before customer-facing discounts."},
		"artifact_refs":         []any{"objective:sales-growth-30"},
	})
	if toolResult == nil {
		t.Fatal("meeting tool returned nil result")
	}
	if toolResult.IsError {
		t.Fatalf("meeting tool error: %s", toolResult.ForLLM)
	}
	for _, want := range []string{
		"Consolidated recommendation: run a two-week existing-pipeline sales sprint",
		"Participants: li-cmo, li-cfo",
		"Timeline: day 1 segment pipeline",
		"Risks: margin pressure; message fatigue",
		"Follow-ups: CRO owns sprint board",
		"Approval needed: Ali approval before customer-facing discounts.",
	} {
		if !strings.Contains(toolResult.ForLLM, want) {
			t.Fatalf("meeting ForLLM missing %q:\n%s", want, toolResult.ForLLM)
		}
	}

	issues, comments = github.snapshot()
	if len(issues) != 1 {
		t.Fatalf("GitHub issues = %d, want 1 executable meeting issue", len(issues))
	}
	if len(comments) != 2 {
		t.Fatalf("GitHub comments = %d, want participant comments", len(comments))
	}
	if strings.Contains(issues[0].Body, "Consolidate this chaired meeting") {
		t.Fatalf("GitHub issue contains raw internal chair prompt:\n%s", issues[0].Body)
	}

	records, err := al.delegationRecords.List(context.Background(), AgentDelegationRecordQuery{IncludePrivateAll: true})
	if err != nil {
		t.Fatalf("delegationRecords.List() error = %v", err)
	}
	if len(records) != 4 {
		t.Fatalf("delegation records = %d, want CEO objective, two participants, and chair synthesis", len(records))
	}
	for _, rec := range records {
		if rec.Status != AgentDelegationStatusCompleted {
			t.Fatalf("delegation %s status = %q, want completed", rec.DelegationID, rec.Status)
		}
		if rec.DurableMemory == nil || rec.DurableMemory.Status != AgentDelegationMemoryStatusFailed {
			t.Fatalf("delegation %s durable memory = %#v, want non-strict Yaad failure recorded", rec.DelegationID, rec.DurableMemory)
		}
	}

	meetings, err := al.meetingRecords.List(context.Background())
	if err != nil {
		t.Fatalf("meetingRecords.List() error = %v", err)
	}
	if len(meetings) != 1 {
		t.Fatalf("meeting records = %d, want 1", len(meetings))
	}
	meeting := meetings[0]
	if meeting.SponsorAgentID != "li-ceo" || meeting.ChairAgentID != "li-cro" {
		t.Fatalf("meeting sponsor/chair = %s/%s, want li-ceo/li-cro", meeting.SponsorAgentID, meeting.ChairAgentID)
	}
	if got := strings.Join(meeting.Participants, ","); got != "li-cmo,li-cfo" {
		t.Fatalf("meeting participants = %q, want li-cmo,li-cfo", got)
	}
	if meeting.GitHubArtifact == nil || meeting.GitHubArtifact.Status != AgentGitHubArtifactStatusCreated {
		t.Fatalf("meeting GitHub artifact = %#v, want created", meeting.GitHubArtifact)
	}

	calls := strings.Join(provider.snapshot(), "\n---\n")
	for _, want := range []string{
		"Ali objective: increase sales by 30%",
		"Target agent: li-cmo",
		"Target agent: li-cfo",
		"Consolidate this chaired meeting",
	} {
		if !strings.Contains(calls, want) {
			t.Fatalf("provider calls missing %q:\n%s", want, calls)
		}
	}

	ceo, ok := al.registry.GetAgent("li-ceo")
	if !ok {
		t.Fatal("li-ceo agent missing")
	}
	if _, ok := ceo.Tools.Get("delegate_to_agent"); !ok {
		t.Fatal("delegate_to_agent should remain a separate tool")
	}
	if _, ok := ceo.Tools.Get("start_agent_meeting"); !ok {
		t.Fatal("start_agent_meeting should be registered when enabled")
	}
	if _, ok := ceo.Tools.Get("spawn"); !ok {
		t.Fatal("spawn should remain registered when enabled")
	}
	if _, ok := ceo.Tools.Get("subagent"); !ok {
		t.Fatal("subagent should remain registered when enabled")
	}

	visibility := collectOutboundMessages(t, msgBus, 11)
	joinedVisibility := make([]string, 0, len(visibility))
	for _, msg := range visibility {
		joinedVisibility = append(joinedVisibility, msg.Content)
	}
	if strings.Contains(strings.Join(joinedVisibility, "\n"), "Consolidate this chaired meeting") {
		t.Fatalf("Discord visibility leaked internal meeting prompt:\n%s", strings.Join(joinedVisibility, "\n"))
	}
}
