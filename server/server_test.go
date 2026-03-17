package server_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"goUp/server"
	"goUp/utils"
	scheduler "goUp/workers"
)

func setupTestEnv(t *testing.T) (*sql.DB, *scheduler.Scheduler, *server.Server, func()) {
	t.Helper()
	dbPath := "./test_server.db"
	cfg := &utils.Config{
		Database_Location: &dbPath,
		Services: map[string]utils.Service{
			"test_service": {Name: "test_service", URL: "http://example.com"},
		},
	}
	utils.Current_Config = cfg
	utils.SetServiceEndpoints([]utils.Service{
		{Name: "test_service", URL: "http://example.com"},
	})

	db, err := utils.InitDB()
	if err != nil {
		t.Fatalf("Failed to init test DB: %v", err)
	}

	scd := scheduler.NewScheduler(db, 60, "seconds")
	srv := server.NewServer(db, scd)

	cleanup := func() {
		scd.Stop()
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database on cleanup: %v", err)
		}
		if err := os.Remove(dbPath); err != nil {
			t.Errorf("Failed to remove database file on cleanup: %v", err)
		}
		utils.Current_Config = nil
		utils.SetServiceEndpoints([]utils.Service{})
	}

	return db, scd, srv, cleanup
}

func TestApiHandler_ReturnsJSON(t *testing.T) {
	_, _, srv, cleanup := setupTestEnv(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	w := httptest.NewRecorder()

	srv.Api(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", ct)
	}

	var data []utils.ServiceData
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		t.Errorf("Failed to decode response body as JSON: %v", err)
	}
}

func TestScheduleApiGet_ReturnsState(t *testing.T) {
	_, _, srv, cleanup := setupTestEnv(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/schedule", nil)
	w := httptest.NewRecorder()

	srv.ScheduleApi(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", res.StatusCode)
	}

	var state scheduler.ScheduleState
	if err := json.NewDecoder(res.Body).Decode(&state); err != nil {
		t.Errorf("Failed to decode schedule state: %v", err)
	}
	if state.Span != 60 || state.Interval != "seconds" {
		t.Errorf("Unexpected schedule state: %+v", state)
	}
}

func TestScheduleApiPost_ValidUpdate(t *testing.T) {
	_, _, srv, cleanup := setupTestEnv(t)
	defer cleanup()

	body := `{"timespan": 30, "interval": "minutes"}`
	req := httptest.NewRequest(http.MethodPost, "/api/schedule", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.ScheduleApi(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", res.StatusCode)
	}

	var resp map[string]bool
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}
	if !resp["updated"] {
		t.Error("Expected updated to be true")
	}
}

func TestScheduleApiPost_InvalidSpan(t *testing.T) {
	_, _, srv, cleanup := setupTestEnv(t)
	defer cleanup()

	// Span of 0 is invalid (must be 1-60)
	body := `{"timespan": 0, "interval": "minutes"}`
	req := httptest.NewRequest(http.MethodPost, "/api/schedule", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.ScheduleApi(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", res.StatusCode)
	}
}

func TestScheduleApiPost_InvalidJSON(t *testing.T) {
	_, _, srv, cleanup := setupTestEnv(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/schedule", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.ScheduleApi(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500 for bad JSON, got %d", res.StatusCode)
	}
}

func TestScheduleApi_InvalidMethod(t *testing.T) {
	_, _, srv, cleanup := setupTestEnv(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPut, "/api/schedule", nil)
	w := httptest.NewRecorder()

	srv.ScheduleApi(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", res.StatusCode)
	}
}

func TestStatusApiGet_ReturnsJSON(t *testing.T) {
	_, _, srv, cleanup := setupTestEnv(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	w := httptest.NewRecorder()

	srv.StatusApi(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", ct)
	}
}

func TestStatusApiPost_BadRequest(t *testing.T) {
	_, _, srv, cleanup := setupTestEnv(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/status", nil)
	w := httptest.NewRecorder()

	srv.StatusApi(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", res.StatusCode)
	}
}

func TestUptimeAPIGet_ReturnsJSON(t *testing.T) {
	_, _, srv, cleanup := setupTestEnv(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/uptime", nil)
	w := httptest.NewRecorder()

	srv.UptimeAPI(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", res.StatusCode)
	}

	var data []utils.AverageData
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		t.Errorf("Failed to decode uptime response: %v", err)
	}
}

func TestGetErrorData_ReturnsJSON(t *testing.T) {
	db, _, srv, cleanup := setupTestEnv(t)
	defer cleanup()

	// Insert an error record
	_ = utils.InsertData(db, utils.ServiceData{
		ServiceName:         "test_service",
		ServiceHTTPResponse: "500",
		Error:               true,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/errors?limit=10&sort=desc", nil)
	w := httptest.NewRecorder()

	srv.GetErrorData(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", res.StatusCode)
	}

	var data []utils.ServiceData
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		t.Errorf("Failed to decode error data response: %v", err)
	}
	if len(data) != 1 {
		t.Errorf("Expected 1 error record, got %d", len(data))
	}
}

func TestGetErrorData_NoLimit(t *testing.T) {
	_, _, srv, cleanup := setupTestEnv(t)
	defer cleanup()

	// No limit param — handler should default to 0 (no limit)
	req := httptest.NewRequest(http.MethodGet, "/api/errors", nil)
	w := httptest.NewRecorder()

	srv.GetErrorData(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", res.StatusCode)
	}
}

func TestGetDatabaseSize_ReturnsSize(t *testing.T) {
	_, _, srv, cleanup := setupTestEnv(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/db/size", nil)
	w := httptest.NewRecorder()

	srv.GetDatabaseSize(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", res.StatusCode)
	}

	var payload utils.DatabaseSizePayload
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Errorf("Failed to decode database size response: %v", err)
	}
	if payload.Size <= 0 {
		t.Errorf("Expected positive DB size, got %d", payload.Size)
	}
}

func TestClearDatabase_ReturnsOK(t *testing.T) {
	db, _, srv, cleanup := setupTestEnv(t)
	defer cleanup()

	_ = utils.InsertData(db, utils.ServiceData{
		ServiceName:         "test_service",
		ServiceHTTPResponse: "200",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/db/clear", nil)
	w := httptest.NewRecorder()

	srv.ClearDatabase(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", res.StatusCode)
	}

	var resp map[string]bool
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		t.Errorf("Failed to decode clear database response: %v", err)
	}
	if !resp["ok"] {
		t.Error("Expected ok to be true")
	}
}

func TestReadConfigData_ReturnsConfig(t *testing.T) {
	_, _, srv, cleanup := setupTestEnv(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	w := httptest.NewRecorder()

	srv.ReadConfigData(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", res.StatusCode)
	}

	var data utils.ConfigData
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		t.Errorf("Failed to decode config data response: %v", err)
	}
	if _, ok := data.Services["test_service"]; !ok {
		t.Error("Expected test_service in config services")
	}
}

func TestNewServer_NotNil(t *testing.T) {
	_, _, srv, cleanup := setupTestEnv(t)
	defer cleanup()

	if srv == nil {
		t.Error("Expected NewServer to return a non-nil server")
	}
}
