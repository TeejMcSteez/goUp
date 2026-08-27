package server

import (
	"context"
	"encoding/json"
	"goUp/utils"
	"log/slog"
	"net/http"
	"time"
)

func (s *Server) Health(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	// use timeout for any blocking/race issues not holding up endpoint
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := utils.CheckDatabaseHealth(ctx, s.db)

	if err != nil {
		if err := json.NewEncoder(w).Encode(map[string]any{
			"ok":    false,
			"error": err,
		}); err != nil {
			slog.Error("error sending health check API message", "error", err)
		}
		return
	}

	if _, err := w.Write([]byte("{\"ok\": true}")); err != nil {
		slog.Error("error sending health check API message", "error", err)
	}
}
