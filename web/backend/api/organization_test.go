package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sipeed/picoclaw/pkg/agent"
	"github.com/sipeed/picoclaw/pkg/config"
)

func TestHandleAgentOrganization_EmptyHierarchyReturnsEmptyActivity(t *testing.T) {
	configPath := writeOrganizationAPIConfig(t, func(cfg *config.Config) {
		cfg.Agents.List = nil
	})

	rec := requestAgentOrganization(t, configPath)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp agentOrganizationSnapshot
	decodeJSONResponse(t, rec, &resp)
	if len(resp.Roots) != 0 {
		t.Fatalf("len(Roots) = %d, want 0", len(resp.Roots))
	}
	if len(resp.Agents) != 0 {
		t.Fatalf("len(Agents) = %d, want 0", len(resp.Agents))
	}
	if resp.Activity.DelegationCount != 0 || resp.Activity.MeetingCount != 0 {
		t.Fatalf("activity = %+v, want empty counts", resp.Activity)
	}
}

func TestHandleAgentOrganization_ExplicitHierarchyBuildsStableTree(t *testing.T) {
	configPath := writeOrganizationAPIConfig(t, func(cfg *config.Config) {
		cfg.Agents.List = []config.AgentConfig{
			{ID: "ceo", Name: "Coordinator"},
			{ID: "cto", Name: "Technology"},
			{ID: "ops", Name: "Operations"},
			{ID: "sales", Name: "Sales"},
		}
		cfg.Agents.Organization = &config.AgentOrganizationConfig{
			Roots: []string{"ceo"},
			Nodes: []config.AgentOrganizationNodeConfig{
				{AgentID: "ops", ParentAgentID: "ceo", Label: "Ops", Group: "delivery", Sort: 20},
				{AgentID: "ceo", Label: "Executive", Group: "exec", Sort: 10},
				{AgentID: "cto", ParentAgentID: "ceo", Label: "Tech", Group: "product", Sort: 10},
			},
		}
	})

	rec := requestAgentOrganization(t, configPath)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp agentOrganizationSnapshot
	decodeJSONResponse(t, rec, &resp)
	if got, want := agentNodeIDs(resp.Roots), []string{"ceo", "sales"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("root IDs = %v, want %v", got, want)
	}
	if got, want := agentNodeIDs(resp.Roots[0].Children), []string{"cto", "ops"}; strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("ceo child IDs = %v, want %v", got, want)
	}
	if resp.Roots[0].Label != "Executive" || resp.Roots[0].Group != "exec" {
		t.Fatalf("root metadata = %+v", resp.Roots[0])
	}
	if resp.Roots[0].Children[0].Label != "Tech" || resp.Roots[0].Children[0].Name != "Technology" {
		t.Fatalf("child metadata = %+v", resp.Roots[0].Children[0])
	}
}

func TestHandleAgentOrganization_ActiveDelegationStatusUsesStructuredRecords(t *testing.T) {
	configPath := writeOrganizationAPIConfig(t, func(cfg *config.Config) {
		cfg.Agents.List = []config.AgentConfig{
			{ID: "lead", Name: "Lead"},
			{ID: "worker", Name: "Worker"},
		}
	})
	cfg := loadOrganizationAPIConfig(t, configPath)
	store := agent.NewDelegationRecordStore(filepath.Join(cfg.WorkspacePath(), "delegations"), nil)
	rec, err := store.Requested(context.Background(), agent.AgentDelegationRequest{
		ParentAgentID: "lead",
		TargetAgentID: "worker",
		Task:          "raw prompt must not leak",
	})
	if err != nil {
		t.Fatalf("Requested() error = %v", err)
	}
	if err := store.Running(context.Background(), rec.DelegationID); err != nil {
		t.Fatalf("Running() error = %v", err)
	}

	httpRec := requestAgentOrganization(t, configPath)
	if httpRec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", httpRec.Code, http.StatusOK, httpRec.Body.String())
	}
	if strings.Contains(httpRec.Body.String(), "raw prompt must not leak") {
		t.Fatalf("response leaked raw delegation task: %s", httpRec.Body.String())
	}

	var resp agentOrganizationSnapshot
	decodeJSONResponse(t, httpRec, &resp)
	if status := resp.Agents["worker"].Status; status != agentOrganizationStatusWorking {
		t.Fatalf("worker status = %q, want %q", status, agentOrganizationStatusWorking)
	}
	if status := resp.Agents["lead"].Status; status != agentOrganizationStatusDelegating {
		t.Fatalf("lead status = %q, want %q", status, agentOrganizationStatusDelegating)
	}
	if current := resp.Agents["worker"].Activity.Current; current == nil || current.RecordID != rec.DelegationID {
		t.Fatalf("worker current = %+v, want delegation %q", current, rec.DelegationID)
	}
}

