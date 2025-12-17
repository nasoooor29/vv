package utils

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"visory/internal/database"
	"visory/internal/database/logs"
)

// LoggerConfig holds the configuration for creating a database-backed logger
type LoggerConfig struct {
	DB           *database.Service
	ServiceGroup string
	Level        slog.Level
}

// DatabaseLogHandler is a handler that logs to both stdout and database
type DatabaseLogHandler struct {
	handler      slog.Handler
	db           *database.Service
	serviceGroup string
	userID       int64
}

// NewDatabaseLogHandler creates a new database log handler
func NewDatabaseLogHandler(handler slog.Handler, db *database.Service, serviceGroup string) *DatabaseLogHandler {
	return &DatabaseLogHandler{
		handler:      handler,
		db:           db,
		serviceGroup: serviceGroup,
		userID:       0, // Will be set by middleware
	}
}

// SetUserID sets the user ID for log entries
func (h *DatabaseLogHandler) SetUserID(userID int64) {
	h.userID = userID
}

// Enabled implements slog.Handler
func (h *DatabaseLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// Handle implements slog.Handler
func (h *DatabaseLogHandler) Handle(ctx context.Context, rec slog.Record) error {
	// Log to stdout first
	if err := h.handler.Handle(ctx, rec); err != nil {
		return err
	}

	// Log to database asynchronously
	go h.logToDatabase(ctx, rec)

	return nil
}

// logToDatabase inserts log entry into database
func (h *DatabaseLogHandler) logToDatabase(ctx context.Context, rec slog.Record) {
	// Don't log to database if no user ID (could log with user_id 0 or skip)
	if h.userID <= 0 {
		return
	}

	// Build details from attributes
	details := rec.Message
	if rec.NumAttrs() > 0 {
		details += " | "
		rec.Attrs(func(a slog.Attr) bool {
			details += fmt.Sprintf("%s: %v | ", a.Key, a.Value)
			return true
		})
	}

	// Create log entry in database
	_, err := h.db.Log.CreateLog(ctx, logs.CreateLogParams{
		UserID:       h.userID,
		Action:       rec.Message,
		Details:      &details,
		ServiceGroup: h.serviceGroup,
		Level:        rec.Level.String(),
	})
	if err != nil {
		// If we can't insert to database, just continue - don't panic
		fmt.Fprintf(os.Stderr, "failed to insert log to database: %v\n", err)
	}
}

// WithAttrs implements slog.Handler
func (h *DatabaseLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &DatabaseLogHandler{
		handler:      h.handler.WithAttrs(attrs),
		db:           h.db,
		serviceGroup: h.serviceGroup,
		userID:       h.userID,
	}
}

// WithGroup implements slog.Handler
func (h *DatabaseLogHandler) WithGroup(name string) slog.Handler {
	return &DatabaseLogHandler{
		handler:      h.handler.WithGroup(name),
		db:           h.db,
		serviceGroup: h.serviceGroup,
		userID:       h.userID,
	}
}

// NewServiceLogger creates a new logger for a service with grouping
func NewServiceLogger(db *database.Service, serviceGroup string) *slog.Logger {
	// Create the DaLog handler with the service group in the style function
	daLogHandler := &DaLog{
		defaultHandler: slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
		Writer: os.Stdout,
		HandlerOptions: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
		style: DaLogStyleLongType1,
	}

	// Wrap it with database logging
	dbHandler := NewDatabaseLogHandler(daLogHandler, db, serviceGroup)

	// Create logger with service group
	logger := slog.New(dbHandler)

	// Add service group as a default attribute
	logger = logger.With("service", serviceGroup)

	return logger
}

// ExtractUserIDFromContext extracts user ID from request context
func ExtractUserIDFromContext(ctx context.Context) int64 {
	userID, ok := ctx.Value("userID").(int64)
	if ok {
		return userID
	}

	// Try to parse from environment or other sources if needed
	userIDStr := os.Getenv("USER_ID")
	if userIDStr != "" {
		if id, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			return id
		}
	}

	return 0
}

// LogRequestDetails logs detailed HTTP request information to database
func LogRequestDetails(ctx context.Context, db *database.Service, userID int64, method, path string, statusCode int, duration time.Duration, remoteAddr string) error {
	details := fmt.Sprintf("Method: %s, Path: %s, Status: %d, Duration: %dms, Remote: %s",
		method, path, statusCode, duration.Milliseconds(), remoteAddr)

	action := fmt.Sprintf("HTTP %s %s", method, path)

	_, err := db.Log.CreateLog(ctx, logs.CreateLogParams{
		UserID:       userID,
		Action:       action,
		Details:      &details,
		ServiceGroup: "http",
		Level:        getLevelFromStatusCode(statusCode),
	})

	return err
}

// getLevelFromStatusCode determines log level based on HTTP status code
func getLevelFromStatusCode(statusCode int) string {
	switch {
	case statusCode >= 500:
		return "ERROR"
	case statusCode >= 400:
		return "WARN"
	case statusCode >= 300:
		return "INFO"
	default:
		return "DEBUG"
	}
}
