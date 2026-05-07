package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

func TestHandleAgentOrganization_FailedRecordTakesPrecedence(t *testing.T) {
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
