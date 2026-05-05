package agent

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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

func TestRunAgentDelegationAsync_ReturnsIDAndPersistsCompletedResult(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "parent", Subagents: &config.SubagentsConfig{AllowAgents: []string{"target"}}},
		{ID: "target"},
	})
	provider := &blockingDelegationProvider{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	al := NewAgentLoop(cfg, bus.NewMessageBus(), provider)

	result, err := al.RunAgentDelegationAsync(context.Background(), AgentDelegationRequest{
		ParentAgentID: "parent",
		TargetAgentID: "target",
		Task:          "prepare async report",
		ThreadKey:     "async-report",
	})
	if err != nil {
		t.Fatalf("RunAgentDelegationAsync() error = %v", err)
	}
	if result.DelegationID == "" {
		t.Fatal("DelegationID is empty")
	}
	if result.TargetAgentID != "target" {
		t.Fatalf("TargetAgentID = %q, want target", result.TargetAgentID)
	}

	select {
	case <-provider.started:
	case <-time.After(2 * time.Second):
		t.Fatal("async target turn did not start")
	}
	running, err := al.delegationRecords.Get(context.Background(), result.DelegationID)
	if err != nil {
		t.Fatalf("Get(running) error = %v", err)
	}
	if running.Status != AgentDelegationStatusRunning {
		t.Fatalf("Status = %q, want running", running.Status)
	}

	close(provider.release)
	deadline := time.After(2 * time.Second)
	for {
		rec, err := al.delegationRecords.Get(context.Background(), result.DelegationID)
		if err != nil {
			t.Fatalf("Get(completed) error = %v", err)
		}
		if rec.Status == AgentDelegationStatusCompleted {
			if rec.Result == nil || rec.Result.Content != "async target answer" {
				t.Fatalf("Result = %#v, want async target answer", rec.Result)
			}
			return
		}
		select {
		case <-deadline:
			t.Fatalf("delegation status = %q, want completed", rec.Status)
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func TestRunAgentDelegationAsync_RejectsWhenExecutorAtCapacity(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "parent", Subagents: &config.SubagentsConfig{AllowAgents: []string{"target"}}},
		{ID: "target"},
	})
	cfg.Agents.Defaults.AsyncDelegation.MaxConcurrent = 1
	provider := &countingBlockingDelegationProvider{
		started: make(chan int, 2),
		release: make(chan struct{}),
	}
	al := NewAgentLoop(cfg, bus.NewMessageBus(), provider)
	defer al.Close()

	first, err := al.RunAgentDelegationAsync(context.Background(), AgentDelegationRequest{
		ParentAgentID: "parent",
		TargetAgentID: "target",
		Task:          "first async task",
	})
	if err != nil {
		t.Fatalf("first RunAgentDelegationAsync() error = %v", err)
	}
	select {
	case <-provider.started:
	case <-time.After(2 * time.Second):
		t.Fatal("first async delegation did not start")
	}

	rejected, err := al.RunAgentDelegationAsync(context.Background(), AgentDelegationRequest{
		ParentAgentID: "parent",
		TargetAgentID: "target",
		Task:          "second async task",
	})
	if !errors.Is(err, ErrAgentDelegationExecutorFull) {
		t.Fatalf("second RunAgentDelegationAsync() error = %v, want ErrAgentDelegationExecutorFull", err)
	}
	if rejected.DelegationID == "" {
		t.Fatal("rejected delegation ID is empty")
	}
	rec, err := al.delegationRecords.Get(context.Background(), rejected.DelegationID)
	if err != nil {
		t.Fatalf("Get(rejected) error = %v", err)
	}
	if rec.Status != AgentDelegationStatusFailed {
		t.Fatalf("rejected status = %q, want failed", rec.Status)
	}
	if rec.Error == nil || !strings.Contains(rec.Error.Message, "capacity") {
		t.Fatalf("rejected error = %#v, want capacity message", rec.Error)
	}

	close(provider.release)
	waitForDelegationStatus(t, al, first.DelegationID, AgentDelegationStatusCompleted)
}

