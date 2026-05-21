package heartbeat

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sipeed/picoclaw/pkg/tools"
)

func TestExecuteHeartbeat_Async(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "heartbeat-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hs := NewHeartbeatService(tmpDir, 30, true)
	hs.stopChan = make(chan struct{}) // Enable for testing

	asyncCalled := false
	asyncResult := &tools.ToolResult{
		ForLLM:  "Background task started",
		ForUser: "Task started in background",
		Silent:  false,
		IsError: false,
		Async:   true,
	}

	hs.SetHandler(func(prompt, channel, chatID string) *tools.ToolResult {
		asyncCalled = true
		if prompt == "" {
			t.Error("Expected non-empty prompt")
		}
		return asyncResult
	})

	// Create HEARTBEAT.md
	os.WriteFile(filepath.Join(tmpDir, "HEARTBEAT.md"), []byte("Test task"), 0o644)

	// Execute heartbeat directly (internal method for testing)
	hs.executeHeartbeat()

	if !asyncCalled {
		t.Error("Expected handler to be called")
	}
}

func TestExecuteHeartbeat_ResultLogging(t *testing.T) {
	tests := []struct {
		name    string
		result  *tools.ToolResult
		wantLog string
	}{
		{
			name: "error result",
			result: &tools.ToolResult{
				ForLLM:  "Heartbeat failed: connection error",
				ForUser: "",
				Silent:  false,
				IsError: true,
				Async:   false,
			},
			wantLog: "error message",
		},
		{
			name: "silent result",
			result: &tools.ToolResult{
				ForLLM:  "Heartbeat completed successfully",
				ForUser: "",
				Silent:  true,
				IsError: false,
				Async:   false,
			},
			wantLog: "completion message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "heartbeat-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			hs := NewHeartbeatService(tmpDir, 30, true)
			hs.stopChan = make(chan struct{}) // Enable for testing

			hs.SetHandler(func(prompt, channel, chatID string) *tools.ToolResult {
				return tt.result
			})

			os.WriteFile(filepath.Join(tmpDir, "HEARTBEAT.md"), []byte("Test task"), 0o644)
			hs.executeHeartbeat()

			logFile := filepath.Join(tmpDir, "heartbeat.log")
			data, err := os.ReadFile(logFile)
			if err != nil {
				t.Fatalf("Failed to read log file: %v", err)
			}
			if string(data) == "" {
				t.Errorf("Expected log file to contain %s", tt.wantLog)
			}
		})
	}
}

func TestHeartbeatService_StartStop(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "heartbeat-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hs := NewHeartbeatService(tmpDir, 1, true)

	err = hs.Start()
	if err != nil {
		t.Fatalf("Failed to start heartbeat service: %v", err)
	}

	hs.Stop()

	time.Sleep(100 * time.Millisecond)
}

func TestHeartbeatService_Disabled(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "heartbeat-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hs := NewHeartbeatService(tmpDir, 1, false)

	if hs.enabled != false {
		t.Error("Expected service to be disabled")
	}

	err = hs.Start()
	_ = err // Disabled service returns nil
}

func TestExecuteHeartbeat_NilResult(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "heartbeat-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hs := NewHeartbeatService(tmpDir, 30, true)
	hs.stopChan = make(chan struct{}) // Enable for testing

	hs.SetHandler(func(prompt, channel, chatID string) *tools.ToolResult {
		return nil
	})

	// Create HEARTBEAT.md
	os.WriteFile(filepath.Join(tmpDir, "HEARTBEAT.md"), []byte("Test task"), 0o644)

	// Should not panic with nil result
	hs.executeHeartbeat()
}

func TestExecuteHeartbeat_SkipsOverlappingRun(t *testing.T) {
	tmpDir := t.TempDir()
	hs := NewHeartbeatService(tmpDir, 30, true)
	hs.stopChan = make(chan struct{})
	hs.SetMaxRunDuration(time.Second)

	started := make(chan struct{})
	release := make(chan struct{})
	var calls int32
	hs.SetHandler(func(prompt, channel, chatID string) *tools.ToolResult {
		if atomic.AddInt32(&calls, 1) == 1 {
			close(started)
			<-release
		}
		return &tools.ToolResult{ForLLM: "ok", Silent: true}
	})

	if err := os.WriteFile(filepath.Join(tmpDir, "HEARTBEAT.md"), []byte("Test task"), 0o644); err != nil {
		t.Fatalf("WriteFile(HEARTBEAT.md) error = %v", err)
	}

	doneFirst := make(chan struct{})
	go func() {
		hs.executeHeartbeat()
		close(doneFirst)
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("first heartbeat did not start")
	}

	hs.executeHeartbeat()
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("handler calls = %d, want 1 while first heartbeat is active", got)
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, heartbeatStatePath))
	if err != nil {
		t.Fatalf("ReadFile(state) error = %v", err)
	}
	var state heartbeatRunState
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("Unmarshal(state) error = %v", err)
	}
	if state.Status != runStatusSkipped {
		t.Fatalf("state.Status = %q, want %q", state.Status, runStatusSkipped)
	}

	close(release)
	select {
	case <-doneFirst:
	case <-time.After(time.Second):
		t.Fatal("first heartbeat did not finish after release")
	}
}

