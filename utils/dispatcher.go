package utils

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type Dispatcher struct {
	store     *ConfigStore
	mu        sync.RWMutex
	lastFired map[string]time.Time
}

// shouldBackoff reports whether a trigger fired within its backoff window.
func shouldBackoff(name string, lastFired time.Time, backoffDuration time.Duration) bool {
	if backoffDuration > 0 && !lastFired.IsZero() && time.Since(lastFired) < backoffDuration {
		slog.Info("Trigger backoff period active, skipping", "trigger", name, "last_fired", time.Since(lastFired))
		return true
	}
	return false
}

// backoffErr wraps ErrBackoff with how long until the trigger can fire again.
func backoffErr(name string, lastFired time.Time, backoffDuration time.Duration) error {
	remaining := (backoffDuration - time.Since(lastFired)).Round(time.Millisecond)
	return fmt.Errorf("%s: %w (%s remaining)", name, ErrBackoff, remaining)
}

func NewDispatcher(store *ConfigStore) *Dispatcher {
	return &Dispatcher{store: store, lastFired: map[string]time.Time{}}
}
func (d *Dispatcher) Send(ctx context.Context, m Message) error {
	t := &d.store.Get().Triggers // current snapshot, read-only
	if len(t.notifiers) == 0 {
		return ErrNoTriggers
	}
	var ers []error
	for _, n := range t.notifiers {
		name := n.Name()
		if err := d.check(name, m.Kind, n.BackoffDuration()); err != nil {
			slog.Info("backoff active, skipping", "notifier", name, "INFO", err)
			continue
		}
		nctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		err := n.Send(nctx, m)
		cancel()
		if err != nil {
			ers = append(ers, err)
			continue
		}
		d.mark(name, m.Kind)
	}
	return errors.Join(ers...)
}

func (d *Dispatcher) check(name string, kind Kind, backoff time.Duration) error {
	if kind != KindAlert {
		return nil
	}
	d.mu.Lock()
	last := d.lastFired[name]
	d.mu.Unlock()
	if shouldBackoff(name, last, backoff) {
		return backoffErr(name, last, backoff)
	}
	return nil
}

func (d *Dispatcher) mark(name string, kind Kind) {
	if kind != KindAlert {
		return
	}
	d.mu.Lock()
	d.lastFired[name] = time.Now()
	d.mu.Unlock()
}
