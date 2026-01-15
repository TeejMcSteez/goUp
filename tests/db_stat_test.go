package utils_test

import (
	"goUp/utils"
	"log"
	"testing"
)

// TODO: Cleanup db files after test
func TestGetStat(t *testing.T) {
	ymlContent := `db_path: "./test_data.db"
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

	data, err := utils.GetDatabaseSize(db)
	if err != nil {
		log.Printf("Error occured getting database statistics: %v", err)
		t.Fail()
	}
	log.Printf("Database file size (bytes): %v", data)

}
