package scheduler

import (
	"database/sql"
	"fmt"
	"goUp/utils"
	"log"
	"strings"
	"time"
)

type ScheduleState struct {
	Span     int
	Interval string
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

// TODO: Replace updating span and inteval from the sharedData struct to the sqlite db
func NewScheduler(db *sql.DB, initialSpan int, initialInterval string) *Scheduler {
	s := &Scheduler{
		state: make(chan ScheduleState),
		get:   make(chan GetState),
		stop:  make(chan struct{}),
	}

	go s.StartScheduler(db, initialSpan, initialInterval)

	return s
}

// TODO: Replace updating span and inteval from the sharedData struct to the sqlite db
func (s *Scheduler) StartScheduler(db *sql.DB, Span int, Interval string) {
	fmt.Println("Starting service data scheduler")

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
			fmt.Println("Scheduler stopped")
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
			fmt.Println("Schedule updated to:", Span, Interval)
		case req := <-s.get:
			req.res <- ScheduleState{Span: Span, Interval: Interval}
		case <-timer.C:
			// time to fetch data
			data := utils.GetServiceData()
			for _, d := range data.AllServices {
				utils.InsertData(db, d)
			}
			checkedData, err := utils.Check(data.AllServices)
			if err != nil {
				log.Print("Received error from checking data in scheduler")
				continue
			}
			if len(checkedData) > 0 {
				utils.Fire(checkedData)
			}
			fmt.Println("Scheduler fetched service data successfully")

			// schedule next run based on the *current* Span/Interval
			dur = computeDuration(Span, Interval)
			timer.Reset(dur)
		}
	}
}

// TODO: Add get handler in scheduler to copy current channel state to get on frontend
func (s *Scheduler) Update(Span int, Interval string) bool {
	if Span < 1 || Span > 60 {
		return false
	}

	low := strings.ToLower(Interval)
	if len(low) == 0 {
		return false
	}

	first := low[0]
	if first != 's' && first != 'm' && first != 'h' {
		return false
	}

	s.state <- ScheduleState{
		Span:     Span,
		Interval: Interval,
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
