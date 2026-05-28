package audio

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestEncodeWAV(t *testing.T) {
	samples := []int16{0, 100, -100, 32767, -32768}
	sampleRate := 16000

	wav, err := encodeWAV(samples, sampleRate)
	if err != nil {
		t.Fatalf("encodeWAV error: %v", err)
	}

	// WAV header is 44 bytes + data
	expectedDataSize := len(samples) * 2
	expectedFileSize := 44 + expectedDataSize

	if len(wav) != expectedFileSize {
		t.Errorf("WAV length = %d, want %d", len(wav), expectedFileSize)
	}

	// Check RIFF header
	if string(wav[0:4]) != "RIFF" {
		t.Errorf("missing RIFF header, got %q", string(wav[0:4]))
	}

	// Check WAVE format
	if string(wav[8:12]) != "WAVE" {
		t.Errorf("missing WAVE format, got %q", string(wav[8:12]))
	}

	// Check fmt chunk
	if string(wav[12:16]) != "fmt " {
		t.Errorf("missing fmt chunk, got %q", string(wav[12:16]))
	}

	// Check data chunk
	if string(wav[36:40]) != "data" {
		t.Errorf("missing data chunk, got %q", string(wav[36:40]))
	}

	// Check data size field
	dataSize := binary.LittleEndian.Uint32(wav[40:44])
	if int(dataSize) != expectedDataSize {
		t.Errorf("data size = %d, want %d", dataSize, expectedDataSize)
	}

	// Check sample rate in header (bytes 24-27)
	sr := binary.LittleEndian.Uint32(wav[24:28])
	if int(sr) != sampleRate {
		t.Errorf("sample rate in WAV = %d, want %d", sr, sampleRate)
	}

	// Check channels (bytes 22-23) = 1 (mono)
	channels := binary.LittleEndian.Uint16(wav[22:24])
	if channels != 1 {
		t.Errorf("channels = %d, want 1", channels)
	}

	// Check bits per sample (bytes 34-35) = 16
	bps := binary.LittleEndian.Uint16(wav[34:36])
	if bps != 16 {
		t.Errorf("bits per sample = %d, want 16", bps)
	}

	// Verify actual samples in data section
	reader := bytes.NewReader(wav[44:])
	for i, expected := range samples {
		var got int16
		if err := binary.Read(reader, binary.LittleEndian, &got); err != nil {
			t.Fatalf("read sample %d: %v", i, err)
		}
		if got != expected {
			t.Errorf("sample[%d] = %d, want %d", i, got, expected)
		}
	}
}

func TestEncodeWAVEmpty(t *testing.T) {
	wav, err := encodeWAV([]int16{}, 16000)
	if err != nil {
		t.Fatalf("encodeWAV error: %v", err)
	}
	// Should still produce a valid header (44 bytes)
	if len(wav) != 44 {
		t.Errorf("empty WAV length = %d, want 44", len(wav))
	}
}

func TestNewRecorder(t *testing.T) {
	r := NewRecorder(16000)
	if r == nil {
		t.Fatal("NewRecorder returned nil")
	}
	if r.sampleRate != 16000 {
		t.Errorf("sampleRate = %d, want 16000", r.sampleRate)
	}
	if r.recording {
		t.Error("new recorder should not be recording")
	}
}

func TestRecorderIsRecording(t *testing.T) {
	r := NewRecorder(16000)
	if r.IsRecording() {
		t.Error("should not be recording initially")
	}
}

func TestRecorderStopWithoutStart(t *testing.T) {
	r := NewRecorder(16000)
	_, err := r.Stop()
	if err == nil {
		t.Fatal("Stop without Start should error")
	}
	if err.Error() != "not recording" {
		t.Errorf("error = %q, want %q", err.Error(), "not recording")
	}
}

func TestRecorderDoubleStart(t *testing.T) {
	r := NewRecorder(16000)
	// Manually set recording flag to simulate already recording
	r.mu.Lock()
	r.recording = true
	r.mu.Unlock()

	err := r.Start()
	if err == nil {
		t.Fatal("double Start should error")
	}
	if err.Error() != "already recording" {
		t.Errorf("error = %q, want %q", err.Error(), "already recording")
	}
}
