package audio

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// buildFFmpegArgs returns ffmpeg arguments for capturing audio on Windows.
func buildFFmpegArgs(sampleRate int) []string {
	return []string{
		"-y",
		"-f", "dshow",
		"-i", "audio=default",
		"-ar", fmt.Sprintf("%d", sampleRate),
		"-ac", "1",
		"-acodec", "pcm_s16le",
		"-f", "wav",
		"pipe:1",
	}
}

// findFFmpeg locates the ffmpeg executable.
// Checks PATH first, then common Windows install locations.
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
	}

	// 3. Check common manual install locations
	commonPaths := []string{
		`C:\ffmpeg\bin\ffmpeg.exe`,
		`C:\Program Files\ffmpeg\bin\ffmpeg.exe`,
		`C:\tools\ffmpeg\bin\ffmpeg.exe`,
	}

	// Also check user profile paths
	home := os.Getenv("USERPROFILE")
	if home != "" {
		commonPaths = append(commonPaths,
			filepath.Join(home, "ffmpeg", "bin", "ffmpeg.exe"),
			filepath.Join(home, "Downloads", "ffmpeg", "bin", "ffmpeg.exe"),
		)
	}

	// Check scoop
	if home != "" {
		scoopPath := filepath.Join(home, "scoop", "shims", "ffmpeg.exe")
		commonPaths = append(commonPaths, scoopPath)
	}

	// Check chocolatey
	chocoInstall := os.Getenv("ChocolateyInstall")
	if chocoInstall != "" {
		commonPaths = append(commonPaths, filepath.Join(chocoInstall, "bin", "ffmpeg.exe"))
	}

	for _, p := range commonPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// 4. Last resort: search winget packages with broader pattern
	if localAppData != "" {
		wingetDir := filepath.Join(localAppData, "Microsoft", "WinGet", "Packages")
		_ = filepath.Walk(wingetDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && strings.EqualFold(info.Name(), "ffmpeg.exe") {
				return filepath.SkipAll
			}
			return nil
		})
		// Do a targeted search
		entries, _ := os.ReadDir(wingetDir)
		for _, e := range entries {
			if strings.Contains(strings.ToLower(e.Name()), "ffmpeg") {
				binPath := filepath.Join(wingetDir, e.Name())
				matches, _ := filepath.Glob(filepath.Join(binPath, "*", "bin", "ffmpeg.exe"))
				if len(matches) > 0 {
					return matches[0]
				}
			}
		}
	}

	return ""
}
