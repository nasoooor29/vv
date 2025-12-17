-- name: GetAllLogs :many
SELECT
  *
FROM
  logs;

-- name: GetLogByID :many
SELECT
  *
FROM
  logs
WHERE
  id = ?;

-- name: CreateLog :one
INSERT INTO logs (user_id, "action", details, service_group, level)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetLogsByServiceGroup :many
SELECT
  *
FROM
  logs
WHERE
  service_group = ?
ORDER BY
  created_at DESC;

-- name: GetLogsByLevel :many
SELECT
  *
FROM
  logs
WHERE
  level = ?
ORDER BY
  created_at DESC;

-- name: GetLogsByServiceGroupAndLevel :many
SELECT
  *
FROM
  logs
WHERE
  service_group = ? AND level = ?
ORDER BY
  created_at DESC;
