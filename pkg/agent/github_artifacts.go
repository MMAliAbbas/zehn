package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	integrationtools "github.com/sipeed/picoclaw/pkg/tools/integration"
)

type AgentGitHubArtifactWriter = integrationtools.GitHubArtifactWriter

type AgentGitHubArtifactStatus string

const (
	AgentGitHubArtifactStatusCreated AgentGitHubArtifactStatus = "created"
	AgentGitHubArtifactStatusFailed  AgentGitHubArtifactStatus = "failed"
)

type AgentGitHubArtifactWrite struct {
	Status    AgentGitHubArtifactStatus `json:"status"`
	IssueURL  string                    `json:"issue_url,omitempty"`
	IssueID   int                       `json:"issue_id,omitempty"`
	Error     string                    `json:"error,omitempty"`
	UpdatedAt time.Time                 `json:"updated_at"`
}

func (al *AgentLoop) SetGitHubArtifactWriter(writer AgentGitHubArtifactWriter) {
	if al == nil {
		return
	}
	al.githubArtifacts = writer
}

func (al *AgentLoop) maybePublishMeetingGitHubArtifact(
	ctx context.Context,
	record AgentMeetingRecord,
	outcome AgentMeetingOutcome,
) (AgentMeetingRecord, error) {
	if al == nil || al.githubArtifacts == nil {
		return record, nil
	}
	if !meetingNeedsGitHubIssue(record, outcome) {
		return record, nil
	}

	issue, err := al.githubArtifacts.CreateIssue(ctx, integrationtools.GitHubIssueRequest{
		SourceType: "meeting",
		SourceID:   record.MeetingID,
		Title:      "Meeting: " + record.Title,
		Body:       buildMeetingGitHubIssueBody(record, outcome),
		Labels:     []string{"meeting", "tracker"},
	})
	if err != nil {
		_ = al.meetingRecords.RecordGitHubArtifact(context.Background(), record.MeetingID, AgentGitHubArtifactWrite{
			Status: AgentGitHubArtifactStatusFailed,
			Error:  err.Error(),
		}, nil)
		return al.meetingRecords.Get(context.Background(), record.MeetingID)
	}

	for _, turn := range record.ParticipantTurns {
		body := curatedGitHubParticipantComment(turn)
		if body == "" {
			continue
		}
		if err := al.githubArtifacts.CreateComment(ctx, integrationtools.GitHubCommentRequest{
			IssueNumber:   issue.Number,
			IssueURL:      issue.URL,
			SourceType:    "meeting",
			SourceID:      record.MeetingID,
			AuthorAgentID: turn.AgentID,
			Body:          body,
		}); err != nil {
			_ = al.meetingRecords.RecordGitHubArtifact(context.Background(), record.MeetingID, AgentGitHubArtifactWrite{
				Status:   AgentGitHubArtifactStatusFailed,
				IssueID:  issue.Number,
				IssueURL: issue.URL,
				Error:    err.Error(),
			}, []string{issue.URL})
			return al.meetingRecords.Get(context.Background(), record.MeetingID)
		}
	}

	if err := al.meetingRecords.RecordGitHubArtifact(ctx, record.MeetingID, AgentGitHubArtifactWrite{
		Status:   AgentGitHubArtifactStatusCreated,
		IssueID:  issue.Number,
		IssueURL: issue.URL,
	}, []string{issue.URL}); err != nil {
		return record, err
	}
	al.publishIssueCreatedSummary(ctx, "meeting", record.MeetingID, issue.URL)
	return al.meetingRecords.Get(ctx, record.MeetingID)
}

func (al *AgentLoop) maybePublishDelegationGitHubArtifact(
	ctx context.Context,
	record AgentDelegationRecord,
	req AgentDelegationRequest,
	result AgentDelegationResult,
) AgentDelegationResult {
	if al == nil || al.githubArtifacts == nil || !delegationNeedsGitHubIssue(req) {
		return result
	}
	issue, err := al.githubArtifacts.CreateIssue(ctx, integrationtools.GitHubIssueRequest{
		SourceType: "delegation",
		SourceID:   record.DelegationID,
		Title:      "Delegation: " + delegationIssueTitle(req),
		Body:       buildDelegationGitHubIssueBody(record, req, result),
		Labels:     []string{"delegation", "tracker"},
	})
	if err != nil {
		_ = al.delegationRecords.RecordGitHubArtifact(context.Background(), record.DelegationID, AgentGitHubArtifactWrite{
			Status: AgentGitHubArtifactStatusFailed,
			Error:  err.Error(),
		}, nil)
		return result
	}
	result.ArtifactRefs = appendUniqueRefs(result.ArtifactRefs, issue.URL)
	_ = al.delegationRecords.RecordGitHubArtifact(ctx, record.DelegationID, AgentGitHubArtifactWrite{
		Status:   AgentGitHubArtifactStatusCreated,
		IssueID:  issue.Number,
		IssueURL: issue.URL,
	}, []string{issue.URL})
	al.publishIssueCreatedSummary(ctx, "delegation", record.DelegationID, issue.URL)
	return result
}

