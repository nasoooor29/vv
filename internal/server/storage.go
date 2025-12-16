package server

import (
	"log/slog"
	"net/http"

	storageService "visory/internal/storage"

	"github.com/labstack/echo/v4"
)

// GetStorageDevices returns list of storage devices
func (s *Server) GetStorageDevices(c echo.Context) error {
	devices, err := storageService.GetBlockDevices()
	if err != nil {
		slog.Error("failed to get block devices", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get storage devices").SetInternal(err)
	}

	return c.JSON(http.StatusOK, devices)
}

// GetMountPoints returns list of mount points
func (s *Server) GetMountPoints(c echo.Context) error {
	mountPoints, err := storageService.GetMountPoints()
	if err != nil {
		slog.Error("failed to get mount points", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get mount points").SetInternal(err)
	}

	return c.JSON(http.StatusOK, mountPoints)
}
