package utils_test

import (
	"errors"
	"goUp/utils"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestEndpointsFollowConfig(t *testing.T) {
	cfg := &utils.Config{Services: map[string]utils.Service{
		"service3": {URL: "https://www.apple.com"},
		"service2": {URL: "https://example.com"},
	}}
	eps, err := utils.Endpoints(cfg)
	if err != nil {
		t.Fatalf("Endpoints() failed: %v", err)
	}
	// Name comes from the map key, order is sorted by name
	if len(eps) != 2 || eps[0].Name != "service2" || eps[1].Name != "service3" {
		t.Fatalf("Expected [service2 service3], got %+v", eps)
	}

	// A new snapshot fully replaces the list: no cached state carries over
	cfg = &utils.Config{Services: map[string]utils.Service{
		"service2": {URL: "https://example.com"},
		"service4": {URL: "https://www.google.com"},
	}}
	eps, err = utils.Endpoints(cfg)
	if err != nil {
		t.Fatalf("Endpoints() failed: %v", err)
	}
	if len(eps) != 2 || eps[0].Name != "service2" || eps[1].Name != "service4" {
		t.Fatalf("Expected [service2 service4], got %+v", eps)
	}
}

// TestEndpointsReflectFieldChanges ensures a field change on a service with
// an unchanged URL (e.g. toggling Active) shows up in the endpoint list.
func TestEndpointsReflectFieldChanges(t *testing.T) {
	inactive := false
	cfg := &utils.Config{Services: map[string]utils.Service{
		"service1": {URL: "https://example.com", Active: &inactive},
	}}
	eps, err := utils.Endpoints(cfg)
	if err != nil {
		t.Fatalf("Endpoints() failed: %v", err)
	}
	if len(eps) != 1 || eps[0].IsActive() {
		t.Fatalf("Expected 1 inactive endpoint, got %+v", eps)
	}
}

func TestEndpointsErrors(t *testing.T) {
	if _, err := utils.Endpoints(nil); err == nil {
		t.Error("Expected error for nil config")
	}
	var noSvc *utils.NoServiceEndpointsError
	if _, err := utils.Endpoints(&utils.Config{}); !errors.As(err, &noSvc) {
		t.Errorf("Expected NoServiceEndpointsError for nil Services, got %v", err)
	}
	eps, err := utils.Endpoints(&utils.Config{Services: map[string]utils.Service{}})
	if err != nil || len(eps) != 0 {
		t.Errorf("Expected empty list and no error for empty Services, got %v, %v", eps, err)
	}
}

// Disabled services are skipped: no fetch, no row, not reported down.
func TestGetServiceDataSkipsInactive(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	inactive := false
	cfg := &utils.Config{Services: map[string]utils.Service{
		"on":  {URL: server.URL},
		"off": {URL: server.URL + "/off", Active: &inactive},
	}}
	svcResponse, err := utils.NewScanner().GetServiceData(cfg)
	if err != nil {
		t.Fatalf("GetServiceData returned an error: %v", err)
	}
	if len(svcResponse.AllServices) != 1 || svcResponse.AllServices[0].ServiceName != "on" {
		t.Errorf("Expected only 'on' in results, got %+v", svcResponse.AllServices)
	}
	if hits != 1 {
		t.Errorf("Expected 1 request, got %d", hits)
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

	cfg := &utils.Config{
		Services: map[string]utils.Service{
			"test-service": services[0],
		},
	}

	svcResponse, err := utils.NewScanner().GetServiceData(cfg)
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

	cfg := &utils.Config{
		Services: map[string]utils.Service{
			"retry-service": services[0],
		},
	}

	svcResponse, err := utils.NewScanner().GetServiceData(cfg)
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

	cfg := &utils.Config{
		Services: map[string]utils.Service{
			"multi-status-service": service,
		},
	}

	// Every configured status code should be reported as up, regardless of
	// which one the endpoint returns on a particular fetch.
	for _, code := range []int{http.StatusOK, http.StatusCreated, http.StatusNoContent, http.StatusMovedPermanently, http.StatusFound} {
		statusCode = code
		svcResponse, err := utils.NewScanner().GetServiceData(cfg)
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
	svcResponse, err := utils.NewScanner().GetServiceData(cfg)
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
