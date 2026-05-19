package agent

import (
	"testing"

	"github.com/sipeed/picoclaw/pkg/config"
)

func TestInferMediaType(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		contentType string
		want        string
	}{
		{
			name:        "png content type",
			filename:    "diagram",
			contentType: "image/png",
			want:        "image",
		},
		{
			name:        "jpeg extension fallback",
			filename:    "photo.JPG",
			contentType: "",
			want:        "image",
		},
		{
			name:        "svg content type is file",
			filename:    "diagram",
			contentType: "image/svg+xml",
			want:        "file",
		},
		{
			name:        "svg content type with parameters is file",
			filename:    "diagram",
			contentType: "image/svg+xml; charset=utf-8",
			want:        "file",
		},
		{
			name:        "svg extension fallback is file",
			filename:    "diagram.SVG",
			contentType: "",
			want:        "file",
		},
		{
			name:        "audio content type",
			filename:    "voice",
			contentType: "audio/ogg",
			want:        "audio",
		},
		{
			name:        "ogg application content type",
			filename:    "voice.ogg",
			contentType: "application/ogg",
			want:        "audio",
		},
		{
			name:        "video extension fallback",
			filename:    "clip.MP4",
			contentType: "",
			want:        "video",
		},
		{
			name:        "unknown type",
			filename:    "archive.bin",
			contentType: "application/octet-stream",
			want:        "file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferMediaType(tt.filename, tt.contentType)
			if got != tt.want {
				t.Fatalf("inferMediaType(%q, %q) = %q, want %q", tt.filename, tt.contentType, got, tt.want)
			}
		})
	}
}

func TestShouldPublishToolFeedbackSkipsInternalChannels(t *testing.T) {
	cfg := &config.Config{
		Agents: config.AgentsConfig{
			Defaults: config.AgentDefaults{
				ToolFeedback: config.ToolFeedbackConfig{Enabled: true},
			},
		},
	}

	for _, channel := range []string{"cli", "system", "subagent", "internal"} {
		t.Run(channel, func(t *testing.T) {
			ts := &turnState{channel: channel}
			if shouldPublishToolFeedback(cfg, ts) {
				t.Fatalf("shouldPublishToolFeedback(%q) = true, want false", channel)
			}
		})
	}
}

func TestShouldPublishToolFeedbackAllowsExternalChannelsWhenEnabled(t *testing.T) {
	cfg := &config.Config{
		Agents: config.AgentsConfig{
			Defaults: config.AgentDefaults{
				ToolFeedback: config.ToolFeedbackConfig{Enabled: true},
			},
		},
	}

	ts := &turnState{channel: "discord"}
	if !shouldPublishToolFeedback(cfg, ts) {
		t.Fatal("shouldPublishToolFeedback(discord) = false, want true")
	}
}
