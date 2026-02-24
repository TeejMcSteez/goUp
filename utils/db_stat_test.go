package utils_test

import (
	"goUp/utils"
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
		t.Fatalf("Error occured during setup: %v", err)
	}

	db, err := utils.InitDB()
	if err != nil {
		t.Fatalf("Error occured during DB init: %v", err)
	}
	defer func() {
		db.Close()
		os.Remove(dbPath)
	}()

	if _, err := utils.GetDatabaseSize(); err != nil {
		t.Fatalf("Error occured getting database statistics: %v", err)
	}
}