func TestHandleAgentOrganization_ActiveMeetingStatusUsesStructuredRecords(t *testing.T) {
	configPath := writeOrganizationAPIConfig(t, func(cfg *config.Config) {
		cfg.Agents.List = []config.AgentConfig{
			{ID: "sponsor", Name: "Sponsor"},
			{ID: "chair", Name: "Chair"},
			{ID: "participant", Name: "Participant"},
		}
	})
	cfg := loadOrganizationAPIConfig(t, configPath)
	store := agent.NewMeetingRecordStore(filepath.Join(cfg.WorkspacePath(), "meetings"), nil)
	rec, err := store.Started(context.Background(), agent.AgentMeetingRequest{
		Title:               "Quarter plan",
		SponsorAgentID:      "sponsor",
		ChairAgentID:        "chair",
		ParticipantAgentIDs: []string{"participant"},
		Goal:                "raw meeting goal must not leak",
	})
	if err != nil {
		t.Fatalf("Started() error = %v", err)
	}

	httpRec := requestAgentOrganization(t, configPath)
	if httpRec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", httpRec.Code, http.StatusOK, httpRec.Body.String())
	}
	if strings.Contains(httpRec.Body.String(), "raw meeting goal must not leak") {
		t.Fatalf("response leaked raw meeting goal: %s", httpRec.Body.String())
	}

	var resp agentOrganizationSnapshot
	decodeJSONResponse(t, httpRec, &resp)
	for _, agentID := range []string{"sponsor", "chair", "participant"} {
		if status := resp.Agents[agentID].Status; status != agentOrganizationStatusMeeting {
			t.Fatalf("%s status = %q, want %q", agentID, status, agentOrganizationStatusMeeting)
		}
		if current := resp.Agents[agentID].Activity.Current; current == nil || current.RecordID != rec.MeetingID {
			t.Fatalf("%s current = %+v, want meeting %q", agentID, current, rec.MeetingID)
		}
	}
}

