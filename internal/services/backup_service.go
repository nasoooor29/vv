package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"visory/internal/database"
	"visory/internal/database/backups"
	"visory/internal/models"
	"visory/internal/utils"

	"github.com/labstack/echo/v4"
)

const (
	maxFirewallBackups = 10
	backupsVMDir       = "backups/vms"
	backupsDockerDir   = "backups/docker"
	backupsFirewallDir = "backups/firewall"
)

type BackupService struct {
	Dispatcher *utils.Dispatcher
	Logger     *slog.Logger
	db         *database.Service
	dataDir    string
}

// NewBackupService creates a new backup service
func NewBackupService(db *database.Service, dispatcher *utils.Dispatcher, logger *slog.Logger) *BackupService {
	service := &BackupService{
		Dispatcher: dispatcher.WithGroup("backup"),
		Logger:     logger.WithGroup("backup"),
		db:         db,
		dataDir:    models.ENV_VARS.Directory,
	}

	// Ensure backup directories exist
	if err := service.ensureBackupDirectories(); err != nil {
		logger.Error("failed to create backup directories", "error", err)
	}

	return service
}

// ensureBackupDirectories creates the backup directory structure
func (s *BackupService) ensureBackupDirectories() error {
	dirs := []string{
		filepath.Join(s.dataDir, backupsVMDir),
		filepath.Join(s.dataDir, backupsDockerDir),
		filepath.Join(s.dataDir, backupsFirewallDir),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// GetBackupDir returns the backup directory for a specific type
func (s *BackupService) GetBackupDir(backupType string) string {
	switch backupType {
	case "vm", "vm_snapshot", "vm_export":
		return filepath.Join(s.dataDir, backupsVMDir)
	case "container", "container_export", "container_commit":
		return filepath.Join(s.dataDir, backupsDockerDir)
	case "firewall":
		return filepath.Join(s.dataDir, backupsFirewallDir)
	default:
		return s.dataDir
	}
}

// ========================================
// Firewall Backup Functions
// ========================================

// CreateFirewallBackup creates a backup of the firewall rules
// This is called automatically when firewall rules change
func (s *BackupService) CreateFirewallBackup(rules []models.FirewallRule) error {
	backupDir := s.GetBackupDir("firewall")

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("firewall_%s.json", timestamp)
	filePath := filepath.Join(backupDir, filename)

	// Marshal rules to JSON
	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	// Write backup file
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	s.Logger.Info("created firewall backup", "file", filename, "rules", len(rules))

	// Record in database
	ctx := context.Background()
	now := time.Now()
	progress := int64(100)
	size := int64(len(data))
	targetName := "firewall"
	destination := filePath

	_, err = s.db.Backup.CreateBackupJob(ctx, backups.CreateBackupJobParams{
		Name:        fmt.Sprintf("Firewall Backup %s", timestamp),
		Type:        string(models.BackupTypeFirewall),
		TargetType:  string(models.BackupTargetFirewall),
		TargetID:    "firewall",
		TargetName:  &targetName,
		Status:      string(models.BackupStatusCompleted),
		Progress:    &progress,
		Destination: &destination,
		SizeBytes:   &size,
		StartedAt:   &now,
		CompletedAt: &now,
	})
	if err != nil {
		s.Logger.Warn("failed to record firewall backup in database", "error", err)
	}

	// Cleanup old backups (keep only last 10)
	if err := s.cleanupOldFirewallBackups(); err != nil {
		s.Logger.Warn("failed to cleanup old firewall backups", "error", err)
	}

	return nil
}

// cleanupOldFirewallBackups removes old firewall backups, keeping only the last 10
func (s *BackupService) cleanupOldFirewallBackups() error {
	backupDir := s.GetBackupDir("firewall")

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	// Filter and sort firewall backup files
	var backupFiles []os.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "firewall_") && strings.HasSuffix(entry.Name(), ".json") {
			backupFiles = append(backupFiles, entry)
		}
	}

	// Sort by name (which includes timestamp, so chronological order)
	sort.Slice(backupFiles, func(i, j int) bool {
		return backupFiles[i].Name() < backupFiles[j].Name()
	})

	// Delete old backups if we have more than maxFirewallBackups
	if len(backupFiles) > maxFirewallBackups {
		toDelete := backupFiles[:len(backupFiles)-maxFirewallBackups]
		for _, file := range toDelete {
			filePath := filepath.Join(backupDir, file.Name())
			if err := os.Remove(filePath); err != nil {
				s.Logger.Warn("failed to delete old backup", "file", file.Name(), "error", err)
			} else {
				s.Logger.Info("deleted old firewall backup", "file", file.Name())
			}
		}
	}

	return nil
}

