package tools

import (
	"context"
	"fmt"
	"strings"
)

type MeetingExecutionRequest struct {
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

type MeetingExecutionResult struct {
	MeetingID      string
	Recommendation string
	Participants   []string
	Timeline       []string
	Risks          []string
	Approvals      []string
	FollowUps      []string
	ArtifactRefs   []string
}

type MeetingRunner interface {
	StartAgentMeeting(ctx context.Context, req MeetingExecutionRequest) (MeetingExecutionResult, error)
}

type MeetingTool struct {
	runner MeetingRunner
}

func NewMeetingTool(runner MeetingRunner) *MeetingTool {
	return &MeetingTool{runner: runner}
}

func (t *MeetingTool) Name() string {
	return "start_agent_meeting"
}

func (t *MeetingTool) Description() string {
	return "Start meeting v1: a private chaired sequential agent meeting. Participants are consulted one at a time, not live real-time debate; user-facing output is the chair's consolidated recommendation."
}

func (t *MeetingTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"title": map[string]any{
				"type":        "string",
				"description": "Meeting v1 title.",
			},
			"sponsor_agent_id": map[string]any{
				"type":        "string",
				"description": "Optional sponsoring agent ID. Defaults to the calling agent.",
			},
			"chair_agent_id": map[string]any{
				"type":        "string",
				"description": "Configured agent ID that chairs the meeting and owns the consolidated recommendation.",
			},
			"participant_agent_ids": map[string]any{
				"type":        "array",
				"description": "Configured participant agent IDs to consult privately in sequential meeting v1 turns.",
				"items": map[string]any{
					"type": "string",
				},
			},
			"goal": map[string]any{
				"type":        "string",
				"description": "Meeting v1 goal or decision to resolve.",
			},
			"constraints": map[string]any{
				"type":        "array",
				"description": "Constraints, approval boundaries, or operating limits.",
				"items": map[string]any{
					"type": "string",
				},
			},
			"notes": map[string]any{
				"type":        "string",
				"description": "Private meeting v1 context or chair notes.",
			},
			"approvals": map[string]any{
				"type":        "array",
				"description": "Known approvals needed before execution.",
				"items": map[string]any{
					"type": "string",
				},
			},
			"artifact_refs": map[string]any{
				"type":        "array",
				"description": "Related paths, issues, PRs, project items, memory IDs, objectives, or meeting IDs.",
				"items": map[string]any{
					"type": "string",
				},
			},
		},
		"required": []string{"title", "chair_agent_id", "participant_agent_ids", "goal"},
	}
}

func (t *MeetingTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	if t.runner == nil {
		return meetingError("execution_failed", "meeting execution failed: meeting runner not configured")
	}
	title := stringArg(args, "title")
	if title == "" {
		return meetingError("missing_title", "title is required and must be a non-empty string")
	}
	chair := stringArg(args, "chair_agent_id")
	if chair == "" {
		return meetingError("missing_chair", "chair_agent_id is required and must be a non-empty string")
	}
	participants := compactStringRefsFromAny(args["participant_agent_ids"])
	if len(participants) == 0 {
		return meetingError("missing_participants", "participant_agent_ids is required and must include at least one agent ID")
	}
	goal := stringArg(args, "goal")
	if goal == "" {
		return meetingError("missing_goal", "goal is required and must be a non-empty string")
	}
	sponsor := stringArg(args, "sponsor_agent_id")
	if sponsor == "" {
		sponsor = strings.TrimSpace(ToolAgentID(ctx))
	}

	result, err := t.runner.StartAgentMeeting(ctx, MeetingExecutionRequest{
		Title:               title,
		SponsorAgentID:      sponsor,
		ChairAgentID:        chair,
		ParticipantAgentIDs: participants,
		Goal:                goal,
		Constraints:         compactStringRefsFromAny(args["constraints"]),
		Notes:               stringArg(args, "notes"),
		Approvals:           compactStringRefsFromAny(args["approvals"]),
		ArtifactRefs:        compactStringRefsFromAny(args["artifact_refs"]),
	})
	if err != nil {
		return meetingError("execution_failed", fmt.Sprintf("meeting execution failed: %v", err)).WithError(err)
	}

	return &ToolResult{
		ForLLM:  formatMeetingResultForLLM(result),
		ForUser: formatMeetingResultForUser(result),
	}
}

func meetingError(code, message string) *ToolResult {
	return ErrorResult(fmt.Sprintf("start_agent_meeting error [%s]: %s", code, message)).
		WithError(fmt.Errorf("%s: %s", code, message))
}

func compactStringRefsFromAny(value any) []string {
	switch refs := value.(type) {
	case []string:
		return compactStringRefs(refs)
	case []any:
		values := make([]string, 0, len(refs))
		for _, ref := range refs {
			if s, ok := ref.(string); ok {
				values = append(values, s)
			}
		}
		return compactStringRefs(values)
	default:
		return nil
	}
}

func formatMeetingResultForLLM(result MeetingExecutionResult) string {
	var sb strings.Builder
	sb.WriteString("Agent meeting v1 completed.")
	sb.WriteString("\nMeeting ID: ")
	sb.WriteString(result.MeetingID)
	if len(result.ArtifactRefs) > 0 {
		sb.WriteString("\nArtifacts: ")
		sb.WriteString(strings.Join(result.ArtifactRefs, ", "))
	}
	if len(result.Participants) > 0 {
		sb.WriteString("\nParticipants: ")
		sb.WriteString(strings.Join(result.Participants, ", "))
	}
	sb.WriteString("\nConsolidated recommendation: ")
	sb.WriteString(strings.TrimSpace(result.Recommendation))
	if len(result.Timeline) > 0 {
		sb.WriteString("\nTimeline: ")
		sb.WriteString(strings.Join(result.Timeline, "; "))
	}
	if len(result.Risks) > 0 {
		sb.WriteString("\nRisks: ")
		sb.WriteString(strings.Join(result.Risks, "; "))
	}
	if len(result.Approvals) > 0 {
		sb.WriteString("\nApproval needed: ")
		sb.WriteString(strings.Join(result.Approvals, "; "))
	}
	if len(result.FollowUps) > 0 {
		sb.WriteString("\nFollow-ups: ")
		sb.WriteString(strings.Join(result.FollowUps, "; "))
	}
	return sb.String()
}

func formatMeetingResultForUser(result MeetingExecutionResult) string {
	var sb strings.Builder
	sb.WriteString(strings.TrimSpace(result.Recommendation))
	if len(result.Timeline) > 0 {
		sb.WriteString("\n\nTimeline: ")
		sb.WriteString(strings.Join(result.Timeline, "; "))
	}
	if len(result.Risks) > 0 {
		sb.WriteString("\nRisks: ")
		sb.WriteString(strings.Join(result.Risks, "; "))
	}
	if len(result.Approvals) > 0 {
		sb.WriteString("\nApproval needed: ")
		sb.WriteString(strings.Join(result.Approvals, "; "))
	}
	if len(result.FollowUps) > 0 {
		sb.WriteString("\nFollow-ups: ")
		sb.WriteString(strings.Join(result.FollowUps, "; "))
	}
	return strings.TrimSpace(sb.String())
}
