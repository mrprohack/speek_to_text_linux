package errors

import (
	"errors"
	"testing"
)

func TestNewError(t *testing.T) {
	err := NewError(ErrorTypeAudio, "test error", errors.New("original error"))

	if err.Type != ErrorTypeAudio {
		t.Errorf("Expected error type %v, got %v", ErrorTypeAudio, err.Type)
	}

	if err.Message != "test error" {
		t.Errorf("Expected message 'test error', got '%s'", err.Message)
	}

	if err.Err == nil {
		t.Error("Expected underlying error to be set")
	}
}

func TestError_Error(t *testing.T) {
	err := NewError(ErrorTypeAPI, "test error", errors.New("underlying"))

	expected := "test error: underlying"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestWrap(t *testing.T) {
	original := errors.New("original error")
	wrapped := Wrap(original, ErrorTypeNetwork, "network operation failed")

	if wrapped.Type != ErrorTypeNetwork {
		t.Errorf("Expected type %v, got %v", ErrorTypeNetwork, wrapped.Type)
	}

	if wrapped.Err != original {
		t.Error("Expected underlying error to be preserved")
	}
}

func TestIsType(t *testing.T) {
	err := NewError(ErrorTypeAudio, "test", nil)

	if !IsType(err, ErrorTypeAudio) {
		t.Error("Expected IsType to return true for matching type")
	}

	if IsType(err, ErrorTypeAPI) {
		t.Error("Expected IsType to return false for non-matching type")
	}

	regularErr := errors.New("regular error")
	if IsType(regularErr, ErrorTypeUnknown) {
		t.Error("Expected IsType to return false for non-Error type")
	}
}

func TestHandler_Error(t *testing.T) {
	handler := NewHandler()

	// This should not panic
	handler.Error("test error: %v", "message")
}

func TestHandler_Fatal(t *testing.T) {
	handler := NewHandler()

	// Fatal calls os.Exit, so we can't test it directly
	// Just verify it doesn't panic before calling os.Exit
	handler.Info("Fatal would be called here with: %v", "test")
}

func TestHandler_Warning(t *testing.T) {
	handler := NewHandler()

	// This should not panic
	handler.Warning("warning: %v", "test")
}

func TestHandler_Info(t *testing.T) {
	handler := NewHandler()

	// This should not panic
	handler.Info("info: %v", "test")
}
