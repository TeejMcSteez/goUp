package scheduler

import (
	"fmt"
	"goUp/utils"
	"strings"
	"sync"
	"time"
)

// Holds the scheduling parameters and a mutex.
type scheduleState struct {
	mu       sync.RWMutex
	span     int
	interval string
}

var schedule = scheduleState{
	span:     30,
	interval: "seconds",
}

// computeDuration converts the current span/interval to time.Duration.
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
		panic("invalid interval (expected something like seconds/minutes/hours)")
	}
}

// StartScheduler runs forever in a goroutine, fetching data on a schedule.
func StartScheduler(currData *utils.SharedData) {
	fmt.Println("Starting service data scheduler")

	go func() {
		for {
			// Read current parameters under read lock
			schedule.mu.RLock()
			span := schedule.span
			interval := schedule.interval
			schedule.mu.RUnlock()

			dur := computeDuration(span, interval)

			// Sleep for the computed duration
			time.Sleep(dur)

			// Do the work
			data := utils.GetServiceData()
			currData.Set(data)
			fmt.Println("Scheduler fetched service data successfully")
		}
	}()
}

// UpdateParameters changes the schedule.
func UpdateParameters(span int, interval string) bool {

	if span < 1 ||
		span > 60 {
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

	schedule.mu.Lock()
	defer schedule.mu.Unlock()

	schedule.span = span
	schedule.interval = interval

	fmt.Println("Schedule updated to:", span, interval)
	return true
}

// GetParameters returns a copy of the current parameters.
func GetParameters() utils.ParamtersData {
	schedule.mu.RLock()
	defer schedule.mu.RUnlock()

	return utils.ParamtersData{
		Span:     schedule.span,
		Interval: schedule.interval,
	}
}
