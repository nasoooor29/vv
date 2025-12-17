package services

import (
	"context"
	"log/slog"
	"time"

	"visory/internal/database"
)

// LogRetentionService handles automatic log cleanup based on retention policy
type LogRetentionService struct {
	db     *database.Service
	logger *slog.Logger
	ticker *time.Ticker
	done   chan bool
}

// NewLogRetentionService creates a new log retention service
func NewLogRetentionService(db *database.Service, logger *slog.Logger) *LogRetentionService {
	return &LogRetentionService{
		db:     db,
		logger: logger.WithGroup("retention"),
		done:   make(chan bool),
	}
}

// Start begins the automatic log cleanup routine
// It runs daily at the specified time (default: 2 AM UTC)
func (s *LogRetentionService) Start(retentionDays int, cleanupHour int) {
	go func() {
		s.logger.Info("log retention service started",
			slog.Int("retention_days", retentionDays),
			slog.Int("cleanup_hour", cleanupHour),
		)

		// Calculate time until next cleanup
		now := time.Now().UTC()
		nextCleanup := time.Date(
			now.Year(), now.Month(), now.Day(),
			cleanupHour, 0, 0, 0,
			time.UTC,
		)

		// If the scheduled time has already passed today, schedule for tomorrow
		if nextCleanup.Before(now) {
			nextCleanup = nextCleanup.Add(24 * time.Hour)
		}

		// Initial sleep until first cleanup
		s.logger.Info("next log cleanup scheduled",
			slog.Time("at", nextCleanup),
		)

		s.ticker = time.NewTicker(24 * time.Hour)

		// Sleep until next cleanup
		time.Sleep(time.Until(nextCleanup))

		// Perform first cleanup
		s.cleanup(retentionDays)

		// Then run daily
		for {
			select {
			case <-s.ticker.C:
				s.cleanup(retentionDays)
			case <-s.done:
				return
			}
		}
	}()
}

// Stop stops the automatic log cleanup routine
func (s *LogRetentionService) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.done <- true
	s.logger.Info("log retention service stopped")
}

// cleanup performs the actual log deletion
func (s *LogRetentionService) cleanup(retentionDays int) {
	ctx := context.Background()
	before := time.Now().AddDate(0, 0, -retentionDays)

	s.logger.Info("starting scheduled log cleanup",
		slog.Int("retention_days", retentionDays),
		slog.Time("before", before),
	)

	// Delete logs older than retention period
	err := s.db.Log.DeleteLogsOlderThan(ctx, before)
	if err != nil {
		s.logger.Error("failed to delete old logs",
			slog.String("error", err.Error()),
			slog.Int("retention_days", retentionDays),
		)
		return
	}

	s.logger.Info("scheduled log cleanup completed successfully",
		slog.Int("retention_days", retentionDays),
		slog.Time("deleted_before", before),
	)
}

// Manual cleanup method - can be called anytime
func (s *LogRetentionService) CleanupNow(retentionDays int) error {
	ctx := context.Background()
	before := time.Now().AddDate(0, 0, -retentionDays)

	s.logger.Info("manual log cleanup requested",
		slog.Int("retention_days", retentionDays),
	)

	err := s.db.Log.DeleteLogsOlderThan(ctx, before)
	if err != nil {
		s.logger.Error("manual log cleanup failed", slog.String("error", err.Error()))
		return err
	}

	s.logger.Info("manual log cleanup completed successfully")
	return nil
}
