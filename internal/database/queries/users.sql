-- name: GetAllUsers :many
SELECT
  *
FROM
  users;

-- name: GetUserByID :many
SELECT
  *
FROM
  users
WHERE
  id = ?;

-- name: GetBySessionToken :one
SELECT
  u.*
FROM
  users u
  JOIN user_sessions us ON u.id = us.user_id
WHERE
  us.session_token = ?;

-- name: UpsertUser :one
INSERT INTO
  users (username, email, password, role)
VALUES
  (?, ?, ?, ?) ON CONFLICT (email) DO
UPDATE
SET
  username = excluded.username,
  password = excluded.password,
  role = excluded.role,
  updated_at = CURRENT_TIMESTAMP RETURNING *;

-- name: GetByEmailOrUsername :one
SELECT
  *
FROM
  users
WHERE
  email = ?
  OR username = ?;

-- name: DeleteUser :exec
DELETE FROM users
WHERE
  id = ?;