// ListFirewallBackups returns a list of available firewall backups
//
//	@Summary      list firewall backups
//	@Description  returns all available firewall backups
//	@Tags         backup
//	@Produce      json
//	@Success      200  {array}   models.FirewallBackup
//	@Failure      401  {object}  models.HTTPError
//	@Failure      403  {object}  models.HTTPError
//	@Failure      500  {object}  models.HTTPError
//	@Router       /backup/firewall [get]
func (s *BackupService) ListFirewallBackups(c echo.Context) error {
	backupDir := s.GetBackupDir("firewall")

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return s.Dispatcher.NewInternalServerError("failed to read backup directory", err)
	}

	var backupList []models.FirewallBackup
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "firewall_") || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Read file to count rules
		filePath := filepath.Join(backupDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var rules []models.FirewallRule
		if err := json.Unmarshal(data, &rules); err != nil {
			continue
		}

		backupList = append(backupList, models.FirewallBackup{
			Filename:  entry.Name(),
			SizeBytes: info.Size(),
			RuleCount: len(rules),
			CreatedAt: info.ModTime(),
		})
	}

	// Sort by creation time (newest first)
	sort.Slice(backupList, func(i, j int) bool {
		return backupList[i].CreatedAt.After(backupList[j].CreatedAt)
	})

	return c.JSON(http.StatusOK, backupList)
}

// GetFirewallBackup returns the contents of a specific firewall backup
//
//	@Summary      get firewall backup
//	@Description  returns the contents of a specific firewall backup
//	@Tags         backup
//	@Produce      json
//	@Param        filename  path      string  true  "Backup filename"
//	@Success      200       {array}   models.FirewallRule
//	@Failure      400       {object}  models.HTTPError
//	@Failure      401       {object}  models.HTTPError
//	@Failure      403       {object}  models.HTTPError
//	@Failure      404       {object}  models.HTTPError
//	@Failure      500       {object}  models.HTTPError
//	@Router       /backup/firewall/{filename} [get]
func (s *BackupService) GetFirewallBackup(c echo.Context) error {
	filename := c.Param("filename")
	if filename == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "filename is required")
	}

	// Security check: prevent directory traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid filename")
	}

	filePath := filepath.Join(s.GetBackupDir("firewall"), filename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return echo.NewHTTPError(http.StatusNotFound, "backup not found")
		}
		return s.Dispatcher.NewInternalServerError("failed to read backup file", err)
	}

	var rules []models.FirewallRule
	if err := json.Unmarshal(data, &rules); err != nil {
		return s.Dispatcher.NewInternalServerError("failed to parse backup file", err)
	}

	return c.JSON(http.StatusOK, rules)
}

// DeleteFirewallBackup deletes a specific firewall backup
//
//	@Summary      delete firewall backup
//	@Description  deletes a specific firewall backup
//	@Tags         backup
//	@Param        filename  path  string  true  "Backup filename"
//	@Success      204
//	@Failure      400  {object}  models.HTTPError
//	@Failure      401  {object}  models.HTTPError
//	@Failure      403  {object}  models.HTTPError
//	@Failure      404  {object}  models.HTTPError
//	@Failure      500  {object}  models.HTTPError
//	@Router       /backup/firewall/{filename} [delete]
func (s *BackupService) DeleteFirewallBackup(c echo.Context) error {
	filename := c.Param("filename")
	if filename == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "filename is required")
	}

	// Security check: prevent directory traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid filename")
	}

	filePath := filepath.Join(s.GetBackupDir("firewall"), filename)

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return echo.NewHTTPError(http.StatusNotFound, "backup not found")
		}
		return s.Dispatcher.NewInternalServerError("failed to delete backup file", err)
	}

	s.Logger.Info("deleted firewall backup", "file", filename)
	return c.NoContent(http.StatusNoContent)
}

// ========================================
// Backup Jobs API
// ========================================

