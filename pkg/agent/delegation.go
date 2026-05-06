package agent

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/routing"
	"github.com/sipeed/picoclaw/pkg/session"
	"github.com/sipeed/picoclaw/pkg/tools"
)

var (
	ErrAgentDelegationPermissionDenied = errors.New("agent delegation permission denied")
	ErrAgentDelegationTargetNotFound   = errors.New("agent delegation target not found")
	ErrAgentDelegationInvalidRequest   = errors.New("invalid agent delegation request")
	ErrAgentDelegationExecutorFull     = errors.New("agent delegation executor at capacity")
	ErrAgentDelegationExecutorClosed   = errors.New("agent delegation executor closed")
)

// AgentDelegationRequest describes a single delegated turn from one configured
// agent to another. It is intentionally source-level: public tools and durable
// queues can wrap it without overloading the existing spawn/subagent behavior.
type AgentDelegationRequest struct {
	ParentAgentID     string
	TargetAgentID     string
	Task              string
	ThreadKey         string
	Mode              string
	Priority          string
	DueAt             *time.Time
	RequestedBy       string
	VisibleToAgentIDs []string
	ApprovalRequired  bool
	ArtifactRefs      []string
}

// AgentDelegationResult is the synchronous outcome of a delegated target-agent
// turn.
type AgentDelegationResult struct {
	DelegationID  string
	ParentAgentID string
	TargetAgentID string
	SessionKey    string
	SessionScope  *session.SessionScope
	Content       string
	Status        TurnEndStatus
	ArtifactRefs  []string
}

// RunAgentDelegation runs req.Task through the real configured target
// AgentInstance, using a private internal delegation session scope.
func (al *AgentLoop) RunAgentDelegation(
	ctx context.Context,
	req AgentDelegationRequest,
) (AgentDelegationResult, error) {
	record, req, result, err := al.prepareAgentDelegation(ctx, req, "sync")
	if err != nil {
		return result, err
	}
	return al.runPreparedAgentDelegation(ctx, req, record, result)
}

func (al *AgentLoop) RunAgentDelegationAsync(
	ctx context.Context,
	req AgentDelegationRequest,
) (AgentDelegationResult, error) {
	recordCtx := ctx
	if err := ctx.Err(); err != nil {
		recordCtx = context.WithoutCancel(ctx)
	}
	record, req, result, err := al.prepareAgentDelegation(recordCtx, req, "async")
	if err != nil {
		return result, err
	}
	if err := al.asyncDelegations.Submit(ctx, func(runCtx context.Context) {
		_, _ = al.runPreparedAgentDelegation(runCtx, req, record, result)
	}); err != nil {
		_ = al.delegationRecords.Failed(context.Background(), record.DelegationID, err)
		_ = al.persistDelegationMemory(context.Background(), record.DelegationID)
		al.publishDelegationBlockerSummary(context.Background(), record, err)
		return result, err
	}
	return result, nil
}

func (al *AgentLoop) RunDelegation(
	ctx context.Context,
	req tools.DelegateExecutionRequest,
) (tools.DelegateExecutionResult, error) {
	result, err := al.RunAgentDelegation(ctx, agentDelegationRequestFromTool(req, "sync"))
	return toolDelegationResult(result), err
}

func (al *AgentLoop) StartDelegation(
	ctx context.Context,
	req tools.DelegateExecutionRequest,
) (tools.DelegateExecutionResult, error) {
	result, err := al.RunAgentDelegationAsync(ctx, agentDelegationRequestFromTool(req, "async"))
	return toolDelegationResult(result), err
}

func (al *AgentLoop) GetDelegationRecord(ctx context.Context, delegationID string) (tools.DelegationRecord, error) {
	if al == nil || al.delegationRecords == nil {
		return tools.DelegationRecord{}, tools.ErrDelegationRecordNotFound
	}
	rec, err := al.delegationRecords.Get(ctx, delegationID)
	if err != nil {
		return tools.DelegationRecord{}, err
	}
	return toolDelegationRecord(rec), nil
}

func (al *AgentLoop) ListDelegationRecords(
	ctx context.Context,
	query tools.DelegationRecordQuery,
) ([]tools.DelegationRecord, error) {
	if al == nil || al.delegationRecords == nil {
		return nil, nil
	}
	records, err := al.delegationRecords.List(ctx, AgentDelegationRecordQuery{
		DelegationID:      query.DelegationID,
		VisibleToAgentID:  routing.NormalizeAgentID(query.VisibleToAgentID),
		ParentAgentID:     routing.NormalizeAgentID(query.ParentAgentID),
		TargetAgentID:     routing.NormalizeAgentID(query.TargetAgentID),
		IncludePrivateAll: query.IncludePrivateAll,
	})
	if err != nil {
		return nil, err
	}
	out := make([]tools.DelegationRecord, 0, len(records))
	for _, rec := range records {
		out = append(out, toolDelegationRecord(rec))
	}
	return out, nil
}

