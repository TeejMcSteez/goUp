package server

import (
	"encoding/json"
	"goUp/utils"
	"log/slog"
	"net/http"
	"strconv"
	"time"
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
// @Description Returns a full snapshot frame (type 0) containing service names, timestamps and per-service response time columns.
// @Description Optionally limited to a time window via `from`/`to` (unix seconds, inclusive).
// @Description Binary layout is described by encodeFrame in server/protocol.go.
// @Tags data
// @Produce octet-stream
// @Param from query int false "Start of window, unix seconds (inclusive). Omit for full history"
// @Param to query int false "End of window, unix seconds (inclusive). Omit for now"
// @Success 200 {file} binary "snapshot frame"
// @Failure 400 {string} string "invalid from/to"
// @Failure 500 {string} string "internal server error"
// @Router /api/rt [get]
func (s *Server) rtSnapshot(w http.ResponseWriter, r *http.Request) {
	f := r.URL.Query().Get("from")
	t := r.URL.Query().Get("to")

	var data utils.ResponseTimePayload
	var err error
	if f == "" {
		// No filter: send all data
		data, err = utils.BuildRtSnapshot(s.db)
	} else {
		fromUnix, perr := strconv.ParseInt(f, 10, 64)
		if perr != nil || fromUnix < 0 {
			http.Error(w, "'from' must be a non-negative unix timestamp", http.StatusBadRequest)
			return
		}
		from := time.Unix(fromUnix, 0)
		to := time.Now()
		if t != "" {
			toUnix, perr := strconv.ParseInt(t, 10, 64)
			if perr != nil {
				http.Error(w, "'to' must be a unix timestamp", http.StatusBadRequest)
				return
			}
			// Stored timestamps carry sub-second precision, so stretch `to` to the
			// end of its second; otherwise rows later in that second are dropped.
			to = time.Unix(toUnix, 0).Add(time.Second - time.Nanosecond)
		}
		if from.After(to) {
			http.Error(w, "'from' must be at or before 'to'", http.StatusBadRequest)
			return
		}
		data, err = utils.BuildRtSnapshotRange(s.db, from, to)
	}
	if err != nil {
		slog.Error("Error occured getting response time data from database", "error", err)
		http.Error(w, "Failed to get response time data", http.StatusInternalServerError)
		return
	}
	frame := encodeFrame(data.Typ, data.Seq, data.Names, data.Ts, data.Cols)
	w.Header().Set("Content-Type", "application/octet-stream")
	// Headers are already sent once Write starts, so a failure can only be logged
	if _, err := w.Write(frame); err != nil {
		slog.Error("Error writing response time frame", "error", err)
	}
}

// SnapshotData is the request body for an append (diff) frame request.
// Both fields are unix seconds, matching the frame's timestamp column, and
// both ends are inclusive of their whole second. To = 0 means "now".
type SnapshotData struct {
	From int64 `json:"from" example:"1793620800"`
	To   int64 `json:"to,omitempty" example:"1793621100"`
}

// @Summary Get response times appended since a previous snapshot
// @Description Returns an append frame (type 1) with response time rows between `from` and `to`, used by the frontend to extend an existing snapshot without refetching it.
// @Tags data
// @Accept json
// @Produce octet-stream
// @Param body body SnapshotData true "Time window (unix seconds, inclusive; omit `to` for now)"
// @Success 200 {file} binary "append frame"
// @Failure 400 {string} string "failed to parse request body"
// @Failure 500 {string} string "internal server error"
// @Router /api/rt [post]
func (s *Server) rtDiff(w http.ResponseWriter, req *http.Request) {
	var data SnapshotData
	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		slog.Error("failed to decode request body", "error", err)
		http.Error(w, "failed to parse response body", http.StatusBadRequest)
		return
	}
	from := time.Unix(data.From, 0)
	// Stored timestamps carry sub-second precision, so stretch `to` to the
	// end of its second; otherwise rows later in that second are dropped.
	to := time.Now()
	if data.To != 0 {
		to = time.Unix(data.To, 0).Add(time.Second - time.Nanosecond)
	}
	if data.From < 0 || from.After(to) {
		http.Error(w, "'from' must be a unix timestamp at or before 'to'", http.StatusBadRequest)
		return
	}
	payload, err := utils.GetRtAppend(s.db, from, to)
	if err != nil {
		slog.Error("failed to get append snapshot", "error", err)
		http.Error(w, "failed to get payload append data", http.StatusInternalServerError)
		return
	}
	frame := encodeFrame(payload.Typ, payload.Seq, payload.Names, payload.Ts, payload.Cols)
	w.Header().Set("Content-Type", "application/octet-stream")
	if _, err := w.Write(frame); err != nil {
		slog.Error("error writing append time frame", "error", err)
	}
}

// GetResponseTimes dispatches /api/rt by method: GET returns a full snapshot
// frame (rtSnapshot), POST returns an append frame for a time window (rtDiff).
// Swagger annotations live on rtSnapshot and rtDiff.
func (s *Server) GetResponseTimes(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case "POST":
		s.rtDiff(w, req)
	case "GET":
		s.rtSnapshot(w, req)
	default:
		http.Error(w, "invalid request", http.StatusBadRequest)
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
