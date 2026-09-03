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

	handlers := map[string]func(w http.ResponseWriter, req *http.Request){
		"/health":                    s.Health,
		"/api":                       s.Api,
		"/api/fire":                  s.ManualFire,
		"/api/schedule":              s.ScheduleApi,
		"/api/status":                s.StatusApi,
		"/api/uptime":                s.UptimeAPI,
		"/api/rt":                    s.GetResponseTimes,
		"/api/tls":                   s.GetTls,
		"/api/errors":                s.GetErrorData,
		"/api/db/size":               s.GetDatabaseSize,
		"/api/db/persist":            s.GetDatabasePersistence,
		"/api/db/clear":              s.ClearDatabase,
		"/api/config":                s.ReadConfigData,
		"/api/config/service":        s.ConfigServiceApi,
		"/api/config/service/active": s.ConfigActiveApi,
		"/api/config/mqtt":           s.ConfigMQTTApi,
		"/api/config/webhook":        s.ConfigWebhookApi,
		"/api/config/smtp":           s.ConfigSMTPApi,
		"/api/config/gotify":         s.ConfigGotifyApi,
		"/api/config/slack":          s.ConfigSlackApi,
		"/api/config/telegram":       s.ConfigTelegramApi,
		"/api/config/ha":             s.ConfigHAApi,
		"/api/config/discord":        s.ConfigDiscordApi,
		"/api/config/backoff":        s.ConfigBackoffApi,
		"/api/config/size":           s.ConfigDatabaseApi,
		"/swagger/":                  httpSwagger.WrapHandler,
		"/ws":                        s.handleWs,
	}

	for url, handler := range handlers {
		http.HandleFunc(url, handler)
	}

	s.handler.next = http.DefaultServeMux
	slog.Info("Starting server", "address", "http://localhost:8101/")
	if err := http.ListenAndServe(":8101", s.handler); err != nil {
		return err
	}

	return nil
}
