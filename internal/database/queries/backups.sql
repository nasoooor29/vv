-- name: CreateBackupJob :one
INSERT INTO backup_jobs (
    name, type, target_type, target_id, target_name, client_id,
    status, progress, destination, size_bytes, error_message,
    started_at, completed_at, created_by
) VALUES (
    ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?,
    ?, ?, ?
) RETURNING *;

-- name: GetBackupJob :one
SELECT * FROM backup_jobs WHERE id = ?;

-- name: ListBackupJobs :many
SELECT * FROM backup_jobs
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListBackupJobsByTarget :many
SELECT * FROM backup_jobs
WHERE target_type = ? AND target_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListBackupJobsByType :many
SELECT * FROM backup_jobs
WHERE type = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateBackupJobStatus :one
UPDATE backup_jobs
SET status = ?, progress = ?, error_message = ?,
    started_at = COALESCE(started_at, ?),
    completed_at = ?
WHERE id = ?
RETURNING *;

-- name: UpdateBackupJobProgress :one
UPDATE backup_jobs
SET progress = ?
WHERE id = ?
RETURNING *;

-- name: UpdateBackupJobCompleted :one
UPDATE backup_jobs
SET status = 'completed', progress = 100, completed_at = ?, size_bytes = ?
WHERE id = ?
RETURNING *;

-- name: UpdateBackupJobFailed :one
UPDATE backup_jobs
SET status = 'failed', error_message = ?, completed_at = ?
WHERE id = ?
RETURNING *;

-- name: DeleteBackupJob :exec
DELETE FROM backup_jobs WHERE id = ?;

-- name: CountBackupJobs :one
SELECT COUNT(*) FROM backup_jobs;

-- name: CountBackupJobsByType :one
SELECT COUNT(*) FROM backup_jobs WHERE type = ?;

-- name: GetBackupStats :one
SELECT
    COUNT(*) as total_backups,
    SUM(CASE WHEN type LIKE 'vm_%' THEN 1 ELSE 0 END) as vm_snapshots,
    SUM(CASE WHEN type LIKE 'container_%' THEN 1 ELSE 0 END) as container_backups,
    SUM(CASE WHEN type = 'firewall' THEN 1 ELSE 0 END) as firewall_backups,
    COALESCE(SUM(size_bytes), 0) as total_size_bytes,
    MAX(created_at) as last_backup_at
FROM backup_jobs
WHERE status = 'completed';

-- name: CreateBackupSchedule :one
INSERT INTO backup_schedules (
    name, type, target_type, target_id, target_name, client_id,
    schedule, schedule_time, destination, retention_count, enabled,
    next_run_at, created_by
) VALUES (
    ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?,
    ?, ?
) RETURNING *;

-- name: GetBackupSchedule :one
SELECT * FROM backup_schedules WHERE id = ?;

-- name: ListBackupSchedules :many
SELECT * FROM backup_schedules
ORDER BY created_at DESC;

-- name: ListEnabledSchedules :many
SELECT * FROM backup_schedules
WHERE enabled = 1
ORDER BY next_run_at ASC;

-- name: ListSchedulesDueForRun :many
SELECT * FROM backup_schedules
WHERE enabled = 1 AND next_run_at <= ?
ORDER BY next_run_at ASC;

-- name: UpdateBackupSchedule :one
UPDATE backup_schedules
SET name = COALESCE(?, name),
    schedule = COALESCE(?, schedule),
    schedule_time = COALESCE(?, schedule_time),
    retention_count = COALESCE(?, retention_count),
    enabled = COALESCE(?, enabled),
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: UpdateScheduleAfterRun :one
UPDATE backup_schedules
SET last_run_at = ?, next_run_at = ?, last_status = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: EnableBackupSchedule :one
UPDATE backup_schedules
SET enabled = 1, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DisableBackupSchedule :one
UPDATE backup_schedules
SET enabled = 0, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteBackupSchedule :exec
DELETE FROM backup_schedules WHERE id = ?;

-- name: CountActiveSchedules :one
SELECT COUNT(*) FROM backup_schedules WHERE enabled = 1;
