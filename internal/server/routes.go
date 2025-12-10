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

func (s *Server) Auth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie(models.COOKIE_NAME)
		if err != nil {
			slog.Error("error happened", "err", err)
			return echo.NewHTTPError(http.StatusUnauthorized, "Failed to get user by session token").SetInternal(err)
		}
		_, err = s.db.User.GetBySessionToken(c.Request().Context(), cookie.Value)
		if err != nil {
			slog.Error("error happened", "err", err)
			return echo.NewHTTPError(http.StatusUnauthorized, "Failed to get user by session token").SetInternal(err)
		}

		return next(c)
	}
}

func (s *Server) RBAC(policies ...models.RBACPolicy) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie(models.COOKIE_NAME)
			if err != nil {
				slog.Error("error happened", "err", err)
				return echo.NewHTTPError(http.StatusUnauthorized, "Failed to get user by session token").SetInternal(err)
			}
			user, err := s.db.User.GetBySessionToken(c.Request().Context(), cookie.Value)
			if err != nil {
				slog.Error("error happened", "err", err)
				return echo.NewHTTPError(http.StatusUnauthorized, "Failed to get user by session token").SetInternal(err)
			}

			user_roles := models.RoleToRBACPolicies(user.Role)
			if _, ok := user_roles[models.RBAC_USER_ADMIN]; ok {
				return next(c)
			}
			for _, policy := range policies {
				if v, ok := user_roles[policy]; !ok || !v {
					return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions")
				}
			}

			return next(c)
		}
	}
}

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
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

	api.POST("/auth/register", s.Register)
	api.POST("/auth/login", s.Login)

	authGroup := api.Group("/auth")
	authGroup.Use(s.Auth)
	authGroup.GET("/me", s.Me)
	authGroup.POST("/logout", s.Logout)

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
		slog.Error("could not open websocket", "error", err)
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
