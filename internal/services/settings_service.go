package services

import (
	"database/sql"
	"log/slog"
	"net/http"

	"visory/internal/database"
	dbnotifications "visory/internal/database/notifications"
	notifs "visory/internal/notifications"
	"visory/internal/utils"

	"github.com/labstack/echo/v4"
)

type SettingsService struct {
	db         *database.Service
	Dispatcher *utils.Dispatcher
	Logger     *slog.Logger
	Notifier   *notifs.Manager
}

// NewSettingsService creates a new SettingsService with dependency injection
func NewSettingsService(db *database.Service, dispatcher *utils.Dispatcher, logger *slog.Logger, notifier *notifs.Manager) *SettingsService {
	return &SettingsService{
		db:         db,
		Dispatcher: dispatcher.WithGroup("settings"),
		Logger:     logger.WithGroup("settings"),
		Notifier:   notifier,
	}
}

// @Summary      get all notification settings
// @Description  fetch all notification provider settings
// @Tags         settings
// @Produce      json
// @Success      200   {array}   notifications.NotificationSetting
// @Failure      401   {object}  models.HTTPError
// @Failure      403   {object}  models.HTTPError
// @Failure      500   {object}  models.HTTPError
// @Router       /settings/notifications [get]
func (s *SettingsService) GetAllNotificationSettings(c echo.Context) error {
	settings, err := s.db.Notification.GetAllNotificationSettings(c.Request().Context())
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to fetch notification settings", err)
	}

	return c.JSON(http.StatusOK, settings)
}

// @Summary      get notification setting by provider
// @Description  fetch notification setting for a specific provider
// @Tags         settings
// @Produce      json
// @Param        provider  path      string  true  "Provider name (e.g., discord, telegram)"
// @Success      200   {object}  notifications.NotificationSetting
// @Failure      401   {object}  models.HTTPError
// @Failure      403   {object}  models.HTTPError
// @Failure      404   {object}  models.HTTPError
// @Failure      500   {object}  models.HTTPError
// @Router       /settings/notifications/{provider} [get]
func (s *SettingsService) GetNotificationSettingByProvider(c echo.Context) error {
	provider := c.Param("provider")

	setting, err := s.db.Notification.GetNotificationSettingByProvider(c.Request().Context(), provider)
	if err != nil {
		if err == sql.ErrNoRows {
			return s.Dispatcher.NewNotFound("Notification setting not found", err)
		}
		return s.Dispatcher.NewInternalServerError("Failed to fetch notification setting", err)
	}

	return c.JSON(http.StatusOK, setting)
}

// NotificationSettingRequest represents the request body for upserting a notification setting
type NotificationSettingRequest struct {
	Provider      string  `json:"provider"`
	Enabled       bool    `json:"enabled"`
	WebhookURL    string  `json:"webhook_url"`
	NotifyOnError bool    `json:"notify_on_error"`
	NotifyOnWarn  bool    `json:"notify_on_warn"`
	NotifyOnInfo  bool    `json:"notify_on_info"`
	Config        *string `json:"config,omitempty"`
}

// @Summary      upsert notification setting
// @Description  create or update a notification provider setting
// @Tags         settings
// @Accept       json
// @Produce      json
// @Param        setting  body      NotificationSettingRequest  true  "Notification setting"
// @Success      200   {object}  notifications.NotificationSetting
// @Failure      400   {object}  models.HTTPError
// @Failure      401   {object}  models.HTTPError
// @Failure      403   {object}  models.HTTPError
// @Failure      500   {object}  models.HTTPError
// @Router       /settings/notifications [post]
func (s *SettingsService) UpsertNotificationSetting(c echo.Context) error {
	var req NotificationSettingRequest
	if err := c.Bind(&req); err != nil {
		return s.Dispatcher.NewBadRequest("Invalid request body", err)
	}

	if req.Provider == "" {
		return s.Dispatcher.NewBadRequest("Provider is required", nil)
	}

	// Convert to nullable pointers for sqlc
	enabled := req.Enabled
	notifyOnError := req.NotifyOnError
	notifyOnWarn := req.NotifyOnWarn
	notifyOnInfo := req.NotifyOnInfo
	webhookURL := req.WebhookURL

	setting, err := s.db.Notification.UpsertNotificationSetting(c.Request().Context(), dbnotifications.UpsertNotificationSettingParams{
		Provider:      req.Provider,
		Enabled:       &enabled,
		WebhookUrl:    &webhookURL,
		NotifyOnError: &notifyOnError,
		NotifyOnWarn:  &notifyOnWarn,
		NotifyOnInfo:  &notifyOnInfo,
		Config:        req.Config,
	})
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to save notification setting", err)
	}

	// Reload notification senders after settings change
	s.reloadNotificationSenders(c)

	return c.JSON(http.StatusOK, setting)
}

