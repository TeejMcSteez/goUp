package workers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"goUp/utils"
	"log/slog"
	"net"
	"strings"
	"time"
)

type GetState struct {
	res chan utils.ScheduleState
}

// Broadcaster pushes fresh bytes out to connected subscribers (e.g. the
// server's websocket hub). Declared here, the consumer, rather than in
// server, so workers doesn't need to import it — anything with a matching
// Broadcast method satisfies this interface.
type Broadcaster interface {
	Broadcast(b []byte)
}

type Scheduler struct {
	ctx    context.Context
	cancel context.CancelFunc
	state  chan utils.ScheduleState
	get    chan GetState
	stop   chan struct{}
	fire   chan struct{}
	pub    Broadcaster
}

type fetchResult struct {
	hasFailed bool
	offline   bool
	// owedOffline is offline time not yet reported by a restored notice
	// (0 = nothing owed). Accumulates across outages whose notice failed.
	owedOffline time.Duration
}

func restoredNotice(offlineFor time.Duration) utils.Message {
	return utils.Message{
		Kind:  utils.KindInfo,
		Title: "Host connection restored",
		Body:  fmt.Sprintf("offline %s; alerts may have been missed", offlineFor.Round(time.Second)),
	}
}

// schedule next run based on the *current* Span/Interval
// FIX: #103
func computeDuration(Span int, Interval string) time.Duration {
	d := time.Duration(Span)
	switch strings.ToLower(Interval)[0] {
	case 's':
		return d * time.Second
	case 'm':
		return d * time.Minute
	case 'h':
		return d * time.Hour
	default:
		slog.Warn("Invalid interval, defaulting to 60 second fetch interval", "span", Span, "interval", Interval)
		return 60 * time.Second
	}
}

const defaultSpan = 30
const defaultInterval = "seconds"

func NewScheduler(ctx context.Context, db *sql.DB, cfg *utils.Config, pub Broadcaster) *Scheduler {
	ctx, cancel := context.WithCancel(ctx)
	span := defaultSpan
	interval := defaultInterval

	if cfg != nil && cfg.Schedule != nil {
		span = cfg.Schedule.Span
		interval = cfg.Schedule.Interval
	}

	s := &Scheduler{
		ctx:    ctx,
		cancel: cancel,
		state:  make(chan utils.ScheduleState),
		get:    make(chan GetState),
		stop:   make(chan struct{}),
		fire:   make(chan struct{}),
		pub:    pub,
	}

	go s.StartScheduler(db, span, interval)

	return s
}

func hostOnline(ctx context.Context) bool {
	d := net.Dialer{Timeout: 3 * time.Second}
	// TODO: allow configurable query
	conn, err := d.DialContext(ctx, "tcp", "1.1.1.1:53")
	if err != nil {
		return false
	}
	if err := conn.Close(); err != nil {
		slog.Error("error closing connection testing host network", "ERROR", err)
	}
	return true
}

// runFetchCycle fetches current service data, persists it, and fires any
// triggers for newly-down services. Split out of StartScheduler's select loop
// so it can run on its own goroutine — service checks can block for a while
// (slow endpoints, retries), and the select loop must stay free to service
// s.get/s.state/s.stop the whole time, or callers like GET/POST /api/schedule
// would hang until the fetch finishes.
func runFetchCycle(ctx context.Context, db *sql.DB, wasFailed bool, offlineSince time.Time, owedOffline time.Duration, pub Broadcaster) fetchResult {
	// owedOffline > 0 means a restored notice failed to send last cycle;
	// retry it before anything else. On failure keep it owed and carry on
	// with the fetch rather than returning, so a broken trigger can't halt
	// monitoring. Skipped while offlineSince is set: the host dropped again
	// before the notice went out, so the owed time rolls into the next
	// restored notice below instead.
	if owedOffline > 0 && offlineSince.IsZero() {
		if err := utils.Current_Config.Triggers.Send(ctx, restoredNotice(owedOffline)); err != nil {
			slog.Error("failed to send restored notice", "ERROR", err)
		} else {
			owedOffline = 0
		}
	}
	data, err := utils.GetServiceData()
	if err != nil {
		slog.Error("Failed fetching service data in scheduler", "error", err)
		// probe instead of defaulting offline to false, which would clear
		// offlineSince and log a false "host back online"
		return fetchResult{hasFailed: wasFailed, offline: !hostOnline(ctx), owedOffline: owedOffline}
	}

	// Down services + no host network means the outage is ours, not theirs.
	// Skip persist/broadcast/send so the DB isn't filled with false downs;
	// the resulting gap in history marks the connectivity loss. hasFailed is
	// left as-is so a real outage is alerted on once the host is back.
	if len(data.DownServices) > 0 && !hostOnline(ctx) {
		slog.Warn("services unreachable and host offline, skipping cycle")
		return fetchResult{hasFailed: wasFailed, offline: true, owedOffline: owedOffline}
	}

	// Past the guard the host is confirmed online. If it was offline coming
	// in, send the restored notice now so it lands before any alert/recovery
	// this cycle sends. The notice reports this outage plus any earlier
	// unreported ones; on failure the running total stays owed.
	if !offlineSince.IsZero() {
		total := owedOffline + time.Since(offlineSince)
		if err := utils.Current_Config.Triggers.Send(ctx, restoredNotice(total)); err != nil {
			slog.Error("failed to send restored notice", "ERROR", err)
			owedOffline = total
		} else {
			owedOffline = 0
		}
	}

	if err := utils.PersistCycle(db, data.AllServices, data.TlsData); err != nil {
		slog.Error("failed to persist fetch cycle in database", "error", err)
	}

	if pub != nil {
		broadcastLiveData(db, pub)
	}
	// Fire when services are down
	if len(data.DownServices) > 0 {
		if err := utils.Current_Config.Triggers.Send(ctx, utils.Message{
			Kind:     utils.KindAlert,
			Title:    "Services are down",
			Body:     "",
			Services: data.DownServices,
		}); err != nil {
			slog.Error("error sending notification for failure", "ERROR", err)
			return fetchResult{hasFailed: true, offline: !hostOnline(ctx), owedOffline: owedOffline}
		}
		return fetchResult{hasFailed: true, offline: false, owedOffline: owedOffline}
	}
	// Only fires on past errors with no new failures
	if wasFailed {
		if err := utils.Current_Config.Triggers.Send(ctx, utils.Message{
			Kind:     utils.KindClear,
			Title:    "Services have recovered",
			Body:     "All services are back online",
			Services: data.AllServices,
		}); err != nil {
			slog.Error("error sending notification for recovery", "ERROR", err)
			return fetchResult{hasFailed: true, offline: !hostOnline(ctx), owedOffline: owedOffline}
		}
	}

	slog.Info("Scheduler fetched service data successfully")
	return fetchResult{hasFailed: false, offline: false, owedOffline: owedOffline}
}

