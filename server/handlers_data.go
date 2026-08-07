package server

import (
	"encoding/json"
	"goUp/utils"
	"log"
	"net/http"
	"strconv"
)

// @Summary Get all recent service data
// @Tags data
// @Produce json
// @Success 200 {array} utils.ServiceData
// @Failure 500 {string} string "internal server error"
// @Router /api [get]
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

// @Summary Get current service status returning any downed services or null for no services down
// @Description Returns all currently down services.
// @Tags data
// @Produce json
// @Success 200 {object} utils.ServiceResponse
// @Failure 500 {string} string "internal server error"
// @Router /api/status [get]
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

// @Summary Get uptime averages for all services
// @Tags data
// @Produce json
// @Success 200 {array} utils.AverageData
// @Failure 500 {string} string "internal server error"
// @Router /api/uptime [get]
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

// @Summary Get error log entries
// @Tags data
// @Produce json
// @Param limit query int false "Max number of results (0 = all)"
// @Param sort query string false "Sort order (asc or desc)"
// @Success 200 {array} utils.ServiceData
// @Failure 500 {string} string "internal server error"
// @Router /api/errors [get]
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
			http.Error(w, "Failed to encode error data to json", http.StatusInternalServerError)
		}
		return
	}
}

// @Summary Get response time history for all services
// @Tags data
// @Produce json
// @Success 200 {array} utils.ServiceResponseTime
// @Failure 500 {string} string "internal server error"
// @Router /api/rt [get]
func (s *Server) GetResponseTimes(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data, err := utils.GetResponseTimes(s.db)
	if err != nil {
		log.Printf("Error occured getting response time data from database: %v", err)
		if err := json.NewEncoder(w).Encode([]byte(err.Error())); err != nil {
			http.Error(w, "Failed to encode error message to json", http.StatusInternalServerError)
		}
		return
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding error data to json: %v", err)
		http.Error(w, "Failed to encode data to json", http.StatusInternalServerError)
	}
}

func (s *Server) ManualFire(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	s.scd.Fire()
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "success"}); err != nil {
		log.Printf("Failed to send response message, %v", err)
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
	}
}
