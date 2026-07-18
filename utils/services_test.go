package utils_test

import (
	"goUp/utils"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

// This helper function creates a temporary services.yml for testing Setup()
func createTestYML(content string, t *testing.T) func() {
	// The path in Setup() is hardcoded to "services.yml".
	err := os.WriteFile("services.yml", []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test services.yml: %v", err)
	}
	return func() {
		if err := os.Remove("services.yml"); err != nil {
			t.Errorf("Failed to remove test config file: %v", err)
		}
	}
}

func TestSetupServiceChanges(t *testing.T) {
	ymlContent1 := `db_path: "./test_data.db"
services:
  service2:
    url: "https://example.com"
    retry: 2
  service3:
    url: "https://www.apple.com"
    retry: 2
`
	cleanup1 := createTestYML(ymlContent1, t)
	defer cleanup1()
	defer func() {
		if err := os.Remove("./test_data.db"); err != nil && !os.IsNotExist(err) {
			t.Errorf("Failed to remove test database: %v", err)
		}
	}()

	// Ensure endpoints are empty before setup
	utils.SetServiceEndpoints([]utils.Service{})

	cfg, err := utils.LoadConfig("services.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	err = utils.Setup(cfg)
	if err != nil {
		t.Fatalf("Initial Setup() failed: %v", err)
	}
	endpoints1 := utils.GetServiceEndpoints()
	if len(endpoints1) != 2 {
		t.Fatalf("Expected 2 endpoints after initial setup, got %d", len(endpoints1))
	}

	ymlContent2 := `db_path: "./test_data.db"
services:
  service2:
    url: "https://example.com"
    retry: 2
  service4:
    url: "https://www.google.com"
    retry: 1
`
	createTestYML(ymlContent2, t)

	cfg, err = utils.LoadConfig("services.yml")
	if err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}

	err = utils.Setup(cfg)
	if err != nil {
		t.Fatalf("Second Setup() failed: %v", err)
	}

	endpoints2 := utils.GetServiceEndpoints()
	if len(endpoints2) != 2 {
		t.Fatalf("Expected 2 endpoints after update, got %d. Endpoints: %+v", len(endpoints2), endpoints2)
	}

	foundService2 := false
	foundService4 := false
	for _, s := range endpoints2 {
		if s.Name == "service2" {
			foundService2 = true
		}
		if s.Name == "service4" {
			foundService4 = true
		}
	}

	if !foundService2 || !foundService4 {
		t.Errorf("Expected service2 and service4 to be present after update, but they were not. Endpoints: %+v", endpoints2)
	}
}

// TestSetupRefreshesExistingServiceFields ensures that when a service's URL is
// unchanged but its other fields change (e.g. toggling Active), Setup()
// refreshes the live endpoint list instead of keeping the stale cached copy.
func TestSetupRefreshesExistingServiceFields(t *testing.T) {
	ymlContent1 := `db_path: "./test_data.db"
services:
  service1:
    url: "https://example.com"
    retry: 2
`
	cleanup1 := createTestYML(ymlContent1, t)
	defer cleanup1()
	defer func() {
		if err := os.Remove("./test_data.db"); err != nil && !os.IsNotExist(err) {
			t.Errorf("Failed to remove test database: %v", err)
		}
	}()

	utils.SetServiceEndpoints([]utils.Service{})

	cfg, err := utils.LoadConfig("services.yml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if err := utils.Setup(cfg); err != nil {
		t.Fatalf("Initial Setup() failed: %v", err)
	}

	endpoints := utils.GetServiceEndpoints()
	if len(endpoints) != 1 || !endpoints[0].IsActive() {
		t.Fatalf("Expected 1 active endpoint after initial setup, got: %+v", endpoints)
	}

	// Same URL, but now explicitly disabled.
	ymlContent2 := `db_path: "./test_data.db"
services:
  service1:
    url: "https://example.com"
    retry: 2
    active: false
`
	createTestYML(ymlContent2, t)

	cfg, err = utils.LoadConfig("services.yml")
	if err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}
	if err := utils.Setup(cfg); err != nil {
		t.Fatalf("Second Setup() failed: %v", err)
	}

	endpoints = utils.GetServiceEndpoints()
	if len(endpoints) != 1 {
		t.Fatalf("Expected 1 endpoint after update, got %d. Endpoints: %+v", len(endpoints), endpoints)
	}
	if endpoints[0].IsActive() {
		t.Errorf("Expected service1 to be inactive after config update, but Setup() kept the stale cached endpoint. Endpoints: %+v", endpoints)
	}
}

