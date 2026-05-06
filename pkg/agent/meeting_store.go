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
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/sipeed/picoclaw/pkg/fileutil"
)

type AgentMeetingStatus string

const (
	AgentMeetingStatusStarted   AgentMeetingStatus = "started"
	AgentMeetingStatusCompleted AgentMeetingStatus = "completed"
	AgentMeetingStatusFailed    AgentMeetingStatus = "failed"
	AgentMeetingStatusCancelled AgentMeetingStatus = "cancelled"
)

type AgentMeetingRecord struct {
	MeetingID        string                    `json:"meeting_id"`
	Status           AgentMeetingStatus        `json:"status"`
	Title            string                    `json:"title"`
	SponsorAgentID   string                    `json:"sponsor_agent_id"`
	ChairAgentID     string                    `json:"chair_agent_id"`
	Participants     []string                  `json:"participants"`
	Goal             string                    `json:"goal"`
	Constraints      []string                  `json:"constraints,omitempty"`
	Notes            string                    `json:"notes,omitempty"`
	Recommendation   string                    `json:"recommendation,omitempty"`
	Timeline         []string                  `json:"timeline,omitempty"`
	Risks            []string                  `json:"risks,omitempty"`
	Approvals        []string                  `json:"approvals,omitempty"`
	FollowUps        []string                  `json:"follow_ups,omitempty"`
	ArtifactRefs     []string                  `json:"artifact_refs,omitempty"`
	ParticipantTurns []AgentMeetingTurn        `json:"participant_turns,omitempty"`
	ChairTurn        *AgentMeetingChairTurn    `json:"chair_turn,omitempty"`
	GitHubArtifact   *AgentGitHubArtifactWrite `json:"github_artifact,omitempty"`
	CreatedAt        time.Time                 `json:"created_at"`
	UpdatedAt        time.Time                 `json:"updated_at"`
	CompletedAt      *time.Time                `json:"completed_at,omitempty"`
	Error            string                    `json:"error,omitempty"`
}

