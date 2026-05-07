package api

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/sipeed/picoclaw/pkg/agent"
	"github.com/sipeed/picoclaw/pkg/config"
)

func TestHandleAgentInbox_ReturnsOnlyTargetAgentRecords(t *testing.T) {
	configPath := writeOrganizationAPIConfig(t, func(cfg *config.Config) {
		cfg.Agents.List = []config.AgentConfig{
			{ID: "lead", Name: "Lead"},
			{ID: "worker", Name: "Worker"},
			{ID: "other", Name: "Other"},
		}
	})
	cfg := loadOrganizationAPIConfig(t, configPath)
	store := agent.NewDelegationRecordStore(filepath.Join(cfg.WorkspacePath(), "delegations"), nil)
	visible := writeDelegationRecord(t, store, agent.AgentDelegationRequest{
		ParentAgentID: "lead",
		TargetAgentID: "worker",
		Task:          "private worker prompt must not leak",
	})
	unrelated := writeDelegationRecord(t, store, agent.AgentDelegationRequest{
		ParentAgentID: "lead",
		TargetAgentID: "other",
		Task:          "private other prompt must not leak",
	})

	rec := requestAgentActivity(t, configPath, "/api/agents/worker/inbox")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), visible.DelegationID) {
		t.Fatalf("response missing target delegation %q: %s", visible.DelegationID, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), unrelated.DelegationID) {
		t.Fatalf("response leaked unrelated delegation %q: %s", unrelated.DelegationID, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "private worker prompt") || strings.Contains(rec.Body.String(), "private other prompt") {
		t.Fatalf("response leaked private delegation prompt: %s", rec.Body.String())
	}
}

func TestHandleAgentOutbox_ReturnsOnlyRequesterRecords(t *testing.T) {
	configPath := writeOrganizationAPIConfig(t, func(cfg *config.Config) {
		cfg.Agents.List = []config.AgentConfig{
			{ID: "lead", Name: "Lead"},
			{ID: "worker", Name: "Worker"},
			{ID: "other", Name: "Other"},
		}
	})
	cfg := loadOrganizationAPIConfig(t, configPath)
	store := agent.NewDelegationRecordStore(filepath.Join(cfg.WorkspacePath(), "delegations"), nil)
	visible := writeDelegationRecord(t, store, agent.AgentDelegationRequest{
		ParentAgentID: "lead",
		TargetAgentID: "worker",
		Task:          "private requested prompt must not leak",
	})
	unrelated := writeDelegationRecord(t, store, agent.AgentDelegationRequest{
		ParentAgentID: "other",
		TargetAgentID: "worker",
		Task:          "private assigned prompt must not leak",
	})

	rec := requestAgentActivity(t, configPath, "/api/agents/lead/outbox")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), visible.DelegationID) {
		t.Fatalf("response missing requester delegation %q: %s", visible.DelegationID, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), unrelated.DelegationID) {
		t.Fatalf("response leaked unrelated delegation %q: %s", unrelated.DelegationID, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "private requested prompt") || strings.Contains(rec.Body.String(), "private assigned prompt") {
		t.Fatalf("response leaked private delegation prompt: %s", rec.Body.String())
	}
}

func TestHandleAgentMeetings_ReturnsOnlyRelatedMeetings(t *testing.T) {
	configPath := writeOrganizationAPIConfig(t, func(cfg *config.Config) {
		cfg.Agents.List = []config.AgentConfig{
			{ID: "sponsor", Name: "Sponsor"},
			{ID: "chair", Name: "Chair"},
			{ID: "participant", Name: "Participant"},
			{ID: "other", Name: "Other"},
		}
	})
	cfg := loadOrganizationAPIConfig(t, configPath)
	store := agent.NewMeetingRecordStore(filepath.Join(cfg.WorkspacePath(), "meetings"), nil)
	visible := writeMeetingRecord(t, store, agent.AgentMeetingRequest{
		Title:               "Visible Meeting",
		SponsorAgentID:      "sponsor",
		ChairAgentID:        "chair",
		ParticipantAgentIDs: []string{"participant"},
		Goal:                "private meeting goal must not leak",
		Notes:               "private notes must not leak",
	})
	unrelated := writeMeetingRecord(t, store, agent.AgentMeetingRequest{
		Title:               "Unrelated Meeting",
		SponsorAgentID:      "other",
		ChairAgentID:        "other",
		ParticipantAgentIDs: []string{"other"},
		Goal:                "private unrelated goal must not leak",
	})

	rec := requestAgentActivity(t, configPath, "/api/agents/participant/meetings")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), visible.MeetingID) {
		t.Fatalf("response missing related meeting %q: %s", visible.MeetingID, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), unrelated.MeetingID) || strings.Contains(rec.Body.String(), "Unrelated Meeting") {
		t.Fatalf("response leaked unrelated meeting: %s", rec.Body.String())
	}
	for _, leaked := range []string{"private meeting goal", "private notes", "private unrelated goal"} {
		if strings.Contains(rec.Body.String(), leaked) {
			t.Fatalf("response leaked %q: %s", leaked, rec.Body.String())
		}
	}
}

