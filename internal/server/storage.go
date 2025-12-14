package server

import (
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"visory/internal/models"

	"github.com/labstack/echo/v4"
)

// GetStorageDevices returns list of storage devices
func (s *Server) GetStorageDevices(c echo.Context) error {
	devices, err := getBlockDevices()
	if err != nil {
		slog.Error("failed to get block devices", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get storage devices").SetInternal(err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"devices": devices,
	})
}

// GetMountPoints returns list of mount points
func (s *Server) GetMountPoints(c echo.Context) error {
	mountPoints, err := getMountPoints()
	if err != nil {
		slog.Error("failed to get mount points", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get mount points").SetInternal(err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"mount_points": mountPoints,
	})
}

// getBlockDevices parses lsblk output
func getBlockDevices() ([]models.StorageDevice, error) {
	cmd := exec.Command("lsblk", "-Pb", "-o", "NAME,SIZE,TYPE,MOUNTPOINT")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var devices []models.StorageDevice
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		device := parseBlockDeviceLine(line)
		if device != nil {
			devices = append(devices, *device)
		}
	}

	return devices, nil
}

// parseBlockDeviceLine parses a single lsblk output line in key="value" format
func parseBlockDeviceLine(line string) *models.StorageDevice {
	device := &models.StorageDevice{}

	// Parse key="value" pairs
	pairs := make(map[string]string)

	// Simple regex-like parsing for key="value" format
	i := 0
	for i < len(line) {
		// Find the key
		eqIdx := strings.IndexByte(line[i:], '=')
		if eqIdx == -1 {
			break
		}
		eqIdx += i

		key := line[i:eqIdx]
		i = eqIdx + 1

		// Expect a quote
		if i >= len(line) || line[i] != '"' {
			break
		}
		i++ // skip opening quote

		// Find the closing quote
		closeIdx := strings.IndexByte(line[i:], '"')
		if closeIdx == -1 {
			break
		}

		value := line[i : i+closeIdx]
		pairs[key] = value
		i += closeIdx + 1

		// Skip space if present
		if i < len(line) && line[i] == ' ' {
			i++
		}
	}

	// Map parsed pairs to device
	if name, ok := pairs["NAME"]; ok {
		device.Name = name
	} else {
		return nil
	}

	if sizeStr, ok := pairs["SIZE"]; ok {
		sizeBytes, _ := strconv.ParseInt(sizeStr, 10, 64)
		device.SizeBytes = sizeBytes
		device.Size = formatBytes(sizeBytes)
	}

	if typeStr, ok := pairs["TYPE"]; ok {
		device.Type = typeStr
	}

	if mountpoint, ok := pairs["MOUNTPOINT"]; ok && mountpoint != "" && mountpoint != "[SWAP]" {
		device.MountPoint = mountpoint
	}

	return device
}

// getMountPoints parses df output
func getMountPoints() ([]models.MountPoint, error) {
	cmd := exec.Command("df", "-B1")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var mountPoints []models.MountPoint
	lines := strings.Split(string(output), "\n")

	// Skip header
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		mp := parseMountPointLine(line)
		if mp != nil {
			mountPoints = append(mountPoints, *mp)
		}
	}

	return mountPoints, nil
}

// parseMountPointLine parses a single df output line
func parseMountPointLine(line string) *models.MountPoint {
	fields := strings.Fields(line)
	if len(fields) < 6 {
		return nil
	}

	device := fields[0]
	total, _ := strconv.ParseInt(fields[1], 10, 64)
	used, _ := strconv.ParseInt(fields[2], 10, 64)
	available, _ := strconv.ParseInt(fields[3], 10, 64)
	usePercent, _ := strconv.ParseInt(strings.TrimSuffix(fields[4], "%"), 10, 32)
	mountPath := strings.Join(fields[5:], " ")

	// Get filesystem type from /etc/mtab or /proc/mounts
	fsType := getFSType(device)

	return &models.MountPoint{
		Path:       mountPath,
		Device:     device,
		FSType:     fsType,
		Total:      total,
		Used:       used,
		Available:  available,
		UsePercent: int32(usePercent),
	}
}

// getFSType gets filesystem type from /etc/mtab
func getFSType(device string) string {
	data, err := os.ReadFile("/etc/mtab")
	if err != nil {
		return "unknown"
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[0] == device {
			return fields[2]
		}
	}

	return "unknown"
}

// formatBytes converts bytes to human-readable format
func formatBytes(bytes int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	size := float64(bytes)

	for i := 0; i < len(units); i++ {
		if size < 1024 {
			if i == 0 {
				return strconv.FormatInt(int64(size), 10) + " " + units[i]
			}
			return strings.TrimRight(strings.TrimRight(strconv.FormatFloat(size, 'f', 2, 64), "0"), ".") + " " + units[i]
		}
		size /= 1024
	}

	return strconv.FormatInt(bytes, 10) + " B"
}
