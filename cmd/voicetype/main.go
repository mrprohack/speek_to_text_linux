// VoiceType - Linux Native Speech-to-Text App
// Main entry point for the application
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"VoiceType/internal/api"
	"VoiceType/internal/audio"
	"VoiceType/internal/clipboard"
	"VoiceType/internal/hotkey"
	"VoiceType/internal/notify"
	"VoiceType/internal/ui"
	"VoiceType/pkg/config"
	"VoiceType/pkg/errors"
)

var (
	version = "1.0.0"
	commit  = "dev"
	date    = "now"
)

func main() {
	// Parse command line flags
	versionFlag := flag.Bool("version", false, "Show version information")
	helpFlag := flag.Bool("help", false, "Show help information")
	verboseFlag := flag.Bool("v", false, "Enable verbose logging")
	hotkeyFlag := flag.String("hotkey", "F6", "Hotkey to trigger recording (default: F6)")
	deviceFlag := flag.String("device", "", "Audio device to use (auto-detect if empty)")
	noNotifyFlag := flag.Bool("no-notify", false, "Disable notifications")
	flag.Parse()

	// Handle version flag
	if *versionFlag {
		fmt.Printf("VoiceType version %s (commit: %s, date: %s)\n", version, commit, date)
		os.Exit(0)
	}

	// Handle help flag
	if *helpFlag {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}

	// Set up logging
	if *verboseFlag {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		log.SetFlags(log.LstdFlags)
		log.SetOutput(os.Stderr)
	}

	log.Printf("VoiceType v%s starting...", version)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Warning: Could not load config file: %v", err)
	}
	cfg.Hotkey = *hotkeyFlag
	cfg.AudioDevice = *deviceFlag
	cfg.DisableNotifications = *noNotifyFlag

	// Check for API key
	if cfg.GROQ_API_KEY == "" {
		log.Fatal("Error: GROQ_API_KEY environment variable is not set. Please set it before running VoiceType.")
	}

	// Create error handler
	errHandler := errors.NewHandler()

	// Create notification system
	notifier := notify.NewNotifier(errHandler)
	if !cfg.DisableNotifications {
		if err := notifier.Initialize(); err != nil {
			log.Printf("Warning: Could not initialize notifications: %v", err)
		}
	}

	// Create UI system
	recordingUI := ui.NewUI(errHandler)

	// Create audio system
	audioSys := audio.NewSystem(errHandler)
	if err := audioSys.Initialize(cfg.AudioDevice); err != nil {
		errHandler.Fatal("Failed to initialize audio system: %v", err)
	}
	defer audioSys.Close()

	// Create API client
	apiClient := api.NewClient(cfg.GROQ_API_KEY, errHandler)

	// Create clipboard system
	clip := clipboard.NewSystem(errHandler)

	// Create hotkey listener
	hotkeyListener := hotkey.NewListener(errHandler)
	if err := hotkeyListener.Initialize(cfg.Hotkey); err != nil {
		errHandler.Fatal("Failed to initialize hotkey listener: %v", err)
	}
	defer hotkeyListener.Close()

	// State management
	var wg sync.WaitGroup
	var stateMu sync.Mutex
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle hotkey events
	hotkeyListener.OnPress(func() {
		stateMu.Lock()
		recordingUI.Show()
		stateMu.Unlock()

		// Start recording
		if err := audioSys.StartRecording(); err != nil {
			errHandler.Error("Failed to start recording: %v", err)
			notifier.Notify("Recording Error", "Failed to start microphone. Check permissions.")
			return
		}
	})

	hotkeyListener.OnRelease(func() {
		stateMu.Lock()
		recordingUI.Hide()
		stateMu.Unlock()

		// Stop recording and get audio data
		audioData, err := audioSys.StopRecording()
		if err != nil {
			errHandler.Error("Failed to stop recording: %v", err)
			return
		}

		if len(audioData) == 0 {
			log.Printf("No audio recorded")
			return
		}

		// Send to transcription
		wg.Add(1)
		go func() {
			defer wg.Done()

			notifier.Notify("Transcribing...", "Processing your speech")

			// Transcribe audio
			text, err := apiClient.Transcribe(ctx, audioData)
			if err != nil {
				errHandler.Error("Transcription failed: %v", err)
				notifier.Notify("Transcription Failed", err.Error())
				return
			}

			if text == "" {
				log.Printf("No text detected in audio")
				return
			}

			// Copy to clipboard and paste
			if err := clip.SetAndPaste(ctx, text); err != nil {
				errHandler.Error("Failed to paste text: %v", err)
				notifier.Notify("Paste Error", "Failed to insert text")
				return
			}

			log.Printf("Successfully transcribed and pasted: %s", text)
		}()
	})

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutdown signal received, cleaning up...")
		cancel()
		hotkeyListener.Close()
		audioSys.Close()
	}()

	log.Println("VoiceType is running. Press and hold the hotkey to record.")
	log.Printf("Hotkey: %s", cfg.Hotkey)
	log.Println("Press Ctrl+C to exit")

	// Wait for completion
	wg.Wait()

	log.Println("VoiceType shutdown complete")
}
