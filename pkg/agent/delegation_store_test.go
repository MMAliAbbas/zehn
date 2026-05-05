package agent

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/providers"
	"github.com/sipeed/picoclaw/pkg/session"
)

func TestDelegationRecordStore_WritesAtomicallyAndReadsAfterRestart(t *testing.T) {
	dir := t.TempDir()
	secret := "scope-secret-12345"
	store := NewDelegationRecordStore(dir, strings.NewReplacer(secret, "[FILTERED]").Replace)
	store.now = fixedDelegationClock(
		time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 5, 10, 1, 0, 0, time.UTC),
		time.Date(2026, 5, 5, 10, 2, 0, 0, time.UTC),
	)

	rec, err := store.Requested(context.Background(), AgentDelegationRequest{
		ParentAgentID: "Parent/Agent",
		TargetAgentID: "Target Agent",
		Task:          "review the rollout",
		ThreadKey:     "Launch Review",
		ArtifactRefs:  []string{"issue:123", " /tmp/report.md "},
	})
	if err != nil {
		t.Fatalf("Requested() error = %v", err)
	}
	if rec.Status != AgentDelegationStatusRequested {
		t.Fatalf("Status = %q, want requested", rec.Status)
	}
	if strings.Contains(rec.DelegationID, "/") || strings.Contains(rec.DelegationID, "\\") {
		t.Fatalf("DelegationID = %q, want filesystem-safe ID", rec.DelegationID)
	}

	if err := store.Running(context.Background(), rec.DelegationID); err != nil {
		t.Fatalf("Running() error = %v", err)
	}
	if err := store.Completed(context.Background(), rec.DelegationID, AgentDelegationResult{
		ParentAgentID: "parent-agent",
		TargetAgentID: "target-agent",
		SessionKey:    "internal:delegation:parent:target:launch-" + secret,
		SessionScope: &session.SessionScope{
			Version:    session.ScopeVersionV1,
			AgentID:    "target",
			Channel:    "internal",
			Dimensions: []string{"delegation"},
			Values:     map[string]string{"delegation": "parent:target:launch-" + secret},
		},
		Content: "ship with guardrails",
		Status:  TurnEndStatusCompleted,
	}); err != nil {
		t.Fatalf("Completed() error = %v", err)
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("record store files = %d, want 1: %v", len(files), files)
	}
	if strings.HasPrefix(files[0].Name(), ".tmp-") {
		t.Fatalf("record file = %q, want final JSON file", files[0].Name())
	}

	restarted := NewDelegationRecordStore(dir, nil)
	loaded, err := restarted.Get(context.Background(), rec.DelegationID)
	if err != nil {
		t.Fatalf("Get() after restart error = %v", err)
	}
	if loaded.Status != AgentDelegationStatusCompleted {
		t.Fatalf("loaded Status = %q, want completed", loaded.Status)
	}
	if loaded.Result == nil || loaded.Result.Content != "ship with guardrails" {
		t.Fatalf("loaded Result = %#v, want completed content", loaded.Result)
	}
	if loaded.Result.SessionKey != "internal:delegation:parent:target:launch-[FILTERED]" {
		t.Fatalf("loaded session key = %q", loaded.Result.SessionKey)
	}
	data, err := os.ReadFile(filepath.Join(dir, loaded.Filename()))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.Contains(string(data), secret) {
		t.Fatalf("record leaked secret: %s", data)
	}
}

