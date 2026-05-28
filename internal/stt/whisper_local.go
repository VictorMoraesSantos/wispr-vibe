package stt

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/victorlui/wispr-vibe/pkg/domain"
)

// WhisperLocal implements Transcriber using whisper.cpp running locally.
// No API key needed — runs 100% offline.
type WhisperLocal struct {
	execPath  string // path to whisper.cpp executable
	modelPath string // path to ggml model file
}

func NewWhisperLocal(execPath, modelPath string) (*WhisperLocal, error) {
	if execPath == "" {
		execPath = findWhisperExe()
	}
	if execPath == "" {
		return nil, fmt.Errorf("whisper.cpp executable not found. Download from: https://github.com/ggerganov/whisper.cpp/releases")
	}

	if modelPath == "" {
		modelPath = findWhisperModel()
	}
	if modelPath == "" {
		return nil, fmt.Errorf("whisper model not found. Download a .bin model from: https://huggingface.co/ggerganov/whisper.cpp")
	}

	return &WhisperLocal{
		execPath:  execPath,
		modelPath: modelPath,
	}, nil
}

func (w *WhisperLocal) Name() string {
	return "whisper_local"
}

func (w *WhisperLocal) Transcribe(ctx context.Context, audio []byte, opts domain.TranscribeOpts) (*domain.TranscribeResult, error) {
	start := time.Now()

	// Write audio to temp file (whisper.cpp needs a file)
	tmpFile, err := os.CreateTemp("", "wispr-audio-*.wav")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(audio); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("write audio: %w", err)
	}
	tmpFile.Close()

	// Build whisper.cpp command
	args := []string{
		"-m", w.modelPath,
		"-f", tmpFile.Name(),
		"--no-timestamps",
		"--output-txt",
	}

	if opts.Language != "" {
		args = append(args, "-l", opts.Language)
	} else {
		args = append(args, "-l", "auto")
	}

	cmd := exec.CommandContext(ctx, w.execPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("whisper.cpp failed: %w\nOutput: %s", err, string(output))
	}

	// Extract text from output (whisper.cpp prints to stdout)
	text := extractText(string(output))

	// Also check for .txt file output
	txtFile := tmpFile.Name() + ".txt"
	if data, err := os.ReadFile(txtFile); err == nil {
		os.Remove(txtFile)
		if t := strings.TrimSpace(string(data)); t != "" {
			text = t
		}
	}

	return &domain.TranscribeResult{
		Text:      text,
		Language:  opts.Language,
		Duration:  time.Since(start),
		CreatedAt: time.Now(),
	}, nil
}

// extractText cleans whisper.cpp stdout output.
func extractText(output string) string {
	var lines []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		// Skip whisper.cpp log lines (start with timestamps or system info)
		if line == "" || strings.HasPrefix(line, "whisper_") || strings.HasPrefix(line, "[") {
			continue
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, " ")
}

// findWhisperExe looks for whisper.cpp executable in common locations.
func findWhisperExe() string {
	names := []string{"whisper-cli", "whisper-cli.exe", "whisper", "whisper.exe", "main.exe"}

	// Check PATH
	for _, name := range names {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}

	// Check common locations
	home, _ := os.UserHomeDir()
	locations := []string{
		filepath.Join(home, ".wispr-vibe", "whisper-bin", "Release", "whisper-cli.exe"),
		filepath.Join(home, ".wispr-vibe", "whisper-bin", "whisper-cli.exe"),
		filepath.Join(home, ".wispr-vibe", "whisper-cli.exe"),
		filepath.Join(home, ".wispr-vibe", "whisper.exe"),
		`C:\whisper\whisper-cli.exe`,
		`C:\whisper.cpp\build\bin\Release\whisper-cli.exe`,
	}

	// Check next to our own executable
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		locations = append(locations,
			filepath.Join(dir, "whisper-cli.exe"),
			filepath.Join(dir, "whisper.exe"),
		)
	}

	for _, p := range locations {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return ""
}

// findWhisperModel looks for a ggml model file.
func findWhisperModel() string {
	home, _ := os.UserHomeDir()
	locations := []string{
		filepath.Join(home, ".wispr-vibe", "models"),
		filepath.Join(home, ".wispr-vibe"),
		`C:\whisper\models`,
	}

	// Check next to our executable
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		locations = append(locations, filepath.Join(dir, "models"), dir)
	}

	// Look for any .bin model file
	modelNames := []string{
		"ggml-base.bin",
		"ggml-base.en.bin",
		"ggml-small.bin",
		"ggml-small.en.bin",
		"ggml-medium.bin",
		"ggml-medium.en.bin",
		"ggml-large-v3.bin",
		"ggml-large-v3-turbo.bin",
	}

	for _, loc := range locations {
		for _, model := range modelNames {
			p := filepath.Join(loc, model)
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}

	return ""
}
