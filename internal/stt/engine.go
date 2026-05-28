package stt

import (
	"context"

	"github.com/victorlui/wispr-vibe/pkg/domain"
)

type Transcriber interface {
	Transcribe(ctx context.Context, audio []byte, opts domain.TranscribeOpts) (*domain.TranscribeResult, error)
	Name() string
}