func meetingNeedsGitHubIssue(record AgentMeetingRecord, outcome AgentMeetingOutcome) bool {
	return len(record.Approvals) > 0 || len(outcome.FollowUps) > 0
}

func delegationNeedsGitHubIssue(req AgentDelegationRequest) bool {
	return req.ApprovalRequired || strings.EqualFold(strings.TrimSpace(req.Mode), "async")
}

func buildMeetingGitHubIssueBody(record AgentMeetingRecord, outcome AgentMeetingOutcome) string {
	var sb strings.Builder
	sb.WriteString("GitHub is a tracker for executable work and approval follow-up. Durable meeting memory remains in the meeting record and configured memory systems.\n\n")
	sb.WriteString("## Meeting\n")
	sb.WriteString("- Meeting ID: ")
	sb.WriteString(record.MeetingID)
	sb.WriteString("\n- Chair: ")
	sb.WriteString(record.ChairAgentID)
	sb.WriteString("\n- Sponsor: ")
	sb.WriteString(record.SponsorAgentID)
	sb.WriteString("\n\n## Goal\n")
	sb.WriteString(record.Goal)
	sb.WriteString("\n\n## Consolidated Recommendation\n")
	sb.WriteString(strings.TrimSpace(outcome.Recommendation))
	appendGitHubSection(&sb, "Timeline", outcome.Timeline)
	appendGitHubSection(&sb, "Risks", outcome.Risks)
	appendGitHubSection(&sb, "Approvals", record.Approvals)
	appendGitHubSection(&sb, "Follow-ups", outcome.FollowUps)
	return strings.TrimSpace(sb.String())
}

func buildDelegationGitHubIssueBody(
	record AgentDelegationRecord,
	req AgentDelegationRequest,
	result AgentDelegationResult,
) string {
	var sb strings.Builder
	sb.WriteString("GitHub is a tracker for delegated executable work or approval follow-up. Durable delegation memory remains in the delegation record and configured memory systems.\n\n")
	sb.WriteString("## Delegation\n")
	sb.WriteString("- Delegation ID: ")
	sb.WriteString(record.DelegationID)
	sb.WriteString("\n- Parent: ")
	sb.WriteString(req.ParentAgentID)
	sb.WriteString("\n- Target: ")
	sb.WriteString(req.TargetAgentID)
	if req.ApprovalRequired {
		sb.WriteString("\n- Approval required: yes")
	}
	if req.Priority != "" {
		sb.WriteString("\n- Priority: ")
		sb.WriteString(req.Priority)
	}
	sb.WriteString("\n\n## Task\n")
	sb.WriteString(req.Task)
	if result.Content != "" {
		sb.WriteString("\n\n## Current Result\n")
		sb.WriteString(strings.TrimSpace(result.Content))
	}
	return strings.TrimSpace(sb.String())
}

func appendGitHubSection(sb *strings.Builder, title string, values []string) {
	if len(values) == 0 {
		return
	}
	sb.WriteString("\n\n## ")
	sb.WriteString(title)
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		sb.WriteString("\n- ")
		sb.WriteString(value)
	}
}

func curatedGitHubParticipantComment(turn AgentMeetingTurn) string {
	lines := make([]string, 0, 4)
	for _, line := range strings.Split(turn.Response, "\n") {
		line = strings.TrimSpace(strings.TrimLeft(line, "-* "))
		if line == "" || isRawInternalLine(line) || !isMaterialGitHubCommentLine(line) {
			continue
		}
		lines = append(lines, line)
		if len(lines) == 6 {
			break
		}
	}
	if len(lines) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("Focused participant note from ")
	sb.WriteString(turn.AgentID)
	sb.WriteString(":")
	for _, line := range lines {
		sb.WriteString("\n- ")
		sb.WriteString(line)
	}
	return sb.String()
}

func isRawInternalLine(line string) bool {
	lower := strings.ToLower(line)
	return strings.Contains(lower, "raw internal") ||
		strings.Contains(lower, "transcript") ||
		strings.Contains(lower, "raw debate")
}

func isMaterialGitHubCommentLine(line string) bool {
	key, _, ok := strings.Cut(line, ":")
	if !ok {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "position", "risk", "risks", "commitment", "commitments", "dependency", "dependencies", "acceptance criteria", "acceptance criterion", "follow-up", "follow-ups":
		return true
	default:
		return false
	}
}

func delegationIssueTitle(req AgentDelegationRequest) string {
	if req.ThreadKey != "" {
		return req.ThreadKey
	}
	task := strings.TrimSpace(req.Task)
	if len(task) > 72 {
		return task[:72]
	}
	if task == "" {
		return fmt.Sprintf("%s to %s", req.ParentAgentID, req.TargetAgentID)
	}
	return task
}

func appendUniqueRefs(refs []string, values ...string) []string {
	out := append([]string(nil), refs...)
	seen := make(map[string]struct{}, len(out)+len(values))
	for _, ref := range out {
		seen[ref] = struct{}{}
	}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
