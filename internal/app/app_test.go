package app

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/victorlui/wispr-vibe/internal/config"
	"github.com/victorlui/wispr-vibe/pkg/domain"
)

// mockTranscriber implements stt.Transcriber for testing.
type mockTranscriber struct {
	result *domain.TranscribeResult
	err    error
	called bool
}

func (m *mockTranscriber) Transcribe(ctx context.Context, audio []byte, opts domain.TranscribeOpts) (*domain.TranscribeResult, error) {
	m.called = true
	return m.result, m.err
}

func (m *mockTranscriber) Name() string {
	return "mock"
}

func TestNewApp(t *testing.T) {
	cfg := config.Default()
	cfg.WhisperAPIKey = "test-key"
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	app := New(cfg, log)
	if app == nil {
		t.Fatal("New() returned nil")
	}
	if app.IsRecording() {
		t.Error("new app should not be recording")
	}
}

func TestNewAppLocalEngine(t *testing.T) {
	cfg := config.Default()
	cfg.STTEngine = "whisper_local"
	cfg.WhisperAPIKey = "fallback-key"
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	// Local whisper won't be found, should fall back to API
	app := New(cfg, log)
	if app == nil {
		t.Fatal("New() returned nil even without whisper.cpp")
	}
}

func TestAppStopWithoutStart(t *testing.T) {
	cfg := config.Default()
	cfg.WhisperAPIKey = "key"
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	app := New(cfg, log)
	_, err := app.StopAndProcess(context.Background())
	if err == nil {
		t.Fatal("StopAndProcess without Start should error")
	}
	if !containsStr(err.Error(), "not recording") {
		t.Errorf("error should mention 'not recording': %v", err)
	}
}

func TestTranscriberInterface(t *testing.T) {
	mock := &mockTranscriber{
		result: &domain.TranscribeResult{
			Text:     "hello world",
			Language: "en",
		},
	}

	result, err := mock.Transcribe(context.Background(), []byte("audio"), domain.TranscribeOpts{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text != "hello world" {
		t.Errorf("text = %q, want %q", result.Text, "hello world")
	}
	if !mock.called {
		t.Error("Transcribe should have been called")
	}
}

func TestTranscriberError(t *testing.T) {
	mock := &mockTranscriber{
		err: errors.New("api timeout"),
	}

	_, err := mock.Transcribe(context.Background(), []byte("audio"), domain.TranscribeOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "api timeout" {
		t.Errorf("error = %q, want %q", err.Error(), "api timeout")
	}
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
