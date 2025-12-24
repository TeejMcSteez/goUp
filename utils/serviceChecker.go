package utils

import (
	"database/sql"
	"fmt"
	"log"
	"slices"
)

type NoConfigError struct {
	Field string
	Reason string
}

func (e *NoConfigError) Error() string {
	return fmt.Sprintf("failed to check data for field '%s': %s", e.Field, e.Reason)
}

// Takes in service data and checks for any bad responses
// Upon bad responses will fire trigger events if any and return bad responses or nil
func Check(data []ServiceData) ([]ServiceData, error) {
	var ret []ServiceData
	if Current_Config == nil {
		return nil, &NoConfigError{"configuration", "cannot be nil"}
	}
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
	
	return ret, nil
}
// Fix to return error or float to make sure GetDataForService is valid
func GetUptimeAverage(db *sql.DB, name string) (float64, error) {
	data, err := GetDataForService(db, name)
	if err != nil {
		return 0.0, err
	} 
	chk, err := Check(data)
	if err != nil {
		log.Fatalf("Error while checking data in getting uptime averages: %v", err)
	}
	numberDown := len(chk)
	totalNumber := len(data)
	if totalNumber == 0 {
		return 0.0,  err
	}
	average := float64(numberDown) / float64(totalNumber)
	return average, nil
}
