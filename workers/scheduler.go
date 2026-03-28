package workers

import (
	"database/sql"
	"goUp/utils"
	"log"
	"strings"
	"time"
)

type GetState struct {
	res chan utils.ScheduleState
}

type Scheduler struct {
	state chan utils.ScheduleState
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

const defaultSpan = 30
const defaultInterval = "seconds"

func NewScheduler(db *sql.DB, cfg *utils.Config) *Scheduler {
	span := defaultSpan
	interval := defaultInterval

	if cfg != nil && cfg.Schedule != nil {
		span = cfg.Schedule.Span
		interval = cfg.Schedule.Interval
	}

	s := &Scheduler{
		state: make(chan utils.ScheduleState),
		get:   make(chan GetState),
		stop:  make(chan struct{}),
	}

	go s.StartScheduler(db, span, interval)

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
			req.res <- utils.ScheduleState{Span: Span, Interval: Interval}
		case <-timer.C:
			// time to fetch data
			data, err := utils.GetServiceData()
			if err != nil {
				log.Printf("Failed fetching service data in scheduler: %v\n", err)
				continue
			}
			for i := range data.AllServices {
				if err := utils.InsertData(db, data.AllServices[i]); err != nil {
					log.Printf("Failed to insert data: %v", err)
				}
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
			log.Printf("Failed to persist schedule to config: %v", err)
		}
	}

	s.state <- utils.ScheduleState{
		Span:     state.Span,
		Interval: state.Interval,
	}

	return true
}

func (s *Scheduler) Get() utils.ScheduleState {
	res := make(chan utils.ScheduleState)
	s.get <- GetState{res: res}
	return <-res
}

func (s *Scheduler) Stop() {
	close(s.stop)
}
