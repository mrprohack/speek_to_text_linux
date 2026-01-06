// Package clipboard provides clipboard and paste functionality for VoiceType
package clipboard

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"VoiceType/pkg/errors"
)

// System represents the clipboard system
type System struct {
	errHandler *errors.Handler
}

// NewSystem creates a new clipboard system
func NewSystem(errHandler *errors.Handler) *System {
	return &System{
		errHandler: errHandler,
	}
}

// SetAndPaste sets text to clipboard and pastes it
func (s *System) SetAndPaste(ctx context.Context, text string) error {
	// Copy to clipboard
	if err := s.SetClipboard(text); err != nil {
		return errors.Wrap(err, errors.ErrorTypeClipboard, "failed to set clipboard")
	}

	// Small delay to ensure clipboard is set
	time.Sleep(50 * time.Millisecond)

	// Simulate paste
	if err := s.Paste(); err != nil {
		return errors.Wrap(err, errors.ErrorTypeClipboard, "failed to paste")
	}

	return nil
}

// SetClipboard sets text to the system clipboard
func (s *System) SetClipboard(text string) error {
	// Try different clipboard tools
	tools := []string{"xclip", "xsel", "wl-copy", "pbpaste"}

	for _, tool := range tools {
		if s.isToolAvailable(tool) {
			return s.setClipboardWithTool(tool, text)
		}
	}

	// Fallback: try to detect desktop environment
	return s.setClipboardFallback(text)
}

// isToolAvailable checks if a tool is available
func (s *System) isToolAvailable(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}

// setClipboardWithTool sets clipboard using a specific tool
func (s *System) setClipboardWithTool(tool, text string) error {
	var cmd *exec.Cmd

	switch tool {
	case "xclip":
		cmd = exec.Command("xclip", "-selection", "clipboard")
	case "xsel":
		cmd = exec.Command("xsel", "--clipboard", "--input")
	case "wl-copy":
		cmd = exec.Command("wl-copy")
	case "pbpaste":
		// pbpaste is for reading, not writing - skip this tool
		return fmt.Errorf("pbpaste cannot be used for writing")
	}

	if cmd == nil {
		return fmt.Errorf("unsupported tool: %s", tool)
	}

	cmd.Stdin = strings.NewReader(text)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set clipboard with %s: %v, output: %s", tool, err, output)
	}

	log.Printf("Set clipboard using %s", tool)
	return nil
}

// setClipboardFallback attempts to set clipboard using available methods
func (s *System) setClipboardFallback(text string) error {
	// Try to detect Wayland or X11
	if strings.Contains(os.Getenv("WAYLAND_DISPLAY"), "wayland") {
		// Try wl-copy again with full path
		return s.setClipboardWithTool("wl-copy", text)
	}

	// Try X11 tools
	if strings.Contains(os.Getenv("DISPLAY"), ":") {
		if s.isToolAvailable("xclip") {
			return s.setClipboardWithTool("xclip", text)
		}
		if s.isToolAvailable("xsel") {
			return s.setClipboardWithTool("xsel", text)
		}
	}

	return fmt.Errorf("no clipboard tool available")
}

// Paste simulates a paste operation
func (s *System) Paste() error {
	// Try different paste methods
	tools := []string{"xdotool", "xte", "wtype"}

	for _, tool := range tools {
		if s.isToolAvailable(tool) {
			return s.pasteWithTool(tool)
		}
	}

	// Fallback: try to detect and use appropriate method
	return s.pasteFallback()
}

// pasteWithTool simulates paste using a specific tool
func (s *System) pasteWithTool(tool string) error {
	var cmd *exec.Cmd

	switch tool {
	case "xdotool":
		// xdotool key ctrl+v
		cmd = exec.Command("xdotool", "key", "ctrl+v")
	case "xte":
		// xte key 'Control+v'
		cmd = exec.Command("xte", "key", "Control+v")
	case "wtype":
		// wtype -P ctrl -k v
		cmd = exec.Command("wtype", "-P", "ctrl", "-k", "v")
	default:
		return fmt.Errorf("unsupported tool: %s", tool)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to paste with %s: %v, output: %s", tool, err, output)
	}

	log.Printf("Performed paste using %s", tool)
	return nil
}

// pasteFallback attempts to paste using available methods
func (s *System) pasteFallback() error {
	// Try to detect Wayland or X11
	if strings.Contains(os.Getenv("WAYLAND_DISPLAY"), "wayland") {
		if s.isToolAvailable("wtype") {
			return s.pasteWithTool("wtype")
		}
	}

	// Try X11 tools
	if strings.Contains(os.Getenv("DISPLAY"), ":") {
		if s.isToolAvailable("xdotool") {
			return s.pasteWithTool("xdotool")
		}
		if s.isToolAvailable("xte") {
			return s.pasteWithTool("xte")
		}
	}

	return fmt.Errorf("no paste tool available")
}

// GetClipboard gets text from the clipboard (for testing)
func (s *System) GetClipboard() (string, error) {
	// Try different tools
	tools := []string{"xclip", "xsel", "wl-paste", "pbpaste"}

	for _, tool := range tools {
		if s.isToolAvailable(tool) {
			return s.getClipboardWithTool(tool)
		}
	}

	return "", fmt.Errorf("no clipboard read tool available")
}

// getClipboardWithTool reads clipboard using a specific tool
func (s *System) getClipboardWithTool(tool string) (string, error) {
	var cmd *exec.Cmd

	switch tool {
	case "xclip":
		cmd = exec.Command("xclip", "-selection", "clipboard", "-o")
	case "xsel":
		cmd = exec.Command("xsel", "--clipboard", "--output")
	case "wl-paste":
		cmd = exec.Command("wl-paste")
	case "pbpaste":
		cmd = exec.Command("pbpaste")
	default:
		return "", fmt.Errorf("unsupported tool: %s", tool)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to read clipboard with %s: %v", tool, err)
	}

	return string(output), nil
}

// Cleanup cleans up resources
func (s *System) Cleanup() {
	// No cleanup needed for command-based approach
}

// Import os for error checking
var _ = fmt.Sprintf
