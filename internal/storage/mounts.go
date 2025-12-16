package storage

import (
	"os/exec"
	"strconv"
	"strings"

	"visory/internal/models"
)

// GetMountPoints retrieves list of mount points
func GetMountPoints() ([]models.MountPoint, error) {
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
	fsType := GetFSType(device)

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
