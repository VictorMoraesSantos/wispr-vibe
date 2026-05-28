package config

import (
	"os"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.STTEngine != "whisper_api" {
		t.Errorf("expected whisper_api, got %s", cfg.STTEngine)
	}
	if cfg.SampleRate != 16000 {
		t.Errorf("expected 16000, got %d", cfg.SampleRate)
	}
	if !cfg.RemoveFillers {
		t.Error("expected RemoveFillers=true")
	}
}

func TestLoadNonExistent(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return defaults
	if cfg.STTEngine != "whisper_api" {
		t.Errorf("expected default config, got engine=%s", cfg.STTEngine)
	}
}

func TestLoadEnvOverride(t *testing.T) {
	os.Setenv("WISPR_API_KEY", "test-key-123")
	defer os.Unsetenv("WISPR_API_KEY")

	cfg, err := Load("/nonexistent/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.WhisperAPIKey != "test-key-123" {
		t.Errorf("expected env override, got %s", cfg.WhisperAPIKey)
	}
}
