// Package wav provides WAV file encoding functionality for VoiceType
package wav

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Writer represents a WAV file writer
type Writer struct {
	w             io.Writer
	sampleRate    int
	channels      int
	bitsPerSample int
	dataSize      int
	headerWritten bool
}

// NewWriter creates a new WAV writer
func NewWriter(w io.Writer, sampleRate, channels, bitsPerSample int) *Writer {
	return &Writer{
		w:             w,
		sampleRate:    sampleRate,
		channels:      channels,
		bitsPerSample: bitsPerSample,
	}
}

// Write writes audio data to the WAV file
func (w *Writer) Write(p []byte) (int, error) {
	if !w.headerWritten {
		if err := w.writeHeader(); err != nil {
			return 0, err
		}
		w.headerWritten = true
	}

	// Write audio data
	n, err := w.w.Write(p)
	w.dataSize += n
	return n, err
}

// writeHeader writes the WAV file header
func (w *Writer) writeHeader() error {
	// WAV file format:
	// RIFF header (12 bytes)
	// fmt chunk (24 bytes for PCM)
	// data chunk (8 bytes header + audio data)

	byterate := w.sampleRate * w.channels * w.bitsPerSample / 8
	blockalign := w.channels * w.bitsPerSample / 8

	// RIFF header
	header := make([]byte, 44)

	// RIFF magic number
	copy(header[0:4], []byte("RIFF"))

	// File size (36 + data size)
	binary.LittleEndian.PutUint32(header[4:8], uint32(36+w.dataSize))

	// WAVE magic number
	copy(header[8:12], []byte("WAVE"))

	// fmt chunk
	copy(header[12:16], []byte("fmt "))

	// fmt chunk size (16 for PCM)
	binary.LittleEndian.PutUint32(header[16:20], 16)

	// Audio format (1 for PCM)
	binary.LittleEndian.PutUint16(header[20:22], 1)

	// Number of channels
	binary.LittleEndian.PutUint16(header[22:24], uint16(w.channels))

	// Sample rate
	binary.LittleEndian.PutUint32(header[24:28], uint32(w.sampleRate))

	// Byte rate
	binary.LittleEndian.PutUint32(header[28:32], uint32(byterate))

	// Block align
	binary.LittleEndian.PutUint16(header[32:34], uint16(blockalign))

	// Bits per sample
	binary.LittleEndian.PutUint16(header[34:36], uint16(w.bitsPerSample))

	// data chunk
	copy(header[36:40], []byte("data"))

	// data chunk size
	binary.LittleEndian.PutUint32(header[40:44], uint32(w.dataSize))

	_, err := w.w.Write(header)
	return err
}

// Close finalizes the WAV file
func (w *Writer) Close() error {
	if !w.headerWritten {
		// Write header with 0 data size
		w.dataSize = 0
		if err := w.writeHeader(); err != nil {
			return err
		}
	}

	// Rewrite the header with the correct data size
	return nil
}

// Encode audio data to WAV format in memory
func Encode(audioData []byte, sampleRate, channels, bitsPerSample int) ([]byte, error) {
	buf := &bufferWriter{
		data: make([]byte, 0, len(audioData)+44),
	}

	writer := NewWriter(buf, sampleRate, channels, bitsPerSample)

	// Write header first
	if err := writer.writeHeader(); err != nil {
		return nil, err
	}

	// Write audio data
	if _, err := buf.Write(audioData); err != nil {
		return nil, err
	}

	return buf.data, nil
}

// bufferWriter is a helper for writing to a byte slice
type bufferWriter struct {
	data []byte
}

func (b *bufferWriter) Write(p []byte) (int, error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

// GetWAVHeaderSize returns the size of a WAV header
func GetWAVHeaderSize() int {
	return 44
}

// CalculateWAVSize calculates the total size of a WAV file
func CalculateWAVSize(audioDataSize int) int {
	return 44 + audioDataSize
}

// ToString returns a string representation of the WAV writer
func (w *Writer) ToString() string {
	return fmt.Sprintf("WAV Writer: %d Hz, %d channel(s), %d bits per sample",
		w.sampleRate, w.channels, w.bitsPerSample)
}
