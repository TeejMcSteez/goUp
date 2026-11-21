package workers_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"goUp/utils"
	"goUp/workers"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// newTestStore returns a ConfigStore seeded with cfg, backed by a yaml file in
// t.TempDir() (removed automatically). It goes through LoadFromFile so the
// store is built exactly like production.
func newTestStore(t *testing.T, cfg *utils.Config) *utils.ConfigStore {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yml")
	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	store := utils.NewConfigStore(path)
	if err := store.LoadFromFile(); err != nil {
		t.Fatalf("Failed to load test config store: %v", err)
	}
	return store
}

// newTestStoreYAML is newTestStore from raw YAML, for configs a marshalled
// Config can't express (e.g. no services key at all -> nil Services).
func newTestStoreYAML(t *testing.T, content string) *utils.ConfigStore {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	store := utils.NewConfigStore(path)
	if err := store.LoadFromFile(); err != nil {
		t.Fatalf("Failed to load test config store: %v", err)
	}
	return store
}

// newTestScheduler creates a scheduler with a long initial duration so the
// timer never fires during tests. The config has no services key, so a fetch
// cycle fails fast in Scanner.GetServiceData (NoServiceEndpointsError) instead of
// making network calls. Callers must defer s.Stop().
func newTestScheduler(t *testing.T) *workers.Scheduler {
	t.Helper()
	var db *sql.DB
	store := newTestStoreYAML(t, "schedule:\n  span: 60\n  interval: minutes\n")
	s := workers.NewScheduler(t.Context(), db, store, nil, utils.NewScanner())
	return s
}

// syncBuffer is a concurrency-safe io.Writer, since fetch cycles run on a
// separate goroutine from the test and both may log concurrently.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

// waitForLog polls buf until it contains substr or the timeout elapses.
func waitForLog(t *testing.T, buf *syncBuffer, substr string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if bytes.Contains([]byte(buf.String()), []byte(substr)) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("Expected log output to contain %q within %s, got:\n%s", substr, timeout, buf.String())
}

func TestSchedulerGetInitialState(t *testing.T) {
	s := newTestScheduler(t)
	defer s.Stop()

	state := s.Get()
	if state.Span != 60 {
		t.Errorf("Expected Span=60, got %d", state.Span)
	}
	if state.Interval != "minutes" {
		t.Errorf("Expected Interval='minutes', got %q", state.Interval)
	}
}

func TestSchedulerUpdateAndGet(t *testing.T) {
	s := newTestScheduler(t)
	defer s.Stop()

	ok := s.Update(utils.ScheduleState{Span: 15, Interval: "seconds"})
	if !ok {
		t.Fatal("Expected Update to return true for valid state")
	}

	state := s.Get()
	if state.Span != 15 {
		t.Errorf("Expected Span=15, got %d", state.Span)
	}
	if state.Interval != "seconds" {
		t.Errorf("Expected Interval='seconds', got %q", state.Interval)
	}
}

func TestSchedulerUpdateInvalidSpan(t *testing.T) {
	s := newTestScheduler(t)
	defer s.Stop()

	cases := []struct {
		span int
		desc string
	}{
		{0, "span=0 (below minimum)"},
		{61, "span=61 (above maximum)"},
		{-1, "span=-1 (negative)"},
	}
	for _, tc := range cases {
		if s.Update(utils.ScheduleState{Span: tc.span, Interval: "minutes"}) {
			t.Errorf("Expected Update to return false for %s", tc.desc)
		}
	}
}

func TestSchedulerUpdateInvalidInterval(t *testing.T) {
	s := newTestScheduler(t)
	defer s.Stop()

	cases := []struct {
		interval string
		desc     string
	}{
		{"", "empty interval"},
		{"weeks", "unsupported unit 'weeks'"},
		{"days", "unsupported unit 'days'"},
		{"x", "unknown single char"},
	}
	for _, tc := range cases {
		if s.Update(utils.ScheduleState{Span: 5, Interval: tc.interval}) {
			t.Errorf("Expected Update to return false for %s", tc.desc)
		}
	}
}

func TestSchedulerUpdateValidIntervalVariants(t *testing.T) {
	s := newTestScheduler(t)
	defer s.Stop()

	// All these should be accepted since the scheduler only checks the first byte
	validIntervals := []string{
		"s", "seconds", "Seconds", "SECONDS",
		"m", "minutes", "Minutes", "MINUTES",
		"h", "hours", "Hours", "HOURS",
	}
	for _, interval := range validIntervals {
		if !s.Update(utils.ScheduleState{Span: 1, Interval: interval}) {
			t.Errorf("Expected Update to return true for interval %q", interval)
		}
	}
}

