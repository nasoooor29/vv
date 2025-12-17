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
  service_group = ? AND created_at >= ?
ORDER BY
  created_at DESC
LIMIT ? OFFSET ?;

-- name: GetLogsByLevel :many
SELECT
  *
FROM
  logs
WHERE
  level = ? AND created_at >= ?
ORDER BY
  created_at DESC
LIMIT ? OFFSET ?;

-- name: GetLogsByServiceGroupAndLevel :many
SELECT
  *
FROM
  logs
WHERE
  service_group = ? AND level = ? AND created_at >= ?
ORDER BY
  created_at DESC
LIMIT ? OFFSET ?;

-- name: GetLogsPaginated :many
SELECT
  *
FROM
  logs
WHERE
  created_at >= ?
ORDER BY
  created_at DESC
LIMIT ? OFFSET ?;

-- name: CountLogs :one
SELECT COUNT(*) as count
FROM logs
WHERE created_at >= ?;

-- name: CountLogsByServiceGroup :one
SELECT COUNT(*) as count
FROM logs
WHERE service_group = ? AND created_at >= ?;

-- name: CountLogsByLevel :one
SELECT COUNT(*) as count
FROM logs
WHERE level = ? AND created_at >= ?;

-- name: CountLogsByServiceGroupAndLevel :one
SELECT COUNT(*) as count
FROM logs
WHERE service_group = ? AND level = ? AND created_at >= ?;

-- name: GetDistinctServiceGroups :many
SELECT DISTINCT service_group
FROM logs
WHERE created_at >= ?
ORDER BY service_group;

-- name: GetDistinctLevels :many
SELECT DISTINCT level
FROM logs
WHERE created_at >= ?
ORDER BY level;

-- name: DeleteLogsOlderThan :exec
DELETE FROM logs
WHERE created_at < ?;
