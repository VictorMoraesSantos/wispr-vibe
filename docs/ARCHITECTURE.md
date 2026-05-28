# Architecture

## Principles

- **Interface-driven**: every external dependency behind an interface
- **Pipeline**: hotkey → audio → STT → processor → auto-type (each step independent)
- **Config-driven**: behavior controlled by JSON config, not hardcoded
- **Testable**: each module testable in isolation with mocks
- **Thread-safe**: Win32 APIs on locked OS threads, Fyne UI via `fyne.Do()`
- **Non-intrusive**: never steals focus, never interrupts workflow

## Directory Structure

```
cmd/
  vibevoice/
    main.go                  # CLI entry point (interactive r/s/q loop)
  vibevoice-gui/
    main.go                  # GUI entry point (Fyne + hotkey + toast)

internal/
  app/
    app.go                   # Pipeline orchestrator (record → STT → process → type)
  audio/
    recorder.go              # FFmpeg subprocess audio capture
    ffmpeg_windows.go        # Windows: device detection, ffmpeg discovery, dshow args
    ffmpeg_linux.go          # Linux: ALSA/PulseAudio args
    ffmpeg_darwin.go         # macOS: avfoundation args
  stt/
    engine.go                # Transcriber interface
    whisper_api.go           # OpenAI/MiniMax API implementation
    whisper_local.go         # Local whisper.cpp subprocess
  processor/
    pipeline.go              # Chain of text transform filters
    filters.go               # Individual filters (fillers, punctuation, trim)
    filters_test.go          # Unit tests for filters
  hotkey/
    hotkey_windows.go        # Win32 RegisterHotKey + message loop (locked thread)
    hotkey_other.go          # Stub for non-Windows
  clipboard/
    clipboard_windows.go     # Win32 clipboard API + SendInput (Ctrl+V)
    clipboard.go             # Non-Windows: pbcopy/xclip + xdotool/osascript
  toast/
    toast_windows.go         # Native Win32 popup overlay (topmost, no-activate)
    toast_other.go           # Stub for non-Windows
  config/
    config.go                # JSON config, Load/Save, setup wizard
    config_test.go           # Config tests
  logger/
    logger.go                # Structured logging with slog

pkg/
  domain/
    types.go                 # Shared types (TranscribeResult, TranscribeOpts)

docs/
  SPEC.md                    # Product specification
  ARCHITECTURE.md            # This file
```

## Key Interfaces

```go
// STT engine — swap between local and cloud transparently
type Transcriber interface {
    Name() string
    Transcribe(ctx context.Context, audio []byte, opts TranscribeOpts) (*TranscribeResult, error)
}

// Implementations:
// - WhisperLocal: runs whisper.cpp as subprocess, reads stdout
// - WhisperAPI: HTTP multipart POST to OpenAI-compatible endpoint
```

## Platform-Specific Components (Windows)

### Audio Capture (FFmpeg + dshow)
- Auto-detects microphone device via `ffmpeg -list_devices true -f dshow -i dummy`
- Auto-discovers ffmpeg binary in winget/scoop/choco/common paths
- Captures 16kHz mono WAV via subprocess

### Global Hotkey (Win32 RegisterHotKey)
- Uses `runtime.LockOSThread()` to keep RegisterHotKey + GetMessage on same thread
- Supports configurable combo string parsing ("Ctrl+Shift+R" → modifiers + VK)
- PostThreadMessage(WM_QUIT) for clean unregister

### Auto-Type (Win32 SendInput)
- Copies text to clipboard via OpenClipboard/SetClipboardData (CF_UNICODETEXT)
- Simulates Ctrl+V via SendInput (4 INPUT structs: Ctrl↓ V↓ V↑ Ctrl↑)
- 50ms delay between clipboard set and keypress for reliability

### Toast Overlay (Win32 CreateWindowEx)
- Styles: WS_EX_TOPMOST | WS_EX_TOOLWINDOW | WS_EX_LAYERED | WS_EX_NOACTIVATE
- WS_POPUP (no border, no title bar)
- Positioned bottom-center, 60px from screen edge
- 86% opacity, dark background, rounded corners
- GDI rendering: RoundRect + DrawText (Segoe UI 16px)

### GUI (Fyne v2.7)
- All UI mutations from background threads wrapped in `fyne.Do()`
- System tray via desktop extension interface
- Window hides on close (never quits, lives in tray)
- Settings window for hotkey configuration

## Data Flow

```
1. User presses Ctrl+Shift+R
2. Win32 message loop fires WM_HOTKEY
3. toggleRecording() called via fyne.Do()
4. FFmpeg subprocess starts capturing mic → temp WAV
5. Toast overlay appears: "🎙 Recording..."
6. User presses Ctrl+Shift+R again
7. FFmpeg stopped, WAV data collected
8. Toast: "⏳ Transcribing..."
9. whisper.cpp runs on WAV → raw text
10. Processor pipeline cleans text (fillers, punctuation, spacing)
11. Text copied to Win32 clipboard (CF_UNICODETEXT)
12. SendInput simulates Ctrl+V into active window
13. Toast hides
14. Text appears at cursor position in whatever app was focused
```

## Configuration

```json
{
  "stt_engine": "whisper_local",
  "provider": "local",
  "whisper_exe_path": "",
  "whisper_model_path": "C:\\Users\\...\\ggml-small.bin",
  "language": "pt",
  "hotkey": "Ctrl+Shift+R",
  "sample_rate": 16000,
  "log_level": "info",
  "remove_fillers": true,
  "fix_punctuation": true
}
```

Auto-discovery paths:
- whisper.cpp: `~/.wispr-vibe/whisper-bin/Release/whisper-cli.exe`
- Models: `~/.wispr-vibe/models/ggml-*.bin` (prefers larger models)
- FFmpeg: winget/scoop/choco install paths + common locations

## Dependencies

| Package | Purpose |
|---------|---------|
| fyne.io/fyne/v2 | Cross-platform GUI framework |
| Go stdlib (syscall, unsafe) | Win32 API calls |
| FFmpeg (external) | Audio capture subprocess |
| whisper.cpp (external) | Local STT subprocess |
