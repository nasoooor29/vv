-- +goose Up
-- +goose StatementBegin

-- Backup jobs table (tracks individual backup/restore operations)
CREATE TABLE IF NOT EXISTS backup_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL,           -- 'vm_snapshot', 'vm_export', 'container_export', 'container_commit', 'firewall'
    target_type TEXT NOT NULL,    -- 'vm', 'container', 'firewall'
    target_id TEXT NOT NULL,      -- VM UUID, container ID, or 'firewall'
    target_name TEXT,
    client_id TEXT,               -- Docker client ID (null for VMs and firewall)
    status TEXT NOT NULL DEFAULT 'pending',  -- 'pending', 'running', 'completed', 'failed'
    progress INTEGER DEFAULT 0,
    destination TEXT,             -- Backup file path or snapshot name
    size_bytes INTEGER,
    error_message TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Backup schedules table (for automated daily backups)
CREATE TABLE IF NOT EXISTS backup_schedules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL,           -- 'vm_snapshot', 'container_export', etc.
    target_type TEXT NOT NULL,    -- 'vm', 'container'
    target_id TEXT NOT NULL,
    target_name TEXT,
    client_id TEXT,
    schedule TEXT NOT NULL,       -- 'daily', 'weekly', 'hourly'
    schedule_time TEXT,           -- Time of day e.g., '02:00' for daily
    destination TEXT NOT NULL,
    retention_count INTEGER DEFAULT 7,  -- Keep last N backups
    enabled INTEGER DEFAULT 1,
    last_run_at TIMESTAMP,
    next_run_at TIMESTAMP,
    last_status TEXT,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for faster queries
CREATE INDEX IF NOT EXISTS idx_backup_jobs_target ON backup_jobs(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_backup_jobs_status ON backup_jobs(status);
CREATE INDEX IF NOT EXISTS idx_backup_jobs_created_at ON backup_jobs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_backup_schedules_enabled ON backup_schedules(enabled);
CREATE INDEX IF NOT EXISTS idx_backup_schedules_next_run ON backup_schedules(next_run_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_backup_jobs_target;
DROP INDEX IF EXISTS idx_backup_jobs_status;
DROP INDEX IF EXISTS idx_backup_jobs_created_at;
DROP INDEX IF EXISTS idx_backup_schedules_enabled;
DROP INDEX IF EXISTS idx_backup_schedules_next_run;
DROP TABLE IF EXISTS backup_schedules;
DROP TABLE IF EXISTS backup_jobs;

-- +goose StatementEnd
