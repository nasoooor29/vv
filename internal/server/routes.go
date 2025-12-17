package server

import (
	"fmt"
	"net/http"
	"time"

	"visory/internal/models"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/coder/websocket"
)

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()

	// Add custom logging middleware before other middlewares
	e.Use(s.LoggingMiddleware())

	e.Use(middleware.Logger())
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

	// api.GET("/health", s.healthHandler, s.RBAC(models.RBAC_HEALTH_CHECKER))
	api.GET("/health", s.healthHandler)

	api.GET("/websocket", s.websocketHandler)

	api.POST("/auth/register", s.authService.Register)
	api.POST("/auth/login", s.authService.Login)

	// OAuth routes
	api.GET("/auth/oauth/:provider", s.authService.OAuthLogin)
	api.GET("/auth/oauth/callback/:provider", s.authService.OAuthCallback)

	authGroup := api.Group("/auth")
	authGroup.Use(s.Auth)
	authGroup.GET("/me", s.authService.Me)
	authGroup.POST("/logout", s.authService.Logout)

	// Storage routes
	storageGroup := api.Group("/storage")
	storageGroup.Use(s.Auth)
	storageGroup.Use(s.RBAC(models.RBAC_SETTINGS_MANAGER))
	storageGroup.GET("/devices", s.storageService.GetStorageDevices)
	storageGroup.GET("/mount-points", s.storageService.GetMountPoints)

	// Users routes
	usersGroup := api.Group("/users")
	usersGroup.Use(s.Auth)
	usersGroup.GET("", s.usersService.GetAllUsers, s.RBAC(models.RBAC_USER_ADMIN))
	usersGroup.GET("/", s.usersService.GetAllUsers, s.RBAC(models.RBAC_USER_ADMIN))
	usersGroup.GET("/:id", s.usersService.GetUserById, s.RBAC(models.RBAC_USER_ADMIN))
	usersGroup.POST("", s.usersService.CreateUser, s.RBAC(models.RBAC_USER_ADMIN))
	usersGroup.PUT("/:id", s.usersService.UpdateUser, s.RBAC(models.RBAC_USER_ADMIN))
	usersGroup.DELETE("/:id", s.usersService.DeleteUser, s.RBAC(models.RBAC_USER_ADMIN))
	usersGroup.PATCH("/:id/role", s.usersService.UpdateUserRole, s.RBAC(models.RBAC_USER_ADMIN))

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
