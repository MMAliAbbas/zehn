package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/sipeed/picoclaw/pkg/routing"
)

// DelegateTool delegates work to a configured peer agent. Zehn keeps the
// durable delegate_to_agent tool contract while adopting upstream's normalized
// target-agent routing semantics.
type DelegateTool struct {
	spawner        SubTurnSpawner
	defaultModel   string
	maxTokens      int
	temperature    float64
	allowlistCheck func(targetAgentID string) bool
	selfAgentID    string
	targetExists   func(targetAgentID string) bool
	targetModel    func(targetAgentID string) string
	runner         DelegationRunner
}

func NewDelegateTool(managers ...*SubagentManager) *DelegateTool {
	tool := &DelegateTool{}
	if len(managers) == 0 || managers[0] == nil {
		return tool
	}
	manager := managers[0]
	tool.defaultModel = manager.defaultModel
	tool.maxTokens = manager.maxTokens
	tool.temperature = manager.temperature
	return tool
}

func (t *DelegateTool) SetSpawner(spawner SubTurnSpawner) {
	t.spawner = spawner
}

func (t *DelegateTool) SetAllowlistChecker(check func(targetAgentID string) bool) {
	t.allowlistCheck = check
}

func (t *DelegateTool) SetSelfAgentID(id string) {
	t.selfAgentID = routing.NormalizeAgentID(id)
}

func (t *DelegateTool) SetTargetExistsChecker(check func(targetAgentID string) bool) {
	t.targetExists = check
}

func (t *DelegateTool) SetTargetModelResolver(resolve func(targetAgentID string) string) {
	t.targetModel = resolve
}

func (t *DelegateTool) SetDelegationRunner(runner DelegationRunner) {
	t.runner = runner
}

func (t *DelegateTool) Name() string {
	return "delegate_to_agent"
}

func (t *DelegateTool) Description() string {
	return "Delegate a task, question, or review to an allowed configured peer agent. Sync mode returns the response; async mode returns a durable delegation ID immediately."
}

func (t *DelegateTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"agent_id": map[string]any{
				"type":        "string",
				"description": "Target configured agent ID.",
			},
			"task": map[string]any{
				"type":        "string",
				"description": "Bounded task, question, review, or work request for the target agent.",
			},
			"mode": map[string]any{
				"type":        "string",
				"description": "Execution mode: sync or async. Defaults to sync.",
				"enum":        []string{"sync", "async"},
			},
			"thread_key": map[string]any{
				"type":        "string",
				"description": "Optional stable key for related delegation follow-ups.",
			},
			"priority": map[string]any{
				"type":        "string",
				"description": "Optional priority label.",
			},
			"due": map[string]any{
				"type":        "string",
				"description": "Optional due date or deadline.",
			},
			"artifact_refs": map[string]any{
				"type":        "array",
				"description": "Optional related files, issues, PRs, memory IDs, project items, or meeting IDs.",
				"items": map[string]any{
					"type": "string",
				},
			},
		},
		"required": []string{"agent_id", "task"},
	}
}