// ListBackupJobs returns all backup jobs
//
//	@Summary      list backup jobs
//	@Description  returns all backup job history
//	@Tags         backup
//	@Produce      json
//	@Param        limit   query     int     false  "Limit"   default(50)
//	@Param        offset  query     int     false  "Offset"  default(0)
//	@Param        type    query     string  false  "Filter by type"
//	@Success      200     {array}   backups.BackupJob
//	@Failure      401     {object}  models.HTTPError
//	@Failure      403     {object}  models.HTTPError
//	@Failure      500     {object}  models.HTTPError
//	@Router       /backup/jobs [get]
func (s *BackupService) ListBackupJobs(c echo.Context) error {
	ctx := c.Request().Context()

	limit := int64(50)
	offset := int64(0)

	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil {
			limit = parsed
		}
	}
	if o := c.QueryParam("offset"); o != "" {
		if parsed, err := strconv.ParseInt(o, 10, 64); err == nil {
			offset = parsed
		}
	}

	typeFilter := c.QueryParam("type")

	var jobs []backups.BackupJob
	var err error

	if typeFilter != "" {
		jobs, err = s.db.Backup.ListBackupJobsByType(ctx, backups.ListBackupJobsByTypeParams{
			Type:   typeFilter,
			Limit:  limit,
			Offset: offset,
		})
	} else {
		jobs, err = s.db.Backup.ListBackupJobs(ctx, backups.ListBackupJobsParams{
			Limit:  limit,
			Offset: offset,
		})
	}

	if err != nil {
		return s.Dispatcher.NewInternalServerError("failed to list backup jobs", err)
	}

	if jobs == nil {
		jobs = []backups.BackupJob{}
	}

	return c.JSON(http.StatusOK, jobs)
}

// GetBackupJob returns a specific backup job
//
//	@Summary      get backup job
//	@Description  returns a specific backup job by ID
//	@Tags         backup
//	@Produce      json
//	@Param        id   path      int  true  "Job ID"
//	@Success      200  {object}  backups.BackupJob
//	@Failure      400  {object}  models.HTTPError
//	@Failure      401  {object}  models.HTTPError
//	@Failure      403  {object}  models.HTTPError
//	@Failure      404  {object}  models.HTTPError
//	@Failure      500  {object}  models.HTTPError
//	@Router       /backup/jobs/{id} [get]
func (s *BackupService) GetBackupJob(c echo.Context) error {
	ctx := c.Request().Context()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid job ID")
	}

	job, err := s.db.Backup.GetBackupJob(ctx, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "backup job not found")
	}

	return c.JSON(http.StatusOK, job)
}

// DeleteBackupJob deletes a backup job record
//
//	@Summary      delete backup job
//	@Description  deletes a backup job record (not the backup file)
//	@Tags         backup
//	@Param        id  path  int  true  "Job ID"
//	@Success      204
//	@Failure      400  {object}  models.HTTPError
//	@Failure      401  {object}  models.HTTPError
//	@Failure      403  {object}  models.HTTPError
//	@Failure      500  {object}  models.HTTPError
//	@Router       /backup/jobs/{id} [delete]
func (s *BackupService) DeleteBackupJob(c echo.Context) error {
	ctx := c.Request().Context()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid job ID")
	}

	if err := s.db.Backup.DeleteBackupJob(ctx, id); err != nil {
		return s.Dispatcher.NewInternalServerError("failed to delete backup job", err)
	}

	return c.NoContent(http.StatusNoContent)
}

// GetBackupStats returns backup statistics
//
//	@Summary      get backup stats
//	@Description  returns backup statistics
//	@Tags         backup
//	@Produce      json
//	@Success      200  {object}  models.BackupStats
//	@Failure      401  {object}  models.HTTPError
//	@Failure      403  {object}  models.HTTPError
//	@Failure      500  {object}  models.HTTPError
//	@Router       /backup/stats [get]
func (s *BackupService) GetBackupStats(c echo.Context) error {
	ctx := c.Request().Context()

	dbStats, err := s.db.Backup.GetBackupStats(ctx)
	if err != nil {
		return s.Dispatcher.NewInternalServerError("failed to get backup stats", err)
	}

	activeSchedules, err := s.db.Backup.CountActiveSchedules(ctx)
	if err != nil {
		activeSchedules = 0
	}

	// Convert interface{} fields to proper types
	var totalSizeBytes int64
	if dbStats.TotalSizeBytes != nil {
		switch v := dbStats.TotalSizeBytes.(type) {
		case int64:
			totalSizeBytes = v
		case float64:
			totalSizeBytes = int64(v)
		}
	}

	var vmSnapshots, containerBackups, firewallBackups int
	if dbStats.VmSnapshots != nil {
		vmSnapshots = int(*dbStats.VmSnapshots)
	}
	if dbStats.ContainerBackups != nil {
		containerBackups = int(*dbStats.ContainerBackups)
	}
	if dbStats.FirewallBackups != nil {
		firewallBackups = int(*dbStats.FirewallBackups)
	}

	stats := models.BackupStats{
		TotalBackups:     int(dbStats.TotalBackups),
		VMSnapshots:      vmSnapshots,
		ContainerBackups: containerBackups,
		FirewallBackups:  firewallBackups,
		TotalSizeBytes:   totalSizeBytes,
		ActiveSchedules:  int(activeSchedules),
	}

	return c.JSON(http.StatusOK, stats)
}

