// Package typing provides direct keyboard input functionality
package typing

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

// System handles direct keyboard input
type System struct{}

// NewSystem creates a new typing system
func NewSystem() *System {
	return &System{}
}

// TypeText simulates typing text directly at the cursor position
func (s *System) TypeText(ctx context.Context, text string, pressEnter bool) error {
	log.Printf("[Typing] Delivering transcription (%d chars) via Multi-Buffer Paste...", len(text))

	tCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	if err := s.SetPrimarySelection(tCtx, text); err != nil {
		log.Printf("[Typing] Clipboard set warning: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if err := s.PasteText(tCtx); err == nil {
		if pressEnter {
			time.Sleep(100 * time.Millisecond)
			_ = s.PressEnter(tCtx)
		}
		return nil
	}

	// 5. Fallback: Direct Typing (Only if separate buffer paste fails)
	log.Printf("[Typing] Auto-paste failed, falling back to direct typing")

	if s.isToolAvailable("ydotool") {
		if err := exec.CommandContext(tCtx, "ydotool", "type", text).Run(); err == nil {
			if pressEnter {
				time.Sleep(100 * time.Millisecond)
				_ = exec.CommandContext(tCtx, "ydotool", "key", "28:1", "28:0").Run()
			}
			return nil
		}
	}

	if s.isToolAvailable("wtype") {
		if err := exec.CommandContext(tCtx, "wtype", text).Run(); err == nil {
			if pressEnter {
				time.Sleep(100 * time.Millisecond)
				_ = exec.CommandContext(tCtx, "wtype", "-k", "Return").Run()
			}
			return nil
		}
	}

	if s.isToolAvailable("xdotool") {
		cmd := exec.CommandContext(tCtx, "xdotool", "type", "--clearmodifiers", "--delay", "2", text)
		if err := cmd.Run(); err == nil {
			if pressEnter {
				time.Sleep(100 * time.Millisecond)
				_ = s.PressEnter(tCtx)
			}
			return nil
		}
	}

	return fmt.Errorf("all typing/pasting methods failed")
}

// PasteText tries various methods to trigger a paste event
func (s *System) PasteText(ctx context.Context) error {
	isWayland := strings.Contains(os.Getenv("WAYLAND_DISPLAY"), "wayland")

	// Priority 1: Ctrl+V (Standard for most GUI apps)
	if isWayland && s.isToolAvailable("wtype") {
		if err := exec.CommandContext(ctx, "wtype", "-M", "ctrl", "-k", "v").Run(); err == nil {
			return nil
		}
	}
	if s.isToolAvailable("xdotool") {
		if err := exec.CommandContext(ctx, "xdotool", "key", "--clearmodifiers", "ctrl+v").Run(); err == nil {
			return nil
		}
	}

	// Priority 2: Shift+Insert (Standard for terminals and many X11 apps)
	if isWayland && s.isToolAvailable("wtype") {
		if err := exec.CommandContext(ctx, "wtype", "-M", "shift", "-k", "Insert").Run(); err == nil {
			return nil
		}
	}
	if s.isToolAvailable("xdotool") {
		if err := exec.CommandContext(ctx, "xdotool", "key", "--clearmodifiers", "shift+Insert").Run(); err == nil {
			return nil
		}
	}

	return fmt.Errorf("no paste trigger tool found or all failed")
}

// SetPrimarySelection copies text to BOTH Primary and Clipboard selections
func (s *System) SetPrimarySelection(ctx context.Context, text string) error {
	tCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	isWayland := strings.Contains(os.Getenv("WAYLAND_DISPLAY"), "wayland")

	if isWayland && s.isToolAvailable("wl-copy") {
		// Set both for Wayland
		_ = exec.CommandContext(tCtx, "wl-copy", text).Run()
		_ = exec.CommandContext(tCtx, "wl-copy", "--primary", text).Run()
		return nil
	}

	// X11 / XWayland
	if s.isToolAvailable("xclip") {
		_ = exec.CommandContext(tCtx, "xclip", "-selection", "clipboard", text).Run()
		_ = exec.CommandContext(tCtx, "xclip", "-selection", "primary", text).Run()
		return nil
	}
	if s.isToolAvailable("xsel") {
		_ = exec.CommandContext(tCtx, "xsel", "--clipboard", "--input", text).Run()
		_ = exec.CommandContext(tCtx, "xsel", "--primary", "--input", text).Run()
		return nil
	}

	return fmt.Errorf("no primary/clipboard selection tool found")
}

// WaitForFocus waits until the focus is no longer on a VoiceType window
func (s *System) WaitForFocus(ctx context.Context) {
	if !s.isToolAvailable("xdotool") {
		// Fallback to simple sleep if we can't verify focus
		time.Sleep(600 * time.Millisecond)
		return
	}

	// Wait up to 2 seconds for focus to shift away from VoiceType
	for i := 0; i < 20; i++ {
		cmd := exec.CommandContext(ctx, "xdotool", "getactivewindow", "getwindowname")
		output, err := cmd.Output()
		if err == nil {
			activeName := strings.ToLower(string(output))
			if !strings.Contains(activeName, "voicetype") {
				// Focus has shifted!
				time.Sleep(50 * time.Millisecond) // Tiny stabilization
				return
			}
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(100 * time.Millisecond):
		}
	}
}

// GetActiveWindowID returns the ID of the currently active window
func (s *System) GetActiveWindowID() string {
	if !s.isToolAvailable("xdotool") {
		return ""
	}
	out, err := exec.Command("xdotool", "getactivewindow").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// ActivateWindow restores focus to a specific window
func (s *System) ActivateWindow(id string) {
	if id == "" || !s.isToolAvailable("xdotool") {
		return
	}
	_ = exec.Command("xdotool", "windowactivate", "--sync", id).Run()
}

// PressEnter simulates pressing the Enter key
func (s *System) PressEnter(ctx context.Context) error {
	tCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if strings.Contains(os.Getenv("WAYLAND_DISPLAY"), "wayland") && s.isToolAvailable("wtype") {
		return exec.CommandContext(tCtx, "wtype", "-k", "Return").Run()
	}
	if s.isToolAvailable("xdotool") {
		return exec.CommandContext(tCtx, "xdotool", "key", "Return").Run()
	}
	return fmt.Errorf("no tool available to press Enter")
}

// isToolAvailable checks if a command-line tool exists
func (s *System) isToolAvailable(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}
