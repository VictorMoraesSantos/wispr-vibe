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
	STTEngine        string `json:"stt_engine"`
	Provider         string `json:"provider"`
	WhisperAPIKey    string `json:"whisper_api_key"`
	WhisperAPIURL    string `json:"whisper_api_url"`
	WhisperModel     string `json:"whisper_model"`
	MiniMaxGroupID   string `json:"minimax_group_id"`
	WhisperExePath   string `json:"whisper_exe_path"`
	WhisperModelPath string `json:"whisper_model_path"`
	SampleRate       int    `json:"sample_rate"`
	Language         string `json:"language"`
	LogLevel         string `json:"log_level"`
	Hotkey           string `json:"hotkey"`
	RemoveFillers    bool   `json:"remove_fillers"`
	FixPunctuation   bool   `json:"fix_punctuation"`
	AutoPaste        bool   `json:"auto_paste"`
}

func Default() *Config {
	return &Config{
		STTEngine:      "whisper_api",
		WhisperAPIURL:  "https://api.openai.com/v1/audio/transcriptions",
		WhisperModel:   "whisper-1",
		SampleRate:     16000,
		LogLevel:       "info",
		Hotkey:         "Ctrl+Shift+R",
		RemoveFillers:  true,
		FixPunctuation: true,
	}
}

func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".wispr-vibe"), nil
}

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

func Save(cfg *Config, path string) error {
	if path == "" {
		p, err := ConfigPath()
		if err != nil {
			return err
		}
		path = p
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func (c *Config) NeedsSetup() bool {
	return c.WhisperAPIKey == ""
}

func RunSetup(cfg *Config) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("")
	fmt.Println("  wispr-vibe — First Time Setup")
	fmt.Println("")
	fmt.Println("Choose your STT provider:")
	fmt.Println("  [1] OpenAI        (whisper-1, gpt-4o-transcribe)")
	fmt.Println("  [2] MiniMax       (hailuo)")
	fmt.Println("  [3] Local Whisper (100% offline)")
	fmt.Print("→ [1/2/3, default=3]: ")

	providerChoice := readLine(reader)

	switch providerChoice {
	case "2":
		cfg.Provider = "minimax"
		cfg.WhisperModel = "hailuo"

		fmt.Print("\nMiniMax API key: ")
		key := readLine(reader)
		if key == "" {
			return fmt.Errorf("API key cannot be empty")
		}
		cfg.WhisperAPIKey = key

		fmt.Print("MiniMax Group ID: ")
		groupID := readLine(reader)
		if groupID == "" {
			return fmt.Errorf("Group ID cannot be empty")
		}
		cfg.MiniMaxGroupID = groupID
		cfg.WhisperAPIURL = fmt.Sprintf("https://api.minimax.chat/v1/audio/transcriptions?GroupId=%s", groupID)

	case "1":
		cfg.Provider = "openai"
		cfg.WhisperAPIURL = "https://api.openai.com/v1/audio/transcriptions"

		fmt.Print("\nOpenAI API key (sk-...): ")
		key := readLine(reader)
		if key == "" {
			return fmt.Errorf("API key cannot be empty")
		}
		cfg.WhisperAPIKey = key

		fmt.Println("\nModel:")
		fmt.Println("  [1] whisper-1")
		fmt.Println("  [2] gpt-4o-transcribe")
		fmt.Println("  [3] gpt-4o-mini-transcribe")
		fmt.Print("→ [1/2/3, default=1]: ")

		switch readLine(reader) {
		case "2":
			cfg.WhisperModel = "gpt-4o-transcribe"
		case "3":
			cfg.WhisperModel = "gpt-4o-mini-transcribe"
		default:
			cfg.WhisperModel = "whisper-1"
		}

	default:
		cfg.Provider = "local"
		cfg.STTEngine = "whisper_local"
		cfg.WhisperAPIKey = "not-needed"

		fmt.Print("\nWhisper executable path (empty=auto): ")
		cfg.WhisperExePath = readLine(reader)

		fmt.Print("Model .bin path (empty=auto): ")
		cfg.WhisperModelPath = readLine(reader)
	}

	fmt.Print("\nLanguage (pt, en, es... empty=auto): ")
	cfg.Language = readLine(reader)

	if err := Save(cfg, ""); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	path, _ := ConfigPath()
	fmt.Printf("\n✅ Saved: %s (model: %s)\n\n", path, cfg.WhisperModel)
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

func readLine(reader *bufio.Reader) string {
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func orDefault(val, def string) string {
	if val == "" {
		return def
	}
	return val
}
