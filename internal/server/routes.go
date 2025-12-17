package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"visory/internal/models"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/coder/websocket"
)

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()

	e.Use(s.RequestLogger())

	// RequestLogger middleware with slog integration
	// e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
	// 	LogStatus:   true,
	// 	LogURI:      true,
	// 	LogMethod:   true,
	// 	LogRemoteIP: true,
	// 	LogError:    true,
	// 	LogLatency:  true,
	// 	HandleError: true,
	// 	Skipper: func(c echo.Context) bool {
	// 		// if options endpoint, skip logging
	// 		if c.Request().Method == http.MethodOptions {
	// 			return true
	// 		}
	// 		// Skip logging for health check and auth/me endpoints
	// 		SKIP_ENDPOINTS := map[string]bool{
	// 			"/api/health":  true,
	// 			"/api/auth/me": true,
	// 		}
	// 		_, skip := SKIP_ENDPOINTS[c.Request().URL.Path]
	// 		return skip
	// 	},
	// 	BeforeNextFunc: func(c echo.Context) {
	// 		// Try to extract user ID from session cookie
	// 		var userID int64
	// 		cookie, err := c.Cookie("session_token")
	// 		if err == nil {
	// 			// Try to get user from session
	// 			user, err := s.db.User.GetBySessionToken(c.Request().Context(), cookie.Value)
	// 			if err == nil {
	// 				userID = user.ID
	// 			}
	// 		}
	// 		// Add user ID to context for later use
	// 		ctx := context.WithValue(c.Request().Context(), "userID", userID)
	// 		c.SetRequest(c.Request().WithContext(ctx))
	// 	},
	// 	LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
	// 		userID, _ := c.Get("userID").(int64)
	//
	// 		// Determine log level based on status code
	// 		logLevel := getLevelFromStatusCode(v.Status)
	//
	// 		// Log to slog with appropriate level
	// 		if v.Error != nil {
	// 			s.logger.Error("HTTP request",
	// 				"status", v.Status,
	// 				"method", v.Method,
	// 				"uri", v.URI,
	// 				"latency_ms", v.Latency.Milliseconds(),
	// 				"user_id", userID,
	// 				"remote_ip", v.RemoteIP,
	// 				"error", v.Error.Error(),
	// 			)
	// 		} else {
	// 			// Use the appropriate log level based on status code
	// 			s.logger.Log(c.Request().Context(), logLevel, "HTTP request",
	// 				"status", v.Status,
	// 				"method", v.Method,
	// 				"uri", v.URI,
	// 				"latency_ms", v.Latency.Milliseconds(),
	// 				"user_id", userID,
	// 				"remote_ip", v.RemoteIP,
	// 			)
	// 		}
	//
	// 		// Log to database asynchronously with background context to avoid cancellation
	// 		go func() {
	// 			jsonifiedDetails, err := json.Marshal(v)
	// 			if err != nil {
	// 				slog.Error("error happened", "err", err)
	// 				return
	// 			}
	//
	// 			action := fmt.Sprintf("%v %v", v.Method, v.URI)
	// 			details := string(jsonifiedDetails)
	// 			ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// 			defer cancel()
	//
	// 			// Use background context to prevent cancellation issues
	// 			_, err = s.db.Log.CreateLog(ctxWithTimeout, logs.CreateLogParams{
	// 				UserID:       userID,
	// 				Action:       action,
	// 				Details:      &details,
	// 				ServiceGroup: "http",
	// 				Level:        logLevel.String(),
	// 			})
	// 			if err != nil {
	// 				s.logger.Debug("could not create log entry", "error", err)
	// 			}
	// 		}()
	//
	// 		return nil
	// 	},
	// }))

	e.Use(middleware.Recover())
	api := e.Group("/api")

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://*", "http://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	api.GET("/", s.HelloWorldHandler)
	Roles := s.authService.RBACMiddleware
	api.GET("/health", s.healthHandler)

	api.GET("/websocket", s.websocketHandler)

	api.POST("/auth/register", s.authService.Register)
	api.POST("/auth/login", s.authService.Login)

	// OAuth routes
	api.GET("/auth/oauth/:provider", s.authService.OAuthLogin)
	api.GET("/auth/oauth/callback/:provider", s.authService.OAuthCallback)

	authGroup := api.Group("/auth")
	authGroup.Use(s.authService.AuthMiddleware)
	authGroup.GET("/me", s.authService.Me)
	authGroup.POST("/logout", s.authService.Logout)

	// Storage routes
	storageGroup := api.Group("/storage")
	storageGroup.Use(s.authService.AuthMiddleware)
	storageGroup.Use(Roles(models.RBAC_SETTINGS_MANAGER))
	storageGroup.GET("/devices", s.storageService.GetStorageDevices)
	storageGroup.GET("/mount-points", s.storageService.GetMountPoints)

	// Users routes
	usersGroup := api.Group("/users")
	usersGroup.Use(s.authService.AuthMiddleware)
	usersGroup.GET("", s.usersService.GetAllUsers, Roles(models.RBAC_USER_ADMIN))
	usersGroup.GET("/", s.usersService.GetAllUsers, Roles(models.RBAC_USER_ADMIN))
	usersGroup.GET("/:id", s.usersService.GetUserById, Roles(models.RBAC_USER_ADMIN))
	usersGroup.POST("", s.usersService.CreateUser, Roles(models.RBAC_USER_ADMIN))
	usersGroup.PUT("/:id", s.usersService.UpdateUser, Roles(models.RBAC_USER_ADMIN))
	usersGroup.DELETE("/:id", s.usersService.DeleteUser, Roles(models.RBAC_USER_ADMIN))
	usersGroup.PATCH("/:id/role", s.usersService.UpdateUserRole, Roles(models.RBAC_USER_ADMIN))

	// Logs routes
	logsGroup := api.Group("/logs")
	logsGroup.Use(s.authService.AuthMiddleware)
	logsGroup.GET("", s.logsService.GetLogs, Roles(models.RBAC_AUDIT_LOG_VIEWER))
	logsGroup.GET("/stats", s.logsService.GetLogStats, Roles(models.RBAC_AUDIT_LOG_VIEWER))
	logsGroup.DELETE("/cleanup", s.logsService.ClearOldLogs, Roles(models.RBAC_USER_ADMIN))

	// Metrics routes
	metricsGroup := api.Group("/metrics")
	metricsGroup.Use(s.authService.AuthMiddleware)
	metricsGroup.GET("", s.metricsService.GetMetrics, Roles(models.RBAC_AUDIT_LOG_VIEWER))
	metricsGroup.GET("/health", s.metricsService.GetHealthMetrics, Roles(models.RBAC_HEALTH_CHECKER))
	metricsGroup.GET("/:service", s.metricsService.GetServiceMetrics, Roles(models.RBAC_AUDIT_LOG_VIEWER))

	return e
}

func (s *Server) HelloWorldHandler(c echo.Context) error {
	resp := map[string]string{
		"message": "Hello World",
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.db.Health())
}

func (s *Server) websocketHandler(c echo.Context) error {
	w := c.Response().Writer
	r := c.Request()
	socket, err := websocket.Accept(w, r, nil)
	if err != nil {
		s.logger.Error("could not open websocket", "error", err)
		_, _ = w.Write([]byte("could not open websocket"))
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}

	defer socket.Close(websocket.StatusGoingAway, "server closing websocket")

	ctx := r.Context()
	socketCtx := socket.CloseRead(ctx)

	for {
		payload := fmt.Sprintf("server timestamp: %d", time.Now().UnixNano())
		err := socket.Write(socketCtx, websocket.MessageText, []byte(payload))
		if err != nil {
			break
		}
		time.Sleep(time.Second * 2)
	}
	return nil
}

// getLevelFromStatusCode determines log level based on HTTP status code
func getLevelFromStatusCode(statusCode int) slog.Level {
	switch {
	case statusCode >= 500:
		return slog.LevelError
	case statusCode >= 400:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}
