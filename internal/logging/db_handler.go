package logging

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"runtime"

	"github.com/jarv/newsgoat/internal/database"
)

type DatabaseHandler struct {
	queries      *database.Queries
	debugEnabled bool
}

func NewDatabaseHandler(queries *database.Queries) *DatabaseHandler {
	return &DatabaseHandler{
		queries:      queries,
		debugEnabled: false,
	}
}

func NewDatabaseHandlerWithDebug(queries *database.Queries, debug bool) *DatabaseHandler {
	return &DatabaseHandler{
		queries:      queries,
		debugEnabled: debug,
	}
}

func (h *DatabaseHandler) Enabled(_ context.Context, level slog.Level) bool {
	// Filter out debug messages unless debug mode is enabled
	if level == slog.LevelDebug && !h.debugEnabled {
		return false
	}
	return true
}

func (h *DatabaseHandler) Handle(ctx context.Context, r slog.Record) error {
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

	// Store the log message in the database
	return h.queries.CreateLogMessage(ctx, database.CreateLogMessageParams{
		Level:      r.Level.String(),
		Message:    r.Message,
		Timestamp:  sql.NullTime{Time: r.Time, Valid: true},
		Attributes: attributesJSON,
	})
}

func (h *DatabaseHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// For simplicity, we'll return the same handler
	// In a more sophisticated implementation, you might want to
	// store these attributes to be added to all future log messages
	return h
}

func (h *DatabaseHandler) WithGroup(_ string) slog.Handler {
	// For simplicity, we'll return the same handler
	return h
}