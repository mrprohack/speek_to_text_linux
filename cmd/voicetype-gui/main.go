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

	"io"
	"speek_to_text_linux/internal/api"
	"speek_to_text_linux/internal/audio"
	"speek_to_text_linux/internal/hotkey"
	"speek_to_text_linux/internal/typing"
	"speek_to_text_linux/internal/ui"
	"speek_to_text_linux/pkg/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var version = "1.0.0"

const configFile = ".voicetype.conf"

type VoiceTypeApp struct {
	a            fyne.App
	cfg          *config.Config
	audioSys     *audio.System
	apiClient    *api.Client
	typer        *typing.System
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
	winPosX      int
	winPosY      int
}

type draggableBackground struct {
	widget.BaseWidget
	app *VoiceTypeApp
}

func newDraggableBackground(a *VoiceTypeApp) *draggableBackground {
	d := &draggableBackground{app: a}
	d.ExtendBaseWidget(d)
	return d
}

func (d *draggableBackground) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(d.app.pillBg)
}

func (d *draggableBackground) Dragged(e *fyne.DragEvent) {
	d.app.mu.Lock()
	d.app.winPosX += int(e.Dragged.DX)
	d.app.winPosY += int(e.Dragged.DY)
	x, y := d.app.winPosX, d.app.winPosY
	title := d.app.winTitle
	d.app.mu.Unlock()

	// Use wmctrl to move the window in real-time
	exec.Command("wmctrl", "-r", title, "-e", fmt.Sprintf("0,%d,%d,-1,-1", x, y)).Run()
}

func (d *draggableBackground) DragEnd() {
	// Optional: Save position to config
}

func (d *draggableBackground) TappedSecondary(e *fyne.PointEvent) {
	d.app.showSettingsWindow()
}

func main() {
	initLogger()
	flagHelp := flag.Bool("help", false, "Show help")
	flagDevice := flag.String("device", "", "Audio device")
	flagToggle := flag.Bool("toggle", false, "Toggle recording on a running instance")
	flagStop := flag.Bool("stop", false, "Stop a running instance")
	flagNoReturn := flag.Bool("no-return", false, "Don't press Enter after typing")
	flagSettings := flag.Bool("settings", false, "Show settings window")
	flag.Parse()

	if *flagHelp {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}

	cfg, _ := config.Load()

	if *flagSettings {
		apiKey := cfg.GROQ_API_KEY
		if apiKey == "" {
			apiKey = os.Getenv("GROQ_API_KEY")
		}
		app := &VoiceTypeApp{
			a:   app.NewWithID("com.voicetype.app"),
			cfg: cfg,
		}
		app.audioSys = audio.NewSystem(nil)
		app.showSettingsWindow()
		app.a.Run()
		os.Exit(0)
	}

	if *flagNoReturn {
		cfg.AutoReturn = false
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

	cfg.AudioDevice = *flagDevice
	if *flagNoReturn {
		cfg.AutoReturn = false
	}
	log.Printf("Config loaded: AutoReturn=%v", cfg.AutoReturn)

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
		app.status.Text = ""
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
		// Smooth fade-in effect
		go app.fadeInWindow()

		// New UI uses specific status colors per section
		app.status.Text = ""
		app.statusIcon.Hide()
		app.status.Refresh()
	})
	app.startWaveAnimation()
	app.startPulseAnimation(color.RGBA{R: 255, G: 255, B: 255, A: 255})

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
		app.status.Text = ""
		app.statusIcon.Hide()
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
				app.pillBg.StrokeColor = color.RGBA{R: 239, G: 68, B: 68, A: 255} // Crimson
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
			app.status.Text = ""
			app.status.Refresh()

			// Capture target window ID before hiding
			prevWindowID := app.typer.GetActiveWindowID()

			// Hide window immediately to return focus to the target software
			app.window.Hide()

			// Actively restore focus to the previous window
			if prevWindowID != "" {
				app.typer.ActivateWindow(prevWindowID)
			}
		})

		// Shorter delay since we actively restore focus
		time.Sleep(500 * time.Millisecond)

		if err := app.typer.TypeText(app.ctx, text, app.cfg.AutoReturn); err != nil {
			log.Printf("Typing failed: %v", err)
			app.safeUIUpdate(func() {
				app.a.Quit()
			})
			return
		}

		app.safeUIUpdate(func() {
			app.status.Text = ""
			app.statusIcon.Hide()
			app.status.Refresh()
			app.statusIcon.Refresh()
		})

		time.Sleep(600 * time.Millisecond)
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
			// Smooth fade out animation
			go func() {
				app.fadeOutWindow()
				app.safeUIUpdate(func() {
					app.status.Text = ""
					app.status.Refresh()
					app.window.Hide()
				})
			}()
		})
	}
}