func (al *AgentLoop) prepareAgentDelegation(
	ctx context.Context,
	req AgentDelegationRequest,
	mode string,
) (AgentDelegationRecord, AgentDelegationRequest, AgentDelegationResult, error) {
	if al == nil || al.registry == nil {
		err := fmt.Errorf(
			"%w: agent loop is not initialized",
			ErrAgentDelegationInvalidRequest,
		)
		return AgentDelegationRecord{}, req, AgentDelegationResult{}, err
	}

	parentAgentID := routing.NormalizeAgentID(req.ParentAgentID)
	targetAgentID := routing.NormalizeAgentID(req.TargetAgentID)
	task := strings.TrimSpace(req.Task)
	if parentAgentID == "" || targetAgentID == "" || task == "" {
		err := fmt.Errorf(
			"%w: parent agent, target agent, and task are required",
			ErrAgentDelegationInvalidRequest,
		)
		return AgentDelegationRecord{}, req, AgentDelegationResult{}, err
	}
	req.ParentAgentID = parentAgentID
	req.TargetAgentID = targetAgentID
	req.Task = task
	req.Mode = strings.TrimSpace(mode)
	if req.Mode == "" {
		req.Mode = "sync"
	}

	record, err := al.delegationRecords.Requested(ctx, req)
	if err != nil {
		return AgentDelegationRecord{}, req, AgentDelegationResult{}, err
	}
	al.publishDelegationCreatedSummary(ctx, record)
	if err := al.persistDelegationMemory(ctx, record.DelegationID); err != nil {
		return record, req, AgentDelegationResult{}, err
	}
	result := AgentDelegationResult{
		DelegationID:  record.DelegationID,
		ParentAgentID: parentAgentID,
		TargetAgentID: targetAgentID,
		ArtifactRefs:  compactDelegationRefs(req.ArtifactRefs),
	}

	if !al.registry.CanSpawnSubagent(parentAgentID, targetAgentID) {
		err := fmt.Errorf(
			"%w: parent %q cannot delegate to target %q",
			ErrAgentDelegationPermissionDenied,
			parentAgentID,
			targetAgentID,
		)
		_ = al.delegationRecords.Failed(context.Background(), record.DelegationID, err)
		_ = al.persistDelegationMemory(context.Background(), record.DelegationID)
		al.publishDelegationBlockerSummary(context.Background(), record, err)
		return record, req, result, err
	}

	target, ok := al.registry.GetAgent(targetAgentID)
	if !ok || target == nil {
		err := fmt.Errorf(
			"%w: target agent %q is not registered",
			ErrAgentDelegationTargetNotFound,
			targetAgentID,
		)
		_ = al.delegationRecords.Failed(context.Background(), record.DelegationID, err)
		_ = al.persistDelegationMemory(context.Background(), record.DelegationID)
		al.publishDelegationBlockerSummary(context.Background(), record, err)
		return record, req, result, err
	}
	return record, req, result, nil
}

func (al *AgentLoop) runPreparedAgentDelegation(
	ctx context.Context,
	req AgentDelegationRequest,
	record AgentDelegationRecord,
	result AgentDelegationResult,
) (AgentDelegationResult, error) {
	target, ok := al.registry.GetAgent(req.TargetAgentID)
	if !ok || target == nil {
		err := fmt.Errorf(
			"%w: target agent %q is not registered",
			ErrAgentDelegationTargetNotFound,
			req.TargetAgentID,
		)
		_ = al.delegationRecords.Failed(context.Background(), record.DelegationID, err)
		_ = al.persistDelegationMemory(context.Background(), record.DelegationID)
		al.publishDelegationBlockerSummary(context.Background(), record, err)
		return result, err
	}

	scope := buildDelegationSessionScope(req.ParentAgentID, req.TargetAgentID, req.ThreadKey)
	sessionKey := session.BuildSessionKey(scope)
	alias := buildDelegationSessionAlias(req.ParentAgentID, req.TargetAgentID, req.ThreadKey)
	dispatch := DispatchRequest{
		SessionKey:     sessionKey,
		SessionAliases: []string{alias},
		SessionScope:   &scope,
		UserMessage:    req.Task,
		InboundContext: &bus.InboundContext{
			Channel:  "internal",
			ChatID:   alias,
			ChatType: "delegation",
			SenderID: req.ParentAgentID,
		},
		RouteResult: &routing.ResolvedRoute{
			AgentID:   req.TargetAgentID,
			Channel:   "internal",
			MatchedBy: "delegation",
			SessionPolicy: routing.SessionPolicy{
				Dimensions: []string{"delegation"},
			},
		},
	}
	opts := processOptions{
		Dispatch:                dispatch,
		SenderID:                req.ParentAgentID,
		SenderDisplayName:       req.ParentAgentID,
		DefaultResponse:         defaultResponse,
		EnableSummary:           false,
		SendResponse:            false,
		SkipInitialSteeringPoll: true,
	}

	ensureSessionMetadata(target.Sessions, sessionKey, &scope, []string{alias})
	if err := al.delegationRecords.Running(ctx, record.DelegationID); err != nil {
		return result, err
	}
	if err := al.persistDelegationMemory(ctx, record.DelegationID); err != nil {
		return result, err
	}

	turnScope := al.newTurnEventScope(
		target.ID,
		sessionKey,
		newTurnContext(dispatch.InboundContext, dispatch.RouteResult, dispatch.SessionScope),
	)
	ts := newTurnState(target, opts, turnScope)
	pipeline := NewPipeline(al)
	turnRes, err := al.runTurn(ctx, ts, pipeline)
	result.SessionKey = sessionKey
	result.SessionScope = session.CloneScope(&scope)
	result.Content = turnRes.finalContent
	result.Status = turnRes.status
	if err != nil {
		_ = al.delegationRecords.Failed(context.Background(), record.DelegationID, err)
		_ = al.persistDelegationMemory(context.Background(), record.DelegationID)
		al.publishDelegationBlockerSummary(context.Background(), record, err)
		return result, err
	}
	if err := al.delegationRecords.Completed(ctx, record.DelegationID, result); err != nil {
		return result, err
	}
	al.publishDelegationCompletedSummary(ctx, result)
	result = al.maybePublishDelegationGitHubArtifact(ctx, record, req, result)
	if err := al.persistDelegationMemory(ctx, record.DelegationID); err != nil {
		return result, err
	}
	return result, nil
}

