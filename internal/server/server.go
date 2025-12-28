package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"visory/internal/database"
	"visory/internal/models"
	"visory/internal/notifications"
	"visory/internal/services"
	"visory/internal/utils"

	"github.com/markbates/goth"
)

type Server struct {
	port int

	fs             *utils.FS
	db             *database.Service
	dispatcher     *utils.Dispatcher
	logger         *slog.Logger
	OAuthProviders map[string]goth.Provider

	// Services
	firewallService  *services.FirewallService
	templatesService *services.TemplatesService
	authService      *services.AuthService
	usersService     *services.UsersService
	storageService   *services.StorageService
	logsService      *services.LogsService
	docsService      *services.DocsService
	metricsService   *services.MetricsService
	qemuService      *services.QemuService
	isoService       *services.ISOService
	dockerService    *services.DockerService
	vncProxy         *services.VNCProxy
	settingsService  *services.SettingsService
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

	// Initialize notification manager with configured senders
	notifier := notifications.NewManager()

	// Register Discord sender if configured
	if models.ENV_VARS.DiscordWebhookURL != "" {
		discordSender := notifications.NewDiscordSender(notifications.DiscordConfig{
			WebhookURL:    models.ENV_VARS.DiscordWebhookURL,
			NotifyOnError: models.ENV_VARS.DiscordNotifyOnError,
			NotifyOnWarn:  models.ENV_VARS.DiscordNotifyOnWarn,
			NotifyOnInfo:  models.ENV_VARS.DiscordNotifyOnInfo,
		})
		notifier.RegisterSender(discordSender)
	}

	dispatcher := utils.NewDispatcher(db, notifier)
	fs := utils.NewFS(models.ENV_VARS.Directory)

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

	// Initialize Docker clients from environment variables
	docsService := services.NewDocsService(db, serverDispatcher, logger)
	settingsService := services.NewSettingsService(db, serverDispatcher, logger, notifier)

	// Load notification settings from database
	loadNotificationSettingsFromDB(db, notifier)
	qemuService := services.NewQemuService(serverDispatcher, fs, logger)
	isoService := services.NewISOService(serverDispatcher, fs, logger)
	vncProxy := services.NewVNCProxy(logger)

	NewServer := &Server{
		port:             port,
		db:               db,
		logger:           logger,
		fs:               fs,
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
		isoService:       isoService,
		firewallService:  firewallService,
		templatesService: templatesService,
		vncProxy:         vncProxy,
		settingsService:  settingsService,
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

// loadNotificationSettingsFromDB loads notification settings from the database and registers senders
func loadNotificationSettingsFromDB(db *database.Service, notifier *notifications.Manager) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	settings, err := db.Notification.GetEnabledNotificationSettings(ctx)
	if err != nil {
		slog.Warn("Failed to load notification settings from database", "error", err)
		return
	}

	for _, setting := range settings {
		if setting.WebhookUrl == nil || *setting.WebhookUrl == "" {
			continue
		}

		switch setting.Provider {
		case "discord":
			notifyOnError := setting.NotifyOnError != nil && *setting.NotifyOnError
			notifyOnWarn := setting.NotifyOnWarn != nil && *setting.NotifyOnWarn
			notifyOnInfo := setting.NotifyOnInfo != nil && *setting.NotifyOnInfo

			sender := notifications.NewDiscordSender(notifications.DiscordConfig{
				WebhookURL:    *setting.WebhookUrl,
				NotifyOnError: notifyOnError,
				NotifyOnWarn:  notifyOnWarn,
				NotifyOnInfo:  notifyOnInfo,
			})
			notifier.RegisterSender(sender)
			slog.Info("Loaded notification setting from database", "provider", setting.Provider)
		}
	}
}
