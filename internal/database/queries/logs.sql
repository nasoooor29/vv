-- name: GetAll :many
SELECT
  *
FROM
  logs;

-- name: ByID :many
SELECT
  *
FROM
  logs
WHERE
  id = ?;
