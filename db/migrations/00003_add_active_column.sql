-- +goose Up
ALTER TABLE service_data ADD COLUMN active INTEGER NOT NULL DEFAULT 1;

-- +goose Down
ALTER TABLE service_data DROP COLUMN active;
