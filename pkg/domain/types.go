package domain

import "time"

// TranscribeResult holds the output of a transcription.
type TranscribeResult struct {
	Text      string
	Language  string
	Duration  time.Duration
	CreatedAt time.Time
}

// TranscribeOpts configures a transcription request.
type TranscribeOpts struct {
	Language    string // e.g. "pt", "en", "" for auto-detect
	Prompt      string // context hint for the model
	Temperature float64
}

// RecordingResult holds raw audio data from a recording session.
type RecordingResult struct {
	Data       []byte
	Format     string // "wav"
	SampleRate int
	Duration   time.Duration
}
