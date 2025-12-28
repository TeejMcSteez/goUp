package utils

import (
	"log"
	"io"
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
		for name, svc := range cfg.Services {
			found := false
			for _, endpoint := range updatedEndpoints {
				if endpoint.URL == svc.URL {
					found = true
					break
				}
			}
			if !found {
				log.Println("Adding", name, "to service endpoints")
				svc.Name = name
				updatedEndpoints = append(updatedEndpoints, svc)
			}
		}
	}

	svcEndpoints.ServiceEndpoint = updatedEndpoints

	return nil
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
		start := time.Now()
		res, err := http.Get(endpoint.URL)
 
		if err != nil {
			// If it errors initially and the user has a configured number of retries
			// This will attempt to retry to get a new valid response the specified number of times
			if endpoint.Retry_Requests != nil {
				for range *endpoint.Retry_Requests {
					log.Printf("Error fetching %s: %v %s\nRe-attempting GET request.", endpoint.URL, err, "❌")
					retry_res, _ := http.Get(endpoint.URL)
					res = retry_res
				}
			// If their are no retries configured it will simply return
			} else {
				log.Printf("Error fetching %s: %v %s", endpoint.URL, err, "❌")
				var sd ServiceData
				sd.ServiceName = endpoint.Name
				sd.ServiceHTTPResponse = "Fetch Error, Check Logs"
				sd.ServiceAPIResponse = ""
				sd.ServiceResponseTime = time.Since(start).String()
				svcResponse.AllServices = append(svcResponse.AllServices, sd)
				continue
			}
		}

		resType := res.StatusCode

		var sd ServiceData
		sd.ServiceName = endpoint.Name
		sd.ServiceHTTPResponse = strconv.Itoa(resType)

		if endpoint.API_URL != nil {
			apiReq, err := http.NewRequest("GET", *endpoint.API_URL, nil)
			if err != nil {
				return nil, err
			}

			if endpoint.API_KEY != nil {
				apiReq.Header.Set("Authorization", "Bearer "+*endpoint.API_KEY)
			}

			apiReq.Header.Set("Content-Type", "application/json")

			apiRes, apiErr := http.DefaultClient.Do(apiReq)

			if apiErr != nil {
				return nil, err
			}

			defer func() {
				if err := apiRes.Body.Close(); err != nil {
					retErr = err
				}
			}()

			apiBody, err := io.ReadAll(apiRes.Body)
			if err != nil {
				return nil, err
			}
			sd.ServiceAPIResponse = string(apiBody)
		}
		elapsed := time.Since(start)
		// log.Printf("Request took: %v\n", elapsed)
		sd.ServiceResponseTime = elapsed.String()
		svcResponse.AllServices = append(svcResponse.AllServices, sd)

	}
	if s, err := Check(svcResponse.AllServices); err != nil {
		log.Printf("Error occured while checking service data: %v", err)
		return nil, err
	} else {
		svcResponse.DownServices = s
	}
	data = &svcResponse
	return data, retErr
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