// TestSchedulerFireRunsFetchCycle verifies that Fire() drives a fetch cycle
// through to completion: the fire case spawns the fetch goroutine, and the
// fetchDone case observes its completion and clears the fetching flag,
// leaving the scheduler's select loop free to keep servicing requests.
func TestSchedulerFireRunsFetchCycle(t *testing.T) {
	buf := &syncBuffer{}
	prevLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(buf, nil)))
	defer slog.SetDefault(prevLogger)

	s := newTestScheduler(t)
	defer s.Stop()

	s.Fire()

	// With no service endpoints configured, GetServiceData fails fast via
	// NoServiceEndpointsError, so runFetchCycle logs this specific error.
	waitForLog(t, buf, "Failed fetching service data in scheduler", 2*time.Second)

	// The fetchDone signal must have been processed by the select loop for
	// this to return; otherwise the loop would still be blocked mid-cycle.
	state := s.Get()
	if state.Span != 60 {
		t.Errorf("Expected Span=60 after Fire() completed, got %d", state.Span)
	}
}

// TestSchedulerFireDoesNotBlockScheduler verifies that handling the fire
// case is non-blocking (the actual fetch runs on its own goroutine), so the
// scheduler's select loop stays responsive to Get() immediately after.
func TestSchedulerFireDoesNotBlockScheduler(t *testing.T) {
	buf := &syncBuffer{}
	prevLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(buf, nil)))
	defer slog.SetDefault(prevLogger)

	s := newTestScheduler(t)
	defer s.Stop()

	s.Fire()

	done := make(chan utils.ScheduleState, 1)
	go func() { done <- s.Get() }()

	select {
	case <-done:
		// success
	case <-time.After(1 * time.Second):
		t.Fatal("Get() did not return after Fire(), scheduler select loop appears stuck")
	}

	// Wait for the fetch goroutine to actually finish before the test returns
	// and the test's logger is restored under it.
	waitForLog(t, buf, "Failed fetching service data in scheduler", 2*time.Second)
}

// TestSchedulerFireSkipsOverlappingFetch verifies the fetching guard: firing
// again while a previous fetch is still in flight logs a "skip" instead of
// starting a second concurrent fetch cycle.
//
// This needs a fetch slow enough to reliably still be running when the
// second Fire() lands - the "no service endpoints" fast-fail path used by
// the other Fire tests completes in-process in a handful of microseconds,
// far too fast to deterministically race against. So this test wires up a
// real (deliberately slow) HTTP endpoint and a real sqlite DB, matching how
// runFetchCycle behaves against an actual configured service.
func TestSchedulerFireSkipsOverlappingFetch(t *testing.T) {
	const fetchDelay = 300 * time.Millisecond

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(fetchDelay)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	dbPath := fmt.Sprintf("./scheduler_fire_skip_%s.db", t.Name())
	cfg := &utils.Config{
		Database_Location: &dbPath,
		Schedule:          &utils.ScheduleState{Span: 60, Interval: "minutes"},
		Services: map[string]utils.Service{
			"Slow": {Name: "Slow", URL: srv.URL},
		},
	}
	// The scanner reads endpoints from this config each cycle, so the
	// fetch always hits the slow server; no shared state to reset.

	db, err := utils.InitDB(cfg)
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("db.Close: %v", err)
		}
		if err := os.Remove(dbPath); err != nil {
			t.Errorf("os.Remove %s: %v", dbPath, err)
		}
	})

	buf := &syncBuffer{}
	prevLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(buf, nil)))
	defer slog.SetDefault(prevLogger)

	s := workers.NewScheduler(t.Context(), db, newTestStore(t, cfg), nil, utils.NewScanner())
	defer s.Stop()

	// Fire() blocks on the unbuffered s.fire channel until the select loop
	// is ready to receive, so these sends are serialized by the loop one
	// iteration at a time. The first Fire() kicks off a fetch that will
	// take at least fetchDelay to finish, so the second Fire(), landing
	// essentially immediately after, is guaranteed to see fetching still
	// true and skip instead of starting a second concurrent fetch cycle.
	s.Fire()
	s.Fire()

	waitForLog(t, buf, "Previous service data fetch still running, skipping this tick", 2*time.Second)

	// Wait for the one real fetch to finish before the test returns and
	// cleanup tears down the DB out from under it.
	waitForLog(t, buf, "Scheduler fetched service data successfully", 2*time.Second)

	// Scheduler must remain responsive after the overlapping fire.
	state := s.Get()
	if state.Span != 60 {
		t.Errorf("Expected Span=60 after overlapping Fire(), got %d", state.Span)
	}
}

func TestSchedulerStop(t *testing.T) {
	s := newTestScheduler(t)

	done := make(chan struct{})
	go func() {
		s.Stop()
		close(done)
	}()

	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Error("Stop() did not complete within timeout")
	}
}
