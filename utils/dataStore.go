package utils

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
	"log"
)

func InitDB() *sql.DB {
	db, err := sql.Open("sqlite", "./serviceData.db?_pragma=journal_mode(WAL)")
	db.SetMaxOpenConns(1)
	if err != nil {
		log.Fatal(err)
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS service_data (
		"id" INTEGER PRIMARY KEY AUTOINCREMENT,
		"service_name" TEXT,
		"service_HTTP_response" TEXT,
		"service_API_response" TEXT,
		"service_response_time" TEXT
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Database and table ready for queries")
	return db
}

func InsertData(db *sql.DB, sd ServiceData) (retErr error) {
	insertSQL := `INSERT INTO service_data (service_name, service_HTTP_response, service_API_response, service_response_time) VALUES (?, ?, ?, ?)`
	statement, err := db.Prepare(insertSQL)
	if err != nil {
		return err
	}

	defer func() {
		if err := statement.Close(); err != nil {
			retErr = err
		}
	}()

	_, err = statement.Exec(sd.ServiceName, sd.ServiceHTTPResponse, sd.ServiceAPIResponse, sd.ServiceResponseTime)
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
		err = row.Scan(&id, &s.ServiceName, &s.ServiceHTTPResponse, &s.ServiceAPIResponse, &s.ServiceResponseTime)
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
		err = row.Scan(&id, &s.ServiceName, &s.ServiceHTTPResponse, &s.ServiceAPIResponse, &s.ServiceResponseTime)
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
		err = row.Scan(&id, &s.ServiceName, &s.ServiceHTTPResponse, &s.ServiceAPIResponse, &s.ServiceResponseTime)
		if err != nil {
			return nil, err
		}
		sd = append(sd, s)
	}
	return sd, retErr
}
