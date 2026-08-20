package workers_test

import (
	"database/sql"
	"goUp/utils"
	"goUp/workers"
	"testing"
	"time"
)

// newTestScheduler creates a scheduler with a long initial duration so the
// timer never fires during tests. Callers must defer s.Stop().
func newTestScheduler(t *testing.T) *workers.Scheduler {
	t.Helper()
	var db *sql.DB
	cfg := &utils.Config{
		Schedule: &utils.ScheduleState{Span: 60, Interval: "minutes"},
	}
	s := workers.NewScheduler(db, cfg)
	return s
}

// withEmptyServiceConfig points utils.Current_Config at a config with no
// service endpoints configured, so a fetch cycle triggered via Fire() fails
// fast inside GetServiceData (NoServiceEndpointsError) instead of making
// real network calls. The previous config is restored on test cleanup.
func withEmptyServiceConfig(t *testing.T) {
	t.Helper()
	prev := utils.Current_Config
	utils.Current_Config = &utils.Config{
		Schedule: &utils.ScheduleState{Span: 60, Interval: "minutes"},
	}
	t.Cleanup(func() { utils.Current_Config = prev })
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
