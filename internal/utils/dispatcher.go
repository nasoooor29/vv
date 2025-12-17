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

type Dispatcher struct {
	db     *database.Service
	Groups []string
}

func NewDispatcher(db *database.Service) *Dispatcher {
	return &Dispatcher{
		db:     db,
		Groups: []string{},
	}
}

func (m *Dispatcher) WithGroup(name string) *Dispatcher {
	newGroups := append(m.Groups, name)

	return &Dispatcher{
		db:     m.db,
		Groups: newGroups,
	}
}

func (m *Dispatcher) InsertIntoDB(data models.LogRequestData) error {
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
	fmt.Printf("m.Groups: %v\n", m.Groups)

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

func (m *Dispatcher) NewHTTPError(status int, message string, internal error, otherInfo ...any) *echo.HTTPError {
	// Convert otherInfo to strings and join them
	var otherErrs []error
	for _, info := range otherInfo {
		otherErrs = append(otherErrs, errors.New(fmt.Sprint(info)))
	}

	// Join all errors: internal and the others
	joinedErr := errors.Join(append([]error{internal}, otherErrs...)...)

	return echo.NewHTTPError(status, message).SetInternal(joinedErr)
}

func (m *Dispatcher) NewBadRequest(message string, internal error, otherInfo ...any) *echo.HTTPError {
	return m.NewHTTPError(http.StatusBadRequest, message, internal, otherInfo...)
}

func (m *Dispatcher) NewInternalServerError(message string, internal error, otherInfo ...any) *echo.HTTPError {
	return m.NewHTTPError(http.StatusInternalServerError, message, internal, otherInfo...)
}

func (m *Dispatcher) NewUnauthorized(message string, internal error, otherInfo ...any) *echo.HTTPError {
	return m.NewHTTPError(http.StatusUnauthorized, message, internal, otherInfo...)
}

func (m *Dispatcher) NewNotFound(message string, internal error, otherInfo ...any) *echo.HTTPError {
	return m.NewHTTPError(http.StatusNotFound, message, internal, otherInfo...)
}

func (m *Dispatcher) NewConflict(message string, internal error, otherInfo ...any) *echo.HTTPError {
	return m.NewHTTPError(http.StatusConflict, message, internal, otherInfo...)
}

func (m *Dispatcher) NewForbidden(message string, internal error, otherInfo ...any) *echo.HTTPError {
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
