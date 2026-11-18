package server

import (
	"encoding/binary"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"goUp/utils"

	"github.com/gorilla/websocket"
)

func dialTestWs(t *testing.T) *websocket.Conn {
	t.Helper()
	dbPath := "./test_ws.db"
	utils.Current_Config = &utils.Config{
		Database_Location: &dbPath,
		Schedule:          &utils.ScheduleState{Span: 60, Interval: "seconds"},
	}
	utils.SetServiceEndpoints([]utils.Service{{Name: "test_service", URL: "http://example.com"}})
	db, err := utils.InitDB()
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	hub := NewHub()
	go hub.Run()
	s := &Server{db: db, hub: hub}
	ts := httptest.NewServer(http.HandlerFunc(s.handleWs))
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(ts.URL, "http"), nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	// first message is the live data seed
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	typ, b, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read seed: %v", err)
	}
	var seed utils.ServiceResponse
	if typ != websocket.TextMessage || json.Unmarshal(b, &seed) != nil || strings.Contains(string(b), `"id"`) {
		t.Fatalf("want unwrapped live data seed, got %d %s", typ, b)
	}
	t.Cleanup(func() {
		_ = conn.Close()
		ts.Close()
		_ = db.Close()
		_ = os.Remove(dbPath)
		utils.Current_Config = nil
		utils.SetServiceEndpoints([]utils.Service{})
	})
	return conn
}

func request(t *testing.T, conn *websocket.Conn, req string) (int, []byte) {
	t.Helper()
	if err := conn.WriteMessage(websocket.TextMessage, []byte(req)); err != nil {
		t.Fatalf("write: %v", err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	typ, b, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	return typ, b
}

func TestWsRequests_EchoID(t *testing.T) {
	conn := dialTestWs(t)

	for _, tc := range []struct {
		req     string
		id      uint32
		wantErr bool
	}{
		{`{"id":7,"type":"live"}`, 7, false},
		{`{"id":1,"type":"errors","errParams":{"limit":10,"sortOrder":"desc"}}`, 1, false},
		{`{"id":2,"type":"uptime","uptimeParams":{"range":"day"}}`, 2, false},
		{`{"id":3,"type":"uptime","uptimeParams":{"range":"bogus"}}`, 3, true},
		{`{"id":4,"type":"rt","rtParams":{"type":1,"from":"abc"}}`, 4, true},
		{`{"id":5,"type":"nope"}`, 5, true},
	} {
		typ, b := request(t, conn, tc.req)
		if typ != websocket.TextMessage {
			t.Fatalf("%s: want text reply, got %d", tc.req, typ)
		}
		var env struct {
			ID    uint32          `json:"id"`
			Data  json.RawMessage `json:"data"`
			Error string          `json:"error"`
		}
		if err := json.Unmarshal(b, &env); err != nil {
			t.Fatalf("%s: unmarshal: %v", tc.req, err)
		}
		if env.ID != tc.id {
			t.Errorf("%s: id = %d, want %d", tc.req, env.ID, tc.id)
		}
		if (env.Error != "") != tc.wantErr {
			t.Errorf("%s: error = %q, wantErr %v", tc.req, env.Error, tc.wantErr)
		}
	}

	for _, req := range []string{
		`{"id":6,"type":"rt","rtParams":{"type":0}}`,
		`{"id":6,"type":"rt","rtParams":{"type":1,"from":"0"}}`,
	} {
		typ, b := request(t, conn, req)
		if typ != websocket.BinaryMessage {
			t.Fatalf("%s: want binary reply, got %d: %s", req, typ, b)
		}
		if id := binary.LittleEndian.Uint32(b); id != 6 {
			t.Errorf("%s: id prefix = %d, want 6", req, id)
		}
		if _, err := DecodeFrame(b[4:]); err != nil {
			t.Errorf("%s: decode frame: %v", req, err)
		}
	}
}
