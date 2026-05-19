package agent

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/routing"
	"github.com/sipeed/picoclaw/pkg/session"
	"github.com/sipeed/picoclaw/pkg/tools"
)

var (
	ErrAgentMeetingInvalidRequest = errors.New("invalid agent meeting request")
	ErrAgentMeetingTargetNotFound = errors.New("agent meeting target not found")
)

type AgentMeetingRequest struct {
	Title               string
	SponsorAgentID      string
	ChairAgentID        string
	ParticipantAgentIDs []string
	Goal                string
	Constraints         []string
	Notes               string
	Approvals           []string
	ArtifactRefs        []string
}

type AgentMeetingOutcome struct {
	Recommendation string
	Timeline       []string
	Risks          []string
	FollowUps      []string
}

func (al *AgentLoop) StartAgentMeeting(
	ctx context.Context,
	req tools.MeetingExecutionRequest,
) (tools.MeetingExecutionResult, error) {
	outcome, record, err := al.RunAgentMeeting(ctx, AgentMeetingRequest{
		Title:               req.Title,
		SponsorAgentID:      req.SponsorAgentID,
		ChairAgentID:        req.ChairAgentID,
		ParticipantAgentIDs: req.ParticipantAgentIDs,
		Goal:                req.Goal,
		Constraints:         req.Constraints,
		Notes:               req.Notes,
		Approvals:           req.Approvals,
		ArtifactRefs:        req.ArtifactRefs,
	})
	if err != nil {
		return tools.MeetingExecutionResult{}, err
	}
	return tools.MeetingExecutionResult{
		MeetingID:      record.MeetingID,
		Recommendation: outcome.Recommendation,
		Participants:   record.Participants,
		Timeline:       outcome.Timeline,
		Risks:          outcome.Risks,
		Approvals:      record.Approvals,
		FollowUps:      outcome.FollowUps,
		ArtifactRefs:   record.ArtifactRefs,
	}, nil
}

func (al *AgentLoop) RunAgentMeeting(
	ctx context.Context,
	req AgentMeetingRequest,
) (AgentMeetingOutcome, AgentMeetingRecord, error) {
	req = normalizeAgentMeetingRequest(req)
	if err := al.validateAgentMeetingRequest(req); err != nil {
		return AgentMeetingOutcome{}, AgentMeetingRecord{}, err
	}

	record, err := al.meetingRecords.Started(ctx, req)
	if err != nil {
		return AgentMeetingOutcome{}, AgentMeetingRecord{}, err
	}
	al.publishMeetingOpenedSummary(ctx, record)

	turns := make([]AgentMeetingTurn, 0, len(req.ParticipantAgentIDs))
	for _, participantID := range req.ParticipantAgentIDs {
		result, err := al.RunAgentDelegation(ctx, AgentDelegationRequest{
			ParentAgentID: req.ChairAgentID,
			TargetAgentID: participantID,
			Task:          buildMeetingParticipantTask(record.MeetingID, req, participantID),
			ThreadKey:     record.MeetingID,
			ArtifactRefs:  append([]string{record.MeetingID}, req.ArtifactRefs...),
		})
		turn := AgentMeetingTurn{
			AgentID:      participantID,
			DelegationID: result.DelegationID,
			Response:     result.Content,
			Status:       string(result.Status),
		}
		if err != nil {
			turn.Status = "failed"
			turn.Response = err.Error()
			_ = al.meetingRecords.AddParticipantTurn(context.Background(), record.MeetingID, turn)
			return AgentMeetingOutcome{}, record, al.failAgentMeetingRecord(record.MeetingID, err)
		}
		if err := al.meetingRecords.AddParticipantTurn(ctx, record.MeetingID, turn); err != nil {
			return AgentMeetingOutcome{}, record, al.failAgentMeetingRecord(record.MeetingID, err)
		}
		turns = append(turns, turn)
	}

	chairResponse, err := al.runAgentMeetingChairTurn(ctx, record.MeetingID, req, turns)
	if err != nil {
		err = al.failAgentMeetingRecord(record.MeetingID, err)
		al.publishDiscordVisibilitySummary(context.Background(), visibilityEventBlockerRaised, fmt.Sprintf(
			"Blocker raised: meeting %s chair=%s failed: %s.",
			record.MeetingID,
			req.ChairAgentID,
			visibilityCompact(err.Error(), 120),
		))
		return AgentMeetingOutcome{}, record, err
	}
	outcome := parseAgentMeetingOutcome(chairResponse)
	if outcome.Recommendation == "" {
		outcome.Recommendation = strings.TrimSpace(chairResponse)
	}
	if err := al.meetingRecords.Completed(ctx, record.MeetingID, AgentMeetingChairTurn{
		AgentID:  req.ChairAgentID,
		Response: chairResponse,
	}, outcome); err != nil {
		return AgentMeetingOutcome{}, record, al.failAgentMeetingRecord(record.MeetingID, err)
	}
	al.publishMeetingRecommendationSummary(ctx, record)
	record, err = al.meetingRecords.Get(ctx, record.MeetingID)
	if err != nil {
		return AgentMeetingOutcome{}, record, err
	}
	al.publishMeetingCompletedSummary(ctx, record)
	record, err = al.maybePublishMeetingGitHubArtifact(ctx, record, outcome)
	if err != nil {
		return AgentMeetingOutcome{}, record, err
	}
	return outcome, record, nil
}

