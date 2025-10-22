package logger

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

var (
	currentLevel = InfoLevel
	mu           sync.RWMutex
)

// Initialize sets the global log level
func Initialize(level LogLevel) {
	mu.Lock()
	defer mu.Unlock()
	currentLevel = level
}

// ParseLevel converts a string to LogLevel
func ParseLevel(levelStr string) LogLevel {
	switch levelStr {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	default:
		return InfoLevel
	}
}

// formatLog creates a properly formatted log message with aligned emojis
// Format: YYYY-MM-DD HH:MM:SS emoji   message
// Using fixed spaces after emoji to ensure perfect alignment
func formatLog(emoji string, args ...any) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	msg := formatMessage(args...)
	// Use exactly 3 spaces after emoji for consistent alignment
	return fmt.Sprintf("%s %s   %s", timestamp, emoji, msg)
}

// formatLogf creates a properly formatted log message with formatting
func formatLogf(emoji string, format string, args ...any) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf(format, args...)
	// Use exactly 3 spaces after emoji for consistent alignment
	return fmt.Sprintf("%s %s   %s", timestamp, emoji, msg)
}

// Debug logs a debug message
func Debug(args ...any) {
	mu.RLock()
	defer mu.RUnlock()
	if currentLevel <= DebugLevel {
		fmt.Fprintln(os.Stderr, formatLog("🔍", args...))
	}
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...any) {
	mu.RLock()
	defer mu.RUnlock()
	if currentLevel <= DebugLevel {
		fmt.Fprintln(os.Stderr, formatLogf("🔍", format, args...))
	}
}

// Info logs an info message
func Info(args ...any) {
	mu.RLock()
	defer mu.RUnlock()
	if currentLevel <= InfoLevel {
		fmt.Fprintln(os.Stderr, formatLog("ℹ️", args...))
	}
}

// Infof logs a formatted info message
func Infof(format string, args ...any) {
	mu.RLock()
	defer mu.RUnlock()
	if currentLevel <= InfoLevel {
		fmt.Fprintln(os.Stderr, formatLogf("ℹ️", format, args...))
	}
}

// Warn logs a warning message
func Warn(args ...any) {
	mu.RLock()
	defer mu.RUnlock()
	if currentLevel <= WarnLevel {
		fmt.Fprintln(os.Stderr, formatLog("⚠️", args...))
	}
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...any) {
	mu.RLock()
	defer mu.RUnlock()
	if currentLevel <= WarnLevel {
		fmt.Fprintln(os.Stderr, formatLogf("⚠️", format, args...))
	}
}

// Error logs an error message
func Error(args ...any) {
	mu.RLock()
	defer mu.RUnlock()
	if currentLevel <= ErrorLevel {
		fmt.Fprintln(os.Stderr, formatLog("❌", args...))
	}
}

// Errorf logs a formatted error message
func Errorf(format string, args ...any) {
	mu.RLock()
	defer mu.RUnlock()
	if currentLevel <= ErrorLevel {
		fmt.Fprintln(os.Stderr, formatLogf("❌", format, args...))
	}
}

// Fatal logs a fatal message and exits the program
func Fatal(args ...any) {
	fmt.Fprintln(os.Stderr, formatLog("🛑", args...))
	os.Exit(1)
}

// Fatalf logs a formatted fatal message and exits the program
func Fatalf(format string, args ...any) {
	fmt.Fprintln(os.Stderr, formatLogf("🛑", format, args...))
	os.Exit(1)
}

// Success logs a success message
func Success(args ...any) {
	mu.RLock()
	defer mu.RUnlock()
	if currentLevel <= InfoLevel {
		fmt.Fprintln(os.Stderr, formatLog("✅", args...))
	}
}

// Successf logs a formatted success message
func Successf(format string, args ...any) {
	mu.RLock()
	defer mu.RUnlock()
	if currentLevel <= InfoLevel {
		fmt.Fprintln(os.Stderr, formatLogf("✅", format, args...))
	}
}

// formatMessage converts variadic arguments to a single message string
func formatMessage(args ...any) string {
	if len(args) == 0 {
		return ""
	}
	return fmt.Sprint(args...)
}
