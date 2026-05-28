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

type WhisperLocal struct {
	execPath  string
	modelPath string
	useGPU    bool
}

func NewWhisperLocal(execPath, modelPath string, useGPU bool) (*WhisperLocal, error) {
	if execPath == "" {
		execPath = findWhisperExe()
	}
	if execPath == "" {
		return nil, fmt.Errorf("whisper.cpp executable not found")
	}

	if modelPath == "" {
		modelPath = findWhisperModel()
	}
	if modelPath == "" {
		return nil, fmt.Errorf("whisper model not found")
	}

	return &WhisperLocal{execPath: execPath, modelPath: modelPath, useGPU: useGPU}, nil
}

func (w *WhisperLocal) Name() string { return "whisper_local" }

// HasGPUSupport reports whether this whisper-cli binary was compiled with CUDA.
// Detection is done by checking for ggml-cuda.dll next to the binary, which is
// only present in CUDA-enabled builds. Parsing --help is unreliable because newer
// whisper.cpp versions always include --no-gpu in help text even for CPU builds.
func (w *WhisperLocal) HasGPUSupport() bool {
	dir := filepath.Dir(w.execPath)
	cudaDLL := filepath.Join(dir, "ggml-cuda.dll")
	_, err := os.Stat(cudaDLL)
	return err == nil
}

// CheckGPUSupport probes a whisper-cli binary for CUDA support without
// constructing a full WhisperLocal. Returns false if the binary is not found.
func CheckGPUSupport(execPath string) bool {
	if execPath == "" {
		execPath = findWhisperExe()
	}
	if execPath == "" {
		return false
	}
	w := &WhisperLocal{execPath: execPath}
	return w.HasGPUSupport()
}

func (w *WhisperLocal) buildArgs(audioFile, lang string) []string {
	if lang == "" {
		lang = "auto"
	}
	args := []string{
		"-m", w.modelPath,
		"-f", audioFile,
		"--no-timestamps",
		"--output-txt",
	}
	if !w.useGPU {
		args = append(args, "-ng")
	}
	args = append(args, "-l", lang)
	return args
}

func (w *WhisperLocal) Transcribe(ctx context.Context, audio []byte, opts domain.TranscribeOpts) (*domain.TranscribeResult, error) {
	start := time.Now()

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

	args := w.buildArgs(tmpFile.Name(), opts.Language)
	cmd := exec.CommandContext(ctx, w.execPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("whisper.cpp failed: %w\nOutput: %s", err, string(output))
	}

	text := extractText(string(output))

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

func extractText(output string) string {
	var lines []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "whisper_") || strings.HasPrefix(line, "[") {
			continue
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, " ")
}

func findWhisperExe() string {
	// 1. Check PATH first
	for _, name := range []string{"whisper-cli", "whisper-cli.exe", "whisper", "whisper.exe"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}

	// 2. Prefer the binary next to the running executable (standard install layout)
	var appDirLocations []string
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		appDirLocations = []string{
			filepath.Join(dir, "whisper-cli.exe"),
			filepath.Join(dir, "whisper.exe"),
		}
	}
	for _, p := range appDirLocations {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// 3. Legacy / fallback locations
	home, _ := os.UserHomeDir()
	locations := []string{
		filepath.Join(home, ".wispr-vibe", "whisper-cli.exe"),
		filepath.Join(home, ".wispr-vibe", "whisper.exe"),
		filepath.Join(home, ".wispr-vibe", "whisper-bin", "Release", "whisper-cli.exe"),
		filepath.Join(home, ".wispr-vibe", "whisper-bin", "whisper-cli.exe"),
		`C:\whisper\whisper-cli.exe`,
		`C:\whisper.cpp\build\bin\Release\whisper-cli.exe`,
	}
	for _, p := range locations {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return ""
}

func findWhisperModel() string {
	home, _ := os.UserHomeDir()
	locations := []string{
		filepath.Join(home, ".wispr-vibe", "models"),
		filepath.Join(home, ".wispr-vibe"),
		`C:\whisper\models`,
	}

	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		locations = append(locations, filepath.Join(dir, "models"), dir)
	}

	modelNames := []string{
		"ggml-large-v3-turbo.bin",
		"ggml-large-v3.bin",
		"ggml-medium.bin",
		"ggml-small.bin",
		"ggml-base.bin",
		"ggml-medium.en.bin",
		"ggml-small.en.bin",
		"ggml-base.en.bin",
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
