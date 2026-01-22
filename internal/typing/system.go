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
	log.Printf("[Typing] Delivering transcription (%d chars) via Separate Buffer...", len(text))

	tCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// 1. Primary Method: Use "PRIMARY" selection (Separate Buffer)
	// This does NOT touch the Ctrl+C/Ctrl+V clipboard.
	if err := s.SetPrimarySelection(tCtx, text); err == nil {
		log.Printf("[Typing] Set separate buffer (PRIMARY). Triggering Shift+Insert...")
		// Small delay for OS to register selection
		time.Sleep(200 * time.Millisecond)

		// 1. Try wtype first (Native Wayland, no daemon needed)
		if s.isToolAvailable("wtype") {
			log.Printf("[Typing] Triggering Shift+Insert via wtype...")
			_ = exec.CommandContext(tCtx, "wtype", "-M", "shift", "-k", "Insert").Run()
			if pressEnter {
				time.Sleep(150 * time.Millisecond)
				_ = s.PressEnter(tCtx)
			}
			return nil
		}

		// 2. Try xdotool (Works for XWayland/X11)
		if s.isToolAvailable("xdotool") {
			log.Printf("[Typing] Triggering Shift+Insert via xdotool...")
			_ = exec.CommandContext(tCtx, "xdotool", "key", "--clearmodifiers", "shift+Insert").Run()
			if pressEnter {
				time.Sleep(150 * time.Millisecond)
				_ = s.PressEnter(tCtx)
			}
			return nil
		}

	}

	// 2. Fallback: Direct Typing (Only if separate buffer paste fails)
	log.Printf("[Typing] Separate buffer paste failed, falling back to direct typing")
	if s.isToolAvailable("ydotool") {
		if err := exec.CommandContext(tCtx, "ydotool", "type", text).Run(); err == nil {
			if pressEnter {
				time.Sleep(100 * time.Millisecond)
				_ = exec.CommandContext(tCtx, "ydotool", "key", "28:1", "28:0").Run()
			}
			return nil
		}
	}

	// 3. Try wtype (Wayland native)
	if s.isToolAvailable("wtype") {
		if err := exec.CommandContext(tCtx, "wtype", text).Run(); err == nil {
			if pressEnter {
				time.Sleep(100 * time.Millisecond)
				_ = exec.CommandContext(tCtx, "wtype", "-k", "Return").Run()
			}
			return nil
		}
	}

	// 4. Try xdotool (X11 / XWayland)
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

	return fmt.Errorf("all typing methods failed")
}

// SetPrimarySelection copies text to the system's PRIMARY selection (separate buffer)
func (s *System) SetPrimarySelection(ctx context.Context, text string) error {
	tCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if strings.Contains(os.Getenv("WAYLAND_DISPLAY"), "wayland") && s.isToolAvailable("wl-copy") {
		cmd = exec.CommandContext(tCtx, "wl-copy", "--primary")
	} else if s.isToolAvailable("xclip") {
		cmd = exec.CommandContext(tCtx, "xclip", "-selection", "primary")
	} else if s.isToolAvailable("xsel") {
		cmd = exec.CommandContext(tCtx, "xsel", "--primary", "--input")
	} else {
		return fmt.Errorf("no primary selection tool found")
	}

	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// SetClipboard copies text to the system clipboard
func (s *System) SetClipboard(ctx context.Context, text string) error {
	tCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if strings.Contains(os.Getenv("WAYLAND_DISPLAY"), "wayland") && s.isToolAvailable("wl-copy") {
		cmd = exec.CommandContext(tCtx, "wl-copy")
	} else if s.isToolAvailable("xclip") {
		cmd = exec.CommandContext(tCtx, "xclip", "-selection", "clipboard")
	} else if s.isToolAvailable("xsel") {
		cmd = exec.CommandContext(tCtx, "xsel", "--clipboard", "--input")
	} else {
		return fmt.Errorf("no clipboard tool found (install wl-copy, xclip or xsel)")
	}

	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
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