func TestRunAgentDelegation_RecordStoreCapturesFailureAndRedactsSecrets(t *testing.T) {
	tmpDir := t.TempDir()
	secret := "sk-delegation-secret-12345"
	cfg := delegationConfigWithAgents(t, []config.AgentConfig{
		{ID: "parent", Subagents: &config.SubagentsConfig{AllowAgents: []string{"target"}}},
		{ID: "target"},
	})
	cfg.Agents.Defaults.Workspace = tmpDir
	cfg.ModelList = config.SecureModelList{
		&config.ModelConfig{
			ModelName: "test-model",
			APIKeys:   config.SecureStrings{config.NewSecureString(secret)},
		},
	}

	providerErr := errors.New("provider rejected " + secret)
	al := NewAgentLoop(cfg, bus.NewMessageBus(), &failingDelegationProvider{err: providerErr})

	result, err := al.RunAgentDelegation(context.Background(), AgentDelegationRequest{
		ParentAgentID: "parent",
		TargetAgentID: "target",
		Task:          "use " + secret + " to check this",
		ThreadKey:     "secret-review-" + secret,
	})
	if err == nil {
		t.Fatal("RunAgentDelegation() error = nil, want provider error")
	}
	if result.DelegationID == "" {
		t.Fatal("result DelegationID is empty")
	}

	loaded, err := al.delegationRecords.Get(context.Background(), result.DelegationID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if loaded.Status != AgentDelegationStatusFailed {
		t.Fatalf("Status = %q, want failed", loaded.Status)
	}
	if loaded.Error == nil || !strings.Contains(loaded.Error.Message, "[FILTERED]") {
		t.Fatalf("Error = %#v, want redacted evidence", loaded.Error)
	}

	data, err := os.ReadFile(filepath.Join(al.delegationRecords.dir, loaded.Filename()))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.Contains(string(data), secret) {
		t.Fatalf("record leaked secret: %s", data)
	}
	if strings.Contains(string(data), "provider rejected") == false {
		t.Fatalf("record missing useful failure evidence: %s", data)
	}
}

func TestDelegationRecordStore_CancelledStatus(t *testing.T) {
	store := NewDelegationRecordStore(t.TempDir(), nil)
	rec, err := store.Requested(context.Background(), AgentDelegationRequest{
		ParentAgentID: "parent",
		TargetAgentID: "target",
		Task:          "prepare a report",
	})
	if err != nil {
		t.Fatalf("Requested() error = %v", err)
	}

	cancelErr := context.Canceled
	if err := store.Failed(context.Background(), rec.DelegationID, cancelErr); err != nil {
		t.Fatalf("Failed(context.Canceled) error = %v", err)
	}
	loaded, err := store.Get(context.Background(), rec.DelegationID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if loaded.Status != AgentDelegationStatusCancelled {
		t.Fatalf("Status = %q, want cancelled", loaded.Status)
	}
	if loaded.Error == nil || loaded.Error.Message == "" {
		t.Fatalf("Error = %#v, want cancellation evidence", loaded.Error)
	}
}

func TestDelegationRecordStore_ListScopesByVisibleAgentAndTarget(t *testing.T) {
	store := NewDelegationRecordStore(t.TempDir(), nil)
	store.now = fixedDelegationClock(
		time.Date(2026, 5, 6, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 6, 10, 1, 0, 0, time.UTC),
		time.Date(2026, 5, 6, 10, 2, 0, 0, time.UTC),
	)
	for _, req := range []AgentDelegationRequest{
		{ParentAgentID: "ceo", TargetAgentID: "cto", Task: "engineering plan"},
		{ParentAgentID: "ceo", TargetAgentID: "cro", Task: "revenue plan"},
		{ParentAgentID: "cfo", TargetAgentID: "legal", Task: "private finance review"},
	} {
		if _, err := store.Requested(context.Background(), req); err != nil {
			t.Fatalf("Requested() error = %v", err)
		}
	}

	visible, err := store.List(context.Background(), AgentDelegationRecordQuery{VisibleToAgentID: "ceo"})
	if err != nil {
		t.Fatalf("List(visible) error = %v", err)
	}
	if len(visible) != 2 {
		t.Fatalf("visible records = %d, want 2: %#v", len(visible), visible)
	}
	for _, rec := range visible {
		if rec.ParentAgentID != "ceo" && rec.TargetAgentID != "ceo" {
			t.Fatalf("visible record leaked unrelated delegation: %#v", rec)
		}
	}

	inbox, err := store.List(context.Background(), AgentDelegationRecordQuery{
		VisibleToAgentID: "cto",
		TargetAgentID:    "cto",
	})
	if err != nil {
		t.Fatalf("List(inbox) error = %v", err)
	}
	if len(inbox) != 1 || inbox[0].TargetAgentID != "cto" {
		t.Fatalf("inbox records = %#v, want one cto record", inbox)
	}
}

func TestDelegationRecordStore_RejectsUnsafeRecordIDs(t *testing.T) {
	store := NewDelegationRecordStore(t.TempDir(), nil)
	if _, err := store.Get(context.Background(), "../escape"); err == nil {
		t.Fatal("Get(../escape) error = nil, want invalid request")
	}
}

func TestDelegationRecordStore_DefaultPathUsesWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.Agents.Defaults.Workspace = filepath.Join(tmpDir, "workspace")
	al := NewAgentLoop(cfg, bus.NewMessageBus(), &mockProvider{})

	want := filepath.Join(cfg.Agents.Defaults.Workspace, "delegations")
	if al.delegationRecords == nil {
		t.Fatal("delegationRecords is nil")
	}
	if al.delegationRecords.dir != want {
		t.Fatalf("delegation store dir = %q, want %q", al.delegationRecords.dir, want)
	}
}

type failingDelegationProvider struct {
	err error
}

func (p *failingDelegationProvider) Chat(
	ctx context.Context,
	messages []providers.Message,
	tools []providers.ToolDefinition,
	model string,
	opts map[string]any,
) (*providers.LLMResponse, error) {
	return nil, p.err
}

func (p *failingDelegationProvider) GetDefaultModel() string {
	return "provider-default"
}

func fixedDelegationClock(times ...time.Time) func() time.Time {
	i := 0
	return func() time.Time {
		if i >= len(times) {
			return times[len(times)-1]
		}
		t := times[i]
		i++
		return t
	}
}

func TestDelegationRecord_JSONSchemaIncludesExpectedFields(t *testing.T) {
	rec := AgentDelegationRecord{
		DelegationID:  "delegation-1",
		ParentAgentID: "parent",
		TargetAgentID: "target",
		Status:        AgentDelegationStatusRequested,
		Request: AgentDelegationRecordRequest{
			Task:         "task",
			ThreadKey:    "thread",
			ArtifactRefs: []string{"issue:1"},
		},
		CreatedAt: time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC),
	}
	data, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	for _, field := range []string{
		`"delegation_id"`,
		`"parent_agent_id"`,
		`"target_agent_id"`,
		`"request"`,
		`"status"`,
		`"created_at"`,
		`"updated_at"`,
		`"artifact_refs"`,
	} {
		if !strings.Contains(string(data), field) {
			t.Fatalf("record JSON missing %s: %s", field, data)
		}
	}
}
