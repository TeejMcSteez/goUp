package server

import (
	"encoding/json"
	"fmt"
	"goUp/utils"
	"net/http"
)

// @Summary Get database file size in bytes for regular files; system-dependent for others https://pkg.go.dev/io/fs#FileInfo.Size
// @Tags database
// @Produce json
// @Success 200 {object} utils.DatabaseSizePayload
// @Failure 500 {string} string "internal server error"
// @Router /api/db/size [get]
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

// @Summary Get whether database persistence is enabled
// @Tags database
// @Produce json
// @Success 200 {boolean} bool "persistence enabled"
// @Failure 500 {string} string "internal server error"
// @Router /api/db/persist [get]
func (s *Server) getDatabasePersistenceGet(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	persists := utils.ReadConfigDatabasePersistence(utils.Current_Config)
	if err := json.NewEncoder(w).Encode(persists); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Summary Toggle database persistence
// @Tags database
// @Produce json
// @Success 200 {string} string "empty body"
// @Failure 500 {string} string "internal server error"
// @Router /api/db/persist [post]
func (s *Server) getDatabasePersistencePost(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := utils.UpdateConfigDatabasePersistence(utils.Current_Config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) GetDatabasePersistence(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		s.getDatabasePersistenceGet(w, req)
	case "POST":
		s.getDatabasePersistencePost(w, req)
	default:
		http.Error(w, "Invalid Request Method", http.StatusBadRequest)
	}
}

// @Summary Clear all data from the database
// @Tags database
// @Produce json
// @Success 200 {object} map[string]bool
// @Failure 500 {string} string "internal server error"
// @Router /api/db/clear [post]
func (s *Server) ClearDatabase(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if err := utils.ClearDatabase(s.db); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if _, err := fmt.Fprint(w, `{ "ok": true }`); err != nil {
		http.Error(w, "Failed to write ok message", http.StatusInternalServerError)
	}
}
