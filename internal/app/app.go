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

	// Clipboard
	if err := clipboard.Copy(text); err != nil {
		return text, fmt.Errorf("copy to clipboard: %w", err)
	}
	a.log.Info("text copied to clipboard")

	// Auto-paste if configured
	if a.cfg.AutoPaste {
		time.Sleep(100 * time.Millisecond) // small delay for clipboard sync
		if err := clipboard.Paste(); err != nil {
			a.log.Warn("auto-paste failed", "error", err)
		} else {
			a.log.Info("auto-pasted")
		}
	}

	return text, nil
}

func (a *App) IsRecording() bool {
	return a.recorder.IsRecording()
}
