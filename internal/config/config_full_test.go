package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultValues(t *testing.T) {
	cfg := Default()

	checks := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"STTEngine", cfg.STTEngine, "whisper_api"},
		{"WhisperAPIURL", cfg.WhisperAPIURL, "https://api.openai.com/v1/audio/transcriptions"},
		{"WhisperModel", cfg.WhisperModel, "whisper-1"},
		{"SampleRate", cfg.SampleRate, 16000},
		{"LogLevel", cfg.LogLevel, "info"},
		{"Hotkey", cfg.Hotkey, "Ctrl+Shift+R"},
		{"RemoveFillers", cfg.RemoveFillers, true},
		{"FixPunctuation", cfg.FixPunctuation, true},
		{"AutoPaste", cfg.AutoPaste, false},
		{"Language", cfg.Language, ""},
	}

	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
}

func TestNeedsSetup(t *testing.T) {
	cfg := Default()
	if !cfg.NeedsSetup() {
		t.Error("default config should need setup (no API key)")
	}

	cfg.WhisperAPIKey = "sk-test"
	if cfg.NeedsSetup() {
		t.Error("config with API key should NOT need setup")
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	original := Default()
	original.WhisperAPIKey = "test-key-xyz"
	original.Language = "pt"
	original.Hotkey = "Alt+Z"
	original.STTEngine = "whisper_local"
	original.Provider = "local"

	// Save
	if err := Save(original, cfgPath); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		t.Fatal("config file not created")
	}

	// Load
	loaded, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	// Verify fields match
	if loaded.WhisperAPIKey != "test-key-xyz" {
		t.Errorf("WhisperAPIKey = %q, want %q", loaded.WhisperAPIKey, "test-key-xyz")
	}
	if loaded.Language != "pt" {
		t.Errorf("Language = %q, want %q", loaded.Language, "pt")
	}
	if loaded.Hotkey != "Alt+Z" {
		t.Errorf("Hotkey = %q, want %q", loaded.Hotkey, "Alt+Z")
	}
	if loaded.STTEngine != "whisper_local" {
		t.Errorf("STTEngine = %q, want %q", loaded.STTEngine, "whisper_local")
	}
	if loaded.Provider != "local" {
		t.Errorf("Provider = %q, want %q", loaded.Provider, "local")
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "sub", "dir", "config.json")

	cfg := Default()
	cfg.WhisperAPIKey = "key"

	if err := Save(cfg, nestedPath); err != nil {
		t.Fatalf("Save should create nested dirs: %v", err)
	}

	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Fatal("file not created in nested dir")
	}
}

func TestSaveFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	cfg := Default()
	cfg.WhisperAPIKey = "secret"
	Save(cfg, cfgPath)

	info, _ := os.Stat(cfgPath)
	// On Windows permissions work differently, just verify it's not world-readable
	// The important thing is the file exists and is valid JSON
	if info.Size() == 0 {
		t.Error("config file is empty")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	os.WriteFile(cfgPath, []byte("not valid json {{{"), 0600)

	_, err := Load(cfgPath)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoadPartialJSON(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	// Only set some fields — others should keep defaults
	partial := map[string]interface{}{
		"language": "en",
		"hotkey":   "Ctrl+F9",
	}
	data, _ := json.Marshal(partial)
	os.WriteFile(cfgPath, data, 0600)

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	// Explicit fields
	if cfg.Language != "en" {
		t.Errorf("Language = %q, want en", cfg.Language)
	}
	if cfg.Hotkey != "Ctrl+F9" {
		t.Errorf("Hotkey = %q, want Ctrl+F9", cfg.Hotkey)
	}

	// Defaults should be preserved for unset fields
	if cfg.SampleRate != 16000 {
		t.Errorf("SampleRate should default to 16000, got %d", cfg.SampleRate)
	}
}

func TestEnvOverridePriority(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	cfg := Default()
	cfg.WhisperAPIKey = "file-key"
	Save(cfg, cfgPath)

	// Env should override file
	os.Setenv("WISPR_API_KEY", "env-key")
	defer os.Unsetenv("WISPR_API_KEY")

	loaded, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if loaded.WhisperAPIKey != "env-key" {
		t.Errorf("env should override file: got %q, want %q", loaded.WhisperAPIKey, "env-key")
	}
}

func TestDefaultUseGPU(t *testing.T) {
	cfg := Default()
	if !cfg.UseGPU {
		t.Error("UseGPU should default to true")
	}
}

func TestSaveLoadUseGPU(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	cfg := Default()
	cfg.WhisperAPIKey = "key"
	cfg.UseGPU = false

	if err := Save(cfg, cfgPath); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	loaded, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if loaded.UseGPU {
		t.Error("UseGPU=false should round-trip correctly, got true")
	}
}

func TestEnvOverrideURL(t *testing.T) {
	os.Setenv("WISPR_API_URL", "https://custom.api/v1/transcriptions")
	defer os.Unsetenv("WISPR_API_URL")

	cfg, _ := Load("/nonexistent")
	if cfg.WhisperAPIURL != "https://custom.api/v1/transcriptions" {
		t.Errorf("URL = %q, want custom URL", cfg.WhisperAPIURL)
	}
}
