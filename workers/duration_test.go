package workers

import (
	"math"
	"testing"
	"time"
)

func TestComputeDuration(t *testing.T) {
	const fallback = 60 * time.Second
	tests := []struct {
		name     string
		span     int
		interval string
		want     time.Duration
	}{
		{"full word seconds", 30, "seconds", 30 * time.Second},
		{"full word minutes", 5, "minutes", 5 * time.Minute},
		{"full word hours", 2, "hours", 2 * time.Hour},
		{"single letter", 15, "m", 15 * time.Minute},
		{"uppercase", 10, "Seconds", 10 * time.Second},
		{"zero span", 0, "seconds", fallback},
		{"negative span", -1, "minutes", fallback},
		{"empty interval", 5, "", fallback},
		{"unknown interval", 5, "days", fallback},
		{"at cap", 24, "hours", maxScheduleDuration},
		{"over cap", 25, "hours", maxScheduleDuration},
		{"overflowing span", math.MaxInt, "hours", maxScheduleDuration},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := computeDuration(tc.span, tc.interval); got != tc.want {
				t.Errorf("computeDuration(%d, %q) = %v, want %v", tc.span, tc.interval, got, tc.want)
			}
		})
	}
}