func (al *AgentLoop) failAgentMeetingRecord(meetingID string, err error) error {
	if err == nil {
		err = errors.New("meeting failed")
	}
	if recordErr := al.meetingRecords.Failed(context.Background(), meetingID, err); recordErr != nil {
		return errors.Join(err, fmt.Errorf("record meeting failure: %w", recordErr))
	}
	return err
}

func normalizeAgentMeetingRequest(req AgentMeetingRequest) AgentMeetingRequest {
	req.Title = strings.TrimSpace(req.Title)
	req.SponsorAgentID = routing.NormalizeAgentID(req.SponsorAgentID)
	req.ChairAgentID = routing.NormalizeAgentID(req.ChairAgentID)
	req.Goal = strings.TrimSpace(req.Goal)
	req.Notes = strings.TrimSpace(req.Notes)
	req.ParticipantAgentIDs = compactAgentIDs(req.ParticipantAgentIDs)
	req.Constraints = compactDelegationRefs(req.Constraints)
	req.Approvals = compactDelegationRefs(req.Approvals)
	req.ArtifactRefs = compactDelegationRefs(req.ArtifactRefs)
	return req
}

func (al *AgentLoop) validateAgentMeetingRequest(req AgentMeetingRequest) error {
	if al == nil || al.registry == nil || al.meetingRecords == nil {
		return fmt.Errorf("%w: agent loop is not initialized", ErrAgentMeetingInvalidRequest)
	}
	if req.Title == "" || req.SponsorAgentID == "" || req.ChairAgentID == "" || req.Goal == "" {
		return fmt.Errorf("%w: title, sponsor, chair, and goal are required", ErrAgentMeetingInvalidRequest)
	}
	if len(req.ParticipantAgentIDs) == 0 {
		return fmt.Errorf("%w: at least one participant is required", ErrAgentMeetingInvalidRequest)
	}
	if _, ok := al.registry.GetAgent(req.SponsorAgentID); !ok {
		return fmt.Errorf("%w: sponsor agent %q is not registered", ErrAgentMeetingTargetNotFound, req.SponsorAgentID)
	}
	if _, ok := al.registry.GetAgent(req.ChairAgentID); !ok {
		return fmt.Errorf("%w: chair agent %q is not registered", ErrAgentMeetingTargetNotFound, req.ChairAgentID)
	}
	if req.SponsorAgentID != req.ChairAgentID && !al.registry.CanSpawnSubagent(req.SponsorAgentID, req.ChairAgentID) {
		return fmt.Errorf("%w: sponsor %q cannot delegate chair synthesis to %q", ErrAgentDelegationPermissionDenied, req.SponsorAgentID, req.ChairAgentID)
	}
	for _, participantID := range req.ParticipantAgentIDs {
		if _, ok := al.registry.GetAgent(participantID); !ok {
			return fmt.Errorf("%w: participant agent %q is not registered", ErrAgentMeetingTargetNotFound, participantID)
		}
		if !al.registry.CanSpawnSubagent(req.ChairAgentID, participantID) {
			return fmt.Errorf("%w: chair %q cannot delegate participant turn to %q", ErrAgentDelegationPermissionDenied, req.ChairAgentID, participantID)
		}
	}
	return nil
}

func (al *AgentLoop) runAgentMeetingChairTurn(
	ctx context.Context,
	meetingID string,
	req AgentMeetingRequest,
	turns []AgentMeetingTurn,
) (string, error) {
	if req.SponsorAgentID != req.ChairAgentID {
		result, err := al.RunAgentDelegation(ctx, AgentDelegationRequest{
			ParentAgentID: req.SponsorAgentID,
			TargetAgentID: req.ChairAgentID,
			Task:          buildMeetingChairTask(meetingID, req, turns),
			ThreadKey:     meetingID + ":chair",
			ArtifactRefs:  append([]string{meetingID}, req.ArtifactRefs...),
		})
		return result.Content, err
	}

	return al.runInternalAgentMeetingTurn(ctx, req.ChairAgentID, meetingID, buildMeetingChairTask(meetingID, req, turns))
}

