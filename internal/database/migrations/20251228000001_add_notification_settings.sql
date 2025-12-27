-- +goose Up
-- +goose StatementBegin
CREATE TABLE notification_settings (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  provider TEXT NOT NULL UNIQUE,
  enabled BOOLEAN DEFAULT TRUE,
  webhook_url TEXT,
  notify_on_error BOOLEAN DEFAULT TRUE,
  notify_on_warn BOOLEAN DEFAULT FALSE,
  notify_on_info BOOLEAN DEFAULT FALSE,
  config TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE UNIQUE INDEX idx_notification_settings_provider ON notification_settings (provider);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_notification_settings_provider;
DROP TABLE IF EXISTS notification_settings;
-- +goose StatementEnd
