package services

import (
	"log/slog"
	"net/http"

	_ "visory/internal/models"
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

//	@Summary      list storage devices
//	@Description  get list of block storage devices
//	@Tags         storage
//	@Produce      json
//	@Success      200  {array}   models.StorageDevice
//	@Failure      401  {object}  models.HTTPError
//	@Failure      403  {object}  models.HTTPError
//	@Failure      500  {object}  models.HTTPError
//	@Router       /storage/devices [get]
//
// GetStorageDevices returns list of storage devices
func (s *StorageService) GetStorageDevices(c echo.Context) error {
	devices, err := storageService.GetBlockDevices()
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to get storage devices", err)
	}

	return c.JSON(http.StatusOK, devices)
}

//	@Summary      list mount points
//	@Description  get list of mount points
//	@Tags         storage
//	@Produce      json
//	@Success      200  {array}   models.MountPoint
//	@Failure      401  {object}  models.HTTPError
//	@Failure      403  {object}  models.HTTPError
//	@Failure      500  {object}  models.HTTPError
//	@Router       /storage/mount-points [get]
//
// GetMountPoints returns list of mount points
func (s *StorageService) GetMountPoints(c echo.Context) error {
	mountPoints, err := storageService.GetMountPoints()
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to get mount points", err)
	}

	return c.JSON(http.StatusOK, mountPoints)
}
