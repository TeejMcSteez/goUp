package utils

import (
	"context"
	"database/sql"
	"fmt"
	database "goUp/internal/db"
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
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	createTableSQL := `CREATE TABLE IF NOT EXISTS service_data (
		"id" INTEGER PRIMARY KEY AUTOINCREMENT,
		"service_url" TEXT,
		"service_name" TEXT,
		"service_description" TEXT,
		"service_HTTP_response" TEXT,
		"service_API_response" TEXT,
		"service_response_time" INTEGER,
		"timestamp" TEXT,
		"error" INTEGER NOT NULL DEFAULT 0,
		"active" INTEGER NOT NULL DEFAULT 1
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
//
// This must run before any sqlc-generated query scans service_response_time,
// since those queries scan it into sql.NullInt64 and a legacy TEXT value like
// "1.234ms" is not a valid int64.
func migrateResponseTimes(db *sql.DB) (err error) {
	rows, err := db.Query("SELECT id, service_response_time FROM service_data")
	if err != nil {
		return err
	}
	// Close should emit err but to be idiomatic and fix IDE warning
	// Check for row errors before closing rows
	defer func() {
		if e := rows.Err(); e != nil {
			err = e
		}
	}()
	defer func() {
		if e := rows.Close(); e != nil {
			err = e
		}
	}()

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

	for _, u := range updates {
		if _, err := db.Exec("UPDATE service_data SET service_response_time = ? WHERE id = ?", u.ns, u.id); err != nil {
			log.Printf("Migration: failed to update row %d: %v", u.id, err)
			return err
		}
	}
	if len(updates) > 0 {
		log.Printf("Migrated %d rows: response time TEXT → INTEGER nanoseconds", len(updates))
	}
	return err
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

// rowToServiceData converts a sqlc-generated row into the ServiceData shape
// used throughout the rest of the app.
func rowToServiceData(d database.ServiceDatum) (ServiceData, error) {
	s := ServiceData{
		ServiceURL:          d.ServiceUrl.String,
		ServiceName:         d.ServiceName.String,
		ServiceDescription:  d.ServiceDescription.String,
		ServiceHTTPResponse: d.ServiceHttpResponse.String,
		ServiceAPIResponse:  d.ServiceApiResponse.String,
		Error:               d.Error != 0,
		Active:              d.Active != 0,
	}
	if d.ServiceResponseTime.Valid {
		s.ServiceResponseTime = formatResponseTime(d.ServiceResponseTime.Int64)
	}
	var err error
	s.Timestamp, err = time.Parse(time.RFC3339Nano, d.Timestamp.String)
	if err != nil {
		log.Printf("Error parsing timestamp in scan: %v", err)
	}
	return s, err
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func InsertData(db *sql.DB, sd ServiceData) error {
	var rtNs int64
	if d, err := time.ParseDuration(sd.ServiceResponseTime); err == nil {
		rtNs = d.Nanoseconds()
	}

	q := database.New(db)
	return q.InsertData(context.Background(), database.InsertDataParams{
		ServiceUrl:          sql.NullString{String: sd.ServiceURL, Valid: true},
		ServiceName:         sql.NullString{String: sd.ServiceName, Valid: true},
		ServiceDescription:  sql.NullString{String: sd.ServiceDescription, Valid: true},
		ServiceHttpResponse: sql.NullString{String: sd.ServiceHTTPResponse, Valid: true},
		ServiceApiResponse:  sql.NullString{String: sd.ServiceAPIResponse, Valid: true},
		ServiceResponseTime: sql.NullInt64{Int64: rtNs, Valid: true},
		Timestamp:           sql.NullString{String: sd.Timestamp.Format(time.RFC3339Nano), Valid: true},
		Error:               boolToInt(sd.Error),
		Active:              boolToInt(sd.Active),
	})
}

// Gets all service data
func GetData(db *sql.DB) ([]ServiceData, error) {
	rows, err := database.New(db).GetAllData(context.Background())
	if err != nil {
		return nil, err
	}

	sd := []ServiceData{}
	// Behavior change with sqlc refactor
	// If a timestamp error occurs will return nill and err out without continuing struct creation
	// Could log the error and continue but still deciding on wanted outcome
	for _, r := range rows {
		s, err := rowToServiceData(r)
		if err != nil {
			return nil, err
		}
		sd = append(sd, s)
	}
	return sd, nil
}

// Gets recent data defined by total number of service endpoints
func GetRecentData(db *sql.DB) ([]ServiceData, error) {
	rows, err := database.New(db).GetRecentData(context.Background())
	if err != nil {
		return nil, err
	}

	// Snapshot current live services to filter out stale DB rows that may
	// exist due to renames or deletes not yet garbage-collected, and to
	// overlay the current Active state — a row's stored Active value is
	// only a snapshot from whenever it was last written, which goes stale
	// the moment a service is toggled and stops being fetched.
	liveServices := make(map[string]Service)
	if Current_Config != nil {
		maps.Copy(liveServices, Current_Config.Services)
	}

	sd := []ServiceData{}
	for _, r := range rows {
		s, err := rowToServiceData(r)
		if err != nil {
			return nil, err
		}
		if len(liveServices) == 0 {
			sd = append(sd, s)
			continue
		}
		if svc, ok := liveServices[s.ServiceName]; ok {
			s.Active = svc.IsActive()
			sd = append(sd, s)
		}
	}
	return sd, nil
}

// Gets data for a specific service
func GetDataForService(db *sql.DB, name string) ([]ServiceData, error) {
	rows, err := database.New(db).GetDataForService(context.Background(), sql.NullString{String: name, Valid: true})
	if err != nil {
		return nil, err
	}

	sd := []ServiceData{}
	for _, r := range rows {
		s, err := rowToServiceData(r)
		if err != nil {
			return nil, err
		}
		sd = append(sd, s)
	}
	return sd, nil
}

// Gets all data from table where errors did occur.
// sortOrder controls timestamp sort direction: "asc" or "desc" (default).
func GetErrorData(db *sql.DB, limit int, sortOrder string) ([]ServiceData, error) {
	q := database.New(db)

	// SQLite treats a negative LIMIT as "no limit".
	l := int64(-1)
	if limit > 0 {
		l = int64(limit)
	}

	var rows []database.ServiceDatum
	var err error
	if sortOrder == "asc" {
		rows, err = q.GetErrorDataAsc(context.Background(), l)
	} else {
		rows, err = q.GetErrorDataDesc(context.Background(), l)
	}
	if err != nil {
		return []ServiceData{}, err
	}

	sd := []ServiceData{}
	for _, r := range rows {
		s, err := rowToServiceData(r)
		if err != nil {
			return nil, err
		}
		sd = append(sd, s)
	}
	return sd, nil
}

// Gets response time for each service
// Returns { service: response_time }
func GetResponseTimes(db *sql.DB) ([]ServiceResponseTime, error) {
	rows, err := database.New(db).GetAllData(context.Background())
	if err != nil {
		return nil, err
	}

	var svcRespTimes []ServiceResponseTime
	for _, r := range rows {
		s, err := rowToServiceData(r)
		if err != nil {
			return nil, err
		}
		svcRespTimes = append(svcRespTimes, ServiceResponseTime{Svc: s, ResponseTime: s.ServiceResponseTime})
	}
	return svcRespTimes, nil
}

// Clears all table information from service_data and reclaims unused pages
func ClearDatabase(db *sql.DB) (retErr error) {
	// Pass context with timeout
	// Otherwise insert can run before or after ClearDatabase
	// This can cause undeterminant outcomes
	// Passing new context will either clear to 0 rows or error
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			retErr = err
		}
	}()

	if err := database.New(conn).ClearServiceData(context.Background()); err != nil {
		return err
	}
	if _, err := conn.ExecContext(context.Background(), `PRAGMA journal_mode=DELETE;`); err != nil {
		return fmt.Errorf("clear completed but could not switch journal mode: %w", err)
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
	if err := database.New(db).DeleteService(context.Background(), sql.NullString{String: service.Name, Valid: true}); err != nil {
		return err
	}

	log.Printf("Removed all instances of %s", service.Name)

	return nil
}

func DbServiceRename(db *sql.DB, oldName string, newName string) error {
	q := database.New(db)
	ctx := context.Background()

	if err := q.ServiceRename(ctx, database.ServiceRenameParams{
		ServiceName:   sql.NullString{String: newName, Valid: true},
		ServiceName_2: sql.NullString{String: oldName, Valid: true},
	}); err != nil {
		return err
	}

	log.Printf("Renamed %s to %s in database", oldName, newName)

	// Removes any orphan rows written under the old name by a concurrent
	// insert that raced with the rename above.
	if err := q.DeleteService(ctx, sql.NullString{String: oldName, Valid: true}); err != nil {
		return err
	}

	log.Printf("Removed all orphan rows in database with old name %s", oldName)

	return nil
}

func DbGarbageCollect(db *sql.DB, conf *Config) error {
	if conf == nil {
		return fmt.Errorf("no config loaded")
	}

	q := database.New(db)
	ctx := context.Background()

	if len(conf.Services) == 0 {
		if err := q.ClearServiceData(ctx); err != nil {
			return err
		}
		log.Printf("GC: removed all rows (no services in config)")
		return nil
	}

	names := make([]sql.NullString, 0, len(conf.Services))
	for name := range conf.Services {
		names = append(names, sql.NullString{String: name, Valid: true})
	}

	if err := q.Cleanup(ctx, names); err != nil {
		return err
	}

	log.Printf("GC: removed orphaned rows not matching current config")
	return nil
}
