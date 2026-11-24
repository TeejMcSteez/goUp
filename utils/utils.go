package utils

import (
	"strconv"
	"time"
)

// DrainTimer stops t and clears any value already delivered to t.C, so the
// timer can be safely Reset. Call only from the goroutine that owns t; it is
// not safe if another goroutine may be receiving on t.C.
func DrainTimer(t *time.Timer) {
	// interrupt current wait and restart with new duration
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// Returns boolean on whether the current time if after the passed time t
func isExpired(t time.Time) bool {
	return time.Now().Compare(t) == 1
}

// parses a string returning its base 10, int64 value
//
// on error, returns 0 and error
func StrToInt(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
