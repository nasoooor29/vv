package models

type StorageDevice struct {
	Name         string `json:"name"`
	Size         string `json:"size"`
	SizeBytes    int64  `json:"size_bytes"`
	Type         string `json:"type"`
	MountPoint   string `json:"mount_point"`
	UsagePercent int32  `json:"usage_percent"`
}

type MountPoint struct {
	Path       string `json:"path"`
	Device     string `json:"device"`
	FSType     string `json:"fs_type"`
	Total      int64  `json:"total"`
	Used       int64  `json:"used"`
	Available  int64  `json:"available"`
	UsePercent int32  `json:"use_percent"`
}
