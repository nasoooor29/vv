package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"visory/internal/database"
	"visory/internal/models"
	"visory/internal/services"
	"visory/internal/utils"

	"github.com/markbates/goth"
)

type Server struct {
	port int

	db             *database.Service
	dispatcher     *utils.Dispatcher
	logger         *slog.Logger
	OAuthProviders map[string]goth.Provider

	// Services
	authService    *services.AuthService
	usersService   *services.UsersService
	storageService *services.StorageService
	logsService    *services.LogsService
	docsService    *services.DocsService
	metricsService *services.MetricsService
	qemuService    *services.QemuService
	dockerService  *services.DockerService
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(models.ENV_VARS.Port)
	logger := slog.New(utils.NewDaLog(
		os.Stdout, utils.DaLogStyleLongType1, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	))
	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(slog.LevelDebug)
	db := database.New()
	dispatcher := utils.NewDispatcher(db)

	// Add server group to logger
	serverDispatcher := dispatcher.WithGroup("server")
	// logger = logger.WithGroup("server")

	authService := services.NewAuthService(db, serverDispatcher, logger)
	usersService := services.NewUsersService(db, serverDispatcher, logger)
	storageService := services.NewStorageService(serverDispatcher, logger)
	logsService := services.NewLogsService(db, serverDispatcher, logger)
	metricsService := services.NewMetricsService(db, serverDispatcher, logger)
	dockerService := services.NewDockerService(serverDispatcher, logger)

	// Initialize Docker clients from environment variables
	docsService := services.NewDocsService(db, serverDispatcher, logger)
	qemuService := services.NewQemuService(serverDispatcher, logger)

	NewServer := &Server{
		port:           port,
		db:             db,
		logger:         logger,
		dispatcher:     serverDispatcher,
		OAuthProviders: authService.OAuthProviders,
		authService:    authService,
		usersService:   usersService,
		storageService: storageService,
		logsService:    logsService,
		metricsService: metricsService,
		dockerService:  dockerService,
		docsService:    docsService,
		qemuService:    qemuService,
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
