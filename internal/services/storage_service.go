package services

import (
	"log/slog"
	"net/http"

	storageService "visory/internal/storage"

	"github.com/labstack/echo/v4"
)

type StorageService struct {
	logger *slog.Logger
}

// NewStorageService creates a new StorageService with dependency injection
func NewStorageService(logger *slog.Logger) *StorageService {
	// Create a grouped logger for storage service
	storageLogger := logger.WithGroup("storage")
	return &StorageService{
		logger: storageLogger,
	}
}

// GetStorageDevices returns list of storage devices
func (s *StorageService) GetStorageDevices(c echo.Context) error {
	devices, err := storageService.GetBlockDevices()
	if err != nil {
		s.logger.Error("failed to get block devices", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get storage devices").SetInternal(err)
	}

	return c.JSON(http.StatusOK, devices)
}

// GetMountPoints returns list of mount points
func (s *StorageService) GetMountPoints(c echo.Context) error {
	mountPoints, err := storageService.GetMountPoints()
	if err != nil {
		s.logger.Error("failed to get mount points", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get mount points").SetInternal(err)
	}

	return c.JSON(http.StatusOK, mountPoints)
}
