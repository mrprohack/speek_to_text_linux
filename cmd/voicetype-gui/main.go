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
	"strings"
	"sync"
	"time"

	"VoiceType/internal/api"
	"VoiceType/internal/audio"
	"VoiceType/internal/clipboard"
	"VoiceType/internal/hotkey"
	"VoiceType/internal/ui"
	"VoiceType/pkg/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

var version = "1.0.0"

const configFile = ".voicetype.conf"

type VoiceTypeApp struct {
	a           fyne.App
	cfg         *config.Config
	audioSys    *audio.System
	apiClient   *api.Client
	clip        *clipboard.System
	hotkey      *hotkey.Listener
	ctx         context.Context
	cancel      context.CancelFunc
	isRecording bool
	mu          sync.Mutex
	window      fyne.Window
	pillBg      *canvas.Rectangle
	waveBars    []*canvas.Rectangle
	status      *canvas.Text
	anim        *fyne.Animation
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

	fmt.Println()
	fmt.Println("VoiceType Minimal is running!")
	fmt.Printf("Toggle with %s\n", strings.ToUpper(app.cfg.Hotkey))
	fmt.Println()

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
		app.pillBg.StrokeColor = color.RGBA{R: 236, G: 72, B: 153, A: 255} // Pink border when recording
		app.status.Text = "Listening"
		app.pillBg.Refresh()
		app.status.Refresh()
	})
	app.startWaveAnimation()

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

	if len(audioData) == 0 {
		app.safeUIUpdate(func() {
			app.status.Text = ""
			app.status.Refresh()
		})
		return
	}

	app.safeUIUpdate(func() {
		app.status.Text = "Transcribing"
		app.status.Refresh()
	})

	go func() {
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
			app.resetUI()
			return
		}

		text = strings.TrimSpace(text)
		if text == "" {
			app.resetUI()
			return
		}

		log.Printf("Transcribed: %s", text)

		if err := app.clip.SetAndPaste(app.ctx, text); err != nil {
			log.Printf("Paste failed: %v", err)
			return
		}

		app.safeUIUpdate(func() {
			app.status.Text = "✓"
			app.pillBg.StrokeColor = color.RGBA{R: 34, G: 197, B: 94, A: 255} // Green
			app.pillBg.Refresh()
			app.status.Refresh()
		})

		time.AfterFunc(1200*time.Millisecond, func() {
			app.resetUI()
		})
	}()
}

func (app *VoiceTypeApp) resetUI() {
	app.mu.Lock()
	defer app.mu.Unlock()
	if !app.isRecording {
		app.safeUIUpdate(func() {
			app.status.Text = ""
			// Revert to cream background/black border
			app.pillBg.StrokeColor = color.Black
			app.pillBg.Refresh()
			app.status.Refresh()
		})
	}
}

func (app *VoiceTypeApp) createWindow() {
	app.a.Settings().SetTheme(&ui.VoiceTypeTheme{})

	// Use a very unique Title to target with xprop/wmctrl reliably
	winTitle := fmt.Sprintf("VoiceTypeUI_%d", time.Now().UnixNano())
	app.window = app.a.NewWindow(winTitle)
	app.window.SetFixedSize(true)
	app.window.Resize(fyne.NewSize(260, 60))
	app.window.CenterOnScreen()

	// Image-inspired pill background (Cream/Off-white)
	cream := color.RGBA{R: 255, G: 254, B: 242, A: 255}
	pink := color.RGBA{R: 225, G: 90, B: 164, A: 255}

	app.pillBg = canvas.NewRectangle(cream)
	app.pillBg.StrokeWidth = 3
	app.pillBg.StrokeColor = pink
	app.pillBg.CornerRadius = 30

	// Status text (centered inside pill)
	app.status = canvas.NewText("", color.RGBA{R: 120, G: 120, B: 120, A: 255})
	app.status.TextSize = 12
	app.status.Alignment = fyne.TextAlignCenter

	// Waveform bars (Black)
	numBars := 18
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

	// Precise Layout: No padding, pure pill
	content := container.NewMax(
		app.pillBg,
		container.NewCenter(
			container.NewVBox(
				container.NewCenter(waveContainer),
				container.NewCenter(app.status),
			),
		),
	)

	app.window.SetContent(content)

	// Force borderless and stay on top using X11 tools
	go func() {
		// Wait for window to be mapped
		time.Sleep(100 * time.Millisecond)
		for i := 0; i < 5; i++ {
			app.stripDecorations(winTitle)
			time.Sleep(200 * time.Millisecond)
		}
	}()

	app.window.Show()
}

func (app *VoiceTypeApp) stripDecorations(title string) {
	// Remove window decorations (border, title bar, buttons)
	exec.Command("xprop", "-name", title, "-f", "_MOTIF_WM_HINTS", "32c", "-set", "_MOTIF_WM_HINTS", "0x2, 0x0, 0x0, 0x0, 0x0").Run()
	// Remove from taskbar and make it float on top
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

	app.anim = fyne.NewAnimation(time.Millisecond*50, func(f float32) {
		// Animate bars
		for i, bar := range app.waveBars {
			if bar == nil {
				continue
			}

			// Sine wave based on time and index for fluid motion
			elapsed := time.Since(startTime).Seconds()
			offset := float64(i) * 0.3
			h := 4 + 45*math.Abs(math.Sin(elapsed*7+offset))

			bar.Resize(fyne.NewSize(3, float32(h)))
			bar.Refresh()
		}
	})
	app.anim.RepeatCount = fyne.AnimationRepeatForever
	app.anim.Start()
	// Timer update
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
				seconds := int(elapsed.Seconds())
				minutes := seconds / 60
				seconds = seconds % 60

				app.safeUIUpdate(func() {
					app.status.Text = fmt.Sprintf("%d:%02d", minutes, seconds)
					app.status.Refresh()
				})
			case <-app.ctx.Done():
				return
			}
		}
	}()
}

func (app *VoiceTypeApp) stopWaveAnimation() {
	if app.anim != nil {
		app.anim.Stop()
		app.anim = nil
	}

	app.safeUIUpdate(func() {
		app.status.Text = ""
		app.status.Refresh()
		for _, bar := range app.waveBars {
			if bar != nil {
				bar.Hide()
				bar.Resize(fyne.NewSize(4, 10))
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
