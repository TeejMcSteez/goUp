package utils_test

import (
	"goUp/utils"
	"testing"
)

func TestServiceCheck(t *testing.T) {
	cfg := &utils.Config{
		Services: map[string]utils.Service{
			"test1": {
				Name:            "test1",
				URL:             "http://test1.com",
				Valid_Responses: &[]string{"200", "201"},
			},
			"test2": {
				Name:            "test2",
				URL:             "http://test2.com",
				Valid_Responses: &[]string{"200"},
			},
		},
	}
	utils.Current_Config = cfg
	defer func() {
		utils.Current_Config = nil
	}()

	sd := []utils.ServiceData{
		{ServiceName: "test1", ServiceHTTPResponse: "200", ServiceAPIResponse: "", ServiceResponseTime: ""},
		{ServiceName: "test2", ServiceHTTPResponse: "303", ServiceAPIResponse: "", ServiceResponseTime: ""},
	}
	res, err := utils.Check(sd)

	if err != nil {
		t.Errorf("Error checking data: %v", err)
	}

	if len(res) != 1 {
		t.Errorf("Expected 1 error, but got %d", len(res))
	}
}

func TestUptimeRounding(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	utils.Current_Config.Services = map[string]utils.Service{
		"roundService": {
			Name:            "roundService",
			URL:             "http://test.com",
			Valid_Responses: &[]string{"200"},
		},
	}

	// 1 down out of 3 = 0.3333... should round to 0.33
	testData := []utils.ServiceData{
		{ServiceName: "roundService", ServiceHTTPResponse: "200"},
		{ServiceName: "roundService", ServiceHTTPResponse: "200"},
		{ServiceName: "roundService", ServiceHTTPResponse: "503"},
	}

	for _, sd := range testData {
		if err := utils.InsertData(db, sd); err != nil {
			t.Fatalf("Failed to insert data: %v", err)
		}
	}

	avg, err := utils.GetUptimeAverage(db, "roundService")
	if err != nil {
		t.Fatalf("Error getting uptime average: %v", err)
	}

	expected := 0.33
	if *avg != expected {
		t.Errorf("Expected average to be %.2f, but got %f", expected, *avg)
	}
}

func TestUptimeCalculation(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	utils.Current_Config.Services = map[string]utils.Service{
		"testService": {
			Name:            "testService",
			URL:             "http://test.com",
			Valid_Responses: &[]string{"200"},
		},
	}

	testData := []utils.ServiceData{
		{ServiceName: "testService", ServiceHTTPResponse: "200"},
		{ServiceName: "testService", ServiceHTTPResponse: "200"},
		{ServiceName: "testService", ServiceHTTPResponse: "503"}, // down
		{ServiceName: "testService", ServiceHTTPResponse: "200"},
	}

	for _, sd := range testData {
		err := utils.InsertData(db, sd)
		if err != nil {
			t.Fatalf("Failed to insert data: %v", err)
		}
	}

	avg, err := utils.GetUptimeAverage(db, "testService")
	if err != nil {
		t.Fatalf("Error getting uptime average: %v", err)
	}

	expectedAvg := 0.25
	if *avg != expectedAvg {
		t.Errorf("Expected average to be %f, but got %f", expectedAvg, *avg)
	}
}
