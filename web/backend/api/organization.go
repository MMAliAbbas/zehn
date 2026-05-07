package api

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
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
	Current       *agentOrganizationActivityRecord `json:"current,omitempty"`
	LastFailure   *agentOrganizationActivityRecord `json:"last_failure,omitempty"`
	LastUpdatedAt *time.Time                       `json:"last_updated_at,omitempty"`
}

type agentOrganizationActivityRecord struct {
	Type      string     `json:"type"`
	RecordID  string     `json:"record_id"`
	Status    string     `json:"status"`
	Role      string     `json:"role,omitempty"`
	AgentID   string     `json:"agent_id,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type agentOrganizationActivitySummary struct {
	DelegationCount int `json:"delegation_count"`
	MeetingCount    int `json:"meeting_count"`
	FailureCount    int `json:"failure_count"`
	ActiveCount     int `json:"active_count"`
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

func (s *agentOrganizationBuildState) applyDelegationRecord(rec agent.AgentDelegationRecord) {
	targetID := strings.TrimSpace(rec.TargetAgentID)
	requesterID := delegationRequesterID(rec)
	activity := organizationRecordActivity("delegation", rec.DelegationID, string(rec.Status), rec.UpdatedAt)

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

func (s *agentOrganizationBuildState) applyMeetingRecord(rec agent.AgentMeetingRecord) {
	participantIDs := meetingParticipantIDs(rec)
	activity := organizationRecordActivity("meeting", rec.MeetingID, string(rec.Status), rec.UpdatedAt)
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
	updatedAt time.Time,
) agentOrganizationActivityRecord {
	activity := agentOrganizationActivityRecord{
		Type:     recordType,
		RecordID: recordID,
		Status:   status,
	}
	if !updatedAt.IsZero() {
		updated := updatedAt.UTC()
		activity.UpdatedAt = &updated
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

const (
	organizationStatusPriorityIdle       = 0
	organizationStatusPriorityDelegating = 1
	organizationStatusPriorityWorking    = 2
	organizationStatusPriorityMeeting    = 3
	organizationStatusPriorityFailed     = 4
)

func organizationPriorityForStatus(record *agentOrganizationActivityRecord) int {
	if record == nil {
		return organizationStatusPriorityIdle
	}
	if record.Status == "failed" {
		return organizationStatusPriorityFailed
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
