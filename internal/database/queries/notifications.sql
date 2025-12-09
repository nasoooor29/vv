-- name: GetAll :many
SELECT
  *
FROM
  notifications;

-- name: ByID :one
SELECT
  *
FROM
  notifications
WHERE
  id = ?;
