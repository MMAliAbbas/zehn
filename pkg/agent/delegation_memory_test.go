package agent

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/sipeed/picoclaw/pkg/config"
)

func TestDelegationYaadMemorySuccessRecordedLocally(t *testing.T) {
	store := NewDelegationRecordStore(t.TempDir(), nil)
	store.now = fixedDelegationClock(
		time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 6, 12, 1, 0, 0, time.UTC),
		time.Date(2026, 5, 6, 12, 2, 0, 0, time.UTC),
	)

	rec, err := store.Requested(context.Background(), AgentDelegationRequest{
		ParentAgentID: "parent",
		TargetAgentID: "target",
		Task:          "review durable memory behavior",
	})
	if err != nil {
		t.Fatalf("Requested() error = %v", err)
	}

	write := AgentDelegationMemoryWrite{
		Provider: "yaad",
		Status:   AgentDelegationMemoryStatusWritten,
		MemoryID: "mem_123",
	}
	if err := store.RecordMemoryWrite(context.Background(), rec.DelegationID, write); err != nil {
		t.Fatalf("RecordMemoryWrite() error = %v", err)
	}

	loaded, err := store.Get(context.Background(), rec.DelegationID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if loaded.DurableMemory == nil {
		t.Fatal("DurableMemory is nil, want Yaad write metadata")
	}
	if loaded.DurableMemory.Status != AgentDelegationMemoryStatusWritten {
		t.Fatalf("DurableMemory.Status = %q, want written", loaded.DurableMemory.Status)
	}
	if loaded.DurableMemory.MemoryID != "mem_123" {
		t.Fatalf("DurableMemory.MemoryID = %q, want mem_123", loaded.DurableMemory.MemoryID)
	}
	if loaded.DurableMemory.UpdatedAt.IsZero() {
		t.Fatal("DurableMemory.UpdatedAt is zero")
	}
}

func TestDelegationYaadMemoryFailureDoesNotLoseResult(t *testing.T) {
	cfg := delegationConfigWithAgents(t, nil)
	al := NewAgentLoop(cfg, nil, &mockProvider{})
	al.delegationRecords = NewDelegationRecordStore(t.TempDir(), nil)

	rec, err := al.delegationRecords.Requested(context.Background(), AgentDelegationRequest{
		ParentAgentID: "parent",
		TargetAgentID: "target",
		Task:          "produce result before Yaad outage",
	})
	if err != nil {
		t.Fatalf("Requested() error = %v", err)
	}
	result := AgentDelegationResult{
		DelegationID:  rec.DelegationID,
		ParentAgentID: "parent",
		TargetAgentID: "target",
		Content:       "completed answer",
		Status:        TurnEndStatusCompleted,
	}
	if err := al.delegationRecords.Completed(context.Background(), rec.DelegationID, result); err != nil {
		t.Fatalf("Completed() error = %v", err)
	}

	writeErr := errors.New("yaad unavailable")
	restoreWriter := overrideDelegationMemoryWriterForTest(t, &recordingDelegationMemoryWriter{err: writeErr})
	defer restoreWriter()
	restoreStrict := overrideDelegationMemoryStrictForTest(t, false)
	defer restoreStrict()

	if err := al.persistDelegationMemory(context.Background(), rec.DelegationID); err != nil {
		t.Fatalf("persistDelegationMemory() error = %v, want non-strict fallback", err)
	}

	loaded, err := al.delegationRecords.Get(context.Background(), rec.DelegationID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if loaded.Status != AgentDelegationStatusCompleted {
		t.Fatalf("Status = %q, want completed", loaded.Status)
	}
	if loaded.Result == nil || loaded.Result.Content != "completed answer" {
		t.Fatalf("Result = %#v, want completed answer preserved", loaded.Result)
	}
	if loaded.DurableMemory == nil || loaded.DurableMemory.Status != AgentDelegationMemoryStatusFailed {
		t.Fatalf("DurableMemory = %#v, want failed write metadata", loaded.DurableMemory)
	}
	if !strings.Contains(loaded.DurableMemory.Error, "yaad unavailable") {
		t.Fatalf("DurableMemory.Error = %q, want Yaad outage evidence", loaded.DurableMemory.Error)
	}
}

func TestDelegationYaadMemoryReceivesRedactedRecord(t *testing.T) {
	secret := "sk-delegation-memory-secret"
	cfg := delegationConfigWithAgents(t, nil)
	cfg.ModelList = append(cfg.ModelList, &config.ModelConfig{
		ModelName: "test-model",
		APIKeys:   config.SecureStrings{config.NewSecureString(secret)},
	})
	al := NewAgentLoop(cfg, nil, &mockProvider{})
	al.delegationRecords = NewDelegationRecordStore(t.TempDir(), delegationRecordRedactor(cfg))

	rec, err := al.delegationRecords.Requested(context.Background(), AgentDelegationRequest{
		ParentAgentID: "parent",
		TargetAgentID: "target",
		Task:          "inspect " + secret,
		ThreadKey:     "thread-" + secret,
		ArtifactRefs:  []string{"ref-" + secret},
	})
	if err != nil {
		t.Fatalf("Requested() error = %v", err)
	}
	if err := al.delegationRecords.Completed(context.Background(), rec.DelegationID, AgentDelegationResult{
		Content: "result includes " + secret,
		Status:  TurnEndStatusCompleted,
	}); err != nil {
		t.Fatalf("Completed() error = %v", err)
	}

	writer := &recordingDelegationMemoryWriter{}
	restoreWriter := overrideDelegationMemoryWriterForTest(t, writer)
	defer restoreWriter()
	restoreStrict := overrideDelegationMemoryStrictForTest(t, false)
	defer restoreStrict()

	if err := al.persistDelegationMemory(context.Background(), rec.DelegationID); err != nil {
		t.Fatalf("persistDelegationMemory() error = %v", err)
	}

	writer.mu.Lock()
	written := writer.records
	writer.mu.Unlock()
	if len(written) != 1 {
		t.Fatalf("Yaad writes = %d, want 1", len(written))
	}
	data := mustMarshalDelegationRecordForTest(t, written[0])
	if strings.Contains(string(data), secret) {
		t.Fatalf("Yaad record leaked secret: %s", data)
	}
	if !strings.Contains(string(data), "[FILTERED]") {
		t.Fatalf("Yaad record missing redaction evidence: %s", data)
	}
}

func TestDelegationYaadMemoryMCPAdapterCallsMemoryAdd(t *testing.T) {
	caller := &recordingDelegationMCPCaller{
		result: &sdkmcp.CallToolResult{
			Content: []sdkmcp.Content{
				&sdkmcp.TextContent{Text: `{"memory_id":"mem_456"}`},
			},
		},
	}
	writer := NewYaadDelegationMemoryWriter(caller, "yaad")
	rec := AgentDelegationRecord{
		DelegationID:  "delegation-1",
		ParentAgentID: "parent",
		TargetAgentID: "target",
		Status:        AgentDelegationStatusCompleted,
		Request: AgentDelegationRecordRequest{
			Task:         "request body",
			ArtifactRefs: []string{"issue:123"},
		},
		Result: &AgentDelegationRecordResult{
			Content: "decision: proceed; follow-up: verify rollout",
			Status:  TurnEndStatusCompleted,
		},
		CreatedAt: time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 6, 12, 1, 0, 0, time.UTC),
	}

	write, err := writer.WriteDelegationMemory(context.Background(), rec)
	if err != nil {
		t.Fatalf("WriteDelegationMemory() error = %v", err)
	}
	if write.Status != AgentDelegationMemoryStatusWritten {
		t.Fatalf("write.Status = %q, want written", write.Status)
	}
	if write.MemoryID != "mem_456" {
		t.Fatalf("write.MemoryID = %q, want mem_456", write.MemoryID)
	}

	caller.mu.Lock()
	defer caller.mu.Unlock()
	if caller.serverName != "yaad" || caller.toolName != "memory_add" {
		t.Fatalf("CallTool() = %s/%s, want yaad/memory_add", caller.serverName, caller.toolName)
	}
	rawContent, ok := caller.arguments["raw_content"].(string)
	if !ok {
		t.Fatalf("raw_content = %#v, want string", caller.arguments["raw_content"])
	}
	for _, want := range []string{
		`"request"`,
		`"result"`,
		`"status":"completed"`,
		`"decisions"`,
		`"follow_ups"`,
		"issue:123",
	} {
		if !strings.Contains(rawContent, want) {
			t.Fatalf("raw_content missing %q: %s", want, rawContent)
		}
	}
}

