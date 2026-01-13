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
var Current_Config *Config

// Sets up service and trigger endpoints from configuration
func Setup() error {
	svcEndpoints.Mux.Lock()
	defer svcEndpoints.Mux.Unlock()

	log.Println("Loading Config . . .")
	cfg, err := LoadConfig("services.yml")

	if err != nil {
		return err
	}

	Current_Config = cfg
	log.Println("Configuration setup finished")
	log.Println("Setting up triggers")
	SetupTrigger(cfg)

	configServices := make(map[string]struct{})
	// Sets each value in the map to the URL of the service
	if cfg.Services != nil {
		for _, svc := range cfg.Services {
			configServices[svc.URL] = struct{}{}
		}
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

	// Add new endpoints from the config.
	if cfg.Services != nil {
		updatedEndpoints = append(updatedEndpoints, UpdateEndoints(cfg, updatedEndpoints)...)
	}

	svcEndpoints.ServiceEndpoint = updatedEndpoints

	return nil
}
// Updates endpoints that arent found in the current config
func UpdateEndoints(cfg *Config, endpoints []Service) []Service {
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

// Gets service data from endpoints
// Also Checks the data before returning for any bad HTTP errors
func GetServiceData() (data *ServiceResponse, retErr error) {
	var svcResponse ServiceResponse
	if len(svcEndpoints.ServiceEndpoint) == 0 {
		log.Println("No service endpoints found looking for config . . .")
		err := Setup()
		if err != nil {
			log.Printf("Error in in setting up config while fetching service data!\n%v", err)
			return nil, err
		}
	}

	log.Println("Scanning services HTTP endpoints . . .")
	for _, endpoint := range svcEndpoints.ServiceEndpoint {
		var sd ServiceData
		sd.ServiceName = endpoint.Name
		start := time.Now()
		res, err := http.Get(endpoint.URL)

		if err != nil {
			res, err = ErrorRetry(&sd, endpoint, err)
			// If all retries fail, err will still be non-nil
			if err != nil {
				log.Printf("Error fetching %s: %v %s", endpoint.URL, err, "❌")
				sd.ServiceHTTPResponse = err.Error()
				sd.ServiceAPIResponse = ""
				sd.ServiceResponseTime = time.Since(start).String()
				svcResponse.AllServices = append(svcResponse.AllServices, sd)
				continue
			}
		}
		defer res.Body.Close()

		resType := res.StatusCode
		sd.ServiceHTTPResponse = strconv.Itoa(resType)
		// If their is an API URL will attempt to get data
		if endpoint.API_URL != nil {
			err := GetAPIData(endpoint, &sd, start)
			if err != nil {
				log.Printf("Error getting API data for %s: %v", endpoint.Name, err)
				sd.Error = true
				sd.ServiceAPIResponse = err.Error()
			}
		}

		if sd.ServiceResponseTime == "" {
			sd.ServiceResponseTime = time.Since(start).String()
		}

		svcResponse.AllServices = append(svcResponse.AllServices, sd)
	}
	if svcResponse.DownServices, retErr = Check(svcResponse.AllServices); retErr != nil {
		return nil, retErr
	}

	return &svcResponse, retErr
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

	apiRes, apiErr := http.DefaultClient.Do(apiReq)

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

// Sets service endpoints (for later frontend use)
func SetServiceEndpoints(validServices []Service) {
	svcEndpoints.Mux.Lock()
	defer svcEndpoints.Mux.Unlock()

	svcEndpoints.ServiceEndpoint = validServices
}
// Will retry the fetch request if configured (N-times)
// If not configured will return the initial error
func ErrorRetry(sd *ServiceData, endpoint Service, initialErr error) (*http.Response, error) {
	sd.Error = true
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
