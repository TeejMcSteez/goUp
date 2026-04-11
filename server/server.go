package server

import (
	"database/sql"
	"embed"
	scheduler "goUp/workers"
	"io/fs"
	"log"
	"net/http"
)

type Server struct {
	db  *sql.DB
	scd *scheduler.Scheduler
}

//go:embed all:static
var content embed.FS

// Returns a new server instance
func NewServer(db *sql.DB, scd *scheduler.Scheduler) *Server {
	return &Server{db: db, scd: scd}
}

// Starts server with all handler functions
//
// Returns an error if a problem with the server occurs
func (s *Server) Start() error {
	staticFS, err := fs.Sub(content, "static")
	if err != nil {
		return err
	}

	http.Handle("/", http.FileServer(http.FS(staticFS)))
	http.HandleFunc("/api", s.Api)
	http.HandleFunc("/api/schedule", s.ScheduleApi)
	http.HandleFunc("/api/status", s.StatusApi)
	http.HandleFunc("/api/uptime", s.UptimeAPI)
	http.HandleFunc("/api/errors", s.GetErrorData)
	http.HandleFunc("/api/db/size", s.GetDatabaseSize)
	http.HandleFunc("/api/db/persist", s.GetDatabasePersistence)
	http.HandleFunc("/api/db/clear", s.ClearDatabase)
	http.HandleFunc("/api/config", s.ReadConfigData)
	http.HandleFunc("/api/config/service", s.ConfigServiceApi)
	http.HandleFunc("/api/config/mqtt", s.ConfigMQTTApi)
	http.HandleFunc("/api/config/webhook", s.ConfigWebhookApi)
	log.Println("Starting server at http://localhost:8101/ . . .")
	if err := http.ListenAndServe(":8101", nil); err != nil {
		return err
	}

	return nil
}
