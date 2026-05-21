package server

import (
	"encoding/json"
	"goUp/utils"
	"log"
	"net/http"
	"strconv"
)

// API handler, returns current service data from the database
func (s *Server) Api(w http.ResponseWriter, req *http.Request) {
	data, err := utils.GetRecentData(s.db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Gets current status in JSON format for automated fetching
//
// Only accepts GET requests which will return currently down services
func (s *Server) StatusApi(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		http.Error(w, "Invalid method", http.StatusBadRequest)
	case "GET":
		w.Header().Add("Content-Type", "application/json")

		recData, err := utils.GetRecentData(s.db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		apiData, err := utils.Check(recData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if err := json.NewEncoder(w).Encode(apiData); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
}

// Gets current uptime averages from the backend
func (s *Server) UptimeAPI(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		http.Error(w, "Invalid method", http.StatusBadRequest)
	case "GET":
		w.Header().Add("Content-Type", "application/json")
		endpoints := utils.GetServiceEndpoints()
		var avgData []utils.AverageData
		for idx := range endpoints {
			endpointName := endpoints[idx].Name
			upAvg, err := utils.GetUptimeAverage(s.db, endpointName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			avgData = append(avgData, utils.AverageData{Name: endpointName, Average: upAvg})
		}
		if err := json.NewEncoder(w).Encode(avgData); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
}

// Gets all service data which has errored
func (s *Server) GetErrorData(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	param := req.URL.Query().Get("limit")
	limit, err := strconv.Atoi(param)
	if err != nil {
		log.Printf("Error occured, invalid limit: %v\nError: %v", limit, err)
		limit = 0
	}
	sortOrder := req.URL.Query().Get("sort")
	data, err := utils.GetErrorData(s.db, limit, sortOrder)
	if err != nil {
		log.Printf("Error occured getting error data from database: %v", err)
		if err := json.NewEncoder(w).Encode([]byte(err.Error())); err != nil {
			http.Error(w, "Failed to encode error message to json", http.StatusInternalServerError)
		}
		return
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding error data to json: %v", err)
		if err := json.NewEncoder(w).Encode([]byte(err.Error())); err != nil {
			http.Error(w, "Failed to encode error message to json", http.StatusInternalServerError)
		}
		return
	}
}
