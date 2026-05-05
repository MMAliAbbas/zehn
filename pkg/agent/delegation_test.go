package agent

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/providers"
	"github.com/sipeed/picoclaw/pkg/session"
)

type recordingDelegationProvider struct {
	mu       sync.Mutex
	messages []providers.Message
	model    string
}

func (p *recordingDelegationProvider) Chat(
	ctx context.Context,
	messages []providers.Message,
	tools []providers.ToolDefinition,
	model string,
	opts map[string]any,
) (*providers.LLMResponse, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.messages = append([]providers.Message(nil), messages...)
	p.model = model
	return &providers.LLMResponse{Content: "target answer", FinishReason: "stop"}, nil
}

func (p *recordingDelegationProvider) GetDefaultModel() string {
	return "provider-default"
}

func (p *recordingDelegationProvider) lastCall() ([]providers.Message, string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]providers.Message(nil), p.messages...), p.model
}

func TestRunAgentDelegation_UsesTargetAgentSessionScopeAndPrompts(t *testing.T) {
	tmpDir := t.TempDir()
	parentWorkspace := filepath.Join(tmpDir, "parent")
	targetWorkspace := filepath.Join(tmpDir, "target")
	writeDelegationPromptFile(t, parentWorkspace, "AGENT.md", "# Parent\nPARENT_ONLY_MARKER")
	writeDelegationPromptFile(t, targetWorkspace, "AGENT.md", "# Target\nTARGET_ONLY_MARKER")
	writeDelegationPromptFile(t, targetWorkspace, "SOUL.md", "# Soul\nTARGET_SOUL_MARKER")
	writeDelegationPromptFile(t, targetWorkspace, "USER.md", "# User\nTARGET_USER_MARKER")

	provider := &recordingDelegationProvider{}
	cfg := &config.Config{
		Agents: config.AgentsConfig{
			Defaults: config.AgentDefaults{
				Workspace:         tmpDir,
				ModelName:         "default-model",
				MaxTokens:         4096,
				MaxToolIterations: 1,
			},
			List: []config.AgentConfig{
				{
					ID:        "parent",
					Workspace: parentWorkspace,
					Subagents: &config.SubagentsConfig{
						AllowAgents: []string{"target"},
					},
				},
				{
					ID:        "target",
					Workspace: targetWorkspace,
					Model:     &config.AgentModelConfig{Primary: "target-model"},
				},
			},
		},
	}
	al := NewAgentLoop(cfg, bus.NewMessageBus(), provider)
	collector, cleanup := newEventCollector(t, al)
	defer cleanup()

	result, err := al.RunAgentDelegation(context.Background(), AgentDelegationRequest{
		ParentAgentID: "parent",
		TargetAgentID: "target",
		Task:          "review this decision",
		ThreadKey:     "launch-review",
	})
	if err != nil {
		t.Fatalf("RunAgentDelegation() error = %v", err)
	}
	if result.TargetAgentID != "target" {
		t.Fatalf("TargetAgentID = %q, want target", result.TargetAgentID)
	}
	if result.SessionScope == nil || result.SessionScope.AgentID != "target" {
		t.Fatalf("SessionScope = %#v, want target agent scope", result.SessionScope)
	}
	if result.SessionScope.Channel != "internal" {
		t.Fatalf("SessionScope.Channel = %q, want internal", result.SessionScope.Channel)
	}
	if result.SessionScope.Values["delegation"] != "parent:target:launch-review" {
		t.Fatalf("delegation scope value = %q", result.SessionScope.Values["delegation"])
	}
	if result.SessionKey != session.BuildSessionKey(*result.SessionScope) {
		t.Fatalf("SessionKey = %q, want key built from result scope", result.SessionKey)
	}
	if result.Content != "target answer" {
		t.Fatalf("Content = %q, want target answer", result.Content)
	}

	messages, model := provider.lastCall()
	if model != "target-model" {
		t.Fatalf("model = %q, want target-model", model)
	}
	prompt := strings.Join(messageContents(messages), "\n")
	for _, marker := range []string{"TARGET_ONLY_MARKER", "TARGET_SOUL_MARKER", "TARGET_USER_MARKER"} {
		if !strings.Contains(prompt, marker) {
			t.Fatalf("target prompt marker %q missing from prompt:\n%s", marker, prompt)
		}
	}
	if strings.Contains(prompt, "PARENT_ONLY_MARKER") {
		t.Fatalf("prompt used parent workspace content:\n%s", prompt)
	}

	collector.mu.Lock()
	events := append([]Event(nil), collector.events...)
	collector.mu.Unlock()
	if len(events) == 0 {
		t.Fatal("expected turn events")
	}
	for _, evt := range events {
		if evt.Meta.AgentID != "target" {
			t.Fatalf("event agent ID = %q, want target for event %+v", evt.Meta.AgentID, evt)
		}
		if evt.Meta.SessionKey != result.SessionKey {
			t.Fatalf("event session key = %q, want %q", evt.Meta.SessionKey, result.SessionKey)
		}
		if evt.Context == nil || evt.Context.Scope == nil || evt.Context.Scope.AgentID != "target" {
			t.Fatalf("event context scope = %#v, want target scope", evt.Context)
		}
	}
}

func TestRunAgentDelegation_PermissionDenied(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "parent", Subagents: &config.SubagentsConfig{AllowAgents: []string{"allowed"}}},
		{ID: "target"},
	})
	al := NewAgentLoop(cfg, bus.NewMessageBus(), &mockProvider{})

	_, err := al.RunAgentDelegation(context.Background(), AgentDelegationRequest{
		ParentAgentID: "parent",
		TargetAgentID: "target",
		Task:          "no permission",
	})
	if !errors.Is(err, ErrAgentDelegationPermissionDenied) {
		t.Fatalf("error = %v, want ErrAgentDelegationPermissionDenied", err)
	}
}

func TestRunAgentDelegation_MissingTarget(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "parent", Subagents: &config.SubagentsConfig{AllowAgents: []string{"*"}}},
	})
	al := NewAgentLoop(cfg, bus.NewMessageBus(), &mockProvider{})

	_, err := al.RunAgentDelegation(context.Background(), AgentDelegationRequest{
		ParentAgentID: "parent",
		TargetAgentID: "missing",
		Task:          "find missing target",
	})
	if !errors.Is(err, ErrAgentDelegationTargetNotFound) {
		t.Fatalf("error = %v, want ErrAgentDelegationTargetNotFound", err)
	}
}

func writeDelegationPromptFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func delegationConfigWithAgents(t *testing.T, agents []config.AgentConfig) *config.Config {
	t.Helper()
	workspace := t.TempDir()
	for i := range agents {
		if agents[i].Workspace == "" {
			agents[i].Workspace = filepath.Join(workspace, agents[i].ID)
		}
	}
	return &config.Config{
		Agents: config.AgentsConfig{
			Defaults: config.AgentDefaults{
				Workspace:         workspace,
				ModelName:         "test-model",
				MaxTokens:         4096,
				MaxToolIterations: 1,
			},
			List: agents,
		},
	}
}

func messageContents(messages []providers.Message) []string {
	contents := make([]string, 0, len(messages))
	for _, msg := range messages {
		contents = append(contents, msg.Content)
	}
	return contents
}