func TestRunAgentDelegationAsync_RecordsCancelledWhenParentContextCanceledBeforeEnqueue(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "parent", Subagents: &config.SubagentsConfig{AllowAgents: []string{"target"}}},
		{ID: "target"},
	})
	al := NewAgentLoop(cfg, bus.NewMessageBus(), &recordingDelegationProvider{})
	defer al.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, err := al.RunAgentDelegationAsync(ctx, AgentDelegationRequest{
		ParentAgentID: "parent",
		TargetAgentID: "target",
		Task:          "cancel before enqueue",
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RunAgentDelegationAsync() error = %v, want context.Canceled", err)
	}
	if result.DelegationID == "" {
		t.Fatal("cancelled delegation ID is empty")
	}
	rec, err := al.delegationRecords.Get(context.Background(), result.DelegationID)
	if err != nil {
		t.Fatalf("Get(cancelled) error = %v", err)
	}
	if rec.Status != AgentDelegationStatusCancelled {
		t.Fatalf("Status = %q, want cancelled", rec.Status)
	}
}

func TestRunAgentDelegationAsync_RecordsFailureAfterEnqueue(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "parent", Subagents: &config.SubagentsConfig{AllowAgents: []string{"target"}}},
		{ID: "target"},
	})
	providerErr := errors.New("target failed after enqueue")
	al := NewAgentLoop(cfg, bus.NewMessageBus(), &failingDelegationProvider{err: providerErr})
	defer al.Close()

	result, err := al.RunAgentDelegationAsync(context.Background(), AgentDelegationRequest{
		ParentAgentID: "parent",
		TargetAgentID: "target",
		Task:          "fail after enqueue",
	})
	if err != nil {
		t.Fatalf("RunAgentDelegationAsync() error = %v", err)
	}

	rec := waitForDelegationStatus(t, al, result.DelegationID, AgentDelegationStatusFailed)
	if rec.Error == nil || !strings.Contains(rec.Error.Message, providerErr.Error()) {
		t.Fatalf("Error = %#v, want provider failure", rec.Error)
	}
}

func TestRunAgentDelegationAsync_CloseCancelsAcceptedWorkAndRejectsNewWork(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "parent", Subagents: &config.SubagentsConfig{AllowAgents: []string{"target"}}},
		{ID: "target"},
	})
	cfg.Agents.Defaults.AsyncDelegation.MaxConcurrent = 1
	provider := &countingBlockingDelegationProvider{
		started: make(chan int, 2),
		release: make(chan struct{}),
	}
	al := NewAgentLoop(cfg, bus.NewMessageBus(), provider)

	result, err := al.RunAgentDelegationAsync(context.Background(), AgentDelegationRequest{
		ParentAgentID: "parent",
		TargetAgentID: "target",
		Task:          "cancel on close",
	})
	if err != nil {
		t.Fatalf("RunAgentDelegationAsync() error = %v", err)
	}
	select {
	case <-provider.started:
	case <-time.After(2 * time.Second):
		t.Fatal("async delegation did not start")
	}

	al.Close()
	waitForDelegationStatus(t, al, result.DelegationID, AgentDelegationStatusCancelled)

	rejected, err := al.RunAgentDelegationAsync(context.Background(), AgentDelegationRequest{
		ParentAgentID: "parent",
		TargetAgentID: "target",
		Task:          "after close",
	})
	if !errors.Is(err, ErrAgentDelegationExecutorClosed) {
		t.Fatalf("RunAgentDelegationAsync() after Close error = %v, want ErrAgentDelegationExecutorClosed", err)
	}
	if rejected.DelegationID == "" {
		t.Fatal("closed-loop rejected delegation ID is empty")
	}
	rec, err := al.delegationRecords.Get(context.Background(), rejected.DelegationID)
	if err != nil {
		t.Fatalf("Get(closed rejected) error = %v", err)
	}
	if rec.Status != AgentDelegationStatusFailed {
		t.Fatalf("closed-loop rejected status = %q, want failed", rec.Status)
	}
}

func TestRunAgentDelegation_PublishesDiscordVisibilitySummariesWhenEnabled(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "parent", Subagents: &config.SubagentsConfig{AllowAgents: []string{"target"}}},
		{ID: "target"},
	})
	cfg.Channels = discordVisibilityChannels(t, true)
	msgBus := bus.NewMessageBus()
	al := NewAgentLoop(cfg, msgBus, &recordingDelegationProvider{})

	result, err := al.RunAgentDelegation(context.Background(), AgentDelegationRequest{
		ParentAgentID: "parent",
		TargetAgentID: "target",
		Task:          "Review private launch transcript: raw transcript must stay private.",
		ThreadKey:     "launch-review",
	})
	if err != nil {
		t.Fatalf("RunAgentDelegation() error = %v", err)
	}
	if result.DelegationID == "" {
		t.Fatal("DelegationID is empty")
	}

	messages := collectOutboundMessages(t, msgBus, 2)
	joined := messages[0].Content + "\n" + messages[1].Content
	for _, want := range []string{"Delegation created", "Delegation completed", result.DelegationID, "parent -> target"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("visibility summaries missing %q:\n%s", want, joined)
		}
	}
	if strings.Contains(joined, "raw transcript must stay private") || strings.Contains(joined, "target answer") {
		t.Fatalf("visibility summaries leaked task/result content:\n%s", joined)
	}
	for _, msg := range messages {
		if msg.Channel != "discord" || msg.ChatID != "visibility-channel" {
			t.Fatalf("summary target = %s:%s, want discord:visibility-channel", msg.Channel, msg.ChatID)
		}
		if got := msg.Context.Raw["visibility_summary_event"]; got == "" {
			t.Fatalf("summary missing visibility event metadata: %#v", msg.Context.Raw)
		}
	}
}

