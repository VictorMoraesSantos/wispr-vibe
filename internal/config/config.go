package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	// STT engine: "whisper_api" or "whisper_local"
	STTEngine string `json:"stt_engine"`

	// Whisper API settings
	WhisperAPIKey  string `json:"whisper_api_key"`
	WhisperAPIURL  string `json:"whisper_api_url"`
	WhisperModel   string `json:"whisper_model"`

	// Audio settings
	SampleRate int `json:"sample_rate"`

	// Language for transcription ("" = auto-detect)
	Language string `json:"language"`

	// Log level: debug, info, warn, error
	LogLevel string `json:"log_level"`

	// Post-processing
	RemoveFillers   bool `json:"remove_fillers"`
	FixPunctuation  bool `json:"fix_punctuation"`
	AutoPaste       bool `json:"auto_paste"`
}

func Default() *Config {
	return &Config{
		STTEngine:      "whisper_api",
		WhisperAPIURL:  "https://api.openai.com/v1/audio/transcriptions",
		WhisperModel:   "whisper-1",
		SampleRate:     16000,
		Language:       "",
		LogLevel:       "info",
		RemoveFillers:  true,
		FixPunctuation: true,
		AutoPaste:      false,
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()

	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return applyEnvOverrides(cfg), nil
		}
		path = filepath.Join(home, ".wispr-vibe", "config.json")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return applyEnvOverrides(cfg), nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return applyEnvOverrides(cfg), nil
}

func applyEnvOverrides(cfg *Config) *Config {
	if key := os.Getenv("WISPR_API_KEY"); key != "" {
		cfg.WhisperAPIKey = key
	}
	if url := os.Getenv("WISPR_API_URL"); url != "" {
		cfg.WhisperAPIURL = url
	}
	return cfg
}
