package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"visory/internal/models"
	"visory/internal/utils"

	"github.com/labstack/echo/v4"
)

// LoggingMiddleware is a custom middleware that logs requests to both stdout and database
func (s *Server) LoggingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Start timing the request
			start := time.Now()

			// Try to extract user ID from session cookie
			var userID int64
			cookie, err := c.Cookie(models.COOKIE_NAME)
			if err == nil {
				// Try to get user from session
				user, err := s.db.User.GetBySessionToken(c.Request().Context(), cookie.Value)
				if err == nil {
					userID = user.ID
				}
			}

			// Add user ID to context for later use
			ctx := context.WithValue(c.Request().Context(), "userID", userID)
			c.SetRequest(c.Request().WithContext(ctx))

			// Continue processing the request
			err = next(c)

			// Get response status code
			statusCode := c.Response().Status
			if err != nil {
				// If there's an error, try to extract status code
				if httpErr, ok := err.(*echo.HTTPError); ok {
					statusCode = httpErr.Code
				} else {
					statusCode = http.StatusInternalServerError
				}
			}

			// Calculate duration
			duration := time.Since(start)

			// Log to database
			go utils.LogRequestDetails(
				c.Request().Context(),
				s.db,
				userID,
				c.Request().Method,
				c.Request().URL.Path,
				statusCode,
				duration,
				c.Request().RemoteAddr,
			)

			// Also log using slog with service group "http"
			logLevel := getLevelFromStatusCode(statusCode)
			logMessage := "HTTP " + c.Request().Method + " " + c.Request().URL.Path

			s.logger.Log(c.Request().Context(), logLevel, logMessage,
				slog.Int("status", statusCode),
				slog.String("method", c.Request().Method),
				slog.String("path", c.Request().URL.Path),
				slog.Int64("duration_ms", duration.Milliseconds()),
				slog.Int64("user_id", userID),
				slog.String("remote_addr", c.Request().RemoteAddr),
			)

			return err
		}
	}
}

// getLevelFromStatusCode determines slog level based on HTTP status code
func getLevelFromStatusCode(statusCode int) slog.Level {
	switch {
	case statusCode >= 500:
		return slog.LevelError
	case statusCode >= 400:
		return slog.LevelWarn
	case statusCode >= 300:
		return slog.LevelInfo
	default:
		return slog.LevelDebug
	}
}
