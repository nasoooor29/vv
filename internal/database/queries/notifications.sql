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
