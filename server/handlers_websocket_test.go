package server

import (
	"encoding/binary"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"goUp/utils"

	"github.com/gorilla/websocket"
)

func dialTestWs(t *testing.T) (*websocket.Conn, *utils.ConfigStore) {
	t.Helper()
	dbPath := "./test_ws.db"
	cfgPath := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(cfgPath, []byte("db_path: "+dbPath+"\nschedule:\n  span: 60\n  interval: seconds\nservices:\n  test_service:\n    url: http://example.com\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	store := utils.NewConfigStore(cfgPath)
	if err := store.LoadFromFile(); err != nil {
		t.Fatalf("load config: %v", err)
	}
	db, err := utils.InitDB(store.Get())
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	hub := NewHub()
	go hub.Run()
	s := &Server{db: db, hub: hub, store: store}
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
	})
	return conn, store
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
	conn, _ := dialTestWs(t)

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

type wsReply struct {
	ID    uint32          `json:"id"`
	Type  WsMessageType   `json:"type"`
	Data  json.RawMessage `json:"data"`
	Error string          `json:"error"`
}

// requestSync sends req followed by a "live" request and returns the replies
// that arrive before the live one. A connection's requests are handled in
// order, so once live is answered req has been fully processed, whether or
// not it produced a reply of its own.
func requestSync(t *testing.T, conn *websocket.Conn, req string) []wsReply {
	t.Helper()
	const syncID = 9999
	if err := conn.WriteMessage(websocket.TextMessage, []byte(req)); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, []byte(`{"id":9999,"type":"live"}`)); err != nil {
		t.Fatalf("write sync: %v", err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var replies []wsReply
	for {
		_, b, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		var r wsReply
		if err := json.Unmarshal(b, &r); err != nil {
			t.Fatalf("unmarshal %s: %v", b, err)
		}
		if r.ID == syncID {
			return replies
		}
		replies = append(replies, r)
	}
}

// upsertRequest builds a ServiceUpsert request with id 8.
func upsertRequest(t *testing.T, name string, svc utils.Service) string {
	t.Helper()
	b, err := json.Marshal(readMessage{ID: 8, Type: ServiceUpsert, UpdateParams: UpdateParams{Name: name, NewService: svc}})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	return string(b)
}

// wantUpsertOK fails unless replies is exactly one success reply.
func wantUpsertOK(t *testing.T, replies []wsReply) {
	t.Helper()
	if len(replies) != 1 {
		t.Fatalf("got %d replies, want 1: %+v", len(replies), replies)
	}
	if r := replies[0]; r.ID != 8 || r.Type != ServiceUpsert || r.Error != "" || string(r.Data) != `{"ok":true}` {
		t.Fatalf("want ok reply to id 8, got %+v (data %s)", r, r.Data)
	}
}

func TestWsServiceUpsert(t *testing.T) {
	for _, tc := range []struct {
		name         string
		key          string
		svcName      string // Name inside newService, must always lose to key
		wantServices int
	}{
		{"update existing", "test_service", "test_service", 1},
		{"update without inner name", "test_service", "", 1},
		{"update with mismatched inner name", "test_service", "other", 1},
		{"insert new", "new_service", "new_service", 2},
		{"insert without inner name", "new_service", "", 2},
		{"insert with mismatched inner name", "new_service", "other", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			conn, store := dialTestWs(t)
			notified := make(chan *utils.Config, 1)
			store.Subscribe(func(_, cur *utils.Config) { notified <- cur })
			before := store.Get()

			desc, retry, active := "new desc", 3, false
			wantUpsertOK(t, requestSync(t, conn, upsertRequest(t, tc.key, utils.Service{
				Name: tc.svcName, URL: "http://upserted.example.com", Description: &desc, Retry_Requests: &retry, Active: &active,
			})))

			check := func(where string, cfg *utils.Config) {
				t.Helper()
				if len(cfg.Services) != tc.wantServices {
					t.Errorf("%s: %d services, want %d", where, len(cfg.Services), tc.wantServices)
				}
				svc, ok := cfg.Services[tc.key]
				if !ok {
					t.Fatalf("%s: %s missing", where, tc.key)
				}
				if svc.Name != tc.key {
					t.Errorf("%s: Name = %q, want %q", where, svc.Name, tc.key)
				}
				if svc.URL != "http://upserted.example.com" {
					t.Errorf("%s: URL = %q, want http://upserted.example.com", where, svc.URL)
				}
				if svc.Description == nil || *svc.Description != desc {
					t.Errorf("%s: Description = %v, want %q", where, svc.Description, desc)
				}
				if svc.Retry_Requests == nil || *svc.Retry_Requests != retry {
					t.Errorf("%s: Retry_Requests = %v, want %d", where, svc.Retry_Requests, retry)
				}
				if svc.IsActive() {
					t.Errorf("%s: service still active", where)
				}
				// an insert must leave the service that was already there alone
				if got := cfg.Services["test_service"].URL; tc.key != "test_service" && got != "http://example.com" {
					t.Errorf("%s: insert changed test_service URL to %q", where, got)
				}
			}

			check("store", store.Get())

			// a fresh store proves the upsert reached disk, not just memory
			onDisk := utils.NewConfigStore(store.Path())
			if err := onDisk.LoadFromFile(); err != nil {
				t.Fatalf("reload config: %v", err)
			}
			check("disk", onDisk.Get())

			select {
			case cfg := <-notified:
				check("subscriber", cfg)
			default:
				t.Error("subscribers not notified of upsert")
			}

			// snapshots handed out earlier must not change underneath their readers
			if len(before.Services) != 1 || before.Services["test_service"].URL != "http://example.com" {
				t.Errorf("old snapshot mutated: %+v", before.Services)
			}
		})
	}
}

func TestWsServiceUpsert_EmptyName(t *testing.T) {
	for _, tc := range []struct{ name, params string }{
		{"empty name", `{"name":"","newService":{"Name":"test_service","URL":"http://upserted.example.com"}}`},
		{"no params", `{}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			conn, store := dialTestWs(t)
			before := store.Get()
			fileBefore, err := os.ReadFile(store.Path())
			if err != nil {
				t.Fatalf("read config: %v", err)
			}

			replies := requestSync(t, conn, `{"id":9,"type":"`+string(ServiceUpsert)+`","updateParams":`+tc.params+`}`)
			if len(replies) != 1 {
				t.Fatalf("got %d replies, want 1: %+v", len(replies), replies)
			}
			if r := replies[0]; r.ID != 9 || r.Type != ServiceUpsert || r.Error == "" {
				t.Errorf("want error reply to id 9, got %+v", r)
			}

			if store.Get() != before {
				t.Errorf("rejected upsert published a new config: %+v", store.Get().Services)
			}
			fileAfter, err := os.ReadFile(store.Path())
			if err != nil {
				t.Fatalf("read config: %v", err)
			}
			if string(fileAfter) != string(fileBefore) {
				t.Errorf("rejected upsert changed config file:\n%s", fileAfter)
			}
		})
	}
}

// A config file without a services section loads with a nil Services map.
func TestWsServiceUpsert_NilServices(t *testing.T) {
	conn, store := dialTestWs(t)
	if err := store.Update(func(c *utils.Config) error {
		c.Services = nil
		return nil
	}); err != nil {
		t.Fatalf("clear services: %v", err)
	}

	wantUpsertOK(t, requestSync(t, conn, upsertRequest(t, "first", utils.Service{URL: "http://first.example.com"})))

	if got := store.Get().Services["first"].URL; got != "http://first.example.com" {
		t.Errorf("URL = %q, want http://first.example.com", got)
	}
}

// A service with long fields must fit under readLimit.
func TestWsServiceUpsert_LargeMessage(t *testing.T) {
	conn, store := dialTestWs(t)
	desc := strings.Repeat("a", 12<<10)

	wantUpsertOK(t, requestSync(t, conn, upsertRequest(t, "test_service", utils.Service{URL: "http://example.com", Description: &desc})))

	if got := store.Get().Services["test_service"].Description; got == nil || *got != desc {
		t.Error("long description not stored")
	}
}
