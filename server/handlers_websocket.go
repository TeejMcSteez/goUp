package server

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type wsConn struct {
	conn *websocket.Conn
	send chan []byte
	done chan struct{}
	once sync.Once
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	pingPeriod = (pongWait * 9) / 10
	// max message size
	_ = 512
)

func (s *Server) handleWs(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		slog.Error("websocket handler error", "error", err)
		return
	}
	wsConn := wsConn{conn: conn, done: make(chan struct{}), send: make(chan []byte)}
	go wsConn.readLoop()
	go wsConn.writeLoop()
}

func (ws *wsConn) readLoop() {
	for {
		_, msg, err := ws.conn.ReadMessage()
		if err != nil {
			slog.Error("error in websocket read loop", "error", err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("unexpected close error in websocket read loop", "error", err)
			}
			return
		}
		slog.Info("read loop input", "info", msg)
	}
}

func (ws *wsConn) writeLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if err := ws.conn.Close(); err != nil {
			slog.Error("error closing websocket connection in write loop", "error", err)
		}
	}()
	for {
		select {
		case <-ticker.C:
			if err := ws.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				slog.Error("error setting read deadline in websocket writeLoop", "error", err)
			}
			err := ws.conn.WriteMessage(1, nil)
			if err != nil {
				slog.Error("error writing websocket message", "error", err)
			}
		case b, ok := <-ws.send:
			if !ok {
				return
			}
			if err := ws.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				slog.Error("error setting read deadline in websocket writeLoop", "error", err)
			}
			if err := ws.conn.WriteMessage(websocket.TextMessage, b); err != nil {
				ws.stop()
				return
			}
		case <-ws.done:
			return
		}
	}
}

func (ws *wsConn) stop() {
	ws.once.Do(func() { close(ws.done) })
}
