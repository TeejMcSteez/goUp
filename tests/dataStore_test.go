package utils_test

import (
	"database/sql"
	"goUp/utils"
	"os"
	"testing"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	dbPath := "./test_db.db"
	cfg := &utils.Config{
		Database_Location: &dbPath,
	}
	utils.Current_Config = cfg

	db, err := utils.InitDB()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
		utils.Current_Config = nil
	}

	return db, cleanup
}

func TestInitDBNoConfig(t *testing.T) {
	utils.Current_Config = nil
	_, err := utils.InitDB()
	if err == nil {
		t.Error("Expected error when initializing DB with nil config, but got nil")
	}
}

func TestInitDBNoDBLocation(t *testing.T) {
	utils.Current_Config = &utils.Config{}
	_, err := utils.InitDB()
	if err == nil {
		t.Error("Expected error when initializing DB with no db_path in config, but got nil")
	}
	utils.Current_Config = nil
}

func TestInsertAndGetData(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	testData := []utils.ServiceData{
		{ServiceName: "service1", ServiceHTTPResponse: "200", Error: false},
		{ServiceName: "service2", ServiceHTTPResponse: "500", Error: true},
	}

	for _, sd := range testData {
		err := utils.InsertData(db, sd)
		if err != nil {
			t.Fatalf("Failed to insert data: %v", err)
		}
	}

	retrievedData, err := utils.GetData(db)
	if err != nil {
		t.Fatalf("Failed to get data: %v", err)
	}

	if len(retrievedData) != len(testData) {
		t.Errorf("Expected %d rows, got %d", len(testData), len(retrievedData))
	}

	// This assumes order is preserved, which is true for this simple case.
	for i, sd := range retrievedData {
		if sd.ServiceName != testData[i].ServiceName || sd.ServiceHTTPResponse != testData[i].ServiceHTTPResponse || sd.Error != testData[i].Error {
			t.Errorf("Mismatch in retrieved data. Expected %+v, got %+v", testData[i], sd)
		}
	}
}

func TestGetDataForService(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	testData := []utils.ServiceData{
		{ServiceName: "service1", ServiceHTTPResponse: "200", Error: false},
		{ServiceName: "service2", ServiceHTTPResponse: "500", Error: true},
		{ServiceName: "service1", ServiceHTTPResponse: "201", Error: false},
	}

	for _, sd := range testData {
		err := utils.InsertData(db, sd)
		if err != nil {
			t.Fatalf("Failed to insert data: %v", err)
		}
	}

	retrievedData, err := utils.GetDataForService(db, "service1")
	if err != nil {
		t.Fatalf("Failed to get data for service: %v", err)
	}

	if len(retrievedData) != 2 {
		t.Errorf("Expected 2 rows for service1, got %d", len(retrievedData))
	}

	for _, sd := range retrievedData {
		if sd.ServiceName != "service1" {
			t.Errorf("Expected serviceName to be 'service1', got '%s'", sd.ServiceName)
		}
	}
}

func TestGetRecentData(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// SetServiceEndpoints will be used to determine the limit.
	utils.SetServiceEndpoints([]utils.Service{
		{Name: "service1", URL: "http://service1.com"},
		{Name: "service2", URL: "http://service2.com"},
	})

	allTestData := []utils.ServiceData{
		{ServiceName: "service1", ServiceHTTPResponse: "200"},
		{ServiceName: "service2", ServiceHTTPResponse: "200"},
		{ServiceName: "service1", ServiceHTTPResponse: "500"},
		{ServiceName: "service2", ServiceHTTPResponse: "503"},
	}

	for _, sd := range allTestData {
		err := utils.InsertData(db, sd)
		if err != nil {
			t.Fatalf("Failed to insert data: %v", err)
		}
	}

	recentData, err := utils.GetRecentData(db)
	if err != nil {
		t.Fatalf("Failed to get recent data: %v", err)
	}

	numOfServices := len(utils.GetServiceEndpoints())
	if len(recentData) != numOfServices {
		t.Errorf("Expected %d recent data points, got %d", numOfServices, len(recentData))
	}

	// It should return the last two inserted records in descending order of ID.
	if recentData[0].ServiceName != "service2" || recentData[0].ServiceHTTPResponse != "503" {
		t.Errorf("Unexpected recent data[0]: %+v, expected service2 with 503", recentData[0])
	}
	if recentData[1].ServiceName != "service1" || recentData[1].ServiceHTTPResponse != "500" {
		t.Errorf("Unexpected recent data[1]: %+v, expected service1 with 500", recentData[1])
	}
}
