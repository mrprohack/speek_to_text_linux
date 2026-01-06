// Package audio provides audio capture functionality for VoiceType
package audio

import (
	"fmt"
	"log"
	"os"
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
}

// NewSystem creates a new audio system
func NewSystem(errHandler *errors.Handler) *System {
	return &System{
		errHandler:    errHandler,
		sampleRate:    16000, // 16kHz for Whisper
		channels:      1,     // Mono
		bitsPerSample: 16,    // 16-bit
	}
}

// Initialize initializes the audio system
func (s *System) Initialize(device string) error {
	s.device = device

	// Try to detect available audio system
	if err := s.detectAudioSystem(); err != nil {
		return errors.Wrap(err, errors.ErrorTypeAudio, "failed to detect audio system")
	}

	log.Printf("Audio system initialized with device: %s", s.device)
	return nil
}

// detectAudioSystem detects available audio systems and selects one
func (s *System) detectAudioSystem() error {
	// Check for PulseAudio
	if os.Getenv("PULSE_SERVER") != "" {
		s.device = "pulse"
		return nil
	}

	// Check for ALSA devices
	if s.device == "" {
		// Try to list ALSA devices
		device, err := s.findALSADevice()
		if err == nil {
			s.device = device
			return nil
		}
	}

	// Default to ALSA default device
	if s.device == "" {
		s.device = "default"
	}

	return nil
}

// findALSADevice finds an available ALSA device
func (s *System) findALSADevice() (string, error) {
	// This is a simplified implementation
	// In a real implementation, we would use alsa-go or similar library
	// For now, we'll use the default device
	return "default", nil
}

// StartRecording starts audio recording
func (s *System) StartRecording() error {
	if s.isRecording {
		return errors.NewError(errors.ErrorTypeAudio, "already recording", nil)
	}

	s.audioBuffer = make([]byte, 0)
	s.isRecording = true

	log.Printf("Started recording audio at %d Hz", s.sampleRate)
	return nil
}

// StopRecording stops recording and returns audio data
func (s *System) StopRecording() ([]byte, error) {
	if !s.isRecording {
		return nil, errors.NewError(errors.ErrorTypeAudio, "not recording", nil)
	}

	s.isRecording = false

	if len(s.audioBuffer) == 0 {
		return nil, errors.ErrAudioTooShort
	}

	log.Printf("Stopped recording, captured %d bytes of audio", len(s.audioBuffer))

	// Return a copy of the audio buffer
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

// addAudioData adds audio data to the buffer
func (s *System) addAudioData(data []byte) {
	s.audioBuffer = append(s.audioBuffer, data...)
}

// ReadAudioDevice reads audio from the device (stub implementation)
func (s *System) ReadAudioDevice() ([]byte, error) {
	// This would be implemented with actual ALSA/PulseAudio calls
	// For now, return silence
	return make([]byte, 1024), nil
}

// Import os for file operations
var _ = fmt.Sprintf
