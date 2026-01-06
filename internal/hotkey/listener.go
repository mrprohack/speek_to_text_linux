package hotkey

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"VoiceType/pkg/errors"
)

type Listener struct {
	errHandler *errors.Handler
	hotkey     string
	onPress    func()
	onRelease  func()
	isRunning  bool
	mu         sync.Mutex
	stopChan   chan struct{}
}

func NewListener(errHandler *errors.Handler) *Listener {
	return &Listener{
		errHandler: errHandler,
	}
}

func (l *Listener) Initialize(hotkey string) error {
	l.hotkey = hotkey
	l.stopChan = make(chan struct{})

	if err := l.detectAndSetup(); err != nil {
		return errors.Wrap(err, errors.ErrorTypeHotkey, "failed to setup hotkey")
	}

	log.Printf("Hotkey listener initialized: %s", hotkey)
	return nil
}

func (l *Listener) detectAndSetup() error {
	if strings.Contains(os.Getenv("WAYLAND_DISPLAY"), "wayland") {
		// Try xdotool first (more reliable)
		if l.isToolAvailable("xdotool") {
			log.Println("Using xdotool for hotkey detection (Wayland)")
			go l.pollKeyPressX11()
			return nil
		}
		// Fall back to ydotool
		return l.setupWaylandHotkey()
	}

	if strings.Contains(os.Getenv("DISPLAY"), ":") {
		return l.setupX11Hotkey()
	}

	return l.setupGlobalHotkey()
}

func (l *Listener) setupX11Hotkey() error {
	if l.isToolAvailable("xdotool") {
		log.Println("Using xdotool for hotkey detection")
		go l.pollKeyPressX11()
		return nil
	}

	log.Println("Warning: No hotkey tool found. Install xdotool: sudo apt install xdotool")
	go l.pollKeyPressGeneric()
	return nil
}

func (l *Listener) pollKeyPressX11() {
	keyboardID := l.findKeyboardID()
	if keyboardID == "" {
		log.Println("Could not find keyboard ID, falling back to generic polling")
		l.pollKeyPressGeneric()
		return
	}

	// Dynamically find keycodes
	ctrlCodes := l.resolveKeycodes("Control_L", "Control_R")
	spaceCodes := l.resolveKeycodes("space")

	if len(ctrlCodes) == 0 || len(spaceCodes) == 0 {
		log.Printf("Warning: Could not resolve keycodes (ctrl: %v, space: %v), using defaults (37, 105 for Ctrl, 65 for Space)", ctrlCodes, spaceCodes)
		ctrlCodes = []string{"37", "105"} // Default for Ctrl_L, Ctrl_R
		spaceCodes = []string{"65"}       // Default for Space
	}

	log.Printf("Monitoring keyboard ID %s for hotkeys (Ctrl: %v, Space: %v)", keyboardID, ctrlCodes, spaceCodes)

	lastToggle := time.Now()
	isPressed := false

	for {
		select {
		case <-l.stopChan:
			return
		default:
		}

		// Query key state using xinput
		cmd := exec.Command("xinput", "query-state", keyboardID)
		output, _ := cmd.CombinedOutput()
		outputStr := string(output)

		ctrlDown := false
		for _, code := range ctrlCodes {
			if strings.Contains(outputStr, "key["+code+"]=down") {
				ctrlDown = true
				break
			}
		}

		spaceDown := false
		for _, code := range spaceCodes {
			if strings.Contains(outputStr, "key["+code+"]=down") {
				spaceDown = true
				break
			}
		}

		currentlyDown := ctrlDown && spaceDown

		if currentlyDown && !isPressed {
			// Key just pressed
			if time.Since(lastToggle) > 400*time.Millisecond {
				log.Println("Hotkey Detected: Ctrl + Space")
				l.firePress()
				isPressed = true
				lastToggle = time.Now()
			}
		} else if !currentlyDown && isPressed {
			// Key released
			isPressed = false
		}

		time.Sleep(40 * time.Millisecond)
	}
}

func (l *Listener) resolveKeycodes(names ...string) []string {
	var codes []string
	cmd := exec.Command("xmodmap", "-pk")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Warning: Could not run xmodmap to resolve keycodes: %v", err)
		return codes
	}

	lines := strings.Split(string(output), "\n")
	for _, name := range names {
		// Prepare common keysym variations
		patterns := []string{
			" " + name + " ",
			" " + name + ")",
			"(" + name + ")",
			"= " + name,
		}

		for _, line := range lines {
			matched := false
			for _, p := range patterns {
				if strings.Contains(line, p) {
					matched = true
					break
				}
			}

			if matched {
				fields := strings.Fields(line)
				if len(fields) < 2 {
					continue
				}
				// Handle both "keycode 37 = ..." and "37 0xffe3 (Control_L) ..."
				code := ""
				if fields[0] == "keycode" {
					code = fields[1]
				} else if _, err := fmt.Sscanf(fields[0], "%d", new(int)); err == nil {
					code = fields[0]
				}

				if code != "" {
					codes = append(codes, code)
				}
			}
		}
	}
	return codes
}

