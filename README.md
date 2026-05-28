# wispr-vibe

Speech-to-text tool for developers, optimized for vibe coding.

Dictate long instructions → transcribe → clean → paste into IDE/terminal/browser.

## Quick Start

### Prerequisites

- Go 1.21+
- [ffmpeg](https://ffmpeg.org/download.html) installed and in PATH
- OpenAI API key (for Whisper API)

### Setup

```bash
# Clone
git clone https://github.com/victorlui/wispr-vibe.git
cd wispr-vibe

# Set API key
export WISPR_API_KEY="sk-..."   # Linux/Mac
set WISPR_API_KEY=sk-...        # Windows

# Build
go build -o vibevoice ./cmd/vibevoice/

# Run
./vibevoice
```

### Usage

```
[r] Start recording
[s] Stop recording & transcribe
[q] Quit
```

Audio is captured via ffmpeg, sent to Whisper API, cleaned up (fillers removed, punctuation fixed), and copied to clipboard.

## Architecture

```
Hotkey → Audio Capture → STT Engine → Text Processor → Clipboard/Paste
```

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for details.

## Configuration

Create `~/.wispr-vibe/config.json`:

```json
{
  "stt_engine": "whisper_api",
  "whisper_api_key": "sk-...",
  "whisper_model": "whisper-1",
  "language": "pt",
  "sample_rate": 16000,
  "log_level": "info",
  "remove_fillers": true,
  "fix_punctuation": true,
  "auto_paste": false
}
```

Or use env vars: `WISPR_API_KEY`, `WISPR_API_URL`.

## Roadmap

- [x] Phase 1: MVP (record, transcribe via API, clean, clipboard)
- [ ] Phase 2: Local Whisper (whisper.cpp)
- [ ] Phase 3: Global hotkey (push-to-talk)
- [ ] Phase 4: Auto-paste into active window
- [ ] Phase 5: LLM-powered text cleanup
- [ ] Phase 6: GUI/tray icon

## License

MIT
