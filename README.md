# Wispr Vibe

Speech-to-text desktop tool for developers — optimized for vibe coding.

Press a hotkey → dictate → text appears at your cursor. No window switching, no typing.

## Features

- **🎙 Global Hotkey** — Record from any app with a configurable shortcut (default: `Ctrl+Shift+R`)
- **🔤 Auto-Type** — Text is pasted directly where your cursor is (IDE, browser, terminal)
- **🖥️ Toast Overlay** — WisprFlow-style floating indicator at bottom of screen while recording
- **🧠 Local Whisper** — 100% offline transcription via whisper.cpp (no API key needed)
- **☁️ Cloud API** — Optional: OpenAI Whisper or MiniMax Hailuo for transcription
- **🧹 Text Cleanup** — Removes fillers (uh, um, tipo), fixes punctuation and spacing
- **🔧 GUI + System Tray** — Fyne-based window, minimizes to tray, never interrupts your work

## Quick Start

### Prerequisites

- **Windows 10/11** (primary platform)
- **Go 1.21+** with CGo enabled
- **GCC** (WinLibs/MinGW — for Fyne GUI compilation)
- **FFmpeg** installed (for audio capture)

### Install & Build

```bash
git clone https://github.com/victorlui/wispr-vibe.git
cd wispr-vibe

# Build GUI version (recommended)
set CGO_ENABLED=1
go build -o vibevoice-gui.exe ./cmd/vibevoice-gui/

# Build CLI version (no GUI dependency)
go build -o vibevoice.exe ./cmd/vibevoice/
```

### First Run

```bash
.\vibevoice-gui.exe
```

On first run, the setup wizard asks for your STT provider:

| Option | Description |
|--------|-------------|
| **Local Whisper** (default) | 100% offline, free. Needs whisper.cpp + model (~465MB) |
| OpenAI | Cloud API, needs API key (`sk-...`) |
| MiniMax | Cloud API, needs API key + Group ID |

### Usage

1. **Start the app** — GUI opens, minimizes to system tray
2. **Press your hotkey** (default: `Ctrl+Shift+R`) from **any application**
3. **Speak** — toast overlay shows "🎙 Recording... 3s" at bottom of screen
4. **Press hotkey again** — audio is transcribed and text is typed at your cursor
5. **Continue working** — window stays hidden, no interruption

### Configure Hotkey

Open Settings (⚙️ icon) → type your preferred combo:

```
Ctrl+Shift+R    (default)
Alt+Z
Ctrl+F9
Win+Shift+V
```

Supports: `Ctrl`, `Shift`, `Alt`, `Win` + any letter, number, F-key, or special key.

## Configuration

Config stored at `~/.wispr-vibe/config.json`:

```json
{
  "stt_engine": "whisper_local",
  "provider": "local",
  "whisper_model_path": "C:\\Users\\You\\.wispr-vibe\\models\\ggml-small.bin",
  "language": "pt",
  "hotkey": "Ctrl+Shift+R",
  "sample_rate": 16000,
  "log_level": "info",
  "remove_fillers": true,
  "fix_punctuation": true
}
```

Environment variables: `WISPR_API_KEY`, `WISPR_API_URL`.

## Whisper Models (Local)

Download models from [huggingface.co/ggerganov/whisper.cpp](https://huggingface.co/ggerganov/whisper.cpp/tree/main):

| Model | Size | Quality | Speed |
|-------|------|---------|-------|
| ggml-base.bin | 141 MB | Basic | Fast |
| **ggml-small.bin** | 465 MB | **Good (recommended)** | ~4s |
| ggml-medium.bin | 1.5 GB | Great | ~8s |
| ggml-large-v3.bin | 3.1 GB | Best | ~15s |

Place in `~/.wispr-vibe/models/`. The app auto-detects the best available model.

## GPU Acceleration (CUDA)

The `Use GPU acceleration` toggle in Settings only works if your `whisper-cli.exe` was compiled with CUDA support. A standard pre-built binary **does not have GPU support** — the flag will appear disabled in Settings.

### How to get a CUDA-enabled whisper-cli

**Option 1 — Auto (recommended):** use the included build script:

```powershell
# Requires: CUDA Toolkit, CMake, Git
.\build.ps1
```

This clones `whisper.cpp`, compiles it with `-DGGML_CUDA=ON`, downloads the model, and places everything in `dist\bin\`.

**Option 2 — Manual:**

```powershell
git clone --depth 1 https://github.com/ggerganov/whisper.cpp.git
cd whisper.cpp
cmake -B build -DCMAKE_BUILD_TYPE=Release -DGGML_CUDA=ON
cmake --build build --config Release --parallel

# Copy the binary
copy build\bin\Release\whisper-cli.exe %USERPROFILE%\.wispr-vibe\
```

**Option 3 — CPU only (no CUDA needed):**

Leave `Use GPU acceleration` unchecked. Whisper still runs well on CPU — a `ggml-small` model takes ~4-6s per transcription.

### Verify GPU support

In Settings, the GPU section shows:
- **"whisper-cli has CUDA support"** — GPU is available, toggle works
- **"NOT compiled with CUDA"** — need to recompile with `build.ps1`

```
Hotkey → Audio Capture → STT Engine → Text Processor → Clipboard + Paste
  ↓                                                         ↓
Toast Overlay                                    SendInput(Ctrl+V)
```

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for details.

## Project Status

- [x] Audio capture via FFmpeg (auto-detect mic device)
- [x] Local Whisper transcription (whisper.cpp, offline)
- [x] Cloud API support (OpenAI, MiniMax)
- [x] Text post-processing (fillers, punctuation, spacing)
- [x] Global hotkey (configurable, Win32 RegisterHotKey)
- [x] Auto-type into active window (Win32 SendInput)
- [x] GUI with Fyne (Record button, Settings, Timer)
- [x] System tray (minimize on close, tray menu)
- [x] Toast overlay (native Win32, always-on-top, no focus steal)
- [x] FFmpeg auto-discovery (winget, scoop, choco paths)
- [x] Interactive setup wizard
- [ ] Global hotkey on Linux/macOS
- [ ] LLM-powered text cleanup
- [ ] Multi-language auto-detect
- [ ] Audio waveform visualization

## License

MIT
