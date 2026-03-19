package utils_test

import (
	"goUp/utils"
	"log"
	"os"
	"strconv"
	"testing"
	"time"
)

// Returns 500 new valid test items
func getTestServiceData() []utils.ServiceData {
	svcData := []utils.ServiceData{}
	for i := range 500 {
		svcData = append(svcData, utils.ServiceData{
			ServiceName:         "example" + strconv.Itoa(i),
			ServiceHTTPResponse: "200",
			ServiceAPIResponse:  "",
			Timestamp:           time.Now(),
			Error:               false,
		})
	}
	return svcData
}

func TestGetDatabaseSize(t *testing.T) {
	ymlContent := `db_path: "./test_data.db"
services:
  service2:
    url: "https://example.com"
    retry: 2
  service3:
    url: "https://www.apple.com"
    retry: 2
`
	db, err := utils.InitDB()
	if err != nil {
		log.Fatalf("Failure to initialize database: %v", err)
	}

	initialSize, err := utils.GetDatabaseSize()
	if err != nil {
		log.Fatalf("Failed to get initial database size: %v", err)
	}

	cleanup := createTestYML(ymlContent, t)
	defer cleanup()
	defer func() {
		if err := os.Remove("./test_data.db"); err != nil {
			t.Errorf("Error removing test database file: %v", err)
		}
	}()

	cfg, err := utils.LoadConfig("services.yml")
	if err != nil {
		t.Fatalf("Failed to load configuration file: %v", err)
	}

	if err := utils.Setup(cfg); err != nil {
		t.Fatalf("Initial Setup() failed: %v", err)
	}

	testData := getTestServiceData()
	for i := range len(testData) {
		if err := utils.InsertData(db, testData[i]); err != nil {
			t.Fatalf("Failed to insert test data into the database")
		}
	}

	newSize, err := utils.GetDatabaseSize()
	if err := db.Close(); err != nil {
		t.Fatalf("Failed to shutdown the database %v: ", err)
	}
	if err := utils.CleanupDbFiles(); err != nil {
		t.Fatalf("Error cleaning up the database: %v", err)
	}
	if newSize <= initialSize {
		t.Fatal("New size is smaller or same as the initital size")
	}

}

func TestGetFilestampSize(t *testing.T) {
	before := time.Now()
	testFileName := "timestamp_test.txt"
	tmpFile, err := os.CreateTemp("", testFileName)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(testFileName); err != nil {
			t.Errorf("Failed to remove the test file: %v", err)
		}
	}()
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	stamp, err := utils.GetFileTimestamp(tmpFile.Name())
	if err != nil {
		t.Fatalf("GetFileStamp failed: %v", err)
	}
	after := time.Now()
	tolerance := 2 * time.Second
	if stamp.Before(before.Add(-tolerance)) || stamp.After(after.Add(tolerance)) {
		t.Fatal("Timestamp between start and file creation was 2 seconds +-")
	}
}
