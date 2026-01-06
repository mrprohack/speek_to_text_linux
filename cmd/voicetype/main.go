package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"image/color"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"speek_to_text_linux/internal/api"
	"speek_to_text_linux/internal/audio"
	"speek_to_text_linux/internal/typing"
	"speek_to_text_linux/pkg/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var version = "1.0.0"

const configFile = ".voicetype.conf"

type VoiceTypeApp struct {
	a           fyne.App
	cfg         *config.Config
	audioSys    *audio.System
	apiClient   *api.Client
	typer       *typing.System
	ctx         context.Context
	cancel      context.CancelFunc
	isRecording bool
	mu          sync.Mutex
	window      fyne.Window
	statusLabel *widget.Label
	icon        *canvas.Text
	running     bool
}

func main() {
	flagHelp := flag.Bool("help", false, "Show help")
	flagDevice := flag.String("device", "", "Audio device")
	flag.Parse()

	if *flagHelp {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}

	log.Println("VoiceType v" + version + " starting...")

	cfg, _ := config.Load()
	cfg.AudioDevice = *flagDevice

	// Load or ask for API key
	apiKey := loadAPIKey()
	if apiKey == "" {
		apiKey = askAPIKey()
		saveAPIKey(apiKey)
	}
	cfg.GROQ_API_KEY = apiKey

	app := &VoiceTypeApp{
		a:       app.NewWithID("com.voicetype.app"),
		cfg:     cfg,
		running: true,
	}

	app.audioSys = audio.NewSystem(nil)
	if err := app.audioSys.Initialize(cfg.AudioDevice); err != nil {
		log.Fatalf("Audio init failed: %v", err)
	}

	app.apiClient = api.NewClient(cfg.GROQ_API_KEY, nil)
	app.typer = typing.NewSystem()
	app.ctx, app.cancel = context.WithCancel(context.Background())

	// Create window
	app.createWindow()

	// Start stdin reader in background
	go app.readStdin()

	fmt.Println()
	fmt.Println("VoiceType is running!")
	fmt.Println("Press ENTER to start/stop recording")
	fmt.Println("Or use the GUI window")
	fmt.Println("Press Ctrl+C to quit")
	fmt.Println()

	// Run Fyne app (this blocks)
	app.a.Run()

	app.shutdown()
}

func (app *VoiceTypeApp) readStdin() {
	reader := bufio.NewReader(os.Stdin)
	for {
		select {
		case <-app.ctx.Done():
			return
		default:
		}

		// Read one line (blocks until Enter pressed)
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		// Only react to pure Enter key
		if strings.TrimSpace(line) == "" {
			app.toggleRecording()
		}
	}
}

func (app *VoiceTypeApp) toggleRecording() {
	app.mu.Lock()
	recording := app.isRecording
	app.mu.Unlock()

	if recording {
		app.stopRecording()
	} else {
		app.startRecording()
	}
}

func (app *VoiceTypeApp) startRecording() {
	if err := app.audioSys.StartRecording(); err != nil {
		log.Printf("❌ Recording error: %v", err)
		app.updateUI("❌", "Error")
		return
	}

	app.mu.Lock()
	app.isRecording = true
	app.mu.Unlock()

	app.updateUI("🔴", "Recording...")
	log.Println("🎤 Recording... (press Enter to stop)")
}

func (app *VoiceTypeApp) stopRecording() {
	audioData, err := app.audioSys.StopRecording()
	if err != nil {
		log.Printf("❌ Stop error: %v", err)
		app.mu.Lock()
		app.isRecording = false
		app.mu.Unlock()
		app.updateUI("🎤", "Ready")
		return
	}

	app.mu.Lock()
	app.isRecording = false
	app.mu.Unlock()

	if len(audioData) == 0 {
		log.Println("⚠️ No audio recorded")
		app.updateUI("🎤", "Ready")
		return
	}

	log.Printf("⏹️ Stopped (%d bytes, transcribing...)", len(audioData))
	app.updateUI("⏳", "Transcribing...")

	// Transcribe in background
	go func() {
		text, err := app.apiClient.Transcribe(app.ctx, audioData)
		if err != nil {
			log.Printf("❌ Transcription failed: %v", err)
			app.updateUI("❌", "Error")
			return
		}

		if text == "" {
			log.Println("⚠️ No speech detected")
			app.updateUI("🎤", "Ready")
			return
		}

		log.Printf("✅ \"%s\"", text)

		if err := app.typer.TypeText(app.ctx, text); err != nil {
			log.Printf("❌ Type error: %v", err)
			app.updateUI("❌", "Type error")
			return
		}

		log.Println("📋 Text pasted!")
		app.updateUI("✅", "Done: "+text[:min(20, len(text))]+"...")
	}()

	// Reset to ready after delay
	time.AfterFunc(3*time.Second, func() {
		app.mu.Lock()
		if !app.isRecording {
			app.updateUI("🎤", "Ready")
		}
		app.mu.Unlock()
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (app *VoiceTypeApp) updateUI(icon, status string) {
	fyne.DoAndWait(func() {
		if app.icon != nil {
			app.icon.Text = icon
		}
		if app.statusLabel != nil {
			app.statusLabel.Text = status
		}
	})
}

func (app *VoiceTypeApp) createWindow() {
	app.window = app.a.NewWindow("VoiceType")
	app.window.SetFixedSize(true)
	app.window.Resize(fyne.NewSize(280, 140))
	app.window.CenterOnScreen()
	app.window.SetCloseIntercept(func() {
		fyne.DoAndWait(func() {
			app.window.Hide()
		})
	})

	app.icon = canvas.NewText("🎤", &color.RGBA{R: 100, G: 200, B: 100, A: 255})
	app.icon.Alignment = fyne.TextAlignCenter
	app.icon.TextSize = 48

	app.statusLabel = widget.NewLabel("Ready")
	app.statusLabel.Alignment = fyne.TextAlignCenter
	app.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	instr := canvas.NewText("Press ENTER to record", &color.RGBA{R: 150, G: 150, B: 150, A: 255})
	instr.Alignment = fyne.TextAlignCenter
	instr.TextSize = 12

	content := container.NewVBox(
		container.NewCenter(app.icon),
		container.NewCenter(app.statusLabel),
		layout.NewSpacer(),
		container.NewCenter(instr),
	)

	app.window.SetContent(content)
}

func (app *VoiceTypeApp) shutdown() {
	log.Println("Shutting down...")
	app.cancel()
	app.audioSys.Close()
	if app.window != nil {
		app.window.Close()
	}
	log.Println("Done")
}

func getConfigPath() string {
	home, _ := os.UserHomeDir()
	return home + "/" + configFile
}

func loadAPIKey() string {
	path := getConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func saveAPIKey(key string) {
	path := getConfigPath()
	os.WriteFile(path, []byte(key), 0600)
}

func askAPIKey() string {
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("  VoiceType - First Time Setup")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("Enter your Groq API key:")
	fmt.Println("(Get it from https://console.groq.com/)")
	fmt.Println()
	fmt.Print("GROQ_API_KEY: ")

	reader := bufio.NewReader(os.Stdin)
	key, _ := reader.ReadString('\n')
	key = strings.TrimSpace(key)

	if key == "" {
		log.Fatal("Error: API key is required")
	}

	return key
}
