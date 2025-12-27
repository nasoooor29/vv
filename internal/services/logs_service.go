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
	"visory/internal/models"
	"visory/internal/utils"

	"github.com/labstack/echo/v4"
)

type LogsService struct {
	db         *database.Service
	Dispatcher *utils.Dispatcher
	Logger     *slog.Logger
}

// NewLogsService creates a new LogsService with dependency injection
func NewLogsService(db *database.Service, dispatcher *utils.Dispatcher, logger *slog.Logger) *LogsService {
	return &LogsService{
		db:         db,
		Dispatcher: dispatcher.WithGroup("logs"),
		Logger:     logger.WithGroup("logs"),
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

//	@Summary      get logs
//	@Description  retrieve logs with filtering and pagination
//	@Tags         logs
//	@Produce      json
//	@Param        service_group  query    string  false  "Filter by service group"
//	@Param        level          query    string  false  "Filter by log level"
//	@Param        page           query    int     false  "Page number (default 1)"
//	@Param        page_size      query    int     false  "Page size (default 20, max 100)"
//	@Param        days           query    int     false  "Number of days to filter (default 7)"
//	@Success      200            {object}  models.GetLogsResponse
//	@Failure      400            {object}  models.HTTPError
//	@Failure      401            {object}  models.HTTPError
//	@Failure      403            {object}  models.HTTPError
//	@Failure      500            {object}  models.HTTPError
//	@Router       /logs [get]
//
// GetLogs retrieves logs with filtering and pagination
func (s *LogsService) GetLogs(c echo.Context) error {
	req := new(GetLogsRequest)
	if err := c.Bind(req); err != nil {
		// s.dispatcher.Error("failed to parse query params", "error", err)
		return s.Dispatcher.NewBadRequest("invalid query parameters", err)
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
			return s.Dispatcher.NewInternalServerError("failed to retrieve logs", err)
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
			return s.Dispatcher.NewInternalServerError("failed to retrieve logs", err)
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
			return s.Dispatcher.NewInternalServerError("failed to retrieve logs", err)
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
			return s.Dispatcher.NewInternalServerError("failed to retrieve logs", err)
		}
		total, err = s.db.Log.CountLogs(ctx, since)
	}

	if err != nil && err != sql.ErrNoRows {
		return s.Dispatcher.NewInternalServerError("failed to count logs", err)
	}

	// Convert to response format
	responseList := make([]models.LogResponse, len(logsList))
	for i, log := range logsList {
		responseList[i] = models.LogResponse{
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

	response := models.GetLogsResponse{
		Logs:       responseList,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}

	// s.dispatcher.Info("logs retrieved",
	// 	slog.Int("count", len(logsList)),
	// 	slog.Int64("total", total),
	// 	slog.String("service_group", req.ServiceGroup),
	// 	slog.String("level", req.Level),
	// )

	return c.JSON(http.StatusOK, response)
}

//	@Summary      get log statistics
//	@Description  retrieve statistics about logs
//	@Tags         logs
//	@Produce      json
//	@Param        days  query    int     false  "Number of days to filter (default 7)"
//	@Success      200   {object}  models.LogStatsResponse
//	@Failure      401   {object}  models.HTTPError
//	@Failure      403   {object}  models.HTTPError
//	@Failure      500   {object}  models.HTTPError
//	@Router       /logs/stats [get]
//
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
		return s.Dispatcher.NewInternalServerError("failed to retrieve stats", err)
	}

	// Get service groups
	serviceGroups, err := s.db.Log.GetDistinctServiceGroups(ctx, since)
	if err != nil && err != sql.ErrNoRows {
		return s.Dispatcher.NewInternalServerError("failed to retrieve stats", err)
	}

	// Get log levels
	levels, err := s.db.Log.GetDistinctLevels(ctx, since)
	if err != nil && err != sql.ErrNoRows {
		return s.Dispatcher.NewInternalServerError("failed to retrieve stats", err)
	}

	response := models.LogStatsResponse{
		Total:         total,
		Days:          days,
		ServiceGroups: serviceGroups,
		Levels:        levels,
		Since:         since,
	}

	return c.JSON(http.StatusOK, response)
}

//	@Summary      clear old logs
//	@Description  delete logs older than specified retention period
//	@Tags         logs
//	@Produce      json
//	@Param        days  query    int     false  "Retention days (default 30)"
//	@Success      200   {object}  models.ClearOldLogsResponse
//	@Failure      401   {object}  models.HTTPError
//	@Failure      403   {object}  models.HTTPError
//	@Failure      500   {object}  models.HTTPError
//	@Router       /logs/cleanup [delete]
//
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
		return s.Dispatcher.NewInternalServerError("failed to delete logs", err)
	}

	// s.dispatcher.Info("old logs deleted",
	// 	slog.Int("days_retained", days),
	// )

	response := models.ClearOldLogsResponse{
		RetentionDays: days,
		Before:        before,
		Message:       "Logs older than " + before.Format("2006-01-02") + " have been deleted",
	}

	return c.JSON(http.StatusOK, response)
}
