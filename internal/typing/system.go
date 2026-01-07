// Package typing provides direct keyboard input functionality
package typing

import (
	"context"
	"fmt"
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
	// Typing timeout
	tCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Wayland: use wtype
	if strings.Contains(os.Getenv("WAYLAND_DISPLAY"), "wayland") {
		if s.isToolAvailable("wtype") {
			// Type the text first
			if err := exec.CommandContext(tCtx, "wtype", text).Run(); err != nil {
				return err
			}
			if pressEnter {
				time.Sleep(100 * time.Millisecond)
				return exec.CommandContext(tCtx, "wtype", "-k", "Return").Run()
			}
			return nil
		}
	}

	// X11: use xdotool
	if s.isToolAvailable("xdotool") {
		// Type the text
		cmd := exec.CommandContext(tCtx, "xdotool", "type", "--clearmodifiers", "--delay", "1", text)
		if err := cmd.Run(); err != nil {
			return err
		}
		// Press enter if requested
		if pressEnter {
			time.Sleep(100 * time.Millisecond)
			return exec.CommandContext(tCtx, "xdotool", "key", "Return").Run()
		}
		return nil
	}

	return fmt.Errorf("no typing tool (wtype or xdotool) available")
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
