package utils

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

// Returns a pointer to the db client
// Db is in WAL mode and the maximum number of open connections are one
func InitDB() (*sql.DB, error) {
	var conn_string string = "./serviceData.db?_pragma=journal_mode(WAL)"
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
		"error" INTEGER NOT NULL DEFAULT 0
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, err
	}

	log.Println("Database and table ready for queries")
	return db, nil
}

func InsertData(db *sql.DB, sd ServiceData) (retErr error) {
	insertSQL := `INSERT INTO service_data (service_name, service_HTTP_response, service_API_response, service_response_time, error) VALUES (?, ?, ?, ?, ?)`
	statement, err := db.Prepare(insertSQL)
	if err != nil {
		return err
	}

	defer func() {
		if err := statement.Close(); err != nil {
			retErr = err
		}
	}()

	_, err = statement.Exec(sd.ServiceName, sd.ServiceHTTPResponse, sd.ServiceAPIResponse, sd.ServiceResponseTime, sd.Error)
	if err != nil {
		return err
	}

	return nil
}

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
		var id int
		var s ServiceData
		err = row.Scan(&id, &s.ServiceName, &s.ServiceHTTPResponse, &s.ServiceAPIResponse, &s.ServiceResponseTime, &s.Error)
		if err != nil {
			return nil, err
		}
		sd = append(sd, s)
	}

	return sd, retErr
}

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
		var id int
		var s ServiceData
		// Might implement a scanner for service data so I can just pass in struct
		err = row.Scan(&id, &s.ServiceName, &s.ServiceHTTPResponse, &s.ServiceAPIResponse, &s.ServiceResponseTime, &s.Error)
		if err != nil {
			return nil, err
		}
		sd = append(sd, s)
	}
	return sd, retErr

}

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
		var id int
		var s ServiceData
		err = row.Scan(&id, &s.ServiceName, &s.ServiceHTTPResponse, &s.ServiceAPIResponse, &s.ServiceResponseTime, &s.Error)
		if err != nil {
			return nil, err
		}
		sd = append(sd, s)
	}
	return sd, retErr
}

// Cleans up database after exit
func CleanupDb() error {
	log.Printf("Cleaning up database file")
	if Current_Config.Database_Location != nil {
		if err := os.Remove(*Current_Config.Database_Location); err != nil {
			return err
		}
	} else {
		if err := os.Remove("./serviceData.db"); err != nil {
			return err
		}
	}
	log.Println("Database file succesfully removed")
	return nil
}

// Sqlite comes built in with dbstat table, this function returns that built in table
func GetDbStat(db *sql.DB) (dbStat DatabaseStatistic, retErr error) {
	statement := "SELECT * FROM dbstat"
	row, err := db.Query(statement)
	if err != nil {
		return DatabaseStatistic{}, nil
	}
	defer func() {
		if err := row.Close(); err != nil {
			retErr = err
		}
	}()
	for row.Next() {
		if err := row.Scan(&dbStat.name, &dbStat.path, &dbStat.pageno, &dbStat.pagetype, &dbStat.ncell, &dbStat.payload, &dbStat.unused, &dbStat.mx_payload, &dbStat.pgoffset, &dbStat.pgsize); err != nil {
			return DatabaseStatistic{}, err
		}
	}
	return dbStat, err
}
