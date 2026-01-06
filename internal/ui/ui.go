// Package ui provides UI functionality for VoiceType
package ui

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"speek_to_text_linux/pkg/errors"
)

// UI represents the user interface system
type UI struct {
	errHandler *errors.Handler
	isVisible  bool
	windowPID  int
}

// NewUI creates a new UI system
func NewUI(errHandler *errors.Handler) *UI {
	return &UI{
		errHandler: errHandler,
	}
}

// Show shows the recording indicator
func (u *UI) Show() {
	if u.isVisible {
		return
	}

	if err := u.showIndicator(); err != nil {
		u.errHandler.Warning("Failed to show recording indicator: %v", err)
		return
	}

	u.isVisible = true
	log.Println("Recording indicator shown")
}

// Hide hides the recording indicator
func (u *UI) Hide() {
	if !u.isVisible {
		return
	}

	if err := u.hideIndicator(); err != nil {
		u.errHandler.Warning("Failed to hide recording indicator: %v", err)
		return
	}

	u.isVisible = false
	log.Println("Recording indicator hidden")
}

// showIndicator shows the recording indicator window
func (u *UI) showIndicator() error {
	// Try different methods based on desktop environment
	if strings.Contains(os.Getenv("WAYLAND_DISPLAY"), "wayland") {
		return u.showWaylandIndicator()
	}

	if strings.Contains(os.Getenv("DISPLAY"), ":") {
		return u.showX11Indicator()
	}

	// Fallback: use terminal bell or similar
	log.Println("Warning: No display detected, using fallback notification")
	return nil
}

// showX11Indicator shows an X11-based recording indicator
func (u *UI) showX11Indicator() error {
	// Try to use xdotool to create a small window or use yad/zenity
	if u.isToolAvailable("yad") {
		return u.showYadIndicator()
	}

	if u.isToolAvailable("zenity") {
		return u.showZenityIndicator()
	}

	// Fallback: create a simple indicator using xdotool
	return u.showXdotoolIndicator()
}

// showWaylandIndicator shows a Wayland-based recording indicator
func (u *UI) showWaylandIndicator() error {
	// For Wayland, we would use similar tools if available
	// Many Wayland compositors support similar tools
	log.Println("Warning: Wayland indicator not fully implemented")
	return nil
}

// showYadIndicator shows recording indicator using yad
func (u *UI) showYadIndicator() error {
	// yad (Yet Another Dialog) is a GTK+ dialog program
	cmd := exec.Command(
		"yad",
		"--notification",
		"--image=audio-input-microphone",
		"--text=Recording...",
		"--no-middle",
		"--timeout=3600", // 1 hour timeout (we'll hide it manually)
		"--kill-parent",
	)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start yad: %v", err)
	}

	u.windowPID = cmd.Process.Pid
	return nil
}

// showZenityIndicator shows recording indicator using zenity
func (u *UI) showZenityIndicator() error {
	// zenity is a dialog program using GTK+
	// We'll use a notification icon approach
	cmd := exec.Command(
		"zenity",
		"--notification",
		"--text=Recording...",
		"--window-icon=audio-input-microphone",
	)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start zenity: %v", err)
	}

	u.windowPID = cmd.Process.Pid
	return nil
}

// showXdotoolIndicator shows recording indicator using xdotool
func (u *UI) showXdotoolIndicator() error {
	// Use xdotool to get current window and show a notification
	cmd := exec.Command(
		"xdotool",
		"getactivewindow",
		"set_window", "--name", "VoiceType Recording",
	)

	if err := cmd.Run(); err != nil {
		log.Printf("Warning: Could not set window name: %v", err)
	}

	// Also try to create a small popup
	log.Println("Recording indicator active (using xdotool fallback)")
	return nil
}

// hideIndicator hides the recording indicator
func (u *UI) hideIndicator() error {
	if u.windowPID > 0 {
		// Kill the indicator process
		cmd := exec.Command("kill", fmt.Sprintf("%d", u.windowPID))
		if err := cmd.Run(); err != nil {
			log.Printf("Warning: Failed to kill indicator process: %v", err)
		}
		u.windowPID = 0
	}

	// Try to clean up any remaining processes
	if u.isToolAvailable("pkill") {
		exec.Command("pkill", "-f", "yad.*VoiceType").Run()
		exec.Command("pkill", "-f", "zenity.*VoiceType").Run()
	}

	return nil
}

// isToolAvailable checks if a tool is available
func (u *UI) isToolAvailable(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}

// Toggle toggles the recording indicator visibility
func (u *UI) Toggle() {
	if u.isVisible {
		u.Hide()
	} else {
		u.Show()
	}
}

// IsVisible returns whether the indicator is visible
func (u *UI) IsVisible() bool {
	return u.isVisible
}

// Close closes the UI system
func (u *UI) Close() {
	u.Hide()
	log.Println("UI system closed")
}

// SetStatus sets the status message
func (u *UI) SetStatus(status string) {
	// This would update the indicator text
	log.Printf("UI Status: %s", status)
}

// Pulse creates a pulsing animation effect (stub)
func (u *UI) Pulse() {
	// For a real implementation, we would animate the indicator
	log.Println("UI Pulse effect")
}

// Import time for timeout calculations
var _ = time.Second
