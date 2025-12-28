package models

import "time"

// BackupType represents the type of backup
type BackupType string

const (
	BackupTypeVMSnapshot      BackupType = "vm_snapshot"
	BackupTypeVMExport        BackupType = "vm_export"
	BackupTypeContainerExport BackupType = "container_export"
	BackupTypeContainerCommit BackupType = "container_commit"
	BackupTypeFirewall        BackupType = "firewall"
)

// BackupTargetType represents the type of target being backed up
type BackupTargetType string

const (
	BackupTargetVM        BackupTargetType = "vm"
	BackupTargetContainer BackupTargetType = "container"
	BackupTargetFirewall  BackupTargetType = "firewall"
)

// BackupStatus represents the status of a backup job
type BackupStatus string

const (
	BackupStatusPending   BackupStatus = "pending"
	BackupStatusRunning   BackupStatus = "running"
	BackupStatusCompleted BackupStatus = "completed"
	BackupStatusFailed    BackupStatus = "failed"
)

// BackupJob represents a backup or restore operation
type BackupJob struct {
	ID           int64            `json:"id"`
	Name         string           `json:"name"`
	Type         BackupType       `json:"type"`
	TargetType   BackupTargetType `json:"target_type"`
	TargetID     string           `json:"target_id"`
	TargetName   string           `json:"target_name"`
	ClientID     *string          `json:"client_id,omitempty"` // For Docker containers
	Status       BackupStatus     `json:"status"`
	Progress     int              `json:"progress"`
	Destination  string           `json:"destination"`
	SizeBytes    *int64           `json:"size_bytes,omitempty"`
	ErrorMessage *string          `json:"error_message,omitempty"`
	StartedAt    *time.Time       `json:"started_at,omitempty"`
	CompletedAt  *time.Time       `json:"completed_at,omitempty"`
	CreatedBy    *int64           `json:"created_by,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
}

// BackupSchedule represents a scheduled backup configuration
type BackupSchedule struct {
	ID             int64            `json:"id"`
	Name           string           `json:"name"`
	Type           BackupType       `json:"type"`
	TargetType     BackupTargetType `json:"target_type"`
	TargetID       string           `json:"target_id"`
	TargetName     string           `json:"target_name"`
	ClientID       *string          `json:"client_id,omitempty"`
	Schedule       string           `json:"schedule"`      // Simple schedule: "daily", "hourly", etc.
	ScheduleTime   string           `json:"schedule_time"` // Time of day for daily backups: "02:00"
	Destination    string           `json:"destination"`
	RetentionCount int              `json:"retention_count"` // Keep last N backups
	Enabled        bool             `json:"enabled"`
	LastRunAt      *time.Time       `json:"last_run_at,omitempty"`
	NextRunAt      *time.Time       `json:"next_run_at,omitempty"`
	LastStatus     *string          `json:"last_status,omitempty"`
	CreatedBy      *int64           `json:"created_by,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

// FirewallBackup represents a firewall configuration backup
type FirewallBackup struct {
	ID        int64     `json:"id"`
	Filename  string    `json:"filename"`
	SizeBytes int64     `json:"size_bytes"`
	RuleCount int       `json:"rule_count"`
	CreatedAt time.Time `json:"created_at"`
}

// VMSnapshot represents a libvirt VM snapshot
type VMSnapshot struct {
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	State       string     `json:"state"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	IsCurrent   bool       `json:"is_current"`
}

// ContainerBackup represents a Docker container backup
type ContainerBackup struct {
	ID          string    `json:"id"`
	Filename    string    `json:"filename"`
	ContainerID string    `json:"container_id"`
	ClientID    string    `json:"client_id"`
	SizeBytes   int64     `json:"size_bytes"`
	CreatedAt   time.Time `json:"created_at"`
}

// Request types

// CreateVMSnapshotRequest is the request to create a VM snapshot
type CreateVMSnapshotRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description,omitempty"`
}

// CreateContainerBackupRequest is the request to backup a container
type CreateContainerBackupRequest struct {
	Name string `json:"name" validate:"required"`
}

// RestoreVMSnapshotRequest is the request to restore a VM snapshot
type RestoreVMSnapshotRequest struct {
	SnapshotName string `json:"snapshot_name" validate:"required"`
}

// RestoreContainerBackupRequest is the request to restore a container from backup
type RestoreContainerBackupRequest struct {
	BackupFile    string `json:"backup_file" validate:"required"`
	ContainerName string `json:"container_name,omitempty"` // If empty, use original name
}

// CreateBackupScheduleRequest is the request to create a backup schedule
type CreateBackupScheduleRequest struct {
	Name           string           `json:"name" validate:"required"`
	Type           BackupType       `json:"type" validate:"required"`
	TargetType     BackupTargetType `json:"target_type" validate:"required"`
	TargetID       string           `json:"target_id" validate:"required"`
	TargetName     string           `json:"target_name"`
	ClientID       *string          `json:"client_id,omitempty"`
	Schedule       string           `json:"schedule" validate:"required"` // "daily", "weekly"
	ScheduleTime   string           `json:"schedule_time"`                // "02:00" for daily
	RetentionCount int              `json:"retention_count"`
}

// UpdateBackupScheduleRequest is the request to update a backup schedule
type UpdateBackupScheduleRequest struct {
	Name           *string `json:"name,omitempty"`
	Schedule       *string `json:"schedule,omitempty"`
	ScheduleTime   *string `json:"schedule_time,omitempty"`
	RetentionCount *int    `json:"retention_count,omitempty"`
	Enabled        *bool   `json:"enabled,omitempty"`
}

// Response types

// BackupActionResponse is a generic response for backup actions
type BackupActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	JobID   *int64 `json:"job_id,omitempty"`
}

// BackupStats represents backup statistics
type BackupStats struct {
	TotalBackups     int        `json:"total_backups"`
	VMSnapshots      int        `json:"vm_snapshots"`
	ContainerBackups int        `json:"container_backups"`
	FirewallBackups  int        `json:"firewall_backups"`
	TotalSizeBytes   int64      `json:"total_size_bytes"`
	ActiveSchedules  int        `json:"active_schedules"`
	LastBackupAt     *time.Time `json:"last_backup_at,omitempty"`
}
