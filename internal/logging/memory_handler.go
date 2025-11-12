package logging

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/jarv/newsgoat/internal/database"
)

// LogEvent represents a log event for subscribers
type LogEvent struct {
	Message   *database.LogMessage
	Timestamp time.Time
}

// MemoryHandler implements slog.Handler and stores logs in memory
type MemoryHandler struct {
	debugEnabled bool
	logs         []*database.LogMessage
	maxLogs      int
	events       chan LogEvent
	mutex        sync.RWMutex
	nextID       int64
}

// NewMemoryHandler creates a new in-memory logging handler
func NewMemoryHandler(maxLogs int) *MemoryHandler {
	return &MemoryHandler{
		debugEnabled: false,
		logs:         make([]*database.LogMessage, 0),
		maxLogs:      maxLogs,
		events:       make(chan LogEvent, 100), // Buffered channel for events
		nextID:       1,
	}
}

// NewMemoryHandlerWithDebug creates a new handler with debug mode enabled
func NewMemoryHandlerWithDebug(maxLogs int, debug bool) *MemoryHandler {
	return &MemoryHandler{
		debugEnabled: debug,
		logs:         make([]*database.LogMessage, 0),
		maxLogs:      maxLogs,
		events:       make(chan LogEvent, 100),
		nextID:       1,
	}
}

func (h *MemoryHandler) Enabled(_ context.Context, level slog.Level) bool {
	// Filter out debug messages unless debug mode is enabled
	if level == slog.LevelDebug && !h.debugEnabled {
		return false
	}
	return true
}

func (h *MemoryHandler) Handle(ctx context.Context, r slog.Record) error {
	// Collect all attributes into a map
	attrs := make(map[string]interface{})
	r.Attrs(func(a slog.Attr) bool {
		// Special handling for error types - convert to string
		if a.Key == "error" {
			if err, ok := a.Value.Any().(error); ok {
				attrs[a.Key] = err.Error()
			} else {
				attrs[a.Key] = a.Value.String()
			}
		} else {
			attrs[a.Key] = a.Value.Any()
		}
		return true
	})

	// Add source location if available
	if r.PC != 0 {
		frames := runtime.CallersFrames([]uintptr{r.PC})
		frame, _ := frames.Next()
		if frame.File != "" {
			attrs["source_file"] = frame.File
			attrs["source_line"] = frame.Line
		}
	}

	// Convert attributes to JSON
	var attributesJSON sql.NullString
	if len(attrs) > 0 {
		jsonData, err := json.Marshal(attrs)
		if err != nil {
			return err
		}
		attributesJSON = sql.NullString{String: string(jsonData), Valid: true}
	}

	// Create log message
	h.mutex.Lock()
	logMsg := &database.LogMessage{
		ID:         h.nextID,
		Level:      r.Level.String(),
		Message:    r.Message,
		Timestamp:  sql.NullTime{Time: r.Time, Valid: true},
		Attributes: attributesJSON,
	}
	h.nextID++

	// Add to logs slice
	h.logs = append(h.logs, logMsg)

	// Trim logs if we exceed maxLogs
	if len(h.logs) > h.maxLogs {
		h.logs = h.logs[len(h.logs)-h.maxLogs:]
	}
	h.mutex.Unlock()

	// Publish event
	select {
	case h.events <- LogEvent{
		Message:   logMsg,
		Timestamp: time.Now(),
	}:
	default:
		// Event channel is full, drop the event
	}

	return nil
}

func (h *MemoryHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// For simplicity, we'll return the same handler
	return h
}

func (h *MemoryHandler) WithGroup(_ string) slog.Handler {
	// For simplicity, we'll return the same handler
	return h
}

// GetLogMessages returns the last N log messages
func (h *MemoryHandler) GetLogMessages(limit int) ([]*database.LogMessage, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	// Return logs in reverse chronological order (newest first)
	result := make([]*database.LogMessage, 0, limit)
	startIdx := len(h.logs) - limit
	if startIdx < 0 {
		startIdx = 0
	}

	// Copy logs in reverse order
	for i := len(h.logs) - 1; i >= startIdx; i-- {
		result = append(result, h.logs[i])
	}

	return result, nil
}

// GetLogMessage returns a single log message by ID
func (h *MemoryHandler) GetLogMessage(id int64) (*database.LogMessage, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, log := range h.logs {
		if log.ID == id {
			return log, nil
		}
	}

	return nil, nil
}

// DeleteAllLogMessages clears all log messages
func (h *MemoryHandler) DeleteAllLogMessages() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.logs = make([]*database.LogMessage, 0)
	return nil
}

// Subscribe returns a channel for log events
func (h *MemoryHandler) Subscribe() <-chan LogEvent {
	return h.events
}

// Close closes the event channel
func (h *MemoryHandler) Close() {
	close(h.events)
}
