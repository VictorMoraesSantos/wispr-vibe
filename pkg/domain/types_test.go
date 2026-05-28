package domain

import (
	"testing"
	"time"
)

func TestTranscribeResultFields(t *testing.T) {
	now := time.Now()
	result := TranscribeResult{
		Text:      "hello world",
		Language:  "en",
		Duration:  2 * time.Second,
		CreatedAt: now,
	}

	if result.Text != "hello world" {
		t.Errorf("Text = %q", result.Text)
	}
	if result.Language != "en" {
		t.Errorf("Language = %q", result.Language)
	}
	if result.Duration != 2*time.Second {
		t.Errorf("Duration = %v", result.Duration)
	}
	if result.CreatedAt != now {
		t.Errorf("CreatedAt mismatch")
	}
}

func TestTranscribeOptsDefaults(t *testing.T) {
	opts := TranscribeOpts{}
	if opts.Language != "" {
		t.Errorf("Language should be empty by default, got %q", opts.Language)
	}
	if opts.Temperature != 0 {
		t.Errorf("Temperature should be 0 by default, got %f", opts.Temperature)
	}
}

func TestRecordingResultFields(t *testing.T) {
	result := RecordingResult{
		Data:       []byte{1, 2, 3, 4},
		Format:     "wav",
		SampleRate: 16000,
		Duration:   5 * time.Second,
	}

	if len(result.Data) != 4 {
		t.Errorf("Data len = %d", len(result.Data))
	}
	if result.Format != "wav" {
		t.Errorf("Format = %q", result.Format)
	}
	if result.SampleRate != 16000 {
		t.Errorf("SampleRate = %d", result.SampleRate)
	}
	if result.Duration != 5*time.Second {
		t.Errorf("Duration = %v", result.Duration)
	}
}
