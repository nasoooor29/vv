-- name: GetAllNotifications :many
SELECT
  *
FROM
  notifications;

-- name: GetNotificationByID :one
SELECT
  *
FROM
  notifications
WHERE
  id = ?;

-- Notification Settings Queries

-- name: GetAllNotificationSettings :many
SELECT
  *
FROM
  notification_settings;

-- name: GetNotificationSettingByProvider :one
SELECT
  *
FROM
  notification_settings
WHERE
  provider = ?;

-- name: UpsertNotificationSetting :one
INSERT INTO notification_settings (
  provider,
  enabled,
  webhook_url,
  notify_on_error,
  notify_on_warn,
  notify_on_info,
  config,
  updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(provider) DO UPDATE SET
  enabled = excluded.enabled,
  webhook_url = excluded.webhook_url,
  notify_on_error = excluded.notify_on_error,
  notify_on_warn = excluded.notify_on_warn,
  notify_on_info = excluded.notify_on_info,
  config = excluded.config,
  updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: DeleteNotificationSetting :exec
DELETE FROM notification_settings WHERE provider = ?;

-- name: GetEnabledNotificationSettings :many
SELECT
  *
FROM
  notification_settings
WHERE
  enabled = TRUE;