func TestExecuteHeartbeat_TimesOutAndRecordsTerminalState(t *testing.T) {
	tmpDir := t.TempDir()
	hs := NewHeartbeatService(tmpDir, 30, true)
	hs.stopChan = make(chan struct{})
	hs.SetMaxRunDuration(20 * time.Millisecond)

	release := make(chan struct{})
	hs.SetHandler(func(prompt, channel, chatID string) *tools.ToolResult {
		<-release
		return &tools.ToolResult{ForLLM: "late", Silent: true}
	})
	defer close(release)

	if err := os.WriteFile(filepath.Join(tmpDir, "HEARTBEAT.md"), []byte("Test task"), 0o644); err != nil {
		t.Fatalf("WriteFile(HEARTBEAT.md) error = %v", err)
	}

	start := time.Now()
	hs.executeHeartbeat()
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Fatalf("executeHeartbeat took %s, want timeout to return quickly", elapsed)
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, heartbeatStatePath))
	if err != nil {
		t.Fatalf("ReadFile(state) error = %v", err)
	}
	var state heartbeatRunState
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("Unmarshal(state) error = %v", err)
	}
	if state.Status != runStatusTimeout {
		t.Fatalf("state.Status = %q, want %q", state.Status, runStatusTimeout)
	}
	if !strings.Contains(state.Error, "exceeded timeout") {
		t.Fatalf("state.Error = %q, want timeout detail", state.Error)
	}
}

// TestLogPath verifies heartbeat log is written to workspace directory
func TestLogPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "heartbeat-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hs := NewHeartbeatService(tmpDir, 30, true)

	// Write a log entry
	hs.logf("INFO", "Test log entry")

	// Verify log file exists at workspace root
	expectedLogPath := filepath.Join(tmpDir, "heartbeat.log")
	if _, err := os.Stat(expectedLogPath); os.IsNotExist(err) {
		t.Errorf("Expected log file at %s, but it doesn't exist", expectedLogPath)
	}
}

// TestHeartbeatFilePath verifies HEARTBEAT.md is at workspace root
func TestHeartbeatFilePath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "heartbeat-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hs := NewHeartbeatService(tmpDir, 30, true)

	// Trigger default template creation
	hs.buildPrompt()

	// Verify HEARTBEAT.md exists at workspace root
	expectedPath := filepath.Join(tmpDir, "HEARTBEAT.md")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected HEARTBEAT.md at %s, but it doesn't exist", expectedPath)
	}
}

func TestBuildPrompt_DefaultTemplateStaysIdle(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "heartbeat-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hs := NewHeartbeatService(tmpDir, 30, true)
	hs.createDefaultHeartbeatTemplate()

	if prompt := hs.buildPrompt(); prompt != "" {
		t.Fatalf("buildPrompt() = %q, want empty prompt for untouched default template", prompt)
	}
}

func TestBuildPrompt_UserTasksAfterMarkerProducePrompt(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "heartbeat-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	hs := NewHeartbeatService(tmpDir, 30, true)
	hs.createDefaultHeartbeatTemplate()

	path := filepath.Join(tmpDir, "HEARTBEAT.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read HEARTBEAT.md: %v", err)
	}
	updated := string(data) + "\n- Check unread Feishu messages\n"
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		t.Fatalf("Failed to update HEARTBEAT.md: %v", err)
	}

	prompt := hs.buildPrompt()
	if prompt == "" {
		t.Fatal("buildPrompt() = empty, want non-empty prompt when user tasks are present")
	}
	if !strings.Contains(prompt, "Check unread Feishu messages") {
		t.Fatalf("prompt = %q, want user task content", prompt)
	}
}
