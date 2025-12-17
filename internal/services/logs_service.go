package services

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"visory/internal/database"
	"visory/internal/database/logs"

	"github.com/labstack/echo/v4"
)

type LogsService struct {
	db     *database.Service
	logger *slog.Logger
}

// NewLogsService creates a new LogsService with dependency injection
func NewLogsService(db *database.Service, logger *slog.Logger) *LogsService {
	return &LogsService{
		db:     db,
		logger: logger.WithGroup("logs"),
	}
}

// GetLogsRequest represents query parameters for log filtering
type GetLogsRequest struct {
	ServiceGroup string `query:"service_group"`
	Level        string `query:"level"`
	Page         int    `query:"page"`
	PageSize     int    `query:"page_size"`
	Days         int    `query:"days"` // Filter logs from last N days
}

// LogResponse represents a log entry for the API
type LogResponse struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	Action       string    `json:"action"`
	Details      *string   `json:"details"`
	ServiceGroup string    `json:"service_group"`
	Level        string    `json:"level"`
	CreatedAt    time.Time `json:"created_at"`
}

// GetLogsResponse represents paginated log results
type GetLogsResponse struct {
	Logs       []LogResponse `json:"logs"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int64         `json:"total_pages"`
}

// GetLogs retrieves logs with filtering and pagination
func (s *LogsService) GetLogs(c echo.Context) error {
	req := new(GetLogsRequest)
	if err := c.Bind(req); err != nil {
		s.logger.Error("failed to parse query params", "error", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid query parameters")
	}

	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	if req.Days < 1 {
		req.Days = 7 // Default to last 7 days
	}

	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	ctx := context.Background()

	// Build the query based on filters
	var logsList []logs.Log
	var total int64
	var err error

	since := time.Now().AddDate(0, 0, -req.Days)

	if req.ServiceGroup != "" && req.Level != "" {
		// Filter by both service_group and level
		logsList, err = s.db.Log.GetLogsByServiceGroupAndLevel(ctx, logs.GetLogsByServiceGroupAndLevelParams{
			ServiceGroup: req.ServiceGroup,
			Level:        req.Level,
			CreatedAt:    since,
			Limit:        limit,
			Offset:       offset,
		})
		if err != nil && err != sql.ErrNoRows {
			s.logger.Error("failed to get logs", "error", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrieve logs")
		}
		// Get total count
		total, err = s.db.Log.CountLogsByServiceGroupAndLevel(ctx, logs.CountLogsByServiceGroupAndLevelParams{
			ServiceGroup: req.ServiceGroup,
			Level:        req.Level,
			CreatedAt:    since,
		})
	} else if req.ServiceGroup != "" {
		// Filter by service_group only
		logsList, err = s.db.Log.GetLogsByServiceGroup(ctx, logs.GetLogsByServiceGroupParams{
			ServiceGroup: req.ServiceGroup,
			CreatedAt:    since,
			Limit:        limit,
			Offset:       offset,
		})
		if err != nil && err != sql.ErrNoRows {
			s.logger.Error("failed to get logs", "error", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrieve logs")
		}
		total, err = s.db.Log.CountLogsByServiceGroup(ctx, logs.CountLogsByServiceGroupParams{
			ServiceGroup: req.ServiceGroup,
			CreatedAt:    since,
		})
	} else if req.Level != "" {
		// Filter by level only
		logsList, err = s.db.Log.GetLogsByLevel(ctx, logs.GetLogsByLevelParams{
			Level:     req.Level,
			CreatedAt: since,
			Limit:     limit,
			Offset:    offset,
		})
		if err != nil && err != sql.ErrNoRows {
			s.logger.Error("failed to get logs", "error", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrieve logs")
		}
		total, err = s.db.Log.CountLogsByLevel(ctx, logs.CountLogsByLevelParams{
			Level:     req.Level,
			CreatedAt: since,
		})
	} else {
		// Get all logs with date filter
		logsList, err = s.db.Log.GetLogsPaginated(ctx, logs.GetLogsPaginatedParams{
			CreatedAt: since,
			Limit:     limit,
			Offset:    offset,
		})
		if err != nil && err != sql.ErrNoRows {
			s.logger.Error("failed to get logs", "error", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrieve logs")
		}
		total, err = s.db.Log.CountLogs(ctx, since)
	}

	if err != nil && err != sql.ErrNoRows {
		s.logger.Error("failed to count logs", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count logs")
	}

	// Convert to response format
	responseList := make([]LogResponse, len(logsList))
	for i, log := range logsList {
		responseList[i] = LogResponse{
			ID:           log.ID,
			UserID:       log.UserID,
			Action:       log.Action,
			Details:      log.Details,
			ServiceGroup: log.ServiceGroup,
			Level:        log.Level,
			CreatedAt:    log.CreatedAt,
		}
	}

	totalPages := (total + int64(req.PageSize) - 1) / int64(req.PageSize)

	response := GetLogsResponse{
		Logs:       responseList,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}

	s.logger.Info("logs retrieved",
		slog.Int("count", len(logsList)),
		slog.Int64("total", total),
		slog.String("service_group", req.ServiceGroup),
		slog.String("level", req.Level),
	)

	return c.JSON(http.StatusOK, response)
}

// GetLogStats returns statistics about logs
func (s *LogsService) GetLogStats(c echo.Context) error {
	daysStr := c.QueryParam("days")
	days := 7
	if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
		days = d
	}

	ctx := context.Background()
	since := time.Now().AddDate(0, 0, -days)

	// Get total logs
	total, err := s.db.Log.CountLogs(ctx, since)
	if err != nil {
		s.logger.Error("failed to count logs", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrieve stats")
	}

	// Get service groups
	serviceGroups, err := s.db.Log.GetDistinctServiceGroups(ctx, since)
	if err != nil && err != sql.ErrNoRows {
		s.logger.Error("failed to get service groups", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrieve stats")
	}

	// Get log levels
	levels, err := s.db.Log.GetDistinctLevels(ctx, since)
	if err != nil && err != sql.ErrNoRows {
		s.logger.Error("failed to get levels", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrieve stats")
	}

	response := map[string]interface{}{
		"total":          total,
		"days":           days,
		"service_groups": serviceGroups,
		"levels":         levels,
		"since":          since,
	}

	return c.JSON(http.StatusOK, response)
}

// ClearOldLogs removes logs older than the specified days (retention policy)
func (s *LogsService) ClearOldLogs(c echo.Context) error {
	// Parse days from query
	daysStr := c.QueryParam("days")
	days := 30 // Default 30 days retention
	if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
		days = d
	}

	ctx := context.Background()
	before := time.Now().AddDate(0, 0, -days)

	// Delete logs older than retention period
	err := s.db.Log.DeleteLogsOlderThan(ctx, before)
	if err != nil {
		s.logger.Error("failed to delete old logs", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete logs")
	}

	s.logger.Info("old logs deleted",
		slog.Int("days_retained", days),
	)

	response := map[string]interface{}{
		"retention_days": days,
		"before":         before,
		"message":        "Logs older than " + before.Format("2006-01-02") + " have been deleted",
	}

	return c.JSON(http.StatusOK, response)
}
