package toast

import (
	"testing"
)

func TestFormatRecordingText(t *testing.T) {
	tests := []struct {
		seconds int
		expect  string
	}{
		{0, "🎙 Recording... 0s"},
		{1, "🎙 Recording... 1s"},
		{5, "🎙 Recording... 5s"},
		{30, "🎙 Recording... 30s"},
		{120, "🎙 Recording... 120s"},
	}

	for _, tt := range tests {
		got := FormatRecordingText(tt.seconds)
		if got != tt.expect {
			t.Errorf("FormatRecordingText(%d) = %q, want %q", tt.seconds, got, tt.expect)
		}
	}
}
