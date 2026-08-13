package utils_test

import (
	"database/sql"
	"goUp/utils"
	"testing"
	"time"
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

func TestGetPastUptimeFunctions(t *testing.T) {
	type uptimeFunc func(*sql.DB, string) (*float64, error)

	cases := []struct {
		name   string
		fn     uptimeFunc
		window time.Duration
	}{
		{"GetPastHourUptime", utils.GetPastHourUptime, time.Hour},
		{"GetPast12HourUptime", utils.GetPast12HourUptime, 12 * time.Hour},
		{"GetPastDayUptime", utils.GetPastDayUptime, 24 * time.Hour},
		{"GetPastWeekUptime", utils.GetPastWeekUptime, 168 * time.Hour},
		{"GetPastMonthUptime", utils.GetPastMonthUptime, 730 * time.Hour},
		{"GetPastYearUptime", utils.GetPastYearUptime, 8760 * time.Hour},
	}

	setupService := func(t *testing.T, serviceName string) *sql.DB {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		utils.Current_Config.Services = map[string]utils.Service{
			serviceName: {
				Name:            serviceName,
				URL:             "http://test.com",
				Valid_Responses: &[]string{"200"},
			},
		}
		return db
	}

	for _, tc := range cases {
		t.Run(tc.name+"/WithinWindow", func(t *testing.T) {
			serviceName := "svc_" + tc.name
			db := setupService(t, serviceName)

			now := time.Now()
			testData := []utils.ServiceData{
				// Within the window: 1 down out of 3
				{ServiceName: serviceName, ServiceHTTPResponse: "200", Timestamp: now.Add(-tc.window / 2)},
				{ServiceName: serviceName, ServiceHTTPResponse: "200", Timestamp: now.Add(-tc.window / 4)},
				{ServiceName: serviceName, ServiceHTTPResponse: "503", Timestamp: now.Add(-time.Minute)},
				// Outside the window: should be excluded from the calculation
				{ServiceName: serviceName, ServiceHTTPResponse: "503", Timestamp: now.Add(-tc.window - time.Hour)},
			}

			for _, sd := range testData {
				if err := utils.InsertData(db, sd); err != nil {
					t.Fatalf("Failed to insert data: %v", err)
				}
			}

			avg, err := tc.fn(db, serviceName)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if avg == nil {
				t.Fatal("Expected non-nil average, got nil")
			}

			expected := 0.33
			if *avg != expected {
				t.Errorf("Expected average %.2f, but got %f", expected, *avg)
			}
		})

		t.Run(tc.name+"/OutsideWindowOnly", func(t *testing.T) {
			serviceName := "svc_outside_" + tc.name
			db := setupService(t, serviceName)

			now := time.Now()
			sd := utils.ServiceData{
				ServiceName:         serviceName,
				ServiceHTTPResponse: "503",
				Timestamp:           now.Add(-tc.window - time.Hour),
			}
			if err := utils.InsertData(db, sd); err != nil {
				t.Fatalf("Failed to insert data: %v", err)
			}

			avg, err := tc.fn(db, serviceName)
			if err != nil {
				t.Errorf("Expected nil error, but got: %v", err)
			}
			if avg != nil {
				t.Errorf("Expected nil average when all data is outside the window, got %v", *avg)
			}
		})

		t.Run(tc.name+"/NoData", func(t *testing.T) {
			serviceName := "svc_empty_" + tc.name
			db := setupService(t, serviceName)

			avg, err := tc.fn(db, serviceName)
			if err != nil {
				t.Errorf("Expected nil error, but got: %v", err)
			}
			if avg != nil {
				t.Errorf("Expected nil average when there is no data, got %v", *avg)
			}
		})
	}
}
