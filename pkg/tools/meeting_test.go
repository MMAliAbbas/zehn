package tools

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type recordingMeetingRunner struct {
	result MeetingExecutionResult
	err    error
	req    MeetingExecutionRequest
	calls  int
}

func (r *recordingMeetingRunner) StartAgentMeeting(ctx context.Context, req MeetingExecutionRequest) (MeetingExecutionResult, error) {
	r.calls++
	r.req = req
	if r.err != nil {
		return MeetingExecutionResult{}, r.err
	}
	if r.result.MeetingID != "" {
		return r.result, nil
	}
	return MeetingExecutionResult{
		MeetingID:      "meeting-123",
		Recommendation: "Launch with a two-week revenue sprint.",
		Timeline:       []string{"Day 1: align offer", "Day 14: review pipeline"},
		Risks:          []string{"Discounting may weaken margin"},
		FollowUps:      []string{"CRO owns pipeline review"},
	}, nil
}

func TestMeetingTool_Parameters(t *testing.T) {
	tool := NewMeetingTool(nil)

	if tool.Name() != "start_agent_meeting" {
		t.Fatalf("Name() = %q, want start_agent_meeting", tool.Name())
	}

	params := tool.Parameters()
	props, ok := params["properties"].(map[string]any)
	if !ok {
		t.Fatal("properties should be a map")
	}
	for _, name := range []string{
		"title",
		"sponsor_agent_id",
		"chair_agent_id",
		"participant_agent_ids",
		"goal",
		"constraints",
		"notes",
		"artifact_refs",
	} {
		if _, ok := props[name]; !ok {
			t.Fatalf("expected parameter %q in schema", name)
		}
	}
	required, ok := params["required"].([]string)
	if !ok {
		t.Fatal("required should be []string")
	}
	if strings.Join(required, ",") != "title,chair_agent_id,participant_agent_ids,goal" {
		t.Fatalf("required = %v", required)
	}
}

func TestMeetingTool_Execute_StartsMeetingAndReturnsConsolidatedRecommendation(t *testing.T) {
	runner := &recordingMeetingRunner{}
	tool := NewMeetingTool(runner)
	ctx := WithToolSessionContext(context.Background(), "ceo", "session", nil)

	result := tool.Execute(ctx, map[string]any{
		"title":                 "Two-week sales lift",
		"chair_agent_id":        "cro",
		"participant_agent_ids": []any{"cmo", "cfo"},
		"goal":                  "Increase sales by 30% in two weeks.",
		"constraints":           []any{"No customer-facing commitments without approval."},
		"notes":                 "Focus on existing pipeline.",
		"artifact_refs":         []any{"objective:sales-lift"},
	})

	if result.IsError {
		t.Fatalf("Execute() returned error: %s", result.ForLLM)
	}
	if result.Async {
		t.Fatal("start_agent_meeting should complete synchronously")
	}
	if strings.Contains(result.ForUser, "cmo") || strings.Contains(result.ForUser, "cfo") {
		t.Fatalf("ForUser exposed participant details: %q", result.ForUser)
	}
	if !strings.Contains(result.ForUser, "Launch with a two-week revenue sprint.") {
		t.Fatalf("ForUser = %q, want consolidated recommendation", result.ForUser)
	}
	if !strings.Contains(result.ForLLM, "meeting-123") {
		t.Fatalf("ForLLM = %q, want meeting ID", result.ForLLM)
	}
	if runner.calls != 1 {
		t.Fatalf("StartAgentMeeting calls = %d, want 1", runner.calls)
	}
	if runner.req.SponsorAgentID != "ceo" {
		t.Fatalf("SponsorAgentID = %q, want ceo", runner.req.SponsorAgentID)
	}
	if runner.req.ChairAgentID != "cro" {
		t.Fatalf("ChairAgentID = %q, want cro", runner.req.ChairAgentID)
	}
	if got := strings.Join(runner.req.ParticipantAgentIDs, ","); got != "cmo,cfo" {
		t.Fatalf("ParticipantAgentIDs = %q, want cmo,cfo", got)
	}
	if runner.req.ArtifactRefs[0] != "objective:sales-lift" {
		t.Fatalf("ArtifactRefs = %v", runner.req.ArtifactRefs)
	}
}

func TestMeetingTool_Execute_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		args map[string]any
		want string
	}{
		{name: "missing title", args: map[string]any{"chair_agent_id": "cro", "participant_agent_ids": []any{"cmo"}, "goal": "sell"}, want: "title is required"},
		{name: "missing chair", args: map[string]any{"title": "Sales", "participant_agent_ids": []any{"cmo"}, "goal": "sell"}, want: "chair_agent_id is required"},
		{name: "missing participants", args: map[string]any{"title": "Sales", "chair_agent_id": "cro", "goal": "sell"}, want: "participant_agent_ids is required"},
		{name: "missing goal", args: map[string]any{"title": "Sales", "chair_agent_id": "cro", "participant_agent_ids": []any{"cmo"}}, want: "goal is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewMeetingTool(&recordingMeetingRunner{})
			result := tool.Execute(context.Background(), tt.args)
			if !result.IsError {
				t.Fatal("expected error result")
			}
			if !strings.Contains(result.ForLLM, tt.want) {
				t.Fatalf("ForLLM = %q, want %q", result.ForLLM, tt.want)
			}
		})
	}
}

func TestMeetingTool_Execute_RunnerFailure(t *testing.T) {
	tool := NewMeetingTool(&recordingMeetingRunner{err: errors.New("boom")})

	result := tool.Execute(context.Background(), map[string]any{
		"title":                 "Sales",
		"sponsor_agent_id":      "ceo",
		"chair_agent_id":        "cro",
		"participant_agent_ids": []any{"cmo"},
		"goal":                  "sell",
	})

	if !result.IsError {
		t.Fatal("expected execution error")
	}
	if !strings.Contains(result.ForLLM, "meeting execution failed") {
		t.Fatalf("ForLLM = %q", result.ForLLM)
	}
}