func TestHandleAgentOrganization_NewerActiveDelegationSupersedesOlderFailure(t *testing.T) {
	configPath := writeOrganizationAPIConfig(t, func(cfg *config.Config) {
		cfg.Agents.List = []config.AgentConfig{
			{ID: "lead", Name: "Lead"},
			{ID: "worker", Name: "Worker"},
		}
	})
	cfg := loadOrganizationAPIConfig(t, configPath)
	store := agent.NewDelegationRecordStore(filepath.Join(cfg.WorkspacePath(), "delegations"), nil)
	failed, err := store.Requested(context.Background(), agent.AgentDelegationRequest{
		ParentAgentID: "lead",
		TargetAgentID: "worker",
		Task:          "failed work",
	})
	if err != nil {
		t.Fatalf("Requested(failed) error = %v", err)
	}
	if err := store.Failed(context.Background(), failed.DelegationID, errors.New("private failure details")); err != nil {
		t.Fatalf("Failed() error = %v", err)
	}
	failed.Status = agent.AgentDelegationStatusFailed
	setDelegationCreatedAt(t, cfg, &failed, time.Date(2026, 5, 7, 9, 0, 0, 0, time.UTC))

	active, err := store.Requested(context.Background(), agent.AgentDelegationRequest{
		ParentAgentID: "lead",
		TargetAgentID: "worker",
		Task:          "active work",
	})
	if err != nil {
		t.Fatalf("Requested(active) error = %v", err)
	}
	if err := store.Running(context.Background(), active.DelegationID); err != nil {
		t.Fatalf("Running(active) error = %v", err)
	}
	active.Status = agent.AgentDelegationStatusRunning
	setDelegationCreatedAt(t, cfg, &active, time.Date(2026, 5, 7, 10, 0, 0, 0, time.UTC))

	httpRec := requestAgentOrganization(t, configPath)
	if httpRec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", httpRec.Code, http.StatusOK, httpRec.Body.String())
	}
	if strings.Contains(httpRec.Body.String(), "private failure details") {
		t.Fatalf("response leaked raw failure details: %s", httpRec.Body.String())
	}

	var resp agentOrganizationSnapshot
	decodeJSONResponse(t, httpRec, &resp)
	worker := resp.Agents["worker"]
	if status := worker.Status; status != agentOrganizationStatusWorking {
		t.Fatalf("worker status = %q, want %q", status, agentOrganizationStatusWorking)
	}
	if current := worker.Activity.Current; current == nil || current.RecordID != active.DelegationID {
		t.Fatalf("worker current = %+v, want active delegation %q", current, active.DelegationID)
	}
	if lastFailure := worker.Activity.LastFailure; lastFailure == nil || lastFailure.RecordID != failed.DelegationID {
		t.Fatalf("worker last failure = %+v, want failed delegation %q", lastFailure, failed.DelegationID)
	}
	if worker.Activity.FailureCount != 1 {
		t.Fatalf("worker failure count = %d, want 1", worker.Activity.FailureCount)
	}
}

func TestHandleAgentOrganization_LastFailureIncludesDrilldownMetadata(t *testing.T) {
	configPath := writeOrganizationAPIConfig(t, func(cfg *config.Config) {
		cfg.Agents.List = []config.AgentConfig{
			{ID: "lead", Name: "Lead"},
			{ID: "worker", Name: "Worker"},
		}
	})
	cfg := loadOrganizationAPIConfig(t, configPath)
	store := agent.NewDelegationRecordStore(filepath.Join(cfg.WorkspacePath(), "delegations"), nil)
	failed, err := store.Requested(context.Background(), agent.AgentDelegationRequest{
		ParentAgentID: "lead",
		TargetAgentID: "worker",
		Task:          "failed work",
		ArtifactRefs:  []string{"github:issue/123"},
	})
	if err != nil {
		t.Fatalf("Requested(failed) error = %v", err)
	}
	if err := store.Failed(context.Background(), failed.DelegationID, errors.New("private failure details")); err != nil {
		t.Fatalf("Failed() error = %v", err)
	}
	failed.Status = agent.AgentDelegationStatusFailed
	createdAt := time.Date(2026, 5, 7, 9, 0, 0, 0, time.UTC)
	completedAt := time.Date(2026, 5, 7, 9, 30, 0, 0, time.UTC)
	setDelegationTimes(t, cfg, &failed, createdAt, completedAt)

	httpRec := requestAgentOrganization(t, configPath)
	if httpRec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", httpRec.Code, http.StatusOK, httpRec.Body.String())
	}

	var resp agentOrganizationSnapshot
	decodeJSONResponse(t, httpRec, &resp)
	lastFailure := resp.Agents["worker"].Activity.LastFailure
	if lastFailure == nil {
		t.Fatal("worker last failure is nil, want drilldown record")
	}
	if lastFailure.RecordID != failed.DelegationID || lastFailure.AgentID != "lead" || lastFailure.Role != "target" {
		t.Fatalf("last failure identity = %+v, want record %q peer lead role target", lastFailure, failed.DelegationID)
	}
	if lastFailure.CreatedAt == nil || !lastFailure.CreatedAt.Equal(createdAt) {
		t.Fatalf("last failure created_at = %+v, want %s", lastFailure.CreatedAt, createdAt)
	}
	if lastFailure.CompletedAt == nil || !lastFailure.CompletedAt.Equal(completedAt) {
		t.Fatalf("last failure completed_at = %+v, want %s", lastFailure.CompletedAt, completedAt)
	}
	if got, want := strings.Join(lastFailure.ArtifactRefs, ","), "github:issue/123"; got != want {
		t.Fatalf("last failure artifact refs = %q, want %q", got, want)
	}
}

