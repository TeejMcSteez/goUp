package workers_test

import (
	"bytes"
	"database/sql"
	"goUp/utils"
	"goUp/workers"
	"log/slog"
	"sync"
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

// TestSchedulerFireRunsFetchCycle verifies that Fire() drives a fetch cycle
// through to completion: the fire case spawns the fetch goroutine, and the
// fetchDone case observes its completion and clears the fetching flag,
// leaving the scheduler's select loop free to keep servicing requests.
func TestSchedulerFireRunsFetchCycle(t *testing.T) {
	withEmptyServiceConfig(t)

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
	withEmptyServiceConfig(t)

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

	// Wait for the fetch goroutine to actually finish reading Current_Config
	// before the test returns and withEmptyServiceConfig's cleanup restores
	// it out from under an in-flight fetch.
	waitForLog(t, buf, "Failed fetching service data in scheduler", 2*time.Second)
}

// TestSchedulerFireSkipsOverlappingFetch verifies the fetching guard: firing
// again while a previous fetch is still in flight logs a "skip" instead of
// starting a second concurrent fetch cycle.
func TestSchedulerFireSkipsOverlappingFetch(t *testing.T) {
	withEmptyServiceConfig(t)

	buf := &syncBuffer{}
	prevLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(buf, nil)))
	defer slog.SetDefault(prevLogger)

	s := newTestScheduler(t)
	defer s.Stop()

	// Fire() blocks on the unbuffered s.fire channel until the select loop
	// is ready to receive, so these sends are serialized by the loop one
	// iteration at a time. The fetch goroutine spawned by the first Fire()
	// needs to be scheduled and run before it clears the fetching flag, so
	// firing again immediately after should observe fetching still true.
	s.Fire()
	s.Fire()

	waitForLog(t, buf, "Previous service data fetch still running, skipping this tick", 2*time.Second)

	// Wait for the one real fetch goroutine to finish reading Current_Config
	// before the test returns and withEmptyServiceConfig's cleanup restores
	// it out from under it.
	waitForLog(t, buf, "Failed fetching service data in scheduler", 2*time.Second)

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
