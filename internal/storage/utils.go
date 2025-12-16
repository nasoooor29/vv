package storage

import (
	"os"
	"strconv"
	"strings"
)

// FormatBytes converts bytes to human-readable format
func FormatBytes(bytes int64) string {
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

// GetFSType gets filesystem type from /etc/mtab
func GetFSType(device string) string {
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