type recordingDelegationMemoryWriter struct {
	mu      sync.Mutex
	records []AgentDelegationRecord
	err     error
}

func (w *recordingDelegationMemoryWriter) WriteDelegationMemory(
	ctx context.Context,
	rec AgentDelegationRecord,
) (AgentDelegationMemoryWrite, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.records = append(w.records, rec)
	if w.err != nil {
		return AgentDelegationMemoryWrite{}, w.err
	}
	return AgentDelegationMemoryWrite{
		Provider: "yaad",
		Status:   AgentDelegationMemoryStatusWritten,
		MemoryID: "fake-memory-id",
	}, nil
}

type recordingDelegationMCPCaller struct {
	mu         sync.Mutex
	serverName string
	toolName   string
	arguments  map[string]any
	result     *sdkmcp.CallToolResult
	err        error
}

func (c *recordingDelegationMCPCaller) CallTool(
	ctx context.Context,
	serverName, toolName string,
	arguments map[string]any,
) (*sdkmcp.CallToolResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.serverName = serverName
	c.toolName = toolName
	c.arguments = arguments
	if c.err != nil {
		return nil, c.err
	}
	return c.result, nil
}

func overrideDelegationMemoryWriterForTest(t *testing.T, writer DelegationMemoryWriter) func() {
	t.Helper()
	previous := delegationMemoryWriterForAgentLoop
	delegationMemoryWriterForAgentLoop = func(al *AgentLoop) DelegationMemoryWriter {
		return writer
	}
	return func() {
		delegationMemoryWriterForAgentLoop = previous
	}
}

func overrideDelegationMemoryStrictForTest(t *testing.T, strict bool) func() {
	t.Helper()
	previous := delegationMemoryStrictForAgentLoop
	delegationMemoryStrictForAgentLoop = func(al *AgentLoop) bool {
		return strict
	}
	return func() {
		delegationMemoryStrictForAgentLoop = previous
	}
}

func mustMarshalDelegationRecordForTest(t *testing.T, rec AgentDelegationRecord) []byte {
	t.Helper()
	data, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	return data
}
