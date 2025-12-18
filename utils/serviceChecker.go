package utils

import (
	"database/sql"
	"slices"
)

// Takes in service data and checks for any bad responses
// Upon bad responses will fire trigger events if any and return bad responses or nil
func Check(data []ServiceData) []ServiceData {
	var ret []ServiceData
	for _, el := range data {
		var valid_responses []string

		service_config, ok := Current_Config.Services[el.ServiceName]
		if ok && service_config.Valid_Responses != nil && len(*service_config.Valid_Responses) > 0 {
			valid_responses = *service_config.Valid_Responses
		} else {
			valid_responses = []string{"200"}
		}
		if !slices.Contains(valid_responses, el.ServiceHTTPResponse) {
			ret = append(ret, el)
		}
	}
	
	return ret
}

func GetUptimeAverage(db *sql.DB, name string) float64 {
	data := GetDataForService(db, name)
	numberDown := len(Check(data))
	totalNumber := len(data)
	if totalNumber == 0 {
		return 0.0
	}
	average := float64(numberDown) / float64(totalNumber)
	return average
}
