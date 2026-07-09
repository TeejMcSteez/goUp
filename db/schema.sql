CREATE TABLE IF NOT EXISTS service_data (
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

CREATE INDEX IF NOT EXISTS idx_service_name_id
	ON service_data(service_name, id);

CREATE INDEX IF NOT EXISTS idx_error_timestamp
	ON service_data(error, timestamp);