// ========================================
// Backup Schedules API
// ========================================

// ListBackupSchedules returns all backup schedules
//
//	@Summary      list backup schedules
//	@Description  returns all configured backup schedules
//	@Tags         backup
//	@Produce      json
//	@Success      200  {array}   backups.BackupSchedule
//	@Failure      401  {object}  models.HTTPError
//	@Failure      403  {object}  models.HTTPError
//	@Failure      500  {object}  models.HTTPError
//	@Router       /backup/schedules [get]
func (s *BackupService) ListBackupSchedules(c echo.Context) error {
	ctx := c.Request().Context()

	schedules, err := s.db.Backup.ListBackupSchedules(ctx)
	if err != nil {
		return s.Dispatcher.NewInternalServerError("failed to list backup schedules", err)
	}

	if schedules == nil {
		schedules = []backups.BackupSchedule{}
	}

	return c.JSON(http.StatusOK, schedules)
}

// GetBackupSchedule returns a specific backup schedule
//
//	@Summary      get backup schedule
//	@Description  returns a specific backup schedule by ID
//	@Tags         backup
//	@Produce      json
//	@Param        id   path      int  true  "Schedule ID"
//	@Success      200  {object}  backups.BackupSchedule
//	@Failure      400  {object}  models.HTTPError
//	@Failure      401  {object}  models.HTTPError
//	@Failure      403  {object}  models.HTTPError
//	@Failure      404  {object}  models.HTTPError
//	@Failure      500  {object}  models.HTTPError
//	@Router       /backup/schedules/{id} [get]
func (s *BackupService) GetBackupSchedule(c echo.Context) error {
	ctx := c.Request().Context()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid schedule ID")
	}

	schedule, err := s.db.Backup.GetBackupSchedule(ctx, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "backup schedule not found")
	}

	return c.JSON(http.StatusOK, schedule)
}

// CreateBackupSchedule creates a new backup schedule
//
//	@Summary      create backup schedule
//	@Description  creates a new automated backup schedule
//	@Tags         backup
//	@Accept       json
//	@Produce      json
//	@Param        schedule  body      models.CreateBackupScheduleRequest  true  "Schedule configuration"
//	@Success      201       {object}  backups.BackupSchedule
//	@Failure      400       {object}  models.HTTPError
//	@Failure      401       {object}  models.HTTPError
//	@Failure      403       {object}  models.HTTPError
//	@Failure      500       {object}  models.HTTPError
//	@Router       /backup/schedules [post]
func (s *BackupService) CreateBackupSchedule(c echo.Context) error {
	ctx := c.Request().Context()

	var req models.CreateBackupScheduleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Validate schedule type
	validSchedules := map[string]bool{
		"hourly": true,
		"daily":  true,
		"weekly": true,
	}
	if !validSchedules[req.Schedule] {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid schedule: must be hourly, daily, or weekly")
	}

	// Calculate next run time
	nextRun := s.calculateNextRunTime(req.Schedule, req.ScheduleTime)

	// Default retention count
	retentionCount := int64(7)
	if req.RetentionCount > 0 {
		retentionCount = int64(req.RetentionCount)
	}

	enabled := int64(1)

	// Determine backup destination
	destination := s.GetBackupDir(string(req.TargetType))

	schedule, err := s.db.Backup.CreateBackupSchedule(ctx, backups.CreateBackupScheduleParams{
		Name:           req.Name,
		Type:           string(req.Type),
		TargetType:     string(req.TargetType),
		TargetID:       req.TargetID,
		TargetName:     &req.TargetName,
		ClientID:       req.ClientID,
		Schedule:       req.Schedule,
		ScheduleTime:   &req.ScheduleTime,
		Destination:    destination,
		RetentionCount: &retentionCount,
		Enabled:        &enabled,
		NextRunAt:      &nextRun,
	})
	if err != nil {
		return s.Dispatcher.NewInternalServerError("failed to create backup schedule", err)
	}

	s.Logger.Info("created backup schedule", "id", schedule.ID, "name", req.Name, "schedule", req.Schedule)
	return c.JSON(http.StatusCreated, schedule)
}

