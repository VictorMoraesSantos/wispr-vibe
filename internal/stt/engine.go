package stt

import (
	"context"

	"github.com/victorlui/wispr-vibe/pkg/domain"
)

// Transcriber is the interface all STT engines must implement.
type Transcriber interface {
	Transcribe(ctx context.Context, audio []byte, opts domain.TranscribeOpts) (*domain.TranscribeResult, error)
	Name() string
}
