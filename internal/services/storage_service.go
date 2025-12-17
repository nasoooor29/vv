package services

import (
	"log/slog"
	"net/http"

	storageService "visory/internal/storage"
	"visory/internal/utils"

	"github.com/labstack/echo/v4"
)

type StorageService struct {
	Dispatcher *utils.Dispatcher
	Logger     *slog.Logger
}

// NewStorageService creates a new StorageService with dependency injection
func NewStorageService(dispatcher *utils.Dispatcher, logger *slog.Logger) *StorageService {
	// Create a grouped logger for storage service
	return &StorageService{
		Dispatcher: dispatcher.WithGroup("storage"),
		Logger:     logger.WithGroup("storage"),
	}
}

// GetStorageDevices returns list of storage devices
func (s *StorageService) GetStorageDevices(c echo.Context) error {
	devices, err := storageService.GetBlockDevices()
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to get storage devices", err)
	}

	return c.JSON(http.StatusOK, devices)
}

// GetMountPoints returns list of mount points
func (s *StorageService) GetMountPoints(c echo.Context) error {
	mountPoints, err := storageService.GetMountPoints()
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to get mount points", err)
	}

	return c.JSON(http.StatusOK, mountPoints)
}
