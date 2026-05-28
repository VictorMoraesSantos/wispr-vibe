package audio

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// buildFFmpegArgs returns ffmpeg arguments for capturing audio on Windows.
func buildFFmpegArgs(sampleRate int) []string {
	device := detectAudioDevice()
	return []string{
		"-y",
		"-f", "dshow",
		"-i", fmt.Sprintf("audio=%s", device),
		"-ar", fmt.Sprintf("%d", sampleRate),
		"-ac", "1",
		"-acodec", "pcm_s16le",
		"-f", "wav",
		"pipe:1",
	}
}

// detectAudioDevice finds the first available audio input device on Windows.
func detectAudioDevice() string {
	ffmpegPath := findFFmpeg()
	if ffmpegPath == "" {
		return "default"
	}

	// Run ffmpeg -list_devices to find audio inputs
	cmd := exec.Command(ffmpegPath, "-list_devices", "true", "-f", "dshow", "-i", "dummy")
	output, _ := cmd.CombinedOutput()

	// Parse output for audio device names
	// Format: "DeviceName" (audio)
	re := regexp.MustCompile(`"([^"]+)"\s+\(audio\)`)
	matches := re.FindAllStringSubmatch(string(output), -1)

	if len(matches) > 0 {
		return matches[0][1] // first audio device
	}

	return "default"
}

// findFFmpeg locates the ffmpeg executable.
func findFFmpeg() string {
	// 1. Check PATH
	if p, err := exec.LookPath("ffmpeg"); err == nil {
		return p
	}
	if p, err := exec.LookPath("ffmpeg.exe"); err == nil {
		return p
	}

	// 2. Check winget packages directory
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData != "" {
		wingetDir := filepath.Join(localAppData, "Microsoft", "WinGet", "Packages")
		if matches, _ := filepath.Glob(filepath.Join(wingetDir, "Gyan.FFmpeg*", "*", "bin", "ffmpeg.exe")); len(matches) > 0 {
			return matches[0]
		}
		// Broader search
		entries, _ := os.ReadDir(wingetDir)
		for _, e := range entries {
			if strings.Contains(strings.ToLower(e.Name()), "ffmpeg") {
				binPath := filepath.Join(wingetDir, e.Name())
				if m, _ := filepath.Glob(filepath.Join(binPath, "*", "bin", "ffmpeg.exe")); len(m) > 0 {
					return m[0]
				}
			}
		}
	}

	// 3. Common manual install locations
	commonPaths := []string{
		`C:\ffmpeg\bin\ffmpeg.exe`,
		`C:\Program Files\ffmpeg\bin\ffmpeg.exe`,
		`C:\tools\ffmpeg\bin\ffmpeg.exe`,
	}

	home := os.Getenv("USERPROFILE")
	if home != "" {
		commonPaths = append(commonPaths,
			filepath.Join(home, "ffmpeg", "bin", "ffmpeg.exe"),
			filepath.Join(home, "scoop", "shims", "ffmpeg.exe"),
		)
	}

	if choco := os.Getenv("ChocolateyInstall"); choco != "" {
		commonPaths = append(commonPaths, filepath.Join(choco, "bin", "ffmpeg.exe"))
	}

	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return ""
}