func TestHandleAgentOrganization_NewerFailureRemainsCurrentWithoutNewerActiveRecord(t *testing.T) {
	configPath := writeOrganizationAPIConfig(t, func(cfg *config.Config) {
		cfg.Agents.List = []config.AgentConfig{
			{ID: "lead", Name: "Lead"},
			{ID: "worker", Name: "Worker"},
		}
	})
	cfg := loadOrganizationAPIConfig(t, configPath)
	store := agent.NewDelegationRecordStore(filepath.Join(cfg.WorkspacePath(), "delegations"), nil)
	active, err := store.Requested(context.Background(), agent.AgentDelegationRequest{
		ParentAgentID: "lead",
		TargetAgentID: "worker",
		Task:          "active work",
	})
	if err != nil {
		t.Fatalf("Requested(active) error = %v", err)
	}
	if err := store.Running(context.Background(), active.DelegationID); err != nil {
		t.Fatalf("Running(active) error = %v", err)
	}
	active.Status = agent.AgentDelegationStatusRunning
	setDelegationCreatedAt(t, cfg, &active, time.Date(2026, 5, 7, 9, 0, 0, 0, time.UTC))

	failed, err := store.Requested(context.Background(), agent.AgentDelegationRequest{
		ParentAgentID: "lead",
		TargetAgentID: "worker",
		Task:          "failed work",
	})
	if err != nil {
		t.Fatalf("Requested(failed) error = %v", err)
	}
	if err := store.Failed(context.Background(), failed.DelegationID, errors.New("private failure details")); err != nil {
		t.Fatalf("Failed() error = %v", err)
	}
	failed.Status = agent.AgentDelegationStatusFailed
	setDelegationCreatedAt(t, cfg, &failed, time.Date(2026, 5, 7, 10, 0, 0, 0, time.UTC))

	httpRec := requestAgentOrganization(t, configPath)
	if httpRec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", httpRec.Code, http.StatusOK, httpRec.Body.String())
	}
	if strings.Contains(httpRec.Body.String(), "private failure details") {
		t.Fatalf("response leaked raw failure details: %s", httpRec.Body.String())
	}

	var resp agentOrganizationSnapshot
	decodeJSONResponse(t, httpRec, &resp)
	if status := resp.Agents["worker"].Status; status != agentOrganizationStatusFailed {
		t.Fatalf("worker status = %q, want %q", status, agentOrganizationStatusFailed)
	}
	if current := resp.Agents["worker"].Activity.Current; current == nil || current.RecordID != failed.DelegationID {
		t.Fatalf("worker current = %+v, want failed delegation %q", current, failed.DelegationID)
	}
	if lastFailure := resp.Agents["worker"].Activity.LastFailure; lastFailure == nil || lastFailure.RecordID != failed.DelegationID {
		t.Fatalf("worker last failure = %+v, want failed delegation %q", lastFailure, failed.DelegationID)
	}
}

