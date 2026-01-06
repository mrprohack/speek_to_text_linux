package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// Log levels
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

var levelNames = []string{"DEBUG", "INFO", "WARN", "ERROR"}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

type Logger struct {
	level  LogLevel
	output io.Writer
	debug  bool
}

// NewLogger creates a new structured logger
func NewLogger(level LogLevel, output io.Writer, debug bool) *Logger {
	return &Logger{
		level:  level,
		output: output,
		debug:  debug,
	}
}

// log logs a message if it matches the logger's level
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	msg := fmt.Sprintf(format, args...)
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     levelNames[level],
		Message:   msg,
	}

	// JSON output
	data, _ := json.Marshal(entry)
	l.output.Write(data)
	l.output.Write([]byte("\n"))

	// Console output for errors
	if level == LogLevelError {
		fmt.Fprintln(os.Stderr, msg)
	}
}

// Convenience methods
func (l *Logger) Debug(format string, args ...interface{}) { l.log(LogLevelDebug, format, args...) }
func (l *Logger) Info(format string, args ...interface{})  { l.log(LogLevelInfo, format, args...) }
func (l *Logger) Warn(format string, args ...interface{})  { l.log(LogLevelWarn, format, args...) }
func (l *Logger) Error(format string, args ...interface{}) { l.log(LogLevelError, format, args...) }
func (l *Logger) SetLevel(level LogLevel)                  { l.level = level }

// Global default logger
var defaultLogger *Logger

func init() {
	defaultLogger = NewLogger(LogLevelInfo, os.Stdout, false)
}

// SetDebug enables debug mode
func SetDebug(debug bool) {
	if defaultLogger != nil {
		defaultLogger.debug = debug
		if debug {
			defaultLogger.level = LogLevelDebug
		}
	}
}

// Global convenience functions
func Debug(format string, args ...interface{}) { defaultLogger.log(LogLevelDebug, format, args...) }
func Info(format string, args ...interface{})  { defaultLogger.log(LogLevelInfo, format, args...) }
func Warn(format string, args ...interface{})  { defaultLogger.log(LogLevelWarn, format, args...) }
func Error(format string, args ...interface{}) { defaultLogger.log(LogLevelError, format, args...) }
