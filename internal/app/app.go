package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/victorlui/wispr-vibe/internal/audio"
	"github.com/victorlui/wispr-vibe/internal/clipboard"
	"github.com/victorlui/wispr-vibe/internal/config"
	"github.com/victorlui/wispr-vibe/internal/processor"
	"github.com/victorlui/wispr-vibe/internal/stt"
	"github.com/victorlui/wispr-vibe/pkg/domain"
)

// App orchestrates the full voice-to-text pipeline.
type App struct {
	cfg        *config.Config
	recorder   *audio.Recorder
	transcriber stt.Transcriber
	pipeline   *processor.Pipeline
	log        *slog.Logger
}

func New(cfg *config.Config, log *slog.Logger) *App {
	var transcriber stt.Transcriber
	switch cfg.STTEngine {
	case "whisper_local":
		t, err := stt.NewWhisperLocal(cfg.WhisperExePath, cfg.WhisperModelPath)
		if err != nil {
			log.Warn("local whisper not available, will fail on transcribe", "error", err)
			transcriber = stt.NewWhisperAPI(cfg.WhisperAPIKey, cfg.WhisperAPIURL, cfg.WhisperModel)
		} else {
			transcriber = t
			log.Info("using local whisper", "exe", t.Name())
		}
	default:
		transcriber = stt.NewWhisperAPI(cfg.WhisperAPIKey, cfg.WhisperAPIURL, cfg.WhisperModel)
	}

	pipeline := processor.DefaultPipeline()
	if cfg.FixPunctuation {
		pipeline.Add(processor.EnsurePeriod)
	}

	return &App{
		cfg:        cfg,
		recorder:   audio.NewRecorder(cfg.SampleRate),
		transcriber: transcriber,
		pipeline:   pipeline,
		log:        log,
	}
}

// StartRecording begins capturing microphone audio.
func (a *App) StartRecording() error {
	a.log.Info("starting recording")
	return a.recorder.Start()
}

// StopAndProcess stops recording, transcribes, processes, and copies to clipboard.
func (a *App) StopAndProcess(ctx context.Context) (string, error) {
	a.log.Info("stopping recording")

	wavData, err := a.recorder.Stop()
	if err != nil {
		return "", fmt.Errorf("stop recording: %w", err)
	}
	a.log.Debug("audio captured", "bytes", len(wavData))

	// Transcribe
	start := time.Now()
	result, err := a.transcriber.Transcribe(ctx, wavData, domain.TranscribeOpts{
		Language: a.cfg.Language,
	})
	if err != nil {
		return "", fmt.Errorf("transcribe: %w", err)
	}
	a.log.Info("transcription complete",
		"text_len", len(result.Text),
		"duration", time.Since(start),
	)

	// Process
	text := a.pipeline.Process(result.Text)
	a.log.Debug("processed text", "text", text)

	// Type text into active window (copy to clipboard + Ctrl+V)
	if err := clipboard.TypeText(text); err != nil {
		a.log.Warn("auto-type failed, text is in clipboard", "error", err)
	} else {
		a.log.Info("text typed into active window")
	}

	return text, nil
}

func (a *App) IsRecording() bool {
	return a.recorder.IsRecording()
}
