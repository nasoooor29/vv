-- +goose Up
-- +goose StatementBegin
ALTER TABLE logs ADD COLUMN service_group TEXT NOT NULL DEFAULT 'general';
ALTER TABLE logs ADD COLUMN level TEXT NOT NULL DEFAULT 'info';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE logs DROP COLUMN service_group;
ALTER TABLE logs DROP COLUMN level;
-- +goose StatementEnd
