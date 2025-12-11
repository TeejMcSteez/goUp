package utils

import (
	"database/sql"
	"fmt"
	"log"
)

// Creates a new empty data store for ServiceData
func NewStore() *SharedData {
	return &SharedData{
		data: make([]ServiceData, 0),
	}
}

// Writer: Sets new data
func (s *SharedData) Set(data []ServiceData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = data
}

// Reader: get a copy of the current slice
func (s *SharedData) Get() []ServiceData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]ServiceData, len(s.data))
	copy(out, s.data)
	return out
}

func InitDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./serviceData.db")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	createTableSQL := `CREATE TBALE IF NOT EXISTS service_data (
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

	row, err := db.Query("SELECT id, service_name, service_HTTP_response, service_API_response from service_data")
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()

	for row.Next() {
		var svc_name string
		var svc_http_res string
		var svc_api_res string
		err = row.Scan(&svc_name, &svc_http_res, &svc_api_res)
		if err != nil {
			log.Fatal(err)
		}
		sd = append(sd, ServiceData{ServiceName: svc_name, ServiceHTTPResponse: svc_http_res, ServiceAPIResponse: svc_api_res})
	}

	return sd
}