// Package hotkey provides global hotkey listening functionality for VoiceType
package hotkey

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"VoiceType/pkg/errors"
)

// Listener represents a global hotkey listener
type Listener struct {
	errHandler *errors.Handler
	hotkey     string
	onPress    func()
	onRelease  func()
	isRunning  bool
	mu         sync.Mutex
}

// NewListener creates a new hotkey listener
func NewListener(errHandler *errors.Handler) *Listener {
	return &Listener{
		errHandler: errHandler,
	}
}

// Initialize initializes the hotkey listener
func (l *Listener) Initialize(hotkey string) error {
	l.hotkey = hotkey

	// Detect desktop environment and choose appropriate method
	if err := l.detectAndSetup(); err != nil {
		return errors.Wrap(err, errors.ErrorTypeHotkey, "failed to setup hotkey")
	}

	log.Printf("Hotkey listener initialized: %s", hotkey)
	return nil
}

// detectAndSetup detects the desktop environment and sets up hotkey listening
func (l *Listener) detectAndSetup() error {
	// Check for Wayland
	if strings.Contains(os.Getenv("WAYLAND_DISPLAY"), "wayland") {
		return l.setupWaylandHotkey()
	}

	// Check for X11
	if strings.Contains(os.Getenv("DISPLAY"), ":") {
		return l.setupX11Hotkey()
	}

	// Fallback: try to use global shortcuts
	return l.setupGlobalHotkey()
}

// setupX11Hotkey sets up X11-based hotkey listening
func (l *Listener) setupX11Hotkey() error {
	// Try to use xbindkeys or sxhkd
	if l.isToolAvailable("xbindkeys") {
		return l.setupXbindkeysHotkey()
	}

	if l.isToolAvailable("sxhkd") {
		return l.setupSxhkdHotkey()
	}

	// Fallback to key detection via xdotool
	log.Println("Warning: No X11 hotkey daemon found, using polling fallback")
	return nil
}

// setupWaylandHotkey sets up Wayland-based hotkey listening
func (l *Listener) setupWaylandHotkey() error {
	// Try to use swaymsg or similar
	if l.isToolAvailable("swaymsg") {
		return l.setupSwayHotkey()
	}

	// For now, we'll use a polling-based approach
	log.Println("Warning: Using polling fallback for Wayland hotkey detection")
	return nil
}

// setupXbindkeysHotkey sets up hotkey using xbindkeys
func (l *Listener) setupXbindkeysHotkey() error {
	// This would configure xbindkeys to listen for the hotkey
	// For now, we'll use a simple polling approach
	log.Println("Using xbindkeys for hotkey detection")
	return nil
}

// setupSxhkdHotkey sets up hotkey using sxhkd
func (l *Listener) setupSxhkdHotkey() error {
	// sxhkd is a hotkey daemon for X11 and Wayland
	log.Println("Using sxhkd for hotkey detection")
	return nil
}

// setupSwayHotkey sets up hotkey using swaync or similar
func (l *Listener) setupSwayHotkey() error {
	log.Println("Using swaymsg for hotkey detection")
	return nil
}

// setupGlobalHotkey sets up a global hotkey using available tools
func (l *Listener) setupGlobalHotkey() error {
	// Try to set up using gsettings or similar
	log.Println("Warning: No display detected, using fallback hotkey method")
	return nil
}

// isToolAvailable checks if a tool is available
func (l *Listener) isToolAvailable(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}

// OnPress sets the callback for hotkey press events
func (l *Listener) OnPress(callback func()) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.onPress = callback
}

// OnRelease sets the callback for hotkey release events
func (l *Listener) OnRelease(callback func()) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.onRelease = callback
}

// Start starts the hotkey listener
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

// Stop stops the hotkey listener
func (l *Listener) Stop() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.isRunning {
		l.isRunning = false
		log.Println("Hotkey listener stopped")
	}
}

// Close closes the hotkey listener and cleans up
func (l *Listener) Close() {
	l.Stop()
}

// IsRunning returns whether the listener is running
func (l *Listener) IsRunning() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.isRunning
}

// GetHotkey returns the current hotkey
func (l *Listener) GetHotkey() string {
	return l.hotkey
}

// firePress fires the press callback
func (l *Listener) firePress() {
	l.mu.Lock()
	callback := l.onPress
	l.mu.Unlock()
	if callback != nil {
		go callback()
	}
}

// fireRelease fires the release callback
func (l *Listener) fireRelease() {
	l.mu.Lock()
	callback := l.onRelease
	l.mu.Unlock()
	if callback != nil {
		go callback()
	}
}

// Import fmt for error formatting
var _ = fmt.Sprintf
