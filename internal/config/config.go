package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// ConfigDir returns the path to the config directory.
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".wispr-vibe"), nil
}

// ConfigPath returns the path to the config file.
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func Load(path string) (*Config, error) {
	cfg := Default()

	if path == "" {
		p, err := ConfigPath()
		if err != nil {
			return applyEnvOverrides(cfg), nil
		}
		path = p
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

// Save writes the config to disk.
func Save(cfg *Config, path string) error {
	if path == "" {
		p, err := ConfigPath()
		if err != nil {
			return err
		}
		path = p
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// NeedsSetup returns true if there's no API key configured.
func (c *Config) NeedsSetup() bool {
	return c.WhisperAPIKey == ""
}

// RunSetup asks the user for API key and model interactively.
func RunSetup(cfg *Config) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║      wispr-vibe — First Time Setup       ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()

	// API Key
	fmt.Println("Enter your OpenAI API key (starts with sk-):")
	fmt.Print("→ ")
	key, _ := reader.ReadString('\n')
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("API key cannot be empty")
	}
	cfg.WhisperAPIKey = key

	// Model selection
	fmt.Println()
	fmt.Println("Choose transcription model:")
	fmt.Println("  [1] whisper-1       (fast, cheap, good for most cases)")
	fmt.Println("  [2] gpt-4o-transcribe  (best accuracy, more expensive)")
	fmt.Println("  [3] gpt-4o-mini-transcribe (good balance)")
	fmt.Print("→ [1/2/3, default=1]: ")
	modelChoice, _ := reader.ReadString('\n')
	modelChoice = strings.TrimSpace(modelChoice)

	switch modelChoice {
	case "2":
		cfg.WhisperModel = "gpt-4o-transcribe"
	case "3":
		cfg.WhisperModel = "gpt-4o-mini-transcribe"
	default:
		cfg.WhisperModel = "whisper-1"
	}

	// Language
	fmt.Println()
	fmt.Println("Preferred language (leave empty for auto-detect):")
	fmt.Println("  Examples: pt, en, es, fr")
	fmt.Print("→ [default=auto]: ")
	lang, _ := reader.ReadString('\n')
	lang = strings.TrimSpace(lang)
	cfg.Language = lang

	// Save
	if err := Save(cfg, ""); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	path, _ := ConfigPath()
	fmt.Printf("\n✅ Config saved to: %s\n", path)
	fmt.Printf("   Model: %s\n", cfg.WhisperModel)
	fmt.Printf("   Language: %s\n", orDefault(cfg.Language, "auto-detect"))
	fmt.Println()

	return nil
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

func orDefault(val, def string) string {
	if val == "" {
		return def
	}
	return val
}