func (l *Listener) findKeyboardID() string {
	cmd := exec.Command("xinput", "list")
	output, _ := cmd.CombinedOutput()
	lines := strings.Split(string(output), "\n")

	// Prioritize slave keyboards which are actual devices
	for _, line := range lines {
		if strings.Contains(line, "slave") && (strings.Contains(line, "keyboard") || strings.Contains(line, "Keyboard")) &&
			!strings.Contains(line, "XTEST") &&
			strings.Contains(line, "id=") {
			parts := strings.Split(line, "id=")
			if len(parts) > 1 {
				idPart := strings.Fields(parts[1])[0]
				return idPart
			}
		}
	}

	// Fallback to master core keyboard (id=3 usually)
	return "3"
}

func (l *Listener) pollKeyPressGeneric() {
	log.Println("Warning: No hotkey detection method available")
	log.Println("Please install xdotool: sudo apt install xdotool")

	for {
		select {
		case <-l.stopChan:
			return
		default:
		}

		if l.isToolAvailable("xdotool") {
			log.Println("xdotool detected, switching to xdotool polling")
			go l.pollKeyPressX11()
			return
		}

		time.Sleep(1 * time.Second)
	}
}

func (l *Listener) setupWaylandHotkey() error {
	if l.isToolAvailable("ydotool") {
		log.Println("Using ydotool for Wayland hotkey detection")
		go l.pollKeyPressWayland()
		return nil
	}

	log.Println("Warning: Using generic polling for Wayland hotkey detection")
	log.Println("Please install ydotool for Wayland support")
	go l.pollKeyPressGeneric()
	return nil
}

func (l *Listener) pollKeyPressWayland() {
	keyName := l.hotkeyToXdotool(l.hotkey)
	log.Printf("Wayland polling for key: %s", keyName)

	prevPressed := false

	for {
		select {
		case <-l.stopChan:
			return
		default:
		}

		// Use ydotool to check if key is being pressed
		cmd := exec.Command("ydotool", "key", "--delay", "0", keyName)
		err := cmd.Run()

		isPressed := err == nil

		if isPressed && !prevPressed {
			log.Println("Hotkey pressed")
			l.firePress()
			prevPressed = true
		} else if !isPressed && prevPressed {
			log.Println("Hotkey released")
			l.fireRelease()
			prevPressed = false
		}

		time.Sleep(30 * time.Millisecond)
	}
}

func (l *Listener) setupGlobalHotkey() error {
	log.Println("Warning: No display detected, using fallback hotkey method")
	go l.pollKeyPressGeneric()
	return nil
}

func (l *Listener) hotkeyToXdotool(hotkey string) string {
	switch hotkey {
	case "Ctrl+Space", "ctrl+space":
		return "ctrl+space"
	case "Enter", "enter", "Return":
		return "Return"
	case "F5", "f5":
		return "F5"
	case "F6", "f6":
		return "F6"
	case "F12", "f12":
		return "F12"
	default:
		return hotkey
	}
}

func (l *Listener) isToolAvailable(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}

func (l *Listener) OnPress(callback func()) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.onPress = callback
}

func (l *Listener) OnRelease(callback func()) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.onRelease = callback
}

func (l *Listener) Start() error {
	l.mu.Lock()
	if l.isRunning {
		l.mu.Unlock()
		return fmt.Errorf("hotkey listener already running")
	}
	l.isRunning = true
	l.mu.Unlock()

	log.Printf("Started hotkey listener for key: %s", l.hotkey)
	return nil
}

func (l *Listener) Stop() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.isRunning {
		l.isRunning = false
		log.Println("Hotkey listener stopped")
	}
}

func (l *Listener) Close() {
	select {
	case <-l.stopChan:
	default:
		close(l.stopChan)
	}
	l.Stop()
}

func (l *Listener) IsRunning() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.isRunning
}

func (l *Listener) GetHotkey() string {
	return l.hotkey
}

func (l *Listener) firePress() {
	l.mu.Lock()
	callback := l.onPress
	l.mu.Unlock()
	if callback != nil {
		go callback()
	}
}

func (l *Listener) fireRelease() {
	l.mu.Lock()
	callback := l.onRelease
	l.mu.Unlock()
	if callback != nil {
		go callback()
	}
}
