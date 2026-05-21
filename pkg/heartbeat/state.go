package heartbeat

import (
	"context"
	"encoding/json"
	"path/filepath"
	"time"

	"github.com/sipeed/picoclaw/pkg/fileutil"
)

const heartbeatStatePath = "heartbeat/state.json"

type runStatus string

const (
	runStatusRunning runStatus = "running"
	runStatusOK      runStatus = "ok"
	runStatusAction  runStatus = "action"
	runStatusError   runStatus = "error"
	runStatusTimeout runStatus = "timeout"
	runStatusSkipped runStatus = "skipped"
)

type heartbeatRunState struct {
	RunID         string    `json:"run_id"`
	Status        runStatus `json:"status"`
	StartedAt     time.Time `json:"started_at"`
	CompletedAt   time.Time `json:"completed_at,omitempty"`
	DurationMS    int64     `json:"duration_ms,omitempty"`
	Channel       string    `json:"channel,omitempty"`
	ChatID        string    `json:"chat_id,omitempty"`
	Error         string    `json:"error,omitempty"`
	ResponseBytes int       `json:"response_bytes,omitempty"`
}

func (hs *HeartbeatService) writeRunState(ctx context.Context, state heartbeatRunState) {
	if hs == nil || hs.workspace == "" {
		return
	}
	if err := ctx.Err(); err != nil {
		return
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return
	}
	_ = fileutil.WriteFileAtomic(filepath.Join(hs.workspace, heartbeatStatePath), data, 0o600)
}
