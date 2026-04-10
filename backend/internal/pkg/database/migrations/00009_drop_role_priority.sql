-- +goose Up
ALTER TABLE roles DROP COLUMN IF EXISTS priority;

-- +goose Down
ALTER TABLE roles ADD COLUMN IF NOT EXISTS priority integer NOT NULL DEFAULT 0;
