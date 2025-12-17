package services

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"visory/internal/database"
	"visory/internal/database/logs"
	"visory/internal/models"
	"visory/internal/utils"

	"github.com/labstack/echo/v4"
)

type MetricsService struct {
	db         *database.Service
	dispatcher *utils.Dispatcher
}

// NewMetricsService creates a new MetricsService with dependency injection
func NewMetricsService(db *database.Service, dispatcher *utils.Dispatcher) *MetricsService {
	return &MetricsService{
		db:         db,
		dispatcher: dispatcher.WithGroup("metrics"),
	}
}

// GetMetricsRequest represents metrics query parameters
type GetMetricsRequest struct {
	Days int `query:"days"`
}

// GetMetrics retrieves comprehensive performance metrics
func (s *MetricsService) GetMetrics(c echo.Context) error {
	req := new(GetMetricsRequest)
	if err := c.Bind(req); err != nil {
		return s.dispatcher.NewBadRequest("invalid query parameters", err)
	}

	// Set defaults
	if req.Days < 1 {
		req.Days = 7
	}

	ctx := context.Background()
	since := time.Now().AddDate(0, 0, -req.Days)

	response := models.MetricsResponse{
		Period: models.MetricsPeriod{
			Days:  req.Days,
			Since: since,
			Until: time.Now(),
		},
	}

	// Get error rate by service
	errorRates, err := s.db.Log.GetErrorRateByService(ctx, since)
	if err != nil && err != sql.ErrNoRows {
		return s.dispatcher.NewInternalServerError("failed to retrieve metrics", err)
	}
	if errorRates != nil {
		for _, rate := range errorRates {
			response.ErrorRateByService = append(response.ErrorRateByService, models.ErrorRateByService{
				ServiceGroup: rate.ServiceGroup,
				ErrorCount:   rate.ErrorCount,
				TotalCount:   rate.TotalCount,
				ErrorRate:    rate.ErrorRate,
			})
		}
	}

	// Get hourly log counts
	hourlyLogs, err := s.db.Log.GetAverageLogCountByHour(ctx, since)
	if err != nil && err != sql.ErrNoRows {
		return s.dispatcher.NewInternalServerError("failed to retrieve metrics", err)
	}
	if hourlyLogs != nil {
		for _, log := range hourlyLogs {
			response.LogCountByHour = append(response.LogCountByHour, models.LogCountByHour{
				Hour:     log.Hour.(string),
				LogCount: log.LogCount,
			})
		}
	}

	// Get log level distribution
	levelDist, err := s.db.Log.GetLogLevelDistribution(ctx, logs.GetLogLevelDistributionParams{
		CreatedAt:   since,
		CreatedAt_2: since,
	})
	if err != nil && err != sql.ErrNoRows {
		return s.dispatcher.NewInternalServerError("failed to retrieve metrics", err)
	}
	if levelDist != nil {
		for _, dist := range levelDist {
			response.LogLevelDistribution = append(response.LogLevelDistribution, models.LogLevelStats{
				Level:      dist.Level,
				Count:      dist.Count,
				Percentage: dist.Percentage,
			})
		}
	}

	// Get service group distribution
	serviceDist, err := s.db.Log.GetServiceGroupDistribution(ctx, logs.GetServiceGroupDistributionParams{
		CreatedAt:   since,
		CreatedAt_2: since,
	})
	if err != nil && err != sql.ErrNoRows {
		return s.dispatcher.NewInternalServerError("failed to retrieve metrics", err)
	}
	if serviceDist != nil {
		for _, dist := range serviceDist {
			response.ServiceGroupDistribution = append(response.ServiceGroupDistribution, models.ServiceStats{
				ServiceGroup: dist.ServiceGroup,
				Count:        dist.Count,
				Percentage:   dist.Percentage,
			})
		}
	}

	// NOTE: log
	// s.dispatcher.Info("metrics retrieved",
	// 	slog.Int("days", req.Days),
	// 	slog.Int("error_rate_entries", len(response.ErrorRateByService)),
	// 	slog.Int("hourly_entries", len(response.LogCountByHour)),
	// )

	return c.JSON(http.StatusOK, response)
}

