package utils

import (
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var mu sync.RWMutex
var svcEndpoints ServiceEndpoints = ServiceEndpoints{Mux: &mu}
var transport = &http.Transport{
	IdleConnTimeout: 30 * time.Second,
	MaxIdleConns:    10,
}
var httpClient = &http.Client{
	Timeout:   30 * time.Second,
	Transport: transport,
}

type NoServiceEndpointsError struct {
	Message string
}

func (e *NoServiceEndpointsError) Error() string {
	return e.Message
}

// Sets up service and trigger endpoints from configuration
func Setup(cfg *Config) error {
	svcEndpoints.Mux.Lock()
	defer svcEndpoints.Mux.Unlock()

	Current_Config = cfg
	slog.Info("Setting up triggers")
	Current_Config.Triggers = *SetupTrigger(cfg)

	var updatedEndpoints []Service
	slog.Info("Setting up service endpoints")
	if cfg.Services != nil {
		updatedEndpoints = append(updatedEndpoints, scanDeadEndpoints(cfg)...)

		updatedEndpoints = append(updatedEndpoints, scanNewEndpoints(cfg, updatedEndpoints)...)
	} else {
		return &NoServiceEndpointsError{"No service endpoints found in current configuration!"}
	}
	svcEndpoints.ServiceEndpoint = updatedEndpoints
	slog.Info("Service endpoints setup finished")
	slog.Info("Configuration setup finished")
	return nil
}

func scanDeadEndpoints(cfg *Config) []Service {
	configServices := make(map[string]Service)
	// Sets each value in the map to the service currently in config, keyed by URL
	for _, svc := range cfg.Services {
		configServices[svc.URL] = svc
	}

	// Remove endpoints that are no longer in the config, and refresh the
	// ones that remain so field changes (e.g. Active) actually take effect.
	var updatedEndpoints []Service
	for _, endpoint := range svcEndpoints.ServiceEndpoint {
		if svc, found := configServices[endpoint.URL]; found {
			updatedEndpoints = append(updatedEndpoints, svc)
		} else {
			slog.Info("Removing endpoint from service endpoints", "endpoint", endpoint.Name)
		}
	}
	return updatedEndpoints
}

// Updates endpoints that arent found in the current config
func scanNewEndpoints(cfg *Config, endpoints []Service) []Service {
	var newEndpoints []Service
	for name, svc := range cfg.Services {
		found := false
		for _, endpoint := range endpoints {
			if endpoint.URL == svc.URL {
				found = true
				break
			}
		}
		if !found {
			slog.Info("Adding endpoint to service endpoints", "endpoint", name)
			svc.Name = name
			newEndpoints = append(newEndpoints, svc)
		}
	}
	return newEndpoints
}

// fetchOne performs the HTTP fetch (plus optional API fetch) for a single endpoint.
func fetchOne(endpoint Service) ServiceData {
	var sd ServiceData
	sd.Timestamp = time.Now()
	sd.ServiceName = endpoint.Name
	sd.ServiceURL = endpoint.URL
	sd.Active = true
	if endpoint.Description != nil {
		sd.ServiceDescription = *endpoint.Description
	}
	req, err := http.NewRequest("GET", endpoint.URL, nil)
	if err != nil {
		slog.Error("Error creating new HTTP request", "error", err)
	}
	req.Header.Add("User-Agent", "GoUp/"+Version)
	start := time.Now()

	res, err := httpClient.Do(req)
	if err != nil {
		res, err = ErrorRetry(&sd, endpoint, err)
		if err != nil {
			slog.Error("Error fetching endpoint", "url", endpoint.URL, "error", err)
			sd.ServiceHTTPResponse = err.Error()
			sd.ServiceResponseTime = time.Since(start).String()
			return sd
		}
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			slog.Error("Failed to close response body", "error", err)
		}
	}()

	sd.ServiceHTTPResponse = strconv.Itoa(res.StatusCode)
	if endpoint.API_URL != nil {
		if err := GetAPIData(endpoint, &sd, start); err != nil {
			slog.Error("Error getting API data", "endpoint", endpoint.Name, "error", err)
			sd.Error = true
			sd.ServiceAPIResponse = err.Error()
		}
	}
	if sd.ServiceResponseTime == "" {
		sd.ServiceResponseTime = time.Since(start).String()
	}
	return sd
}

