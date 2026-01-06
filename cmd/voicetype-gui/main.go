package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"speek_to_text_linux/internal/api"
	"speek_to_text_linux/internal/audio"
	"speek_to_text_linux/internal/clipboard"
	"speek_to_text_linux/internal/hotkey"
	"speek_to_text_linux/internal/ui"
	"speek_to_text_linux/pkg/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

var version = "1.0.0"

const configFile = ".voicetype.conf"

type VoiceTypeApp struct {
	a            fyne.App
	cfg          *config.Config
	audioSys     *audio.System
	apiClient    *api.Client
	clip         *clipboard.System
	hotkey       *hotkey.Listener
	ctx          context.Context
	cancel       context.CancelFunc
	isRecording  bool
	mu           sync.Mutex
	window       fyne.Window
	pillBg       *canvas.Rectangle
	glowLayers   []*canvas.Rectangle // Kept for logic compatibility but will be empty/ignored
	waveBars     []*canvas.Rectangle
	status       *canvas.Text
	anim         *fyne.Animation
	pulseAnim    *fyne.Animation
	running      bool
	lastToggle   time.Time
	winTitle     string
	isProcessing bool
	statusIcon   *canvas.Image
	smoothLevel  float64
}

func main() {
	flagHelp := flag.Bool("help", false, "Show help")
	flagDevice := flag.String("device", "", "Audio device")
	flagToggle := flag.Bool("toggle", false, "Toggle recording on a running instance")
	flagStop := flag.Bool("stop", false, "Stop a running instance")
	flag.Parse()

	if *flagHelp {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}

	pidFile := filepath.Join(os.TempDir(), "voicetype-gui.pid")

	// Handle --toggle or --stop by sending signals to existing process
	if *flagToggle || *flagStop {
		data, err := os.ReadFile(pidFile)
		if err == nil {
			var pid int
			fmt.Sscanf(string(data), "%d", &pid)
			process, err := os.FindProcess(pid)
			if err == nil {
				if *flagToggle {
					_ = process.Signal(syscall.SIGUSR1)
					fmt.Println("Sent toggle signal to running instance.")
				} else {
					_ = process.Signal(syscall.SIGTERM)
					fmt.Println("Sent stop signal to running instance.")
				}
				os.Exit(0)
			}
		}
		if *flagToggle {
			fmt.Println("No running instance found. Starting new instance...")
		} else {
			fmt.Println("No running instance found.")
			os.Exit(1)
		}
	}

	// Single instance check for main app
	if _, err := os.Stat(pidFile); err == nil {
		// Verify if process really exists
		data, _ := os.ReadFile(pidFile)
		var oldPid int
		fmt.Sscanf(string(data), "%d", &oldPid)
		if p, err := os.FindProcess(oldPid); err == nil && p.Signal(syscall.Signal(0)) == nil {
			// Instead of just exiting, toggle the already running instance
			_ = p.Signal(syscall.SIGUSR1)
			fmt.Println("VoiceType is already running. Sent toggle signal.")
			os.Exit(0)
		}
	}
	os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
	defer os.Remove(pidFile)

	log.Println("VoiceType v" + version + " starting...")

	cfg, _ := config.Load()
	cfg.AudioDevice = *flagDevice

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
	app.clip = clipboard.NewSystem(nil)
	app.hotkey = hotkey.NewListener(nil)
	app.ctx, app.cancel = context.WithCancel(context.Background())

	// Handle Signals for toggling and quitting
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGUSR1, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range sigChan {
			switch sig {
			case syscall.SIGUSR1:
				app.toggleRecording()
			case syscall.SIGINT, syscall.SIGTERM:
				fyne.Do(func() {
					app.a.Quit()
				})
			}
		}
	}()

	// Restore background listener for persistent sessions
	if err := app.hotkey.Initialize(app.cfg.Hotkey); err != nil {
		log.Printf("Hotkey init failed: %v", err)
	}
	app.hotkey.OnPress(func() {
		app.toggleRecording()
	})
	if err := app.hotkey.Start(); err != nil {
		log.Printf("Hotkey start failed: %v", err)
	}

	app.createWindow()
	app.window.Show()
	app.safeUIUpdate(func() {
		app.status.Text = "VoiceType Ready"
		app.status.Refresh()
	})

	// Debounce toggle from same hotkey that launched the app
	app.lastToggle = time.Now()

	// One-shot auto-start on launch
	app.startRecording()

	// Safety shutdown: If the app is left idle for more than 60 seconds, quit.
	// This handles cases where --toggle was used but something hung.
	time.AfterFunc(60*time.Second, func() {
		app.mu.Lock()
		recording := app.isRecording
		app.mu.Unlock()
		if !recording {
			log.Println("Auto-shutting down due to inactivity")
			app.a.Quit()
		}
	})

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

		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		if strings.TrimSpace(line) == "" {
			app.toggleRecording()
		}
	}
}

