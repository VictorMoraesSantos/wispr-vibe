package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/victorlui/wispr-vibe/internal/audio"
	"github.com/victorlui/wispr-vibe/internal/clipboard"
	"github.com/victorlui/wispr-vibe/internal/config"
	"github.com/victorlui/wispr-vibe/internal/processor"
	"github.com/victorlui/wispr-vibe/internal/stt"
	"github.com/victorlui/wispr-vibe/pkg/domain"
)

type App struct {
	cfg         *config.Config
	recorder    *audio.Recorder
	transcriber stt.Transcriber
	pipeline    *processor.Pipeline
	log         *slog.Logger
}

func New(cfg *config.Config, log *slog.Logger) *App {
	transcriber := buildTranscriber(cfg, log)

	pipeline := processor.DefaultPipeline()
	if cfg.FixPunctuation {
		pipeline.Add(processor.EnsurePeriod)
	}

	return &App{
		cfg:         cfg,
		recorder:    audio.NewRecorder(cfg.SampleRate),
		transcriber: transcriber,
		pipeline:    pipeline,
		log:         log,
	}
}

func buildTranscriber(cfg *config.Config, log *slog.Logger) stt.Transcriber {
	if cfg.STTEngine == "whisper_local" {
		t, err := stt.NewWhisperLocal(cfg.WhisperExePath, cfg.WhisperModelPath, cfg.UseGPU)
		if err == nil {
			return t
		}
		log.Warn("local whisper unavailable, falling back to API", "error", err)
	}
	return stt.NewWhisperAPI(cfg.WhisperAPIKey, cfg.WhisperAPIURL, cfg.WhisperModel)
}

func (a *App) StartRecording() error {
	return a.recorder.Start()
}

func (a *App) StopAndProcess(ctx context.Context) (string, error) {
	wavData, err := a.recorder.Stop()
	if err != nil {
		return "", fmt.Errorf("stop recording: %w", err)
	}

	result, err := a.transcriber.Transcribe(ctx, wavData, domain.TranscribeOpts{
		Language: a.cfg.Language,
	})
	if err != nil {
		return "", fmt.Errorf("transcribe: %w", err)
	}

	text := a.pipeline.Process(result.Text)

	if err := clipboard.TypeText(text); err != nil {
		a.log.Warn("auto-type failed, text in clipboard", "error", err)
	}

	return text, nil
}

func (a *App) IsRecording() bool {
	return a.recorder.IsRecording()
}
