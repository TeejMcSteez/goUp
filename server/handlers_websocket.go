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
	hub  *Hub
}

// Hub owns the set of connected wsConns and is the only goroutine that
// reads or writes that set, so no mutex is needed around it. Workers push
// fresh data in via broadcast; wsConns come and go via register/unregister.
type Hub struct {
	clients    map[*wsConn]struct{}
	broadcast  chan []byte
	register   chan *wsConn
	unregister chan *wsConn
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*wsConn]struct{}),
		broadcast:  make(chan []byte),
		register:   make(chan *wsConn),
		unregister: make(chan *wsConn),
	}
}

// Broadcast satisfies workers.Broadcaster, letting the scheduler push fresh
// data out to every connected client without importing the server package.
func (h *Hub) Broadcast(b []byte) {
	h.broadcast <- b
}

func (h *Hub) Run() {
	for {
		select {
		case ws := <-h.register:
			h.clients[ws] = struct{}{}
		case ws := <-h.unregister:
			if _, ok := h.clients[ws]; ok {
				delete(h.clients, ws)
				close(ws.send)
			}
		case b := <-h.broadcast:
			for ws := range h.clients {
				select {
				case ws.send <- b:
				default:
					// client too slow to keep up, drop it rather than
					// block delivery to everyone else
					delete(h.clients, ws)
					close(ws.send)
				}
			}
		}
	}
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

// @Summary Start Websocket connection to receive data on every fetch
// @Description Upgrades the HTTP connection to a websocket. Not a REST call — no request/response body. Once connected, the server pushes a utils.ServiceResponse JSON message on every scheduler fetch cycle; the client sends nothing.
// @Tags websocket
// @Success 101 {object} utils.ServiceResponse "Switching Protocols; subsequent messages pushed to the client are utils.ServiceResponse JSON"
// @Failure 500 {string} string "internal server error"
// @Router /ws [get]
func (s *Server) handleWs(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		slog.Error("websocket handler error", "error", err)
		return
	}
	wsConn := &wsConn{conn: conn, done: make(chan struct{}), send: make(chan []byte), hub: s.hub}
	s.hub.register <- wsConn
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
			ws.stop()
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
				ws.stop()
				return
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
	ws.once.Do(func() {
		close(ws.done)
		ws.hub.unregister <- ws
	})
}