// UpdateBackupSchedule updates an existing backup schedule
//
//	@Summary      update backup schedule
//	@Description  updates an existing backup schedule
//	@Tags         backup
//	@Accept       json
//	@Produce      json
//	@Param        id        path      int                                 true  "Schedule ID"
//	@Param        schedule  body      models.UpdateBackupScheduleRequest  true  "Schedule updates"
//	@Success      200       {object}  backups.BackupSchedule
//	@Failure      400       {object}  models.HTTPError
//	@Failure      401       {object}  models.HTTPError
//	@Failure      403       {object}  models.HTTPError
//	@Failure      404       {object}  models.HTTPError
//	@Failure      500       {object}  models.HTTPError
//	@Router       /backup/schedules/{id} [put]
func (s *BackupService) UpdateBackupSchedule(c echo.Context) error {
	ctx := c.Request().Context()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid schedule ID")
	}

	var req models.UpdateBackupScheduleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Get current schedule to validate it exists
	current, err := s.db.Backup.GetBackupSchedule(ctx, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "backup schedule not found")
	}

	// Build update params
	updateParams := backups.UpdateBackupScheduleParams{
		ID: id,
	}

	if req.Name != nil {
		updateParams.Name = *req.Name
	} else {
		updateParams.Name = current.Name
	}

	if req.Schedule != nil {
		updateParams.Schedule = *req.Schedule
	} else {
		updateParams.Schedule = current.Schedule
	}

	if req.ScheduleTime != nil {
		updateParams.ScheduleTime = req.ScheduleTime
	} else {
		updateParams.ScheduleTime = current.ScheduleTime
	}

	if req.RetentionCount != nil {
		rc := int64(*req.RetentionCount)
		updateParams.RetentionCount = &rc
	} else {
		updateParams.RetentionCount = current.RetentionCount
	}

	if req.Enabled != nil {
		var e int64
		if *req.Enabled {
			e = 1
		}
		updateParams.Enabled = &e
	} else {
		updateParams.Enabled = current.Enabled
	}

	schedule, err := s.db.Backup.UpdateBackupSchedule(ctx, updateParams)
	if err != nil {
		return s.Dispatcher.NewInternalServerError("failed to update backup schedule", err)
	}

	s.Logger.Info("updated backup schedule", "id", id)
	return c.JSON(http.StatusOK, schedule)
}

// DeleteBackupSchedule deletes a backup schedule
//
//	@Summary      delete backup schedule
//	@Description  deletes a backup schedule
//	@Tags         backup
//	@Param        id  path  int  true  "Schedule ID"
//	@Success      204
//	@Failure      400  {object}  models.HTTPError
//	@Failure      401  {object}  models.HTTPError
//	@Failure      403  {object}  models.HTTPError
//	@Failure      500  {object}  models.HTTPError
//	@Router       /backup/schedules/{id} [delete]
func (s *BackupService) DeleteBackupSchedule(c echo.Context) error {
	ctx := c.Request().Context()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid schedule ID")
	}

	if err := s.db.Backup.DeleteBackupSchedule(ctx, id); err != nil {
		return s.Dispatcher.NewInternalServerError("failed to delete backup schedule", err)
	}

	s.Logger.Info("deleted backup schedule", "id", id)
	return c.NoContent(http.StatusNoContent)
}

