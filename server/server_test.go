package server_test

import (
	"database/sql"
	"encoding/json"
	"log/slog"
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
		Schedule:          &utils.ScheduleState{Span: 60, Interval: "seconds"},
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

	scd := scheduler.NewScheduler(db, cfg)
	val := "y"
	ptr := &val
	srv := server.NewServer(db, scd, ptr)

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

	var state utils.ScheduleState
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

// setupTestEnvWithConfig is like setupTestEnv but also writes a temp config file
// so handlers that call writeConfig (config mutations) don't fail on empty ConfigPath.
func setupTestEnvWithConfig(t *testing.T) (*sql.DB, *scheduler.Scheduler, *server.Server, func()) {
	t.Helper()
	dbPath := "./test_server_cfg.db"
	cfgFile, err := os.CreateTemp("", "goup-test-*.yml")
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	cfgPath := cfgFile.Name()
	if err := cfgFile.Close(); err != nil {
		slog.Error("Error occured closing config file", "error", err)
	}

	cfg := &utils.Config{
		Database_Location: &dbPath,
		ConfigPath:        cfgPath,
		Schedule:          &utils.ScheduleState{Span: 60, Interval: "seconds"},
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

	scd := scheduler.NewScheduler(db, cfg)
	val := "y"
	srv := server.NewServer(db, scd, &val)

	cleanup := func() {
		scd.Stop()
		if err := db.Close(); err != nil {
			slog.Error("Failed to close database connection", "error", err)
		}
		if err := os.Remove(dbPath); err != nil {
			slog.Error("Failed to remove database file", "error", err)
		}
		if err := os.Remove(cfgPath); err != nil {
			slog.Error("Failed to remove config file", "error", err)
		}
		utils.Current_Config = nil
		utils.SetServiceEndpoints([]utils.Service{})
	}
	return db, scd, srv, cleanup
}

// --- GetResponseTimes ---

func TestGetResponseTimes_ReturnsJSON(t *testing.T) {
	_, _, srv, cleanup := setupTestEnv(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/rt", nil)
	w := httptest.NewRecorder()

	srv.GetResponseTimes(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Expected application/json, got %s", ct)
	}
}

// --- UptimeAPI ---

func TestUptimeAPI_InvalidMethod(t *testing.T) {
	_, _, srv, cleanup := setupTestEnv(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/uptime", nil)
	w := httptest.NewRecorder()

	srv.UptimeAPI(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Result().StatusCode)
	}
}

// --- GetDatabasePersistence ---

func TestGetDatabasePersistence_Get(t *testing.T) {
	_, _, srv, cleanup := setupTestEnv(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/db/persist", nil)
	w := httptest.NewRecorder()

	srv.GetDatabasePersistence(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", res.StatusCode)
	}
	var val bool
	if err := json.NewDecoder(res.Body).Decode(&val); err != nil {
		t.Errorf("Failed to decode persistence response: %v", err)
	}
}

func TestGetDatabasePersistence_Post(t *testing.T) {
	_, _, srv, cleanup := setupTestEnvWithConfig(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/db/persist", nil)
	w := httptest.NewRecorder()

	srv.GetDatabasePersistence(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Result().StatusCode)
	}
}

func TestGetDatabasePersistence_InvalidMethod(t *testing.T) {
	_, _, srv, cleanup := setupTestEnv(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodDelete, "/api/db/persist", nil)
	w := httptest.NewRecorder()

	srv.GetDatabasePersistence(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Result().StatusCode)
	}
}

// --- ConfigServiceApi ---

func TestConfigServiceApi_Post_AddsService(t *testing.T) {
	_, _, srv, cleanup := setupTestEnvWithConfig(t)
	defer cleanup()

	body := `{"name":"new_svc","url":"http://new.example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/config/service", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.ConfigServiceApi(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Result().StatusCode)
	}
}

func TestConfigServiceApi_Post_InvalidJSON(t *testing.T) {
	_, _, srv, cleanup := setupTestEnvWithConfig(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/config/service", strings.NewReader("not json"))
	w := httptest.NewRecorder()

	srv.ConfigServiceApi(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Result().StatusCode)
	}
}

func TestConfigServiceApi_Put_UpdatesService(t *testing.T) {
	db, _, srv, cleanup := setupTestEnvWithConfig(t)
	defer cleanup()

	body := `{"old_name":"test_service","service":{"name":"test_service","url":"http://updated.example.com"}}`
	req := httptest.NewRequest(http.MethodPut, "/api/config/service", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.ConfigServiceApi(w, req)

	_ = db
	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Result().StatusCode)
	}
}

func TestConfigServiceApi_Delete_RemovesService(t *testing.T) {
	db, _, srv, cleanup := setupTestEnvWithConfig(t)
	defer cleanup()

	body := `{"name":"test_service","url":"http://example.com"}`
	req := httptest.NewRequest(http.MethodDelete, "/api/config/service", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.ConfigServiceApi(w, req)

	_ = db
	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Result().StatusCode)
	}
}

func TestConfigServiceApi_InvalidMethod(t *testing.T) {
	_, _, srv, cleanup := setupTestEnv(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/config/service", nil)
	w := httptest.NewRecorder()

	srv.ConfigServiceApi(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Result().StatusCode)
	}
}

// --- ConfigMQTTApi ---

func TestConfigMQTTApi_Post_SetsMQTT(t *testing.T) {
	_, _, srv, cleanup := setupTestEnvWithConfig(t)
	defer cleanup()

	broker := "tcp://broker.example.com:1883"
	user := "user"
	key := "pass"
	body := `{"mqtt_broker":"` + broker + `","mqtt_user":"` + user + `","mqtt_key":"` + key + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/config/mqtt", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.ConfigMQTTApi(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Result().StatusCode)
	}
}

func TestConfigMQTTApi_Delete_RemovesMQTT(t *testing.T) {
	_, _, srv, cleanup := setupTestEnvWithConfig(t)
	defer cleanup()

	// Seed MQTT config first so there is something to delete.
	postBody := `{"mqtt_broker":"tcp://broker.example.com:1883","mqtt_user":"u","mqtt_key":"k"}`
	postReq := httptest.NewRequest(http.MethodPost, "/api/config/mqtt", strings.NewReader(postBody))
	postReq.Header.Set("Content-Type", "application/json")
	srv.ConfigMQTTApi(httptest.NewRecorder(), postReq)

	req := httptest.NewRequest(http.MethodDelete, "/api/config/mqtt", nil)
	w := httptest.NewRecorder()
	srv.ConfigMQTTApi(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Result().StatusCode)
	}
}

func TestConfigMQTTApi_InvalidMethod(t *testing.T) {
	_, _, srv, cleanup := setupTestEnv(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/config/mqtt", nil)
	w := httptest.NewRecorder()

	srv.ConfigMQTTApi(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Result().StatusCode)
	}
}

// --- ConfigWebhookApi ---

func TestConfigWebhookApi_Post_SetsWebhook(t *testing.T) {
	_, _, srv, cleanup := setupTestEnvWithConfig(t)
	defer cleanup()

	body := `{"webhook_url":"http://hook.example.com","webhook_key":"Bearer token123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/config/webhook", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.ConfigWebhookApi(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Result().StatusCode)
	}
}

func TestConfigWebhookApi_Delete_RemovesWebhook(t *testing.T) {
	_, _, srv, cleanup := setupTestEnvWithConfig(t)
	defer cleanup()

	// Seed webhook config first so there is something to delete.
	postBody := `{"webhook_url":"http://hook.example.com","webhook_key":"Bearer token123"}`
	postReq := httptest.NewRequest(http.MethodPost, "/api/config/webhook", strings.NewReader(postBody))
	postReq.Header.Set("Content-Type", "application/json")
	srv.ConfigWebhookApi(httptest.NewRecorder(), postReq)

	req := httptest.NewRequest(http.MethodDelete, "/api/config/webhook", nil)
	w := httptest.NewRecorder()
	srv.ConfigWebhookApi(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Result().StatusCode)
	}
}

func TestConfigWebhookApi_InvalidMethod(t *testing.T) {
	_, _, srv, cleanup := setupTestEnv(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/config/webhook", nil)
	w := httptest.NewRecorder()

	srv.ConfigWebhookApi(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Result().StatusCode)
	}
}

// --- HandleNoUi ---

func TestHandleNoUi_ReturnsNotImplemented(t *testing.T) {
	_, _, srv, cleanup := setupTestEnv(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	srv.HandleNoUi(w, req)

	if w.Result().StatusCode != http.StatusNotImplemented {
		t.Errorf("Expected 501, got %d", w.Result().StatusCode)
	}
}
