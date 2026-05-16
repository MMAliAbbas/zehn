package tools

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type recordingDelegateSpawner struct {
	result *ToolResult
	err    error
	cfg    SubTurnConfig
	calls  int
}

func (s *recordingDelegateSpawner) SpawnSubTurn(_ context.Context, cfg SubTurnConfig) (*ToolResult, error) {
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

type recordingDelegationRunner struct {
	runResult   DelegateExecutionResult
	runErr      error
	runReq      DelegateExecutionRequest
	runCalls    int
	startResult DelegateExecutionResult
	startErr    error
	startReq    DelegateExecutionRequest
	startCalls  int
}

func (r *recordingDelegationRunner) RunDelegation(_ context.Context, req DelegateExecutionRequest) (DelegateExecutionResult, error) {
	r.runCalls++
	r.runReq = req
	if r.runErr != nil {
		return DelegateExecutionResult{}, r.runErr
	}
	if r.runResult.DelegationID != "" {
		return r.runResult, nil
	}
	return DelegateExecutionResult{DelegationID: "delegation-sync", TargetAgentID: req.TargetAgentID, Content: "sync response"}, nil
}

func (r *recordingDelegationRunner) StartDelegation(_ context.Context, req DelegateExecutionRequest) (DelegateExecutionResult, error) {
	r.startCalls++
	r.startReq = req
	if r.startErr != nil {
		return DelegateExecutionResult{}, r.startErr
	}
	if r.startResult.DelegationID != "" {
		return r.startResult, nil
	}
	return DelegateExecutionResult{DelegationID: "delegation-123", TargetAgentID: req.TargetAgentID}, nil
}

func TestDelegateTool_Name(t *testing.T) {
	tool := NewDelegateTool()
	if tool.Name() != "delegate_to_agent" {
		t.Fatalf("Name() = %q, want delegate_to_agent", tool.Name())
	}
}

func TestDelegateTool_Parameters(t *testing.T) {
	tool := NewDelegateTool()

	params := tool.Parameters()
	props, ok := params["properties"].(map[string]any)
	if !ok {
		t.Fatal("properties should be a map")
	}
	for _, name := range []string{"agent_id", "task", "mode", "thread_key", "priority", "due", "artifact_refs"} {
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

func TestDelegateTool_Execute_AsyncReturnsDelegationIDImmediately(t *testing.T) {
	runner := &recordingDelegationRunner{}
	tool := NewDelegateTool()
	tool.SetDelegationRunner(runner)
	tool.SetAllowlistChecker(func(targetAgentID string) bool { return targetAgentID == "engineering" })
	tool.SetTargetExistsChecker(func(targetAgentID string) bool { return targetAgentID == "engineering" })

	start := time.Now()
	result := tool.Execute(context.Background(), map[string]any{
		"agent_id":   "Engineering",
		"task":       "Inspect the repository and report risks.",
		"mode":       "async",
		"thread_key": "repo-risk",
	})

	if time.Since(start) > time.Second {
		t.Fatal("async delegation should return immediately")
	}
	if result.IsError {
		t.Fatalf("Execute() returned error: %s", result.ForLLM)
	}
	if !result.Async {
		t.Fatal("async delegation result should be marked async")
	}
	if !strings.Contains(result.ForLLM, "delegation-123") {
		t.Fatalf("ForLLM = %q, want delegation ID", result.ForLLM)
	}
	if runner.startCalls != 1 {
		t.Fatalf("StartDelegation calls = %d, want 1", runner.startCalls)
	}
	if runner.startReq.TargetAgentID != "engineering" {
		t.Fatalf("TargetAgentID = %q, want engineering", runner.startReq.TargetAgentID)
	}
	if runner.startReq.Mode != "async" {
		t.Fatalf("Mode = %q, want async", runner.startReq.Mode)
	}
	if runner.startReq.ThreadKey != "repo-risk" {
		t.Fatalf("ThreadKey = %q, want repo-risk", runner.startReq.ThreadKey)
	}
}

func TestDelegateTool_Execute_SyncRunnerSuccess(t *testing.T) {
	runner := &recordingDelegationRunner{}
	tool := NewDelegateTool()
	tool.SetDelegationRunner(runner)
	tool.SetTargetExistsChecker(func(targetAgentID string) bool { return targetAgentID == "cto" })

	result := tool.Execute(context.Background(), map[string]any{
		"agent_id": "CTO",
		"task":     "Review risk.",
	})

	if result.IsError {
		t.Fatalf("Execute() returned error: %s", result.ForLLM)
	}
	if result.Async {
		t.Fatal("sync delegation should not be async")
	}
	if runner.runCalls != 1 {
		t.Fatalf("RunDelegation calls = %d, want 1", runner.runCalls)
	}
	if runner.runReq.TargetAgentID != "cto" {
		t.Fatalf("TargetAgentID = %q, want cto", runner.runReq.TargetAgentID)
	}
	if !strings.Contains(result.ForLLM, "delegation-sync") {
		t.Fatalf("ForLLM = %q, want delegation ID", result.ForLLM)
	}
}

func TestDelegateTool_Execute_SubturnFallbackSuccess(t *testing.T) {
	spawner := &recordingDelegateSpawner{result: UserResult("risk looks acceptable")}
	tool := NewDelegateTool()
	tool.SetSpawner(spawner)
	tool.SetAllowlistChecker(func(targetAgentID string) bool { return targetAgentID == "ciso" })
	tool.SetTargetExistsChecker(func(targetAgentID string) bool { return targetAgentID == "ciso" })

	result := tool.Execute(context.Background(), map[string]any{
		"agent_id":      "CISO",
		"task":          "Review launch risk.",
		"thread_key":    "launch-risk",
		"priority":      "high",
		"due":           "2026-05-06",
		"artifact_refs": []any{"docs/launch.md", "issue:123"},
	})

	if result.IsError {
		t.Fatalf("Execute() returned error: %s", result.ForLLM)
	}
	if got := result.ForUser; got != "risk looks acceptable" {
		t.Fatalf("ForUser = %q, want target response", got)
	}
	if spawner.calls != 1 {
		t.Fatalf("spawner calls = %d, want 1", spawner.calls)
	}
	if spawner.cfg.TargetAgentID != "ciso" {
		t.Fatalf("TargetAgentID = %q, want ciso", spawner.cfg.TargetAgentID)
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
		{
			name: "invalid mode",
			args: map[string]any{"agent_id": "ciso", "task": "review", "mode": "later"},
			want: "mode must be either sync or async",
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

func TestDelegateTool_Execute_SelfDelegation(t *testing.T) {
	tool := NewDelegateTool()
	tool.SetSpawner(&recordingDelegateSpawner{})
	tool.SetSelfAgentID("alpha")

	for _, agentID := range []string{"alpha", "ALPHA", " Alpha "} {
		t.Run(agentID, func(t *testing.T) {
			result := tool.Execute(context.Background(), map[string]any{
				"agent_id": agentID,
				"task":     "test",
			})
			if !result.IsError {
				t.Fatal("expected error for self-delegation")
			}
			if !strings.Contains(result.ForLLM, "cannot delegate to self") {
				t.Fatalf("ForLLM = %q", result.ForLLM)
			}
		})
	}
}

func TestDelegateTool_Execute_NilResult(t *testing.T) {
	tool := NewDelegateTool()
	tool.SetSpawner(&nilResultSpawner{})

	result := tool.Execute(context.Background(), map[string]any{
		"agent_id": "researcher",
		"task":     "test",
	})

	if !result.IsError {
		t.Fatal("expected error for nil result")
	}
	if !strings.Contains(result.ForLLM, "nil result") {
		t.Fatalf("ForLLM = %q", result.ForLLM)
	}
}

type nilResultSpawner struct{}

func (m *nilResultSpawner) SpawnSubTurn(_ context.Context, _ SubTurnConfig) (*ToolResult, error) {
	return nil, nil
}