func (app *VoiceTypeApp) createWindow() {
	app.a.Settings().SetTheme(&ui.VoiceTypeTheme{})

	screenSize := fyne.NewSize(1920, 1080)
	if d, ok := app.a.Driver().(interface {
		AllScreens() []interface {
			Size() fyne.Size
		}
	}); ok {
		screens := d.AllScreens()
		if len(screens) > 0 {
			screenSize = screens[0].Size()
		}
	}

	pillWidth := float32(120.0)
	pillHeight := float32(28.0)

	app.winTitle = fmt.Sprintf("VoiceTypeUI_%d", time.Now().UnixNano())
	app.window = app.a.NewWindow(app.winTitle)
	app.window.SetFixedSize(true)
	app.window.SetPadded(false)
	app.window.Resize(fyne.NewSize(pillWidth, pillHeight))

	app.winPosX = int((screenSize.Width - pillWidth) / 2)
	app.winPosY = int(screenSize.Height - pillHeight - 60)

	numBars := 12
	app.waveBars = make([]*canvas.Rectangle, numBars)

	hbox := layout.NewHBoxLayout()
	waveContainer := container.New(hbox)

	for i := 0; i < numBars; i++ {
		bar := canvas.NewRectangle(color.Transparent)
		bar.SetMinSize(fyne.NewSize(2, 2))
		bar.CornerRadius = 1.0
		bar.Hide()
		app.waveBars[i] = bar
		waveContainer.Add(container.NewCenter(bar))
	}

	app.status = canvas.NewText("", color.Transparent)
	app.status.Hide()
	app.statusIcon = canvas.NewImageFromResource(theme.InfoIcon())
	app.statusIcon.Hide()

	app.pillBg = canvas.NewRectangle(color.RGBA{R: 15, G: 15, B: 20, A: 210})
	app.pillBg.CornerRadius = pillHeight / 2
	app.pillBg.StrokeWidth = 1.0
	app.pillBg.StrokeColor = color.RGBA{R: 255, G: 255, B: 255, A: 30}

	app.glowLayers = make([]*canvas.Rectangle, 0)

	bg := newDraggableBackground(app)

	content := container.NewStack(
		bg,
		container.NewCenter(waveContainer),
	)

	app.window.SetContent(content)

	app.window.Canvas().SetOnTypedKey(func(k *fyne.KeyEvent) {
		if k.Name == fyne.KeyEscape {
			app.a.Quit()
		}
	})

	go func() {
		time.Sleep(100 * time.Millisecond)
		for i := 0; i < 10; i++ {
			app.stripDecorations(app.winTitle)
			exec.Command("wmctrl", "-r", app.winTitle, "-e", fmt.Sprintf("0,%d,%d,-1,-1", app.winPosX, app.winPosY)).Run()
			time.Sleep(200 * time.Millisecond)
			if i == 5 {
				app.safeUIUpdate(func() {
					app.window.SetTitle("")
				})
			}
		}
	}()

	app.window.Show()
}

func (app *VoiceTypeApp) stripDecorations(title string) {
	if title == "" {
		return
	}
	var winID string
	out, err := exec.Command("xdotool", "search", "--name", title).Output()
	if err == nil {
		winID = strings.TrimSpace(string(out))
		if strings.Contains(winID, "\n") {
			winID = strings.Split(winID, "\n")[0]
		}
	}

	if winID != "" {
		exec.Command("xprop", "-id", winID, "-f", "_MOTIF_WM_HINTS", "32c", "-set", "_MOTIF_WM_HINTS", "0x2, 0x0, 0x0, 0x0, 0x0").Run()
		exec.Command("xprop", "-id", winID, "-f", "_NET_WM_WINDOW_TYPE", "32a", "-set", "_NET_WM_WINDOW_TYPE", "_NET_WM_WINDOW_TYPE_NOTIFICATION").Run()
		exec.Command("xprop", "-id", winID, "-f", "_NET_WM_STATE", "32a", "-set", "_NET_WM_STATE", "_NET_WM_STATE_SKIP_TASKBAR,_NET_WM_STATE_SKIP_PAGER,_NET_WM_STATE_ABOVE,_NET_WM_STATE_STAY_ON_TOP").Run()
		exec.Command("xprop", "-id", winID, "-f", "_NET_WM_ALLOWED_ACTIONS", "32a", "-set", "_NET_WM_ALLOWED_ACTIONS", "").Run()
	} else {
		exec.Command("xprop", "-name", title, "-f", "_MOTIF_WM_HINTS", "32c", "-set", "_MOTIF_WM_HINTS", "0x2, 0x0, 0x0, 0x0, 0x0").Run()
		exec.Command("xprop", "-name", title, "-f", "_NET_WM_WINDOW_TYPE", "32a", "-set", "_NET_WM_WINDOW_TYPE", "_NET_WM_WINDOW_TYPE_NOTIFICATION").Run()
		exec.Command("wmctrl", "-r", title, "-b", "add,above,skip_taskbar,skip_pager").Run()
	}
}

