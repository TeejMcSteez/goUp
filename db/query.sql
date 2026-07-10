-- name: InsertData :exec
INSERT INTO service_data (service_url, service_name, service_description, service_HTTP_response, service_API_response, service_response_time, timestamp, error, active) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
-- name: GetAllData :many
SELECT * FROM service_data;
-- name: GetRecentData :many
SELECT * FROM service_data WHERE id IN (
	SELECT MAX(id) FROM service_data GROUP BY service_name
) ORDER BY id DESC;
-- name: GetDataForService :many
SELECT * FROM service_data WHERE service_name = ? ORDER BY id DESC;
-- name: GetErrorDataAsc :many
SELECT * FROM service_data WHERE error = 1 ORDER BY timestamp ASC LIMIT ?;
-- name: GetErrorDataDesc :many
SELECT * FROM service_data WHERE error = 1 ORDER BY timestamp DESC LIMIT ?;
-- name: ClearServiceData :exec
DELETE FROM service_data;
-- name: DeleteService :exec
DELETE FROM service_data WHERE service_name = ?;
-- name: ServiceRename :exec
UPDATE service_data SET service_name = ? WHERE service_name = ?;
-- name: Cleanup :exec
DELETE FROM service_data WHERE service_name NOT IN (sqlc.slice('names'));
