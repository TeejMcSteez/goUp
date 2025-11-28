package utils

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
)

var mu sync.RWMutex
var svcEndpoints ServiceEndpoints = ServiceEndpoints{Mux: &mu, ServiceEndpoint: make([]Service, 0)}

func GetServiceData(serviceEndpoints []Service) []ServiceData {
	var svcData []ServiceData

	d := make([]Service, len(serviceEndpoints))
	copy(d, serviceEndpoints)

	svcEndpoints.Mux.Lock()
	svcEndpoints.ServiceEndpoint = d
	svcEndpoints.Mux.Unlock()

	fmt.Println("Scanning services HTTP endpoints . . .")
	for i, endpoint := range d {
		res, err := http.Get(endpoint.URL)
		if err != nil {
			log.Printf("error fetching %s: %v %s", endpoint.URL, err, "❌")
			var sd ServiceData
			sd.ServiceName = endpoint.URL
			sd.ServiceHTTPResponse = err.Error()
			sd.ServiceAPIResponse = ""
			svcData = append(svcData, sd)
			continue
		}

		resType := res.StatusCode

		var sd ServiceData
		sd.ServiceName = d[i].URL
		sd.ServiceHTTPResponse = strconv.Itoa(resType)

		if resType == 200 {
			fmt.Println("Service ", d[i].URL, "responded with 200, Ok ️✅")
		} else {
			fmt.Println("response type was invalid: ", d[i], "->", resType, "❌")
		}
		if d[i].API_URL != nil {
			apiReq, err := http.NewRequest("GET", *endpoint.API_URL, nil)
			if err != nil {
				panic(err)
			}

			if d[i].API_KEY != nil {
				apiReq.Header.Set("Authorization", "Bearer "+*endpoint.API_KEY)
			}

			apiReq.Header.Set("Content-Type", "application/json")

			apiRes, apiErr := http.DefaultClient.Do(apiReq)

			if apiErr != nil {
				panic(apiErr)
			}

			defer apiRes.Body.Close()
			apiBody, err := io.ReadAll(apiRes.Body)
			if err != nil {
				panic(err)
			}
			sd.ServiceAPIResponse = string(apiBody)
		}
		svcData = append(svcData, sd)

	}
	return svcData
}

func GetServiceEndpoints() []Service {
	svcEndpoints.Mux.Lock()
	defer svcEndpoints.Mux.Unlock()

	out := make([]Service, len(svcEndpoints.ServiceEndpoint))
	copy(out, svcEndpoints.ServiceEndpoint)

	return out
}
