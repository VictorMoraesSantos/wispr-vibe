# Wispr Vibe — Product Specification

## Vision

Desktop speech-to-text tool for developers, optimized for vibe coding.
Press a hotkey → dictate → text appears at your cursor in any application.
No window switching, no typing, no interruption.

## Core Requirements

### Functional

1. ✅ Global hotkey (configurable) starts/stops recording from any app
2. ✅ Capture microphone audio via FFmpeg (auto-detect device on Windows)
3. ✅ Transcribe via local Whisper (offline, free) or cloud API (OpenAI/MiniMax)
4. ✅ Post-process transcription:
   - Remove filler words (uh, um, tipo, né, então)
   - Fix spacing and punctuation
   - Capitalize sentences
   - Clean verbal auxiliary commands
5. ✅ Auto-paste result into active window at cursor position (SendInput Ctrl+V)
6. ✅ Toast overlay at bottom of screen while recording (WisprFlow-style)
7. ✅ GUI with system tray (minimize, never interrupts workflow)
8. ✅ Structured logging for debug (slog)
9. ✅ Interactive setup wizard for first-time configuration

### Non-Functional

- ✅ Low latency (~3-5s for local Whisper with ggml-small model)
- ✅ Modular architecture — swap STT backends without touching other code
- ✅ Offline-first (local Whisper as primary path)
- ✅ Windows-first (Win32 native APIs for hotkey, clipboard, toast)
- ✅ Minimal resource footprint when idle (no polling, event-driven)
- Cross-platform stubs for future Linux/macOS support

## Supported STT Engines

| Engine | Type | Key Required | Quality |
|--------|------|-------------|---------|
| **whisper.cpp (local)** | Offline | No | Depends on model size |
| OpenAI Whisper API | Cloud | Yes (sk-...) | High |
| MiniMax Hailuo | Cloud | Yes + GroupID | High |

## Architecture

```
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌───────────┐    ┌───────────┐
│  Hotkey  │───▶│  Audio   │───▶│   STT    │───▶│ Processor │───▶│ Auto-Type │
│ (Global) │    │ (FFmpeg) │    │ (Engine) │    │ (Pipeline)│    │(SendInput)│
└──────────┘    └──────────┘    └──────────┘    └───────────┘    └───────────┘
     │                                │                                │
     ▼                          ┌─────┴─────┐                         ▼
┌──────────┐               │ whisper.cpp │               ┌───────────┐
│  Toast   │               │  or Cloud   │               │  Cursor   │
│ Overlay  │               └───────────┘               │ Position  │
└──────────┘                                              └───────────┘
```

## Modules

| Module     | Responsibility                                      | Status |
|-----------|------------------------------------------------------|--------|
| audio     | FFmpeg subprocess, WAV capture, device detection      | ✅ Done |
| stt       | Transcriber interface + WhisperLocal + WhisperAPI     | ✅ Done |
| processor | Text cleanup pipeline (fillers, punctuation, spaces)  | ✅ Done |
| hotkey    | Win32 RegisterHotKey, configurable combo parsing      | ✅ Done |
| clipboard | Win32 clipboard + SendInput (Ctrl+V) auto-paste      | ✅ Done |
| toast     | Native Win32 popup overlay (recording indicator)      | ✅ Done |
| config    | JSON config, setup wizard, env overrides              | ✅ Done |
| logger    | Structured logging with slog                          | ✅ Done |
| app       | Pipeline orchestrator (record → STT → process → type)| ✅ Done |

## Completed Phases

- **Phase 1**: MVP — record, transcribe (API), clean, clipboard ✅
- **Phase 2**: Local Whisper via whisper.cpp (100% offline) ✅
- **Phase 3**: Global hotkey (push-to-talk, configurable) ✅
- **Phase 4**: Auto-paste into active window (Win32 SendInput) ✅
- **Phase 5**: GUI + System Tray (Fyne) ✅
- **Phase 6**: Toast overlay (native Win32, WisprFlow-style) ✅

## Future Phases

- **Phase 7**: LLM-powered text cleanup (GPT context-aware rewriting)
- **Phase 8**: Linux/macOS support (PulseAudio, xdotool, osascript)
- **Phase 9**: Audio waveform visualization in toast
- **Phase 10**: Multiple language auto-detection
- **Phase 11**: Voice commands ("new line", "delete that", "undo")
