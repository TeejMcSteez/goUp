package server

import (
	"encoding/json"
	"goUp/utils"
	"log"
	"net/http"
)

// @Summary Get the polling schedule
// @Tags schedule
// @Produce json
// @Success 200 {object} utils.ScheduleState
// @Failure 500 {string} string "internal server error"
// @Router /api/schedule [get]
func (s *Server) scheduleApiGet(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	state := s.scd.Get()
	if err := json.NewEncoder(w).Encode(state); err != nil {
		log.Printf("failed to encode polling schedule: %v", err)
	}
}

// @Summary Update the polling schedule
// @Tags schedule
// @Accept json
// @Produce json
// @Param schedule body utils.ScheduleState true "Schedule to set"
// @Success 200 {object} map[string]bool "{updated:true}"
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "internal server error"
// @Router /api/schedule [post]
func (s *Server) scheduleApiPost(w http.ResponseWriter, req *http.Request) {
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
			log.Printf("error writing polling update response: %v", err)
		}
	} else {
		w.Header().Add("Content-Type", "application/json")
		http.Error(w, "{ \"updated\": false }", http.StatusBadRequest)
		return
	}
}

func (s *Server) ScheduleApi(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		s.scheduleApiPost(w, req)
	case "GET":
		s.scheduleApiGet(w, req)
	default:
		http.Error(w, "Invalid API Request", http.StatusBadRequest)
		return
	}
}

// Updates schedule parameters from schedule API
func (s *Server) scheduleUpdater(state utils.ScheduleState) bool {
	return s.scd.Update(state)
}