func (app *VoiceTypeApp) toggleRecording() {
	app.mu.Lock()
	// Debounce and "Already Processing" check (prevents loop from xdotool CTRL+V)
	if time.Since(app.lastToggle) < 600*time.Millisecond || app.isProcessing {
		app.mu.Unlock()
		return
	}
	app.lastToggle = time.Now()
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
		log.Printf("Recording error: %v", err)
		return
	}

	app.mu.Lock()
	app.isRecording = true
	app.mu.Unlock()

	app.safeUIUpdate(func() {
		app.window.Show()
		app.window.RequestFocus()
		// Re-apply "Always on top" every time we show, just in case
		go app.stripDecorations(app.winTitle)

		// New UI uses specific status colors per section
		app.status.Text = "Listening... "
		app.statusIcon.Resource = theme.MediaRecordIcon()
		app.statusIcon.Show()
		app.status.Refresh()
		app.statusIcon.Refresh()
	})
	app.startWaveAnimation()
	app.startPulseAnimation(color.RGBA{R: 0, G: 195, B: 255, A: 255})

	log.Println("Recording started")
}

func (app *VoiceTypeApp) stopRecording() {
	audioData, err := app.audioSys.StopRecording()
	if err != nil {
		log.Printf("Stop error: %v", err)
		app.mu.Lock()
		app.isRecording = false
		app.mu.Unlock()
		return
	}

	app.mu.Lock()
	app.isRecording = false
	app.mu.Unlock()

	app.stopWaveAnimation()
	app.stopPulseAnimation()

	if len(audioData) == 0 {
		app.safeUIUpdate(func() {
			app.status.Text = ""
			app.status.Refresh()
		})
		return
	}

	app.safeUIUpdate(func() {
		app.status.Text = "Analyzing... "
		app.statusIcon.Resource = theme.ViewRefreshIcon() // Spinner-like
		app.status.Refresh()
		app.statusIcon.Refresh()
	})

	app.mu.Lock()
	app.isProcessing = true
	app.mu.Unlock()

	go func() {
		defer func() {
			app.mu.Lock()
			app.isProcessing = false
			app.mu.Unlock()
		}()

		text, err := app.apiClient.Transcribe(app.ctx, audioData)
		if err != nil {
			log.Printf("Transcription failed: %v", err)
			app.safeUIUpdate(func() {
				app.status.Text = "Error"
				app.pillBg.StrokeColor = color.RGBA{R: 239, G: 68, B: 68, A: 255} // Red
				app.pillBg.Refresh()
				app.status.Refresh()
			})
			time.Sleep(1500 * time.Millisecond)
			app.safeUIUpdate(func() {
				app.a.Quit()
			})
			return
		}

		text = strings.TrimSpace(text)
		if text == "" {
			app.safeUIUpdate(func() {
				app.a.Quit()
			})
			return
		}

		log.Printf("Transcribed: %s", text)

		app.safeUIUpdate(func() {
			app.status.Text = "Typing Result... "
			app.statusIcon.Resource = theme.HistoryIcon()
			app.status.Refresh()
			app.statusIcon.Refresh()
		})

		if err := app.clip.TypeDirectly(app.ctx, text); err != nil {
			log.Printf("Typing failed: %v", err)
			app.safeUIUpdate(func() {
				app.a.Quit()
			})
			return
		}

		app.safeUIUpdate(func() {
			app.status.Text = "✓ Done "
			app.statusIcon.Resource = theme.ConfirmIcon()
			app.status.Refresh()
			app.statusIcon.Refresh()
		})

		time.Sleep(1200 * time.Millisecond)
		app.safeUIUpdate(func() {
			app.a.Quit()
		})
	}()
}

