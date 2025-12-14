package utils

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
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

	if err != nil {
		panic(err)
	}

	if cfg.Services == nil {
		panic("No services specified in the configuration file!")
	}

	fmt.Println("Setting up triggers")
	SetupTrigger(cfg)

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

func findValidFaviconEndpoints() []Service {
	validServices := []Service{}

	for _, endpoint := range svcEndpoints.ServiceEndpoint {
		u, err := url.Parse(endpoint.URL)
		if err != nil {
			fmt.Printf("Error parsing URL %s: %v\n", endpoint.URL, err)
			continue
		}

		assets := u.JoinPath("assets/favicon.ico")
		static := u.JoinPath("static/assets/favicon.ico")
		u = u.JoinPath("favicon.ico")
		
		if res, err := http.Get(assets.String()); err != nil {
			fmt.Printf("Error checking assets favicon endpoint: %v\n", err)
		} else {
			if res.StatusCode == 200 {
				validServices = append(validServices, Service{URL: assets.String()})
				continue
			}
		}

		if res, err := http.Get(static.String()); err != nil {
			fmt.Printf("Error checking static/assets favicon endpoint: %v\n", err)
		} else {
			if res.StatusCode == 200 {
				validServices = append(validServices, Service{URL: static.String()})
				continue
			}
		}

		if res, err := http.Get(u.String()); err != nil {
			fmt.Printf("Error checking basic favicon endpoint: %v\n", err)
		} else {
			if res.StatusCode == 200 {
				validServices = append(validServices, Service{URL: u.String()})
				continue
			}
		}
	}

	return validServices
}

func GetServiceFavicons() []FaviconData {
	// Need to change to test different endpoints
	// Some websites store there images at /assets/favicon.ico, /static/assets/favicon.png, etc.
	// Request most common and keep URL that responds with 200 and (maybe) valid image data
	var imageData []FaviconData
	
	validEndpoints := findValidFaviconEndpoints()

	if len(validEndpoints) == 0 {
		fmt.Println("No valid favicon endpoint found . . .")
		return []FaviconData{}
	}
	for _, endpoint := range validEndpoints {
		imageData = append(imageData, FaviconData{FaviconURL: endpoint.URL})
	}

	return imageData
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
