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

	baseRequestLogger := RequestLogger(s.logger, s.dispatcher)

	e.Use(middleware.Recover())
	api := e.Group("/api")

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://*", "http://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	api.GET("/", s.HelloWorldHandler, baseRequestLogger)
	Roles := s.authService.RBACMiddleware
	api.GET("/health", s.healthHandler, baseRequestLogger)

	api.GET("/websocket", s.websocketHandler)

	authlogger := RequestLogger(s.authService.Logger, s.authService.Dispatcher)
	api.POST("/auth/register", s.authService.Register, authlogger)
	api.POST("/auth/login", s.authService.Login, authlogger)

	// OAuth routes
	api.GET("/auth/oauth/:provider", s.authService.OAuthLogin, authlogger)
	api.GET("/auth/oauth/callback/:provider", s.authService.OAuthCallback, authlogger)

	authGroup := api.Group("/auth", s.authService.AuthMiddleware, authlogger)
	authGroup.GET("/me", s.authService.Me)
	authGroup.POST("/logout", s.authService.Logout)

	// Storage routes
	storageGroup := api.Group("/storage", RequestLogger(s.storageService.Logger, s.storageService.Dispatcher))
	storageGroup.Use(s.authService.AuthMiddleware)
	storageGroup.Use(Roles(models.RBAC_SETTINGS_MANAGER))
	storageGroup.GET("/devices", s.storageService.GetStorageDevices)
	storageGroup.GET("/mount-points", s.storageService.GetMountPoints)

	// Users routes
	usersGroup := api.Group("/users", RequestLogger(s.usersService.Logger, s.usersService.Dispatcher))
	usersGroup.Use(s.authService.AuthMiddleware)
	usersGroup.GET("", s.usersService.GetAllUsers, Roles(models.RBAC_USER_ADMIN))
	usersGroup.GET("/", s.usersService.GetAllUsers, Roles(models.RBAC_USER_ADMIN))
	usersGroup.GET("/:id", s.usersService.GetUserById, Roles(models.RBAC_USER_ADMIN))
	usersGroup.POST("", s.usersService.CreateUser, Roles(models.RBAC_USER_ADMIN))
	usersGroup.PUT("/:id", s.usersService.UpdateUser, Roles(models.RBAC_USER_ADMIN))
	usersGroup.DELETE("/:id", s.usersService.DeleteUser, Roles(models.RBAC_USER_ADMIN))
	usersGroup.PATCH("/:id/role", s.usersService.UpdateUserRole, Roles(models.RBAC_USER_ADMIN))

	// Logs routes
	logsGroup := api.Group("/logs", s.authService.AuthMiddleware, RequestLogger(s.logsService.Logger, s.logsService.Dispatcher))
	logsGroup.GET("", s.logsService.GetLogs, Roles(models.RBAC_AUDIT_LOG_VIEWER))
	logsGroup.GET("/stats", s.logsService.GetLogStats, Roles(models.RBAC_AUDIT_LOG_VIEWER))
	logsGroup.DELETE("/cleanup", s.logsService.ClearOldLogs, Roles(models.RBAC_USER_ADMIN))

	// Metrics routes
	metricsGroup := api.Group("/metrics", s.authService.AuthMiddleware, RequestLogger(s.metricsService.Logger, s.metricsService.Dispatcher))
	metricsGroup.GET("", s.metricsService.GetMetrics, Roles(models.RBAC_AUDIT_LOG_VIEWER))
	metricsGroup.GET("/health", s.metricsService.GetHealthMetrics, Roles(models.RBAC_HEALTH_CHECKER))
	metricsGroup.GET("/:service", s.metricsService.GetServiceMetrics, Roles(models.RBAC_AUDIT_LOG_VIEWER))

	docsGroup := api.Group("/docs", RequestLogger(s.docsService.Logger, s.docsService.Dispatcher))
	docsGroup.GET("/swagger", s.docsService.ServeSwagger)
	docsGroup.GET("/redoc", s.docsService.ServeRedoc)
	docsGroup.GET("/spec", s.docsService.ServeSpec)

	return e
}

// @Summary      hello world
// @Description  simple hello world endpoint
// @Tags         general
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       / [get]
func (s *Server) HelloWorldHandler(c echo.Context) error {
	resp := map[string]string{
		"message": "Hello World",
	}

	return c.JSON(http.StatusOK, resp)
}

// @Summary      health check
// @Description  check database health status
// @Tags         health
// @Produce      json
// @Success      200  {object}  database.Health
// @Failure      500  {object}  models.HTTPError
// @Router       /health [get]
func (s *Server) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.db.Health())
}

// websocketHandler handles WebSocket connections
//
//	@Summary      websocket connection
//	@Description  establishes websocket connection for real-time updates
//	@Tags         websocket
//	@Router       /websocket [get]
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