func TestRunAgentDelegation_DiscordVisibilitySummariesDisabledByDefault(t *testing.T) {
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "parent", Subagents: &config.SubagentsConfig{AllowAgents: []string{"target"}}},
		{ID: "target"},
	})
	cfg.Channels = discordVisibilityChannels(t, false)
	msgBus := bus.NewMessageBus()
	al := NewAgentLoop(cfg, msgBus, &recordingDelegationProvider{})

	_, err := al.RunAgentDelegation(context.Background(), AgentDelegationRequest{
		ParentAgentID: "parent",
		TargetAgentID: "target",
		Task:          "Review launch.",
	})
	if err != nil {
		t.Fatalf("RunAgentDelegation() error = %v", err)
	}

	select {
	case msg := <-msgBus.OutboundChan():
		t.Fatalf("unexpected visibility summary when disabled: %#v", msg)
	default:
	}
}

type blockingDelegationProvider struct {
	startOnce sync.Once
	started   chan struct{}
	release   chan struct{}
}

func (p *blockingDelegationProvider) Chat(
	ctx context.Context,
	messages []providers.Message,
	tools []providers.ToolDefinition,
	model string,
	opts map[string]any,
) (*providers.LLMResponse, error) {
	p.startOnce.Do(func() { close(p.started) })
	select {
	case <-p.release:
		return &providers.LLMResponse{Content: "async target answer", FinishReason: "stop"}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *blockingDelegationProvider) GetDefaultModel() string {
	return "provider-default"
}

type countingBlockingDelegationProvider struct {
	calls   atomic.Int64
	started chan int
	release chan struct{}
}

func (p *countingBlockingDelegationProvider) Chat(
	ctx context.Context,
	messages []providers.Message,
	tools []providers.ToolDefinition,
	model string,
	opts map[string]any,
) (*providers.LLMResponse, error) {
	call := int(p.calls.Add(1))
	p.started <- call
	select {
	case <-p.release:
		return &providers.LLMResponse{Content: "async target answer", FinishReason: "stop"}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *countingBlockingDelegationProvider) GetDefaultModel() string {
	return "provider-default"
}

func waitForDelegationStatus(
	t *testing.T,
	al *AgentLoop,
	delegationID string,
	want AgentDelegationStatus,
) AgentDelegationRecord {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		rec, err := al.delegationRecords.Get(context.Background(), delegationID)
		if err != nil {
			t.Fatalf("Get(%s) error = %v", delegationID, err)
		}
		if rec.Status == want {
			return rec
		}
		select {
		case <-deadline:
			t.Fatalf("delegation status = %q, want %q", rec.Status, want)
		default:
			time.Sleep(10 * time.Millisecond)
		}
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

func discordVisibilityChannels(t *testing.T, enabled bool) config.ChannelsConfig {
	t.Helper()
	return config.ChannelsConfig{
		"discord": &config.Channel{
			Enabled:   true,
			Type:      config.ChannelDiscord,
			AllowFrom: config.FlexibleStringSlice{"owner"},
			Settings:  config.RawNode(`{"visibility_summaries":{"enabled":` + boolString(enabled) + `,"chat_id":"visibility-channel"}}`),
		},
	}
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func collectOutboundMessages(t *testing.T, msgBus *bus.MessageBus, count int) []bus.OutboundMessage {
	t.Helper()
	out := make([]bus.OutboundMessage, 0, count)
	for len(out) < count {
		select {
		case msg := <-msgBus.OutboundChan():
			out = append(out, msg)
		case <-time.After(2 * time.Second):
			t.Fatalf("timeout waiting for outbound message %d/%d", len(out)+1, count)
		}
	}
	return out
}

func messageContents(messages []providers.Message) []string {
	contents := make([]string, 0, len(messages))
	for _, msg := range messages {
		contents = append(contents, msg.Content)
	}
	return contents
}
