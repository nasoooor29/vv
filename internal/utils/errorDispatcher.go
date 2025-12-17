package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"visory/internal/database"
	"visory/internal/database/logs"
	"visory/internal/models"

	"github.com/labstack/echo/v4"
)

type ErrorDispatcher struct {
	db     *database.Service
	Groups []string
}

func NewErrorDispatcher(logger *slog.Logger, db *database.Service) *ErrorDispatcher {
	return &ErrorDispatcher{
		db:     db,
		Groups: []string{},
	}
}

func (m *ErrorDispatcher) WithGroup(name string) *ErrorDispatcher {
	newGroups := append(m.Groups, name)

	return &ErrorDispatcher{
		db:     m.db,
		Groups: newGroups,
	}
}

func (m *ErrorDispatcher) InsertIntoDB(data models.LogRequestData) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	jsonifiedDetails, err := json.Marshal(data)
	if err != nil {
		slog.Error("error happened", "err", err)
		return err
	}
	details := string(jsonifiedDetails)
	grp := "???"
	if len(m.Groups) > 0 {
		grp = m.Groups[len(m.Groups)-1]
	}

	levelStr, _ := StatusCodeToLogLevel(data.Status)

	_, err = m.db.Log.CreateLog(ctx, logs.CreateLogParams{
		UserID:       data.UserId,
		Action:       fmt.Sprintf("%v %v", data.Method, data.Uri),
		Details:      &details,
		ServiceGroup: grp,
		Level:        levelStr,
	})
	if err != nil {
		slog.Error("error inserting log into DB", "err", err)
		return err
	}
	return nil
}

func (m *ErrorDispatcher) NewHTTPError(status int, message string, internal error, otherInfo ...any) *echo.HTTPError {
	// Convert otherInfo to strings and join them
	var otherErrs []error
	for _, info := range otherInfo {
		otherErrs = append(otherErrs, errors.New(fmt.Sprint(info)))
	}

	// Join all errors: internal and the others
	joinedErr := errors.Join(append([]error{internal}, otherErrs...)...)

	return echo.NewHTTPError(status, message).SetInternal(joinedErr)
}

func (m *ErrorDispatcher) NewBadRequest(message string, internal error, otherInfo ...any) *echo.HTTPError {
	return m.NewHTTPError(http.StatusBadRequest, message, internal, otherInfo...)
}

func (m *ErrorDispatcher) NewInternalServerError(message string, internal error, otherInfo ...any) *echo.HTTPError {
	return m.NewHTTPError(http.StatusInternalServerError, message, internal, otherInfo...)
}

func (m *ErrorDispatcher) NewUnauthorized(message string, internal error, otherInfo ...any) *echo.HTTPError {
	return m.NewHTTPError(http.StatusUnauthorized, message, internal, otherInfo...)
}

func (m *ErrorDispatcher) NewNotFound(message string, internal error, otherInfo ...any) *echo.HTTPError {
	return m.NewHTTPError(http.StatusNotFound, message, internal, otherInfo...)
}

func (m *ErrorDispatcher) NewConflict(message string, internal error, otherInfo ...any) *echo.HTTPError {
	return m.NewHTTPError(http.StatusConflict, message, internal, otherInfo...)
}

func (m *ErrorDispatcher) NewForbidden(message string, internal error, otherInfo ...any) *echo.HTTPError {
	return m.NewHTTPError(http.StatusForbidden, message, internal, otherInfo...)
}

func StatusCodeToLogLevel(status int) (string, slog.Level) {
	switch {
	case status >= 500:
		return "ERROR", slog.LevelError
	case status >= 400:
		return "WARN", slog.LevelWarn
	default:
		return "INFO", slog.LevelInfo
	}
}
