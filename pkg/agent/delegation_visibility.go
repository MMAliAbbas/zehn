package agent

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/logger"
)

const (
	visibilityEventDelegationCreated   = "delegation_created"
	visibilityEventMeetingOpened       = "meeting_opened"
	visibilityEventRecommendationReady = "recommendation_ready"
	visibilityEventApprovalNeeded      = "approval_needed"
	visibilityEventIssueCreated        = "issue_created"
	visibilityEventBlockerRaised       = "blocker_raised"
	visibilityEventCompletion          = "completion"
)

func (al *AgentLoop) publishDelegationCreatedSummary(ctx context.Context, record AgentDelegationRecord) {
	al.publishDiscordVisibilitySummary(ctx, visibilityEventDelegationCreated, fmt.Sprintf(
		"Delegation created: %s %s -> %s (%s).",
		record.DelegationID,
		record.ParentAgentID,
		record.TargetAgentID,
		visibilityValueOrDefault(record.Request.Mode, "sync"),
	))
	if record.Request.ApprovalRequired {
		al.publishDiscordVisibilitySummary(ctx, visibilityEventApprovalNeeded, fmt.Sprintf(
			"Approval needed: delegation %s requires approval before execution.",
			record.DelegationID,
		))
	}
}

func (al *AgentLoop) publishDelegationCompletedSummary(ctx context.Context, result AgentDelegationResult) {
	al.publishDiscordVisibilitySummary(ctx, visibilityEventCompletion, fmt.Sprintf(
		"Delegation completed: %s target=%s status=%s.",
		result.DelegationID,
		result.TargetAgentID,
		visibilityValueOrDefault(string(result.Status), "completed"),
	))
}

func (al *AgentLoop) publishDelegationBlockerSummary(ctx context.Context, record AgentDelegationRecord, err error) {
	if err == nil {
		return
	}
	al.publishDiscordVisibilitySummary(ctx, visibilityEventBlockerRaised, fmt.Sprintf(
		"Blocker raised: delegation %s target=%s failed: %s.",
		record.DelegationID,
		record.TargetAgentID,
		visibilityCompact(err.Error(), 120),
	))
}

func (al *AgentLoop) publishMeetingOpenedSummary(ctx context.Context, record AgentMeetingRecord) {
	al.publishDiscordVisibilitySummary(ctx, visibilityEventMeetingOpened, fmt.Sprintf(
		"Meeting opened: %s chair=%s participants=%d.",
		record.MeetingID,
		record.ChairAgentID,
		len(record.Participants),
	))
	if len(record.Approvals) > 0 {
		al.publishDiscordVisibilitySummary(ctx, visibilityEventApprovalNeeded, fmt.Sprintf(
			"Approval needed: meeting %s has %d approval item(s).",
			record.MeetingID,
			len(record.Approvals),
		))
	}
}

func (al *AgentLoop) publishMeetingRecommendationSummary(ctx context.Context, record AgentMeetingRecord) {
	al.publishDiscordVisibilitySummary(ctx, visibilityEventRecommendationReady, fmt.Sprintf(
		"Recommendation ready: meeting %s chair=%s.",
		record.MeetingID,
		record.ChairAgentID,
	))
}

func (al *AgentLoop) publishMeetingCompletedSummary(ctx context.Context, record AgentMeetingRecord) {
	al.publishDiscordVisibilitySummary(ctx, visibilityEventCompletion, fmt.Sprintf(
		"Meeting completed: %s chair=%s status=%s.",
		record.MeetingID,
		record.ChairAgentID,
		visibilityValueOrDefault(string(record.Status), "completed"),
	))
}

func (al *AgentLoop) publishIssueCreatedSummary(ctx context.Context, sourceType, sourceID, issueURL string) {
	al.publishDiscordVisibilitySummary(ctx, visibilityEventIssueCreated, fmt.Sprintf(
		"Issue created: %s %s -> %s.",
		sourceType,
		sourceID,
		issueURL,
	))
}

func (al *AgentLoop) publishDiscordVisibilitySummary(ctx context.Context, event, content string) {
	channel, chatID, ok := al.discordVisibilitySummaryTarget(event)
	if !ok {
		return
	}
	pubCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	msg := bus.OutboundMessage{
		Context: bus.InboundContext{
			Channel:  channel,
			ChatID:   chatID,
			ChatType: "channel",
			Raw: map[string]string{
				"message_kind":             "visibility_summary",
				"visibility_summary_event": event,
			},
		},
		Content: visibilityCompact(content, 300),
	}
	if err := al.bus.PublishOutbound(pubCtx, msg); err != nil {
		logger.WarnCF("agent", "Discord visibility summary publish failed", map[string]any{
			"event": event,
			"error": err.Error(),
		})
	}
}

func (al *AgentLoop) discordVisibilitySummaryTarget(event string) (string, string, bool) {
	if al == nil || al.cfg == nil || al.bus == nil {
		return "", "", false
	}
	names := slices.Sorted(maps.Keys(al.cfg.Channels))
	for _, name := range names {
		ch := al.cfg.Channels[name]
		if ch == nil || !ch.Enabled || ch.Type != config.ChannelDiscord {
			continue
		}
		decoded, err := ch.GetDecoded()
		if err != nil {
			continue
		}
		settings, ok := decoded.(*config.DiscordSettings)
		if !ok || !settings.VisibilitySummaries.Enabled {
			continue
		}
		chatID := strings.TrimSpace(settings.VisibilitySummaries.ChatID)
		if chatID == "" || !visibilityEventAllowed(settings.VisibilitySummaries.Events, event) {
			continue
		}
		return name, chatID, true
	}
	return "", "", false
}

func visibilityEventAllowed(events []string, event string) bool {
	if len(events) == 0 {
		return true
	}
	for _, allowed := range events {
		if strings.EqualFold(strings.TrimSpace(allowed), event) {
			return true
		}
	}
	return false
}

func visibilityValueOrDefault(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func visibilityCompact(value string, maxRunes int) string {
	value = strings.Join(strings.Fields(value), " ")
	if maxRunes <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= maxRunes {
		return value
	}
	if maxRunes <= 3 {
		return string(runes[:maxRunes])
	}
	return string(runes[:maxRunes-3]) + "..."
}
