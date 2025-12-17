package revenium

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Logger interface defines the logging methods
// debug only shows if REVENIUM_DEBUG=true or config.debug=true
type Logger interface {
	Debug(message string, args ...interface{})
	Info(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Error(message string, args ...interface{})
}

// DefaultLogger is the default console logger implementation
type DefaultLogger struct {
}

// NewDefaultLogger creates a new default logger
func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{}
}

// Debug logs a debug message
// Debug messages are ONLY shown if:
// 1. The global config has Debug=true, OR
// 2. The REVENIUM_DEBUG environment variable is set to "true"
func (l *DefaultLogger) Debug(message string, args ...interface{}) {
	if getGlobalDebugFlag() || os.Getenv("REVENIUM_DEBUG") == "true" {
		l.log("Debug", message, args...)
	}
}

// Info logs an info message (always shown)
func (l *DefaultLogger) Info(message string, args ...interface{}) {
	l.log("", message, args...)
}

// Warn logs a warning message (always shown)
func (l *DefaultLogger) Warn(message string, args ...interface{}) {
	l.log("Warning", message, args...)
}

// Error logs an error message (always shown)
func (l *DefaultLogger) Error(message string, args ...interface{}) {
	l.log("Error", message, args...)
}

// log is the internal logging method
func (l *DefaultLogger) log(level, message string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	var prefix string
	if level == "" {
		prefix = fmt.Sprintf("[%s] [Revenium]", timestamp)
	} else {
		prefix = fmt.Sprintf("[%s] [Revenium %s]", timestamp, level)
	}

	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}

	log.Printf("%s %s", prefix, message)
}

// Global logger instance
var globalLogger Logger = NewDefaultLogger()

// GetLogger returns the global logger instance
func GetLogger() Logger {
	return globalLogger
}

// SetLogger sets a custom global logger
func SetLogger(logger Logger) {
	globalLogger = logger
}

// InitializeLogger initializes the logger from environment variables
// Only debug mode can be toggled via REVENIUM_DEBUG
func InitializeLogger() {
	// Logger is ready to use, no initialization needed
	// Debug mode is controlled by REVENIUM_DEBUG environment variable
}

// Convenience functions for global logger
func Debug(message string, args ...interface{}) {
	globalLogger.Debug(message, args...)
}

func Info(message string, args ...interface{}) {
	globalLogger.Info(message, args...)
}

func Warn(message string, args ...interface{}) {
	globalLogger.Warn(message, args...)
}

func Error(message string, args ...interface{}) {
	globalLogger.Error(message, args...)
}

// getGlobalDebugFlag returns the debug flag from the global configuration
// This is used by the logger to check if debug mode is enabled programmatically
func getGlobalDebugFlag() bool {
	// This function will be implemented to access the global config
	// For now, we'll use a package-level variable that can be set
	return globalDebugEnabled
}

// Package-level variable to track debug state
var globalDebugEnabled bool

// SetGlobalDebug sets the global debug flag
// This is called when the configuration is initialized
func SetGlobalDebug(enabled bool) {
	globalDebugEnabled = enabled
}
