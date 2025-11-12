package logging

import (
	"log/slog"

	"github.com/jarv/newsgoat/internal/database"
)

var logger *slog.Logger
var memoryHandler *MemoryHandler

// SetLogger sets the global logger instance
func SetLogger(l *slog.Logger) {
	logger = l
}

// SetMemoryHandler sets the global memory handler instance
func SetMemoryHandler(h *MemoryHandler) {
	memoryHandler = h
}

// GetLogger returns the global logger instance
func GetLogger() *slog.Logger {
	return logger
}

// GetMemoryHandler returns the global memory handler instance
func GetMemoryHandler() *MemoryHandler {
	return memoryHandler
}

// GetLogMessages returns the last N log messages from the in-memory handler
func GetLogMessages(limit int) ([]*database.LogMessage, error) {
	if memoryHandler == nil {
		return []*database.LogMessage{}, nil
	}
	return memoryHandler.GetLogMessages(limit)
}

// GetLogMessage returns a single log message by ID from the in-memory handler
func GetLogMessage(id int64) (*database.LogMessage, error) {
	if memoryHandler == nil {
		return nil, nil
	}
	return memoryHandler.GetLogMessage(id)
}

// DeleteAllLogMessages clears all log messages from the in-memory handler
func DeleteAllLogMessages() error {
	if memoryHandler == nil {
		return nil
	}
	return memoryHandler.DeleteAllLogMessages()
}

// Subscribe returns a channel for log events from the in-memory handler
func Subscribe() <-chan LogEvent {
	if memoryHandler == nil {
		ch := make(chan LogEvent)
		close(ch)
		return ch
	}
	return memoryHandler.Subscribe()
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