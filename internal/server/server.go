package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"visory/internal/database"
	"visory/internal/models"
	"visory/internal/services"

	"github.com/markbates/goth"
)

type Server struct {
	port int

	db             *database.Service
	logger         *slog.Logger
	OAuthProviders map[string]goth.Provider

	// Services
	authService    *services.AuthService
	usersService   *services.UsersService
	storageService *services.StorageService
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(models.ENV_VARS.Port)
	logger := slog.Default()

	db := database.New()
	authService := services.NewAuthService(db, logger)
	usersService := services.NewUsersService(db, logger)
	storageService := services.NewStorageService(logger)

	NewServer := &Server{
		port:           port,
		db:             db,
		logger:         logger,
		OAuthProviders: authService.OAuthProviders,
		authService:    authService,
		usersService:   usersService,
		storageService: storageService,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
