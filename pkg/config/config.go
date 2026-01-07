package config

import (
	"encoding/json"
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
	AutoReturn           bool    `json:"auto_return"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Hotkey:      "ctrl+space",
		AudioDevice: "",
		Model:       "whisper-large-v3",
		Temperature: 0.0,
		AutoReturn:  true,
	}
}

// Load loads configuration from environment and files
func Load() (*Config, error) {
	cfg := DefaultConfig()

	// 1. Try to load from file
	path, err := GetConfigPath()
	if err == nil {
		if data, err := os.ReadFile(path); err == nil {
			var fileCfg Config
			if err := json.Unmarshal(data, &fileCfg); err == nil {
				// Merge values from file
				if fileCfg.GROQ_API_KEY != "" {
					cfg.GROQ_API_KEY = fileCfg.GROQ_API_KEY
				}
				if fileCfg.Hotkey != "" {
					cfg.Hotkey = fileCfg.Hotkey
				}
				if fileCfg.AudioDevice != "" {
					cfg.AudioDevice = fileCfg.AudioDevice
				}
				if fileCfg.Model != "" {
					cfg.Model = fileCfg.Model
				}
				cfg.AutoReturn = fileCfg.AutoReturn
				cfg.DisableNotifications = fileCfg.DisableNotifications
				cfg.Verbose = fileCfg.Verbose
				cfg.Temperature = fileCfg.Temperature
			}
		}
	}

	// 2. Override with environment variables
	if ek := os.Getenv("GROQ_API_KEY"); ek != "" {
		cfg.GROQ_API_KEY = ek
	}

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
	if path == "" {
		var err error
		path, err = GetConfigPath()
		if err != nil {
			return err
		}
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
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
