package server

import (
	"database/sql"
	"embed"
	_ "goUp/docs"
	scheduler "goUp/workers"
	"io/fs"
	"log/slog"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
)

type Server struct {
	db      *sql.DB
	scd     *scheduler.Scheduler
	serveUi bool
	handler *GoupHandler
}

type GoupHandler struct {
	Cors   bool
	Origin *string
	next   http.Handler
}

func (g *GoupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	getOrigin := func(handler *GoupHandler) string {
		if handler.Origin == nil {
			return "*"
		}
		return *handler.Origin
	}(g)
	if g.Cors {
		w.Header().Add("Access-Control-Allow-Origin", getOrigin)
	}
	g.next.ServeHTTP(w, r)
}

//go:embed all:static
var content embed.FS

// Returns a new server instance
func NewServer(db *sql.DB, scd *scheduler.Scheduler, serveUi bool, handler GoupHandler) *Server {
	return &Server{db: db, scd: scd, serveUi: serveUi, handler: &handler}
}

// Starts server with all handler functions
//
// Returns an error if a problem with the server occurs
func (s *Server) Start() error {
	if s.serveUi {
		staticFS, err := fs.Sub(content, "static")
		if err != nil {
			return err
		}
		http.Handle("/", http.FileServer(http.FS(staticFS)))
	} else {
		http.HandleFunc("/", s.HandleNoUi)
	}
	http.HandleFunc("/health", s.Health)
	http.HandleFunc("/api", s.Api)
	http.HandleFunc("/api/fire", s.ManualFire)
	http.HandleFunc("/api/schedule", s.ScheduleApi)
	http.HandleFunc("/api/status", s.StatusApi)
	http.HandleFunc("/api/uptime", s.UptimeAPI)
	http.HandleFunc("/api/rt", s.GetResponseTimes)
	http.HandleFunc("/api/tls", s.GetTls)
	http.HandleFunc("/api/errors", s.GetErrorData)
	http.HandleFunc("/api/db/size", s.GetDatabaseSize)
	http.HandleFunc("/api/db/persist", s.GetDatabasePersistence)
	http.HandleFunc("/api/db/clear", s.ClearDatabase)
	http.HandleFunc("/api/config", s.ReadConfigData)
	http.HandleFunc("/api/config/service", s.ConfigServiceApi)
	http.HandleFunc("/api/config/service/active", s.ConfigActiveApi)
	http.HandleFunc("/api/config/mqtt", s.ConfigMQTTApi)
	http.HandleFunc("/api/config/webhook", s.ConfigWebhookApi)
	http.HandleFunc("/api/config/smtp", s.ConfigSMTPApi)
	http.HandleFunc("/api/config/gotify", s.ConfigGotifyApi)
	http.HandleFunc("/api/config/slack", s.ConfigSlackApi)
	http.HandleFunc("/api/config/telegram", s.ConfigTelegramApi)
	http.HandleFunc("/api/config/ha", s.ConfigHAApi)
	http.HandleFunc("/api/config/discord", s.ConfigDiscordApi)
	http.HandleFunc("/api/config/backoff", s.ConfigBackoffApi)
	http.HandleFunc("/api/config/size", s.ConfigDatabaseApi)
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)
	s.handler.next = http.DefaultServeMux
	slog.Info("Starting server", "address", "http://localhost:8101/")
	if err := http.ListenAndServe(":8101", s.handler); err != nil {
		return err
	}

	return nil
}
