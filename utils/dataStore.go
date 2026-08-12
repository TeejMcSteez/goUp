package utils

import (
	"context"
	"database/sql"
	"fmt"
	migrations "goUp/db/migrations"
	database "goUp/internal/db"
	"log/slog"
	"maps"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pressly/goose/v3"
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

	if err := runMigrations(db); err != nil {
		return nil, err
	}

	slog.Info("Database ready for queries")
	return db, nil
}

// runMigrations applies the goose migration set in db/migrations.
//
// Databases that predate goose (every database created before this change
// shipped) already have some or all of that schema history applied via the
// old hand-rolled DDL in this file. bootstrapLegacyVersion detects that case
// and marks the already-applied migrations as done without re-running them,
// so goose only executes genuinely new migrations against those databases.
func runMigrations(db *sql.DB) error {
	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}
	if err := bootstrapLegacyVersion(db); err != nil {
		return fmt.Errorf("failed to baseline goose version for existing database: %w", err)
	}
	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("failed to run database migrations: %w", err)
	}
	return nil
}

// bootstrapLegacyVersion marks migrations that already ran via the old
// hand-rolled schema setup as applied, without re-running them, so goose
// only executes new migrations against a database that predates it.
//
// goose.EnsureDBVersion creates its own version-tracking table
// (goose_db_version) the first time it's called against a database. A
// version of 0 there means goose has never touched this database before —
// which is true for every database that exists today, since goose is only
// being introduced now. If service_data already exists at that point, this
// is one of those pre-goose databases, and we inspect its current schema to
// figure out which migrations already happened before marking them applied.
func bootstrapLegacyVersion(db *sql.DB) error {
	current, err := goose.EnsureDBVersion(db)
	if err != nil {
		return err
	}
	if current > 0 {
		// goose has run against this database before; nothing to baseline.
		return nil
	}

	var tableExists int
	if err := db.QueryRow(
		`SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'service_data'`,
	).Scan(&tableExists); err != nil {
		return err
	}
	if tableExists == 0 {
		// Brand-new database — let goose create everything for real.
		return nil
	}

	// service_data already exists, so migration 1 (table + indexes) and
	// migration 2 (response time normalization, which the old code ran on
	// every boot) already happened under the old hand-rolled setup.
	baseline := []int64{1, 2}

	var hasActive int
	if err := db.QueryRow(
		`SELECT count(*) FROM pragma_table_info('service_data') WHERE name = 'active'`,
	).Scan(&hasActive); err != nil {
		return err
	}
	if hasActive == 1 {
		baseline = append(baseline, 3)
	}

	for _, version := range baseline {
		if _, err := db.Exec(
			`INSERT INTO goose_db_version (version_id, is_applied) VALUES (?, 1)`, version,
		); err != nil {
			return err
		}
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
		slog.Error("Error parsing timestamp in scan", "error", err)
	}
	return s, err
}

func rowToTlsStatus(t database.TlsStatus) TlsStatus {
	firstSeen, err := time.Parse(time.RFC3339Nano, t.FirstSeen)
	if err != nil {
		slog.Error("failed to parse row first seen to time", "error", err)
	}
	lastChecked, err := time.Parse(time.RFC3339Nano, t.LastChecked)
	if err != nil {
		slog.Error("failed to parse row last checked to time", "error", err)
	}
	return TlsStatus{
		ServiceName:  t.ServiceName,
		Fingerprint:  t.Fingerprint,
		Not_after:    time.Unix(t.NotAfter, 0),
		Subject:      t.Subject.String,
		Issuer:       t.Issuer.String,
		Is_expired:   t.IsExpired == 1,
		Chain:        t.Chain.String,
		First_seen:   firstSeen,
		Last_checked: lastChecked,
	}
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

func UpsertTls(db *sql.DB, t TlsStatus) error {
	payload := database.InsertTlsStatusParams{
		ServiceName: t.ServiceName,
		Fingerprint: t.Fingerprint,
		NotAfter:    t.Not_after.Unix(),
		Subject:     sql.NullString{String: t.Subject},
		Issuer:      sql.NullString{String: t.Issuer},
		IsExpired:   boolToInt(t.Is_expired),
		Chain:       sql.NullString{String: t.Chain},
		FirstSeen:   t.First_seen.Format(time.RFC3339Nano),
		LastChecked: t.Last_checked.Format(time.RFC3339Nano),
	}
	return database.New(db).InsertTlsStatus(context.Background(), payload)
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

func GetExpiredTls(db *sql.DB) ([]TlsStatus, error) {
	rows, err := database.New(db).GetTlsStatus(context.Background())
	if err != nil {
		return nil, err
	}

	var status []TlsStatus
	for _, r := range rows {
		status = append(status, rowToTlsStatus(r))
	}
	return status, nil
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
	slog.Info("Cleaning up database file")
	if Current_Config == nil || Current_Config.Database_Location == nil {
		return fmt.Errorf("no database location configured, skipping cleanup")
	}
	if err := os.Remove(*Current_Config.Database_Location); err != nil {
		return err
	}
	slog.Info("Database file succesfully removed")
	return nil
}

func DbServiceDelete(db *sql.DB, service Service) error {
	if err := database.New(db).DeleteService(context.Background(), sql.NullString{String: service.Name, Valid: true}); err != nil {
		return err
	}

	slog.Info("Removed all instances of service", "service", service.Name)

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

	slog.Info("Renamed service in database", "old_name", oldName, "new_name", newName)

	// Removes any orphan rows written under the old name by a concurrent
	// insert that raced with the rename above.
	if err := q.DeleteService(ctx, sql.NullString{String: oldName, Valid: true}); err != nil {
		return err
	}

	slog.Info("Removed all orphan rows in database", "old_name", oldName)

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
		slog.Info("GC: removed all rows (no services in config)")
		return nil
	}

	names := make([]sql.NullString, 0, len(conf.Services))
	for name := range conf.Services {
		names = append(names, sql.NullString{String: name, Valid: true})
	}

	if err := q.Cleanup(ctx, names); err != nil {
		return err
	}

	slog.Info("GC: removed orphaned rows not matching current config")
	return nil
}
