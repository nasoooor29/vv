-- CREATE TABLE IF NOT EXISTS user_sessions (
--   id INTEGER PRIMARY KEY AUTOINCREMENT,
--   user_id INTEGER NOT NULL,
--   session_token TEXT NOT NULL UNIQUE,
--   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
--   updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
--   FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
-- );
-- name: GetAllSessions :many
SELECT
  *
FROM
  user_sessions;

-- name: GetSessionByID :one
SELECT
  *
FROM
  user_sessions
WHERE
  id = ?;

-- name: UpsertSession :one
INSERT INTO
  user_sessions (user_id, session_token)
VALUES
  (?, ?) ON CONFLICT (session_token) DO
UPDATE
SET
  user_id = excluded.user_id,
  updated_at = CURRENT_TIMESTAMP RETURNING *;

-- name: DeleteBySessionToken :exec
DELETE FROM user_sessions
WHERE
  session_token = ?;
