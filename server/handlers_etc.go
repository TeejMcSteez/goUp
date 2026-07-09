package server

import (
	"log"
	"net/http"
)

func (s *Server) Health(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if _, err := w.Write([]byte("{\"ok\": true}")); err != nil {
		log.Printf("error sending health check API message: %v", err)
	}
}
