package server

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"visory/internal/database"
	"visory/internal/models"

	"github.com/markbates/goth"
)

type Server struct {
	port int

	db             *database.Service
	OAuthProviders map[string]goth.Provider
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(models.ENV_VARS.Port)
	providers := InitializeOAuth()
	NewServer := &Server{
		port:           port,
		db:             database.New(),
		OAuthProviders: providers,
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
