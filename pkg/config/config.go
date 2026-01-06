// Package config provides configuration management for VoiceType
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	GROQ_API_KEY         string  `json:"groq_api_key"`
	Hotkey               string  `json:"hotkey"`
	AudioDevice          string  `json:"audio_device"`
	DisableNotifications bool    `json:"disable_notifications"`
	Verbose              bool    `json:"verbose"`
	Model                string  `json:"model"`
	Temperature          float64 `json:"temperature"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Hotkey:      "F6",
		AudioDevice: "",
		Model:       "whisper-large-v3",
		Temperature: 0.0,
	}
}

// Load loads configuration from environment and files
func Load() (*Config, error) {
	cfg := DefaultConfig()

	// Load from environment
	cfg.GROQ_API_KEY = os.Getenv("GROQ_API_KEY")

	// Override with environment variables if set
	if hotkey := os.Getenv("VOICE_TYPE_HOTKEY"); hotkey != "" {
		cfg.Hotkey = hotkey
	}

	if device := os.Getenv("VOICE_TYPE_AUDIO_DEVICE"); device != "" {
		cfg.AudioDevice = device
	}

	if model := os.Getenv("VOICE_TYPE_MODEL"); model != "" {
		cfg.Model = model
	}

	if tempStr := os.Getenv("VOICE_TYPE_TEMPERATURE"); tempStr != "" {
		var temp float64
		if _, err := fmt.Sscanf(tempStr, "%f", &temp); err == nil {
			cfg.Temperature = temp
		}
	}

	if os.Getenv("VOICE_TYPE_NOTIFICATIONS") == "0" {
		cfg.DisableNotifications = true
	}

	if os.Getenv("VOICE_TYPE_VERBOSE") == "1" {
		cfg.Verbose = true
	}

	return cfg, nil
}

// Save saves configuration to a file
func (c *Config) Save(path string) error {
	// Implementation would go here for persistent config
	// For now, we only support environment-based config
	_ = path
	return nil
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(home, ".config", "voicetype")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.json"), nil
}
