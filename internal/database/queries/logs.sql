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
