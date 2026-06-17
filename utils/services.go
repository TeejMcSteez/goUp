package utils

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var mu sync.RWMutex
var svcEndpoints ServiceEndpoints = ServiceEndpoints{Mux: &mu}
var httpClient = &http.Client{
	Timeout: 30 * time.Second,
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
	log.Println("Setting up triggers")
	SetupTrigger(cfg)

	var updatedEndpoints []Service
	log.Println("Setting up service endpoints")
	if cfg.Services != nil {
		updatedEndpoints = append(updatedEndpoints, scanDeadEndpoints(cfg)...)

		updatedEndpoints = append(updatedEndpoints, scanNewEndpoints(cfg, updatedEndpoints)...)
	} else {
		return &NoServiceEndpointsError{"No service endpoints found in current configuration!"}
	}
	svcEndpoints.ServiceEndpoint = updatedEndpoints
	log.Println("Service endpoints setup finished")
	log.Println("Configuration setup finished")
	return nil
}

func scanDeadEndpoints(cfg *Config) []Service {
	configServices := make(map[string]struct{})
	// Sets each value in the map to the URL of the service
	for _, svc := range cfg.Services {
		configServices[svc.URL] = struct{}{}
	}

	// Remove endpoints that are no longer in the config.
	var updatedEndpoints []Service
	for _, endpoint := range svcEndpoints.ServiceEndpoint {
		if _, found := configServices[endpoint.URL]; found {
			updatedEndpoints = append(updatedEndpoints, endpoint)
		} else {
			log.Println("Removing", endpoint.Name, "from service endpoints")
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
			log.Println("Adding", name, "to service endpoints")
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
	if endpoint.Description != nil {
		sd.ServiceDescription = *endpoint.Description
	}
	req, err := http.NewRequest("GET", endpoint.URL, nil)
	req.Header.Add("User-Agent", "GoUp/"+Version)
	start := time.Now()

	res, err := httpClient.Do(req)
	if err != nil {
		res, err = ErrorRetry(&sd, endpoint, err)
		if err != nil {
			log.Printf("Error fetching %s: %v %s", endpoint.URL, err, "❌")
			sd.ServiceHTTPResponse = err.Error()
			sd.ServiceResponseTime = time.Since(start).String()
			return sd
		}
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("Failed to close response body: %v", err)
		}
	}()

	sd.ServiceHTTPResponse = strconv.Itoa(res.StatusCode)
	if endpoint.API_URL != nil {
		if err := GetAPIData(endpoint, &sd, start); err != nil {
			log.Printf("Error getting API data for %s: %v", endpoint.Name, err)
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
		log.Println("No service endpoints found looking for config . . .")
		if err := Setup(Current_Config); err != nil {
			log.Printf("Error setting up config while fetching service data!\n%v", err)
			return nil, err
		}
	}

	endpoints := svcEndpoints.ServiceEndpoint
	results := make([]ServiceData, len(endpoints))
	var wg sync.WaitGroup

	log.Println("Scanning services HTTP endpoints . . .")
	for i, ep := range endpoints {
		wg.Add(1)
		go func(i int, ep Service) {
			defer wg.Done()
			results[i] = fetchOne(ep)
		}(i, ep)
	}
	wg.Wait()

	var svcResponse ServiceResponse
	svcResponse.AllServices = results
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
		log.Printf("Error occured during API request: %v\n", apiErr)
		return apiErr
	}

	defer func() {
		if err := apiRes.Body.Close(); err != nil {
			log.Printf("Error closing API response body: %v\n", err)
		}
	}()

	apiBody, err := io.ReadAll(apiRes.Body)
	if err != nil {
		return err
	}
	sd.ServiceAPIResponse = string(apiBody)
	elapsed := time.Since(start)
	sd.ServiceResponseTime = elapsed.String()
	log.Print("Fetched API data successfully\n")
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
			log.Printf("Error fetching %s: %v %s\nRe-attempting GET request.", endpoint.URL, currentErr, "❌")
			retry_res, retry_err := http.Get(endpoint.URL)
			if retry_err == nil {
				return retry_res, nil
			}
			currentErr = retry_err
		}
	}
	return nil, currentErr
}
