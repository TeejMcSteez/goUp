package server

import (
	"encoding/json"
	"fmt"
	"goUp/utils"
	"net/http"
)

// Returns size of database in bytes
func (s *Server) GetDatabaseSize(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	size, err := utils.GetDatabaseSize()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if err = json.NewEncoder(w).Encode(&utils.DatabaseSizePayload{Size: size}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) GetDatabasePersistence(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		w.Header().Add("Content-Type", "application/json")
		persists := utils.ReadConfigDatabasePersistence(utils.Current_Config)
		if err := json.NewEncoder(w).Encode(persists); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case "POST":
		w.Header().Set("Content-Type", "application/json")
		err := utils.UpdateConfigDatabasePersistence(utils.Current_Config)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	default:
		http.Error(w, "Invalid Request Method", http.StatusBadRequest)
	}
}

// Clears database
func (s *Server) ClearDatabase(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if err := utils.ClearDatabase(s.db); err != nil {
		if err := json.NewEncoder(w).Encode([]byte(err.Error())); err != nil {
			http.Error(w, "Failed to parse error message as json", http.StatusInternalServerError)
		}
	}
	if _, err := fmt.Fprint(w, `{ "ok": true }`); err != nil {
		http.Error(w, "Failed to write ok message", http.StatusInternalServerError)
	}
}
