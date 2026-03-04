package utils

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"slices"
)

type NoConfigError struct {
	Field  string
	Reason string
}

func (e *NoConfigError) Error() string {
	return fmt.Sprintf("failed to check data for field '%s': %s", e.Field, e.Reason)
}

// Takes in service data and checks for any bad responses
// Valid responses will either be configured or will be the default 200
func Check(data []ServiceData) ([]ServiceData, error) {
	var ret []ServiceData
	if Current_Config == nil {
		return nil, &NoConfigError{"configuration", "cannot be nil"}
	}
	for i := range data {
		var valid_responses []string

		service_config, ok := Current_Config.Services[data[i].ServiceName]
		if ok && service_config.Valid_Responses != nil && len(*service_config.Valid_Responses) > 0 {
			valid_responses = *service_config.Valid_Responses
		} else {
			valid_responses = []string{"200"}
		}
		if !slices.Contains(valid_responses, data[i].ServiceHTTPResponse) {
			data[i].Error = true
			ret = append(ret, data[i])
		}
	}

	return ret, nil
}

// Returns the uptime average for a service or error
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
		return 0.0, err
	}
	average := float64(numberDown) / float64(totalNumber)
	// Rounds to nearest 2nd decmial place
	rounded_average := math.Round(average*100) / 100
	return rounded_average, nil
}
