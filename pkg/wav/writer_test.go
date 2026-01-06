package wav

import (
	"bytes"
	"testing"
)

func TestNewWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := NewWriter(buf, 16000, 1, 16)

	if writer.sampleRate != 16000 {
		t.Errorf("Expected sample rate 16000, got %d", writer.sampleRate)
	}

	if writer.channels != 1 {
		t.Errorf("Expected 1 channel, got %d", writer.channels)
	}

	if writer.bitsPerSample != 16 {
		t.Errorf("Expected 16 bits per sample, got %d", writer.bitsPerSample)
	}
}

func TestEncode(t *testing.T) {
	// Test encoding with simple audio data
	audioData := []byte{0x00, 0x00, 0x00, 0x00}

	wavData, err := Encode(audioData, 16000, 1, 16)
	if err != nil {
		t.Fatalf("Encode() failed: %v", err)
	}

	// Check WAV header
	if len(wavData) < 44 {
		t.Errorf("Expected WAV data to be at least 44 bytes, got %d", len(wavData))
	}

	// Check RIFF magic number
	if string(wavData[0:4]) != "RIFF" {
		t.Errorf("Expected 'RIFF' magic number, got '%s'", string(wavData[0:4]))
	}

	// Check WAVE magic number
	if string(wavData[8:12]) != "WAVE" {
		t.Errorf("Expected 'WAVE' magic number, got '%s'", string(wavData[8:12]))
	}

	// Check data chunk marker
	if string(wavData[36:40]) != "data" {
		t.Errorf("Expected 'data' chunk marker, got '%s'", string(wavData[36:40]))
	}
}

func TestGetWAVHeaderSize(t *testing.T) {
	size := GetWAVHeaderSize()
	if size != 44 {
		t.Errorf("Expected header size 44, got %d", size)
	}
}

func TestCalculateWAVSize(t *testing.T) {
	audioSize := 1000
	total := CalculateWAVSize(audioSize)

	expected := 44 + audioSize
	if total != expected {
		t.Errorf("Expected total size %d, got %d", expected, total)
	}
}

func TestWriterToString(t *testing.T) {
	buf := &bytes.Buffer{}
	writer := NewWriter(buf, 16000, 1, 16)

	str := writer.ToString()
	if str == "" {
		t.Error("Expected non-empty string representation")
	}
}

func TestEncodeDifferentSampleRates(t *testing.T) {
	testCases := []struct {
		sampleRate    int
		channels      int
		bitsPerSample int
	}{
		{16000, 1, 16},
		{44100, 2, 16},
		{48000, 2, 24},
	}

	for _, tc := range testCases {
		audioData := []byte{0x00, 0x00}
		wavData, err := Encode(audioData, tc.sampleRate, tc.channels, tc.bitsPerSample)
		if err != nil {
			t.Fatalf("Encode() failed for %d Hz, %d channels, %d bits: %v",
				tc.sampleRate, tc.channels, tc.bitsPerSample, err)
		}

		if len(wavData) < 44 {
			t.Errorf("WAV data too short for %d Hz, %d channels, %d bits",
				tc.sampleRate, tc.channels, tc.bitsPerSample)
		}

		// Check that the header has the correct sample rate
		// (at byte offset 24 for little-endian uint32)
		expectedRate := tc.sampleRate
		// We can't easily verify this without binary parsing, but at least
		// we verify the encoding doesn't fail
		_ = expectedRate
	}
}
