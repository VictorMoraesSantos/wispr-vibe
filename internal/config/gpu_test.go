package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// utf8BOM is the three-byte sequence prepended by PowerShell Set-Content -Encoding UTF8.
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// TestLoadStripsUTF8BOM verifies that a config file written with a UTF-8 BOM
// (as PowerShell does) is parsed correctly instead of crashing with
// "invalid character" — the bug that caused the app to exit immediately on startup.
func TestLoadStripsUTF8BOM(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	cfg := Default()
	cfg.WhisperAPIKey = "not-needed"
	cfg.STTEngine = "whisper_local"
	cfg.Provider = "local"
	cfg.UseGPU = true
	cfg.WhisperExePath = `C:\dist\bin\whisper-cli.exe`

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	// Prepend BOM exactly as PowerShell does
	withBOM := append(utf8BOM, data...)
	if err := os.WriteFile(cfgPath, withBOM, 0600); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load should succeed with BOM-prefixed JSON, got: %v", err)
	}

	if loaded.STTEngine != "whisper_local" {
		t.Errorf("STTEngine = %q, want %q", loaded.STTEngine, "whisper_local")
	}
	if !loaded.UseGPU {
		t.Error("UseGPU should be true after BOM-stripped parse")
	}
	if loaded.WhisperExePath != `C:\dist\bin\whisper-cli.exe` {
		t.Errorf("WhisperExePath = %q, want path preserved", loaded.WhisperExePath)
	}
}

// TestLoadBOMWithInvalidJSON ensures that BOM stripping does not hide parse
// errors: a BOM followed by garbage must still return an error.
func TestLoadBOMWithInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	garbage := append(utf8BOM, []byte(`{not: valid}`)...)
	os.WriteFile(cfgPath, garbage, 0600)

	_, err := Load(cfgPath)
	if err == nil {
		t.Fatal("expected JSON parse error for BOM + invalid JSON, got nil")
	}
}

// TestSaveProducesNoBOM verifies that Save writes valid JSON without a BOM so
// the file can be re-read by any JSON parser without special handling.
func TestSaveProducesNoBOM(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	cfg := Default()
	cfg.WhisperAPIKey = "key"
	if err := Save(cfg, cfgPath); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(data) < 3 {
		t.Fatal("config file is too short")
	}
	if data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		t.Error("Save must not write a UTF-8 BOM; Go's json package cannot parse it")
	}
	if data[0] != '{' {
		t.Errorf("first byte of config should be '{', got 0x%02X", data[0])
	}
}

// TestUseGPUTrueRoundTrip verifies that UseGPU=true is written to disk and
// read back correctly — the flag that controls whether -ng is passed to whisper-cli.
func TestUseGPUTrueRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	cfg := Default()
	cfg.WhisperAPIKey = "key"
	cfg.UseGPU = true

	if err := Save(cfg, cfgPath); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !loaded.UseGPU {
		t.Error("UseGPU=true should survive a save/load cycle")
	}
}

// TestUseGPUFalseRoundTrip mirrors the above for UseGPU=false (CPU-only mode).
func TestUseGPUFalseRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	cfg := Default()
	cfg.WhisperAPIKey = "key"
	cfg.UseGPU = false

	if err := Save(cfg, cfgPath); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.UseGPU {
		t.Error("UseGPU=false should survive a save/load cycle")
	}
}
