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
-- name: InsertTlsStatus :exec
INSERT INTO tls_status (service_name, fingerprint, not_after, subject, issuer,
                        is_expired, chain, first_seen, last_checked)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(service_name) DO UPDATE SET
    fingerprint  = excluded.fingerprint,
    not_after    = excluded.not_after,
    subject      = excluded.subject,
    issuer       = excluded.issuer,
    is_expired   = excluded.is_expired,
    chain        = excluded.chain,
    last_checked = excluded.last_checked,
    first_seen   = CASE
        WHEN tls_status.fingerprint = excluded.fingerprint
        THEN tls_status.first_seen   -- same cert, keep it
        ELSE excluded.first_seen     -- renewed, resets, giving cert age for free
    END;
-- name: GetServiceTlsStatus :one
SELECT * FROM tls_status WHERE service_name = ? LIMIT 1;
-- name: GetTlsStatus :many
SELECT * FROM tls_status;
-- name: DeleteServiceTlsStatus :exec
DELETE FROM tls_status WHERE service_name = ?;
