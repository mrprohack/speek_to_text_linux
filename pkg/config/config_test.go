package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Hotkey != "F6" {
		t.Errorf("Expected default hotkey 'F6', got '%s'", cfg.Hotkey)
	}

	if cfg.AudioDevice != "" {
		t.Errorf("Expected empty audio device, got '%s'", cfg.AudioDevice)
	}

	if cfg.Model != "whisper-large-v3" {
		t.Errorf("Expected default model 'whisper-large-v3', got '%s'", cfg.Model)
	}

	if cfg.Temperature != 0.0 {
		t.Errorf("Expected default temperature 0.0, got %f", cfg.Temperature)
	}
}

func TestLoad(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("VOICE_TYPE_HOTKEY")
	os.Unsetenv("VOICE_TYPE_AUDIO_DEVICE")
	os.Unsetenv("VOICE_TYPE_MODEL")
	os.Unsetenv("VOICE_TYPE_TEMPERATURE")
	os.Unsetenv("VOICE_TYPE_NOTIFICATIONS")
	os.Unsetenv("VOICE_TYPE_VERBOSE")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.GROQ_API_KEY != "" {
		t.Errorf("Expected empty GROQ_API_KEY, got '%s'", cfg.GROQ_API_KEY)
	}
}

func TestLoadWithEnvironment(t *testing.T) {
	// Set environment variables
	os.Setenv("VOICE_TYPE_HOTKEY", "F12")
	os.Setenv("VOICE_TYPE_AUDIO_DEVICE", "hw:0")
	os.Setenv("VOICE_TYPE_MODEL", "whisper-1")
	os.Setenv("VOICE_TYPE_TEMPERATURE", "0.5")
	os.Setenv("VOICE_TYPE_NOTIFICATIONS", "0")
	os.Setenv("VOICE_TYPE_VERBOSE", "1")

	defer func() {
		os.Unsetenv("VOICE_TYPE_HOTKEY")
		os.Unsetenv("VOICE_TYPE_AUDIO_DEVICE")
		os.Unsetenv("VOICE_TYPE_MODEL")
		os.Unsetenv("VOICE_TYPE_TEMPERATURE")
		os.Unsetenv("VOICE_TYPE_NOTIFICATIONS")
		os.Unsetenv("VOICE_TYPE_VERBOSE")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Hotkey != "F12" {
		t.Errorf("Expected hotkey 'F12', got '%s'", cfg.Hotkey)
	}

	if cfg.AudioDevice != "hw:0" {
		t.Errorf("Expected audio device 'hw:0', got '%s'", cfg.AudioDevice)
	}

	if cfg.Model != "whisper-1" {
		t.Errorf("Expected model 'whisper-1', got '%s'", cfg.Model)
	}

	if cfg.Temperature != 0.5 {
		t.Errorf("Expected temperature 0.5, got %f", cfg.Temperature)
	}

	if !cfg.DisableNotifications {
		t.Error("Expected DisableNotifications to be true")
	}

	if !cfg.Verbose {
		t.Error("Expected Verbose to be true")
	}
}