func (app *VoiceTypeApp) fadeInWindow() {
	steps := 10
	for i := 0; i <= steps; i++ {
		opacity := float64(i) / float64(steps)
		exec.Command("xprop", "-name", app.winTitle, "-f", "_NET_WM_WINDOW_OPACITY", "32c", "-set", "_NET_WM_WINDOW_OPACITY", fmt.Sprintf("%d", uint32(opacity*0xFFFFFFFF))).Run()
		time.Sleep(20 * time.Millisecond)
	}
}

func (app *VoiceTypeApp) fadeOutWindow() {
	steps := 8
	for i := steps; i >= 0; i-- {
		opacity := float64(i) / float64(steps)
		exec.Command("xprop", "-name", app.winTitle, "-f", "_NET_WM_WINDOW_OPACITY", "32c", "-set", "_NET_WM_WINDOW_OPACITY", fmt.Sprintf("%d", uint32(opacity*0xFFFFFFFF))).Run()
		time.Sleep(25 * time.Millisecond)
	}
}

func (app *VoiceTypeApp) safeUIUpdate(f func()) {
	fyne.Do(f)
}

func (app *VoiceTypeApp) smoothColorTransition(targetColor color.RGBA, duration time.Duration) {
	steps := 15
	stepDuration := duration / time.Duration(steps)

	startColor := app.status.Color
	startR, startG, startB, startA := startColor.RGBA()

	for i := 0; i <= steps; i++ {
		progress := float64(i) / float64(steps)

		// Interpolate colors
		r := uint8(float64(startR>>8) + (float64(targetColor.R)-float64(startR>>8))*progress)
		g := uint8(float64(startG>>8) + (float64(targetColor.G)-float64(startG>>8))*progress)
		b := uint8(float64(startB>>8) + (float64(targetColor.B)-float64(startB>>8))*progress)
		a := uint8(float64(startA>>8) + (float64(targetColor.A)-float64(startA>>8))*progress)

		newColor := color.RGBA{R: r, G: g, B: b, A: a}

		app.safeUIUpdate(func() {
			app.status.Color = newColor
			app.status.Refresh()
		})

		time.Sleep(stepDuration)
	}
}

