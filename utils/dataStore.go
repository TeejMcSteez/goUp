package utils

import (
	"database/sql"
	"fmt"
	"log"
	"os"
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
		"service_name" TEXT,
		"service_HTTP_response" TEXT,
		"service_API_response" TEXT,
		"service_response_time" TEXT,
		"timestamp" TEXT,
		"error" INTEGER NOT NULL DEFAULT 0
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, err
	}

	log.Println("Database and table ready for queries")
	return db, nil
}

// scanServiceDataRow scans a single row from *sql.Rows into an int (id) and a ServiceData struct.
func scanServiceDataRow(row *sql.Rows) (int, ServiceData, error) {
	var id int
	var s ServiceData
	var tBuff string
	err := row.Scan(
		&id,
		&s.ServiceName,
		&s.ServiceHTTPResponse,
		&s.ServiceAPIResponse,
		&s.ServiceResponseTime,
		&tBuff,
		&s.Error,
	)
	if err != nil {
		log.Printf("Failed scanning database row: %v", err)
	}
	s.Timestamp, err = time.Parse(time.RFC3339Nano, tBuff)
	if err != nil {
		log.Printf("Error parsing timestamp in scan: %v", err)
	}
	return id, s, err
}

func InsertData(db *sql.DB, sd ServiceData) (retErr error) {
	insertSQL := `INSERT INTO service_data (service_name, service_HTTP_response, service_API_response, service_response_time, timestamp, error) VALUES (?, ?, ?, ?, ?, ?)`
	statement, err := db.Prepare(insertSQL)
	if err != nil {
		return err
	}

	defer func() {
		if err := statement.Close(); err != nil {
			retErr = err
		}
	}()
	_, err = statement.Exec(sd.ServiceName, sd.ServiceHTTPResponse, sd.ServiceAPIResponse, sd.ServiceResponseTime, sd.Timestamp.Format(time.RFC3339Nano), sd.Error)
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
	numOfServices := len(GetServiceEndpoints())
	sd := []ServiceData{}

	statement := fmt.Sprintf("SELECT * FROM service_data ORDER BY id DESC LIMIT %v;", numOfServices)

	row, err := db.Query(statement)
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

// Clears all table information from service_data and reclaims unused pages
func ClearDatabase(db *sql.DB) error {
	statement := `DELETE FROM service_data;`

	res, err := db.Exec(statement)
	if err != nil {
		return err
	}
	if rowsAffected, err := res.RowsAffected(); err != nil {
		return err
	} else {
		log.Printf("Cleared databased, Rows Affected: %v", rowsAffected)
	}

	// Temporarily switch out of WAL mode so VACUUM replaces the file rather
	// than checkpointing in-place, which physically shrinks the file on disk.
	if _, err := db.Exec("PRAGMA journal_mode=DELETE;"); err != nil {
		return fmt.Errorf("clear succeeded but could not switch journal mode: %w", err)
	}
	if _, err := db.Exec("VACUUM;"); err != nil {
		return fmt.Errorf("clear succeeded but vacuum failed: %w", err)
	}
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return fmt.Errorf("vacuum succeeded but could not restore WAL mode: %w", err)
	}

	return nil
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

	statement := `DELETE FROM service_data WHERE service_name = ?`

	res, err := db.Exec(statement, service.Name)
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
