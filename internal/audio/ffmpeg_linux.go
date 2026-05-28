//go:build linux

package audio

import (
	"fmt"
	"os/exec"
)

// buildFFmpegArgs returns ffmpeg arguments for capturing audio on Linux.
func buildFFmpegArgs(sampleRate int) []string {
	return []string{
		"-y",
		"-f", "pulse",
		"-i", "default",
		"-ar", fmt.Sprintf("%d", sampleRate),
		"-ac", "1",
		"-acodec", "pcm_s16le",
		"-f", "wav",
		"pipe:1",
	}
}

func findFFmpeg() string {
	if p, err := exec.LookPath("ffmpeg"); err == nil {
		return p
	}
	return ""
}