func agentDelegationRequestFromTool(req tools.DelegateExecutionRequest, mode string) AgentDelegationRequest {
	dueAt := parseDelegationDue(req.Due)
	return AgentDelegationRequest{
		ParentAgentID:     routing.NormalizeAgentID(req.ParentAgentID),
		TargetAgentID:     routing.NormalizeAgentID(req.TargetAgentID),
		Task:              req.Task,
		ThreadKey:         req.ThreadKey,
		Mode:              mode,
		Priority:          req.Priority,
		DueAt:             dueAt,
		RequestedBy:       req.RequestedBy,
		VisibleToAgentIDs: compactDelegationRefs(req.VisibleToAgentIDs),
		ApprovalRequired:  req.ApprovalRequired,
		ArtifactRefs:      req.ArtifactRefs,
	}
}

func parseDelegationDue(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339, time.DateOnly} {
		if t, err := time.Parse(layout, value); err == nil {
			return &t
		}
	}
	return nil
}

func toolDelegationResult(result AgentDelegationResult) tools.DelegateExecutionResult {
	return tools.DelegateExecutionResult{
		DelegationID:  result.DelegationID,
		ParentAgentID: result.ParentAgentID,
		TargetAgentID: result.TargetAgentID,
		Content:       result.Content,
		Status:        string(result.Status),
		ArtifactRefs:  result.ArtifactRefs,
	}
}

func toolDelegationRecord(rec AgentDelegationRecord) tools.DelegationRecord {
	out := tools.DelegationRecord{
		DelegationID:      rec.DelegationID,
		Status:            string(rec.Status),
		ParentAgentID:     rec.ParentAgentID,
		TargetAgentID:     rec.TargetAgentID,
		Task:              rec.Request.Task,
		ThreadKey:         rec.Request.ThreadKey,
		Mode:              rec.Request.Mode,
		Priority:          rec.Request.Priority,
		RequestedBy:       rec.Request.RequestedBy,
		VisibleToAgentIDs: rec.VisibleToAgentIDs,
		ArtifactRefs:      rec.Request.ArtifactRefs,
		CreatedAt:         rec.CreatedAt,
		UpdatedAt:         rec.UpdatedAt,
		StartedAt:         rec.StartedAt,
		CompletedAt:       rec.CompletedAt,
	}
	if rec.Result != nil {
		out.Result = rec.Result.Content
	}
	if rec.Error != nil {
		out.Error = rec.Error.Message
	}
	return out
}

func buildDelegationSessionScope(parentAgentID, targetAgentID, threadKey string) session.SessionScope {
	return session.SessionScope{
		Version:    session.ScopeVersionV1,
		AgentID:    targetAgentID,
		Channel:    "internal",
		Dimensions: []string{"delegation"},
		Values: map[string]string{
			"delegation": buildDelegationSessionValue(parentAgentID, targetAgentID, threadKey),
		},
	}
}

func buildDelegationSessionAlias(parentAgentID, targetAgentID, threadKey string) string {
	return "internal:delegation:" + buildDelegationSessionValue(parentAgentID, targetAgentID, threadKey)
}

func buildDelegationSessionValue(parentAgentID, targetAgentID, threadKey string) string {
	threadKey = strings.TrimSpace(strings.ToLower(threadKey))
	if threadKey == "" {
		threadKey = "default"
	}
	return fmt.Sprintf("%s:%s:%s", parentAgentID, targetAgentID, threadKey)
}

func compactDelegationRefs(refs []string) []string {
	values := make([]string, 0, len(refs))
	for _, ref := range refs {
		ref = strings.TrimSpace(ref)
		if ref != "" {
			values = append(values, ref)
		}
	}
	return values
}
