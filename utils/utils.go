package utils

import "time"

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
