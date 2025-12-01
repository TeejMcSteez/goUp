package scheduler

import (
	"fmt"
	"goUp/utils"
	"strings"
	"time"
)

// Holds the scheduling parameters and a mutex.
type scheduleState struct {
	span     int
	interval string
}

type Scheduler struct {
	state chan scheduleState
	stop  chan struct{}
}

func computeDuration(span int, interval string) time.Duration {
	d := time.Duration(span)
	switch strings.ToLower(interval)[0] {
	case 's':
		return d * time.Second
	case 'm':
		return d * time.Minute
	case 'h':
		return d * time.Hour
	default:
		panic("Invalid interval (expected seconds/minutes/hours)")
	}
}

func NewScheduler(currData *utils.SharedData, initialSpan int, initialInterval string) *Scheduler {
	s := &Scheduler{
		state: make(chan scheduleState),
		stop:  make(chan struct{}),
	}

	go s.StartScheduler(currData, initialSpan, initialInterval)

	return s
}

func (s *Scheduler) StartScheduler(currData *utils.SharedData, span int, interval string) {
	fmt.Println("Starting service data scheduler")

	dur := computeDuration(span, interval)
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
			fmt.Println("Scheduler stopping")
			return

		case upd := <-s.state:
			// update schedule parameters
			span = upd.span
			interval = upd.interval
			dur = computeDuration(span, interval)

			// interrupt current wait and restart with new duration
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(dur)
			fmt.Println("Schedule updated to:", span, interval)

		case <-timer.C:
			// time to fetch data
			data := utils.GetServiceData()
			currData.Set(data)
			fmt.Println("Scheduler fetched service data successfully")

			// schedule next run based on the *current* span/interval
			dur = computeDuration(span, interval)
			timer.Reset(dur)
		}
	}
}

// TODO: Add get handler in scheduler to copy current channel state to get on frontend

func (s *Scheduler) Update(span int, interval string) bool {
	if span < 1 || span > 60 {
		return false
	}

	low := strings.ToLower(interval)
	if len(low) == 0 {
		return false
	}

	first := low[0]
	if first != 's' && first != 'm' && first != 'h' {
		return false
	}

	s.state <- scheduleState{
		span:     span,
		interval: interval,
	}

	return true
}

func (s *Scheduler) Stop() {
	close(s.stop)
}
