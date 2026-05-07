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

func TestHandleAgentOrganization_RecentEventsIgnoreUnrelatedAndMalformedLogs(t *testing.T) {
	withIsolatedGatewayLogs(t)
	configPath := writeOrganizationAPIConfig(t, func(cfg *config.Config) {
		cfg.Agents.List = []config.AgentConfig{
			{ID: "worker", Name: "Worker"},
			{ID: "other", Name: "Other"},
		}
	})
	gateway.logs.Append(`{"level":"info","agent_id":"other","message":"other event"}`)
	gateway.logs.Append(`not json agent_id=worker message="should not break the page"`)
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
