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

	cfg, err := utils.LoadConfig("services.yml")
	if err != nil {
		t.Fatalf("Failed to load configuration file: %v", err)
	}

	if err := utils.Setup(cfg); err != nil {
		t.Fatalf("Error occured during setup: %v", err)
	}

	db, err := utils.InitDB()
	if err != nil {
		t.Fatalf("Error occured during DB init: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
		if err := os.Remove(dbPath); err != nil {
			t.Errorf("Error removing database file: %v", err)
		}
	}()

	if _, err := utils.GetDatabaseSize(); err != nil {
		t.Fatalf("Error occured getting database statistics: %v", err)
	}
}

func TestStat(t *testing.T) {
	dbPath := "./test_size.db"
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

	cfg, err := utils.LoadConfig("services.yml")
	if err != nil {
		t.Fatalf("Failed to load configuration file: %v", err)
	}

	if err := utils.Setup(cfg); err != nil {
		t.Fatalf("Error occured during setup: %v", err)
	}

	db, err := utils.InitDB()
	if err != nil {
		t.Fatalf("Error occured during DB init: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
		if err := os.Remove(dbPath); err != nil {
			t.Errorf("Error removing database file: %v", err)
		}
	}()

	size, err := utils.GetDatabaseSize()
	if err != nil {
		t.Fatalf("Error occured getting database statistics: %v", err)
	}
	dbInfo, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("failed to get file stat")
	}
	shmInfo, err := os.Stat(dbPath + "-shm")
	if err != nil {
		t.Fatalf("failed to get file stat")
	}
	walInfo, err := os.Stat(dbPath + "-wal")
	if err != nil {
		t.Fatalf("failed to get file stat")
	}
	if size-(shmInfo.Size()+walInfo.Size()) != dbInfo.Size() {
		t.Fatalf("main db file minus shm+wal is not the same size on no write\nGot size: %v\nStat size: %v\n", size, dbInfo.Size())
	}
	if size-(dbInfo.Size()+walInfo.Size()) != shmInfo.Size() {
		t.Fatalf("shm db file minus db+wal is not the same size on no write\nGot size: %v\nStat size: %v\n", size, dbInfo.Size())
	}
	if size-(dbInfo.Size()+shmInfo.Size()) != walInfo.Size() {
		t.Fatalf("wal db file minus db+shm is not the same size on no write\nGot size: %v\nStat size: %v\n", size, dbInfo.Size())
	}
	if size != (dbInfo.Size() + shmInfo.Size() + walInfo.Size()) {
		t.Fatalf("got not equal to Stat added sizes\nGot size: %v\ndb+shm+wal: %d", size, (dbInfo.Size() + shmInfo.Size() + walInfo.Size()))
	}
}