// broadcastLiveData pushes the freshly persisted state read back from the DB
// rather than the raw fetch result, so the payload matches what the REST
// endpoints return (formatted response times, inactive services, stored TLS).
func broadcastLiveData(db *sql.DB, pub Broadcaster) {
	live, err := utils.GetLiveData(db)
	if err != nil {
		slog.Error("failed to read live data for broadcast", "error", err)
		return
	}
	b, err := json.Marshal(live)
	if err != nil {
		slog.Error("failed to marshal service data for broadcast", "error", err)
		return
	}
	pub.Broadcast(b)
}

func (s *Scheduler) StartScheduler(db *sql.DB, Span int, Interval string) {
	slog.Info("Starting service data scheduler")

	dur := computeDuration(Span, Interval)
	timer := time.NewTimer(dur)
	fetching := false
	hasFailed := false
	var offlineSince time.Time
	var owedOffline time.Duration
	fetchDone := make(chan fetchResult, 1) // buffered goroutine never blocks on send
	// starts fetchCycle
	startFetch := func() {
		if fetching {
			slog.Info("Previous service data fetch still running, skipping this tick")
			return
		}
		fetching = true
		go func(wasFailed bool, offlineSince time.Time, owedOffline time.Duration) {
			fetchDone <- runFetchCycle(s.ctx, db, wasFailed, offlineSince, owedOffline, s.pub)
		}(hasFailed, offlineSince, owedOffline) // snapshot loop-owned state
	}

	for {
		select {
		case <-s.stop:
			utils.DrainTimer(timer)
			slog.Info("Scheduler stopped")
			return

		case upd := <-s.state:
			// update schedule parameters
			Span = upd.Span
			Interval = upd.Interval
			// compute new duration
			dur = computeDuration(Span, Interval)
			utils.DrainTimer(timer)
			timer.Reset(dur)
			slog.Info("Schedule updated", "span", Span, "interval", Interval)
		case req := <-s.get:
			req.res <- utils.ScheduleState{Span: Span, Interval: Interval}
		case res := <-fetchDone:
			fetching = false
			hasFailed = res.hasFailed
			owedOffline = res.owedOffline
			switch {
			case res.offline && offlineSince.IsZero():
				offlineSince = time.Now()
				slog.Warn("host appears offline", "since", offlineSince)
			case !res.offline && !offlineSince.IsZero():
				timeOffline := time.Since(offlineSince)
				offlineSince = time.Time{}
				// restored notice is sent by runFetchCycle
				slog.Info("host back online", "down for", timeOffline)
			}
		case <-s.fire:
			utils.DrainTimer(timer)
			hasFailed = false
			startFetch()
			dur = computeDuration(Span, Interval)
			timer.Reset(dur)
		case <-timer.C:
			startFetch()
			dur = computeDuration(Span, Interval)
			timer.Reset(dur)
		}
	}
}

func (s *Scheduler) Update(state utils.ScheduleState) bool {
	if state.Span < 1 || state.Span > 60 {
		return false
	}

	low := strings.ToLower(state.Interval)
	if len(low) == 0 {
		return false
	}

	first := low[0]
	if first != 's' && first != 'm' && first != 'h' {
		return false
	}

	if utils.Current_Config != nil {
		if err := utils.UpdateConfigSchedule(utils.Current_Config, state); err != nil {
			slog.Error("Failed to persist schedule to config", "error", err)
		}
	}

	s.state <- utils.ScheduleState{
		Span:     state.Span,
		Interval: state.Interval,
	}

	return true
}

func (s *Scheduler) Fire() {
	s.fire <- struct{}{}
}

func (s *Scheduler) Get() utils.ScheduleState {
	res := make(chan utils.ScheduleState)
	s.get <- GetState{res: res}
	return <-res
}

func (s *Scheduler) Stop() {
	s.cancel()
	close(s.stop)
}
