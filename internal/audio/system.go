// Package audio provides audio capture functionality for VoiceType
package audio

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"VoiceType/pkg/errors"
)

// System represents the audio capture system
type System struct {
	errHandler    *errors.Handler
	sampleRate    int
	channels      int
	bitsPerSample int
	device        string
	isRecording   bool
	audioBuffer   []byte
	cmd           *exec.Cmd
	stdout        io.ReadCloser
}

// NewSystem creates a new audio system
func NewSystem(errHandler *errors.Handler) *System {
	return &System{
		errHandler:    errHandler,
		sampleRate:    16000,
		channels:      1,
		bitsPerSample: 16,
		device:        "default",
	}
}

// Initialize initializes the audio system
func (s *System) Initialize(device string) error {
	if device != "" {
		s.device = device
	}
	log.Printf("Audio system initialized with device: %s", s.device)
	return nil
}

// StartRecording starts audio recording from microphone
func (s *System) StartRecording() error {
	if s.isRecording {
		return errors.NewError(errors.ErrorTypeAudio, "already recording", nil)
	}

	s.audioBuffer = make([]byte, 0)

	// Use arecord to capture real audio from microphone
	args := []string{
		"-D", s.device,
		"-f", "S16_LE",
		"-r", fmt.Sprintf("%d", s.sampleRate),
		"-c", fmt.Sprintf("%d", s.channels),
		"-t", "raw",
	}

	s.cmd = exec.Command("arecord", args...)

	var err error
	s.stdout, err = s.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start arecord: %w", err)
	}

	s.isRecording = true

	// Read audio data in background
	go s.readAudio()

	log.Printf("Started recording audio at %d Hz", s.sampleRate)
	return nil
}

// readAudio reads audio data from arecord
func (s *System) readAudio() {
	buffer := make([]byte, 4096)
	for s.isRecording {
		n, err := s.stdout.Read(buffer)
		if n > 0 {
			s.audioBuffer = append(s.audioBuffer, buffer[:n]...)
		}
		if err != nil {
			break
		}
	}
}

// StopRecording stops recording and returns audio data
func (s *System) StopRecording() ([]byte, error) {
	if !s.isRecording {
		return nil, errors.NewError(errors.ErrorTypeAudio, "not recording", nil)
	}

	s.isRecording = false

	// Stop arecord
	if s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Kill()
		s.cmd.Wait()
	}

	if s.stdout != nil {
		s.stdout.Close()
	}

	if len(s.audioBuffer) == 0 {
		return nil, errors.ErrAudioTooShort
	}

	log.Printf("Stopped recording, captured %d bytes of audio", len(s.audioBuffer))

	result := make([]byte, len(s.audioBuffer))
	copy(result, s.audioBuffer)
	s.audioBuffer = nil

	return result, nil
}

// Close closes the audio system
func (s *System) Close() error {
	if s.isRecording {
		s.StopRecording()
	}
	log.Println("Audio system closed")
	return nil
}

// GetAudioBuffer returns the current audio buffer
func (s *System) GetAudioBuffer() []byte {
	return s.audioBuffer
}

// IsRecording returns whether the system is currently recording
func (s *System) IsRecording() bool {
	return s.isRecording
}

// SampleRate returns the sample rate
func (s *System) SampleRate() int {
	return s.sampleRate
}

// Channels returns the number of channels
func (s *System) Channels() int {
	return s.channels
}

// BitsPerSample returns the bits per sample
func (s *System) BitsPerSample() int {
	return s.bitsPerSample
}

// Duration returns the duration of recorded audio
func (s *System) Duration() time.Duration {
	if len(s.audioBuffer) == 0 {
		return 0
	}

	bytesPerSample := s.bitsPerSample / 8
	samples := len(s.audioBuffer) / (bytesPerSample * s.channels)
	return time.Duration(samples) * time.Second / time.Duration(s.sampleRate)
}

// SaveToFile saves audio buffer to a WAV file (for testing)
func (s *System) SaveToFile(filename string) error {
	if len(s.audioBuffer) == 0 {
		return fmt.Errorf("no audio data to save")
	}

	// Create WAV file manually
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write WAV header
	dataSize := len(s.audioBuffer)
	fileSize := 36 + dataSize

	// RIFF header
	file.Write([]byte("RIFF"))
	writeInt32(file, fileSize)
	file.Write([]byte("WAVE"))

	// fmt chunk
	file.Write([]byte("fmt "))
	writeInt32(file, 16)                                        // Subchunk1Size
	writeInt16(file, 1)                                         // AudioFormat (PCM)
	writeInt16(file, int16(s.channels))                         // NumChannels
	writeInt32(file, s.sampleRate)                              // SampleRate
	writeInt32(file, s.sampleRate*s.bitsPerSample/8*s.channels) // ByteRate
	writeInt16(file, int16(s.bitsPerSample/8*s.channels))       // BlockAlign
	writeInt16(file, int16(s.bitsPerSample))                    // BitsPerSample

	// data chunk
	file.Write([]byte("data"))
	writeInt32(file, dataSize)
	file.Write(s.audioBuffer)

	log.Printf("Saved audio to %s (%d bytes)", filename, fileSize)
	return nil
}

func writeInt16(file *os.File, v int16) {
	b := []byte{byte(v), byte(v >> 8)}
	file.Write(b)
}

func writeInt32(file *os.File, v int) {
	b := []byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)}
	file.Write(b)
}
