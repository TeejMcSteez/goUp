package utils

import (
	"database/sql"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"time"
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
func GetUptimeAverage(db *sql.DB, name string) (*float64, error) {
	data, err := GetDataForService(db, name)
	if err != nil {
		return nil, err
	}
	chk, err := Check(data)
	if err != nil {
		slog.Error("Error while checking data in getting uptime averages", "error", err)
		return nil, err
	}
	numberDown := len(chk)
	totalNumber := len(data)
	if totalNumber == 0 {
		return nil, err
	}
	average := float64(numberDown) / float64(totalNumber)
	// Rounds to 2 decimal places
	rounded_average, err := strconv.ParseFloat(strconv.FormatFloat(average, 'f', 2, 64), 64)
	if err != nil {
		return nil, err
	}
	return &rounded_average, nil
}

// Proposed: Could change SQL from text to integer to avoid all these helper functions and use the WHERE
// clause to filter data over a range however this doesn't require any database mechanics
// and stays simpler by just collecting data over a range in Go over SQL
func GetPastHourUptime(db *sql.DB, name string) (*float64, error) {
	now := time.Now()
	since := now.Add(-time.Hour)
	data, err := GetDataForService(db, name)
	if err != nil {
		return nil, err
	}
	var pastHr []ServiceData
	for _, d := range data {
		if d.Timestamp.After(since) && d.Timestamp.Before(now) {
			pastHr = append(pastHr, d)
		}
	}
	chk, err := Check(pastHr)
	if err != nil {
		slog.Error("Error while checking data in getting uptime averages", "error", err)
		return nil, err
	}
	numberDown := len(chk)
	totalNumber := len(pastHr)
	if totalNumber == 0 {
		return nil, err
	}
	average := float64(numberDown) / float64(totalNumber)
	// Rounds to 2 decimal places
	rounded_average, err := strconv.ParseFloat(strconv.FormatFloat(average, 'f', 2, 64), 64)
	if err != nil {
		return nil, err
	}
	return &rounded_average, nil
}

func GetPast12HourUptime(db *sql.DB, name string) (*float64, error) {
	now := time.Now()
	since := now.Add(-time.Hour * 12)
	data, err := GetDataForService(db, name)
	if err != nil {
		return nil, err
	}
	var pastHr []ServiceData
	for _, d := range data {
		if d.Timestamp.After(since) && d.Timestamp.Before(now) {
			pastHr = append(pastHr, d)
		}
	}
	chk, err := Check(pastHr)
	if err != nil {
		slog.Error("Error while checking data in getting uptime averages", "error", err)
		return nil, err
	}
	numberDown := len(chk)
	totalNumber := len(pastHr)
	if totalNumber == 0 {
		return nil, err
	}
	average := float64(numberDown) / float64(totalNumber)
	// Rounds to 2 decimal places
	rounded_average, err := strconv.ParseFloat(strconv.FormatFloat(average, 'f', 2, 64), 64)
	if err != nil {
		return nil, err
	}
	return &rounded_average, nil
}

func GetPastDayUptime(db *sql.DB, name string) (*float64, error) {
	now := time.Now()
	since := now.Add(-time.Hour * 24)
	data, err := GetDataForService(db, name)
	if err != nil {
		return nil, err
	}
	var pastHr []ServiceData
	for _, d := range data {
		if d.Timestamp.After(since) && d.Timestamp.Before(now) {
			pastHr = append(pastHr, d)
		}
	}
	chk, err := Check(pastHr)
	if err != nil {
		slog.Error("Error while checking data in getting uptime averages", "error", err)
		return nil, err
	}
	numberDown := len(chk)
	totalNumber := len(pastHr)
	if totalNumber == 0 {
		return nil, err
	}
	average := float64(numberDown) / float64(totalNumber)
	// Rounds to 2 decimal places
	rounded_average, err := strconv.ParseFloat(strconv.FormatFloat(average, 'f', 2, 64), 64)
	if err != nil {
		return nil, err
	}
	return &rounded_average, nil
}

func GetPastWeekUptime(db *sql.DB, name string) (*float64, error) {
	now := time.Now()
	since := now.Add(-time.Hour * 168) // 168 hours in a week
	data, err := GetDataForService(db, name)
	if err != nil {
		return nil, err
	}
	var pastHr []ServiceData
	for _, d := range data {
		if d.Timestamp.After(since) && d.Timestamp.Before(now) {
			pastHr = append(pastHr, d)
		}
	}
	chk, err := Check(pastHr)
	if err != nil {
		slog.Error("Error while checking data in getting uptime averages", "error", err)
		return nil, err
	}
	numberDown := len(chk)
	totalNumber := len(pastHr)
	if totalNumber == 0 {
		return nil, err
	}
	average := float64(numberDown) / float64(totalNumber)
	// Rounds to 2 decimal places
	rounded_average, err := strconv.ParseFloat(strconv.FormatFloat(average, 'f', 2, 64), 64)
	if err != nil {
		return nil, err
	}
	return &rounded_average, nil
}

func GetPastMonthUptime(db *sql.DB, name string) (*float64, error) {
	now := time.Now()
	since := now.Add(-time.Hour * 730) // 730 hours in a month (roughly)
	data, err := GetDataForService(db, name)
	if err != nil {
		return nil, err
	}
	var pastHr []ServiceData
	for _, d := range data {
		if d.Timestamp.After(since) && d.Timestamp.Before(now) {
			pastHr = append(pastHr, d)
		}
	}
	chk, err := Check(pastHr)
	if err != nil {
		slog.Error("Error while checking data in getting uptime averages", "error", err)
		return nil, err
	}
	numberDown := len(chk)
	totalNumber := len(pastHr)
	if totalNumber == 0 {
		return nil, err
	}
	average := float64(numberDown) / float64(totalNumber)
	// Rounds to 2 decimal places
	rounded_average, err := strconv.ParseFloat(strconv.FormatFloat(average, 'f', 2, 64), 64)
	if err != nil {
		return nil, err
	}
	return &rounded_average, nil
}

func GetPastYearUptime(db *sql.DB, name string) (*float64, error) {
	now := time.Now()
	since := now.Add(-time.Hour * 8760) // 8760 hours in a year (roughly)
	data, err := GetDataForService(db, name)
	if err != nil {
		return nil, err
	}
	var pastHr []ServiceData
	for _, d := range data {
		if d.Timestamp.After(since) && d.Timestamp.Before(now) {
			pastHr = append(pastHr, d)
		}
	}
	chk, err := Check(pastHr)
	if err != nil {
		slog.Error("Error while checking data in getting uptime averages", "error", err)
		return nil, err
	}
	numberDown := len(chk)
	totalNumber := len(pastHr)
	if totalNumber == 0 {
		return nil, err
	}
	average := float64(numberDown) / float64(totalNumber)
	// Rounds to 2 decimal places
	rounded_average, err := strconv.ParseFloat(strconv.FormatFloat(average, 'f', 2, 64), 64)
	if err != nil {
		return nil, err
	}
	return &rounded_average, nil
}
