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
	logger         *utils.MySlog
	OAuthProviders map[string]goth.Provider

	// Services
	authService    *services.AuthService
	usersService   *services.UsersService
	storageService *services.StorageService
	logsService    *services.LogsService
	metricsService *services.MetricsService
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(models.ENV_VARS.Port)
	// serverLogger := slog.New(utils.NewDaLog(
	// 	os.Stdout,
	// 	utils.DaLogStyleLongType1,
	// 	&slog.HandlerOptions{
	// 		Level: slog.LevelDebug,
	// 	},
	// ))
	// serverLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	// 	AddSource: true,
	// 	Level:     slog.LevelDebug,
	// 	ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
	// 		if a.Key == slog.SourceKey {
	// 			_, file, line, ok := runtime.Caller(6)
	// 			if !ok {
	// 				return a
	// 			}
	// 			shortFile := file
	// 			lastSlash := -1
	// 			for i := len(file) - 1; i >= 0; i-- {
	// 				if file[i] == '/' {
	// 					lastSlash = i
	// 					break
	// 				}
	// 			}
	// 			if lastSlash != -1 {
	// 				shortFile = file[lastSlash+1:]
	// 			}
	// 			a.Value = slog.StringValue(fmt.Sprintf("%s:%d", shortFile, line))
	// 			return a
	// 		}
	// 		if a.Key == "error" {
	// 			a.Value = slog.StringValue(fmt.Sprintf("%v", a.Value))
	// 			return a
	// 		}
	// 		return a
	// 	},
	// }))
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(slog.LevelDebug)
	db := database.New()
	myLogger := utils.NewMySlog(logger, db)

	// Add server group to logger
	serverLogger := myLogger.WithGroup("server")
	// logger = logger.WithGroup("server")

	authService := services.NewAuthService(db, serverLogger)
	usersService := services.NewUsersService(db, logger)
	storageService := services.NewStorageService(logger)
	logsService := services.NewLogsService(db, logger)
	metricsService := services.NewMetricsService(db, logger)

	NewServer := &Server{
		port:           port,
		db:             db,
		logger:         serverLogger,
		OAuthProviders: authService.OAuthProviders,
		authService:    authService,
		usersService:   usersService,
		storageService: storageService,
		logsService:    logsService,
		metricsService: metricsService,
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
