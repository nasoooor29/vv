package server

import (
	"fmt"
	"log/slog"
	"time"

	"visory/internal/database/user"
	"visory/internal/models"
	"visory/internal/utils"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
)

var IGNORED_ROUTES = map[string]bool{
	"/api/health":  true,
	"/api/auth/me": true,
}

func RequestLogger(logger *slog.Logger, dispatcher *utils.Dispatcher) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// if options endpoint, skip logging
			if c.Request().Method == "OPTIONS" {
				return next(c)
			}
			if _, ok := IGNORED_ROUTES[c.Path()]; ok {
				return next(c)
			}
			start := time.Now()

			rid := uuid.Nil
			if id, err := uuid.NewV4(); err != nil {
				logger.Error("failed to generate request id", "error", err)
			} else {
				rid = id
			}

			c.Set("RequestId", rid.String())

			err := next(c)

			req := c.Request()
			res := c.Response()

			status := res.Status
			var errMsg any
			if err != nil {
				if he, ok := err.(*echo.HTTPError); ok {
					errMsg = he.Message
					status = he.Code
				}
				c.Error(err)
			}

			var userID int64 = -1
			if u, ok := c.Get("userWithSession").(user.GetUserAndSessionByTokenRow); ok {
				userID = u.User.ID
			}
			data := models.LogRequestData{
				RequestId: rid.String(),
				UserId:    userID,
				Method:    req.Method,
				Path:      c.Path(),
				Uri:       req.RequestURI,
				Status:    status,
				Latency:   time.Since(start),
				RemoteIp:  c.RealIP(),
				UserAgent: req.UserAgent(),
				Protocol:  req.Proto,
				Bytes:     res.Size,
				Error:     errMsg,
			}

			if models.ENV_VARS.VerboseNotifications {
				notifyAndInsert(dispatcher, data)
			} else if req.Method != "GET" {
				notifyAndInsert(dispatcher, data)
			}

			_, logLevel := utils.StatusCodeToLogLevel(data.Status)
			switch logLevel {
			case slog.LevelError:
				logger.Error("[REQUEST]", "data", data)
			case slog.LevelWarn:
				logger.Warn("[REQUEST]", "data", data)
			default:
				logger.Info("[REQUEST]", "data", data)
			}

			return err
		}
	}
}

func notifyAndInsert(dispatcher *utils.Dispatcher, data models.LogRequestData) {
	dispatcher.InsertIntoDB(data)

	// Send notification based on log level
	_, logLevel := utils.StatusCodeToLogLevel(data.Status)
	fmt.Printf("logLevel: %v\n", logLevel)
	if logLevel == slog.LevelError {
		dispatcher.SendError(
			"Request Error",
			fmt.Sprintf("%s %s returned %d", data.Method, data.Path, data.Status),
			map[string]string{
				"Method":    data.Method,
				"Path":      data.Path,
				"Status":    fmt.Sprintf("%d", data.Status),
				"Remote IP": data.RemoteIp,
				"Latency":   data.Latency.String(),
			},
		)
	} else if data.Method == "DELETE" {
		dispatcher.SendWarning(
			"Request Warning",
			fmt.Sprintf("%s %s returned %d", data.Method, data.Path, data.Status),
			map[string]string{
				"Method":    data.Method,
				"Path":      data.Path,
				"Status":    fmt.Sprintf("%d", data.Status),
				"Remote IP": data.RemoteIp,
				"Latency":   data.Latency.String(),
			},
		)
	} else if logLevel == slog.LevelWarn {
		dispatcher.SendWarning(
			"Request Warning",
			fmt.Sprintf("%s %s returned %d", data.Method, data.Path, data.Status),
			map[string]string{
				"Method":    data.Method,
				"Path":      data.Path,
				"Status":    fmt.Sprintf("%d", data.Status),
				"Remote IP": data.RemoteIp,
				"Latency":   data.Latency.String(),
			},
		)
	} else {
		fmt.Printf("logLevel: %v\n", logLevel)
		dispatcher.SendInfo(
			"Request Info",
			fmt.Sprintf("%s %s returned %d", data.Method, data.Path, data.Status),
			map[string]string{
				"Method":    data.Method,
				"Path":      data.Path,
				"Status":    fmt.Sprintf("%d", data.Status),
				"Remote IP": data.RemoteIp,
				"Latency":   data.Latency.String(),
			},
		)
	}
}
