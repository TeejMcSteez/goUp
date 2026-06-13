package utils

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Returns a pointer to the db client
// Db is in WAL mode and the maximum number of open connections are one
func InitDB() (*sql.DB, error) {
	var conn_string string
	if Current_Config == nil {
		return nil, &NoConfigError{"Configuration", "Not found"}
	}
	if Current_Config.Database_Location != nil {
		conn_string = *Current_Config.Database_Location + "?_pragma=journal_mode(WAL)"
	} else {
		return nil, &NoConfigError{"db_path", "Not found"}
	}
	db, err := sql.Open("sqlite", conn_string)
	db.SetMaxOpenConns(1)
	if err != nil {
		return nil, err
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS service_data (
		"id" INTEGER PRIMARY KEY AUTOINCREMENT,
		"service_url" TEXT,
		"service_name" TEXT,
		"service_description" TEXT,
		"service_HTTP_response" TEXT,
		"service_API_response" TEXT,
		"service_response_time" INTEGER,
		"timestamp" TEXT,
		"error" INTEGER NOT NULL DEFAULT 0
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, err
	}

	// Drop superseded index from earlier schema versions before creating replacements.
	if _, err = db.Exec(`DROP INDEX IF EXISTS idx_service_data_lookup`); err != nil {
		return nil, err
	}

	indexes := []string{
		// Covers: GetRecentData subquery (GROUP BY service_name, MAX(id)),
		//         GetDataForService (WHERE service_name = ? ORDER BY id DESC),
		//         DbServiceDelete / DbServiceRename / DbGarbageCollect (WHERE service_name = ?)
		`CREATE INDEX IF NOT EXISTS idx_service_name_id
		 ON service_data(service_name, id)`,

		// Covers: GetErrorData (WHERE error = 1 ORDER BY timestamp ASC/DESC)
		`CREATE INDEX IF NOT EXISTS idx_error_timestamp
		 ON service_data(error, timestamp)`,
	}
	for _, idx := range indexes {
		if _, err = db.Exec(idx); err != nil {
			return nil, err
		}
	}

	if err := migrateResponseTimes(db); err != nil {
		log.Printf("Response time migration warning: %v", err)
	}

	log.Println("Database ready for queries")
	return db, nil
}

// migrateResponseTimes converts legacy TEXT response time values (e.g. "1.234ms")
// to INTEGER nanoseconds for existing rows written before the schema change.
func migrateResponseTimes(db *sql.DB) error {
	rows, err := db.Query("SELECT id, service_response_time FROM service_data")
	if err != nil {
		return err
	}
	defer rows.Close()

	type pending struct {
		id int
		ns int64
	}
	var updates []pending

	for rows.Next() {
		var id int
		var raw string
		if err := rows.Scan(&id, &raw); err != nil {
			continue
		}
		// Already an integer — no migration needed
		if _, err := strconv.ParseInt(raw, 10, 64); err == nil {
			continue
		}
		d, err := time.ParseDuration(raw)
		if err != nil {
			continue
		}
		updates = append(updates, pending{id, d.Nanoseconds()})
	}
	rows.Close()

	for _, u := range updates {
		if _, err := db.Exec("UPDATE service_data SET service_response_time = ? WHERE id = ?", u.ns, u.id); err != nil {
			log.Printf("Migration: failed to update row %d: %v", u.id, err)
		}
	}
	if len(updates) > 0 {
		log.Printf("Migrated %d rows: response time TEXT → INTEGER nanoseconds", len(updates))
	}
	return nil
}

// formatResponseTime formats nanoseconds into a frontend-friendly string.
// Output is always in ms or s to match the frontend parseMs regex.
// Trailing zeros and the decimal point are stripped (e.g. "12.00ms" → "12ms").
func formatResponseTime(ns int64) string {
	trimFloat := func(f float64, prec int) string {
		s := strconv.FormatFloat(f, 'f', prec, 64)
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
		return s
	}
	switch {
	case ns >= int64(time.Second):
		return trimFloat(float64(ns)/float64(time.Second), 2) + "s"
	case ns >= int64(time.Millisecond):
		return trimFloat(float64(ns)/float64(time.Millisecond), 2) + "ms"
	default:
		// Sub-millisecond: express as fractional ms so the frontend regex still matches
		return trimFloat(float64(ns)/float64(time.Millisecond), 3) + "ms"
	}
}

// scanServiceDataRow scans a single row from *sql.Rows into an int (id) and a ServiceData struct.
func scanServiceDataRow(row *sql.Rows) (int, ServiceData, error) {
	var id int
	var s ServiceData
	var tBuff string
	// Scan response time as []byte so the driver converts any stored type (INTEGER or TEXT)
	// to its string representation, avoiding interface{} type-assertion fragility.
	var rtBytes []byte
	err := row.Scan(
		&id,
		&s.ServiceURL,
		&s.ServiceName,
		&s.ServiceDescription,
		&s.ServiceHTTPResponse,
		&s.ServiceAPIResponse,
		&rtBytes,
		// timestamp is stored as RFC3339Nano string and parsed after scan
		&tBuff,
		&s.Error,
	)
	if err != nil {
		log.Printf("Failed scanning database row: %v", err)
	}
	if len(rtBytes) > 0 {
		raw := string(rtBytes)
		if ns, err := strconv.ParseInt(raw, 10, 64); err == nil {
			// INTEGER nanoseconds (current format)
			s.ServiceResponseTime = formatResponseTime(ns)
		} else {
			// Legacy TEXT value (e.g. "1.234ms") — pass through unchanged
			s.ServiceResponseTime = raw
		}
	}
	s.Timestamp, err = time.Parse(time.RFC3339Nano, tBuff)
	if err != nil {
		log.Printf("Error parsing timestamp in scan: %v", err)
	}
	return id, s, err
}

func InsertData(db *sql.DB, sd ServiceData) (retErr error) {
	insertSQL := `INSERT INTO service_data (service_url, service_name, service_description, service_HTTP_response, service_API_response, service_response_time, timestamp, error) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	statement, err := db.Prepare(insertSQL)
	if err != nil {
		return err
	}

	defer func() {
		if err := statement.Close(); err != nil {
			retErr = err
		}
	}()

	var rtNs int64
	if d, err := time.ParseDuration(sd.ServiceResponseTime); err == nil {
		rtNs = d.Nanoseconds()
	}

	_, err = statement.Exec(sd.ServiceURL, sd.ServiceName, sd.ServiceDescription, sd.ServiceHTTPResponse, sd.ServiceAPIResponse, rtNs, sd.Timestamp.Format(time.RFC3339Nano), sd.Error)
	if err != nil {
		return err
	}

	return nil
}

// Gets all service data
func GetData(db *sql.DB) (retSd []ServiceData, retErr error) {
	sd := []ServiceData{}

	row, err := db.Query("SELECT * FROM service_data;")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := row.Close(); err != nil {
			retErr = err
		}
	}()

	for row.Next() {
		_, s, err := scanServiceDataRow(row)
		if err != nil {
			return nil, err
		}
		sd = append(sd, s)
	}

	return sd, retErr
}

// Gets recent data defined by total number of service endpoints
func GetRecentData(db *sql.DB) (retSd []ServiceData, retErr error) {
	sd := []ServiceData{}

	statement := `SELECT * FROM service_data WHERE id IN (
		SELECT MAX(id) FROM service_data GROUP BY service_name
	) ORDER BY id DESC;`

	row, err := db.Query(statement)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := row.Close(); err != nil {
			retErr = err
		}
	}()

	// Snapshot current live service names to filter out stale DB rows
	// that may exist due to renames or deletes not yet garbage-collected.
	liveNames := make(map[string]struct{})
	if Current_Config != nil {
		for name := range Current_Config.Services {
			liveNames[name] = struct{}{}
		}
	}

	for row.Next() {
		_, s, err := scanServiceDataRow(row)
		if err != nil {
			return nil, err
		}
		if len(liveNames) == 0 {
			sd = append(sd, s)
			continue
		}
		if _, ok := liveNames[s.ServiceName]; ok {
			sd = append(sd, s)
		}
	}
	return sd, retErr
}

// Gets data for a specific service
func GetDataForService(db *sql.DB, name string) (retSd []ServiceData, retErr error) {
	sd := []ServiceData{}

	statement := "SELECT * FROM service_data WHERE service_name = ? ORDER BY id DESC"

	row, err := db.Query(statement, name)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := row.Close(); err != nil {
			retErr = err
		}
	}()
	for row.Next() {
		_, s, err := scanServiceDataRow(row)
		if err != nil {
			return nil, err
		}
		sd = append(sd, s)
	}
	return sd, retErr
}

// Gets all data from table where errors did occur.
// sortOrder controls timestamp sort direction: "asc" or "desc" (default).
func GetErrorData(db *sql.DB, limit int, sortOrder string) (retSd []ServiceData, retErr error) {
	sd := []ServiceData{}

	order := "DESC"
	if sortOrder == "asc" {
		order = "ASC"
	}

	statement := fmt.Sprintf("SELECT * FROM service_data WHERE error = 1 ORDER BY timestamp %s", order)
	if limit > 0 {
		statement = fmt.Sprintf("%s LIMIT %d", statement, limit)
	}
	row, err := db.Query(statement)
	if err != nil {
		return sd, err
	}
	defer func() {
		if err := row.Close(); err != nil {
			retErr = err
		}
	}()
	for row.Next() {
		_, s, err := scanServiceDataRow(row)
		if err != nil {
			return nil, err
		}
		sd = append(sd, s)
	}
	return sd, retErr
}

// Gets response time for each service
// Returns { service: response_time }
func GetResponseTimes(db *sql.DB) (svcRespTimes []ServiceResponseTime, retErr error) {
	statement := "SELECT * FROM service_data;"
	row, err := db.Query(statement)
	if err != nil {
		return svcRespTimes, err
	}
	defer func() {
		if err := row.Close(); err != nil {
			retErr = err
		}
	}()
	for row.Next() {
		_, s, err := scanServiceDataRow(row)
		if err != nil {
			return nil, err
		}
		svcRespTimes = append(svcRespTimes, ServiceResponseTime{Svc: s, ResponseTime: s.ServiceResponseTime})
	}
	return svcRespTimes, retErr
}

// Clears all table information from service_data and reclaims unused pages
func ClearDatabase(db *sql.DB) (retErr error) {

	conn, err := db.Conn(context.Background())
	if err != nil {
		return err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			retErr = err
		}
	}()
	statement := `DELETE FROM service_data;`

	if _, err := conn.ExecContext(context.Background(), statement); err != nil {
		return err
	}
	if _, err := conn.ExecContext(context.Background(), `PRAGMA journal_mode=DELETE;`); err != nil {
		return fmt.Errorf("clear completed but could not swtich journal mode: %w", err)
	}
	if _, err := conn.ExecContext(context.Background(), `VACUUM;`); err != nil {
		return fmt.Errorf("clear completed but VACUUM failed: %w", err)
	}
	if _, err := conn.ExecContext(context.Background(), `PRAGMA journal_mode=WAL;`); err != nil {
		return fmt.Errorf("vacuum completed but could not restore WAL mode: %w", err)
	}
	return retErr
}

// Cleans up database after exit
func CleanupDbFiles() error {
	log.Printf("Cleaning up database file")
	if Current_Config == nil || Current_Config.Database_Location == nil {
		return fmt.Errorf("no database location configured, skipping cleanup")
	}
	if err := os.Remove(*Current_Config.Database_Location); err != nil {
		return err
	}
	log.Println("Database file succesfully removed")
	return nil
}

func DbServiceDelete(db *sql.DB, service Service) error {
	res, err := db.Exec(`DELETE FROM service_data WHERE service_name = ?`, service.Name)
	if err != nil {
		return err
	}

	rows_affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	log.Printf("Removed all instances of %s, rows affected: %d", service.Name, rows_affected)

	return nil
}

func DbServiceRename(db *sql.DB, oldName string, newName string) error {
	res, err := db.Exec(`UPDATE service_data SET service_name = ? WHERE service_name = ?`, newName, oldName)
	if err != nil {
		return err
	}

	rows_affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	log.Printf("Renamed %s to %s in database, rows affected: %d", oldName, newName, rows_affected)

	res, err = db.Exec(`DELETE FROM service_data WHERE service_name = ?`, oldName)
	if err != nil {
		return err
	}
	rows_affected, err = res.RowsAffected()
	if err != nil {
		return err
	}

	log.Printf("Removed all orphan rows in database with old name %s, rows affected: %d", oldName, rows_affected)

	return nil
}

func DbGarbageCollect(db *sql.DB, conf *Config) error {
	if conf == nil {
		return fmt.Errorf("no config loaded")
	}

	names := make([]any, 0, len(conf.Services))
	for name := range conf.Services {
		names = append(names, name)
	}

	if len(names) == 0 {
		res, err := db.Exec(`DELETE FROM service_data`)
		if err != nil {
			return err
		}
		rows, _ := res.RowsAffected()
		log.Printf("GC: removed %d orphaned rows (no services in config)", rows)
		return nil
	}

	placeholders := make([]byte, 0, len(names)*2)
	for i := range names {
		if i > 0 {
			placeholders = append(placeholders, ',', '?')
		} else {
			placeholders = append(placeholders, '?')
		}
	}

	stmt := fmt.Sprintf(`DELETE FROM service_data WHERE service_name NOT IN (%s)`, placeholders)
	res, err := db.Exec(stmt, names...)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows > 0 {
		log.Printf("GC: removed %d orphaned rows not matching current config", rows)
	}
	return nil
}
