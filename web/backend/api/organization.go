package api

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/sipeed/picoclaw/pkg/agent"
	"github.com/sipeed/picoclaw/pkg/config"
)

const (
	agentOrganizationStatusIdle       = "idle"
	agentOrganizationStatusWorking    = "working"
	agentOrganizationStatusDelegating = "delegating"
	agentOrganizationStatusMeeting    = "meeting"
	agentOrganizationStatusFailed     = "failed"
)

type agentOrganizationSnapshot struct {
	Roots    []agentOrganizationNode           `json:"roots"`
	Agents   map[string]agentOrganizationAgent `json:"agents"`
	Activity agentOrganizationActivitySummary  `json:"activity"`
	Metadata agentOrganizationSnapshotMetadata `json:"metadata"`
}

type agentOrganizationSnapshotMetadata struct {
	Source       string    `json:"source"`
	GeneratedAt  time.Time `json:"generated_at"`
	HasHierarchy bool      `json:"has_hierarchy"`
}

type agentOrganizationNode struct {
	ID        string                         `json:"id"`
	Name      string                         `json:"name,omitempty"`
	Label     string                         `json:"label,omitempty"`
	Group     string                         `json:"group,omitempty"`
	Workspace string                         `json:"workspace,omitempty"`
	Status    string                         `json:"status"`
	Activity  agentOrganizationAgentActivity `json:"activity"`
	Children  []agentOrganizationNode        `json:"children,omitempty"`
}

type agentOrganizationAgent struct {
	ID        string                         `json:"id"`
	Name      string                         `json:"name,omitempty"`
	Label     string                         `json:"label,omitempty"`
	Group     string                         `json:"group,omitempty"`
	Workspace string                         `json:"workspace,omitempty"`
	Status    string                         `json:"status"`
	Activity  agentOrganizationAgentActivity `json:"activity"`
}

type agentOrganizationAgentActivity struct {
	InboxCount    int                              `json:"inbox_count"`
	OutboxCount   int                              `json:"outbox_count"`
	MeetingCount  int                              `json:"meeting_count"`
	FailureCount  int                              `json:"failure_count"`
	RecentEvents  []agentOrganizationRecentEvent   `json:"recent_events"`
	Current       *agentOrganizationActivityRecord `json:"current,omitempty"`
	LastFailure   *agentOrganizationActivityRecord `json:"last_failure,omitempty"`
	LastUpdatedAt *time.Time                       `json:"last_updated_at,omitempty"`
}

