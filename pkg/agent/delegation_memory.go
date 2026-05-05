package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	picoclawmcp "github.com/sipeed/picoclaw/pkg/mcp"
)

type DelegationMemoryWriter interface {
	WriteDelegationMemory(ctx context.Context, rec AgentDelegationRecord) (AgentDelegationMemoryWrite, error)
}

type delegationMCPToolCaller interface {
	CallTool(
		ctx context.Context,
		serverName, toolName string,
		arguments map[string]any,
	) (*sdkmcp.CallToolResult, error)
}

var delegationMemoryWriterForAgentLoop = func(al *AgentLoop) DelegationMemoryWriter {
	if al == nil {
		return nil
	}
	manager := al.mcp.getManager()
	if manager == nil {
		return nil
	}
	return NewYaadDelegationMemoryWriter(manager, yaadDelegationMemoryServerName(manager))
}

var delegationMemoryStrictForAgentLoop = func(al *AgentLoop) bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("PICOCLAW_DELEGATION_MEMORY_STRICT")), "true")
}

type YaadDelegationMemoryWriter struct {
	caller     delegationMCPToolCaller
	serverName string
}

func NewYaadDelegationMemoryWriter(caller delegationMCPToolCaller, serverName string) *YaadDelegationMemoryWriter {
	serverName = strings.TrimSpace(serverName)
	if serverName == "" {
		serverName = "yaad"
	}
	return &YaadDelegationMemoryWriter{
		caller:     caller,
		serverName: serverName,
	}
}

func (w *YaadDelegationMemoryWriter) WriteDelegationMemory(
	ctx context.Context,
	rec AgentDelegationRecord,
) (AgentDelegationMemoryWrite, error) {
	if w == nil || w.caller == nil {
		return AgentDelegationMemoryWrite{}, errors.New("yaad delegation memory writer is unavailable")
	}
	args := yaadDelegationMemoryAddArgs(rec)
	result, err := w.caller.CallTool(ctx, w.serverName, "memory_add", args)
	if err != nil {
		return AgentDelegationMemoryWrite{}, err
	}
	if result == nil {
		return AgentDelegationMemoryWrite{}, errors.New("yaad memory_add returned nil result")
	}
	if result.IsError {
		return AgentDelegationMemoryWrite{}, fmt.Errorf("yaad memory_add returned error: %s", delegationMCPContentText(result.Content))
	}
	return AgentDelegationMemoryWrite{
		Provider: "yaad",
		Status:   AgentDelegationMemoryStatusWritten,
		MemoryID: delegationMemoryIDFromMCPResult(result),
	}, nil
}

func (al *AgentLoop) persistDelegationMemory(ctx context.Context, delegationID string) error {
	if al == nil || al.delegationRecords == nil {
		return nil
	}
	rec, err := al.delegationRecords.Get(ctx, delegationID)
	if err != nil {
		return err
	}
	if !shouldWriteDelegationMemory(rec) {
		return al.delegationRecords.RecordMemorySkipped(
			context.Background(),
			delegationID,
			rec.Status,
			"yaad delegation memory skipped until terminal status",
		)
	}
	if rec.DurableMemory != nil && rec.DurableMemory.Status == AgentDelegationMemoryStatusWritten {
		return al.delegationRecords.RecordMemorySkipped(
			context.Background(),
			delegationID,
			rec.Status,
			"yaad delegation memory already written; no update or upsert tool is configured",
		)
	}

	strict := delegationMemoryStrictForAgentLoop(al)
	writer := delegationMemoryWriterForAgentLoop(al)
	if writer == nil {
		write := AgentDelegationMemoryWrite{
			Provider: "yaad",
			Status:   AgentDelegationMemoryStatusUnavailable,
			Error:    "yaad delegation memory writer unavailable",
		}
		if err := al.delegationRecords.RecordMemoryWrite(context.Background(), delegationID, write); err != nil {
			return err
		}
		if strict {
			return errors.New(write.Error)
		}
		return nil
	}
	write, err := writer.WriteDelegationMemory(ctx, rec)
	if err != nil {
		write = AgentDelegationMemoryWrite{
			Provider: "yaad",
			Status:   AgentDelegationMemoryStatusFailed,
			Error:    err.Error(),
		}
		if recordErr := al.delegationRecords.RecordMemoryWrite(context.Background(), delegationID, write); recordErr != nil {
			return recordErr
		}
		if strict {
			return err
		}
		return nil
	}
	if write.Provider == "" {
		write.Provider = "yaad"
	}
	if write.Status == "" {
		write.Status = AgentDelegationMemoryStatusWritten
	}
	return al.delegationRecords.RecordMemoryWrite(ctx, delegationID, write)
}

