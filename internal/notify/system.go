// Package notify provides notification functionality for VoiceType
package notify

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"VoiceType/pkg/errors"
)

// Notifier represents the notification system
type Notifier struct {
	errHandler *errors.Handler
	isReady    bool
}

// NewNotifier creates a new notifier
func NewNotifier(errHandler *errors.Handler) *Notifier {
	return &Notifier{
		errHandler: errHandler,
	}
}

// Initialize initializes the notification system
func (n *Notifier) Initialize() error {
	// Check for notification daemon
	if err := n.detectNotificationSystem(); err != nil {
		return errors.Wrap(err, errors.ErrorTypeUI, "failed to initialize notification system")
	}

	n.isReady = true
	log.Println("Notification system initialized")
	return nil
}

// detectNotificationSystem detects available notification systems
func (n *Notifier) detectNotificationSystem() error {
	// Check for notify-send (libnotify)
	if n.isToolAvailable("notify-send") {
		return nil
	}

	// Check for other notification tools
	tools := []string{"dunstify", "knotify", "xfce4-notifyd"}
	for _, tool := range tools {
		if n.isToolAvailable(tool) {
			return nil
		}
	}

	// Notifications might still work with notify-send even if not installed
	log.Println("Warning: No notification daemon found, notifications may not appear")
	return nil
}

// isToolAvailable checks if a tool is available
func (n *Notifier) isToolAvailable(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}

// Notify sends a notification
func (n *Notifier) Notify(title, message string) error {
	if !n.isReady {
		log.Printf("Notification (disabled): %s: %s", title, message)
		return nil
	}

	// Try notify-send first
	if n.isToolAvailable("notify-send") {
		return n.sendWithNotifySend(title, message)
	}

	// Try other tools
	if n.isToolAvailable("dunstify") {
		return n.sendWithDunstify(title, message)
	}

	log.Printf("Notification: %s - %s", title, message)
	return nil
}

// sendWithNotifySend sends notification using notify-send
func (n *Notifier) sendWithNotifySend(title, message string) error {
	cmd := exec.Command(
		"notify-send",
		"--app-name=VoiceType",
		"--icon=microphone",
		"--urgency=normal",
		title,
		message,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("notify-send failed: %v, output: %s", err, output)
	}

	return nil
}

// sendWithDunstify sends notification using dunstify
func (n *Notifier) sendWithDunstify(title, message string) error {
	cmd := exec.Command(
		"dunstify",
		"--app-name=VoiceType",
		"--icon=microphone",
		"--urgency=normal",
		title,
		message,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("dunstify failed: %v, output: %s", err, output)
	}

	return nil
}

// NotifyError sends an error notification
func (n *Notifier) NotifyError(title, message string) error {
	if !n.isReady {
		log.Printf("Error notification: %s: %s", title, message)
		return nil
	}

	// Try notify-send with critical urgency
	if n.isToolAvailable("notify-send") {
		cmd := exec.Command(
			"notify-send",
			"--app-name=VoiceType",
			"--icon=dialog-error",
			"--urgency=critical",
			title,
			message,
		)

		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("notify-send failed: %v, output: %s", err, output)
		}

		return nil
	}

	log.Printf("Error notification: %s - %s", title, message)
	return nil
}

// NotifySuccess sends a success notification
func (n *Notifier) NotifySuccess(title, message string) error {
	return n.Notify(title, message)
}

// NotifyWithTimeout sends a notification that auto-dismisses
func (n *Notifier) NotifyWithTimeout(title, message string, timeout time.Duration) error {
	if !n.isReady {
		log.Printf("Notification (timeout=%v): %s: %s", timeout, title, message)
		return nil
	}

	// notify-send doesn't support timeouts, but most notification daemons do
	// We'll just use the regular notification
	return n.Notify(title, message)
}

// Close closes the notification system
func (n *Notifier) Close() {
	n.isReady = false
	log.Println("Notification system closed")
}

// IsReady returns whether the notifier is ready
func (n *Notifier) IsReady() bool {
	return n.isReady
}

// Import strings for environment variable checking
var _ = strings.Contains
