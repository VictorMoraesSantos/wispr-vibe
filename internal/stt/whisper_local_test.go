package stt

import (
	"os"
	"slices"
	"testing"
)

func TestWhisperLocalName(t *testing.T) {
	w := &WhisperLocal{execPath: "whisper-cli", modelPath: "model.bin", useGPU: true}
	if w.Name() != "whisper_local" {
		t.Errorf("Name() = %q, want %q", w.Name(), "whisper_local")
	}
}

func TestNewWhisperLocalWithExplicitPaths(t *testing.T) {
	tmpDir := t.TempDir()

	exeFile, err := os.CreateTemp(tmpDir, "fake-whisper*.exe")
	if err != nil {
		t.Fatal(err)
	}
	exeFile.Close()

	modelFile, err := os.CreateTemp(tmpDir, "ggml-model*.bin")
	if err != nil {
		t.Fatal(err)
	}
	modelFile.Close()

	w, err := NewWhisperLocal(exeFile.Name(), modelFile.Name(), true)
	if err != nil {
		t.Fatalf("NewWhisperLocal with explicit paths: %v", err)
	}
	if w.execPath != exeFile.Name() {
		t.Errorf("execPath = %q, want %q", w.execPath, exeFile.Name())
	}
	if w.modelPath != modelFile.Name() {
		t.Errorf("modelPath = %q, want %q", w.modelPath, modelFile.Name())
	}
	if !w.useGPU {
		t.Error("useGPU should be true")
	}
}

func TestNewWhisperLocalGPUFlagStored(t *testing.T) {
	tmpDir := t.TempDir()
	exe, _ := os.CreateTemp(tmpDir, "whisper*.exe")
	exe.Close()
	model, _ := os.CreateTemp(tmpDir, "model*.bin")
	model.Close()

	wGPU, _ := NewWhisperLocal(exe.Name(), model.Name(), true)
	wCPU, _ := NewWhisperLocal(exe.Name(), model.Name(), false)

	if !wGPU.useGPU {
		t.Error("useGPU=true should be stored")
	}
	if wCPU.useGPU {
		t.Error("useGPU=false should be stored")
	}
}

func TestNewWhisperLocalMissingExe(t *testing.T) {
	if findWhisperExe() != "" {
		t.Skip("whisper.cpp is installed on this system; skipping missing-exe test")
	}
	_, err := NewWhisperLocal("", "", false)
	if err == nil {
		t.Fatal("expected error when whisper executable not found")
	}
}

func TestNewWhisperLocalMissingModel(t *testing.T) {
	if findWhisperModel() != "" {
		t.Skip("whisper model found on this system; skipping missing-model test")
	}

	tmpDir := t.TempDir()
	exe, _ := os.CreateTemp(tmpDir, "fake-whisper*.exe")
	exe.Close()

	_, err := NewWhisperLocal(exe.Name(), "", false)
	if err == nil {
		t.Fatal("expected error when whisper model not found")
	}
}

func TestBuildArgsGPUEnabled(t *testing.T) {
	w := &WhisperLocal{execPath: "whisper-cli", modelPath: "model.bin", useGPU: true}
	args := w.buildArgs("audio.wav", "en")

	if slices.Contains(args, "-ng") {
		t.Error("GPU enabled: args should NOT contain -ng")
	}
}

func TestBuildArgsGPUDisabled(t *testing.T) {
	w := &WhisperLocal{execPath: "whisper-cli", modelPath: "model.bin", useGPU: false}
	args := w.buildArgs("audio.wav", "en")

	if !slices.Contains(args, "-ng") {
		t.Error("GPU disabled: args should contain -ng")
	}
}

func TestBuildArgsLanguageDefault(t *testing.T) {
	w := &WhisperLocal{execPath: "whisper-cli", modelPath: "model.bin", useGPU: true}
	args := w.buildArgs("audio.wav", "")

	idx := slices.Index(args, "-l")
	if idx == -1 {
		t.Fatal("args should contain -l flag")
	}
	if idx+1 >= len(args) || args[idx+1] != "auto" {
		t.Errorf("empty language should resolve to 'auto', got: %v", args[idx+1:])
	}
}

func TestBuildArgsLanguageExplicit(t *testing.T) {
	w := &WhisperLocal{execPath: "whisper-cli", modelPath: "model.bin", useGPU: true}
	args := w.buildArgs("audio.wav", "pt")

	idx := slices.Index(args, "-l")
	if idx == -1 {
		t.Fatal("args should contain -l flag")
	}
	if idx+1 >= len(args) || args[idx+1] != "pt" {
		t.Errorf("explicit language 'pt' not set, got: %v", args[idx+1:])
	}
}

func TestBuildArgsRequiredFlags(t *testing.T) {
	w := &WhisperLocal{execPath: "whisper-cli", modelPath: "/path/to/model.bin", useGPU: true}
	args := w.buildArgs("/tmp/audio.wav", "en")

	required := []struct {
		flag  string
		value string
	}{
		{"-m", "/path/to/model.bin"},
		{"-f", "/tmp/audio.wav"},
	}

	for _, req := range required {
		idx := slices.Index(args, req.flag)
		if idx == -1 {
			t.Errorf("args missing %s flag", req.flag)
			continue
		}
		if idx+1 >= len(args) || args[idx+1] != req.value {
			t.Errorf("%s value = %q, want %q", req.flag, args[idx+1], req.value)
		}
	}

	if !slices.Contains(args, "--no-timestamps") {
		t.Error("args should contain --no-timestamps")
	}
	if !slices.Contains(args, "--output-txt") {
		t.Error("args should contain --output-txt")
	}
}

func TestBuildArgsModelAndAudioFile(t *testing.T) {
	w := &WhisperLocal{execPath: "x", modelPath: "my-model.bin", useGPU: true}
	args := w.buildArgs("my-audio.wav", "fr")

	mIdx := slices.Index(args, "-m")
	if mIdx == -1 || args[mIdx+1] != "my-model.bin" {
		t.Errorf("model path not set correctly in args: %v", args)
	}

	fIdx := slices.Index(args, "-f")
	if fIdx == -1 || args[fIdx+1] != "my-audio.wav" {
		t.Errorf("audio file not set correctly in args: %v", args)
	}
}
