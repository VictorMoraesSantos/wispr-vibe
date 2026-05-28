package stt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/victorlui/wispr-vibe/pkg/domain"
)

// WhisperAPI implements Transcriber using OpenAI-compatible Whisper API.
type WhisperAPI struct {
	apiKey string
	apiURL string
	model  string
	client *http.Client
}

func NewWhisperAPI(apiKey, apiURL, model string) *WhisperAPI {
	if apiURL == "" {
		apiURL = "https://api.openai.com/v1/audio/transcriptions"
	}
	if model == "" {
		model = "whisper-1"
	}
	return &WhisperAPI{
		apiKey: apiKey,
		apiURL: apiURL,
		model:  model,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (w *WhisperAPI) Name() string {
	return "whisper_api"
}

func (w *WhisperAPI) Transcribe(ctx context.Context, audio []byte, opts domain.TranscribeOpts) (*domain.TranscribeResult, error) {
	start := time.Now()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "audio.wav")
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := part.Write(audio); err != nil {
		return nil, fmt.Errorf("write audio: %w", err)
	}

	writer.WriteField("model", w.model)
	if opts.Language != "" {
		writer.WriteField("language", opts.Language)
	}
	if opts.Prompt != "" {
		writer.WriteField("prompt", opts.Prompt)
	}
	if opts.Temperature > 0 {
		writer.WriteField("temperature", fmt.Sprintf("%.2f", opts.Temperature))
	}
	writer.WriteField("response_format", "json")
	writer.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", w.apiURL, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+w.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := w.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &domain.TranscribeResult{
		Text:      result.Text,
		Language:  opts.Language,
		Duration:  time.Since(start),
		CreatedAt: time.Now(),
	}, nil
}