func TestHandleAgentOrganization_MalformedConfigReturnsError(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "config.json")
	if err := os.WriteFile(configPath, []byte(`{"version":3,"agents":{"list":[{"id":"main"}],"organization":{"nodes":[{"agent_id":"missing"}]}}}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	rec := requestAgentOrganization(t, configPath)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
}

func TestHandleAgentOrganization_NoLogsReturnsEmptyRecentEvents(t *testing.T) {
	withIsolatedGatewayLogs(t)
	configPath := writeOrganizationAPIConfig(t, func(cfg *config.Config) {
		cfg.Agents.List = []config.AgentConfig{{ID: "worker", Name: "Worker"}}
	})

	rec := requestAgentOrganization(t, configPath)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp agentOrganizationSnapshot
	decodeJSONResponse(t, rec, &resp)
	if events := resp.Agents["worker"].Activity.RecentEvents; len(events) != 0 {
		t.Fatalf("recent events = %+v, want empty", events)
	}
}

func TestHandleAgentOrganization_RecentEventsIncludeMatchingStructuredGatewayLogs(t *testing.T) {
	withIsolatedGatewayLogs(t)
	configPath := writeOrganizationAPIConfig(t, func(cfg *config.Config) {
		cfg.Agents.List = []config.AgentConfig{{ID: "worker", Name: "Worker"}}
	})
	gateway.logs.Append(`{"level":"info","time":"2026-05-07T10:11:12Z","agent_id":"worker","event":"turn_started","message":"agent turn started"}`)

	rec := requestAgentOrganization(t, configPath)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp agentOrganizationSnapshot
	decodeJSONResponse(t, rec, &resp)
	agentState := resp.Agents["worker"]
	if agentState.Status != agentOrganizationStatusIdle {
		t.Fatalf("status = %q, want structured idle status", agentState.Status)
	}
	if len(agentState.Activity.RecentEvents) != 1 {
		t.Fatalf("recent events = %+v, want one event", agentState.Activity.RecentEvents)
	}
	event := agentState.Activity.RecentEvents[0]
	if event.AgentID != "worker" || event.Level != "info" || event.Message != "agent turn started" || event.Event != "turn_started" {
		t.Fatalf("event = %+v, want parsed structured log fields", event)
	}
	if event.Timestamp == nil || !event.Timestamp.Equal(time.Date(2026, 5, 7, 10, 11, 12, 0, time.UTC)) {
		t.Fatalf("timestamp = %+v, want parsed UTC timestamp", event.Timestamp)
	}
}

func TestHandleAgentOrganization_ActivityFeedSummarizesRecentOrgEvents(t *testing.T) {
	withIsolatedGatewayLogs(t)
	configPath := writeOrganizationAPIConfig(t, func(cfg *config.Config) {
		cfg.Agents.List = []config.AgentConfig{
			{ID: "lead", Name: "Lead"},
			{ID: "worker", Name: "Worker"},
			{ID: "chair", Name: "Chair"},
		}
	})
	cfg := loadOrganizationAPIConfig(t, configPath)

	delegations := agent.NewDelegationRecordStore(filepath.Join(cfg.WorkspacePath(), "delegations"), nil)
	failed, err := delegations.Requested(context.Background(), agent.AgentDelegationRequest{
		ParentAgentID: "lead",
		TargetAgentID: "worker",
		Task:          "private failure prompt must not leak",
	})
	if err != nil {
		t.Fatalf("Requested(failed) error = %v", err)
	}
	if err := delegations.Failed(context.Background(), failed.DelegationID, errors.New("private failure details")); err != nil {
		t.Fatalf("Failed() error = %v", err)
	}
	failed.Status = agent.AgentDelegationStatusFailed
	setDelegationCreatedAt(t, cfg, &failed, time.Date(2026, 5, 7, 9, 0, 0, 0, time.UTC))

	running, err := delegations.Requested(context.Background(), agent.AgentDelegationRequest{
		ParentAgentID: "lead",
		TargetAgentID: "worker",
		Task:          "private running prompt must not leak",
	})
	if err != nil {
		t.Fatalf("Requested(running) error = %v", err)
	}
	if err := delegations.Running(context.Background(), running.DelegationID); err != nil {
		t.Fatalf("Running() error = %v", err)
	}
	running.Status = agent.AgentDelegationStatusRunning
	setDelegationCreatedAt(t, cfg, &running, time.Date(2026, 5, 7, 10, 0, 0, 0, time.UTC))

	meetings := agent.NewMeetingRecordStore(filepath.Join(cfg.WorkspacePath(), "meetings"), nil)
	meeting := writeMeetingRecord(t, meetings, agent.AgentMeetingRequest{
		Title:               "Quarter plan",
		SponsorAgentID:      "lead",
		ChairAgentID:        "chair",
		ParticipantAgentIDs: []string{"worker"},
		Goal:                "private meeting goal must not leak",
	})
	setMeetingCreatedAt(t, cfg, &meeting, time.Date(2026, 5, 7, 11, 0, 0, 0, time.UTC))

	gateway.logs.Append(`{"level":"info","time":"2026-05-07T12:00:00Z","agent_id":"worker","event":"turn_finished","message":"turn finished"}`)

	httpRec := requestAgentOrganization(t, configPath)
	if httpRec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", httpRec.Code, http.StatusOK, httpRec.Body.String())
	}
	body := httpRec.Body.String()
	for _, private := range []string{"private failure prompt", "private running prompt", "private meeting goal", "private failure details"} {
		if strings.Contains(body, private) {
			t.Fatalf("response leaked private content %q: %s", private, body)
		}
	}

	var resp agentOrganizationSnapshot
	decodeJSONResponse(t, httpRec, &resp)
	if len(resp.Activity.Recent) != 4 {
		t.Fatalf("recent feed = %+v, want four entries", resp.Activity.Recent)
	}

	wantTypes := []string{"event", "meeting", "delegation", "failure"}
	wantAgents := []string{"worker", "chair", "worker", "worker"}
	for i, entry := range resp.Activity.Recent {
		if entry.Type != wantTypes[i] || entry.AgentID != wantAgents[i] {
			t.Fatalf("recent[%d] = %+v, want type %q agent %q", i, entry, wantTypes[i], wantAgents[i])
		}
		if entry.Status == "" {
			t.Fatalf("recent[%d] missing status: %+v", i, entry)
		}
		if entry.Timestamp == nil {
			t.Fatalf("recent[%d] missing timestamp: %+v", i, entry)
		}
	}
	if resp.Activity.Recent[1].RecordID != meeting.MeetingID {
		t.Fatalf("meeting record id = %q, want %q", resp.Activity.Recent[1].RecordID, meeting.MeetingID)
	}
	if resp.Activity.Recent[2].RecordID != running.DelegationID {
		t.Fatalf("running delegation id = %q, want %q", resp.Activity.Recent[2].RecordID, running.DelegationID)
	}
	if resp.Activity.Recent[3].RecordID != failed.DelegationID {
		t.Fatalf("failed delegation id = %q, want %q", resp.Activity.Recent[3].RecordID, failed.DelegationID)
	}
}

func TestHandleAgentOrganization_RecentEventsIgnoreUnrelatedAndMalformedLogs(t *testing.T) {
	withIsolatedGatewayLogs(t)
	configPath := writeOrganizationAPIConfig(t, func(cfg *config.Config) {
		cfg.Agents.List = []config.AgentConfig{
			{ID: "worker", Name: "Worker"},
			{ID: "other", Name: "Other"},
		}
	})
	gateway.logs.Append(`{"level":"info","agent_id":"other","message":"other event"}`)
	gateway.logs.Append(`not json mentioning worker without an agent key should not match`)
	gateway.logs.Append(`{"level":"info","message":"missing agent should not match"}`)

	rec := requestAgentOrganization(t, configPath)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp agentOrganizationSnapshot
	decodeJSONResponse(t, rec, &resp)
	if events := resp.Agents["worker"].Activity.RecentEvents; len(events) != 0 {
		t.Fatalf("worker recent events = %+v, want empty", events)
	}
	if events := resp.Agents["other"].Activity.RecentEvents; len(events) != 1 {
		t.Fatalf("other recent events = %+v, want one matching event", events)
	}
}

func TestHandleAgentOrganization_RecentEventsIncludeExplicitTextGatewayLogs(t *testing.T) {
	withIsolatedGatewayLogs(t)
	configPath := writeOrganizationAPIConfig(t, func(cfg *config.Config) {
		cfg.Agents.List = []config.AgentConfig{{ID: "worker", Name: "Worker"}}
	})
	gateway.logs.Append(`10:11:12 INF agent/pipeline.go:129 > turn finished agent_id=worker event=turn_finished status=ok`)

	rec := requestAgentOrganization(t, configPath)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp agentOrganizationSnapshot
	decodeJSONResponse(t, rec, &resp)
	agentState := resp.Agents["worker"]
	if agentState.Status != agentOrganizationStatusIdle {
		t.Fatalf("status = %q, want structured idle status", agentState.Status)
	}
	if len(agentState.Activity.RecentEvents) != 1 {
		t.Fatalf("recent events = %+v, want one text-derived event", agentState.Activity.RecentEvents)
	}
	event := agentState.Activity.RecentEvents[0]
	if event.AgentID != "worker" || event.Level != "info" || event.Event != "turn_finished" {
		t.Fatalf("event = %+v, want parsed text log fields", event)
	}
	if !strings.Contains(event.Message, "turn finished") {
		t.Fatalf("message = %q, want text log content", event.Message)
	}
}

func TestHandleAgentOrganization_TextRecentEventsRequireExactAgentKeyValue(t *testing.T) {
	withIsolatedGatewayLogs(t)
	configPath := writeOrganizationAPIConfig(t, func(cfg *config.Config) {
		cfg.Agents.List = []config.AgentConfig{
			{ID: "worker", Name: "Worker"},
			{ID: "work", Name: "Work"},
		}
	})
	gateway.logs.Append(`10:11:12 INF agent/pipeline.go:129 > turn finished agent_id=workerish event=turn_finished`)
	gateway.logs.Append(`10:11:13 INF agent/pipeline.go:129 > worker completed without explicit key`)

	rec := requestAgentOrganization(t, configPath)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp agentOrganizationSnapshot
	decodeJSONResponse(t, rec, &resp)
	if events := resp.Agents["worker"].Activity.RecentEvents; len(events) != 0 {
		t.Fatalf("worker recent events = %+v, want no partial text matches", events)
	}
	if events := resp.Agents["work"].Activity.RecentEvents; len(events) != 0 {
		t.Fatalf("work recent events = %+v, want no substring text matches", events)
	}
}

func TestHandleAgentOrganization_RecentEventsRedactSensitiveAndLongMessages(t *testing.T) {
	withIsolatedGatewayLogs(t)
	configPath := writeOrganizationAPIConfig(t, func(cfg *config.Config) {
		cfg.Agents.List = []config.AgentConfig{{ID: "worker", Name: "Worker"}}
	})
	longSecretMessage := "authorization=Bearer very-secret-token token=abc123 api_key=sk-private " + strings.Repeat("x", 220)
	gateway.logs.Append(`{"level":"error","agent_id":"worker","message":` + strconv.Quote(longSecretMessage) + `}`)

	rec := requestAgentOrganization(t, configPath)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	for _, secret := range []string{"very-secret-token", "abc123", "sk-private"} {
		if strings.Contains(body, secret) {
			t.Fatalf("response leaked secret %q: %s", secret, body)
		}
	}

	var resp agentOrganizationSnapshot
	decodeJSONResponse(t, rec, &resp)
	events := resp.Agents["worker"].Activity.RecentEvents
	if len(events) != 1 {
		t.Fatalf("recent events = %+v, want one redacted event", events)
	}
	if got := events[0].Message; len(got) > 180 || !strings.Contains(got, "[redacted]") {
		t.Fatalf("redacted message = %q, want bounded message with redaction", got)
	}
}

func TestHandleAgentOrganization_TextRecentEventsRedactSensitiveAndMalformedLines(t *testing.T) {
	withIsolatedGatewayLogs(t)
	configPath := writeOrganizationAPIConfig(t, func(cfg *config.Config) {
		cfg.Agents.List = []config.AgentConfig{{ID: "worker", Name: "Worker"}}
	})
	gateway.logs.Append(`10:11:12 INF agent/pipeline.go:129 > turn finished agent_id=worker token=abc123 api_key=sk-private authorization=Bearer very-secret-token`)
	gateway.logs.Append(`10:11:13 INF agent/pipeline.go:129 > malformed quoted field agent_id="worker`)

	rec := requestAgentOrganization(t, configPath)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	body := rec.Body.String()
	for _, secret := range []string{"very-secret-token", "abc123", "sk-private"} {
		if strings.Contains(body, secret) {
			t.Fatalf("response leaked secret %q: %s", secret, body)
		}
	}

	var resp agentOrganizationSnapshot
	decodeJSONResponse(t, rec, &resp)
	events := resp.Agents["worker"].Activity.RecentEvents
	if len(events) != 1 {
		t.Fatalf("recent events = %+v, want only well-formed text event", events)
	}
	if got := events[0].Message; !strings.Contains(got, "[redacted]") {
		t.Fatalf("redacted message = %q, want redaction", got)
	}
}

func writeOrganizationAPIConfig(t *testing.T, mutate func(*config.Config)) string {
	t.Helper()

	tmp := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.Agents.Defaults.Workspace = filepath.Join(tmp, "workspace")
	cfg.Agents.List = []config.AgentConfig{{ID: "main", Name: "Main"}}
	cfg.ModelList = nil
	if mutate != nil {
		mutate(cfg)
	}
	configPath := filepath.Join(tmp, "config.json")
	if err := config.SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}
	return configPath
}