// ToggleBackupSchedule enables or disables a backup schedule
//
//	@Summary      toggle backup schedule
//	@Description  enables or disables a backup schedule
//	@Tags         backup
//	@Produce      json
//	@Param        id  path      int  true  "Schedule ID"
//	@Success      200  {object}  backups.BackupSchedule
//	@Failure      400  {object}  models.HTTPError
//	@Failure      401  {object}  models.HTTPError
//	@Failure      403  {object}  models.HTTPError
//	@Failure      404  {object}  models.HTTPError
//	@Failure      500  {object}  models.HTTPError
//	@Router       /backup/schedules/{id}/toggle [post]
func (s *BackupService) ToggleBackupSchedule(c echo.Context) error {
	ctx := c.Request().Context()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid schedule ID")
	}

	// Get current schedule
	current, err := s.db.Backup.GetBackupSchedule(ctx, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "backup schedule not found")
	}

	var schedule backups.BackupSchedule
	if current.Enabled != nil && *current.Enabled == 1 {
		schedule, err = s.db.Backup.DisableBackupSchedule(ctx, id)
	} else {
		schedule, err = s.db.Backup.EnableBackupSchedule(ctx, id)
	}

	if err != nil {
		return s.Dispatcher.NewInternalServerError("failed to toggle backup schedule", err)
	}

	s.Logger.Info("toggled backup schedule", "id", id, "enabled", schedule.Enabled)
	return c.JSON(http.StatusOK, schedule)
}

// calculateNextRunTime calculates the next run time based on schedule
func (s *BackupService) calculateNextRunTime(schedule, scheduleTime string) time.Time {
	now := time.Now()

	switch schedule {
	case "hourly":
		return now.Add(time.Hour).Truncate(time.Hour)
	case "daily":
		// Parse schedule time (e.g., "02:00")
		hour, minute := 2, 0
		if scheduleTime != "" {
			parts := strings.Split(scheduleTime, ":")
			if len(parts) >= 1 {
				if h, err := strconv.Atoi(parts[0]); err == nil {
					hour = h
				}
			}
			if len(parts) >= 2 {
				if m, err := strconv.Atoi(parts[1]); err == nil {
					minute = m
				}
			}
		}
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
		if next.Before(now) {
			next = next.Add(24 * time.Hour)
		}
		return next
	case "weekly":
		// Next Sunday at the specified time
		hour, minute := 2, 0
		if scheduleTime != "" {
			parts := strings.Split(scheduleTime, ":")
			if len(parts) >= 1 {
				if h, err := strconv.Atoi(parts[0]); err == nil {
					hour = h
				}
			}
			if len(parts) >= 2 {
				if m, err := strconv.Atoi(parts[1]); err == nil {
					minute = m
				}
			}
		}
		daysUntilSunday := (7 - int(now.Weekday())) % 7
		if daysUntilSunday == 0 {
			daysUntilSunday = 7
		}
		next := time.Date(now.Year(), now.Month(), now.Day()+daysUntilSunday, hour, minute, 0, 0, now.Location())
		return next
	default:
		return now.Add(24 * time.Hour)
	}
}

// ========================================
// Container Backup Functions
// ========================================

