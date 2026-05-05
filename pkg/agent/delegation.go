package agent

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/routing"
	"github.com/sipeed/picoclaw/pkg/session"
)

var (
	ErrAgentDelegationPermissionDenied = errors.New("agent delegation permission denied")
	ErrAgentDelegationTargetNotFound   = errors.New("agent delegation target not found")
	ErrAgentDelegationInvalidRequest   = errors.New("invalid agent delegation request")
)

// AgentDelegationRequest describes a single delegated turn from one configured
// agent to another. It is intentionally source-level: public tools and durable
// queues can wrap it without overloading the existing spawn/subagent behavior.
type AgentDelegationRequest struct {
	ParentAgentID string
	TargetAgentID string
	Task          string
	ThreadKey     string
}

// AgentDelegationResult is the synchronous outcome of a delegated target-agent
// turn.
type AgentDelegationResult struct {
	ParentAgentID string
	TargetAgentID string
	SessionKey    string
	SessionScope  *session.SessionScope
	Content       string
	Status        TurnEndStatus
}

// RunAgentDelegation runs req.Task through the real configured target
// AgentInstance, using a private internal delegation session scope.
func (al *AgentLoop) RunAgentDelegation(
	ctx context.Context,
	req AgentDelegationRequest,
) (AgentDelegationResult, error) {
	if al == nil || al.registry == nil {
		return AgentDelegationResult{}, fmt.Errorf(
			"%w: agent loop is not initialized",
			ErrAgentDelegationInvalidRequest,
		)
	}

	parentAgentID := routing.NormalizeAgentID(req.ParentAgentID)
	targetAgentID := routing.NormalizeAgentID(req.TargetAgentID)
	task := strings.TrimSpace(req.Task)
	if parentAgentID == "" || targetAgentID == "" || task == "" {
		return AgentDelegationResult{}, fmt.Errorf(
			"%w: parent agent, target agent, and task are required",
			ErrAgentDelegationInvalidRequest,
		)
	}

	if !al.registry.CanSpawnSubagent(parentAgentID, targetAgentID) {
		return AgentDelegationResult{}, fmt.Errorf(
			"%w: parent %q cannot delegate to target %q",
			ErrAgentDelegationPermissionDenied,
			parentAgentID,
			targetAgentID,
		)
	}

	target, ok := al.registry.GetAgent(targetAgentID)
	if !ok || target == nil {
		return AgentDelegationResult{}, fmt.Errorf(
			"%w: target agent %q is not registered",
			ErrAgentDelegationTargetNotFound,
			targetAgentID,
		)
	}

	scope := buildDelegationSessionScope(parentAgentID, targetAgentID, req.ThreadKey)
	sessionKey := session.BuildSessionKey(scope)
	alias := buildDelegationSessionAlias(parentAgentID, targetAgentID, req.ThreadKey)
	dispatch := DispatchRequest{
		SessionKey:     sessionKey,
		SessionAliases: []string{alias},
		SessionScope:   &scope,
		UserMessage:    task,
		InboundContext: &bus.InboundContext{
			Channel:  "internal",
			ChatID:   alias,
			ChatType: "delegation",
			SenderID: parentAgentID,
		},
		RouteResult: &routing.ResolvedRoute{
			AgentID:   targetAgentID,
			Channel:   "internal",
			MatchedBy: "delegation",
			SessionPolicy: routing.SessionPolicy{
				Dimensions: []string{"delegation"},
			},
		},
	}
	opts := processOptions{
		Dispatch:                dispatch,
		SenderID:                parentAgentID,
		SenderDisplayName:       parentAgentID,
		DefaultResponse:         defaultResponse,
		EnableSummary:           false,
		SendResponse:            false,
		SkipInitialSteeringPoll: true,
	}

	ensureSessionMetadata(target.Sessions, sessionKey, &scope, []string{alias})

	turnScope := al.newTurnEventScope(
		target.ID,
		sessionKey,
		newTurnContext(dispatch.InboundContext, dispatch.RouteResult, dispatch.SessionScope),
	)
	ts := newTurnState(target, opts, turnScope)
	pipeline := NewPipeline(al)
	turnRes, err := al.runTurn(ctx, ts, pipeline)
	result := AgentDelegationResult{
		ParentAgentID: parentAgentID,
		TargetAgentID: targetAgentID,
		SessionKey:    sessionKey,
		SessionScope:  session.CloneScope(&scope),
		Content:       turnRes.finalContent,
		Status:        turnRes.status,
	}
	if err != nil {
		return result, err
	}
	return result, nil
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