func (app *VoiceTypeApp) startWaveAnimation() {
	app.safeUIUpdate(func() {
		for _, bar := range app.waveBars {
			bar.Show()
		}
	})

	startTime := time.Now()

	app.anim = fyne.NewAnimation(time.Millisecond*16, func(f float32) {
		app.mu.Lock()
		level := app.audioSys.GetLevel()
		app.smoothLevel = app.smoothLevel*0.7 + level*0.3
		app.mu.Unlock()

		center := len(app.waveBars) / 2
		for i, bar := range app.waveBars {
			if bar == nil {
				continue
			}
			elapsed := time.Since(startTime).Seconds()

			dist := float64(math.Abs(float64(i - center)))
			maxDist := float64(center)

			falloff := math.Exp(-math.Pow(dist/(maxDist*0.8), 2.0))

			idle := 1.0 * math.Sin(elapsed*3.5+float64(i)*0.15)
			idle += 0.5 * math.Sin(elapsed*2.0+float64(i)*0.25)

			vocal := app.smoothLevel * 40.0 * falloff

			h := 2.5 + math.Abs(idle) + vocal
			if h > 18 {
				h = 18
			}

			opacity := uint8(180 + 75*(h/18.0))
			barColor := color.RGBA{R: 255, G: 255, B: 255, A: opacity}

			bar.FillColor = barColor
			bar.Resize(fyne.NewSize(1.5, float32(h)))
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
	app.pulseAnim = fyne.NewAnimation(time.Duration(float64(time.Second)*2.0), func(f float32) {
		app.safeUIUpdate(func() {
			val := math.Sin(float64(f)*2*math.Pi - math.Pi/2)
			normVal := (val + 1) / 2 // 0 to 1

			// Breathing border
			app.pillBg.StrokeWidth = 1.0 + 0.8*float32(normVal)
			alphaBase := uint8(30 + 50*normVal)
			app.pillBg.StrokeColor = color.RGBA{R: pulseColor.R, G: pulseColor.G, B: pulseColor.B, A: alphaBase}

			// Multi-layered bloom glow
			for i, glow := range app.glowLayers {
				layerFactor := float64(len(app.glowLayers) - i)
				glowAlpha := uint8((15 + 25*normVal) * (layerFactor / float64(len(app.glowLayers))))
				glow.StrokeColor = color.RGBA{R: pulseColor.R, G: pulseColor.G, B: pulseColor.B, A: glowAlpha}
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

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		key := strings.TrimSpace(scanner.Text())
		if key == "" {
			log.Fatal("Error: API key is required")
		}
		return key
	}
	return ""
}

func (app *VoiceTypeApp) showSettingsWindow() {
	w := app.a.NewWindow("VoiceType Settings")

	keyEntry := widget.NewPasswordEntry()
	keyEntry.SetText(app.cfg.GROQ_API_KEY)
	keyEntry.SetPlaceHolder("gsk_...")

	hotkeyEntry := widget.NewEntry()
	hotkeyEntry.SetText(app.cfg.Hotkey)
	hotkeyEntry.SetPlaceHolder("ctrl+space")

	devices := app.audioSys.GetDevices()
	deviceSelect := widget.NewSelect(devices, nil)
	deviceSelect.SetSelected(app.cfg.AudioDevice)
	if deviceSelect.Selected == "" && len(devices) > 0 {
		deviceSelect.SetSelected("default")
	}

	autoReturnCheck := widget.NewCheck("Auto-press Enter after typing", nil)
	autoReturnCheck.Checked = app.cfg.AutoReturn

	models := []string{"whisper-large-v3", "distil-whisper-large-v3-en"}
	modelSelect := widget.NewSelect(models, nil)
	modelSelect.SetSelected(app.cfg.Model)
	if modelSelect.Selected == "" {
		modelSelect.SetSelected("whisper-large-v3")
	}

	form := widget.NewForm(
		widget.NewFormItem("GROQ API Key", keyEntry),
		widget.NewFormItem("Hotkey", hotkeyEntry),
		widget.NewFormItem("Audio Device", deviceSelect),
		widget.NewFormItem("Model", modelSelect),
		widget.NewFormItem("", autoReturnCheck),
	)

	saveBtn := widget.NewButton("Save & Exit", func() {
		app.cfg.GROQ_API_KEY = keyEntry.Text
		app.cfg.Hotkey = hotkeyEntry.Text
		app.cfg.AudioDevice = deviceSelect.Selected
		app.cfg.AutoReturn = autoReturnCheck.Checked
		app.cfg.Model = modelSelect.Selected

		if err := app.cfg.Save(""); err != nil {
			log.Printf("Failed to save config: %v", err)
		} else {
			log.Println("Config saved successfully")
		}
		app.a.Quit()
	})
	saveBtn.Importance = widget.HighImportance

	title := widget.NewLabelWithStyle("VoiceType Settings", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		form,
		layout.NewSpacer(),
		saveBtn,
	)

	bg := canvas.NewRectangle(color.RGBA{R: 25, G: 25, B: 30, A: 255})
	w.SetContent(container.NewStack(bg, container.NewPadded(content)))

	w.Resize(fyne.NewSize(420, 320))
	w.SetFixedSize(true)
	w.CenterOnScreen()

	go func() {
		for i := 0; i < 5; i++ {
			exec.Command("wmctrl", "-r", "VoiceType Settings", "-b", "add,above").Run()
			time.Sleep(200 * time.Millisecond)
		}
	}()

	w.Show()
}

func initLogger() {
	path, err := config.GetConfigPath()
	if err != nil {
		return
	}
	logPath := filepath.Join(filepath.Dir(path), "debug.log")

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Warning: Failed to open log file: %v\n", err)
		return
	}

	// Write to both stdout and file
	multiWriter := io.MultiWriter(os.Stdout, file)
	log.SetOutput(multiWriter)
	log.Printf("--- Logging started at %v ---", time.Now().Format(time.RFC3339))
	log.Printf("Log file: %s", logPath)
}