func (al *AgentLoop) runInternalAgentMeetingTurn(
	ctx context.Context,
	chairAgentID, meetingID, task string,
) (string, error) {
	target, ok := al.registry.GetAgent(chairAgentID)
	if !ok || target == nil {
		return "", fmt.Errorf("%w: chair agent %q is not registered", ErrAgentMeetingTargetNotFound, chairAgentID)
	}
	scope := session.SessionScope{
		Version:    session.ScopeVersionV1,
		AgentID:    chairAgentID,
		Channel:    "internal",
		Dimensions: []string{"meeting"},
		Values: map[string]string{
			"meeting": meetingID + ":chair",
		},
	}
	sessionKey := session.BuildSessionKey(scope)
	alias := "internal:meeting:" + meetingID + ":chair"
	dispatch := DispatchRequest{
		SessionKey:     sessionKey,
		SessionAliases: []string{alias},
		SessionScope:   &scope,
		UserMessage:    task,
		InboundContext: &bus.InboundContext{
			Channel:  "internal",
			ChatID:   alias,
			ChatType: "meeting",
			SenderID: chairAgentID,
		},
		RouteResult: &routing.ResolvedRoute{
			AgentID:   chairAgentID,
			Channel:   "internal",
			MatchedBy: "meeting",
			SessionPolicy: routing.SessionPolicy{
				Dimensions: []string{"meeting"},
			},
		},
	}
	opts := processOptions{
		Dispatch:                dispatch,
		SenderID:                chairAgentID,
		SenderDisplayName:       chairAgentID,
		DefaultResponse:         defaultResponse,
		EnableSummary:           false,
		SendResponse:            false,
		SuppressToolFeedback:    true,
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
	if err != nil {
		return turnRes.finalContent, err
	}
	return turnRes.finalContent, nil
}

func buildMeetingParticipantTask(meetingID string, req AgentMeetingRequest, participantID string) string {
	var sb strings.Builder
	sb.WriteString("You are participating in a private chaired agent meeting.\n")
	sb.WriteString("Meeting ID: ")
	sb.WriteString(meetingID)
	sb.WriteString("\nTitle: ")
	sb.WriteString(req.Title)
	sb.WriteString("\nSponsor: ")
	sb.WriteString(req.SponsorAgentID)
	sb.WriteString("\nChair: ")
	sb.WriteString(req.ChairAgentID)
	sb.WriteString("\nTarget agent: ")
	sb.WriteString(participantID)
	sb.WriteString("\nParticipant: ")
	sb.WriteString(participantID)
	sb.WriteString("\nGoal: ")
	sb.WriteString(req.Goal)
	appendMeetingList(&sb, "Constraints", req.Constraints)
	if req.Notes != "" {
		sb.WriteString("\nNotes: ")
		sb.WriteString(req.Notes)
	}
	appendMeetingList(&sb, "Approvals", req.Approvals)
	appendMeetingList(&sb, "Artifacts", req.ArtifactRefs)
	sb.WriteString("\n\nReturn your concise domain position, timeline concerns, risks, and follow-ups for the chair. Do not address the user directly.")
	return sb.String()
}

func buildMeetingChairTask(meetingID string, req AgentMeetingRequest, turns []AgentMeetingTurn) string {
	var sb strings.Builder
	sb.WriteString("Consolidate this chaired meeting into one recommendation.\n")
	sb.WriteString("Meeting ID: ")
	sb.WriteString(meetingID)
	sb.WriteString("\nTitle: ")
	sb.WriteString(req.Title)
	sb.WriteString("\nSponsor: ")
	sb.WriteString(req.SponsorAgentID)
	sb.WriteString("\nChair: ")
	sb.WriteString(req.ChairAgentID)
	sb.WriteString("\nGoal: ")
	sb.WriteString(req.Goal)
	appendMeetingList(&sb, "Constraints", req.Constraints)
	appendMeetingList(&sb, "Approvals", req.Approvals)
	appendMeetingList(&sb, "Artifacts", req.ArtifactRefs)
	if req.Notes != "" {
		sb.WriteString("\nNotes: ")
		sb.WriteString(req.Notes)
	}
	sb.WriteString("\n\nParticipant positions:")
	for _, turn := range turns {
		sb.WriteString("\n- ")
		sb.WriteString(turn.AgentID)
		sb.WriteString(": ")
		sb.WriteString(strings.TrimSpace(turn.Response))
	}
	sb.WriteString("\n\nReturn a consolidated recommendation with explicit Timeline, Risks, and Follow-ups sections. Do not include raw debate.")
	return sb.String()
}

func appendMeetingList(sb *strings.Builder, label string, values []string) {
	if len(values) == 0 {
		return
	}
	sb.WriteString("\n")
	sb.WriteString(label)
	sb.WriteString(": ")
	sb.WriteString(strings.Join(values, "; "))
}

func parseAgentMeetingOutcome(content string) AgentMeetingOutcome {
	var outcome AgentMeetingOutcome
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(strings.TrimLeft(line, "-*"))
		if line == "" {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			if outcome.Recommendation == "" {
				outcome.Recommendation = line
			}
			continue
		}
		value = strings.TrimSpace(value)
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "recommendation", "consolidated recommendation":
			outcome.Recommendation = value
		case "timeline":
			outcome.Timeline = splitMeetingItems(value)
		case "risks", "risk":
			outcome.Risks = splitMeetingItems(value)
		case "follow-ups", "follow ups", "followups", "follow-up tasks":
			outcome.FollowUps = splitMeetingItems(value)
		}
	}
	if outcome.Recommendation == "" {
		outcome.Recommendation = strings.TrimSpace(content)
	}
	return outcome
}

func splitMeetingItems(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ';'
	})
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func compactAgentIDs(ids []string) []string {
	out := make([]string, 0, len(ids))
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		id = routing.NormalizeAgentID(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
