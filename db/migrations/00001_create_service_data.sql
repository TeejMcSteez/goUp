-- +goose Up
CREATE TABLE service_data (
	"id" INTEGER PRIMARY KEY AUTOINCREMENT,
	"service_url" TEXT,
	"service_name" TEXT,
	"service_description" TEXT,
	"service_http_response" TEXT,
	"service_api_response" TEXT,
	"service_response_time" INTEGER,
	"timestamp" TEXT,
	"error" INTEGER NOT NULL DEFAULT 0
);

-- Superseded by idx_service_name_id / idx_error_timestamp below; dropping it
-- here is a no-op on a fresh database and cleans up any database that still
-- has it from before those two indexes existed.
DROP INDEX IF EXISTS idx_service_data_lookup;

CREATE INDEX idx_service_name_id
	ON service_data(service_name, id);

CREATE INDEX idx_error_timestamp
	ON service_data(error, timestamp);

-- +goose Down
DROP TABLE service_data;
