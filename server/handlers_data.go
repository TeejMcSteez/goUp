package server

import (
	"encoding/json"
	"goUp/utils"
	"log/slog"
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

		timeRange := req.URL.Query().Get("range")
		endpoints := utils.GetServiceEndpoints()

		switch timeRange {
		case "":
			var avgData []utils.AverageData
			for idx := range endpoints {
				endpointName := endpoints[idx].Name
				upAvg, err := utils.GetUptimeAverage(s.db, endpointName)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				// If upAvg is nil an error occured during the uptime calculation, ignore
				if upAvg != nil {
					avgData = append(avgData, utils.AverageData{Name: endpointName, Average: *upAvg})
				}
			}
			if err := json.NewEncoder(w).Encode(avgData); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "1hr":
			var avgData []utils.AverageData
			for idx := range endpoints {
				endpointName := endpoints[idx].Name
				upAvg, err := utils.GetPastHourUptime(s.db, endpointName)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				// If upAvg is nil an error occured during the uptime calculation, ignore
				if upAvg != nil {
					avgData = append(avgData, utils.AverageData{Name: endpointName, Average: *upAvg})
				}
			}
			if err := json.NewEncoder(w).Encode(avgData); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "12hr":
			var avgData []utils.AverageData
			for idx := range endpoints {
				endpointName := endpoints[idx].Name
				upAvg, err := utils.GetPast12HourUptime(s.db, endpointName)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				// If upAvg is nil an error occured during the uptime calculation, ignore
				if upAvg != nil {
					avgData = append(avgData, utils.AverageData{Name: endpointName, Average: *upAvg})
				}
			}
			if err := json.NewEncoder(w).Encode(avgData); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "week":
			var avgData []utils.AverageData
			for idx := range endpoints {
				endpointName := endpoints[idx].Name
				upAvg, err := utils.GetPastWeekUptime(s.db, endpointName)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				// If upAvg is nil an error occured during the uptime calculation, ignore
				if upAvg != nil {
					avgData = append(avgData, utils.AverageData{Name: endpointName, Average: *upAvg})
				}
			}
			if err := json.NewEncoder(w).Encode(avgData); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "day":
			var avgData []utils.AverageData
			for idx := range endpoints {
				endpointName := endpoints[idx].Name
				upAvg, err := utils.GetPastDayUptime(s.db, endpointName)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				// If upAvg is nil an error occured during the uptime calculation, ignore
				if upAvg != nil {
					avgData = append(avgData, utils.AverageData{Name: endpointName, Average: *upAvg})
				}
			}
			if err := json.NewEncoder(w).Encode(avgData); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "month":
			var avgData []utils.AverageData
			for idx := range endpoints {
				endpointName := endpoints[idx].Name
				upAvg, err := utils.GetPastMonthUptime(s.db, endpointName)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				// If upAvg is nil an error occured during the uptime calculation, ignore
				if upAvg != nil {
					avgData = append(avgData, utils.AverageData{Name: endpointName, Average: *upAvg})
				}
			}
			if err := json.NewEncoder(w).Encode(avgData); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "year":
			var avgData []utils.AverageData
			for idx := range endpoints {
				endpointName := endpoints[idx].Name
				upAvg, err := utils.GetPastYearUptime(s.db, endpointName)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				// If upAvg is nil an error occured during the uptime calculation, ignore
				if upAvg != nil {
					avgData = append(avgData, utils.AverageData{Name: endpointName, Average: *upAvg})
				}
			}
			if err := json.NewEncoder(w).Encode(avgData); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		default:
			http.Error(w, "bad range provided to search over", http.StatusBadRequest)
			slog.Error("client requested uptime API over range with bad range provided", "error", "bad request")
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
		slog.Error("Error occured, invalid limit", "limit", limit, "error", err)
		limit = 0
	}
	sortOrder := req.URL.Query().Get("sort")
	data, err := utils.GetErrorData(s.db, limit, sortOrder)
	if err != nil {
		slog.Error("Error occured getting error data from database", "error", err)
		if err := json.NewEncoder(w).Encode([]byte(err.Error())); err != nil {
			http.Error(w, "Failed to encode error message to json", http.StatusInternalServerError)
		}
		return
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("Error encoding error data to json", "error", err)
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
		slog.Error("Error occured getting response time data from database", "error", err)
		if err := json.NewEncoder(w).Encode([]byte(err.Error())); err != nil {
			http.Error(w, "Failed to encode error message to json", http.StatusInternalServerError)
		}
		return
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("Error encoding error data to json", "error", err)
		http.Error(w, "Failed to encode data to json", http.StatusInternalServerError)
	}
}

// @Summary Manually trigger a service data fetch
// @Description Fires the scheduler immediately instead of waiting for the next scheduled interval
// @Tags data
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {string} string "internal server error"
// @Router /api/fire [post]
func (s *Server) ManualFire(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if req.Method != "POST" {
		http.Error(w, "invalid method", http.StatusInternalServerError)
		return
	}
	s.scd.Fire()
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "success"}); err != nil {
		slog.Error("Failed to send response message", "error", err)
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
	}
}

// @Summary Gets all TLS certificate(s)
// @Description Returns all TLS certificates - null if none
// @Tags data
// @Produce json
// @Success 200 {array} utils.TlsStatus
// @Failure 500 {string} string "internal server error"
// @Router /api/tls [get]
func (s *Server) GetTls(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if req.Method != "GET" {
		http.Error(w, "invalid method", http.StatusInternalServerError)
		return
	}
	data, err := utils.GetExpiredTls(s.db)
	if err != nil {
		http.Error(w, "failed to get TLS data", http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("Error encoding tls data to json", "error", err)
		http.Error(w, "Failed to encode data to json", http.StatusInternalServerError)
	}

}
