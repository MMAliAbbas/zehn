package tools

import (
	"context"
	"strings"
	"testing"
	"time"
)

type fakeDelegationRecordReader struct {
	records []DelegationRecord
}

func (r *fakeDelegationRecordReader) GetDelegationRecord(ctx context.Context, delegationID string) (DelegationRecord, error) {
	for _, rec := range r.records {
		if rec.DelegationID == delegationID {
			return rec, nil
		}
	}
	return DelegationRecord{}, ErrDelegationRecordNotFound
}

func (r *fakeDelegationRecordReader) ListDelegationRecords(ctx context.Context, query DelegationRecordQuery) ([]DelegationRecord, error) {
	var out []DelegationRecord
	for _, rec := range r.records {
		if query.VisibleToAgentID != "" && rec.ParentAgentID != query.VisibleToAgentID && rec.TargetAgentID != query.VisibleToAgentID {
			continue
		}
		if query.TargetAgentID != "" && rec.TargetAgentID != query.TargetAgentID {
			continue
		}
		out = append(out, rec)
	}
	return out, nil
}

func TestDelegationStatusTool_ListScopesRecordsToCallingAgent(t *testing.T) {
	reader := &fakeDelegationRecordReader{records: []DelegationRecord{
		{
			DelegationID:  "delegation-owned",
			Status:        "running",
			ParentAgentID: "ceo",
			TargetAgentID: "cto",
			Task:          "Inspect launch readiness.",
			CreatedAt:     time.Date(2026, 5, 6, 10, 0, 0, 0, time.UTC),
		},
		{
			DelegationID:  "delegation-private",
			Status:        "completed",
			ParentAgentID: "cfo",
			TargetAgentID: "legal",
			Task:          "Private finance review.",
			CreatedAt:     time.Date(2026, 5, 6, 11, 0, 0, 0, time.UTC),
		},
	}}
	tool := NewDelegationStatusTool(reader)
	ctx := WithToolSessionContext(context.Background(), "ceo", "session", nil)

	result := tool.Execute(ctx, map[string]any{})

	if result.IsError {
		t.Fatalf("Execute() returned error: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "delegation-owned") {
		t.Fatalf("ForLLM missing owned delegation:\n%s", result.ForLLM)
	}
	if strings.Contains(result.ForLLM, "delegation-private") {
		t.Fatalf("ForLLM leaked unrelated private delegation:\n%s", result.ForLLM)
	}
}

func TestDelegationStatusTool_GetByIDRejectsUnrelatedRecord(t *testing.T) {
	reader := &fakeDelegationRecordReader{records: []DelegationRecord{
		{
			DelegationID:  "delegation-private",
			Status:        "completed",
			ParentAgentID: "cfo",
			TargetAgentID: "legal",
			Task:          "Private finance review.",
		},
	}}
	tool := NewDelegationStatusTool(reader)
	ctx := WithToolSessionContext(context.Background(), "ceo", "session", nil)

	result := tool.Execute(ctx, map[string]any{"delegation_id": "delegation-private"})

	if !result.IsError {
		t.Fatalf("expected unrelated record lookup to be rejected, got: %s", result.ForLLM)
	}
}

func TestDelegationInboxTool_ListsOnlyTargetAgentWork(t *testing.T) {
	reader := &fakeDelegationRecordReader{records: []DelegationRecord{
		{
			DelegationID:  "delegation-for-cto",
			Status:        "requested",
			ParentAgentID: "ceo",
			TargetAgentID: "cto",
			Task:          "Prepare engineering plan.",
		},
		{
			DelegationID:  "delegation-for-cro",
			Status:        "requested",
			ParentAgentID: "ceo",
			TargetAgentID: "cro",
			Task:          "Prepare revenue plan.",
		},
	}}
	tool := NewDelegationInboxTool(reader)
	ctx := WithToolSessionContext(context.Background(), "cto", "session", nil)

	result := tool.Execute(ctx, map[string]any{})

	if result.IsError {
		t.Fatalf("Execute() returned error: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "delegation-for-cto") {
		t.Fatalf("ForLLM missing target delegation:\n%s", result.ForLLM)
	}
	if strings.Contains(result.ForLLM, "delegation-for-cro") {
		t.Fatalf("ForLLM leaked another agent's inbox:\n%s", result.ForLLM)
	}
}
