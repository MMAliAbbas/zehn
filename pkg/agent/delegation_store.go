package agent

import (
	"cmp"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/sipeed/picoclaw/pkg/fileutil"
	"github.com/sipeed/picoclaw/pkg/session"
)

type AgentDelegationStatus string

const (
	AgentDelegationStatusRequested AgentDelegationStatus = "requested"
	AgentDelegationStatusRunning   AgentDelegationStatus = "running"
	AgentDelegationStatusCompleted AgentDelegationStatus = "completed"
	AgentDelegationStatusFailed    AgentDelegationStatus = "failed"
	AgentDelegationStatusCancelled AgentDelegationStatus = "cancelled"
)

type AgentDelegationRecord struct {
	DelegationID   string                       `json:"delegation_id"`
	Status         AgentDelegationStatus        `json:"status"`
	ParentAgentID  string                       `json:"parent_agent_id"`
	TargetAgentID  string                       `json:"target_agent_id"`
	Request        AgentDelegationRecordRequest `json:"request"`
	CreatedAt      time.Time                    `json:"created_at"`
	UpdatedAt      time.Time                    `json:"updated_at"`
	StartedAt      *time.Time                   `json:"started_at,omitempty"`
	CompletedAt    *time.Time                   `json:"completed_at,omitempty"`
	Result         *AgentDelegationRecordResult `json:"result,omitempty"`
	Error          *AgentDelegationRecordError  `json:"error,omitempty"`
	DurableMemory  *AgentDelegationMemoryWrite  `json:"durable_memory,omitempty"`
	GitHubArtifact *AgentGitHubArtifactWrite    `json:"github_artifact,omitempty"`
}

type AgentDelegationRecordQuery struct {
	DelegationID      string
	VisibleToAgentID  string
	ParentAgentID     string
	TargetAgentID     string
	IncludePrivateAll bool
}

type AgentDelegationRecordRequest struct {
	Task             string     `json:"task"`
	ThreadKey        string     `json:"thread_key,omitempty"`
	Mode             string     `json:"mode,omitempty"`
	Priority         string     `json:"priority,omitempty"`
	DueAt            *time.Time `json:"due_at,omitempty"`
	RequestedBy      string     `json:"requested_by,omitempty"`
	ApprovalRequired bool       `json:"approval_required,omitempty"`
	ArtifactRefs     []string   `json:"artifact_refs,omitempty"`
}

type AgentDelegationRecordResult struct {
	Content      string                `json:"content,omitempty"`
	Status       TurnEndStatus         `json:"status"`
	SessionKey   string                `json:"session_key,omitempty"`
	SessionScope *session.SessionScope `json:"session_scope,omitempty"`
	ArtifactRefs []string              `json:"artifact_refs,omitempty"`
}

type AgentDelegationRecordError struct {
	Message string `json:"message"`
	Type    string `json:"type,omitempty"`
}

type AgentDelegationMemoryStatus string

const (
	AgentDelegationMemoryStatusUnavailable AgentDelegationMemoryStatus = "unavailable"
	AgentDelegationMemoryStatusWritten     AgentDelegationMemoryStatus = "written"
	AgentDelegationMemoryStatusFailed      AgentDelegationMemoryStatus = "failed"
	AgentDelegationMemoryStatusSkipped     AgentDelegationMemoryStatus = "skipped"
)

type AgentDelegationMemoryWrite struct {
	Provider        string                      `json:"provider"`
	Status          AgentDelegationMemoryStatus `json:"status"`
	MemoryID        string                      `json:"memory_id,omitempty"`
	Error           string                      `json:"error,omitempty"`
	SkippedStatuses []AgentDelegationStatus     `json:"skipped_statuses,omitempty"`
	UpdatedAt       time.Time                   `json:"updated_at"`
}

func (r AgentDelegationRecord) Filename() string {
	return r.DelegationID + ".json"
}