// GetServiceMetrics retrieves metrics for a specific service
func (s *MetricsService) GetServiceMetrics(c echo.Context) error {
	serviceGroup := c.Param("service")
	days := 7
	if d, err := strconv.Atoi(c.QueryParam("days")); err == nil && d > 0 {
		days = d
	}

	ctx := context.Background()
	since := time.Now().AddDate(0, 0, -days)

	// Count logs by level for this service
	levelDist, err := s.db.Log.GetLogLevelDistribution(ctx, logs.GetLogLevelDistributionParams{
		CreatedAt:   since,
		CreatedAt_2: since,
	})
	if err != nil && err != sql.ErrNoRows {
		return s.dispatcher.NewInternalServerError("failed to retrieve metrics", err, "service", serviceGroup)
	}

	// Count total logs for this service
	total, err := s.db.Log.CountLogsByServiceGroup(ctx, logs.CountLogsByServiceGroupParams{
		ServiceGroup: serviceGroup,
		CreatedAt:    since,
	})
	if err != nil && err != sql.ErrNoRows {
		return s.dispatcher.NewInternalServerError("failed to retrieve metrics", err, "service", serviceGroup)
	}

	// Count errors for this service
	errorCount, err := s.db.Log.CountLogsByServiceGroupAndLevel(ctx, logs.CountLogsByServiceGroupAndLevelParams{
		ServiceGroup: serviceGroup,
		Level:        "ERROR",
		CreatedAt:    since,
	})
	if err != nil && err != sql.ErrNoRows {
		return s.dispatcher.NewInternalServerError("failed to retrieve metrics", err, "service", serviceGroup)
	}

	errorRate := 0.0
	if total > 0 {
		errorRate = float64(errorCount) / float64(total) * 100
	}

	// Convert level distribution
	levelDistribution := []models.LogLevelStats{}
	if levelDist != nil {
		for _, dist := range levelDist {
			levelDistribution = append(levelDistribution, models.LogLevelStats{
				Level:      dist.Level,
				Count:      dist.Count,
				Percentage: dist.Percentage,
			})
		}
	}

	response := models.ServiceMetricsResponse{
		ServiceGroup:      serviceGroup,
		Days:              days,
		Since:             since,
		TotalLogs:         total,
		ErrorCount:        errorCount,
		ErrorRate:         errorRate,
		LevelDistribution: levelDistribution,
	}

	// NOTE: log
	// s.dispatcher.Info("service metrics retrieved",
	// 	slog.String("service", serviceGroup),
	// 	slog.Int("days", days),
	// 	slog.Int64("total", total),
	// 	slog.Float64("error_rate", errorRate),
	// )

	return c.JSON(http.StatusOK, response)
}

// GetHealthMetrics returns system health based on error rates
func (s *MetricsService) GetHealthMetrics(c echo.Context) error {
	ctx := context.Background()
	since := time.Now().Add(-1 * time.Hour) // Last hour

	errorRates, err := s.db.Log.GetErrorRateByService(ctx, since)
	if err != nil && err != sql.ErrNoRows {
		return s.dispatcher.NewInternalServerError("failed to retrieve health metrics", err)
	}

	health := models.HealthMetricsResponse{
		Timestamp:     time.Now(),
		Period:        "last_hour",
		Services:      []models.ServiceHealth{},
		OverallStatus: "healthy",
		Alerts:        []string{},
	}

	alerts := []string{}
	maxErrorRate := 0.0

	if errorRates != nil {
		for _, rate := range errorRates {
			service := models.ServiceHealth{
				ServiceGroup: rate.ServiceGroup,
				ErrorRate:    rate.ErrorRate,
				ErrorCount:   rate.ErrorCount,
				TotalCount:   rate.TotalCount,
				Status:       "healthy",
			}

			// Mark as warning if error rate > 5%
			if rate.ErrorRate > 5 {
				service.Status = "warning"
				alerts = append(alerts, rate.ServiceGroup+" has error rate of "+string(rune(int(rate.ErrorRate)))+"%")
			}

			// Mark as critical if error rate > 10%
			if rate.ErrorRate > 10 {
				service.Status = "critical"
				alerts = append(alerts, rate.ServiceGroup+" has CRITICAL error rate of "+string(rune(int(rate.ErrorRate)))+"%")
			}

			if rate.ErrorRate > maxErrorRate {
				maxErrorRate = rate.ErrorRate
			}

			health.Services = append(health.Services, service)
		}
	}

	health.Alerts = alerts

	// Determine overall status
	if maxErrorRate > 10 {
		health.OverallStatus = "critical"
	} else if maxErrorRate > 5 {
		health.OverallStatus = "warning"
	}

	// NOTE: log
	// s.dispatcher.Info("health metrics retrieved",
	// 	slog.String("overall_status", health.OverallStatus),
	// 	slog.Int("alerts", len(alerts)),
	// )

	return c.JSON(http.StatusOK, health)
}
