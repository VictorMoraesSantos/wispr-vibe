# Architecture

## Principles

- **Interface-driven**: every external dependency behind an interface
- **Pipeline**: audio → STT → processor → output (each step independent)
- **Config-driven**: behavior controlled by config, not hardcoded
- **Testable**: each module testable in isolation with mocks

## Directory Structure

```
cmd/
  vibevoice/
    main.go              # CLI entry, wire dependencies, run app

internal/
  app/
    app.go               # Orchestrator: ties pipeline together
  audio/
    recorder.go          # Microphone capture interface + impl
    wav.go               # WAV encoding utilities
  stt/
    engine.go            # STT interface (Transcriber)
    whisper_api.go       # OpenAI Whisper API implementation
    whisper_local.go     # (future) local whisper.cpp
  processor/
    pipeline.go          # Chain of text transforms
    filters.go           # Individual filter functions
  hotkey/
    hotkey.go            # Global hotkey interface
    hotkey_windows.go    # Windows implementation (future)
  clipboard/
    clipboard.go         # Copy/paste interface + impl
  config/
    config.go            # Load config from file/env
  logger/
    logger.go            # Structured logging setup

pkg/
  domain/
    types.go             # Shared domain types (TranscribeResult, etc.)

tests/                   # Integration tests
go.mod
go.sum
```

## Key Interfaces

```go
// STT engine
type Transcriber interface {
    Transcribe(ctx context.Context, audio []byte, opts TranscribeOpts) (string, error)
}

// Audio capture
type Recorder interface {
    Start() error
    Stop() ([]byte, error) // returns WAV bytes
}

// Text processor
type Processor interface {
    Process(text string) string
}

// Clipboard
type Clipboard interface {
    Copy(text string) error
    Paste() error
}
```
