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
		if query.DelegationID != "" && rec.DelegationID != query.DelegationID {
			continue
		}
		if !query.IncludePrivateAll && query.VisibleToAgentID != "" && !delegationRecordVisibleToAgent(rec, query.VisibleToAgentID) {
			continue
		}
		if query.TargetAgentID != "" && rec.TargetAgentID != query.TargetAgentID {
			continue
		}
		out = append(out, rec)
	}
	return out, nil
}

func TestDelegationStatusTool_MissingCallerIdentityCannotListRecords(t *testing.T) {
	reader := &fakeDelegationRecordReader{records: []DelegationRecord{
		{
			DelegationID:  "delegation-private",
			Status:        "running",
			ParentAgentID: "ceo",
			TargetAgentID: "cto",
			Task:          "Inspect launch readiness.",
		},
	}}
	tool := NewDelegationStatusTool(reader)

	result := tool.Execute(context.Background(), map[string]any{})

	if !result.IsError {
		t.Fatalf("expected missing caller identity to be rejected, got: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "calling agent identity is required") {
		t.Fatalf("unexpected error for missing caller identity: %s", result.ForLLM)
	}
}

func TestDelegationStatusTool_MissingCallerIdentityCannotGetByID(t *testing.T) {
	reader := &fakeDelegationRecordReader{records: []DelegationRecord{
		{
			DelegationID:  "delegation-private",
			Status:        "running",
			ParentAgentID: "ceo",
			TargetAgentID: "cto",
			Task:          "Inspect launch readiness.",
		},
	}}
	tool := NewDelegationStatusTool(reader)

	result := tool.Execute(context.Background(), map[string]any{"delegation_id": "delegation-private"})

	if !result.IsError {
		t.Fatalf("expected missing caller identity to be rejected, got: %s", result.ForLLM)
	}
	if strings.Contains(result.ForLLM, "delegation-private") {
		t.Fatalf("missing caller error leaked delegation ID: %s", result.ForLLM)
	}
}

func TestDelegationStatusTool_RunningRecordBeforeToolStartIsStale(t *testing.T) {
	startedAt := time.Date(2026, 6, 4, 19, 56, 0, 0, time.UTC)
	reader := &fakeDelegationRecordReader{records: []DelegationRecord{
		{
			DelegationID:  "delegation-stale",
			Status:        "running",
			ParentAgentID: "zehn-main",
			TargetAgentID: "li-ceo",
			CreatedAt:     startedAt.Add(-2 * time.Hour),
			UpdatedAt:     startedAt.Add(-time.Hour),
			Task:          "Execute async company lane.",
		},
	}}
	tool := NewDelegationStatusTool(reader)
	tool.startedAt = startedAt
	ctx := WithToolSessionContext(context.Background(), "zehn-main", "session", nil)

	listResult := tool.Execute(ctx, map[string]any{})
	if listResult.IsError {
		t.Fatalf("list returned error: %s", listResult.ForLLM)
	}
	if !strings.Contains(listResult.ForLLM, "status=running_stale") {
		t.Fatalf("list result = %q, want running_stale status", listResult.ForLLM)
	}

	getResult := tool.Execute(ctx, map[string]any{"delegation_id": "delegation-stale"})
	if getResult.IsError {
		t.Fatalf("get returned error: %s", getResult.ForLLM)
	}
	if !strings.Contains(getResult.ForLLM, "status=running_stale") {
		t.Fatalf("get result = %q, want running_stale status", getResult.ForLLM)
	}
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

func TestDelegationStatusTool_GetByIDAllowsVisibleRecordRoles(t *testing.T) {
	reader := &fakeDelegationRecordReader{records: []DelegationRecord{
		{
			DelegationID:      "delegation-requester",
			Status:            "completed",
			ParentAgentID:     "ceo",
			TargetAgentID:     "cto",
			RequestedBy:       "chief-of-staff",
			VisibleToAgentIDs: []string{"pm"},
			Task:              "Launch readiness review.",
		},
	}}
	tool := NewDelegationStatusTool(reader)
	for _, callerAgentID := range []string{"ceo", "cto", "chief-of-staff", "pm"} {
		t.Run(callerAgentID, func(t *testing.T) {
			ctx := WithToolSessionContext(context.Background(), callerAgentID, "session", nil)

			result := tool.Execute(ctx, map[string]any{"delegation_id": "delegation-requester"})

			if result.IsError {
				t.Fatalf("Execute() returned error for visible caller %q: %s", callerAgentID, result.ForLLM)
			}
			if !strings.Contains(result.ForLLM, "delegation-requester") {
				t.Fatalf("ForLLM missing visible delegation:\n%s", result.ForLLM)
			}
		})
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
	if strings.Contains(result.ForLLM, "delegation-private") {
		t.Fatalf("unrelated lookup leaked delegation ID: %s", result.ForLLM)
	}
}

func TestDelegationStatusTool_ListFiltersRequestedByAndExplicitVisibleRecords(t *testing.T) {
	reader := &fakeDelegationRecordReader{records: []DelegationRecord{
		{
			DelegationID:  "delegation-requested",
			Status:        "running",
			ParentAgentID: "ceo",
			TargetAgentID: "cto",
			RequestedBy:   "chief-of-staff",
			Task:          "Requested by visible caller.",
		},
		{
			DelegationID:      "delegation-explicit",
			Status:            "running",
			ParentAgentID:     "cfo",
			TargetAgentID:     "legal",
			VisibleToAgentIDs: []string{"chief-of-staff"},
			Task:              "Explicitly visible to caller.",
		},
		{
			DelegationID:  "delegation-private",
			Status:        "running",
			ParentAgentID: "cro",
			TargetAgentID: "sales",
			Task:          "Private revenue work.",
		},
	}}
	tool := NewDelegationStatusTool(reader)
	ctx := WithToolSessionContext(context.Background(), "chief-of-staff", "session", nil)

	result := tool.Execute(ctx, map[string]any{})

	if result.IsError {
		t.Fatalf("Execute() returned error: %s", result.ForLLM)
	}
	for _, want := range []string{"delegation-requested", "delegation-explicit"} {
		if !strings.Contains(result.ForLLM, want) {
			t.Fatalf("ForLLM missing visible delegation %q:\n%s", want, result.ForLLM)
		}
	}
	if strings.Contains(result.ForLLM, "delegation-private") {
		t.Fatalf("ForLLM leaked unrelated private delegation:\n%s", result.ForLLM)
	}
}

func TestDelegationStatusTool_ZehnMainCanInspectNestedDelegations(t *testing.T) {
	reader := &fakeDelegationRecordReader{records: []DelegationRecord{
		{
			DelegationID:  "delegation-nested",
			Status:        "completed",
			ParentAgentID: "li-ceo",
			TargetAgentID: "li-cto",
			Task:          "Classify release ladder.",
		},
	}}
	tool := NewDelegationStatusTool(reader)
	ctx := WithToolSessionContext(context.Background(), "zehn-main", "session", nil)

	result := tool.Execute(ctx, map[string]any{"delegation_id": "delegation-nested"})

	if result.IsError {
		t.Fatalf("Execute() returned error for zehn-main supervisor: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "delegation-nested") {
		t.Fatalf("ForLLM missing nested delegation:\n%s", result.ForLLM)
	}
}

func TestDelegationStatusTool_RuntimeSupervisorSendersCanInspectDelegations(t *testing.T) {
	reader := &fakeDelegationRecordReader{records: []DelegationRecord{
		{
			DelegationID:  "delegation-runtime-visible",
			Status:        "completed",
			ParentAgentID: "li-ceo",
			TargetAgentID: "li-coo",
			Task:          "Reconcile operating cycle.",
			CreatedAt:     time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC),
		},
	}}
	tool := NewDelegationStatusTool(reader)
	for _, callerAgentID := range []string{"heartbeat", "cron"} {
		t.Run(callerAgentID, func(t *testing.T) {
			ctx := WithToolSessionContext(context.Background(), callerAgentID, "session", nil)

			result := tool.Execute(ctx, map[string]any{})

			if result.IsError {
				t.Fatalf("Execute() returned error for runtime supervisor %q: %s", callerAgentID, result.ForLLM)
			}
			if !strings.Contains(result.ForLLM, "delegation-runtime-visible") {
				t.Fatalf("ForLLM missing runtime-visible delegation for %q:\n%s", callerAgentID, result.ForLLM)
			}
		})
	}
}

func TestDelegationStatusTool_ListCapsNewestRecords(t *testing.T) {
	records := make([]DelegationRecord, 0, 25)
	for i := range 25 {
		records = append(records, DelegationRecord{
			DelegationID:  "delegation-cap-" + string(rune('a'+i)),
			Status:        "completed",
			ParentAgentID: "li-ceo",
			TargetAgentID: "li-coo",
			Task:          "Reconcile operating cycle.",
			CreatedAt:     time.Date(2026, 6, 4, 12, i, 0, 0, time.UTC),
		})
	}
	tool := NewDelegationStatusTool(&fakeDelegationRecordReader{records: records})
	ctx := WithToolSessionContext(context.Background(), "zehn-main", "session", nil)

	result := tool.Execute(ctx, map[string]any{"target_agent_id": "li-coo"})

	if result.IsError {
		t.Fatalf("Execute() returned error: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "showing newest 20 of 25") {
		t.Fatalf("ForLLM missing capped-count header:\n%s", result.ForLLM)
	}
	if strings.Contains(result.ForLLM, "delegation-cap-a") {
		t.Fatalf("ForLLM included oldest record despite cap:\n%s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "delegation-cap-y") {
		t.Fatalf("ForLLM missing newest record:\n%s", result.ForLLM)
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
