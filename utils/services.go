package utils

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
    "io"
)

func GetServiceData(serviceEndpoints []Service) []ServiceData{
    var svcData []ServiceData
	fmt.Println("Scanning services HTTP endpoints . . .")
    for i, endpoint := range serviceEndpoints {
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
        sd.ServiceName = serviceEndpoints[i].URL
        sd.ServiceHTTPResponse = strconv.Itoa(resType)
        

		if (resType == 200) {
			fmt.Println("Service ", serviceEndpoints[i], "responded with 200, Ok ️✅")
		} else {
            fmt.Println("response type was invalid: ", serviceEndpoints[i], "->", resType, "❌")
        }

        if serviceEndpoints[i].API_URL != nil {
            apiReq, err := http.NewRequest("GET", *endpoint.API_URL, nil)
            if err != nil {
                panic(err)
            }
            apiReq.Header.Set("Authorization", "Bearer "+*endpoint.API_KEY)
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