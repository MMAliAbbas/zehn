package constants

import "testing"

func TestIsInternalChannelIncludesDelegationInternal(t *testing.T) {
	for _, channel := range []string{"cli", "system", "subagent", "internal"} {
		if !IsInternalChannel(channel) {
			t.Fatalf("IsInternalChannel(%q) = false, want true", channel)
		}
	}
}

func TestIsInternalChannelRejectsExternalChannels(t *testing.T) {
	for _, channel := range []string{"discord", "telegram", "pico", ""} {
		if IsInternalChannel(channel) {
			t.Fatalf("IsInternalChannel(%q) = true, want false", channel)
		}
	}
}
