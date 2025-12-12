package utils

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"fmt"
	"log"
)

func InitDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./serviceData.db")
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

func InsertData(db *sql.DB, sd ServiceData) {
	insertSQL := `INSERT INTO service_data (service_name, service_HTTP_response, service_API_response, service_response_time) VALUES (?, ?, ?, ?)`
	statement, err := db.Prepare(insertSQL)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := statement.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	_, err = statement.Exec(sd.ServiceName, sd.ServiceHTTPResponse, sd.ServiceAPIResponse, sd.ServiceResponseTime)
	if err != nil {
		log.Fatal(err)
	}
}

func GetData(db *sql.DB) []ServiceData {
	sd := []ServiceData{}

	row, err := db.Query("SELECT * FROM service_data;")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := row.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	for row.Next() {
		var id int
		var s ServiceData
		err = row.Scan(&id, &s.ServiceName, &s.ServiceHTTPResponse, &s.ServiceAPIResponse, &s.ServiceResponseTime)
		if err != nil {
			log.Fatal(err)
		}
		sd = append(sd, s)
	}

	return sd
}

func GetRecentData(db *sql.DB) []ServiceData {
	numOfServices := len(GetServiceEndpoints())
	sd := []ServiceData{}

	statement := fmt.Sprintf("SELECT * FROM service_data ORDER BY id DESC LIMIT %v;", numOfServices)

	row, err := db.Query(statement)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := row.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	for row.Next() {
		var id int
		var s ServiceData
		// Might implement a scanner for service data so I can just pass in struct
		err = row.Scan(&id, &s.ServiceName, &s.ServiceHTTPResponse, &s.ServiceAPIResponse, &s.ServiceResponseTime)
		if err != nil {
			log.Fatal(err)
		}
		sd = append(sd, s)
	}
	return sd

}

func GetDataForService(db *sql.DB, name string) []ServiceData {
	sd := []ServiceData{}

	statement := "SELECT * FROM service_data WHERE service_name = ? ORDER BY id DESC"

	row, err := db.Query(statement, name)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := row.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	for row.Next() {
		var id int
		var s ServiceData
		err = row.Scan(&id, &s.ServiceName, &s.ServiceHTTPResponse, &s.ServiceAPIResponse, &s.ServiceResponseTime)
		if err != nil {
			log.Fatal(err)
		}
		sd = append(sd, s)
	}
	return sd
}