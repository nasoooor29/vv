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
	"visory/internal/notifications"

	"github.com/labstack/echo/v4"
)

type Dispatcher struct {
	db       *database.Service
	Groups   []string
	notifier *notifications.Manager
}

func NewDispatcher(db *database.Service, notifier *notifications.Manager) *Dispatcher {
	return &Dispatcher{
		db:       db,
		Groups:   []string{},
		notifier: notifier,
	}
}

func (m *Dispatcher) WithGroup(name string) *Dispatcher {
	newGroups := append(m.Groups, name)

	return &Dispatcher{
		db:       m.db,
		Groups:   newGroups,
		notifier: m.notifier,
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
	fmt.Printf("internal: %v\n", internal)
	fmt.Printf("message: %v\n", message)
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

// getGroupName returns the current group name or a default
func (m *Dispatcher) getGroupName() string {
	if len(m.Groups) > 0 {
		return m.Groups[len(m.Groups)-1]
	}
	return "Visory"
}

// SendError sends an error notification to all registered senders
func (m *Dispatcher) SendError(title string, message string, fields map[string]string) {
	if m.notifier == nil {
		return
	}
	m.notifier.SendError(title, message, fields, m.getGroupName(), models.ENV_VARS.APP_VERSION)
}

// SendWarning sends a warning notification to all registered senders
func (m *Dispatcher) SendWarning(title string, message string, fields map[string]string) {
	if m.notifier == nil {
		return
	}
	m.notifier.SendWarning(title, message, fields, m.getGroupName(), models.ENV_VARS.APP_VERSION)
}

// SendInfo sends an info notification to all registered senders
func (m *Dispatcher) SendInfo(title string, message string, fields map[string]string) {
	if m.notifier == nil {
		return
	}
	m.notifier.SendInfo(title, message, fields, m.getGroupName(), models.ENV_VARS.APP_VERSION)
}

// SendSuccess sends a success notification to all registered senders
func (m *Dispatcher) SendSuccess(title string, message string, fields map[string]string) {
	if m.notifier == nil {
		return
	}
	m.notifier.SendSuccess(title, message, fields, m.getGroupName(), models.ENV_VARS.APP_VERSION)
}
