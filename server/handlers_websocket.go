package server

import (
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"goUp/utils"
	"goUp/workers"
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
	scd   *workers.Scheduler
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

	readLimit = 512
)

// @Summary Start Websocket connection to receive data on every fetch
// @Description Upgrades the HTTP connection to a websocket. Not a REST call — no request/response body. Once connected, the server pushes a utils.ServiceResponse JSON message on every scheduler fetch cycle.
// @Description The first message after connecting is the same payload, read from the DB.
// @Description The client may also send JSON requests `{"id":N,"type":"live"|"errors"|"uptime"|"rt", "errParams"|"uptimeParams"|"rtParams":{...}}`.
// @Description live/errors/uptime reply with JSON `{"id":N,"type":...,"data":...}` (or `"error"` on failure); rt replies with a binary message: u32 LE id followed by an encodeFrame frame.
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
	wsConn := &wsConn{db: s.db, conn: conn, done: make(chan struct{}), send: make(chan []byte, 16), reply: make(chan outMsg, 16), hub: s.hub, scd: s.scd}
	// Seed the client with the current services/status/TLS state read
	// straight from the DB, so it has data before the next fetch cycle
	// broadcasts. Queued before register so it's always the first message
	// and any broadcast that follows is at least as new.
	if b, err := liveSnapshot(s.db); err != nil {
		slog.Error("failed to build initial live data for websocket client", "error", err)
	} else {
		wsConn.send <- b // fresh buffered channel, nothing else writes yet
	}
	s.hub.register <- wsConn
	go wsConn.readLoop()
	go wsConn.writeLoop()
}

type WsMessageType string

const (
	ReadErrors        WsMessageType = "errors"
	ReadUptime        WsMessageType = "uptime"
	ReadResponseTimes WsMessageType = "rt"
	// ReadLive re-pulls the same services/status/TLS payload the hub
	// broadcasts (and sends on connect), wrapped in a reply envelope.
	ReadLive      WsMessageType = "live"
	Fire          WsMessageType = "fire"
	ServiceUpdate WsMessageType = "serviceUpdate"
)

type ErrorParams struct {
	Limit     int    `json:"limit"`
	SortOrder string `json:"sortOrder"`
}

type ResponseTimeParams struct {
	// 0: full snapshot, 1: snapshot of [from, to]
	Type byte   `json:"type"`
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
	// ID is chosen by the client and echoed back on the reply so it can
	// match replies to requests (several can be in flight at once).
	ID           uint32             `json:"id"`
	Type         WsMessageType      `json:"type"`
	ErrParams    ErrorParams        `json:"errParams"`
	RtParams     ResponseTimeParams `json:"rtParams"`
	UptimeParams UptimeParams       `json:"uptimeParams"`
	Services     []utils.Service    `json:"services"`
}

// replyEnvelope wraps every JSON reply to a client request. Exactly one of
// Data or Error is meaningful. Hub broadcasts are sent unwrapped, so clients
// tell the two apart by the presence of "id".
type replyEnvelope struct {
	ID    uint32        `json:"id"`
	Type  WsMessageType `json:"type"`
	Data  any           `json:"data"`
	Error string        `json:"error,omitempty"`
}

func (ws *wsConn) replyJSON(id uint32, typ WsMessageType, data any) {
	b, err := json.Marshal(replyEnvelope{ID: id, Type: typ, Data: data})
	if err != nil {
		slog.Error("failed to marshal websocket reply", "type", typ, "ERROR", err)
		ws.replyError(id, typ, errors.New("failed to encode reply"))
		return
	}
	ws.enqueue(websocket.TextMessage, b)
}

func (ws *wsConn) replyError(id uint32, typ WsMessageType, err error) {
	b, mErr := json.Marshal(replyEnvelope{ID: id, Type: typ, Error: err.Error()})
	if mErr != nil {
		slog.Error("failed to marshal websocket error reply", "ERROR", mErr)
		return
	}
	ws.enqueue(websocket.TextMessage, b)
}

// replyFrame sends a response time frame as a binary message prefixed with
// the request id (u32 little-endian), followed by the encodeFrame bytes.
func (ws *wsConn) replyFrame(id uint32, data utils.ResponseTimePayload) {
	frame := encodeFrame(data.Typ, data.Seq, data.Names, data.Ts, data.Cols)
	b := make([]byte, 4+len(frame))
	binary.LittleEndian.PutUint32(b, id)
	copy(b[4:], frame)
	ws.enqueue(websocket.BinaryMessage, b)
}

