package server

import (
	"encoding/json"
	"goUp/utils"
	"net/http"
)

// @Summary Get or update the polling schedule
// @Tags schedule
// @Accept json
// @Produce json
// @Param schedule body utils.ScheduleState false "Schedule to set (POST only)"
// @Success 200 {object} utils.ScheduleState "current schedule (GET) or {updated:true} (POST)"
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "internal server error"
// @Router /api/schedule [get]
// @Router /api/schedule [post]
func (s *Server) ScheduleApi(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		dec := json.NewDecoder(req.Body)
		var jsonData utils.ScheduleState

		if err := dec.Decode(&jsonData); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		updated := s.scheduleUpdater(jsonData)

		if updated {
			w.Header().Add("Content-Type", "application/json")
			if _, err := w.Write([]byte("{ \"updated\": true }")); err != nil {
				panic(err)
			}
		} else {
			w.Header().Add("Content-Type", "application/json")
			http.Error(w, "{ \"updated\": false }", http.StatusBadRequest)
			return
		}
	case "GET":
		w.Header().Add("Content-Type", "application/json")
		state := s.scd.Get()
		if err := json.NewEncoder(w).Encode(state); err != nil {
			panic(err)
		}
	default:
		http.Error(w, "Invalid API Request", http.StatusBadRequest)
		return
	}
}

// Updates schedule parameters from schedule API
func (s *Server) scheduleUpdater(state utils.ScheduleState) bool {
	return s.scd.Update(state)
}