func TestHandleAgentActivity_UnknownAgentReturnsClientError(t *testing.T) {
	configPath := writeOrganizationAPIConfig(t, func(cfg *config.Config) {
		cfg.Agents.List = []config.AgentConfig{{ID: "known", Name: "Known"}}
	})

	rec := requestAgentActivity(t, configPath, "/api/agents/missing/inbox")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "unknown agent") {
		t.Fatalf("response = %q, want clear unknown-agent error", rec.Body.String())
	}
}

func TestHandleAgentActivity_MissingStoresReturnEmptyLists(t *testing.T) {
	configPath := writeOrganizationAPIConfig(t, func(cfg *config.Config) {
		cfg.Agents.List = []config.AgentConfig{{ID: "known", Name: "Known"}}
	})

	for _, path := range []string{
		"/api/agents/known/inbox",
		"/api/agents/known/outbox",
		"/api/agents/known/meetings",
	} {
		t.Run(path, func(t *testing.T) {
			rec := requestAgentActivity(t, configPath, path)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), `"records":[]`) {
				t.Fatalf("response = %s, want empty records list", rec.Body.String())
			}
		})
	}
}

func TestHandleAgentActivity_LimitUsesNewestFirstStableOrder(t *testing.T) {
	configPath := writeOrganizationAPIConfig(t, func(cfg *config.Config) {
		cfg.Agents.List = []config.AgentConfig{
			{ID: "lead", Name: "Lead"},
			{ID: "worker", Name: "Worker"},
		}
	})
	cfg := loadOrganizationAPIConfig(t, configPath)
	store := agent.NewDelegationRecordStore(filepath.Join(cfg.WorkspacePath(), "delegations"), nil)
	older := writeDelegationRecord(t, store, agent.AgentDelegationRequest{
		ParentAgentID: "lead",
		TargetAgentID: "worker",
		Task:          "older",
	})
	setDelegationCreatedAt(t, cfg, &older, time.Date(2026, 5, 7, 10, 0, 0, 0, time.UTC))
	newer := writeDelegationRecord(t, store, agent.AgentDelegationRequest{
		ParentAgentID: "lead",
		TargetAgentID: "worker",
		Task:          "newer",
	})
	setDelegationCreatedAt(t, cfg, &newer, time.Date(2026, 5, 7, 11, 0, 0, 0, time.UTC))

	rec := requestAgentActivity(t, configPath, "/api/agents/worker/inbox?limit=1")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), newer.DelegationID) {
		t.Fatalf("response missing newest delegation %q: %s", newer.DelegationID, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), older.DelegationID) {
		t.Fatalf("response included older delegation despite limit=1: %s", rec.Body.String())
	}
}

