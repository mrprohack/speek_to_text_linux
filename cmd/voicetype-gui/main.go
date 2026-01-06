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
	glowLayers   []*canvas.Rectangle
	waveBars     []*canvas.Rectangle
	status       *canvas.Text
	anim         *fyne.Animation
	pulseAnim    *fyne.Animation
	running      bool
	lastToggle   time.Time
	winTitle     string
	isProcessing bool
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

		app.pillBg.StrokeColor = color.RGBA{R: 225, G: 90, B: 164, A: 255}
		app.status.Text = "Listening..."
		app.pillBg.Refresh()
		app.status.Refresh()
		for _, glow := range app.glowLayers {
			glow.StrokeColor = color.RGBA{R: 225, G: 90, B: 164, A: 60}
			glow.Refresh()
		}
	})
	app.startWaveAnimation()
	app.startPulseAnimation()

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
		app.status.Text = "Transcribing..."
		app.status.Refresh()
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
			app.status.Text = "Typing..."
			app.status.Refresh()
		})

		if err := app.clip.TypeDirectly(app.ctx, text); err != nil {
			log.Printf("Typing failed: %v", err)
			app.safeUIUpdate(func() {
				app.a.Quit()
			})
			return
		}

		app.safeUIUpdate(func() {
			app.status.Text = "✓ Done"
			app.pillBg.StrokeColor = color.RGBA{R: 34, G: 197, B: 94, A: 255}
			app.pillBg.Refresh()
			app.status.Refresh()
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
			app.pillBg.StrokeColor = color.Black
			app.pillBg.StrokeWidth = 3
			for _, glow := range app.glowLayers {
				glow.StrokeColor = color.Transparent
				glow.Refresh()
			}
			app.pillBg.Refresh()
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
	// Larger window to accommodate the outer glow (shadow)
	app.window.Resize(fyne.NewSize(320, 110))
	app.window.CenterOnScreen()

	cream := color.RGBA{R: 255, G: 254, B: 242, A: 255}

	// Create glow layers (stacked rectangles with decreasing alpha)
	app.glowLayers = make([]*canvas.Rectangle, 4)
	for i := 0; i < 4; i++ {
		glow := canvas.NewRectangle(color.Transparent)
		glow.StrokeWidth = float32(i + 5)
		glow.StrokeColor = color.Transparent
		glow.CornerRadius = 38
		app.glowLayers[i] = glow
	}

	app.pillBg = canvas.NewRectangle(cream)
	app.pillBg.StrokeWidth = 3
	app.pillBg.StrokeColor = color.Black
	app.pillBg.CornerRadius = 30

	app.status = canvas.NewText("", color.RGBA{R: 80, G: 80, B: 80, A: 255})
	app.status.TextSize = 14
	app.status.TextStyle = fyne.TextStyle{Bold: true}
	app.status.Alignment = fyne.TextAlignCenter

	numBars := 22
	app.waveBars = make([]*canvas.Rectangle, numBars)
	for i := 0; i < numBars; i++ {
		bar := canvas.NewRectangle(color.Black)
		bar.SetMinSize(fyne.NewSize(3, 12))
		bar.Resize(fyne.NewSize(3, 12))
		bar.CornerRadius = 1.5
		bar.Hide()
		app.waveBars[i] = bar
	}

	waveContainer := container.NewHBox()
	for _, bar := range app.waveBars {
		waveContainer.Add(bar)
	}

	// Layout with Glow -> Pill -> Content
	glowStack := container.NewMax()
	for i := len(app.glowLayers) - 1; i >= 0; i-- {
		glowStack.Add(app.glowLayers[i])
	}

	mainContent := container.NewMax(
		app.pillBg,
		container.NewCenter(
			container.NewVBox(
				container.NewCenter(waveContainer),
				container.NewCenter(app.status),
			),
		),
	)

	app.window.SetContent(container.NewMax(
		container.NewPadded(glowStack),
		container.NewPadded(mainContent),
	))

	go func() {
		time.Sleep(150 * time.Millisecond)
		for i := 0; i < 3; i++ {
			app.stripDecorations(app.winTitle)
			time.Sleep(250 * time.Millisecond)
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

	app.anim = fyne.NewAnimation(time.Millisecond*40, func(f float32) {
		for i, bar := range app.waveBars {
			if bar == nil {
				continue
			}
			elapsed := time.Since(startTime).Seconds()
			offset := float64(i) * 0.25
			h := 2 + 50*math.Abs(math.Sin(elapsed*8+offset))
			bar.Resize(fyne.NewSize(3, float32(h)))
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

func (app *VoiceTypeApp) startPulseAnimation() {
	pink := color.RGBA{R: 225, G: 90, B: 164, A: 255}
	app.pulseAnim = fyne.NewAnimation(time.Second*2, func(f float32) {
		app.safeUIUpdate(func() {
			val := math.Sin(float64(f) * 2 * math.Pi)
			app.pillBg.StrokeWidth = 3 + 1.2*float32(val)
			alpha := uint8(100 + 100*val)
			for i, glow := range app.glowLayers {
				glow.StrokeColor = color.RGBA{R: pink.R, G: pink.G, B: pink.B, A: uint8(float64(alpha) / float64(i+1))}
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
