package main
import (
"encoding/json"
"os"
"path/filepath"
"log"
)
type Config struct {
	GROQ_API_KEY         string  `json:"groq_api_key"`
	Hotkey               string  `json:"hotkey"`
	AudioDevice          string  `json:"audio_device"`
	DisableNotifications bool    `json:"disable_notifications"`
	Verbose              bool    `json:"verbose"`
	Model                string  `json:"model"`
	Temperature          float64 `json:"temperature"`
}
func main() {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".config", "voicetype", "config.json")
	
	cfg := Config{
		GROQ_API_KEY: os.Getenv("GROQ_API_KEY"),
		Hotkey: "ctrl+space",
		Model: "whisper-large-v3",
	}
	
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(path, data, 0600)
}