func TestGetServiceData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api" {
			auth := r.Header.Get("Authorization")
			if auth != "Bearer test-key" {
				t.Errorf("Expected Bearer token, got %s", auth)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			if _, err := w.Write([]byte(`{"status": "ok"}`)); err != nil {
				t.Errorf("Failed to write status to server: %v", err)
			}
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	apiKey := "test-key"
	apiURL := server.URL + "/api"

	services := []utils.Service{
		{Name: "test-service", URL: server.URL, API_URL: &apiURL, API_KEY: &apiKey},
	}
	utils.SetServiceEndpoints(services)

	// Also need to set Current_Config for the Check function inside GetServiceData
	utils.Current_Config = &utils.Config{
		Services: map[string]utils.Service{
			"test-service": services[0],
		},
	}
	defer func() { utils.Current_Config = nil }()

	svcResponse, err := utils.GetServiceData()
	if err != nil {
		t.Fatalf("GetServiceData returned an error: %v", err)
	}

	if len(svcResponse.AllServices) != 1 {
		t.Fatalf("Expected 1 service data, got %d", len(svcResponse.AllServices))
	}

	sd := svcResponse.AllServices[0]
	if sd.ServiceName != "test-service" {
		t.Errorf("Expected service name 'test-service', got '%s'", sd.ServiceName)
	}
	if sd.ServiceHTTPResponse != "200" {
		t.Errorf("Expected HTTP response '200', got '%s'", sd.ServiceHTTPResponse)
	}
	if sd.ServiceAPIResponse != `{"status": "ok"}` {
		t.Errorf("Expected API response `{\"status\": \"ok\"}`, got '%s'", sd.ServiceAPIResponse)
	}
	if len(svcResponse.DownServices) != 0 {
		t.Errorf("Expected 0 down services, got %d", len(svcResponse.DownServices))
	}
}

func TestGetServiceDataRetry(t *testing.T) {
	retryCount := 0
	maxRetries := 2
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if retryCount < maxRetries {
			retryCount++
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	retryNum := maxRetries
	services := []utils.Service{
		{Name: "retry-service", URL: server.URL, Retry_Requests: &retryNum},
	}
	utils.SetServiceEndpoints(services)

	// Also need to set Current_Config for the Check function inside GetServiceData
	utils.Current_Config = &utils.Config{
		Services: map[string]utils.Service{
			"retry-service": services[0],
		},
	}
	defer func() { utils.Current_Config = nil }()

	svcResponse, err := utils.GetServiceData()
	if err != nil {
		t.Fatalf("GetServiceData returned an error: %v", err)
	}

	if len(svcResponse.AllServices) != 1 {
		t.Fatalf("Expected 1 service data, got %d", len(svcResponse.AllServices))
	}

}

// TestGetServiceDataValidResponses ensures that Check() correctly treats
// every status code in a service's configured Valid_Responses list as "up",
// no matter which one the endpoint happens to return on a given fetch, and
// still flags the service as down once the response changes to a code that
// isn't on that list.
func TestGetServiceDataValidResponses(t *testing.T) {
	var statusCode int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
	}))
	defer server.Close()

	validResponses := []string{"200", "201", "204", "301", "302"}
	service := utils.Service{Name: "multi-status-service", URL: server.URL, Valid_Responses: &validResponses}
	utils.SetServiceEndpoints([]utils.Service{service})

	utils.Current_Config = &utils.Config{
		Services: map[string]utils.Service{
			"multi-status-service": service,
		},
	}
	// Clear for other tests
	defer func() { utils.Current_Config = nil }()

	// Every configured status code should be reported as up, regardless of
	// which one the endpoint returns on a particular fetch.
	for _, code := range []int{http.StatusOK, http.StatusCreated, http.StatusNoContent, http.StatusMovedPermanently, http.StatusFound} {
		statusCode = code
		svcResponse, err := utils.GetServiceData()
		if err != nil {
			t.Fatalf("GetServiceData returned an error for status %d: %v", code, err)
		}
		if len(svcResponse.AllServices) != 1 {
			t.Fatalf("Expected 1 service data for status %d, got %d", code, len(svcResponse.AllServices))
		}
		got := svcResponse.AllServices[0].ServiceHTTPResponse
		want := strconv.Itoa(code)
		if got != want {
			t.Errorf("Expected HTTP response '%s', got '%s'", want, got)
		}
		if len(svcResponse.DownServices) != 0 {
			t.Errorf("Expected status %d to be treated as valid (0 down services), got %d down: %+v", code, len(svcResponse.DownServices), svcResponse.DownServices)
		}
	}

	// A response outside the configured list must be flagged as down.
	statusCode = http.StatusInternalServerError
	svcResponse, err := utils.GetServiceData()
	if err != nil {
		t.Fatalf("GetServiceData returned an error: %v", err)
	}
	if len(svcResponse.DownServices) != 1 {
		t.Fatalf("Expected service to be down for an unconfigured status code, got %d down services", len(svcResponse.DownServices))
	}
	if svcResponse.DownServices[0].ServiceName != "multi-status-service" {
		t.Errorf("Expected 'multi-status-service' to be marked down, got '%s'", svcResponse.DownServices[0].ServiceName)
	}
}
