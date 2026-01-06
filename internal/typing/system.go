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
func (s *System) TypeText(ctx context.Context, text string) error {
	// Typing timeout
	tCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Wayland: use wtype
	if strings.Contains(os.Getenv("WAYLAND_DISPLAY"), "wayland") {
		if s.isToolAvailable("wtype") {
			cmd := exec.CommandContext(tCtx, "wtype", text)
			return cmd.Run()
		}
	}

	// X11: use xdotool
	if s.isToolAvailable("xdotool") {
		cmd := exec.CommandContext(tCtx, "xdotool", "type", "--clearmodifiers", "--delay", "1", text)
		return cmd.Run()
	}

	return fmt.Errorf("no typing tool (wtype or xdotool) available")
}

// isToolAvailable checks if a command-line tool exists
func (s *System) isToolAvailable(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}
