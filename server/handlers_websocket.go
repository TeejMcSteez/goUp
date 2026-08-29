package server

import (
	"encoding/json"
	"goUp/utils"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Unused for now - comment out for linter
// const (
// 	writeWait = 10 * time.Second

// 	pongWait = 60 * time.Second

// 	pingPeriod = (pongWait * 9) / 10

// 	maxMessageSize = 512
// )

func (s *Server) handleWs(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		slog.Error("websocket handler error", "error", err)
		return
	}
	go s.writeLoop(conn)
}

func (s *Server) writeLoop(c *websocket.Conn) {
	ticker := time.NewTicker(5 * time.Second)
	defer func() {
		if err := c.Close(); err != nil {
			slog.Error("error closing websocket connection", "error", err)
		}
	}()
	for {
		<-ticker.C
		recentData, err := utils.GetRecentData(s.db)
		if err != nil {
			slog.Error("error getting recent data in websocket write loop", "error", err)
			break
		}
		jsonData, err := json.Marshal(recentData)
		if err != nil {
			slog.Error("error encoding json data", "error", err)
			break
		}
		err = c.WriteMessage(1, jsonData)
		if err != nil {
			slog.Error("error writing websocket message", "error", err)
			break
		}
	}
}
