// Package errors provides error handling utilities for VoiceType
package errors

import (
	"fmt"
	"log"
	"runtime"
	"sync"
)

// ErrorType represents the type of error
type ErrorType int

const (
	// ErrorTypeUnknown represents an unknown error
	ErrorTypeUnknown ErrorType = iota
	// ErrorTypeAudio represents audio-related errors
	ErrorTypeAudio
	// ErrorTypeAPI represents API-related errors
	ErrorTypeAPI
	// ErrorTypeTyping represents typing-related errors
	ErrorTypeTyping
	// ErrorTypeHotkey represents hotkey-related errors
	ErrorTypeHotkey
	// ErrorTypeUI represents UI-related errors
	ErrorTypeUI
	// ErrorTypeConfig represents configuration errors
	ErrorTypeConfig
	// ErrorTypeNetwork represents network-related errors
	ErrorTypeNetwork
)

// Error represents an error with metadata
type Error struct {
	Type    ErrorType
	Message string
	Err     error
	Stack   []byte
}

// NewError creates a new error with stack trace
func NewError(errType ErrorType, message string, err error) *Error {
	stack := make([]byte, 4096)
	n := runtime.Stack(stack, false)
	return &Error{
		Type:    errType,
		Message: message,
		Err:     err,
		Stack:   stack[:n],
	}
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Handler provides centralized error handling
type Handler struct {
	mu        sync.Mutex
	callbacks []func(*Error)
	logger    *log.Logger
}

// NewHandler creates a new error handler
func NewHandler() *Handler {
	return &Handler{
		logger: log.New(log.Writer(), "VoiceType: ", log.LstdFlags|log.Lshortfile),
	}
}

// OnError registers a callback for error events
func (h *Handler) OnError(callback func(*Error)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.callbacks = append(h.callbacks, callback)
}

// Handle processes an error
func (h *Handler) Handle(err error) {
	if err == nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Log the error
	h.logger.Print(err)

	// Notify callbacks
	for _, callback := range h.callbacks {
		go callback(NewError(ErrorTypeUnknown, "Error occurred", err))
	}
}

// Error logs an error message with format
func (h *Handler) Error(format string, args ...interface{}) {
	err := fmt.Errorf(format, args...)
	h.Handle(err)
}

// Fatal logs a fatal error and exits
func (h *Handler) Fatal(format string, args ...interface{}) {
	err := fmt.Errorf(format, args...)
	h.Handle(err)
	h.logger.Fatal("Fatal error, exiting...")
}

// Warning logs a warning message
func (h *Handler) Warning(format string, args ...interface{}) {
	h.logger.Printf("WARNING: "+format, args...)
}

// Info logs an info message
func (h *Handler) Info(format string, args ...interface{}) {
	h.logger.Printf("INFO: "+format, args...)
}

// Debug logs a debug message (only if verbose is enabled)
func (h *Handler) Debug(format string, args ...interface{}) {
	h.logger.Printf("DEBUG: "+format, args...)
}

// IsType checks if an error is of a specific type
func IsType(err error, errType ErrorType) bool {
	if e, ok := err.(*Error); ok {
		return e.Type == errType
	}
	return false
}

// Wrap wraps an error with additional context
func Wrap(err error, errType ErrorType, message string) *Error {
	if err == nil {
		return nil
	}
	return NewError(errType, message, err)
}

// Common errors
var (
	ErrNotSupported     = fmt.Errorf("operation not supported")
	ErrNotImplemented   = fmt.Errorf("not implemented")
	ErrPermissionDenied = fmt.Errorf("permission denied")
	ErrDeviceNotFound   = fmt.Errorf("device not found")
	ErrConnectionFailed = fmt.Errorf("connection failed")
	ErrTimeout          = fmt.Errorf("operation timed out")
	ErrAPIKeyMissing    = fmt.Errorf("API key is missing")
	ErrAPIKeyInvalid    = fmt.Errorf("API key is invalid")
	ErrRateLimited      = fmt.Errorf("rate limited")
	ErrAudioTooShort    = fmt.Errorf("audio recording is too short")
	ErrNoMicrophone     = fmt.Errorf("no microphone found")
	ErrTypingFailed     = fmt.Errorf("typing operation failed")
)
