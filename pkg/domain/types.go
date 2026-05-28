package domain

import "time"

type TranscribeResult struct {
	Text      string
	Language  string
	Duration  time.Duration
	CreatedAt time.Time
}

type TranscribeOpts struct {
	Language    string
	Prompt      string
	Temperature float64
}

type RecordingResult struct {
	Data       []byte
	Format     string
	SampleRate int
	Duration   time.Duration
}
