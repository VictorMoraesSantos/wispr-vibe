# Wispr Vibe — Product Specification

## Vision

Desktop speech-to-text tool for developers, optimized for vibe coding.
Dictate long, clear instructions → transcribe → clean → insert into IDE/terminal/browser.

## Core Requirements

### Functional

1. CLI app starts listening on hotkey/command
2. Capture microphone audio while key held or toggle active
3. Send audio to STT engine (local Whisper preferred, API fallback)
4. Post-process transcription:
   - Remove filler words (uh, um, tipo, né)
   - Fix spacing and punctuation
   - Clean verbal auxiliary commands
5. Copy result to clipboard
6. Optionally paste into active window (simulate Ctrl+V or type)
7. Structured logging for debug

### Non-Functional

- Low latency (< 2s end-to-end for short phrases)
- Modular architecture, swap STT backends without touching other code
- Offline-first (local Whisper as primary)
- Cross-platform target: Windows first, Linux/macOS later
- Minimal resource footprint when idle

## Architecture

```
┌─────────┐    ┌─────────┐    ┌──────────┐    ┌───────────┐    ┌───────────┐
│  Hotkey  │───▶│  Audio  │───▶│   STT    │───▶│ Processor │───▶│ Clipboard │
│  Trigger │    │ Capture │    │  Engine  │    │  Pipeline │    │  / Paste  │
└─────────┘    └─────────┘    └──────────┘    └───────────┘    └───────────┘
                                    │
                              ┌─────┴─────┐
                              │  Whisper   │
                              │  Local/API │
                              └───────────┘
```

## Modules

| Module       | Responsibility                          |
|-------------|------------------------------------------|
| audio       | Microphone capture, WAV encoding         |
| stt         | Transcription interface + implementations |
| processor   | Text cleanup pipeline                    |
| hotkey      | Global hotkey registration               |
| clipboard   | Copy text, simulate paste                |
| config      | YAML/env config loading                  |
| logger      | Structured logging (slog)                |
| app         | Orchestration, lifecycle                 |

## MVP Scope (Phase 1)

Smallest vertical slice:
1. CLI start
2. Record audio on key press (or simple stdin trigger)
3. Transcribe via Whisper API (OpenAI-compatible endpoint)
4. Basic text cleanup
5. Copy to clipboard
6. Log everything

## Phases

- **Phase 1**: MVP — record, transcribe (API), clean, clipboard
- **Phase 2**: Local Whisper via whisper.cpp / go bindings
- **Phase 3**: Global hotkey (push-to-talk)
- **Phase 4**: Auto-paste into active window
- **Phase 5**: Advanced processing (LLM cleanup, context-aware)
- **Phase 6**: GUI/tray icon, settings UI
