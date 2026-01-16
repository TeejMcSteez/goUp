package utils_test

import (
	"goUp/utils"
	"log"
	"os"
	"testing"
)

func TestGetStat(t *testing.T) {
	dbPath := "./test_data.db"
	ymlContent := `db_path: "` + dbPath + `"
services:
  service2:
    url: "https://example.com"
    retry: 2
  service4:
    url: "https://www.google.com"
    retry: 1
`
	cleanup := createTestYML(ymlContent, t)
	defer cleanup()

	if err := utils.Setup(); err != nil {
		log.Printf("Error occured during setup: %v", err)
		t.Fail()
	}

	db, err := utils.InitDB()
	if err != nil {
		log.Printf("Error occured during DB init: %v", err)
		t.Fail()
	}
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

	data, err := utils.GetDatabaseSize()
	if err != nil {
		log.Printf("Error occured getting database statistics: %v", err)
		t.Fail()
	}
	log.Printf("Database file size (bytes): %v", data)

}