func shouldWriteDelegationMemory(rec AgentDelegationRecord) bool {
	switch rec.Status {
	case AgentDelegationStatusCompleted, AgentDelegationStatusFailed, AgentDelegationStatusCancelled:
		return true
	default:
		return false
	}
}

func yaadDelegationMemoryServerName(manager *picoclawmcp.Manager) string {
	if manager == nil {
		return "yaad"
	}
	if _, ok := manager.GetServer("yaad"); ok {
		return "yaad"
	}
	for name, conn := range manager.GetServers() {
		for _, tool := range conn.Tools {
			if tool != nil && tool.Name == "memory_add" {
				return name
			}
		}
	}
	return "yaad"
}

func yaadDelegationMemoryAddArgs(rec AgentDelegationRecord) map[string]any {
	return map[string]any{
		"memory_class": "summary",
		"title":        fmt.Sprintf("Delegation %s: %s to %s", rec.DelegationID, rec.ParentAgentID, rec.TargetAgentID),
		"raw_content":  yaadDelegationMemoryContent(rec),
		"scopes": []map[string]any{
			{
				"scope": map[string]any{
					"scope_type":   "project",
					"external_key": "zehn",
				},
			},
			{
				"scope": map[string]any{
					"scope_type":   "agent",
					"external_key": rec.ParentAgentID,
				},
			},
			{
				"scope": map[string]any{
					"scope_type":   "agent",
					"external_key": rec.TargetAgentID,
				},
			},
		},
		"labels": []string{
			"zehn",
			"delegation",
			string(rec.Status),
			rec.ParentAgentID,
			rec.TargetAgentID,
		},
		"source": "picoclaw-delegation",
	}
}

func yaadDelegationMemoryContent(rec AgentDelegationRecord) string {
	payload := map[string]any{
		"delegation_id":   rec.DelegationID,
		"parent_agent_id": rec.ParentAgentID,
		"target_agent_id": rec.TargetAgentID,
		"status":          rec.Status,
		"request":         rec.Request,
		"result":          rec.Result,
		"error":           rec.Error,
		"decisions":       delegationMemoryDecisions(rec),
		"follow_ups":      delegationMemoryFollowUps(rec),
		"created_at":      rec.CreatedAt,
		"updated_at":      rec.UpdatedAt,
		"started_at":      rec.StartedAt,
		"completed_at":    rec.CompletedAt,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Sprintf("delegation_id=%s status=%s parent=%s target=%s", rec.DelegationID, rec.Status, rec.ParentAgentID, rec.TargetAgentID)
	}
	return string(data)
}

func delegationMemoryDecisions(rec AgentDelegationRecord) []string {
	if rec.Result == nil || strings.TrimSpace(rec.Result.Content) == "" {
		return nil
	}
	return []string{rec.Result.Content}
}

func delegationMemoryFollowUps(rec AgentDelegationRecord) []string {
	if len(rec.Request.ArtifactRefs) == 0 {
		return nil
	}
	return append([]string(nil), rec.Request.ArtifactRefs...)
}

func delegationMemoryIDFromMCPResult(result *sdkmcp.CallToolResult) string {
	text := delegationMCPContentText(result.Content)
	if text == "" {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(text), &payload); err != nil {
		return ""
	}
	for _, key := range []string{"memory_id", "id"} {
		if value, ok := payload[key].(string); ok {
			return strings.TrimSpace(value)
		}
	}
	if memory, ok := payload["memory"].(map[string]any); ok {
		if value, ok := memory["id"].(string); ok {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func delegationMCPContentText(content []sdkmcp.Content) string {
	var parts []string
	for _, item := range content {
		if text, ok := item.(*sdkmcp.TextContent); ok {
			parts = append(parts, text.Text)
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}
