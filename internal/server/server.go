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
	authService      *services.AuthService
	usersService     *services.UsersService
	storageService   *services.StorageService
	logsService      *services.LogsService
	docsService      *services.DocsService
	metricsService   *services.MetricsService
	qemuService      *services.QemuService
	dockerService    *services.DockerService
	firewallService  *services.FirewallService
	templatesService *services.TemplatesService
	backupService    *services.BackupService
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
	firewallService := services.NewFirewallService(serverDispatcher, logger)
	templatesService := services.NewTemplatesService(serverDispatcher, logger, dockerService.ClientManager)
	backupService := services.NewBackupService(db, serverDispatcher, logger)

	// Wire firewall backup callback - creates backup on every rule change
	firewallService.SetOnRulesChanged(func(rules []models.FirewallRule) {
		if err := backupService.CreateFirewallBackup(rules); err != nil {
			logger.Error("failed to create firewall backup", "error", err)
		}
	})

	// Initialize Docker clients from environment variables
	docsService := services.NewDocsService(db, serverDispatcher, logger)
	qemuService := services.NewQemuService(serverDispatcher, logger)

	NewServer := &Server{
		port:             port,
		db:               db,
		logger:           logger,
		dispatcher:       serverDispatcher,
		OAuthProviders:   authService.OAuthProviders,
		authService:      authService,
		usersService:     usersService,
		storageService:   storageService,
		logsService:      logsService,
		metricsService:   metricsService,
		dockerService:    dockerService,
		docsService:      docsService,
		qemuService:      qemuService,
		firewallService:  firewallService,
		templatesService: templatesService,
		backupService:    backupService,
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
