package utils

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strconv"
	"sync"
	"time"
)

var mu sync.RWMutex
var svcEndpoints ServiceEndpoints = ServiceEndpoints{Mux: &mu}

// Sets up service and trigger endpoints from configuration
func Setup() {
	svcEndpoints.Mux.Lock()
	defer svcEndpoints.Mux.Unlock()

	fmt.Print("Loading Config . . .\n\n")
	cfg, err := LoadConfig("services.yml")

	fmt.Println("Setting up triggers")
	SetupTrigger(cfg)

	if err != nil {
		panic(err)
	}
	if cfg.Services == nil {
		panic("No services specified in the configuration file!")
	}
	for name, svc := range cfg.Services {
		if !slices.Contains(svcEndpoints.ServiceEndpoint, Service{URL: svc.URL}) {
			fmt.Println("Adding ", name, "to service endpoints")
			svcEndpoints.ServiceEndpoint = append(svcEndpoints.ServiceEndpoint, Service{URL: svc.URL, API_URL: svc.API_URL, API_KEY: svc.API_KEY})
		}
	}
}

// Gets service data from endpoints
// Also Checks the data before returning for any bad HTTP errors
func GetServiceData() []ServiceData {
	var svcData []ServiceData
	if len(svcEndpoints.ServiceEndpoint) == 0 {
		fmt.Println("No service endpoints found looking for config . . .")
		Setup()
	}

	fmt.Println("Scanning services HTTP endpoints . . .")
	for i, endpoint := range svcEndpoints.ServiceEndpoint {
		start := time.Now()
		res, err := http.Get(endpoint.URL)
		if err != nil {
			log.Printf("error fetching %s: %v %s", endpoint.URL, err, "❌")
			var sd ServiceData
			sd.ServiceName = endpoint.URL
			sd.ServiceHTTPResponse = err.Error()
			sd.ServiceAPIResponse = ""
			elapsed := time.Since(start)
			sd.ServiceResponseTime = elapsed.String()
			svcData = append(svcData, sd)
			continue
		}

		resType := res.StatusCode

		var sd ServiceData
		sd.ServiceName = svcEndpoints.ServiceEndpoint[i].URL
		sd.ServiceHTTPResponse = strconv.Itoa(resType)

		if resType == 200 {
			fmt.Println("Service ", svcEndpoints.ServiceEndpoint[i].URL, "responded with 200, Ok ️✅")
		} else {
			fmt.Println("response type was invalid: ", svcEndpoints.ServiceEndpoint[i], "->", resType, "❌")
		}
		if svcEndpoints.ServiceEndpoint[i].API_URL != nil {
			apiReq, err := http.NewRequest("GET", *endpoint.API_URL, nil)
			if err != nil {
				panic(err)
			}

			if svcEndpoints.ServiceEndpoint[i].API_KEY != nil {
				apiReq.Header.Set("Authorization", "Bearer "+*endpoint.API_KEY)
			}

			apiReq.Header.Set("Content-Type", "application/json")

			apiRes, apiErr := http.DefaultClient.Do(apiReq)

			if apiErr != nil {
				fmt.Println("Api Response Error")
				return svcData
			}

			defer func() {
				if err := apiRes.Body.Close(); err != nil {
					fmt.Println("Error closing body: " + err.Error())
				}
			}()

			apiBody, err := io.ReadAll(apiRes.Body)
			if err != nil {
				fmt.Println("Error reading response body")
				return svcData
			}
			sd.ServiceAPIResponse = string(apiBody)
		}
		elapsed := time.Since(start)
		fmt.Printf("Request took: %v\n", elapsed)
		sd.ServiceResponseTime = elapsed.String()
		svcData = append(svcData, sd)

	}
	Check(svcData)
	return svcData
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
func SetServiceEndpoints(svcs []Service) {
	svcEndpoints.Mux.Lock()
	defer svcEndpoints.Mux.Unlock()

	svcEndpoints.ServiceEndpoint = svcs
}
