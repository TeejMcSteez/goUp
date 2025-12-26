package utils

import (
	"fmt"
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

	fmt.Print("Loading Config . . .\n\n")
	cfg, err := LoadConfig("services.yml")

	if err != nil {
		return err
	}

	if cfg.Services == nil {
		return err
	}

	Current_Config = cfg
	fmt.Println("Configuration setup finished")
	fmt.Println("Setting up triggers")
	SetupTrigger(cfg)

	for name, svc := range cfg.Services {
		found := false
		for _, es := range svcEndpoints.ServiceEndpoint {
			if es.URL == svc.URL {
				found = true
				break
			}
		}
		if !found {
			fmt.Println("Adding ", name, "to service endpoints")
			svc.Name = name
			svcEndpoints.ServiceEndpoint = append(svcEndpoints.ServiceEndpoint, svc)
		}
	}
	return nil
}

// Gets service data from endpoints
// Also Checks the data before returning for any bad HTTP errors
func GetServiceData() ServiceResponse {
	var svcResponse ServiceResponse
	if len(svcEndpoints.ServiceEndpoint) == 0 {
		fmt.Println("No service endpoints found looking for config . . .")
		err := Setup()
		if err != nil {
			log.Fatalf("Error in in setting up config while fetching service data!\n%v", err)
		}
	}

	fmt.Println("Scanning services HTTP endpoints . . .")
	for _, endpoint := range svcEndpoints.ServiceEndpoint {
		start := time.Now()
		res, err := http.Get(endpoint.URL)
		if err != nil {
			log.Printf("error fetching %s: %v %s", endpoint.URL, err, "❌")
			var sd ServiceData
			sd.ServiceName = endpoint.Name
			sd.ServiceHTTPResponse = "Fetch Error, Check Logs"
			sd.ServiceAPIResponse = ""
			sd.ServiceResponseTime = time.Since(start).String()
			svcResponse.AllServices = append(svcResponse.AllServices, sd)
			continue
		}

		resType := res.StatusCode

		var sd ServiceData
		sd.ServiceName = endpoint.Name
		sd.ServiceHTTPResponse = strconv.Itoa(resType)

		if endpoint.API_URL != nil {
			apiReq, err := http.NewRequest("GET", *endpoint.API_URL, nil)
			if err != nil {
				panic(err)
			}

			if endpoint.API_KEY != nil {
				apiReq.Header.Set("Authorization", "Bearer "+*endpoint.API_KEY)
			}

			apiReq.Header.Set("Content-Type", "application/json")

			apiRes, apiErr := http.DefaultClient.Do(apiReq)

			if apiErr != nil {
				fmt.Println("Api Response Error")
			}

			defer func() {
				if err := apiRes.Body.Close(); err != nil {
					fmt.Println("Error closing body: " + err.Error())
				}
			}()

			apiBody, err := io.ReadAll(apiRes.Body)
			if err != nil {
				fmt.Println("Error reading response body")
			}
			sd.ServiceAPIResponse = string(apiBody)
		}
		elapsed := time.Since(start)
		// fmt.Printf("Request took: %v\n", elapsed)
		sd.ServiceResponseTime = elapsed.String()
		svcResponse.AllServices = append(svcResponse.AllServices, sd)

	}
	if s, err := Check(svcResponse.AllServices); err != nil {
		log.Fatalf("Error occured while getting service data: %v", err)
	} else {
		svcResponse.DownServices = s
	}
	return svcResponse
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
