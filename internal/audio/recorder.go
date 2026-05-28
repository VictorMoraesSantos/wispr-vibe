package audio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os/exec"
	"sync"
)

// Recorder captures audio from the default input device using ffmpeg.
// This avoids CGo dependency — just needs ffmpeg in PATH.
type Recorder struct {
	sampleRate int
	cmd        *exec.Cmd
	buf        bytes.Buffer
	mu         sync.Mutex
	recording  bool
	done       chan struct{}
}

func NewRecorder(sampleRate int) *Recorder {
	return &Recorder{
		sampleRate: sampleRate,
	}
}

func (r *Recorder) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.recording {
		return fmt.Errorf("already recording")
	}

	// Use ffmpeg to capture from default audio input device
	// On Windows: dshow; On Linux: pulse/alsa; On macOS: avfoundation
	r.buf.Reset()
	r.done = make(chan struct{})

	args := buildFFmpegArgs(r.sampleRate)
	r.cmd = exec.Command("ffmpeg", args...)
	r.cmd.Stdout = &r.buf
	r.cmd.Stderr = nil // suppress ffmpeg stderr noise

	if err := r.cmd.Start(); err != nil {
		return fmt.Errorf("start ffmpeg: %w (is ffmpeg installed?)", err)
	}

	r.recording = true

	go func() {
		r.cmd.Wait()
		close(r.done)
	}()

	return nil
}

func (r *Recorder) Stop() ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.recording {
		return nil, fmt.Errorf("not recording")
	}

	r.recording = false

	// Send quit signal to ffmpeg (write 'q' to stdin won't work without stdin pipe)
	// Kill the process gracefully
	if r.cmd != nil && r.cmd.Process != nil {
		r.cmd.Process.Kill()
	}

	<-r.done

	data := r.buf.Bytes()
	if len(data) == 0 {
		return nil, fmt.Errorf("no audio captured — check microphone and ffmpeg")
	}

	return data, nil
}

func (r *Recorder) IsRecording() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.recording
}

func encodeWAV(samples []int16, sampleRate int) ([]byte, error) {
	var buf bytes.Buffer

	numSamples := len(samples)
	dataSize := numSamples * 2
	fileSize := 36 + dataSize

	buf.WriteString("RIFF")
	binary.Write(&buf, binary.LittleEndian, int32(fileSize))
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	binary.Write(&buf, binary.LittleEndian, int32(16))
	binary.Write(&buf, binary.LittleEndian, int16(1))
	binary.Write(&buf, binary.LittleEndian, int16(1))
	binary.Write(&buf, binary.LittleEndian, int32(sampleRate))
	binary.Write(&buf, binary.LittleEndian, int32(sampleRate*2))
	binary.Write(&buf, binary.LittleEndian, int16(2))
	binary.Write(&buf, binary.LittleEndian, int16(16))
	buf.WriteString("data")
	binary.Write(&buf, binary.LittleEndian, int32(dataSize))
	for _, s := range samples {
		binary.Write(&buf, binary.LittleEndian, s)
	}

	return buf.Bytes(), nil
}