// ExportContainer exports a container to a tar file
//
//	@Summary      export container
//	@Description  exports a container filesystem to a tar file
//	@Tags         backup
//	@Accept       json
//	@Produce      json
//	@Param        clientId  path      string                              true  "Docker client ID"
//	@Param        id        path      string                              true  "Container ID"
//	@Param        request   body      models.CreateContainerBackupRequest true  "Backup configuration"
//	@Success      201       {object}  models.BackupActionResponse
//	@Failure      400       {object}  models.HTTPError
//	@Failure      401       {object}  models.HTTPError
//	@Failure      403       {object}  models.HTTPError
//	@Failure      500       {object}  models.HTTPError
//	@Router       /backup/container/{clientId}/{id}/export [post]
func (s *BackupService) ExportContainer(c echo.Context, dockerClient interface {
	ContainerExport(ctx context.Context, containerID string) (io.ReadCloser, error)
},
) error {
	ctx := c.Request().Context()

	clientID := c.Param("clientId")
	containerID := c.Param("id")

	var req models.CreateContainerBackupRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Name == "" {
		req.Name = fmt.Sprintf("container_%s", containerID[:12])
	}

	// Generate backup filename
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.tar", req.Name, timestamp)
	filePath := filepath.Join(s.GetBackupDir("container"), filename)

	// Create backup job record
	now := time.Now()
	progress := int64(0)
	targetName := req.Name
	destination := filePath

	job, err := s.db.Backup.CreateBackupJob(ctx, backups.CreateBackupJobParams{
		Name:        fmt.Sprintf("Container Export: %s", req.Name),
		Type:        string(models.BackupTypeContainerExport),
		TargetType:  string(models.BackupTargetContainer),
		TargetID:    containerID,
		TargetName:  &targetName,
		ClientID:    &clientID,
		Status:      string(models.BackupStatusRunning),
		Progress:    &progress,
		Destination: &destination,
		StartedAt:   &now,
	})
	if err != nil {
		return s.Dispatcher.NewInternalServerError("failed to create backup job", err)
	}

	// Export container
	reader, err := dockerClient.ContainerExport(ctx, containerID)
	if err != nil {
		errMsg := err.Error()
		s.db.Backup.UpdateBackupJobFailed(ctx, backups.UpdateBackupJobFailedParams{
			ID:           job.ID,
			ErrorMessage: &errMsg,
			CompletedAt:  &now,
		})
		return s.Dispatcher.NewInternalServerError("failed to export container", err)
	}
	defer reader.Close()

	// Write to file
	file, err := os.Create(filePath)
	if err != nil {
		errMsg := err.Error()
		s.db.Backup.UpdateBackupJobFailed(ctx, backups.UpdateBackupJobFailedParams{
			ID:           job.ID,
			ErrorMessage: &errMsg,
			CompletedAt:  &now,
		})
		return s.Dispatcher.NewInternalServerError("failed to create backup file", err)
	}
	defer file.Close()

	written, err := io.Copy(file, reader)
	if err != nil {
		errMsg := err.Error()
		s.db.Backup.UpdateBackupJobFailed(ctx, backups.UpdateBackupJobFailedParams{
			ID:           job.ID,
			ErrorMessage: &errMsg,
			CompletedAt:  &now,
		})
		return s.Dispatcher.NewInternalServerError("failed to write backup file", err)
	}

	// Update job as completed
	completedAt := time.Now()
	s.db.Backup.UpdateBackupJobCompleted(ctx, backups.UpdateBackupJobCompletedParams{
		ID:          job.ID,
		CompletedAt: &completedAt,
		SizeBytes:   &written,
	})

	s.Logger.Info("exported container", "container", containerID, "file", filename, "size", written)

	return c.JSON(http.StatusCreated, models.BackupActionResponse{
		Success: true,
		Message: fmt.Sprintf("Container exported to %s", filename),
		JobID:   &job.ID,
	})
}

// ListContainerBackups returns all container backups
//
//	@Summary      list container backups
//	@Description  returns all available container backups
//	@Tags         backup
//	@Produce      json
//	@Success      200  {array}   models.ContainerBackup
//	@Failure      401  {object}  models.HTTPError
//	@Failure      403  {object}  models.HTTPError
//	@Failure      500  {object}  models.HTTPError
//	@Router       /backup/containers [get]
func (s *BackupService) ListContainerBackups(c echo.Context) error {
	backupDir := s.GetBackupDir("container")

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return s.Dispatcher.NewInternalServerError("failed to read backup directory", err)
	}

	var backupList []models.ContainerBackup
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tar") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		backupList = append(backupList, models.ContainerBackup{
			Filename:  entry.Name(),
			SizeBytes: info.Size(),
			CreatedAt: info.ModTime(),
		})
	}

	// Sort by creation time (newest first)
	sort.Slice(backupList, func(i, j int) bool {
		return backupList[i].CreatedAt.After(backupList[j].CreatedAt)
	})

	return c.JSON(http.StatusOK, backupList)
}

// DeleteContainerBackup deletes a container backup file
//
//	@Summary      delete container backup
//	@Description  deletes a container backup file
//	@Tags         backup
//	@Param        filename  path  string  true  "Backup filename"
//	@Success      204
//	@Failure      400  {object}  models.HTTPError
//	@Failure      401  {object}  models.HTTPError
//	@Failure      403  {object}  models.HTTPError
//	@Failure      404  {object}  models.HTTPError
//	@Failure      500  {object}  models.HTTPError
//	@Router       /backup/containers/{filename} [delete]
func (s *BackupService) DeleteContainerBackup(c echo.Context) error {
	filename := c.Param("filename")
	if filename == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "filename is required")
	}

	// Security check: prevent directory traversal
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid filename")
	}

	filePath := filepath.Join(s.GetBackupDir("container"), filename)

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return echo.NewHTTPError(http.StatusNotFound, "backup not found")
		}
		return s.Dispatcher.NewInternalServerError("failed to delete backup file", err)
	}

	s.Logger.Info("deleted container backup", "file", filename)
	return c.NoContent(http.StatusNoContent)
}
