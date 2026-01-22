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
		AutoReturn:  false,
	}
}

// Load loads configuration from environment and files
func Load() (*Config, error) {
	cfg := DefaultConfig()

	// 1. Try to load from file
	path, err := GetConfigPath()
	if err == nil {
		if data, err := os.ReadFile(path); err == nil {
			// Use a map to check if field exists in JSON
			var raw map[string]interface{}
			if err := json.Unmarshal(data, &raw); err == nil {
				if val, ok := raw["auto_return"]; ok {
					if b, ok := val.(bool); ok {
						cfg.AutoReturn = b
					}
				}
				// Also merge other fields safely
				if val, ok := raw["groq_api_key"].(string); ok && val != "" {
					cfg.GROQ_API_KEY = val
				}
				if val, ok := raw["hotkey"].(string); ok && val != "" {
					cfg.Hotkey = val
				}
				if val, ok := raw["audio_device"].(string); ok && val != "" {
					cfg.AudioDevice = val
				}
				if val, ok := raw["model"].(string); ok && val != "" {
					cfg.Model = val
				}
				if val, ok := raw["disable_notifications"].(bool); ok {
					cfg.DisableNotifications = val
				}
				if val, ok := raw["verbose"].(bool); ok {
					cfg.Verbose = val
				}
				if val, ok := raw["temperature"].(float64); ok {
					cfg.Temperature = val
				}
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