func (app *VoiceTypeApp) resetUI() {
	app.mu.Lock()
	defer app.mu.Unlock()
	if !app.isRecording {
		app.safeUIUpdate(func() {
			// Fade out effect start
			app.status.Text = ""
			app.status.Refresh()

			// Small delay before hiding to let the user see the "✓" or "Error"
			time.AfterFunc(200*time.Millisecond, func() {
				app.safeUIUpdate(func() {
					app.window.Hide()
				})
			})
		})
	}
}

func (app *VoiceTypeApp) createWindow() {
	app.a.Settings().SetTheme(&ui.VoiceTypeTheme{})

	app.winTitle = fmt.Sprintf("VoiceTypeUI_%d", time.Now().UnixNano())
	app.window = app.a.NewWindow(app.winTitle)
	app.window.SetFixedSize(true)

	// Minimal wave + timer window
	app.window.Resize(fyne.NewSize(350, 50))

	// Wave Bars (Simple, Clean)
	numBars := 32
	app.waveBars = make([]*canvas.Rectangle, numBars)

	waveContainer := container.NewHBox()
	for i := 0; i < numBars; i++ {
		cyan := color.RGBA{R: 0, G: 200, B: 255, A: 255}
		bar := canvas.NewRectangle(cyan)
		bar.SetMinSize(fyne.NewSize(4, 6))
		bar.Resize(fyne.NewSize(4, 6))
		bar.CornerRadius = 2
		bar.Hide()
		app.waveBars[i] = bar
		waveContainer.Add(bar)
	}

	// Timer display (visible, white text)
	app.status = canvas.NewText("0:00", color.White)
	app.status.TextSize = 16
	app.status.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	app.status.Alignment = fyne.TextAlignCenter

	app.statusIcon = canvas.NewImageFromResource(theme.InfoIcon())
	app.statusIcon.Hide()

	// Placeholder for pillBg (not visible)
	app.pillBg = canvas.NewRectangle(color.Transparent)

	// Layout: Wave bars + Timer
	mainLayout := container.NewHBox(
		waveContainer,
		container.NewCenter(app.status),
	)

	app.window.SetContent(container.NewCenter(mainLayout))

	// Position at bottom center of screen
	go func() {
		time.Sleep(100 * time.Millisecond)
		for i := 0; i < 4; i++ {
			app.stripDecorations(app.winTitle)
			// Move window to bottom center
			exec.Command("wmctrl", "-r", app.winTitle, "-e", "0,-1,900,-1,-1").Run()
			time.Sleep(200 * time.Millisecond)
		}
	}()

	app.window.Show()
}

func (app *VoiceTypeApp) stripDecorations(title string) {
	exec.Command("xprop", "-name", title, "-f", "_MOTIF_WM_HINTS", "32c", "-set", "_MOTIF_WM_HINTS", "0x2, 0x0, 0x0, 0x0, 0x0").Run()
	exec.Command("wmctrl", "-r", title, "-b", "add,above,skip_taskbar,skip_pager").Run()
}

func (app *VoiceTypeApp) safeUIUpdate(f func()) {
	fyne.Do(f)
}

