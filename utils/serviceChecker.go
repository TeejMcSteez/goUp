package utils

import (
	"fmt"
	"database/sql"
)

// Takes in service data and checks for any bad responses
// Upon bad responses will fire trigger events if any and return bad responses or nil
func Check(data []ServiceData) []ServiceData {

	fmt.Println("Checking service data for errors")

	badRes := false
	var ret []ServiceData
	for _, el := range data {
		if el.ServiceHTTPResponse != "200" {
			badRes = true
			ret = append(ret, el)
		}
	}
	if badRes {
		Fire(ret)
	}
	return ret
}

func GetUptimeAverages(db *sql.DB, name string) float64 {
	data := GetDataForService(db, name)
	var total float64 = 0
	for idx := range data {
		if (data[idx].ServiceHTTPResponse == "200") {
			total += 1
		}
	}
	if len(data) == 0 {
		return 0.0
	}
	return total / float64(len(data))
}