type agentOrganizationActivityRecord struct {
	Type         string     `json:"type"`
	RecordID     string     `json:"record_id"`
	Status       string     `json:"status"`
	Role         string     `json:"role,omitempty"`
	AgentID      string     `json:"agent_id,omitempty"`
	ArtifactRefs []string   `json:"artifact_refs,omitempty"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

type agentOrganizationRecentEvent struct {
	Source    string     `json:"source"`
	AgentID   string     `json:"agent_id"`
	Level     string     `json:"level,omitempty"`
	Event     string     `json:"event,omitempty"`
	Message   string     `json:"message"`
	Timestamp *time.Time `json:"timestamp,omitempty"`
}

type agentOrganizationActivitySummary struct {
	DelegationCount int                             `json:"delegation_count"`
	MeetingCount    int                             `json:"meeting_count"`
	FailureCount    int                             `json:"failure_count"`
	ActiveCount     int                             `json:"active_count"`
	Recent          []agentOrganizationActivityFeed `json:"recent"`
}

type agentOrganizationActivityFeed struct {
	Type      string     `json:"type"`
	AgentID   string     `json:"agent_id,omitempty"`
	RecordID  string     `json:"record_id,omitempty"`
	Status    string     `json:"status,omitempty"`
	Summary   string     `json:"summary,omitempty"`
	Timestamp *time.Time `json:"timestamp,omitempty"`
}

type agentActivityListResponse[T any] struct {
	AgentID string `json:"agent_id"`
	Kind    string `json:"kind"`
	Limit   int    `json:"limit"`
	Records []T    `json:"records"`
}

type agentDelegationActivitySummary struct {
	DelegationID  string     `json:"delegation_id"`
	Status        string     `json:"status"`
	ParentAgentID string     `json:"parent_agent_id"`
	TargetAgentID string     `json:"target_agent_id"`
	RequesterID   string     `json:"requester_id,omitempty"`
	Role          string     `json:"role"`
	Mode          string     `json:"mode,omitempty"`
	Priority      string     `json:"priority,omitempty"`
	ArtifactRefs  []string   `json:"artifact_refs,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}

type agentMeetingActivitySummary struct {
	MeetingID      string     `json:"meeting_id"`
	Status         string     `json:"status"`
	Title          string     `json:"title,omitempty"`
	SponsorAgentID string     `json:"sponsor_agent_id"`
	ChairAgentID   string     `json:"chair_agent_id"`
	Participants   []string   `json:"participants,omitempty"`
	Role           string     `json:"role"`
	ArtifactRefs   []string   `json:"artifact_refs,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

type agentOrganizationBuildState struct {
	agents      map[string]*agentOrganizationAgent
	delegations []agent.AgentDelegationRecord
	meetings    []agent.AgentMeetingRecord
	summary     agentOrganizationActivitySummary
}

// registerAgentOrganizationRoutes binds read-only configured agent organization endpoints.
func (h *Handler) registerAgentOrganizationRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/agents/organization", h.handleGetAgentOrganization)
	mux.HandleFunc("GET /api/agents/{id}/activity", h.handleGetAgentActivity)
	mux.HandleFunc("GET /api/agents/{id}/inbox", h.handleGetAgentInbox)
	mux.HandleFunc("GET /api/agents/{id}/outbox", h.handleGetAgentOutbox)
	mux.HandleFunc("GET /api/agents/{id}/meetings", h.handleGetAgentMeetings)
	mux.HandleFunc("GET /api/agents/{id}/failures", h.handleGetAgentFailures)
}

// handleGetAgentOrganization returns a normalized configured agent hierarchy plus structured activity.
//
//	GET /api/agents/organization
func (h *Handler) handleGetAgentOrganization(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}

	resp, err := buildAgentOrganizationSnapshot(r.Context(), cfg)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load agent organization: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) handleGetAgentActivity(w http.ResponseWriter, r *http.Request) {
	cfg, agentID, _, ok := h.loadAgentActivityRequest(w, r)
	if !ok {
		return
	}

	snapshot, err := buildAgentOrganizationSnapshot(r.Context(), cfg)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load agent activity: %v", err), http.StatusInternalServerError)
		return
	}
	agentState, ok := snapshot.Agents[agentID]
	if !ok {
		http.Error(w, fmt.Sprintf("unknown agent %q", agentID), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(agentState); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) handleGetAgentInbox(w http.ResponseWriter, r *http.Request) {
	cfg, agentID, limit, ok := h.loadAgentActivityRequest(w, r)
	if !ok {
		return
	}

	records, err := agent.NewDelegationRecordStore(
		filepath.Join(cfg.WorkspacePath(), "delegations"),
		nil,
	).List(r.Context(), agent.AgentDelegationRecordQuery{
		VisibleToAgentID: agentID,
		TargetAgentID:    agentID,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load agent inbox: %v", err), http.StatusInternalServerError)
		return
	}

	summaries := make([]agentDelegationActivitySummary, 0, min(len(records), limit))
	for _, rec := range newestDelegationRecords(records, limit) {
		summaries = append(summaries, summarizeDelegationActivity(rec, agentID, "target"))
	}
	writeAgentActivityResponse(w, agentActivityListResponse[agentDelegationActivitySummary]{
		AgentID: agentID,
		Kind:    "inbox",
		Limit:   limit,
		Records: summaries,
	})
}

func (h *Handler) handleGetAgentOutbox(w http.ResponseWriter, r *http.Request) {
	cfg, agentID, limit, ok := h.loadAgentActivityRequest(w, r)
	if !ok {
		return
	}

	records, err := agent.NewDelegationRecordStore(
		filepath.Join(cfg.WorkspacePath(), "delegations"),
		nil,
	).List(r.Context(), agent.AgentDelegationRecordQuery{IncludePrivateAll: true})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load agent outbox: %v", err), http.StatusInternalServerError)
		return
	}

	related := make([]agent.AgentDelegationRecord, 0, len(records))
	for _, rec := range records {
		if delegationRequesterID(rec) == agentID {
			related = append(related, rec)
		}
	}
	summaries := make([]agentDelegationActivitySummary, 0, min(len(related), limit))
	for _, rec := range newestDelegationRecords(related, limit) {
		summaries = append(summaries, summarizeDelegationActivity(rec, agentID, "requester"))
	}
	writeAgentActivityResponse(w, agentActivityListResponse[agentDelegationActivitySummary]{
		AgentID: agentID,
		Kind:    "outbox",
		Limit:   limit,
		Records: summaries,
	})
}

func (h *Handler) handleGetAgentMeetings(w http.ResponseWriter, r *http.Request) {
	cfg, agentID, limit, ok := h.loadAgentActivityRequest(w, r)
	if !ok {
		return
	}

	records, err := agent.NewMeetingRecordStore(
		filepath.Join(cfg.WorkspacePath(), "meetings"),
		nil,
	).List(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load agent meetings: %v", err), http.StatusInternalServerError)
		return
	}

	related := make([]agent.AgentMeetingRecord, 0, len(records))
	for _, rec := range records {
		if _, ok := meetingParticipantIDs(rec)[agentID]; ok {
			related = append(related, rec)
		}
	}
	summaries := make([]agentMeetingActivitySummary, 0, min(len(related), limit))
	for _, rec := range newestMeetingRecords(related, limit) {
		summaries = append(summaries, summarizeMeetingActivity(rec, agentID))
	}
	writeAgentActivityResponse(w, agentActivityListResponse[agentMeetingActivitySummary]{
		AgentID: agentID,
		Kind:    "meetings",
		Limit:   limit,
		Records: summaries,
	})
}

func (h *Handler) handleGetAgentFailures(w http.ResponseWriter, r *http.Request) {
	cfg, agentID, limit, ok := h.loadAgentActivityRequest(w, r)
	if !ok {
		return
	}

	delegations, err := agent.NewDelegationRecordStore(
		filepath.Join(cfg.WorkspacePath(), "delegations"),
		nil,
	).List(r.Context(), agent.AgentDelegationRecordQuery{IncludePrivateAll: true})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load agent failures: %v", err), http.StatusInternalServerError)
		return
	}
	meetings, err := agent.NewMeetingRecordStore(
		filepath.Join(cfg.WorkspacePath(), "meetings"),
		nil,
	).List(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load agent failures: %v", err), http.StatusInternalServerError)
		return
	}

	records := make([]agentOrganizationActivityRecord, 0)
	for _, rec := range delegations {
		if rec.Status != agent.AgentDelegationStatusFailed {
			continue
		}
		role := delegationRoleForAgent(rec, agentID)
		if role != "target" && role != "requester" {
			continue
		}
		records = append(records, summarizeDelegationFailureActivity(rec, agentID, role))
	}
	for _, rec := range meetings {
		if rec.Status != agent.AgentMeetingStatusFailed {
			continue
		}
		if _, ok := meetingParticipantIDs(rec)[agentID]; !ok {
			continue
		}
		records = append(records, summarizeMeetingFailureActivity(rec, agentID))
	}

	writeAgentActivityResponse(w, agentActivityListResponse[agentOrganizationActivityRecord]{
		AgentID: agentID,
		Kind:    "failures",
		Limit:   limit,
		Records: newestActivityRecords(records, limit),
	})
}

func (h *Handler) loadAgentActivityRequest(
	w http.ResponseWriter,
	r *http.Request,
) (*config.Config, string, int, bool) {
	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return nil, "", 0, false
	}

	agentID := strings.TrimSpace(r.PathValue("id"))
	if !configuredAgentExists(cfg, agentID) {
		http.Error(w, fmt.Sprintf("unknown agent %q", agentID), http.StatusNotFound)
		return nil, "", 0, false
	}
	return cfg, agentID, parseAgentActivityLimit(r), true
}

func configuredAgentExists(cfg *config.Config, agentID string) bool {
	if cfg == nil || agentID == "" {
		return false
	}
	return slices.ContainsFunc(cfg.Agents.List, func(configured config.AgentConfig) bool {
		return strings.TrimSpace(configured.ID) == agentID
	})
}

func parseAgentActivityLimit(r *http.Request) int {
	const (
		defaultLimit = 50
		maxLimit     = 100
	)
	raw := strings.TrimSpace(r.URL.Query().Get("limit"))
	if raw == "" {
		return defaultLimit
	}
	limit, err := strconv.Atoi(raw)
	if err != nil || limit <= 0 {
		return defaultLimit
	}
	return min(limit, maxLimit)
}

func newestDelegationRecords(records []agent.AgentDelegationRecord, limit int) []agent.AgentDelegationRecord {
	records = append([]agent.AgentDelegationRecord(nil), records...)
	slices.SortFunc(records, func(a, b agent.AgentDelegationRecord) int {
		if byCreated := cmp.Compare(b.CreatedAt.UnixNano(), a.CreatedAt.UnixNano()); byCreated != 0 {
			return byCreated
		}
		return cmp.Compare(b.DelegationID, a.DelegationID)
	})
	if len(records) > limit {
		records = records[:limit]
	}
	return records
}

func newestMeetingRecords(records []agent.AgentMeetingRecord, limit int) []agent.AgentMeetingRecord {
	records = append([]agent.AgentMeetingRecord(nil), records...)
	slices.SortFunc(records, func(a, b agent.AgentMeetingRecord) int {
		if byCreated := cmp.Compare(b.CreatedAt.UnixNano(), a.CreatedAt.UnixNano()); byCreated != 0 {
			return byCreated
		}
		return cmp.Compare(b.MeetingID, a.MeetingID)
	})
	if len(records) > limit {
		records = records[:limit]
	}
	return records
}

func newestActivityRecords(records []agentOrganizationActivityRecord, limit int) []agentOrganizationActivityRecord {
	records = append([]agentOrganizationActivityRecord(nil), records...)
	slices.SortFunc(records, func(a, b agentOrganizationActivityRecord) int {
		if byUpdated := cmp.Compare(activityRecordUnixNano(b), activityRecordUnixNano(a)); byUpdated != 0 {
			return byUpdated
		}
		if byCreated := cmp.Compare(activityRecordCreatedUnixNano(b), activityRecordCreatedUnixNano(a)); byCreated != 0 {
			return byCreated
		}
		if byType := cmp.Compare(a.Type, b.Type); byType != 0 {
			return byType
		}
		return cmp.Compare(a.RecordID, b.RecordID)
	})
	if len(records) > limit {
		records = records[:limit]
	}
	return records
}

func summarizeDelegationActivity(
	rec agent.AgentDelegationRecord,
	agentID string,
	role string,
) agentDelegationActivitySummary {
	if role == "" {
		role = delegationRoleForAgent(rec, agentID)
	}
	return agentDelegationActivitySummary{
		DelegationID:  rec.DelegationID,
		Status:        string(rec.Status),
		ParentAgentID: rec.ParentAgentID,
		TargetAgentID: rec.TargetAgentID,
		RequesterID:   delegationRequesterID(rec),
		Role:          role,
		Mode:          rec.Request.Mode,
		Priority:      rec.Request.Priority,
		ArtifactRefs:  append([]string(nil), rec.Request.ArtifactRefs...),
		CreatedAt:     rec.CreatedAt,
		UpdatedAt:     rec.UpdatedAt,
		StartedAt:     rec.StartedAt,
		CompletedAt:   rec.CompletedAt,
	}
}

func summarizeDelegationFailureActivity(
	rec agent.AgentDelegationRecord,
	agentID string,
	role string,
) agentOrganizationActivityRecord {
	activity := organizationRecordActivity(
		"delegation",
		rec.DelegationID,
		string(rec.Status),
		rec.CreatedAt,
		rec.UpdatedAt,
		rec.CompletedAt,
		rec.Request.ArtifactRefs,
	)
	activity.Role = role
	if role == "target" {
		activity.AgentID = delegationRequesterID(rec)
	} else {
		activity.AgentID = strings.TrimSpace(rec.TargetAgentID)
	}
	return activity
}

func delegationRoleForAgent(rec agent.AgentDelegationRecord, agentID string) string {
	switch agentID {
	case rec.TargetAgentID:
		return "target"
	case delegationRequesterID(rec):
		return "requester"
	default:
		return "visible"
	}
}

func summarizeMeetingActivity(rec agent.AgentMeetingRecord, agentID string) agentMeetingActivitySummary {
	role := meetingParticipantIDs(rec)[agentID]
	return agentMeetingActivitySummary{
		MeetingID:      rec.MeetingID,
		Status:         string(rec.Status),
		Title:          rec.Title,
		SponsorAgentID: rec.SponsorAgentID,
		ChairAgentID:   rec.ChairAgentID,
		Participants:   append([]string(nil), rec.Participants...),
		Role:           role,
		ArtifactRefs:   append([]string(nil), rec.ArtifactRefs...),
		CreatedAt:      rec.CreatedAt,
		UpdatedAt:      rec.UpdatedAt,
		CompletedAt:    rec.CompletedAt,
	}
}

func summarizeMeetingFailureActivity(rec agent.AgentMeetingRecord, agentID string) agentOrganizationActivityRecord {
	activity := organizationRecordActivity(
		"meeting",
		rec.MeetingID,
		string(rec.Status),
		rec.CreatedAt,
		rec.UpdatedAt,
		rec.CompletedAt,
		rec.ArtifactRefs,
	)
	activity.Role = meetingParticipantIDs(rec)[agentID]
	switch activity.Role {
	case "chair":
		activity.AgentID = strings.TrimSpace(rec.SponsorAgentID)
	default:
		activity.AgentID = strings.TrimSpace(rec.ChairAgentID)
	}
	return activity
}

func writeAgentActivityResponse[T any](w http.ResponseWriter, resp agentActivityListResponse[T]) {
	if resp.Records == nil {
		resp.Records = []T{}
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func buildAgentOrganizationSnapshot(ctx context.Context, cfg *config.Config) (agentOrganizationSnapshot, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	if err := cfg.ValidateAgentOrganization(); err != nil {
		return agentOrganizationSnapshot{}, err
	}

	delegations, err := agent.NewDelegationRecordStore(
		filepath.Join(cfg.WorkspacePath(), "delegations"),
		nil,
	).List(ctx, agent.AgentDelegationRecordQuery{IncludePrivateAll: true})
	if err != nil {
		return agentOrganizationSnapshot{}, err
	}
	meetings, err := agent.NewMeetingRecordStore(
		filepath.Join(cfg.WorkspacePath(), "meetings"),
		nil,
	).List(ctx)
	if err != nil {
		return agentOrganizationSnapshot{}, err
	}

	state := agentOrganizationBuildState{
		agents:      buildAgentOrganizationAgentMap(cfg),
		delegations: delegations,
		meetings:    meetings,
	}
	state.applyActivity()
	recentEvents := gatewayLogRecentEvents(state.agents)
	state.applyRecentEvents(recentEvents)
	state.summary.Recent = state.recentActivityFeed(recentEvents)

	snapshot := agentOrganizationSnapshot{
		Roots:    buildAgentOrganizationRoots(cfg, state.agents),
		Agents:   dereferenceOrganizationAgents(state.agents),
		Activity: state.summary,
		Metadata: agentOrganizationSnapshotMetadata{
			Source:       "launcher_config",
			GeneratedAt:  time.Now().UTC(),
			HasHierarchy: cfg.Agents.Organization != nil,
		},
	}
	if snapshot.Roots == nil {
		snapshot.Roots = []agentOrganizationNode{}
	}
	if snapshot.Agents == nil {
		snapshot.Agents = map[string]agentOrganizationAgent{}
	}
	if snapshot.Activity.Recent == nil {
		snapshot.Activity.Recent = []agentOrganizationActivityFeed{}
	}
	return snapshot, nil
}

func buildAgentOrganizationAgentMap(cfg *config.Config) map[string]*agentOrganizationAgent {
	agents := make(map[string]*agentOrganizationAgent, len(cfg.Agents.List))
	for _, configured := range cfg.Agents.List {
		agentID := strings.TrimSpace(configured.ID)
		if agentID == "" {
			continue
		}
		node := organizationNodeForAgent(cfg.Agents.Organization, agentID)
		agents[agentID] = &agentOrganizationAgent{
			ID:        agentID,
			Name:      strings.TrimSpace(configured.Name),
			Label:     organizationLabel(configured, node),
			Group:     strings.TrimSpace(node.Group),
			Workspace: organizationAgentWorkspace(cfg, configured),
			Status:    agentOrganizationStatusIdle,
			Activity: agentOrganizationAgentActivity{
				RecentEvents: []agentOrganizationRecentEvent{},
			},
		}
	}
	return agents
}

func organizationAgentWorkspace(cfg *config.Config, agentCfg config.AgentConfig) string {
	if workspace := strings.TrimSpace(agentCfg.Workspace); workspace != "" {
		return workspace
	}
	return cfg.WorkspacePath()
}

func organizationNodeForAgent(
	org *config.AgentOrganizationConfig,
	agentID string,
) config.AgentOrganizationNodeConfig {
	if org == nil {
		return config.AgentOrganizationNodeConfig{}
	}
	for _, node := range org.Nodes {
		if strings.TrimSpace(node.AgentID) == agentID {
			return node
		}
	}
	return config.AgentOrganizationNodeConfig{}
}

func organizationLabel(agentCfg config.AgentConfig, node config.AgentOrganizationNodeConfig) string {
	if label := strings.TrimSpace(node.Label); label != "" {
		return label
	}
	return strings.TrimSpace(agentCfg.Name)
}

func buildAgentOrganizationRoots(
	cfg *config.Config,
	agents map[string]*agentOrganizationAgent,
) []agentOrganizationNode {
	if len(agents) == 0 {
		return nil
	}

	rootIDs := organizationRootIDs(cfg)
	seen := make(map[string]struct{}, len(rootIDs))
	roots := make([]agentOrganizationNode, 0, len(rootIDs))
	for _, agentID := range rootIDs {
		if _, ok := seen[agentID]; ok {
			continue
		}
		node, ok := buildAgentOrganizationNode(cfg, agents, agentID)
		if !ok {
			continue
		}
		roots = append(roots, node)
		seen[agentID] = struct{}{}
	}
	return roots
}

func organizationRootIDs(cfg *config.Config) []string {
	if cfg.Agents.Organization == nil {
		ids := make([]string, 0, len(cfg.Agents.List))
		for _, configured := range cfg.Agents.List {
			if agentID := strings.TrimSpace(configured.ID); agentID != "" {
				ids = append(ids, agentID)
			}
		}
		return ids
	}

	ids := cfg.Agents.Organization.RootAgentIDs()
	nodeAgents := make(map[string]struct{}, len(cfg.Agents.Organization.Nodes))
	for _, node := range cfg.Agents.Organization.Nodes {
		nodeAgents[strings.TrimSpace(node.AgentID)] = struct{}{}
	}
	for _, configured := range cfg.Agents.List {
		agentID := strings.TrimSpace(configured.ID)
		if agentID == "" {
			continue
		}
		if _, ok := nodeAgents[agentID]; !ok {
			ids = append(ids, agentID)
		}
	}
	return ids
}

func buildAgentOrganizationNode(
	cfg *config.Config,
	agents map[string]*agentOrganizationAgent,
	agentID string,
) (agentOrganizationNode, bool) {
	agentState, ok := agents[agentID]
	if !ok {
		return agentOrganizationNode{}, false
	}
	node := agentOrganizationNode{
		ID:        agentState.ID,
		Name:      agentState.Name,
		Label:     agentState.Label,
		Group:     agentState.Group,
		Workspace: agentState.Workspace,
		Status:    agentState.Status,
		Activity:  agentState.Activity,
	}
	if cfg.Agents.Organization != nil {
		for _, child := range cfg.Agents.Organization.ChildrenOf(agentID) {
			childNode, ok := buildAgentOrganizationNode(cfg, agents, child.AgentID)
			if ok {
				node.Children = append(node.Children, childNode)
			}
		}
	}
	return node, true
}

func dereferenceOrganizationAgents(
	agents map[string]*agentOrganizationAgent,
) map[string]agentOrganizationAgent {
	out := make(map[string]agentOrganizationAgent, len(agents))
	for agentID, agentState := range agents {
		if agentState != nil {
			out[agentID] = *agentState
		}
	}
	return out
}

func (s *agentOrganizationBuildState) applyActivity() {
	for _, rec := range s.delegations {
		if s.configuredDelegationAgentID(rec) == "" {
			continue
		}
		s.summary.DelegationCount++
		if isActiveDelegationStatus(rec.Status) {
			s.summary.ActiveCount++
		}
		if rec.Status == agent.AgentDelegationStatusFailed {
			s.summary.FailureCount++
		}
		s.applyDelegationRecord(rec)
	}
	for _, rec := range s.meetings {
		if s.configuredMeetingAgentID(rec) == "" {
			continue
		}
		s.summary.MeetingCount++
		if rec.Status == agent.AgentMeetingStatusStarted {
			s.summary.ActiveCount++
		}
		if rec.Status == agent.AgentMeetingStatusFailed {
			s.summary.FailureCount++
		}
		s.applyMeetingRecord(rec)
	}
}

func (s *agentOrganizationBuildState) applyRecentEvents(events map[string][]agentOrganizationRecentEvent) {
	for agentID, agentEvents := range events {
		agentState := s.agents[agentID]
		if agentState == nil {
			continue
		}
		agentState.Activity.RecentEvents = append([]agentOrganizationRecentEvent(nil), agentEvents...)
	}
}

func (s *agentOrganizationBuildState) applyDelegationRecord(rec agent.AgentDelegationRecord) {
	targetID := strings.TrimSpace(rec.TargetAgentID)
	requesterID := delegationRequesterID(rec)
	activity := organizationRecordActivity(
		"delegation",
		rec.DelegationID,
		string(rec.Status),
		rec.CreatedAt,
		rec.UpdatedAt,
		rec.CompletedAt,
		rec.Request.ArtifactRefs,
	)

	if target := s.agents[targetID]; target != nil {
		target.Activity.InboxCount++
		target.Activity.LastUpdatedAt = laterTime(target.Activity.LastUpdatedAt, rec.UpdatedAt)
		roleActivity := activity
		roleActivity.Role = "target"
		roleActivity.AgentID = requesterID
		applyAgentOrganizationStatus(target, roleActivity, delegationRecordPriority(rec.Status, "target"))
	}
	if requester := s.agents[requesterID]; requester != nil {
		requester.Activity.OutboxCount++
		requester.Activity.LastUpdatedAt = laterTime(requester.Activity.LastUpdatedAt, rec.UpdatedAt)
		roleActivity := activity
		roleActivity.Role = "requester"
		roleActivity.AgentID = targetID
		applyAgentOrganizationStatus(requester, roleActivity, delegationRecordPriority(rec.Status, "requester"))
	}
}

const agentOrganizationActivityFeedLimit = 20

func (s *agentOrganizationBuildState) recentActivityFeed(
	recentEvents map[string][]agentOrganizationRecentEvent,
) []agentOrganizationActivityFeed {
	feed := make([]agentOrganizationActivityFeed, 0, len(s.delegations)+len(s.meetings))
	for _, rec := range s.delegations {
		agentID := s.configuredDelegationAgentID(rec)
		if agentID == "" {
			continue
		}
		entryType := "delegation"
		if rec.Status == agent.AgentDelegationStatusFailed {
			entryType = "failure"
		}
		feed = append(feed, agentOrganizationActivityFeed{
			Type:      entryType,
			AgentID:   agentID,
			RecordID:  rec.DelegationID,
			Status:    string(rec.Status),
			Summary:   delegationFeedSummary(rec),
			Timestamp: timePointer(rec.UpdatedAt),
		})
	}
	for _, rec := range s.meetings {
		agentID := s.configuredMeetingAgentID(rec)
		if agentID == "" {
			continue
		}
		entryType := "meeting"
		if rec.Status == agent.AgentMeetingStatusFailed {
			entryType = "failure"
		}
		feed = append(feed, agentOrganizationActivityFeed{
			Type:      entryType,
			AgentID:   agentID,
			RecordID:  rec.MeetingID,
			Status:    string(rec.Status),
			Summary:   meetingFeedSummary(rec),
			Timestamp: timePointer(rec.UpdatedAt),
		})
	}
	for agentID, events := range recentEvents {
		for _, event := range events {
			feed = append(feed, agentOrganizationActivityFeed{
				Type:      "event",
				AgentID:   agentID,
				Status:    firstNonEmpty(event.Level, event.Event, "log"),
				Summary:   event.Message,
				Timestamp: event.Timestamp,
			})
		}
	}
	slices.SortFunc(feed, func(a, b agentOrganizationActivityFeed) int {
		if byTime := cmp.Compare(feedEntryUnixNano(b), feedEntryUnixNano(a)); byTime != 0 {
			return byTime
		}
		if byType := cmp.Compare(a.Type, b.Type); byType != 0 {
			return byType
		}
		if byAgent := cmp.Compare(a.AgentID, b.AgentID); byAgent != 0 {
			return byAgent
		}
		return cmp.Compare(a.RecordID, b.RecordID)
	})
	if len(feed) > agentOrganizationActivityFeedLimit {
		feed = feed[:agentOrganizationActivityFeedLimit]
	}
	return feed
}

func (s *agentOrganizationBuildState) configuredDelegationAgentID(rec agent.AgentDelegationRecord) string {
	targetID := strings.TrimSpace(rec.TargetAgentID)
	if _, ok := s.agents[targetID]; ok {
		return targetID
	}
	requesterID := delegationRequesterID(rec)
	if _, ok := s.agents[requesterID]; ok {
		return requesterID
	}
	return ""
}

func (s *agentOrganizationBuildState) configuredMeetingAgentID(rec agent.AgentMeetingRecord) string {
	if chairID := strings.TrimSpace(rec.ChairAgentID); chairID != "" {
		if _, ok := s.agents[chairID]; ok {
			return chairID
		}
	}
	if sponsorID := strings.TrimSpace(rec.SponsorAgentID); sponsorID != "" {
		if _, ok := s.agents[sponsorID]; ok {
			return sponsorID
		}
	}
	for _, participantID := range rec.Participants {
		participantID = strings.TrimSpace(participantID)
		if _, ok := s.agents[participantID]; ok {
			return participantID
		}
	}
	return ""
}

func delegationFeedSummary(rec agent.AgentDelegationRecord) string {
	return fmt.Sprintf("Delegation %s: %s -> %s", rec.Status, delegationRequesterID(rec), rec.TargetAgentID)
}

func meetingFeedSummary(rec agent.AgentMeetingRecord) string {
	return fmt.Sprintf("Meeting %s: %s", rec.Status, firstNonEmpty(rec.Title, rec.MeetingID))
}

func feedEntryUnixNano(entry agentOrganizationActivityFeed) int64 {
	if entry.Timestamp == nil {
		return 0
	}
	return entry.Timestamp.UnixNano()
}

func timePointer(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	value = value.UTC()
	return &value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func (s *agentOrganizationBuildState) applyMeetingRecord(rec agent.AgentMeetingRecord) {
	participantIDs := meetingParticipantIDs(rec)
	activity := organizationRecordActivity(
		"meeting",
		rec.MeetingID,
		string(rec.Status),
		rec.CreatedAt,
		rec.UpdatedAt,
		rec.CompletedAt,
		rec.ArtifactRefs,
	)
	for agentID, role := range participantIDs {
		agentState := s.agents[agentID]
		if agentState == nil {
			continue
		}
		agentState.Activity.MeetingCount++
		agentState.Activity.LastUpdatedAt = laterTime(agentState.Activity.LastUpdatedAt, rec.UpdatedAt)
		roleActivity := activity
		roleActivity.Role = role
		applyAgentOrganizationStatus(agentState, roleActivity, meetingRecordPriority(rec.Status))
	}
}

func delegationRequesterID(rec agent.AgentDelegationRecord) string {
	if requestedBy := strings.TrimSpace(rec.Request.RequestedBy); requestedBy != "" {
		return requestedBy
	}
	return strings.TrimSpace(rec.ParentAgentID)
}

func meetingParticipantIDs(rec agent.AgentMeetingRecord) map[string]string {
	ids := map[string]string{}
	if sponsor := strings.TrimSpace(rec.SponsorAgentID); sponsor != "" {
		ids[sponsor] = "sponsor"
	}
	if chair := strings.TrimSpace(rec.ChairAgentID); chair != "" {
		ids[chair] = "chair"
	}
	for _, participant := range rec.Participants {
		participant = strings.TrimSpace(participant)
		if participant != "" {
			if _, ok := ids[participant]; !ok {
				ids[participant] = "participant"
			}
		}
	}
	return ids
}

func organizationRecordActivity(
	recordType string,
	recordID string,
	status string,
	createdAt time.Time,
	updatedAt time.Time,
	completedAt *time.Time,
	artifactRefs []string,
) agentOrganizationActivityRecord {
	activity := agentOrganizationActivityRecord{
		Type:         recordType,
		RecordID:     recordID,
		Status:       status,
		ArtifactRefs: append([]string(nil), artifactRefs...),
	}
	if !createdAt.IsZero() {
		created := createdAt.UTC()
		activity.CreatedAt = &created
	}
	if !updatedAt.IsZero() {
		updated := updatedAt.UTC()
		activity.UpdatedAt = &updated
	}
	if completedAt != nil && !completedAt.IsZero() {
		completed := completedAt.UTC()
		activity.CompletedAt = &completed
	}
	return activity
}

func applyAgentOrganizationStatus(
	agentState *agentOrganizationAgent,
	record agentOrganizationActivityRecord,
	priority int,
) {
	if agentState == nil || priority <= 0 {
		return
	}
	if priority == organizationStatusPriorityFailed {
		agentState.Activity.FailureCount++
		agentState.Activity.LastFailure = newerActivityRecord(agentState.Activity.LastFailure, record)
	}
	if !shouldReplaceCurrentActivity(agentState.Activity.Current, record, priority) {
		return
	}
	agentState.Activity.Current = &record
	agentState.Status = statusForOrganizationPriority(priority)
}

func shouldReplaceCurrentActivity(
	current *agentOrganizationActivityRecord,
	candidate agentOrganizationActivityRecord,
	priority int,
) bool {
	if current == nil {
		return true
	}
	currentPriority := organizationPriorityForStatus(current)
	if byTime := cmp.Compare(activityRecordUnixNano(candidate), activityRecordUnixNano(*current)); byTime != 0 {
		return byTime > 0
	}
	if priority != currentPriority {
		return priority > currentPriority
	}
	return compareActivityRecord(candidate, *current) > 0
}

func newerActivityRecord(
	current *agentOrganizationActivityRecord,
	candidate agentOrganizationActivityRecord,
) *agentOrganizationActivityRecord {
	if current == nil || compareActivityRecord(candidate, *current) > 0 {
		return &candidate
	}
	return current
}

func compareActivityRecord(a, b agentOrganizationActivityRecord) int {
	aTime := activityRecordUnixNano(a)
	bTime := activityRecordUnixNano(b)
	if byTime := cmp.Compare(aTime, bTime); byTime != 0 {
		return byTime
	}
	if byType := cmp.Compare(a.Type, b.Type); byType != 0 {
		return byType
	}
	return cmp.Compare(a.RecordID, b.RecordID)
}

func activityRecordUnixNano(record agentOrganizationActivityRecord) int64 {
	if record.UpdatedAt == nil {
		return 0
	}
	return record.UpdatedAt.UnixNano()
}

func activityRecordCreatedUnixNano(record agentOrganizationActivityRecord) int64 {
	if record.CreatedAt == nil {
		return 0
	}
	return record.CreatedAt.UnixNano()
}

const (
	organizationStatusPriorityIdle       = 0
	organizationStatusPriorityCompleted  = 1
	organizationStatusPriorityDelegating = 2
	organizationStatusPriorityWorking    = 3
	organizationStatusPriorityMeeting    = 4
	organizationStatusPriorityFailed     = 5
)

func organizationPriorityForStatus(record *agentOrganizationActivityRecord) int {
	if record == nil {
		return organizationStatusPriorityIdle
	}
	if record.Status == "failed" {
		return organizationStatusPriorityFailed
	}
	if record.Status == "completed" {
		return organizationStatusPriorityCompleted
	}
	if record.Type == "meeting" && record.Status == string(agent.AgentMeetingStatusStarted) {
		return organizationStatusPriorityMeeting
	}
	if record.Type == "delegation" &&
		(record.Status == string(agent.AgentDelegationStatusRequested) ||
			record.Status == string(agent.AgentDelegationStatusRunning)) {
		if record.Role == "target" {
			return organizationStatusPriorityWorking
		}
		return organizationStatusPriorityDelegating
	}
	return organizationStatusPriorityIdle
}

func delegationRecordPriority(status agent.AgentDelegationStatus, role string) int {
	switch status {
	case agent.AgentDelegationStatusFailed:
		return organizationStatusPriorityFailed
	case agent.AgentDelegationStatusRequested, agent.AgentDelegationStatusRunning:
		if role == "target" {
			return organizationStatusPriorityWorking
		}
		return organizationStatusPriorityDelegating
	case agent.AgentDelegationStatusCompleted:
		return organizationStatusPriorityCompleted
	default:
		return organizationStatusPriorityIdle
	}
}

func meetingRecordPriority(status agent.AgentMeetingStatus) int {
	switch status {
	case agent.AgentMeetingStatusFailed:
		return organizationStatusPriorityFailed
	case agent.AgentMeetingStatusStarted:
		return organizationStatusPriorityMeeting
	case agent.AgentMeetingStatusCompleted:
		return organizationStatusPriorityCompleted
	default:
		return organizationStatusPriorityIdle
	}
}

func statusForOrganizationPriority(priority int) string {
	switch priority {
	case organizationStatusPriorityFailed:
		return agentOrganizationStatusFailed
	case organizationStatusPriorityMeeting:
		return agentOrganizationStatusMeeting
	case organizationStatusPriorityWorking:
		return agentOrganizationStatusWorking
	case organizationStatusPriorityDelegating:
		return agentOrganizationStatusDelegating
	default:
		return agentOrganizationStatusIdle
	}
}

func isActiveDelegationStatus(status agent.AgentDelegationStatus) bool {
	return status == agent.AgentDelegationStatusRequested || status == agent.AgentDelegationStatusRunning
}

func laterTime(current *time.Time, candidate time.Time) *time.Time {
	if candidate.IsZero() {
		return current
	}
	candidate = candidate.UTC()
	if current == nil || candidate.After(*current) {
		return &candidate
	}
	return current
}

const (
	agentOrganizationRecentEventLimit       = 10
	agentOrganizationRecentEventMaxMessage  = 160
	agentOrganizationRecentEventSource      = "gateway_log"
	agentOrganizationRecentEventDefaultText = "gateway event"
)

var (
	bearerAssignmentLogValuePattern = regexp.MustCompile(`(?i)authorization=Bearer\s+\S+`)
	sensitiveAssignmentLogPattern   = regexp.MustCompile(`(?i)(authorization|token|api[_-]?key|secret|password)=\S+`)
	bearerLogValuePattern           = regexp.MustCompile(`(?i)Bearer\s+\S+`)
	openAIKeyLogValuePattern        = regexp.MustCompile(`sk-[A-Za-z0-9_-]+`)
)

func gatewayLogRecentEvents(
	agents map[string]*agentOrganizationAgent,
) map[string][]agentOrganizationRecentEvent {
	if len(agents) == 0 || gateway.logs == nil {
		return nil
	}
	lines, _, _ := gateway.logs.LinesSince(0)
	if len(lines) == 0 {
		return nil
	}

	events := make(map[string][]agentOrganizationRecentEvent)
	for i := len(lines) - 1; i >= 0; i-- {
		event, ok := parseGatewayLogRecentEvent(lines[i])
		if !ok {
			continue
		}
		for _, agentID := range matchingLogAgentIDs(event.fields, agents) {
			if len(events[agentID]) >= agentOrganizationRecentEventLimit {
				continue
			}
			agentEvent := event.toRecentEvent(agentID)
			events[agentID] = append(events[agentID], agentEvent)
		}
	}
	return events
}

type parsedGatewayLogEvent struct {
	fields    map[string]any
	level     string
	event     string
	message   string
	timestamp *time.Time
}

func parseGatewayLogRecentEvent(line string) (parsedGatewayLogEvent, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return parsedGatewayLogEvent{}, false
	}

	var fields map[string]any
	if err := json.Unmarshal([]byte(line), &fields); err != nil || len(fields) == 0 {
		return parseTextGatewayLogRecentEvent(line)
	}

	message := firstStringLogField(fields, "message", "msg")
	if message == "" {
		message = firstStringLogField(fields, "event", "type")
	}
	if message == "" {
		message = agentOrganizationRecentEventDefaultText
	}

	timestamp := parseGatewayLogTimestamp(firstStringLogField(fields, "time", "timestamp", "ts"))
	return parsedGatewayLogEvent{
		fields:    fields,
		level:     firstStringLogField(fields, "level", "severity"),
		event:     firstStringLogField(fields, "event"),
		message:   sanitizeRecentEventMessage(message),
		timestamp: timestamp,
	}, true
}

func parseTextGatewayLogRecentEvent(line string) (parsedGatewayLogEvent, bool) {
	fields, ok := parseGatewayLogKeyValues(line)
	if !ok || len(matchingKnownLogAgentIDs(fields)) == 0 {
		return parsedGatewayLogEvent{}, false
	}

	return parsedGatewayLogEvent{
		fields:    fields,
		level:     textGatewayLogLevel(fields, line),
		event:     firstStringLogField(fields, "event", "type"),
		message:   sanitizeRecentEventMessage(line),
		timestamp: parseGatewayLogTimestamp(firstStringLogField(fields, "time", "timestamp", "ts")),
	}, true
}

func parseGatewayLogKeyValues(line string) (map[string]any, bool) {
	fields := map[string]any{}
	for i := 0; i < len(line); {
		for i < len(line) && line[i] <= ' ' {
			i++
		}
		start := i
		for i < len(line) && isGatewayLogKeyByte(line[i]) {
			i++
		}
		if i == start || i >= len(line) || line[i] != '=' {
			for i < len(line) && line[i] > ' ' {
				i++
			}
			continue
		}

		key := line[start:i]
		i++
		if i >= len(line) {
			fields[key] = ""
			break
		}

		var value string
		if line[i] == '"' {
			parsed, next, ok := parseQuotedGatewayLogValue(line, i+1)
			if !ok {
				return nil, false
			}
			value = parsed
			i = next
		} else {
			valueStart := i
			for i < len(line) && line[i] > ' ' {
				i++
			}
			value = line[valueStart:i]
		}
		fields[key] = value
	}
	return fields, len(fields) > 0
}

func parseQuotedGatewayLogValue(line string, start int) (string, int, bool) {
	var value strings.Builder
	escaped := false
	for i := start; i < len(line); i++ {
		ch := line[i]
		if escaped {
			value.WriteByte(ch)
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == '"' {
			next := i + 1
			if next < len(line) && line[next] > ' ' {
				return "", 0, false
			}
			return value.String(), next, true
		}
		value.WriteByte(ch)
	}
	return "", 0, false
}

func isGatewayLogKeyByte(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '_' || ch == '-'
}

func textGatewayLogLevel(fields map[string]any, line string) string {
	if level := firstStringLogField(fields, "level", "severity"); level != "" {
		return normalizeGatewayLogLevel(level)
	}
	for _, token := range strings.Fields(line) {
		switch strings.Trim(token, "[]") {
		case "DBG", "DEBUG":
			return "debug"
		case "INF", "INFO":
			return "info"
		case "WRN", "WARN", "WARNING":
			return "warn"
		case "ERR", "ERROR":
			return "error"
		case "FTL", "FATAL":
			return "fatal"
		}
	}
	return ""
}

func normalizeGatewayLogLevel(level string) string {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "DBG", "DEBUG":
		return "debug"
	case "INF", "INFO":
		return "info"
	case "WRN", "WARN", "WARNING":
		return "warn"
	case "ERR", "ERROR":
		return "error"
	case "FTL", "FATAL":
		return "fatal"
	default:
		return strings.TrimSpace(level)
	}
}

func (e parsedGatewayLogEvent) toRecentEvent(agentID string) agentOrganizationRecentEvent {
	return agentOrganizationRecentEvent{
		Source:    agentOrganizationRecentEventSource,
		AgentID:   agentID,
		Level:     e.level,
		Event:     e.event,
		Message:   e.message,
		Timestamp: e.timestamp,
	}
}

func matchingLogAgentIDs(
	fields map[string]any,
	agents map[string]*agentOrganizationAgent,
) []string {
	matches := make([]string, 0, 1)
	seen := map[string]struct{}{}
	for _, field := range gatewayLogAgentReferenceFields {
		for _, agentID := range stringValuesFromLogField(fields[field]) {
			agentID = strings.TrimSpace(agentID)
			if agentID == "" || agents[agentID] == nil {
				continue
			}
			if _, ok := seen[agentID]; ok {
				continue
			}
			seen[agentID] = struct{}{}
			matches = append(matches, agentID)
		}
	}
	return matches
}

func matchingKnownLogAgentIDs(fields map[string]any) []string {
	matches := make([]string, 0, 1)
	for _, field := range gatewayLogAgentReferenceFields {
		matches = append(matches, stringValuesFromLogField(fields[field])...)
	}
	return matches
}

var gatewayLogAgentReferenceFields = []string{
	"agent_id",
	"target_agent_id",
	"parent_agent_id",
	"requester_id",
	"sponsor_agent_id",
	"chair_agent_id",
	"child_agent_id",
	"route_agent_id",
	"scope_agent_id",
}

func firstStringLogField(fields map[string]any, keys ...string) string {
	for _, key := range keys {
		values := stringValuesFromLogField(fields[key])
		if len(values) > 0 {
			return strings.TrimSpace(values[0])
		}
	}
	return ""
}

func stringValuesFromLogField(value any) []string {
	switch typed := value.(type) {
	case string:
		if strings.TrimSpace(typed) == "" {
			return nil
		}
		return []string{typed}
	case []any:
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				values = append(values, s)
			}
		}
		return values
	default:
		return nil
	}
}

func parseGatewayLogTimestamp(raw string) *time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return nil
	}
	parsed = parsed.UTC()
	return &parsed
}

func sanitizeRecentEventMessage(message string) string {
	message = strings.TrimSpace(message)
	message = bearerAssignmentLogValuePattern.ReplaceAllString(message, "authorization=[redacted]")
	message = sensitiveAssignmentLogPattern.ReplaceAllString(message, "$1=[redacted]")
	message = bearerLogValuePattern.ReplaceAllString(message, "Bearer [redacted]")
	message = openAIKeyLogValuePattern.ReplaceAllString(message, "[redacted]")
	if len(message) <= agentOrganizationRecentEventMaxMessage {
		return message
	}
	return strings.TrimSpace(message[:agentOrganizationRecentEventMaxMessage-3]) + "..."
}