func (app *VoiceTypeApp) startWaveAnimation() {
	app.safeUIUpdate(func() {
		for _, bar := range app.waveBars {
			bar.Show()
		}
	})

	startTime := time.Now()

	app.anim = fyne.NewAnimation(time.Millisecond*30, func(f float32) {
		app.mu.Lock()
		level := app.audioSys.GetLevel()
		// Smoothing: move toward target level
		app.smoothLevel = app.smoothLevel*0.7 + level*0.3
		app.mu.Unlock()

		center := len(app.waveBars) / 2
		for i, bar := range app.waveBars {
			if bar == nil {
				continue
			}
			elapsed := time.Since(startTime).Seconds()

			// Symmetric index (distance from center)
			dist := float64(math.Abs(float64(i - center)))
			maxDist := float64(center)

			// Gaussian-like falloff for a pill shape
			falloff := math.Exp(-math.Pow(dist/(maxDist*0.8), 2))

			// Base idle movement
			idle := 2.0 * math.Sin(elapsed*5+float64(i)*0.2)

			// Vocal reactivity scaling
			vocal := app.smoothLevel * 60.0 * falloff

			h := 4.0 + math.Abs(idle) + vocal

			// Spectral color shift based on height
			percent := h / 50.0
			if percent > 1.0 {
				percent = 1.0
			}

			bar.FillColor = color.RGBA{
				R: uint8(0 + 100*percent),
				G: uint8(210 - 50*percent),
				B: 255,
				A: 255,
			}

			bar.Resize(fyne.NewSize(2, float32(h)))
			bar.Refresh()
		}
	})
	app.anim.RepeatCount = fyne.AnimationRepeatForever
	app.anim.Start()

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			app.mu.Lock()
			recording := app.isRecording
			app.mu.Unlock()
			if !recording {
				return
			}

			select {
			case <-ticker.C:
				elapsed := time.Since(startTime)
				mins := int(elapsed.Minutes())
				secs := int(elapsed.Seconds()) % 60
				app.safeUIUpdate(func() {
					app.status.Text = fmt.Sprintf("%d:%02d", mins, secs)
					app.status.Refresh()
				})
			case <-app.ctx.Done():
				return
			}
		}
	}()
}

func (app *VoiceTypeApp) startPulseAnimation(pulseColor color.RGBA) {
	app.pulseAnim = fyne.NewAnimation(time.Duration(float64(time.Second)*1.5), func(f float32) {
		app.safeUIUpdate(func() {
			// Phase-shifted sine for a breathing effect
			val := math.Sin(float64(f)*2*math.Pi - math.Pi/2)
			normVal := (val + 1) / 2 // 0 to 1

			app.pillBg.StrokeWidth = 2 + 1.2*float32(normVal)
			alphaBase := uint8(40 + 80*normVal)

			for i, glow := range app.glowLayers {
				alpha := uint8(float64(alphaBase) / float64(i+1))
				glow.StrokeColor = color.RGBA{R: pulseColor.R, G: pulseColor.G, B: pulseColor.B, A: alpha}
				glow.Refresh()
			}
			app.pillBg.Refresh()
		})
	})
	app.pulseAnim.RepeatCount = fyne.AnimationRepeatForever
	app.pulseAnim.Start()
}

func (app *VoiceTypeApp) stopPulseAnimation() {
	if app.pulseAnim != nil {
		app.pulseAnim.Stop()
		app.pulseAnim = nil
	}
	app.safeUIUpdate(func() {
		for _, glow := range app.glowLayers {
			glow.StrokeColor = color.Transparent
			glow.Refresh()
		}
	})
}

func (app *VoiceTypeApp) stopWaveAnimation() {
	if app.anim != nil {
		app.anim.Stop()
		app.anim = nil
	}
	app.resetUI()
	app.safeUIUpdate(func() {
		for _, bar := range app.waveBars {
			if bar != nil {
				bar.Hide()
				bar.Refresh()
			}
		}
	})
}

func (app *VoiceTypeApp) shutdown() {
	log.Println("Shutting down...")
	app.cancel()
	app.audioSys.Close()
	log.Println("Done")
}

func loadAPIKey() string {
	cfg, _ := config.Load()
	return cfg.GROQ_API_KEY
}

func saveAPIKey(key string) {
	cfg, _ := config.Load()
	cfg.GROQ_API_KEY = key
	cfg.Save("")
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
