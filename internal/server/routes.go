package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"visory/internal/database"
	"visory/internal/database/user"
	"visory/internal/models"
	"visory/internal/utils"

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

	openapi := utils.New()
	api := openapi.Group("/api")
	api.Returns(http.StatusInternalServerError, utils.JSON(models.HTTPError{}), "application/json")

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://*", "http://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	api.GET("/health", s.healthHandler).
		Returns(http.StatusOK, database.Health{}, "Database connection is healthy").
		Returns(http.StatusInternalServerError, models.HTTPError{}, "Database connection failed").
		Description("Health Check Endpoint")

	api.POST("/auth/register", s.Register).
		Input(user.UpsertUserParams{}).
		Returns(http.StatusOK, user.User{}).
		Returns(http.StatusBadRequest, models.HTTPError{}).
		Returns(http.StatusConflict, models.HTTPError{}).
		Description("User Registration Endpoint")

	api.POST("/auth/login", s.Login).
		Input(models.Login{}).
		Returns(http.StatusOK, user.User{}).
		Returns(http.StatusBadRequest, models.HTTPError{}).
		Returns(http.StatusUnauthorized, models.HTTPError{}).
		Description("User Login Endpoint")

	api.GET("/auth/oauth/:provider", s.OAuthLogin).
		Description("OAuth Login Endpoint")

	api.GET("/auth/oauth/callback/:provider", s.OAuthCallback).
		Returns(http.StatusOK, user.User{}).
		Returns(http.StatusBadRequest, models.HTTPError{}).
		Returns(http.StatusUnauthorized, models.HTTPError{}).
		Description("OAuth Callback Endpoint")

	authGroup := api.Group("/auth", s.Auth).
		Returns(http.StatusUnauthorized, models.HTTPError{})

	authGroup.GET("/me", s.Me).
		Returns(http.StatusOK, user.GetUserAndSessionByTokenRow{}).
		Description("Get Current User Endpoint")

	authGroup.POST("/logout", s.Logout).
		Returns(http.StatusOK, nil).
		Returns(http.StatusBadRequest, models.HTTPError{}).
		Description("User Logout Endpoint")

	spec := openapi.OpenAPI()
	data, err := spec.MarshalJSON()
	if err != nil {
		slog.Error("error happened", "err", err)
	} else {
		slog.Info("OpenAPI spec generated")
		api.GET("/openapi.json", func(c echo.Context) error {
			return c.JSONBlob(http.StatusOK, data)
		})

		// Swagger UI endpoint
		api.GET("/docs", func(c echo.Context) error {
			html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>API Documentation - Swagger UI</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@3/swagger-ui.css">
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin:0; padding:0; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@3/swagger-ui-bundle.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@3/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: "/api/openapi.json",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            })
            window.ui = ui
        }
    </script>
</body>
</html>`
			return c.HTML(http.StatusOK, html)
		})

	}

	openapi.Mount(e)

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