type DelegationRecordStore struct {
	dir    string
	filter func(string) string
	now    func() time.Time
	mu     sync.Mutex
}

func NewDelegationRecordStore(dir string, filter func(string) string) *DelegationRecordStore {
	return &DelegationRecordStore{
		dir:    dir,
		filter: filter,
		now:    time.Now,
	}
}

func (s *DelegationRecordStore) Requested(ctx context.Context, req AgentDelegationRequest) (AgentDelegationRecord, error) {
	if s == nil {
		return AgentDelegationRecord{}, nil
	}
	if err := ctx.Err(); err != nil {
		return AgentDelegationRecord{}, err
	}
	now := s.now().UTC()
	rec := AgentDelegationRecord{
		DelegationID:  s.newID(req, now),
		Status:        AgentDelegationStatusRequested,
		ParentAgentID: s.redact(req.ParentAgentID),
		TargetAgentID: s.redact(req.TargetAgentID),
		Request: AgentDelegationRecordRequest{
			Task:             s.redact(req.Task),
			ThreadKey:        s.redact(req.ThreadKey),
			Mode:             s.redact(req.Mode),
			Priority:         s.redact(req.Priority),
			DueAt:            req.DueAt,
			RequestedBy:      s.redact(req.RequestedBy),
			ApprovalRequired: req.ApprovalRequired,
			ArtifactRefs:     s.redactRefs(req.ArtifactRefs),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	return rec, s.write(ctx, rec)
}

func (s *DelegationRecordStore) Running(ctx context.Context, delegationID string) error {
	return s.update(ctx, delegationID, func(rec *AgentDelegationRecord, now time.Time) {
		rec.Status = AgentDelegationStatusRunning
		rec.StartedAt = &now
		rec.Error = nil
	})
}

func (s *DelegationRecordStore) Completed(ctx context.Context, delegationID string, result AgentDelegationResult) error {
	return s.update(ctx, delegationID, func(rec *AgentDelegationRecord, now time.Time) {
		rec.Status = AgentDelegationStatusCompleted
		rec.CompletedAt = &now
		rec.Result = &AgentDelegationRecordResult{
			Content:      s.redact(result.Content),
			Status:       result.Status,
			SessionKey:   s.redact(result.SessionKey),
			SessionScope: s.redactSessionScope(result.SessionScope),
			ArtifactRefs: s.redactRefs(result.ArtifactRefs),
		}
		rec.Error = nil
	})
}

func (s *DelegationRecordStore) Failed(ctx context.Context, delegationID string, err error) error {
	if err == nil {
		err = errors.New("delegation failed")
	}
	status := AgentDelegationStatusFailed
	if errors.Is(err, context.Canceled) {
		status = AgentDelegationStatusCancelled
	}
	return s.update(ctx, delegationID, func(rec *AgentDelegationRecord, now time.Time) {
		rec.Status = status
		rec.CompletedAt = &now
		rec.Error = &AgentDelegationRecordError{
			Message: s.redact(err.Error()),
			Type:    fmt.Sprintf("%T", err),
		}
	})
}

func (s *DelegationRecordStore) RecordMemoryWrite(
	ctx context.Context,
	delegationID string,
	write AgentDelegationMemoryWrite,
) error {
	return s.update(ctx, delegationID, func(rec *AgentDelegationRecord, now time.Time) {
		if rec.DurableMemory != nil {
			write.SkippedStatuses = appendUniqueDelegationStatuses(
				rec.DurableMemory.SkippedStatuses,
				write.SkippedStatuses...,
			)
		}
		write.Provider = s.redact(write.Provider)
		write.MemoryID = s.redact(write.MemoryID)
		write.Error = s.redact(write.Error)
		write.UpdatedAt = now
		rec.DurableMemory = &write
	})
}

func (s *DelegationRecordStore) RecordMemorySkipped(
	ctx context.Context,
	delegationID string,
	status AgentDelegationStatus,
	reason string,
) error {
	return s.update(ctx, delegationID, func(rec *AgentDelegationRecord, now time.Time) {
		write := AgentDelegationMemoryWrite{
			Provider: "yaad",
			Status:   AgentDelegationMemoryStatusSkipped,
			Error:    reason,
		}
		if rec.DurableMemory != nil {
			write = *rec.DurableMemory
			if write.Provider == "" {
				write.Provider = "yaad"
			}
			if write.Status != AgentDelegationMemoryStatusWritten {
				write.Status = AgentDelegationMemoryStatusSkipped
				write.Error = reason
			}
		}
		write.Provider = s.redact(write.Provider)
		write.MemoryID = s.redact(write.MemoryID)
		write.Error = s.redact(write.Error)
		write.SkippedStatuses = appendUniqueDelegationStatuses(write.SkippedStatuses, status)
		write.UpdatedAt = now
		rec.DurableMemory = &write
	})
}

func (s *DelegationRecordStore) RecordGitHubArtifact(
	ctx context.Context,
	delegationID string,
	write AgentGitHubArtifactWrite,
	artifactRefs []string,
) error {
	return s.update(ctx, delegationID, func(rec *AgentDelegationRecord, now time.Time) {
		write.UpdatedAt = now
		rec.GitHubArtifact = &write
		rec.Request.ArtifactRefs = appendUniqueRefs(rec.Request.ArtifactRefs, artifactRefs...)
		if rec.Result != nil {
			rec.Result.ArtifactRefs = appendUniqueRefs(rec.Result.ArtifactRefs, artifactRefs...)
		}
	})
}

func (s *DelegationRecordStore) Get(ctx context.Context, delegationID string) (AgentDelegationRecord, error) {
	if s == nil {
		return AgentDelegationRecord{}, os.ErrNotExist
	}
	if err := ctx.Err(); err != nil {
		return AgentDelegationRecord{}, err
	}
	path, err := s.pathForID(delegationID)
	if err != nil {
		return AgentDelegationRecord{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return AgentDelegationRecord{}, err
	}
	var rec AgentDelegationRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return AgentDelegationRecord{}, err
	}
	return rec, nil
}

func (s *DelegationRecordStore) List(ctx context.Context, query AgentDelegationRecordQuery) ([]AgentDelegationRecord, error) {
	if s == nil {
		return nil, nil
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(query.DelegationID) != "" {
		rec, err := s.Get(ctx, query.DelegationID)
		if err != nil {
			return nil, err
		}
		if !recordMatchesDelegationQuery(rec, query) {
			return nil, nil
		}
		return []AgentDelegationRecord{rec}, nil
	}

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	records := make([]AgentDelegationRecord, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".json")
		rec, err := s.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		if recordMatchesDelegationQuery(rec, query) {
			records = append(records, rec)
		}
	}
	slices.SortFunc(records, func(a, b AgentDelegationRecord) int {
		if cmpCreated := cmp.Compare(a.CreatedAt.UnixNano(), b.CreatedAt.UnixNano()); cmpCreated != 0 {
			return cmpCreated
		}
		return cmp.Compare(a.DelegationID, b.DelegationID)
	})
	return records, nil
}

func (s *DelegationRecordStore) update(ctx context.Context, delegationID string, mutate func(*AgentDelegationRecord, time.Time)) error {
	if s == nil {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	rec, err := s.getUnlocked(ctx, delegationID)
	if err != nil {
		return err
	}
	now := s.now().UTC()
	mutate(&rec, now)
	rec.UpdatedAt = now
	return s.writeUnlocked(rec)
}

func recordMatchesDelegationQuery(rec AgentDelegationRecord, query AgentDelegationRecordQuery) bool {
	if query.DelegationID != "" && rec.DelegationID != strings.TrimSpace(query.DelegationID) {
		return false
	}
	if query.ParentAgentID != "" && rec.ParentAgentID != strings.TrimSpace(query.ParentAgentID) {
		return false
	}
	if query.TargetAgentID != "" && rec.TargetAgentID != strings.TrimSpace(query.TargetAgentID) {
		return false
	}
	if !query.IncludePrivateAll {
		visibleTo := strings.TrimSpace(query.VisibleToAgentID)
		if visibleTo != "" && rec.ParentAgentID != visibleTo && rec.TargetAgentID != visibleTo {
			return false
		}
	}
	return true
}

func (s *DelegationRecordStore) write(ctx context.Context, rec AgentDelegationRecord) error {
	if s == nil {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.writeUnlocked(rec)
}

func (s *DelegationRecordStore) getUnlocked(ctx context.Context, delegationID string) (AgentDelegationRecord, error) {
	if err := ctx.Err(); err != nil {
		return AgentDelegationRecord{}, err
	}
	path, err := s.pathForID(delegationID)
	if err != nil {
		return AgentDelegationRecord{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return AgentDelegationRecord{}, err
	}
	var rec AgentDelegationRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return AgentDelegationRecord{}, err
	}
	return rec, nil
}

func (s *DelegationRecordStore) writeUnlocked(rec AgentDelegationRecord) error {
	path, err := s.pathForID(rec.DelegationID)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	return fileutil.WriteFileAtomic(path, data, 0o600)
}

func (s *DelegationRecordStore) pathForID(delegationID string) (string, error) {
	delegationID = strings.TrimSpace(delegationID)
	if delegationID == "" || !delegationIDPattern.MatchString(delegationID) {
		return "", fmt.Errorf("%w: invalid delegation ID %q", ErrAgentDelegationInvalidRequest, delegationID)
	}
	filename := delegationID + ".json"
	if !filepath.IsLocal(filename) {
		return "", fmt.Errorf("%w: invalid delegation filename %q", ErrAgentDelegationInvalidRequest, filename)
	}
	return filepath.Join(s.dir, filename), nil
}

func (s *DelegationRecordStore) newID(req AgentDelegationRequest, now time.Time) string {
	input := strings.Join([]string{
		req.ParentAgentID,
		req.TargetAgentID,
		req.ThreadKey,
		req.Task,
		now.Format(time.RFC3339Nano),
	}, "\x00")
	sum := sha256.Sum256([]byte(input))
	return fmt.Sprintf("delegation-%s-%s", now.Format("20060102T150405.000000000Z"), hex.EncodeToString(sum[:])[:12])
}

func (s *DelegationRecordStore) redact(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || s.filter == nil {
		return value
	}
	return s.filter(value)
}

func (s *DelegationRecordStore) redactRefs(refs []string) []string {
	out := make([]string, 0, len(refs))
	for _, ref := range refs {
		ref = s.redact(ref)
		if ref != "" {
			out = append(out, ref)
		}
	}
	return out
}

func (s *DelegationRecordStore) redactSessionScope(scope *session.SessionScope) *session.SessionScope {
	if scope == nil {
		return nil
	}
	redacted := session.CloneScope(scope)
	redacted.AgentID = s.redact(redacted.AgentID)
	redacted.Channel = s.redact(redacted.Channel)
	redacted.Account = s.redact(redacted.Account)
	for i, dim := range redacted.Dimensions {
		redacted.Dimensions[i] = s.redact(dim)
	}
	for key, value := range redacted.Values {
		delete(redacted.Values, key)
		redacted.Values[s.redact(key)] = s.redact(value)
	}
	return redacted
}

func appendUniqueDelegationStatuses(
	values []AgentDelegationStatus,
	additions ...AgentDelegationStatus,
) []AgentDelegationStatus {
	out := append([]AgentDelegationStatus(nil), values...)
	for _, addition := range additions {
		if addition == "" {
			continue
		}
		if slices.Contains(out, addition) {
			continue
		}
		out = append(out, addition)
	}
	return out
}

var delegationIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
