package stt

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/victorlui/wispr-vibe/pkg/domain"
)

func TestExtractText(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "clean output",
			input:  "Hello world this is a test",
			expect: "Hello world this is a test",
		},
		{
			name:   "skips whisper log lines",
			input:  "whisper_init_from_file: loading model\n[00:00:00.000 --> 00:00:03.000]  Hello world\nwhisper_print_timings: done",
			expect: "",
		},
		{
			name:   "skips empty and bracket lines",
			input:  "\n\n[some timestamp]\nActual transcription text\n\n",
			expect: "Actual transcription text",
		},
		{
			name:   "multiple valid lines joined",
			input:  "whisper_init: model loaded\nFirst line\nSecond line\nwhisper_done",
			expect: "First line Second line",
		},
		{
			name:   "empty input",
			input:  "",
			expect: "",
		},
		{
			name:   "only log lines",
			input:  "whisper_init: loading\n[00:00.000]\nwhisper_print: done\n",
			expect: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractText(tt.input)
			if got != tt.expect {
				t.Errorf("extractText() = %q, want %q", got, tt.expect)
			}
		})
	}
}

func TestWhisperAPIName(t *testing.T) {
	api := NewWhisperAPI("key", "http://example.com", "whisper-1")
	if api.Name() != "whisper_api" {
		t.Errorf("Name() = %q, want %q", api.Name(), "whisper_api")
	}
}

func TestWhisperAPIDefaultURL(t *testing.T) {
	api := NewWhisperAPI("key", "", "")
	if api.apiURL != "https://api.openai.com/v1/audio/transcriptions" {
		t.Errorf("default URL = %q", api.apiURL)
	}
	if api.model != "whisper-1" {
		t.Errorf("default model = %q", api.model)
	}
}

func TestWhisperAPITranscribeSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			t.Errorf("missing Bearer auth header")
		}
		contentType := r.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "multipart/form-data") {
			t.Errorf("expected multipart, got %s", contentType)
		}

		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("parse multipart: %v", err)
		}
		if r.FormValue("model") != "whisper-1" {
			t.Errorf("model = %q, want whisper-1", r.FormValue("model"))
		}
		if r.FormValue("language") != "pt" {
			t.Errorf("language = %q, want pt", r.FormValue("language"))
		}
		if r.FormValue("response_format") != "json" {
			t.Errorf("response_format = %q, want json", r.FormValue("response_format"))
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("no file uploaded: %v", err)
		}
		file.Close()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"text": "Olá mundo este é um teste",
		})
	}))
	defer server.Close()

	api := NewWhisperAPI("test-key", server.URL, "whisper-1")
	result, err := api.Transcribe(context.Background(), []byte("fake-audio-data"), domain.TranscribeOpts{
		Language: "pt",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text != "Olá mundo este é um teste" {
		t.Errorf("text = %q, want %q", result.Text, "Olá mundo este é um teste")
	}
	if result.Language != "pt" {
		t.Errorf("language = %q, want pt", result.Language)
	}
}

func TestWhisperAPITranscribeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": {"message": "Invalid API key"}}`))
	}))
	defer server.Close()

	api := NewWhisperAPI("bad-key", server.URL, "whisper-1")
	_, err := api.Transcribe(context.Background(), []byte("audio"), domain.TranscribeOpts{})

	if err == nil {
		t.Fatal("expected error for 401 response")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error should mention 401: %v", err)
	}
}

func TestWhisperAPITranscribeCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	api := NewWhisperAPI("key", server.URL, "whisper-1")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := api.Transcribe(ctx, []byte("audio"), domain.TranscribeOpts{})
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
}
