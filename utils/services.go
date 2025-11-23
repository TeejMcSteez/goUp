package utils

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func GetServiceData(serviceEndpoints []string) []ServiceData{
    var svcData []ServiceData
	fmt.Println("Scanning services HTTP endpoints . . .")
    for i, endpoint := range serviceEndpoints {
        res, err := http.Get(endpoint)
        if err != nil {
            log.Printf("error fetching %s: %v %s", endpoint, err, "❌")
            var sd ServiceData
            sd.ServiceName = serviceEndpoints[i]
            sd.ServiceHTTPResponse = err.Error()
            svcData = append(svcData, sd)
            continue
        }

		resType := res.StatusCode

        var sd ServiceData
        sd.ServiceName = serviceEndpoints[i]
        sd.ServiceHTTPResponse = strconv.Itoa(resType)
        
        svcData = append(svcData, sd)

		if (resType == 200) {
			fmt.Println("Service ", serviceEndpoints[i], "responded with 200, Ok ️✅")
		} else {
            fmt.Println("response type was invalid: ", serviceEndpoints[i], "->", resType, "❌")
        }
    }
    return svcData
}