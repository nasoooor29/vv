package services

import (
	"net/http"

	storageService "visory/internal/storage"
	"visory/internal/utils"

	"github.com/labstack/echo/v4"
)

type StorageService struct {
	dispatcher *utils.Dispatcher
}

// NewStorageService creates a new StorageService with dependency injection
func NewStorageService(logger *utils.Dispatcher) *StorageService {
	// Create a grouped logger for storage service
	storageLogger := logger.WithGroup("storage")
	return &StorageService{
		dispatcher: storageLogger,
	}
}

// GetStorageDevices returns list of storage devices
func (s *StorageService) GetStorageDevices(c echo.Context) error {
	devices, err := storageService.GetBlockDevices()
	if err != nil {
		return s.dispatcher.NewInternalServerError("Failed to get storage devices", err)
	}

	return c.JSON(http.StatusOK, devices)
}

// GetMountPoints returns list of mount points
func (s *StorageService) GetMountPoints(c echo.Context) error {
	mountPoints, err := storageService.GetMountPoints()
	if err != nil {
		return s.dispatcher.NewInternalServerError("Failed to get mount points", err)
	}

	return c.JSON(http.StatusOK, mountPoints)
}
