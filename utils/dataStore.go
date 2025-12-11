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
		"service_API_response" TEXT
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Database and table ready for queries")
	return db
}

func InsertData(db *sql.DB, sd ServiceData) {
	insertSQL := `INSERT INTO service_data (service_name, service_HTTP_response, service_API_response) VALUES (?, ?, ?)`
	statement, err := db.Prepare(insertSQL)
	if err != nil {
		log.Fatal(err)
	}

	defer statement.Close()

	_, err = statement.Exec(sd.ServiceName, sd.ServiceHTTPResponse, sd.ServiceAPIResponse)
	if err != nil {
		log.Fatal(err)
	}
}

func GetData(db *sql.DB) []ServiceData {
	sd := []ServiceData{}

	row, err := db.Query("SELECT id, service_name, service_HTTP_response, service_API_response FROM service_data;")
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()

	for row.Next() {
		var id int
		var svc_name string
		var svc_http_res string
		var svc_api_res string
		err = row.Scan(&id, &svc_name, &svc_http_res, &svc_api_res)
		if err != nil {
			log.Fatal(err)
		}
		sd = append(sd, ServiceData{ServiceName: svc_name, ServiceHTTPResponse: svc_http_res, ServiceAPIResponse: svc_api_res})
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
	defer row.Close()

	for row.Next() {
		var id int
		var svc_name string
		var svc_http_res string
		var svc_api_res string
		err = row.Scan(&id, &svc_name, &svc_http_res, &svc_api_res)
		if err != nil {
			log.Fatal(err)
		}
		sd = append(sd, ServiceData{ServiceName: svc_name, ServiceHTTPResponse: svc_http_res, ServiceAPIResponse: svc_api_res})
	}
	return sd

}