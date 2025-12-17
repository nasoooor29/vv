package utils

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"visory/internal/database"
	"visory/internal/database/logs"

	"github.com/labstack/echo/v4"
)

type MySlog struct {
	*slog.Logger
	db     *database.Service
	Groups []string
}

func NewMySlog(logger *slog.Logger, db *database.Service) *MySlog {
	return &MySlog{
		Logger: logger,
		db:     db,
		Groups: []string{},
	}
}

func (m *MySlog) WithGroup(name string) *MySlog {
	newGroups := append(m.Groups, name)

	return &MySlog{
		Logger: m.Logger.WithGroup(name),
		db:     m.db,
		Groups: newGroups,
	}
}

func (m *MySlog) insertIntoDB(level string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err := m.db.Log.CreateLog(ctx, logs.CreateLogParams{
		UserID:       -1,
		Action:       "",
		Details:      new(string),
		ServiceGroup: "",
		Level:        level,
	})
	if err != nil {
		m.Logger.Error("Failed to insert log into DB", "err", err)
		return err
	}
	return nil
}

func (m *MySlog) NewHTTPError(status int, message string, internal error) *echo.HTTPError {
	levelStr, _ := statusCodeToLogLevel(status)
	m.insertIntoDB(levelStr)
	fmt.Printf("m.Groups: %v\n", m.Groups)
	return echo.NewHTTPError(status, message, "under the hell").SetInternal(internal)
}

func (m *MySlog) NewBadRequest(message string, internal error) *echo.HTTPError {
	return m.NewHTTPError(http.StatusBadRequest, message, internal)
}

func (m *MySlog) NewInternalServerError(message string, internal error) *echo.HTTPError {
	return m.NewHTTPError(http.StatusInternalServerError, message, internal)
}

func (m *MySlog) NewUnauthorized(message string, internal error) *echo.HTTPError {
	return m.NewHTTPError(http.StatusUnauthorized, message, internal)
}

func (m *MySlog) NewNotFound(message string, internal error) *echo.HTTPError {
	return m.NewHTTPError(http.StatusNotFound, message, internal)
}

func (m *MySlog) NewConflict(message string, internal error) *echo.HTTPError {
	return m.NewHTTPError(http.StatusConflict, message, internal)
}

func (m *MySlog) NewForbidden(message string, internal error) *echo.HTTPError {
	return m.NewHTTPError(http.StatusForbidden, message, internal)
}

func statusCodeToLogLevel(status int) (string, slog.Level) {
	switch {
	case status >= 500:
		return "ERROR", slog.LevelError
	case status >= 400:
		return "WARN", slog.LevelWarn
	default:
		return "INFO", slog.LevelInfo
	}
}
