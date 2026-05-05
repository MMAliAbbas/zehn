package tools

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type recordingDelegateSpawner struct {
	result *ToolResult
	err    error
	cfg    SubTurnConfig
	calls  int
}

func (s *recordingDelegateSpawner) SpawnSubTurn(ctx context.Context, cfg SubTurnConfig) (*ToolResult, error) {
	s.calls++
	s.cfg = cfg
	if s.err != nil {
		return nil, s.err
	}
	if s.result != nil {
		return s.result, nil
	}
	return NewToolResult("target response"), nil
}

func TestDelegateTool_Parameters(t *testing.T) {
	tool := NewDelegateTool()

	if tool.Name() != "delegate_to_agent" {
		t.Fatalf("Name() = %q, want delegate_to_agent", tool.Name())
	}

	params := tool.Parameters()
	props, ok := params["properties"].(map[string]any)
	if !ok {
		t.Fatal("properties should be a map")
	}
	for _, name := range []string{"agent_id", "task", "thread_key", "priority", "due", "artifact_refs"} {
		if _, ok := props[name]; !ok {
			t.Fatalf("expected parameter %q in schema", name)
		}
	}
	required, ok := params["required"].([]string)
	if !ok {
		t.Fatal("required should be []string")
	}
	if len(required) != 2 || required[0] != "agent_id" || required[1] != "task" {
		t.Fatalf("required = %v, want [agent_id task]", required)
	}
}

func TestDelegateTool_Execute_SyncSuccess(t *testing.T) {
	spawner := &recordingDelegateSpawner{result: UserResult("risk looks acceptable")}
	tool := NewDelegateTool()
	tool.SetSpawner(spawner)
	tool.SetAllowlistChecker(func(targetAgentID string) bool { return targetAgentID == "ciso" })
	tool.SetTargetExistsChecker(func(targetAgentID string) bool { return targetAgentID == "ciso" })

	result := tool.Execute(context.Background(), map[string]any{
		"agent_id":      "ciso",
		"task":          "Review launch risk.",
		"thread_key":    "launch-risk",
		"priority":      "high",
		"due":           "2026-05-06",
		"artifact_refs": []any{"docs/launch.md", "issue:123"},
	})

	if result.IsError {
		t.Fatalf("Execute() returned error: %s", result.ForLLM)
	}
	if result.Async {
		t.Fatal("delegate_to_agent should execute synchronously")
	}
	if got := result.ForUser; got != "risk looks acceptable" {
		t.Fatalf("ForUser = %q, want target response", got)
	}
	if !strings.Contains(result.ForLLM, "Delegation to ciso completed") {
		t.Fatalf("ForLLM = %q, want delegation summary", result.ForLLM)
	}
	if spawner.calls != 1 {
		t.Fatalf("spawner calls = %d, want 1", spawner.calls)
	}
	if spawner.cfg.Async {
		t.Fatal("SubTurnConfig.Async = true, want false")
	}
	for _, want := range []string{"Target agent: ciso", "Task: Review launch risk.", "Thread key: launch-risk", "Artifacts: docs/launch.md, issue:123"} {
		if !strings.Contains(spawner.cfg.SystemPrompt, want) {
			t.Fatalf("SystemPrompt missing %q:\n%s", want, spawner.cfg.SystemPrompt)
		}
	}
}

func TestDelegateTool_Execute_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		args map[string]any
		want string
	}{
		{
			name: "missing agent ID",
			args: map[string]any{"task": "review"},
			want: "agent_id is required",
		},
		{
			name: "missing task",
			args: map[string]any{"agent_id": "ciso"},
			want: "task is required",
		},
		{
			name: "empty task",
			args: map[string]any{"agent_id": "ciso", "task": " \n\t"},
			want: "task is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewDelegateTool()
			result := tool.Execute(context.Background(), tt.args)
			if !result.IsError {
				t.Fatal("expected error result")
			}
			if !strings.Contains(result.ForLLM, tt.want) {
				t.Fatalf("ForLLM = %q, want %q", result.ForLLM, tt.want)
			}
			if result.Err == nil {
				t.Fatal("Err should be set")
			}
		})
	}
}

func TestDelegateTool_Execute_DeniedTarget(t *testing.T) {
	tool := NewDelegateTool()
	tool.SetAllowlistChecker(func(targetAgentID string) bool { return false })
	tool.SetTargetExistsChecker(func(targetAgentID string) bool { return true })

	result := tool.Execute(context.Background(), map[string]any{
		"agent_id": "cto",
		"task":     "review",
	})

	if !result.IsError {
		t.Fatal("expected denied target error")
	}
	if !strings.Contains(result.ForLLM, "not allowed to delegate to agent \"cto\"") {
		t.Fatalf("ForLLM = %q", result.ForLLM)
	}
}

func TestDelegateTool_Execute_MissingTarget(t *testing.T) {
	tool := NewDelegateTool()
	tool.SetAllowlistChecker(func(targetAgentID string) bool { return true })
	tool.SetTargetExistsChecker(func(targetAgentID string) bool { return false })

	result := tool.Execute(context.Background(), map[string]any{
		"agent_id": "ghost",
		"task":     "review",
	})

	if !result.IsError {
		t.Fatal("expected missing target error")
	}
	if !strings.Contains(result.ForLLM, "target agent \"ghost\" not found") {
		t.Fatalf("ForLLM = %q", result.ForLLM)
	}
}

func TestDelegateTool_Execute_ExecutionFailure(t *testing.T) {
	spawner := &recordingDelegateSpawner{err: errors.New("boom")}
	tool := NewDelegateTool()
	tool.SetSpawner(spawner)
	tool.SetAllowlistChecker(func(targetAgentID string) bool { return true })
	tool.SetTargetExistsChecker(func(targetAgentID string) bool { return true })

	result := tool.Execute(context.Background(), map[string]any{
		"agent_id": "ciso",
		"task":     "review",
	})

	if !result.IsError {
		t.Fatal("expected execution failure")
	}
	if !strings.Contains(result.ForLLM, "delegated execution failed") {
		t.Fatalf("ForLLM = %q", result.ForLLM)
	}
	if result.Err == nil {
		t.Fatal("Err should be set")
	}
}
