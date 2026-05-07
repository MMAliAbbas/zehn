package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
