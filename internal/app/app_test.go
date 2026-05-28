package app

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/victorlui/wispr-vibe/internal/audio"
	"github.com/victorlui/wispr-vibe/internal/config"
	"github.com/victorlui/wispr-vibe/internal/processor"
	"github.com/victorlui/wispr-vibe/internal/stt"
	"github.com/victorlui/wispr-vibe/pkg/domain"
)

// mockTranscriber implements stt.Transcriber for testing.
type mockTranscriber struct {
	result *domain.TranscribeResult
	err    error
	called bool
}

func (m *mockTranscriber) Transcribe(_ context.Context, _ []byte, opts domain.TranscribeOpts) (*domain.TranscribeResult, error) {
	m.called = true
	return m.result, m.err
}

func (m *mockTranscriber) Name() string { return "mock" }

// newForTesting constructs an App with an injected transcriber. Test-only.
func newForTesting(cfg *config.Config, log *slog.Logger, t stt.Transcriber) *App {
	pipeline := processor.DefaultPipeline()
	if cfg.FixPunctuation {
		pipeline.Add(processor.EnsurePeriod)
	}
	return &App{
		cfg:         cfg,
		recorder:    audio.NewRecorder(cfg.SampleRate),
		transcriber: t,
		pipeline:    pipeline,
		log:         log,
	}
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestNewApp(t *testing.T) {
	cfg := config.Default()
	cfg.WhisperAPIKey = "test-key"

	app := New(cfg, newTestLogger())
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

	// Local whisper won't be found in CI; should fall back to API.
	app := New(cfg, newTestLogger())
	if app == nil {
		t.Fatal("New() returned nil even without whisper.cpp")
	}
}

func TestAppStopWithoutStart(t *testing.T) {
	cfg := config.Default()
	cfg.WhisperAPIKey = "key"

	app := New(cfg, newTestLogger())
	_, err := app.StopAndProcess(context.Background())
	if err == nil {
		t.Fatal("StopAndProcess without Start should error")
	}
	if !strings.Contains(err.Error(), "not recording") {
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
	mock := &mockTranscriber{err: errors.New("api timeout")}

	_, err := mock.Transcribe(context.Background(), []byte("audio"), domain.TranscribeOpts{})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "api timeout" {
		t.Errorf("error = %q, want %q", err.Error(), "api timeout")
	}
}

func TestAppPipelineRemovesFillers(t *testing.T) {
	cfg := config.Default()
	cfg.FixPunctuation = false

	app := newForTesting(cfg, newTestLogger(), &mockTranscriber{})

	input := "uh I want to create a function"
	got := app.pipeline.Process(input)

	if strings.Contains(strings.ToLower(got), "uh ") {
		t.Errorf("pipeline should remove filler 'uh', got: %q", got)
	}
	if got == "" {
		t.Error("pipeline should not discard non-filler content")
	}
	if got[0] < 'A' || got[0] > 'Z' {
		t.Errorf("pipeline should capitalize first letter, got: %q", got)
	}
}

func TestAppPipelineAddsPeriodWhenConfigured(t *testing.T) {
	cfg := config.Default()
	cfg.FixPunctuation = true

	app := newForTesting(cfg, newTestLogger(), &mockTranscriber{})
	got := app.pipeline.Process("hello world")

	if !strings.HasSuffix(got, ".") {
		t.Errorf("FixPunctuation=true: text should end with '.', got: %q", got)
	}
}

func TestAppPipelineNoPeriodWhenDisabled(t *testing.T) {
	cfg := config.Default()
	cfg.FixPunctuation = false

	app := newForTesting(cfg, newTestLogger(), &mockTranscriber{})
	got := app.pipeline.Process("hello world")

	if strings.HasSuffix(got, ".") {
		t.Errorf("FixPunctuation=false: text should not end with '.', got: %q", got)
	}
}

func TestAppIsRecordingInitiallyFalse(t *testing.T) {
	app := newForTesting(config.Default(), newTestLogger(), &mockTranscriber{})
	if app.IsRecording() {
		t.Error("app should not be recording on creation")
	}
}
