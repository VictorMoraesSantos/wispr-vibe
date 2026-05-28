package audio

import "fmt"

// buildFFmpegArgs returns ffmpeg arguments for capturing audio on Windows.
// Uses DirectShow (dshow) to capture from default audio device.
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
