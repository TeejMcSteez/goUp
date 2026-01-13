package workers

import (
	"database/sql"
	"goUp/utils"
	"log"
	"strings"
	"time"
)

type ScheduleState struct {
	// Number (30, 60, etc.)
	Span int `json:"timespan"`
	// Interval of time (sec, min, hrs)
	Interval string `json:"interval"`
}

type GetState struct {
	res chan ScheduleState
}

type Scheduler struct {
	state chan ScheduleState
	get   chan GetState
	stop  chan struct{}
}

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
		panic("Invalid Interval (expected seconds/minutes/hours)")
	}
}

func NewScheduler(db *sql.DB, initialSpan int, initialInterval string) *Scheduler {
	s := &Scheduler{
		state: make(chan ScheduleState),
		get:   make(chan GetState),
		stop:  make(chan struct{}),
	}

	go s.StartScheduler(db, initialSpan, initialInterval)

	return s
}

func (s *Scheduler) StartScheduler(db *sql.DB, Span int, Interval string) {
	log.Println("Starting service data scheduler")

	dur := computeDuration(Span, Interval)
	timer := time.NewTimer(dur)

	for {
		select {
		case <-s.stop:
			// clean shutdown: stop timer and drain channel if needed
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			log.Println("Scheduler stopped")
			return

		case upd := <-s.state:
			// update schedule parameters
			Span = upd.Span
			Interval = upd.Interval
			dur = computeDuration(Span, Interval)

			// interrupt current wait and restart with new duration
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(dur)
			log.Println("Schedule updated to:", Span, Interval)
		case req := <-s.get:
			req.res <- ScheduleState{Span: Span, Interval: Interval}
		case <-timer.C:
			// time to fetch data
			data, err := utils.GetServiceData()
			if err != nil {
				log.Printf("Failed fetching service data in scheduler: %v\n", err)
				continue
			}
			for i := range data.AllServices {
				utils.InsertData(db, data.AllServices[i])
			}
			checkedData, err := utils.Check(data.AllServices)
			if err != nil {
				log.Printf("Received error from checking data in scheduler: %v", err)
				continue
			}
			if len(checkedData) > 0 {
				utils.Current_Config.Triggers.Fire(checkedData)
			}
			log.Println("Scheduler fetched service data successfully")

			// schedule next run based on the *current* Span/Interval
			dur = computeDuration(Span, Interval)
			timer.Reset(dur)
		}
	}
}

func (s *Scheduler) Update(state ScheduleState) bool {
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

	s.state <- ScheduleState{
		Span:     state.Span,
		Interval: state.Interval,
	}

	return true
}

func (s *Scheduler) Get() ScheduleState {
	res := make(chan ScheduleState)
	s.get <- GetState{res: res}
	return <-res
}

func (s *Scheduler) Stop() {
	close(s.stop)
}