// @Summary      delete notification setting
// @Description  delete a notification provider setting
// @Tags         settings
// @Produce      json
// @Param        provider  path      string  true  "Provider name"
// @Success      204
// @Failure      401   {object}  models.HTTPError
// @Failure      403   {object}  models.HTTPError
// @Failure      500   {object}  models.HTTPError
// @Router       /settings/notifications/{provider} [delete]
func (s *SettingsService) DeleteNotificationSetting(c echo.Context) error {
	provider := c.Param("provider")

	err := s.db.Notification.DeleteNotificationSetting(c.Request().Context(), provider)
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to delete notification setting", err)
	}

	// Reload notification senders after settings change
	s.reloadNotificationSenders(c)

	return c.NoContent(http.StatusNoContent)
}

// @Summary      test notification
// @Description  send a test notification to verify the configuration
// @Tags         settings
// @Accept       json
// @Produce      json
// @Param        provider  path      string  true  "Provider name"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  models.HTTPError
// @Failure      401   {object}  models.HTTPError
// @Failure      403   {object}  models.HTTPError
// @Failure      500   {object}  models.HTTPError
// @Router       /settings/notifications/{provider}/test [post]
func (s *SettingsService) TestNotification(c echo.Context) error {
	provider := c.Param("provider")

	setting, err := s.db.Notification.GetNotificationSettingByProvider(c.Request().Context(), provider)
	if err != nil {
		if err == sql.ErrNoRows {
			return s.Dispatcher.NewNotFound("Notification setting not found", err)
		}
		return s.Dispatcher.NewInternalServerError("Failed to fetch notification setting", err)
	}

	if setting.WebhookUrl == nil || *setting.WebhookUrl == "" {
		return s.Dispatcher.NewBadRequest("Webhook URL is not configured", nil)
	}

	// Create a temporary sender to test
	switch provider {
	case "discord":
		sender := notifs.NewDiscordSender(notifs.DiscordConfig{
			WebhookURL:    *setting.WebhookUrl,
			NotifyOnError: true,
			NotifyOnWarn:  true,
			NotifyOnInfo:  true,
		})

		err := sender.Send(notifs.Notification{
			Level:   notifs.LevelInfo,
			Title:   "Test Notification",
			Message: "This is a test notification from Visory to verify your webhook configuration.",
			Fields: map[string]string{
				"Provider": provider,
				"Status":   "Testing",
			},
			Group:   "settings",
			Version: "test",
		})

		if err != nil {
			return s.Dispatcher.NewInternalServerError("Failed to send test notification", err)
		}
	default:
		return s.Dispatcher.NewBadRequest("Unsupported provider", nil)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Test notification sent successfully",
	})
}

// reloadNotificationSenders reloads the notification senders from the database
func (s *SettingsService) reloadNotificationSenders(c echo.Context) {
	if s.Notifier == nil {
		return
	}

	settings, err := s.db.Notification.GetEnabledNotificationSettings(c.Request().Context())
	if err != nil {
		s.Logger.Error("Failed to reload notification settings", "error", err)
		return
	}

	// Clear existing senders and re-register from DB
	s.Notifier.ClearSenders()

	for _, setting := range settings {
		if setting.WebhookUrl == nil || *setting.WebhookUrl == "" {
			continue
		}

		switch setting.Provider {
		case "discord":
			notifyOnError := setting.NotifyOnError != nil && *setting.NotifyOnError
			notifyOnWarn := setting.NotifyOnWarn != nil && *setting.NotifyOnWarn
			notifyOnInfo := setting.NotifyOnInfo != nil && *setting.NotifyOnInfo

			sender := notifs.NewDiscordSender(notifs.DiscordConfig{
				WebhookURL:    *setting.WebhookUrl,
				NotifyOnError: notifyOnError,
				NotifyOnWarn:  notifyOnWarn,
				NotifyOnInfo:  notifyOnInfo,
			})
			s.Notifier.RegisterSender(sender)
		}
	}
}
