package server

import (
	"log/slog"
	"net/http"
)

func (s *Server) Health(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if _, err := w.Write([]byte("{\"ok\": true}")); err != nil {
		slog.Error("error sending health check API message", "error", err)
	}
}