func loadOrganizationAPIConfig(t *testing.T, configPath string) *config.Config {
	t.Helper()

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	return cfg
}

func requestAgentOrganization(t *testing.T, configPath string) *httptest.ResponseRecorder {
	t.Helper()

	h := NewHandler(configPath)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/agents/organization", nil)
	mux.ServeHTTP(rec, req)
	return rec
}

func withIsolatedGatewayLogs(t *testing.T) {
	t.Helper()

	previous := gateway.logs
	gateway.logs = NewLogBuffer(200)
	gateway.logs.Reset()
	t.Cleanup(func() {
		gateway.logs = previous
	})
}

func decodeJSONResponse(t *testing.T, rec *httptest.ResponseRecorder, out any) {
	t.Helper()

	if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
		t.Fatalf("Unmarshal() error = %v, body=%s", err, rec.Body.String())
	}
}

func agentNodeIDs(nodes []agentOrganizationNode) []string {
	ids := make([]string, 0, len(nodes))
	for _, node := range nodes {
		ids = append(ids, node.ID)
	}
	return ids
}

func setMeetingCreatedAt(
	t *testing.T,
	cfg *config.Config,
	rec *agent.AgentMeetingRecord,
	createdAt time.Time,
) {
	t.Helper()

	rec.CreatedAt = createdAt
	rec.UpdatedAt = createdAt
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}
	path := filepath.Join(cfg.WorkspacePath(), "meetings", rec.MeetingID+".json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func setDelegationTimes(
	t *testing.T,
	cfg *config.Config,
	rec *agent.AgentDelegationRecord,
	createdAt time.Time,
	completedAt time.Time,
) {
	t.Helper()

	rec.CreatedAt = createdAt
	rec.UpdatedAt = completedAt
	rec.CompletedAt = &completedAt
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}
	path := filepath.Join(cfg.WorkspacePath(), "delegations", rec.DelegationID+".json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