func TestHandleAgentActivitySummary_ReturnsOrgPageStateForAgent(t *testing.T) {
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
	delegation := writeDelegationRecord(t, delegations, agent.AgentDelegationRequest{
		ParentAgentID: "lead",
		TargetAgentID: "worker",
		Task:          "private activity prompt must not leak",
	})
	if err := delegations.Running(context.Background(), delegation.DelegationID); err != nil {
		t.Fatalf("Running() error = %v", err)
	}
	meetings := agent.NewMeetingRecordStore(filepath.Join(cfg.WorkspacePath(), "meetings"), nil)
	meeting := writeMeetingRecord(t, meetings, agent.AgentMeetingRequest{
		Title:               "Planning",
		SponsorAgentID:      "lead",
		ChairAgentID:        "chair",
		ParticipantAgentIDs: []string{"worker"},
		Goal:                "private meeting goal must not leak",
	})
	gateway.logs.Append(`{"level":"info","time":"2026-05-07T10:11:12Z","agent_id":"worker","event":"turn_finished","message":"worker turn finished"}`)

	rec := requestAgentActivity(t, configPath, "/api/agents/worker/activity")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	for _, leaked := range []string{"private activity prompt", "private meeting goal"} {
		if strings.Contains(rec.Body.String(), leaked) {
			t.Fatalf("response leaked %q: %s", leaked, rec.Body.String())
		}
	}

	var resp agentOrganizationAgent
	decodeJSONResponse(t, rec, &resp)
	if resp.ID != "worker" {
		t.Fatalf("agent id = %q, want worker", resp.ID)
	}
	if resp.Status != agentOrganizationStatusMeeting {
		t.Fatalf("status = %q, want %q", resp.Status, agentOrganizationStatusMeeting)
	}
	if resp.Activity.InboxCount != 1 || resp.Activity.OutboxCount != 0 || resp.Activity.MeetingCount != 1 {
		t.Fatalf("activity counts = %+v, want one inbox and one meeting", resp.Activity)
	}
	if resp.Activity.Current == nil || resp.Activity.Current.RecordID != meeting.MeetingID {
		t.Fatalf("current = %+v, want meeting %q", resp.Activity.Current, meeting.MeetingID)
	}
	if len(resp.Activity.RecentEvents) != 1 || resp.Activity.RecentEvents[0].Event != "turn_finished" {
		t.Fatalf("recent events = %+v, want matching gateway event", resp.Activity.RecentEvents)
	}
}

func TestAgentOrganizationReadOnlyFlowDoesNotRewriteFiles(t *testing.T) {
	withIsolatedGatewayLogs(t)
	configPath := writeOrganizationAPIConfig(t, func(cfg *config.Config) {
		cfg.Agents.List = []config.AgentConfig{
			{ID: "lead", Name: "Lead"},
			{ID: "worker", Name: "Worker"},
		}
	})
	cfg := loadOrganizationAPIConfig(t, configPath)
	store := agent.NewDelegationRecordStore(filepath.Join(cfg.WorkspacePath(), "delegations"), nil)
	writeDelegationRecord(t, store, agent.AgentDelegationRequest{
		ParentAgentID: "lead",
		TargetAgentID: "worker",
		Task:          "read-only check",
	})
	before := snapshotRegularFiles(t, filepath.Dir(configPath))

	for _, path := range []string{
		"/api/agents/organization",
		"/api/agents/worker/activity",
		"/api/agents/worker/inbox",
		"/api/agents/lead/outbox",
		"/api/agents/worker/meetings",
	} {
		rec := requestAgentActivity(t, configPath, path)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want %d, body=%s", path, rec.Code, http.StatusOK, rec.Body.String())
		}
	}

	after := snapshotRegularFiles(t, filepath.Dir(configPath))
	if strings.Join(before, "\n") != strings.Join(after, "\n") {
		t.Fatalf("read-only flow changed files\nbefore:\n%s\nafter:\n%s", strings.Join(before, "\n"), strings.Join(after, "\n"))
	}
}

func writeDelegationRecord(
	t *testing.T,
	store *agent.DelegationRecordStore,
	req agent.AgentDelegationRequest,
) agent.AgentDelegationRecord {
	t.Helper()

	rec, err := store.Requested(context.Background(), req)
	if err != nil {
		t.Fatalf("Requested() error = %v", err)
	}
	return rec
}

func writeMeetingRecord(
	t *testing.T,
	store *agent.MeetingRecordStore,
	req agent.AgentMeetingRequest,
) agent.AgentMeetingRecord {
	t.Helper()

	rec, err := store.Started(context.Background(), req)
	if err != nil {
		t.Fatalf("Started() error = %v", err)
	}
	return rec
}

func setDelegationCreatedAt(
	t *testing.T,
	cfg *config.Config,
	rec *agent.AgentDelegationRecord,
	createdAt time.Time,
) {
	t.Helper()

	rec.CreatedAt = createdAt
	rec.UpdatedAt = createdAt
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent() error = %v", err)
	}
	path := filepath.Join(cfg.WorkspacePath(), "delegations", rec.DelegationID+".json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func requestAgentActivity(t *testing.T, configPath string, path string) *httptest.ResponseRecorder {
	t.Helper()

	h := NewHandler(configPath)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	mux.ServeHTTP(rec, req)
	return rec
}

func snapshotRegularFiles(t *testing.T, root string) []string {
	t.Helper()

	var snapshot []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		snapshot = append(snapshot, fmt.Sprintf("%s %x", filepath.ToSlash(rel), sum))
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir(%q) error = %v", root, err)
	}
	sort.Strings(snapshot)
	return snapshot
}
