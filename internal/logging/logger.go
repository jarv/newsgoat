package logging

import (
	"log/slog"
)

var logger *slog.Logger

// SetLogger sets the global logger instance
func SetLogger(l *slog.Logger) {
	logger = l
}

// GetLogger returns the global logger instance
func GetLogger() *slog.Logger {
	return logger
}

// Info logs at info level
func Info(msg string, args ...any) {
	if logger != nil {
		logger.Info(msg, args...)
	}
}

// Warn logs at warn level
func Warn(msg string, args ...any) {
	if logger != nil {
		logger.Warn(msg, args...)
	}
}

// Error logs at error level
func Error(msg string, args ...any) {
	if logger != nil {
		logger.Error(msg, args...)
	}
}

// Debug logs at debug level
func Debug(msg string, args ...any) {
	if logger != nil {
		logger.Debug(msg, args...)
	}
}