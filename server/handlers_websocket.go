package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"goUp/utils"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type outMsg struct {
	typ  int
	data []byte
}

type wsConn struct {
	conn  *websocket.Conn
	send  chan []byte
	reply chan outMsg
	done  chan struct{}
	once  sync.Once
	hub   *Hub
	db    *sql.DB
}

// Hub owns the set of connected wsConns and is the only goroutine that
// reads or writes that set, so no mutex is needed around it. Workers push
// fresh data in via broadcast; wsConns come and go via register/unregister.
type Hub struct {
	clients    map[*wsConn]struct{}
	broadcast  chan []byte
	register   chan *wsConn
	unregister chan *wsConn
	latest     []byte // last thing broadcast, only touched by Run()
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

func (ws *wsConn) enqueue(typ int, b []byte) {
	select {
	case ws.reply <- outMsg{typ, b}:
	case <-ws.done: // conn shutting down drop
	}
}

func (h *Hub) Run() {
	for {
		select {
		case ws := <-h.register:
			h.clients[ws] = struct{}{}
			if h.latest != nil {
				ws.send <- h.latest // fresh channel with buffer, won't block
			}
		case ws := <-h.unregister:
			if _, ok := h.clients[ws]; ok {
				delete(h.clients, ws)
				close(ws.send)
			}
		case b := <-h.broadcast:
			h.latest = b
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

	readLimit = 512
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
	wsConn := &wsConn{db: s.db, conn: conn, done: make(chan struct{}), send: make(chan []byte, 16), hub: s.hub}
	s.hub.register <- wsConn
	go wsConn.readLoop()
	go wsConn.writeLoop()
}

type ReadType string

const (
	ReadErrors        ReadType = "errors"
	ReadUptime        ReadType = "uptime"
	ReadResponseTimes ReadType = "rt"
)

type ErrorParams struct {
	Limit     int    `json:"limit"`
	SortOrder string `json:"sortOrder"`
}

type ResponseTimeParams struct {
	Type byte
	From string `json:"from"`
	To   string `json:"to"`
}

type UptimeParams struct {
	// Range matches the /api/uptime ?range= values: "" (all time),
	// "1hr", "12hr", "day", "week", "month", "year".
	Range string `json:"range"`
}

// uptimeFetchers maps an UptimeParams.Range to the utils helper that
// computes a single service's uptime average over that window.
var uptimeFetchers = map[string]func(*sql.DB, string) (*float64, error){
	"":      utils.GetUptimeAverage,
	"1hr":   utils.GetPastHourUptime,
	"12hr":  utils.GetPast12HourUptime,
	"day":   utils.GetPastDayUptime,
	"week":  utils.GetPastWeekUptime,
	"month": utils.GetPastMonthUptime,
	"year":  utils.GetPastYearUptime,
}

type readMessage struct {
	Type         ReadType           `json:"type"`
	ErrParams    ErrorParams        `json:"errParams"`
	RtParams     ResponseTimeParams `json:"rtParams"`
	UptimeParams UptimeParams       `json:"uptimeParams"`
}

func (ws *wsConn) handleRead(t int, p []byte) {
	if t > 1 {
		// 1: websocket.TextMessage
		// 2: websocket.BinaryMessage
		// rest are control signals
		return
	}

	var msg readMessage
	err := json.Unmarshal(p, &msg)
	if err != nil {
		slog.Error("error reading message from websocket client", "ERROR", err)
		return
	}
	switch msg.Type {
	case "errors":
		data, err := utils.GetErrorData(ws.db, msg.ErrParams.Limit, msg.ErrParams.SortOrder)
		if err != nil {
			slog.Error("failed to get error data for websocket client", "ERROR", err)
			return
		}
		b, err := json.Marshal(data)
		if err != nil {
			return
		}
		ws.enqueue(websocket.TextMessage, b)
	case "rt":
		switch msg.RtParams.Type {
		case 0:
			data, err := utils.BuildRtSnapshot(ws.db)
			if err != nil {
				slog.Error("failed to build response time snapshot for websocket client", "ERROR", err)
				return
			}
			frame := encodeFrame(data.Typ, data.Seq, data.Names, data.Ts, data.Cols)
			ws.enqueue(websocket.BinaryMessage, frame)
		case 1:
			// TODO: factor out of datastore.go to be a general util helper for accepting from/to
			// as string to convert to unix time format
			fromUnix, perr := strconv.ParseInt(msg.RtParams.From, 10, 64)
			if perr != nil || fromUnix < 0 {
				slog.Error("failed to parse to in websocket read", "ERROR", perr)
				return
			}
			from := time.Unix(fromUnix, 0)
			to := time.Now()
			if msg.RtParams.To != "" {
				toUnix, perr := strconv.ParseInt(msg.RtParams.To, 10, 64)
				if perr != nil {
					slog.Error("invalid to time in websocket client", "ERROR", perr)
					return
				}
				// Stored timestamps carry sub-second precision, so stretch `to` to the
				// end of its second; otherwise rows later in that second are dropped.
				to = time.Unix(toUnix, 0).Add(time.Second - time.Nanosecond)
			}
			if from.After(to) {
				slog.Error("'from' must be at or before 'to'", "ERROR", perr)
				return
			}
			data, err := utils.BuildRtSnapshotRange(ws.db, from, to)
			if err != nil {
				slog.Error("failed to build response time append for websocket client", "ERROR", err)
				return
			}
			frame := encodeFrame(data.Typ, data.Seq, data.Names, data.Ts, data.Cols)
			ws.enqueue(websocket.BinaryMessage, frame)

		default:
			slog.Error("invalid type based for response time fetch", "ERROR", errors.New("invalid input"))
			return
		}
	case "uptime":
		fetch, ok := uptimeFetchers[msg.UptimeParams.Range]
		if !ok {
			slog.Error("websocket client requested uptime with bad range", "range", msg.UptimeParams.Range)
			return
		}
		var data []utils.AverageData
		for _, endpoint := range utils.GetServiceEndpoints() {
			upAvg, err := fetch(ws.db, endpoint.Name)
			if err != nil {
				slog.Error("failed to get uptime data for websocket client", "service", endpoint.Name, "ERROR", err)
				return
			}
			// nil means the uptime calculation had nothing usable, skip it
			// (same behavior as UptimeAPI)
			if upAvg != nil {
				data = append(data, utils.AverageData{Name: endpoint.Name, Average: *upAvg})
			}
		}
		b, err := json.Marshal(data)
		if err != nil {
			slog.Error("failed to marshal uptime data for websocket client", "ERROR", err)
			return
		}
		ws.enqueue(websocket.TextMessage, b)
	default:
		// invalid key
	}
}

func (ws *wsConn) readLoop() {
	defer ws.stop()
	ws.conn.SetReadLimit(readLimit)
	_ = ws.conn.SetReadDeadline(time.Now().Add(pongWait))
	ws.conn.SetPongHandler(func(string) error {
		return ws.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		if t, p, err := ws.conn.ReadMessage(); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				slog.Warn("websocket closed unexpectedly", "error", err)
			}
			return
		} else {
			// handles reads to push data back to client
			// based off ReadType type
			ws.handleRead(t, p)
		}
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
			err := ws.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait))
			if err != nil {
				slog.Error("error writing websocket control messsage", "error", err)
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
		case m := <-ws.reply:
			if err := ws.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				slog.Error("failed to set write deadline for websocket clinet", "ERROR", err)
			}
			if err := ws.conn.WriteMessage(m.typ, m.data); err != nil {
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