// GetServiceData fetches all service endpoints concurrently and returns the results.
func GetServiceData() (*ServiceResponse, error) {
	if len(svcEndpoints.ServiceEndpoint) == 0 {
		slog.Info("No service endpoints found looking for config . . .")
		if err := Setup(Current_Config); err != nil {
			slog.Error("Error setting up config while fetching service data", "error", err)
			return nil, err
		}
	}

	endpoints := svcEndpoints.ServiceEndpoint
	results := make([]ServiceData, len(endpoints))
	var wg sync.WaitGroup

	slog.Info("Scanning services HTTP endpoints . . .")
	for i, ep := range endpoints {
		wg.Add(1)
		go func(i int, ep Service) {
			defer wg.Done()
			if ep.IsActive() {
				results[i] = fetchOne(ep)
			} else {
				slog.Info("Endpoint is disabled, continuing", "endpoint", ep.Name)
			}
		}(i, ep)
	}
	wg.Wait()

	var svcResponse ServiceResponse
	for _, r := range results {
		// Disabled endpoints are never fetched, leaving their slot as a
		// zero-value ServiceData{}; skip them so we don't insert blank
		// rows into the DB or have Check() flag them as down.
		if r.ServiceName == "" {
			continue
		}
		svcResponse.AllServices = append(svcResponse.AllServices, r)
	}
	var err error
	if svcResponse.DownServices, err = Check(svcResponse.AllServices); err != nil {
		return nil, err
	}
	return &svcResponse, nil
}

// Gets API data
func GetAPIData(endpoint Service, sd *ServiceData, start time.Time) error {
	apiReq, err := http.NewRequest("GET", *endpoint.API_URL, nil)
	if err != nil {
		return err
	}

	if endpoint.API_KEY != nil {
		apiReq.Header.Set("Authorization", "Bearer "+*endpoint.API_KEY)
	}

	apiReq.Header.Set("Content-Type", "application/json")
	apiReq.Header.Set("User-Agent", "GoUp/"+Version)

	apiRes, apiErr := httpClient.Do(apiReq)

	if apiErr != nil {
		slog.Error("Error occured during API request", "error", apiErr)
		return apiErr
	}

	defer func() {
		if err := apiRes.Body.Close(); err != nil {
			slog.Error("Error closing API response body", "error", err)
		}
	}()

	apiBody, err := io.ReadAll(apiRes.Body)
	if err != nil {
		return err
	}
	sd.ServiceAPIResponse = string(apiBody)
	elapsed := time.Since(start)
	sd.ServiceResponseTime = elapsed.String()
	slog.Info("Fetched API data successfully")
	return nil
}

// Returns current service endpoints for fetching
func GetServiceEndpoints() []Service {
	svcEndpoints.Mux.Lock()
	defer svcEndpoints.Mux.Unlock()

	out := make([]Service, len(svcEndpoints.ServiceEndpoint))
	copy(out, svcEndpoints.ServiceEndpoint)

	return out
}

// Sets service endpoints
func SetServiceEndpoints(validServices []Service) {
	svcEndpoints.Mux.Lock()
	defer svcEndpoints.Mux.Unlock()

	svcEndpoints.ServiceEndpoint = validServices
}

// Will retry the fetch request if configured (N-times)
// If not configured will return the initial error
func ErrorRetry(sd *ServiceData, endpoint Service, initialErr error) (*http.Response, error) {
	currentErr := initialErr
	// If it errors initially and the user has a configured number of retries
	// This will attempt to retry to get a new valid response the specified number of times
	if endpoint.Retry_Requests != nil {
		for range *endpoint.Retry_Requests {
			slog.Warn("Error fetching endpoint, re-attempting GET request", "url", endpoint.URL, "error", currentErr)
			retry_res, retry_err := httpClient.Get(endpoint.URL)
			if retry_err == nil {
				return retry_res, nil
			}
			currentErr = retry_err
		}
	}
	return nil, currentErr
}