func (ws *wsConn) handleRead(t int, p []byte) {
	if t != websocket.TextMessage {
		// requests are JSON text, ignore anything else
		return
	}

	var msg readMessage
	if err := json.Unmarshal(p, &msg); err != nil {
		slog.Error("error reading message from websocket client", "ERROR", err)
		return
	}
	slog.Debug("websocket request", "id", msg.ID, "type", msg.Type)

	switch msg.Type {
	case ReadErrors:
		data, err := utils.GetErrorData(ws.db, msg.ErrParams.Limit, msg.ErrParams.SortOrder)
		if err != nil {
			slog.Error("failed to get error data for websocket client", "ERROR", err)
			ws.replyError(msg.ID, msg.Type, errors.New("failed to get error data"))
			return
		}
		ws.replyJSON(msg.ID, msg.Type, data)
	case ReadResponseTimes:
		data, err := ws.responseTimes(msg.RtParams)
		if err != nil {
			slog.Error("failed to build response times for websocket client", "ERROR", err)
			ws.replyError(msg.ID, msg.Type, err)
			return
		}
		ws.replyFrame(msg.ID, data)
	case ReadLive:
		data, err := utils.GetLiveData(ws.db)
		if err != nil {
			slog.Error("failed to get live data for websocket client", "ERROR", err)
			ws.replyError(msg.ID, msg.Type, errors.New("failed to get live data"))
			return
		}
		ws.replyJSON(msg.ID, msg.Type, data)
	case ReadUptime:
		fetch, ok := uptimeFetchers[msg.UptimeParams.Range]
		if !ok {
			slog.Error("websocket client requested uptime with bad range", "range", msg.UptimeParams.Range)
			ws.replyError(msg.ID, msg.Type, fmt.Errorf("invalid uptime range %q", msg.UptimeParams.Range))
			return
		}
		data := []utils.AverageData{}
		for _, endpoint := range utils.GetServiceEndpoints() {
			upAvg, err := fetch(ws.db, endpoint.Name)
			if err != nil {
				slog.Error("failed to get uptime data for websocket client", "service", endpoint.Name, "ERROR", err)
				ws.replyError(msg.ID, msg.Type, errors.New("failed to get uptime data"))
				return
			}
			// nil means the uptime calculation had nothing usable, skip it
			// (same behavior as UptimeAPI)
			if upAvg != nil {
				data = append(data, utils.AverageData{Name: endpoint.Name, Average: *upAvg})
			}
		}
		ws.replyJSON(msg.ID, msg.Type, data)
	case Fire:
		ws.scd.Fire()
	case ServiceUpdate:
	// TODO: implement
	default:
		ws.replyError(msg.ID, msg.Type, fmt.Errorf("unknown request type %q", msg.Type))
	}
}

// responseTimes builds the snapshot a "rt" request asks for: type 0 is the
// full history, type 1 is the [from, to] window (unix seconds, `to`
// defaults to now). Type 1 serves both windowed views and appends.
func (ws *wsConn) responseTimes(p ResponseTimeParams) (utils.ResponseTimePayload, error) {
	switch p.Type {
	case 0:
		return utils.BuildRtSnapshot(ws.db)
	case 1:
		// TODO: factor out of datastore.go to be a general util helper for accepting from/to
		// as string to convert to unix time format
		fromUnix, err := strconv.ParseInt(p.From, 10, 64)
		if err != nil || fromUnix < 0 {
			return utils.ResponseTimePayload{}, fmt.Errorf("invalid from %q", p.From)
		}
		from := time.Unix(fromUnix, 0)
		to := time.Now()
		if p.To != "" {
			toUnix, err := strconv.ParseInt(p.To, 10, 64)
			if err != nil {
				return utils.ResponseTimePayload{}, fmt.Errorf("invalid to %q", p.To)
			}
			// Stored timestamps carry sub-second precision, so stretch `to` to the
			// end of its second; otherwise rows later in that second are dropped.
			to = time.Unix(toUnix, 0).Add(time.Second - time.Nanosecond)
		}
		if from.After(to) {
			return utils.ResponseTimePayload{}, errors.New("'from' must be at or before 'to'")
		}
		return utils.BuildRtSnapshotRange(ws.db, from, to)
	default:
		return utils.ResponseTimePayload{}, fmt.Errorf("invalid response time request type %d", p.Type)
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

// liveSnapshot is the hub broadcast payload built on demand, for seeding a
// newly connected client.
func liveSnapshot(db *sql.DB) ([]byte, error) {
	live, err := utils.GetLiveData(db)
	if err != nil {
		return nil, err
	}
	return json.Marshal(live)
}
