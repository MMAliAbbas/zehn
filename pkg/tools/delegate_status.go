package tools

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

var ErrDelegationRecordNotFound = errors.New("delegation record not found")

type DelegateExecutionRequest struct {
	ParentAgentID    string
	TargetAgentID    string
	Task             string
	ThreadKey        string
	Mode             string
	Priority         string
	Due              string
	RequestedBy      string
	ApprovalRequired bool
	ArtifactRefs     []string
}

type DelegateExecutionResult struct {
	DelegationID  string
	ParentAgentID string
	TargetAgentID string
	Content       string
	Status        string
	ArtifactRefs  []string
}

type DelegationRunner interface {
	RunDelegation(ctx context.Context, req DelegateExecutionRequest) (DelegateExecutionResult, error)
	StartDelegation(ctx context.Context, req DelegateExecutionRequest) (DelegateExecutionResult, error)
}

type DelegationRecord struct {
	DelegationID  string
	Status        string
	ParentAgentID string
	TargetAgentID string
	Task          string
	ThreadKey     string
	Mode          string
	Priority      string
	RequestedBy   string
	ArtifactRefs  []string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	StartedAt     *time.Time
	CompletedAt   *time.Time
	Result        string
	Error         string
}

type DelegationRecordQuery struct {
	DelegationID      string
	VisibleToAgentID  string
	ParentAgentID     string
	TargetAgentID     string
	IncludePrivateAll bool
}

type DelegationRecordReader interface {
	GetDelegationRecord(ctx context.Context, delegationID string) (DelegationRecord, error)
	ListDelegationRecords(ctx context.Context, query DelegationRecordQuery) ([]DelegationRecord, error)
}

type DelegationStatusTool struct {
	reader DelegationRecordReader
}

func NewDelegationStatusTool(reader DelegationRecordReader) *DelegationStatusTool {
	return &DelegationStatusTool{reader: reader}
}

func (t *DelegationStatusTool) Name() string {
	return "delegation_status"
}

func (t *DelegationStatusTool) Description() string {
	return "List visible delegation records or inspect one delegation by delegation_id. Records are scoped to the calling agent unless private-all visibility is configured by the host."
}

func (t *DelegationStatusTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"delegation_id": map[string]any{
				"type":        "string",
				"description": "Optional delegation ID to inspect.",
			},
			"target_agent_id": map[string]any{
				"type":        "string",
				"description": "Optional target agent filter for visible delegations.",
			},
		},
		"required": []string{},
	}
}

func (t *DelegationStatusTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	if t.reader == nil {
		return ErrorResult("delegation_status error: delegation record reader not configured")
	}
	callerAgentID := strings.TrimSpace(ToolAgentID(ctx))
	delegationID := stringArg(args, "delegation_id")
	targetAgentID := stringArg(args, "target_agent_id")

	if delegationID != "" {
		rec, err := t.reader.GetDelegationRecord(ctx, delegationID)
		if err != nil {
			return ErrorResult(fmt.Sprintf("delegation_status error: delegation %q not found", delegationID)).WithError(err)
		}
		if callerAgentID != "" && rec.ParentAgentID != callerAgentID && rec.TargetAgentID != callerAgentID {
			err := fmt.Errorf("delegation %q is not visible to agent %q", delegationID, callerAgentID)
			return ErrorResult("delegation_status error: delegation not found").WithError(err)
		}
		return NewToolResult(formatDelegationRecord(rec))
	}

	records, err := t.reader.ListDelegationRecords(ctx, DelegationRecordQuery{
		VisibleToAgentID: callerAgentID,
		TargetAgentID:    targetAgentID,
	})
	if err != nil {
		return ErrorResult(fmt.Sprintf("delegation_status error: %v", err)).WithError(err)
	}
	if len(records) == 0 {
		return NewToolResult("No delegations found.")
	}
	return NewToolResult(formatDelegationRecords("Delegation status", records))
}

type DelegationInboxTool struct {
	reader DelegationRecordReader
}

func NewDelegationInboxTool(reader DelegationRecordReader) *DelegationInboxTool {
	return &DelegationInboxTool{reader: reader}
}

func (t *DelegationInboxTool) Name() string {
	return "delegation_inbox"
}

func (t *DelegationInboxTool) Description() string {
	return "List delegation work assigned to the calling target agent."
}

func (t *DelegationInboxTool) Parameters() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
		"required":   []string{},
	}
}

func (t *DelegationInboxTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	if t.reader == nil {
		return ErrorResult("delegation_inbox error: delegation record reader not configured")
	}
	callerAgentID := strings.TrimSpace(ToolAgentID(ctx))
	if callerAgentID == "" {
		return ErrorResult("delegation_inbox error: calling agent identity is required")
	}
	records, err := t.reader.ListDelegationRecords(ctx, DelegationRecordQuery{
		VisibleToAgentID: callerAgentID,
		TargetAgentID:    callerAgentID,
	})
	if err != nil {
		return ErrorResult(fmt.Sprintf("delegation_inbox error: %v", err)).WithError(err)
	}
	if len(records) == 0 {
		return NewToolResult("Delegation inbox is empty.")
	}
	return NewToolResult(formatDelegationRecords("Delegation inbox", records))
}

func formatDelegationRecords(title string, records []DelegationRecord) string {
	records = append([]DelegationRecord(nil), records...)
	slices.SortFunc(records, func(a, b DelegationRecord) int {
		if cmpCreated := cmp.Compare(a.CreatedAt.UnixNano(), b.CreatedAt.UnixNano()); cmpCreated != 0 {
			return cmpCreated
		}
		return cmp.Compare(a.DelegationID, b.DelegationID)
	})
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s (%d total):", title, len(records)))
	for _, rec := range records {
		sb.WriteString("\n")
		sb.WriteString(formatDelegationRecord(rec))
	}
	return sb.String()
}

func formatDelegationRecord(rec DelegationRecord) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s] status=%s parent=%s target=%s", rec.DelegationID, rec.Status, rec.ParentAgentID, rec.TargetAgentID))
	if !rec.CreatedAt.IsZero() {
		sb.WriteString(" created=")
		sb.WriteString(rec.CreatedAt.UTC().Format("2006-01-02 15:04:05 UTC"))
	}
	if rec.Task != "" {
		sb.WriteString("\n  task:   ")
		sb.WriteString(delegateCompact(rec.Task))
	}
	if rec.Result != "" {
		sb.WriteString("\n  result: ")
		sb.WriteString(delegateCompact(rec.Result))
	}
	if rec.Error != "" {
		sb.WriteString("\n  error:  ")
		sb.WriteString(delegateCompact(rec.Error))
	}
	return sb.String()
}

func stringArg(args map[string]any, name string) string {
	raw, _ := args[name].(string)
	return strings.TrimSpace(raw)
}
