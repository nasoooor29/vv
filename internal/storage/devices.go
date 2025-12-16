package storage

import (
	"os/exec"
	"strconv"
	"strings"

	"visory/internal/models"
)

// GetBlockDevices retrieves list of storage devices
func GetBlockDevices() ([]models.StorageDevice, error) {
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
		device.Size = FormatBytes(sizeBytes)
	}

	if typeStr, ok := pairs["TYPE"]; ok {
		device.Type = typeStr
	}

	if mountpoint, ok := pairs["MOUNTPOINT"]; ok && mountpoint != "" && mountpoint != "[SWAP]" {
		device.MountPoint = mountpoint
	}

	return device
}
