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

-- name: GetErrorRateByService :many
SELECT 
  service_group,
  COUNT(CASE WHEN level = 'ERROR' THEN 1 END) as error_count,
  COUNT(*) as total_count,
  ROUND(100.0 * COUNT(CASE WHEN level = 'ERROR' THEN 1 END) / COUNT(*), 2) as error_rate
FROM logs
WHERE created_at >= ?
GROUP BY service_group
ORDER BY error_rate DESC;

-- name: GetAverageLogCountByHour :many
SELECT 
  strftime('%Y-%m-%d %H:00:00', created_at) as hour,
  COUNT(*) as log_count
FROM logs
WHERE created_at >= ?
GROUP BY hour
ORDER BY hour DESC
LIMIT 24;

-- name: GetLogLevelDistribution :many
SELECT 
  level,
  COUNT(*) as count,
  ROUND(100.0 * COUNT(*) / (SELECT COUNT(*) FROM logs WHERE logs.created_at >= ?), 2) as percentage
FROM logs
WHERE logs.created_at >= ?
GROUP BY level
ORDER BY count DESC;

-- name: GetServiceGroupDistribution :many
SELECT 
  service_group,
  COUNT(*) as count,
  ROUND(100.0 * COUNT(*) / (SELECT COUNT(*) FROM logs WHERE logs.created_at >= ?), 2) as percentage
FROM logs
WHERE logs.created_at >= ?
GROUP BY service_group
ORDER BY count DESC;