func (t *DelegateTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	rawAgentID, ok := args["agent_id"].(string)
	if !ok || strings.TrimSpace(rawAgentID) == "" {
		return delegateError("missing_agent_id", "agent_id is required and must be a non-empty string")
	}
	agentID := routing.NormalizeAgentID(rawAgentID)

	task, ok := args["task"].(string)
	task = strings.TrimSpace(task)
	if !ok || task == "" {
		return delegateError("missing_task", "task is required and must be a non-empty string")
	}

	if t.selfAgentID != "" && agentID == t.selfAgentID {
		return delegateError("self_delegation", "cannot delegate to self")
	}
	if t.allowlistCheck != nil && !t.allowlistCheck(agentID) {
		return delegateError("denied_target", fmt.Sprintf("not allowed to delegate to agent %q", agentID))
	}
	if t.targetExists != nil && !t.targetExists(agentID) {
		return delegateError("missing_target", fmt.Sprintf("target agent %q not found", agentID))
	}

	threadKey, _ := args["thread_key"].(string)
	priority, _ := args["priority"].(string)
	due, _ := args["due"].(string)
	artifactRefs := delegateArtifactRefs(args["artifact_refs"])
	mode, _ := args["mode"].(string)
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		mode = "sync"
	}
	if mode != "sync" && mode != "async" {
		return delegateError("invalid_mode", "mode must be either sync or async")
	}

	req := DelegateExecutionRequest{
		ParentAgentID: ToolAgentID(ctx),
		TargetAgentID: agentID,
		Task:          task,
		ThreadKey:     strings.TrimSpace(threadKey),
		Mode:          mode,
		Priority:      strings.TrimSpace(priority),
		Due:           strings.TrimSpace(due),
		ArtifactRefs:  artifactRefs,
	}

	if mode == "async" {
		if t.runner == nil {
			return delegateError("execution_failed", "delegated async execution failed: delegation runner not configured")
		}
		result, err := t.runner.StartDelegation(ctx, req)
		if err != nil {
			return delegateError("execution_failed", fmt.Sprintf("delegated async execution failed: %v", err)).WithError(err)
		}
		return &ToolResult{
			ForLLM: fmt.Sprintf(
				"Async delegation to %s started with delegation_id=%s. Use delegation_status to inspect progress.",
				agentID,
				result.DelegationID,
			),
			ForUser: fmt.Sprintf("Delegation %s started for %s.", result.DelegationID, agentID),
			Async:   true,
		}
	}

	if t.runner != nil {
		result, err := t.runner.RunDelegation(ctx, req)
		if err != nil {
			return delegateError("execution_failed", fmt.Sprintf("delegated execution failed: %v", err)).WithError(err)
		}
		return &ToolResult{
			ForLLM:  fmt.Sprintf("Delegation to %s completed.\nDelegation ID: %s\nResult: %s", agentID, result.DelegationID, result.Content),
			ForUser: delegateCompact(result.Content),
			Silent:  false,
			IsError: false,
			Async:   false,
		}
	}
	if t.spawner == nil {
		return delegateError("execution_failed", "delegated execution failed: delegate tool not configured")
	}

	model := t.defaultModel
	if t.targetModel != nil {
		if targetModel := strings.TrimSpace(t.targetModel(agentID)); targetModel != "" {
			model = targetModel
		}
	}

	result, err := t.spawner.SpawnSubTurn(ctx, SubTurnConfig{
		TargetAgentID: agentID,
		Model:         model,
		SystemPrompt:  delegatePrompt(agentID, task, threadKey, priority, due, artifactRefs),
		MaxTokens:     t.maxTokens,
		Temperature:   t.temperature,
		Async:         false,
	})
	if err != nil {
		return delegateError("execution_failed", fmt.Sprintf("delegated execution failed: %v", err)).WithError(err)
	}
	if result == nil {
		err := fmt.Errorf("delegate_to_agent returned nil result")
		return delegateError("execution_failed", "delegated execution failed: nil result").WithError(err)
	}

	targetResponse := result.ForLLM
	if strings.TrimSpace(result.ForUser) != "" {
		targetResponse = result.ForUser
	}

	return &ToolResult{
		ForLLM:  fmt.Sprintf("Delegation to %s completed.\nResult: %s", agentID, result.ContentForLLM()),
		ForUser: delegateCompact(targetResponse),
		Silent:  false,
		IsError: result.IsError,
		Async:   false,
		Err:     result.Err,
	}
}

func delegateError(code, message string) *ToolResult {
	return ErrorResult(fmt.Sprintf("delegate_to_agent error [%s]: %s", code, message)).
		WithError(fmt.Errorf("%s: %s", code, message))
}

func delegatePrompt(agentID, task, threadKey, priority, due string, artifactRefs []string) string {
	var b strings.Builder
	b.WriteString("You are the configured target agent receiving a bounded synchronous delegation.\n")
	b.WriteString("Return concise advice, review, or work output for the calling agent.\n\n")
	b.WriteString("Target agent: ")
	b.WriteString(agentID)
	b.WriteString("\nTask: ")
	b.WriteString(task)
	if strings.TrimSpace(threadKey) != "" {
		b.WriteString("\nThread key: ")
		b.WriteString(strings.TrimSpace(threadKey))
	}
	if strings.TrimSpace(priority) != "" {
		b.WriteString("\nPriority: ")
		b.WriteString(strings.TrimSpace(priority))
	}
	if strings.TrimSpace(due) != "" {
		b.WriteString("\nDue: ")
		b.WriteString(strings.TrimSpace(due))
	}
	if len(artifactRefs) > 0 {
		b.WriteString("\nArtifacts: ")
		b.WriteString(strings.Join(artifactRefs, ", "))
	}
	return b.String()
}

func delegateArtifactRefs(value any) []string {
	switch refs := value.(type) {
	case []string:
		return compactStringRefs(refs)
	case []any:
		values := make([]string, 0, len(refs))
		for _, ref := range refs {
			if s, ok := ref.(string); ok {
				values = append(values, s)
			}
		}
		return compactStringRefs(values)
	default:
		return nil
	}
}

func compactStringRefs(refs []string) []string {
	values := make([]string, 0, len(refs))
	for _, ref := range refs {
		ref = strings.TrimSpace(ref)
		if ref != "" {
			values = append(values, ref)
		}
	}
	return values
}

func delegateCompact(value string) string {
	value = strings.TrimSpace(value)
	const maxRunes = 500
	runes := []rune(value)
	if len(runes) <= maxRunes {
		return value
	}
	return strings.TrimSpace(string(runes[:maxRunes])) + "..."
}
