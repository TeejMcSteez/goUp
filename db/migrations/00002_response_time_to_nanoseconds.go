package migrations

import (
	"context"
	"database/sql"
	"log"
	"strconv"
	"time"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upResponseTimeToNanoseconds, downResponseTimeToNanoseconds)
}

// upResponseTimeToNanoseconds converts legacy TEXT response time values
// (e.g. "1.234ms") to INTEGER nanoseconds. Values that already parse as an
// integer are left untouched, so this is safe to run against a table that's
// a mix of old and new rows.
func upResponseTimeToNanoseconds(ctx context.Context, tx *sql.Tx) (err error) {
	rows, err := tx.QueryContext(ctx, "SELECT id, service_response_time FROM service_data")
	if err != nil {
		return err
	}
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
			// NULL or otherwise unscannable as a string — nothing to convert.
			continue
		}
		if _, err := strconv.ParseInt(raw, 10, 64); err == nil {
			continue
		}
		d, err := time.ParseDuration(raw)
		if err != nil {
			continue
		}
		updates = append(updates, pending{id, d.Nanoseconds()})
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, u := range updates {
		if _, err := tx.ExecContext(ctx, "UPDATE service_data SET service_response_time = ? WHERE id = ?", u.ns, u.id); err != nil {
			return err
		}
	}
	if len(updates) > 0 {
		log.Printf("Migrated %d rows: response time TEXT -> INTEGER nanoseconds", len(updates))
	}
	return err
}

// downResponseTimeToNanoseconds is a no-op: the original TEXT formatting
// (e.g. "1.234ms" vs "1234000ns") isn't recoverable from the nanosecond
// integer, and nanoseconds remain a valid value for every downstream reader.
func downResponseTimeToNanoseconds(_ context.Context, _ *sql.Tx) error {
	return nil
}
