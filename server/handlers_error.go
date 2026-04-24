package server

import "net/http"

func (s *Server) HandleNoUi(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "UI is not enabled on this server", http.StatusNotImplemented)
}