type AgentMeetingTurn struct {
	AgentID      string    `json:"agent_id"`
	DelegationID string    `json:"delegation_id,omitempty"`
	Response     string    `json:"response,omitempty"`
	Status       string    `json:"status,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type AgentMeetingChairTurn struct {
	AgentID  string `json:"agent_id"`
	Response string `json:"response"`
}

func (r AgentMeetingRecord) Filename() string {
	return r.MeetingID + ".json"
}

type MeetingRecordStore struct {
	dir    string
	filter func(string) string
	now    func() time.Time
	mu     sync.Mutex
}

func NewMeetingRecordStore(dir string, filter func(string) string) *MeetingRecordStore {
	return &MeetingRecordStore{
		dir:    dir,
		filter: filter,
		now:    time.Now,
	}
}

func (s *MeetingRecordStore) Started(ctx context.Context, req AgentMeetingRequest) (AgentMeetingRecord, error) {
	if s == nil {
		return AgentMeetingRecord{}, nil
	}
	if err := ctx.Err(); err != nil {
		return AgentMeetingRecord{}, err
	}
	now := s.now().UTC()
	rec := AgentMeetingRecord{
		MeetingID:      s.newID(req, now),
		Status:         AgentMeetingStatusStarted,
		Title:          s.redact(req.Title),
		SponsorAgentID: s.redact(req.SponsorAgentID),
		ChairAgentID:   s.redact(req.ChairAgentID),
		Participants:   s.redactRefs(req.ParticipantAgentIDs),
		Goal:           s.redact(req.Goal),
		Constraints:    s.redactRefs(req.Constraints),
		Notes:          s.redact(req.Notes),
		Approvals:      s.redactRefs(req.Approvals),
		ArtifactRefs:   s.redactRefs(req.ArtifactRefs),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	return rec, s.write(ctx, rec)
}

func (s *MeetingRecordStore) AddParticipantTurn(ctx context.Context, meetingID string, turn AgentMeetingTurn) error {
	return s.update(ctx, meetingID, func(rec *AgentMeetingRecord, now time.Time) {
		turn.AgentID = s.redact(turn.AgentID)
		turn.DelegationID = s.redact(turn.DelegationID)
		turn.Response = s.redact(turn.Response)
		turn.Status = s.redact(turn.Status)
		if turn.CreatedAt.IsZero() {
			turn.CreatedAt = now
		}
		rec.ParticipantTurns = append(rec.ParticipantTurns, turn)
	})
}

func (s *MeetingRecordStore) Completed(ctx context.Context, meetingID string, chair AgentMeetingChairTurn, result AgentMeetingOutcome) error {
	return s.update(ctx, meetingID, func(rec *AgentMeetingRecord, now time.Time) {
		rec.Status = AgentMeetingStatusCompleted
		rec.Recommendation = s.redact(result.Recommendation)
		rec.Timeline = s.redactRefs(result.Timeline)
		rec.Risks = s.redactRefs(result.Risks)
		rec.FollowUps = s.redactRefs(result.FollowUps)
		rec.ChairTurn = &AgentMeetingChairTurn{
			AgentID:  s.redact(chair.AgentID),
			Response: s.redact(chair.Response),
		}
		rec.CompletedAt = &now
		rec.Error = ""
	})
}

func (s *MeetingRecordStore) Failed(ctx context.Context, meetingID string, err error) error {
	if err == nil {
		err = errors.New("meeting failed")
	}
	status := AgentMeetingStatusFailed
	if errors.Is(err, context.Canceled) {
		status = AgentMeetingStatusCancelled
	}
	return s.update(ctx, meetingID, func(rec *AgentMeetingRecord, now time.Time) {
		rec.Status = status
		rec.CompletedAt = &now
		rec.Error = s.redact(err.Error())
	})
}

func (s *MeetingRecordStore) RecordGitHubArtifact(
	ctx context.Context,
	meetingID string,
	write AgentGitHubArtifactWrite,
	artifactRefs []string,
) error {
	return s.update(ctx, meetingID, func(rec *AgentMeetingRecord, now time.Time) {
		write.IssueURL = s.redact(write.IssueURL)
		write.Error = s.redact(write.Error)
		write.UpdatedAt = now
		rec.GitHubArtifact = &write
		rec.ArtifactRefs = appendUniqueRefs(rec.ArtifactRefs, s.redactRefs(artifactRefs)...)
	})
}

func (s *MeetingRecordStore) Get(ctx context.Context, meetingID string) (AgentMeetingRecord, error) {
	if s == nil {
		return AgentMeetingRecord{}, os.ErrNotExist
	}
	if err := ctx.Err(); err != nil {
		return AgentMeetingRecord{}, err
	}
	path, err := s.pathForID(meetingID)
	if err != nil {
		return AgentMeetingRecord{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return AgentMeetingRecord{}, err
	}
	var rec AgentMeetingRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return AgentMeetingRecord{}, err
	}
	return rec, nil
}

func (s *MeetingRecordStore) List(ctx context.Context) ([]AgentMeetingRecord, error) {
	if s == nil {
		return nil, nil
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	records := make([]AgentMeetingRecord, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		rec, err := s.Get(ctx, strings.TrimSuffix(entry.Name(), ".json"))
		if err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	slices.SortFunc(records, func(a, b AgentMeetingRecord) int {
		if cmpCreated := cmp.Compare(a.CreatedAt.UnixNano(), b.CreatedAt.UnixNano()); cmpCreated != 0 {
			return cmpCreated
		}
		return cmp.Compare(a.MeetingID, b.MeetingID)
	})
	return records, nil
}

func (s *MeetingRecordStore) update(ctx context.Context, meetingID string, mutate func(*AgentMeetingRecord, time.Time)) error {
	if s == nil {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	rec, err := s.getUnlocked(ctx, meetingID)
	if err != nil {
		return err
	}
	now := s.now().UTC()
	mutate(&rec, now)
	rec.UpdatedAt = now
	return s.writeUnlocked(rec)
}

func (s *MeetingRecordStore) write(ctx context.Context, rec AgentMeetingRecord) error {
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

func (s *MeetingRecordStore) getUnlocked(ctx context.Context, meetingID string) (AgentMeetingRecord, error) {
	if err := ctx.Err(); err != nil {
		return AgentMeetingRecord{}, err
	}
	path, err := s.pathForID(meetingID)
	if err != nil {
		return AgentMeetingRecord{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return AgentMeetingRecord{}, err
	}
	var rec AgentMeetingRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return AgentMeetingRecord{}, err
	}
	return rec, nil
}

func (s *MeetingRecordStore) writeUnlocked(rec AgentMeetingRecord) error {
	path, err := s.pathForID(rec.MeetingID)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	return fileutil.WriteFileAtomic(path, data, 0o600)
}

func (s *MeetingRecordStore) pathForID(meetingID string) (string, error) {
	meetingID = strings.TrimSpace(meetingID)
	if meetingID == "" || !delegationIDPattern.MatchString(meetingID) {
		return "", fmt.Errorf("%w: invalid meeting ID %q", ErrAgentMeetingInvalidRequest, meetingID)
	}
	filename := meetingID + ".json"
	if !filepath.IsLocal(filename) {
		return "", fmt.Errorf("%w: invalid meeting filename %q", ErrAgentMeetingInvalidRequest, filename)
	}
	return filepath.Join(s.dir, filename), nil
}

func (s *MeetingRecordStore) newID(req AgentMeetingRequest, now time.Time) string {
	input := strings.Join([]string{
		req.SponsorAgentID,
		req.ChairAgentID,
		strings.Join(req.ParticipantAgentIDs, ","),
		req.Title,
		req.Goal,
		now.Format(time.RFC3339Nano),
	}, "\x00")
	sum := sha256.Sum256([]byte(input))
	return fmt.Sprintf("meeting-%s-%s", now.Format("20060102T150405.000000000Z"), hex.EncodeToString(sum[:])[:12])
}

func (s *MeetingRecordStore) redact(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || s.filter == nil {
		return value
	}
	return s.filter(value)
}

func (s *MeetingRecordStore) redactRefs(refs []string) []string {
	out := make([]string, 0, len(refs))
	for _, ref := range refs {
		ref = s.redact(ref)
		if ref != "" {
			out = append(out, ref)
		}
	}
	return out
}
